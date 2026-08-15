package handler

import (
	"Hyper/config"
	"Hyper/middleware"
	"Hyper/models"
	"Hyper/pkg/context"
	"Hyper/pkg/log"
	"Hyper/pkg/response"
	"Hyper/service"
	"Hyper/types"
	base "context"
	"encoding/json"
	"errors"
	"fmt"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"strconv"

	rmq_client "github.com/apache/rocketmq-clients/golang/v5"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	_ "golang.org/x/image/webp"
	"gorm.io/gorm"
)

type Note struct {
	OssService     service.IOssService
	NoteService    service.INoteService
	LikeService    service.ILikeService
	CollectService service.ICollectService
	TopicService   service.ITopicService
	Config         *config.Config
	Producer       rmq_client.Producer
	Db             *gorm.DB
}

func (n *Note) RegisterRouter(r gin.IRouter) {
	authorize := middleware.Auth([]byte(n.Config.Jwt.Secret))
	optionalAuth := middleware.OptionalAuth([]byte(n.Config.Jwt.Secret))
	g := r.Group("/v1/note")

	g.GET("/gen", authorize, context.Wrap(n.Gen))
	g.POST("/upload", authorize, context.Wrap(n.UploadImage))
	g.GET("/upload/:image_id/tags", authorize, context.Wrap(n.GetUploadImageTags))
	g.POST("/create", authorize, context.Wrap(n.CreateNote))
	g.GET("/my", authorize, context.Wrap(n.GetMyNotes))
	g.GET("/my/likes", authorize, context.Wrap(n.GetMyLikes))
	g.GET("/my/collects", authorize, context.Wrap(n.GetMyCollections))

	g.GET("/list", optionalAuth, context.Wrap(n.ListNote))
	g.GET("/related", optionalAuth, context.Wrap(n.ListRelatedNotes))
	g.GET("/followed", authorize, context.Wrap(n.ListFollowedNotes))
	// Like APIs
	g.POST("/:note_id/like", authorize, context.Wrap(n.Like))
	g.DELETE("/:note_id/like", authorize, context.Wrap(n.Unlike))
	g.GET("/:note_id/like", authorize, context.Wrap(n.GetLikeStatus))
	g.GET("/:note_id/likes/count", context.Wrap(n.GetLikeCount))
	// Collection APIs
	g.POST("/:note_id/collect", authorize, context.Wrap(n.Collect))
	g.DELETE("/:note_id/collect", authorize, context.Wrap(n.Uncollect))
	g.GET("/:note_id/collect", authorize, context.Wrap(n.GetCollectStatus))
	g.GET("/:note_id/collections/count", context.Wrap(n.GetCollectCount))
	g.POST("/:note_id/share", authorize, context.Wrap(n.RecordShare))
	g.DELETE("/:note_id", authorize, context.Wrap(n.DeleteNote))
	g.PATCH("/:note_id/relation", authorize, context.Wrap(n.UpdateNoteRelation))
	g.PUT("/:note_id/relation", authorize, context.Wrap(n.UpdateNoteRelation))
	g.GET("/:note_id", optionalAuth, context.Wrap(n.GetNoteDetail))
}

func (n *Note) RecordShare(c *gin.Context) error {
	userID, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(http.StatusUnauthorized, "未登录")
	}
	noteID, err := strconv.ParseUint(c.Param("note_id"), 10, 64)
	if err != nil || noteID == 0 {
		return response.NewError(http.StatusBadRequest, "笔记ID格式错误")
	}
	var req types.RecordNoteShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, "参数格式错误: "+err.Error())
	}
	if err := n.NoteService.RecordShare(c.Request.Context(), uint64(userID), noteID, req.Channel); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewError(http.StatusNotFound, "动态不存在")
		}
		return response.NewError(http.StatusBadRequest, err.Error())
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

