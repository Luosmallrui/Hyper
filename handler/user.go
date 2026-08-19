package handler

import (
	"Hyper/config"
	"Hyper/middleware"
	"Hyper/models"
	"Hyper/pkg/context"
	"Hyper/pkg/response"
	"Hyper/pkg/utils"
	"Hyper/service"
	"Hyper/types"
	base "context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	Config         *config.Config
	UserService    service.IUserService
	OssService     service.IOssService
	FollowService  service.FollowService
	LikeService    service.LikeService
	CollectService service.CollectService
	NoteService    service.INoteService
	DB             *gorm.DB
}

func (u *User) RegisterRouter(r gin.IRouter) {
	authorize := middleware.Auth([]byte(u.Config.Jwt.Secret))
	optionalAuth := middleware.OptionalAuth([]byte(u.Config.Jwt.Secret))
	g := r.Group("/v1/user")
	g.POST("/info", authorize, context.Wrap(u.UpdateUserInfo))
	g.POST("/avatar", authorize, context.Wrap(u.UploadAvatar))
	g.GET("/info", optionalAuth, context.Wrap(u.GetUserInfo))
	g.GET("/customer-service", authorize, context.Wrap(u.GetCustomerServiceContact))
	g.GET("/note", optionalAuth, context.Wrap(u.GetUserNote))
	g.GET("/my-notes", authorize, context.Wrap(u.GetMyNotes))

}

// GetCustomerServiceContact exposes the configured platform customer-service
// account to signed-in clients. Message delivery remains the normal single
// chat flow; this endpoint avoids a fragile hard-coded peer ID in the client.
func (u *User) GetCustomerServiceContact(c *gin.Context) error {
	if u.DB == nil {
		return response.NewError(http.StatusInternalServerError, "客服配置服务不可用")
	}
	var setting models.PlatformSetting
	if err := u.DB.WithContext(c.Request.Context()).Where("setting_key = ?", "customer_service_user_id").First(&setting).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewError(http.StatusNotFound, "客服账号暂未配置")
		}
		return response.NewError(http.StatusInternalServerError, "读取客服配置失败")
	}
	userID, err := strconv.Atoi(strings.TrimSpace(setting.Value))
	if err != nil || userID <= 0 {
		return response.NewError(http.StatusNotFound, "客服账号暂未配置")
	}
	var account models.Users
	if err := u.DB.WithContext(c.Request.Context()).Where("id = ? AND status = ?", userID, 1).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewError(http.StatusNotFound, "客服账号不存在或已停用")
		}
		return response.NewError(http.StatusInternalServerError, "读取客服账号失败")
	}
	response.Success(c, types.CustomerServiceContact{
		UserID: userID, Nickname: account.Nickname, AvatarURL: account.Avatar, Signature: account.Motto,
	})
	return nil
}

func (u *User) GetUserNote(c *gin.Context) error {
	userId := c.GetInt("user_id")
	var req types.FeedRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		return response.NewError(http.StatusBadRequest, "参数错误")
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}
	notes, err := u.NoteService.ListNoteByUser(
		c.Request.Context(),
		req.Cursor,
		req.PageSize,
		userId,
		req.UserId,
	)
	avatar, nickName, err := u.UserService.GetUserAvatar(c.Request.Context(), int64(req.UserId))
	notes.Avatar = avatar
	notes.Nickname = nickName
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	response.Success(c, notes)
	return nil
}

func (u *User) GetUserInfo(c *gin.Context) error {
	ctx := c.Request.Context()

	// 登录态可选；没有 user_id 参数时仍然只能查看本人资料。
	loginUID, loginErr := context.GetUserID(c)

	//  确定要查询的用户ID（默认自己）
	queryID := int(loginUID)
	isQueryOther := false

	if userIDStr := c.Query("user_id"); userIDStr != "" {
		parsedID, err := strconv.Atoi(userIDStr)
		if err != nil {
			return response.NewError(http.StatusBadRequest, "user_id 非法")
		}
		queryID = parsedID
		isQueryOther = true
	} else if loginErr != nil || loginUID == 0 {
		return response.NewError(http.StatusUnauthorized, "请先登录后查看本人资料")
	}

	//  获取用户基础信息
	userInfo, err := u.UserService.GetUserInfo(ctx, queryID)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, "获取用户信息失败: "+err.Error())
	}

	//  获取统计数据（失败则降级为 0）
	following, _ := u.FollowService.GetFollowingCount(ctx, uint64(userInfo.Id))
	follower, _ := u.FollowService.GetFollowerCount(ctx, uint64(userInfo.Id))
	totalLikes, _ := u.LikeService.GetUserTotalLikes(ctx, uint64(userInfo.Id))
	totalCollects, _ := u.CollectService.GetUserTotalCollects(ctx, uint64(userInfo.Id))

	// 是否关注
	isFollowing := false
	if isQueryOther && loginUID > 0 {
		isFollowing, _ = u.FollowService.IsFollowing(
			ctx,
			uint64(loginUID),
			uint64(queryID),
		)
	}

	var verifierInfo *types.VerifierInfo
	if !isQueryOther && loginUID > 0 && u.DB != nil {
		var verifier models.Verifier
		if err := u.DB.WithContext(ctx).
			Where("user_id = ? AND status = ?", loginUID, models.VerifierStatusActive).
			Order("id desc").
			First(&verifier).Error; err == nil {
			info := &types.VerifierInfo{
				ID:          verifier.ID,
				Name:        verifier.Name,
				Phone:       verifier.Phone,
				Status:      verifier.Status,
				OrganizerID: verifier.OrganizerID,
			}
			var org models.Organizer
			if err := u.DB.WithContext(ctx).First(&org, verifier.OrganizerID).Error; err == nil {
				info.OrganizerName = org.Name
			}
			verifierInfo = info
		}
	}

	isSelf := loginUID > 0 && int64(queryID) == loginUID
	phoneNumber := ""
	if isSelf {
		phoneNumber = userInfo.Mobile
	}
	rep := types.UserProfileResp{
		User: types.UserBasicInfo{
			Id:          userInfo.Id,
			UserID:      utils.GenHashID(u.Config.Jwt.Secret, userInfo.Id),
			NumericID:   userInfo.Id,
			Nickname:    userInfo.Nickname,
			PhoneNumber: phoneNumber,
			AvatarURL:   userInfo.Avatar,
			Signature:   userInfo.Motto,
			CreatedAt:   userInfo.CreatedAt,
			IsVerifier:  verifierInfo != nil,
		},
		Stats: types.UserStats{
			Following: following,
			Follower:  follower,
			Likes:     totalLikes + totalCollects,
		},
		IsFollowing: isFollowing,
		IsVerifier:  verifierInfo != nil,
	}
	if isSelf {
		rep.Organizer = u.currentOrganizerIdentity(ctx, loginUID)
	}
	if verifierInfo != nil {
		rep.VerifierID = verifierInfo.ID
		rep.Verifier = verifierInfo
		rep.User.VerifierID = verifierInfo.ID
		rep.User.Verifier = verifierInfo
	}

	response.Success(c, rep)
	return nil
}

