package handler

import (
	"Hyper/config"
	"Hyper/middleware"
	"Hyper/pkg/context"
	"Hyper/pkg/response"
	"Hyper/pkg/utils"
	"Hyper/service"
	"Hyper/types"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type User struct {
	Config         *config.Config
	UserService    service.IUserService
	OssService     service.IOssService
	FollowService  service.FollowService
	LikeService    service.LikeService
	CollectService service.CollectService
	NoteService    service.INoteService
}

func (u *User) RegisterRouter(r gin.IRouter) {
	authorize := middleware.Auth([]byte(u.Config.Jwt.Secret))
	g := r.Group("/v1/user")
	g.Use(authorize)
	g.POST("/info", context.Wrap(u.UpdateUserInfo))
	g.POST("/avatar", context.Wrap(u.UploadAvatar))
	g.GET("/info", context.Wrap(u.GetUserInfo))
	g.GET("/note", context.Wrap(u.GetUserNote))
	g.GET("/my-notes", context.Wrap(u.GetMyNotes))

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

	// 1. 获取当前登录用户ID
	loginUID, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(http.StatusUnauthorized, "获取用户身份失败")
	}

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
	if isQueryOther {
		isFollowing, _ = u.FollowService.IsFollowing(
			ctx,
			uint64(loginUID),
			uint64(queryID),
		)
	}

	rep := types.UserProfileResp{
		User: types.UserBasicInfo{
			Id:          userInfo.Id,
			UserID:      utils.GenHashID(u.Config.Jwt.Secret, userInfo.Id),
			Nickname:    userInfo.Nickname,
			PhoneNumber: userInfo.Mobile,
			AvatarURL:   userInfo.Avatar,
			CreatedAt:   userInfo.CreatedAt,
		},
		Stats: types.UserStats{
			Following: following,
			Follower:  follower,
			Likes:     totalLikes + totalCollects,
		},
		IsFollowing: isFollowing,
	}

	response.Success(c, rep)
	return nil
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