// Gen 兜底：把积压的未分类笔记（channel_id = 0）逐条发到 MQ，由 conn-server 消费分类。
// 供定时任务/手动触发（/v1/note/gen），不再在发帖时触发。
func (n *Note) Gen(c *gin.Context) error {
	go func() {
		ctx := base.Background()
		var notes []models.Note
		_ = n.Db.WithContext(ctx).Where("channel_id = ?", 0).FindInBatches(&notes, 100, func(tx *gorm.DB, batch int) error {
			for _, v := range notes {
				n.publishNoteClassify(v.ID)
			}
			return nil
		})
	}()
	return nil
}

// publishNoteClassify 发送笔记分类消息到 RocketMQ。
func (n *Note) publishNoteClassify(noteID uint64) {
	if n.Producer == nil {
		log.L.Warn("note classify producer is nil")
		return
	}
	body, _ := json.Marshal(&types.NoteClassifyMessage{NoteID: noteID})
	msg := &rmq_client.Message{Topic: types.NoteClassifyTopic, Body: body}
	if _, err := n.Producer.Send(base.Background(), msg); err != nil {
		log.L.Error("send note classify msg error", zap.Error(err))
	}
}

// publishNoteImageTag 将耗时的视觉标签生成移到 MQ 消费端，避免阻塞上传接口。
func (n *Note) publishNoteImageTag(imageID int64, userID uint64, url string) error {
	if n.Producer == nil {
		return errors.New("note image tag producer is nil")
	}
	body, err := json.Marshal(&types.NoteImageTagMessage{
		ImageID: imageID,
		UserID:  userID,
		URL:     url,
	})
	if err != nil {
		return err
	}
	_, err = n.Producer.Send(base.Background(), &rmq_client.Message{
		Topic: types.NoteImageTagTopic,
		Body:  body,
	})
	if err == nil {
		log.L.Info("note image tag task enqueued", zap.Int64("image_id", imageID), zap.Uint64("user_id", userID))
	}
	return err
}

// CreateNote 创建笔记
func (n *Note) CreateNote(c *gin.Context) error {
	//从 context 获取用户 ID
	userID, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	//userID := uint64(1)
	// 绑定请求参数
	var req types.CreateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, "参数格式错误: "+err.Error())
	}

	// 调用 MessageService 层创建笔记
	noteID, err := n.NoteService.CreateNote(c.Request.Context(), uint64(userID), &req)
	if err != nil {
		if errors.Is(err, service.ErrContentUnsafe) {
			return response.NewError(http.StatusBadRequest, "内容含违规信息")
		}
		if errors.Is(err, service.ErrContentSafetyUnavailable) {
			return response.NewError(http.StatusServiceUnavailable, "内容安全验证失败，请稍后重试")
		}
		return response.NewError(http.StatusInternalServerError, "创建笔记失败: "+err.Error())
	}
	n.publishNoteClassify(noteID)
	// 返回成功响应
	response.Success(c, types.CreateNoteResponse{
		NoteID: noteID,
	})
	return nil
}
func (n *Note) GetNoteDetail(c *gin.Context) error {
	// 获取笔记ID
	noteIDStr := c.Param("note_id")
	noteID, err := strconv.ParseUint(noteIDStr, 10, 64)
	if err != nil {
		return response.NewError(http.StatusBadRequest, "笔记ID格式错误")
	}

	// 获取当前用户ID (可选,未登录为0)
	currentUserID, _ := context.GetUserID(c)

	// 调用 Service 获取详情
	detail, err := n.NoteService.GetNoteDetail(c.Request.Context(), noteID, uint64(currentUserID))
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}

	// 返回成功响应
	response.Success(c, detail)
	return nil
}