func (u *User) currentOrganizerIdentity(ctx base.Context, userID int64) *types.OrganizerIdentity {
	if userID <= 0 || u.DB == nil {
		return nil
	}
	var org models.Organizer
	err := u.DB.WithContext(ctx).
		Where("user_id = ? AND status = ? AND enabled = 1", userID, models.OrganizerStatusApproved).
		First(&org).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = u.DB.WithContext(ctx).
			Table("organizer_staff AS os").
			Select("o.*").
			Joins("JOIN organizers o ON o.id = os.organizer_id").
			Where("os.user_id = ? AND os.status = 1 AND o.status = ? AND o.enabled = 1", userID, models.OrganizerStatusApproved).
			Order("os.id DESC").
			Limit(1).
			Scan(&org).Error
	}
	if err != nil || org.ID == 0 {
		return nil
	}
	return &types.OrganizerIdentity{ID: org.ID, Type: org.Type, Name: org.Name, Logo: org.Logo}
}

func (u *User) UpdateUserInfo(c *gin.Context) error {
	userID, err := context.GetUserID(c) // 这里的 userID 是 int
	if err != nil {
		return response.NewError(http.StatusUnauthorized, "未登录")
	}

	var req types.UpdateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, "参数格式错误")
	}

	if err = u.UserService.Update(c.Request.Context(), int(userID), &req); err != nil {
		return response.NewError(http.StatusInternalServerError, "更新失败: "+err.Error())
	}
	response.Success(c, "更新成功")
	return nil
}

func (u *User) UploadAvatar(c *gin.Context) error {
	userID, _ := context.GetUserID(c)

	header, err := c.FormFile("image")
	if err != nil {
		return response.NewError(400, "请选择图片")
	}

	//  大小校验（10MB）
	if header.Size > 10<<20 {
		return response.NewError(400, "图片不能超过10MB")
	}

	file, err := header.Open()
	if err != nil {
		return response.NewError(400, "读取文件失败")
	}
	defer file.Close()

	buf := make([]byte, 512)
	_, _ = file.Read(buf)
	file.Seek(0, io.SeekStart)

	contentType := http.DetectContentType(buf)
	switch contentType {
	case "image/jpeg", "image/png", "image/webp":
	default:
		return response.NewError(400, "不支持的图片格式")
	}

	objectKey := fmt.Sprintf(
		"avatars/%02d/%d/%s%s",
		userID%100,
		userID,
		uuid.NewString()[:8],
		path.Ext(header.Filename),
	)

	if err := u.OssService.UploadReader(c.Request.Context(), file, objectKey); err != nil {
		return response.NewError(500, "上传云端失败")
	}
	fullUrl := fmt.Sprintf("https://cdn.hypercn.cn/%s", objectKey)
	response.Success(c, types.UploadAvatarRes{Url: fullUrl})
	return nil
}

// GetMyNotes 获取我的笔记
func (u *User) GetMyNotes(c *gin.Context) error {
	userID := c.GetInt("user_id")
	var req types.FeedRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		return response.NewError(http.StatusBadRequest, "参数错误")
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}
	notes, nextCursor, hasMore, err := u.NoteService.GetMyNotesFeed(
		c.Request.Context(),
		userID,
		req.Cursor,
		req.PageSize,
	)
	respNotes := make([]*types.Note, len(notes))
	for i, note := range notes {
		respNotes[i] = &types.Note{
			ID:          int64(note.ID),
			UserID:      int64(note.UserID),
			Title:       note.Title,
			Content:     note.Content,
			TopicIDs:    make([]int64, 0),
			Location:    types.Location{},
			MediaData:   make([]types.NoteMedia, 0),
			Type:        note.Type,
			Status:      note.Status,
			VisibleConf: note.VisibleConf,
			CreatedAt:   note.CreatedAt,
			UpdatedAt:   note.UpdatedAt,
		}
		_ = json.Unmarshal([]byte(note.TopicIDs), &respNotes[i].TopicIDs)
		_ = json.Unmarshal([]byte(note.Location), &respNotes[i].Location)
		_ = json.Unmarshal([]byte(note.MediaData), &respNotes[i].MediaData)
	}
	u.NoteService.EnrichNoteCards(c.Request.Context(), respNotes, uint64(userID))

	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	response.Success(c, types.FeedResponse{
		List:       respNotes,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	})
	return nil
}
