package handler

import (
	"Hyper/config"
	"Hyper/middleware"
	"Hyper/models"
	"Hyper/pkg/context"
	"Hyper/pkg/llm"
	"Hyper/pkg/response"
	"Hyper/service"
	"Hyper/types"
	base "context"
	"encoding/json"
	"fmt"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "golang.org/x/image/webp"
	"gorm.io/gorm"
)

type Note struct {
	OssService     service.IOssService
	NoteService    service.INoteService
	LikeService    service.ILikeService
	CollectService service.ICollectService
	Config         *config.Config
	Channel        service.IChannelService
	Db             *gorm.DB
}

func (n *Note) RegisterRouter(r gin.IRouter) {
	authorize := middleware.Auth([]byte(n.Config.Jwt.Secret))
	g := r.Group("/v1/note")

	g.GET("/gen", authorize, context.Wrap(n.Gen))
	g.POST("/upload", authorize, context.Wrap(n.UploadImage))
	g.POST("/create", authorize, context.Wrap(n.CreateNote))
	g.GET("/my", authorize, context.Wrap(n.GetMyNotes))
	g.GET("/my/collects", authorize, context.Wrap(n.GetMyCollections))

	g.GET("/list", authorize, context.Wrap(n.ListNote))
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
	g.GET("/:note_id", authorize, context.Wrap(n.GetNoteDetail))
}

func (n *Note) Gen(c *gin.Context) error {
	// 1. 预先获取频道映射，避免在循环里调接口
	tags, err := n.Channel.ListChannels(c.Request.Context(), &types.ListChannelsReq{})
	if err != nil {
		return err
	}
	tagsSlice := make([]string, 0)
	tagsMap := make(map[string]int) // ID 建议用 uint32，与数据库对应
	for _, v := range tags.Channels {
		tagsSlice = append(tagsSlice, v.Name)
		tagsMap[v.Name] = v.Id
	}

	// 2. 异步处理，不阻塞接口返回
	go func() {
		// 使用一个独立的 Context，不要用 Gin 的 c.Request.Context()，因为请求结束它会 cancel
		ctx := base.Background()

		// 3. 分批处理（例如每次取 100 条），只取未分类的 (channel_id = 0)
		var notes []models.Note
		n.Db.WithContext(ctx).Where("channel_id = ?", 0).FindInBatches(&notes, 100, func(tx *gorm.DB, batch int) error {
			for _, v := range notes {
				// 4. 解析媒体数据
				var noteMedia []types.NoteMedia
				_ = json.Unmarshal([]byte(v.MediaData), &noteMedia)
				urlImages := make([]string, 0)
				for _, m := range noteMedia {
					urlImages = append(urlImages, m.URL)
				}

				// 5. 调用大模型
				label := llm.ClassifyMultiImageNote(ctx, v.Title, v.Content, urlImages, tagsSlice)

				// 6. 安全匹配 ID
				if labelId, ok := tagsMap[label]; ok {
					err := n.Db.Model(&models.Note{}).Where("id = ?", v.ID).Update("channel_id", labelId).Error
					if err != nil {
						log.Printf("更新笔记 %d 失败: %v", v.ID, err)
					}
				}
			}
			return nil
		})
	}()

	// 接口立即返回
	c.JSON(200, gin.H{"msg": "已开始异步分类任务"})
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
	response.Success(c, img)
	return nil
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
		return response.NewError(http.StatusInternalServerError, "创建笔记失败: "+err.Error())
	}

	// 返回成功响应
	response.Success(c, types.CreateNoteResponse{
		NoteID: noteID,
	})
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