func (n *Note) UpdateNoteRelation(c *gin.Context) error {
	userID := c.GetInt("user_id")
	if userID == 0 {
		return response.NewError(http.StatusUnauthorized, "未登录")
	}
	noteID, err := strconv.ParseUint(c.Param("note_id"), 10, 64)
	if err != nil || noteID == 0 {
		return response.NewError(http.StatusBadRequest, "笔记ID格式错误")
	}
	var req types.UpdateNoteRelationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, "参数格式错误: "+err.Error())
	}
	if err := n.NoteService.UpdateNoteRelation(c.Request.Context(), uint64(userID), noteID, req); err != nil {
		return response.NewError(http.StatusBadRequest, err.Error())
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (n *Note) UploadImage(c *gin.Context) error {
	userID, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	header, err := c.FormFile("image")
	if err != nil {
		return response.NewError(400, "missing image")
	}
	img, err := n.OssService.UploadImage(c.Request.Context(), int(userID), header)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	if err := n.publishNoteImageTag(img.ImageID, uint64(userID), img.Url); err != nil {
		log.L.Error("send note image tag msg error", zap.Error(err), zap.Int64("image_id", img.ImageID))
		if n.Db != nil {
			_ = n.Db.WithContext(c.Request.Context()).Model(&models.Image{}).
				Where("id = ? AND user_id = ?", img.ImageID, userID).
				Updates(map[string]interface{}{
					"tag_status": types.ImageTagStatusFailed,
					"tag_error":  "标签任务投递失败",
				}).Error
		}
		img.TagStatus = "failed"
	}

	response.Success(c, img)
	return nil
}

// GetUploadImageTags 为 Socket 离线或丢包场景提供标签结果兜底查询。
func (n *Note) GetUploadImageTags(c *gin.Context) error {
	userID, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(http.StatusUnauthorized, "未登录")
	}
	imageID, err := strconv.ParseInt(c.Param("image_id"), 10, 64)
	if err != nil || imageID <= 0 {
		return response.NewError(http.StatusBadRequest, "图片ID格式错误")
	}

	var image models.Image
	if err := n.Db.WithContext(c.Request.Context()).
		Where("id = ? AND user_id = ?", imageID, userID).
		First(&image).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewError(http.StatusNotFound, "图片不存在")
		}
		return response.NewError(http.StatusInternalServerError, "查询图片标签失败")
	}

	tags := make([]types.CreateOrGetTopicResponse, 0)
	if image.Tags != "" {
		if err := json.Unmarshal([]byte(image.Tags), &tags); err != nil {
			log.L.Warn("unmarshal image tags error", zap.Error(err), zap.Int64("image_id", imageID))
			tags = make([]types.CreateOrGetTopicResponse, 0)
		}
	}
	response.Success(c, types.NoteImageTagResult{
		ImageID: image.ID,
		URL:     "https://cdn.hypercn.cn/" + image.OssKey,
		Status:  imageTagStatusName(image.TagStatus),
		Tags:    tags,
		Error:   imageTagError(image.TagStatus),
	})
	return nil
}

func imageTagStatusName(status int) string {
	switch status {
	case types.ImageTagStatusCompleted:
		return "completed"
	case types.ImageTagStatusFailed:
		return "failed"
	default:
		return "pending"
	}
}

func imageTagError(status int) string {
	if status == types.ImageTagStatusFailed {
		return "标签生成失败"
	}
	return ""
}

func (n *Note) ListNote(c *gin.Context) error {
	userId := c.GetInt("user_id")

	var req types.ListNotesReq
	if err := c.ShouldBindQuery(&req); err != nil {
		return response.NewError(http.StatusBadRequest, "参数错误: "+err.Error())
	}

	if req.PageSize == 0 {
		req.PageSize = types.DefaultPageSize
	}
	var resp types.ListNotesRep
	var err error
	if req.SearchType == "follow" {
		resp, err = n.NoteService.GetFollowedPosts(c.Request.Context(), userId, req.Cursor, req.PageSize)
		if err != nil {
			return response.NewError(http.StatusInternalServerError, "获取笔记失败: "+err.Error())
		}
		response.Success(c, resp)
		return nil
	}

	if req.ChannelID > 0 {
		resp, err = n.NoteService.GetNoteByChannelID(c.Request.Context(), userId, req.Cursor, req.PageSize, int(req.ChannelID))
		if err != nil {
			return response.NewError(http.StatusInternalServerError, "获取笔记失败: "+err.Error())
		}
	} else {
		resp, err = n.NoteService.ListNote(c.Request.Context(), req.Cursor, req.PageSize, uint64(userId))
		if err != nil {
			return response.NewError(http.StatusInternalServerError, "获取笔记失败: "+err.Error())
		}
	}

	response.Success(c, resp)
	return nil
}

func (n *Note) ListRelatedNotes(c *gin.Context) error {
	userID := c.GetInt("user_id")
	var req types.ListRelatedNotesReq
	if err := c.ShouldBindQuery(&req); err != nil {
		return response.NewError(http.StatusBadRequest, "参数错误: "+err.Error())
	}
	resp, err := n.NoteService.GetRelatedNotes(c.Request.Context(), req, uint64(userID))
	if err != nil {
		return response.NewError(http.StatusInternalServerError, "获取相关动态失败: "+err.Error())
	}
	response.Success(c, resp)
	return nil
}

func (n *Note) ListFollowedNotes(c *gin.Context) error {
	userID := c.GetInt("user_id")
	var req types.ListNotesReq
	if err := c.ShouldBindQuery(&req); err != nil {
		return response.NewError(http.StatusBadRequest, "参数错误: "+err.Error())
	}

	if req.PageSize == 0 {
		req.PageSize = types.DefaultPageSize
	}

	rep, err := n.NoteService.GetFollowedPosts(c.Request.Context(), userID, req.Cursor, req.PageSize)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, "获取笔记失败: "+err.Error())
	}

	response.Success(c, rep)
	return nil
}

// GetMyNotes 查询自己的笔记列表
func (n *Note) GetMyNotes(c *gin.Context) error {
	// 1. 获取当前登录用户ID
	userID, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(http.StatusUnauthorized, "未登录")
	}
	// fmt.Printf("[GetMyNotes] 查询用户ID: %d\n", userID)

	// 2. 绑定查询参数
	var req types.GetMyNotesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		return response.NewError(http.StatusBadRequest, "参数错误: "+err.Error())
	}

	// 3. 设置默认值
	if req.Page == 0 {
		req.Page = types.DefaultPage
	}
	if req.PageSize == 0 {
		req.PageSize = types.DefaultPageSize
	}
	// 仅当未提供 status 参数时，默认查询公开状态
	if c.Query("status") == "" {
		req.Status = types.NoteStatusDefaultQuery
	}
	// 计算 limit 和 offset
	limit := req.PageSize
	offset := (req.Page - 1) * req.PageSize
	// fmt.Printf("[GetMyNotes] 查询参数 - Status: %d, Page: %d, PageSize: %d, Offset: %d\n", req.Status, req.Page, req.PageSize, offset)

	// 4. 调用 MessageService 层查询
	notes, err := n.NoteService.GetUserNotes(c.Request.Context(), uint64(userID), req.Status, limit, offset)
	if err != nil {
		// fmt.Printf("[GetMyNotes] 查询错误: %v\n", err)
		return response.NewError(http.StatusInternalServerError, "查询失败: "+err.Error())
	}
	// fmt.Printf("[GetMyNotes] 查询结果数量: %d\n", len(notes))

	// 5. 返回成功响应
	total := 0
	if notes != nil {
		total = len(notes)
	}

	res := make([]*types.Note, 0)
	for _, note := range notes {

		k := &types.Note{
			ID:          int64(note.ID),
			UserID:      int64(note.UserID),
			Title:       note.Title,
			Content:     note.Content,
			Type:        note.Type,
			Status:      note.Status,
			VisibleConf: note.VisibleConf,
			CreatedAt:   note.CreatedAt,
			UpdatedAt:   note.UpdatedAt,
		}
		_ = json.Unmarshal([]byte(note.TopicIDs), &k.TopicIDs)
		_ = json.Unmarshal([]byte(note.Location), &k.Location)
		_ = json.Unmarshal([]byte(note.MediaData), &k.MediaData)
		res = append(res, k)
	}

	response.Success(c, types.GetMyNotesResponse{
		Notes: res,
		Total: total,
	})
	return nil
}

// GetMyCollections 查询自己的收藏列表
func (n *Note) GetMyCollections(c *gin.Context) error {
	userID, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(http.StatusUnauthorized, "未登录")
	}

	var req types.GetMyCollectionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		return response.NewError(http.StatusBadRequest, "参数错误: "+err.Error())
	}

	if req.Page == 0 {
		req.Page = types.DefaultPage
	}
	if req.PageSize == 0 {
		req.PageSize = types.DefaultPageSize
	}

	limit := req.PageSize
	offset := (req.Page - 1) * req.PageSize

	notes, total, err := n.CollectService.GetUserCollections(c.Request.Context(), uint64(userID), limit, offset)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, "查询失败: "+err.Error())
	}

	response.Success(c, types.GetMyCollectionsResponse{
		Notes: notes,
		Total: int(total),
	})
	return nil
}

// GetMyLikes queries the current user's own like history. It must not be
// confused with the profile's received-like statistic.
func (n *Note) GetMyLikes(c *gin.Context) error {
	userID, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(http.StatusUnauthorized, "未登录")
	}
	var req types.GetMyLikesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		return response.NewError(http.StatusBadRequest, "参数错误: "+err.Error())
	}
	if req.Page == 0 {
		req.Page = types.DefaultPage
	}
	if req.PageSize == 0 {
		req.PageSize = types.DefaultPageSize
	}
	notes, total, err := n.LikeService.GetUserLikes(c.Request.Context(), uint64(userID), req.PageSize, (req.Page-1)*req.PageSize)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, "查询失败: "+err.Error())
	}
	response.Success(c, types.GetMyLikesResponse{Notes: notes, Total: int(total)})
	return nil
}

func (n *Note) DeleteNote(c *gin.Context) error {
	userID := c.GetInt("user_id")
	if userID == 0 {
		return response.NewError(http.StatusUnauthorized, "未登录")
	}
	noteID, err := strconv.ParseUint(c.Param("note_id"), 10, 64)
	if err != nil || noteID == 0 {
		return response.NewError(http.StatusBadRequest, "笔记ID格式错误")
	}
	var note models.Note
	if err := n.Db.WithContext(c.Request.Context()).Where("id = ?", noteID).First(&note).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.NewError(http.StatusNotFound, "动态不存在")
		}
		return response.NewError(http.StatusInternalServerError, "查询动态失败: "+err.Error())
	}
	if note.UserID != uint64(userID) {
		return response.NewError(http.StatusForbidden, "只能删除自己的动态")
	}
	if note.Status == -1 {
		return response.NewError(http.StatusBadRequest, "动态已删除")
	}
	if err := n.Db.WithContext(c.Request.Context()).
		Model(&models.Note{}).
		Where("id = ? AND user_id = ?", noteID, userID).
		Updates(map[string]any{"status": -1}).Error; err != nil {
		return response.NewError(http.StatusInternalServerError, "删除动态失败: "+err.Error())
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "删除成功", "data": gin.H{"success": true}})
	return nil
}

// Like 点赞笔记
func (n *Note) Like(c *gin.Context) error {
	userID, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(http.StatusUnauthorized, "未登录")
	}
	noteIDParam := c.Param("note_id")
	if noteIDParam == "" {
		return response.NewError(http.StatusBadRequest, "缺少 note_id")
	}
	var noteID uint64
	_, err = fmt.Sscanf(noteIDParam, "%d", &noteID)
	if err != nil {
		return response.NewError(http.StatusBadRequest, "note_id 格式错误")
	}

	err = n.LikeService.LikeNote(c.Request.Context(), uint64(userID), noteID)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	response.Success(c, gin.H{"liked": true})
	return nil
}

// Unlike 取消点赞
func (n *Note) Unlike(c *gin.Context) error {
	userID, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(http.StatusUnauthorized, "未登录")
	}
	noteIDParam := c.Param("note_id")
	if noteIDParam == "" {
		return response.NewError(http.StatusBadRequest, "缺少 note_id")
	}
	var noteID uint64
	_, err = fmt.Sscanf(noteIDParam, "%d", &noteID)
	if err != nil {
		return response.NewError(http.StatusBadRequest, "note_id 格式错误")
	}

	err = n.LikeService.UnlikeNote(c.Request.Context(), uint64(userID), noteID)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	response.Success(c, gin.H{"liked": false})
	return nil
}

// GetLikeStatus 查询是否已点赞
func (n *Note) GetLikeStatus(c *gin.Context) error {
	userID, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(http.StatusUnauthorized, "未登录")
	}
	noteIDParam := c.Param("note_id")
	if noteIDParam == "" {
		return response.NewError(http.StatusBadRequest, "缺少 note_id")
	}
	var noteID uint64
	_, err = fmt.Sscanf(noteIDParam, "%d", &noteID)
	if err != nil {
		return response.NewError(http.StatusBadRequest, "note_id 格式错误")
	}

	liked, err := n.LikeService.IsLiked(c.Request.Context(), uint64(userID), noteID)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	response.Success(c, gin.H{"liked": liked})
	return nil
}

// GetLikeCount 查询点赞数量
func (n *Note) GetLikeCount(c *gin.Context) error {
	noteIDParam := c.Param("note_id")
	if noteIDParam == "" {
		return response.NewError(http.StatusBadRequest, "缺少 note_id")
	}
	var noteID uint64
	_, err := fmt.Sscanf(noteIDParam, "%d", &noteID)
	if err != nil {
		return response.NewError(http.StatusBadRequest, "note_id 格式错误")
	}

	cnt, err := n.LikeService.GetLikeCount(c.Request.Context(), noteID)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	response.Success(c, gin.H{"like_count": cnt})
	return nil
}

// Collect 收藏笔记
func (n *Note) Collect(c *gin.Context) error {
	userID, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(http.StatusUnauthorized, "未登录")
	}
	noteIDParam := c.Param("note_id")
	if noteIDParam == "" {
		return response.NewError(http.StatusBadRequest, "缺少 note_id")
	}
	var noteID uint64
	_, err = fmt.Sscanf(noteIDParam, "%d", &noteID)
	if err != nil {
		return response.NewError(http.StatusBadRequest, "note_id 格式错误")
	}

	err = n.CollectService.Collect(c.Request.Context(), uint64(userID), noteID)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	response.Success(c, gin.H{"collected": true})
	return nil
}

// Uncollect 取消收藏
func (n *Note) Uncollect(c *gin.Context) error {
	userID, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(http.StatusUnauthorized, "未登录")
	}
	noteIDParam := c.Param("note_id")
	if noteIDParam == "" {
		return response.NewError(http.StatusBadRequest, "缺少 note_id")
	}
	var noteID uint64
	_, err = fmt.Sscanf(noteIDParam, "%d", &noteID)
	if err != nil {
		return response.NewError(http.StatusBadRequest, "note_id 格式错误")
	}

	err = n.CollectService.Uncollect(c.Request.Context(), uint64(userID), noteID)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	response.Success(c, gin.H{"collected": false})
	return nil
}

// GetCollectStatus 查询是否已收藏
func (n *Note) GetCollectStatus(c *gin.Context) error {
	userID, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(http.StatusUnauthorized, "未登录")
	}
	noteIDParam := c.Param("note_id")
	if noteIDParam == "" {
		return response.NewError(http.StatusBadRequest, "缺少 note_id")
	}
	var noteID uint64
	_, err = fmt.Sscanf(noteIDParam, "%d", &noteID)
	if err != nil {
		return response.NewError(http.StatusBadRequest, "note_id 格式错误")
	}

	collected, err := n.CollectService.IsCollected(c.Request.Context(), uint64(userID), noteID)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	response.Success(c, gin.H{"collected": collected})
	return nil
}

// GetCollectCount 查询收藏数量
func (n *Note) GetCollectCount(c *gin.Context) error {
	noteIDParam := c.Param("note_id")
	if noteIDParam == "" {
		return response.NewError(http.StatusBadRequest, "缺少 note_id")
	}
	var noteID uint64
	_, err := fmt.Sscanf(noteIDParam, "%d", &noteID)
	if err != nil {
		return response.NewError(http.StatusBadRequest, "note_id 格式错误")
	}

	cnt, err := n.CollectService.GetCollectionCount(c.Request.Context(), noteID)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	response.Success(c, gin.H{"collect_count": cnt})
	return nil
}
