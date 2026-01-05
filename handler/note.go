package handler

import (
	"Hyper/config"
	"Hyper/middleware"
	"Hyper/pkg/context"
	"Hyper/pkg/response"
	"Hyper/service"
	"Hyper/types"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "golang.org/x/image/webp"
)

type Note struct {
	OssService  service.IOssService
	NoteService service.INoteService
	LikeService service.ILikeService
	Config      *config.Config
}

func (n *Note) RegisterRouter(r gin.IRouter) {
	authorize := middleware.Auth([]byte(n.Config.Jwt.Secret))
	g := r.Group("/v1/note")
	g.POST("/upload", context.Wrap(n.UploadImage))
	g.POST("/create", authorize, context.Wrap(n.CreateNote))
	g.GET("/my", authorize, context.Wrap(n.GetMyNotes))
	// Like APIs
	g.POST("/:note_id/like", authorize, context.Wrap(n.Like))
	g.DELETE("/:note_id/like", authorize, context.Wrap(n.Unlike))
	g.GET("/:note_id/like", authorize, context.Wrap(n.GetLikeStatus))
	g.GET("/:note_id/likes/count", context.Wrap(n.GetLikeCount))
}
func (n *Note) UploadImage(c *gin.Context) error {
	header, err := c.FormFile("image")
	if err != nil {
		return response.NewError(500, err.Error())
	}
	const maxSize = 10 << 20
	if header.Size > maxSize {
		return fmt.Errorf("image size exceeds 10MB")
	}

	file, err := header.Open()
	if err != nil {
		return response.NewError(500, err.Error())
	}
	defer file.Close()
	var (
		width  int
		height int
	)

	if seeker, ok := file.(io.Seeker); ok {
		cfg, format, err := image.DecodeConfig(file)
		if err == nil {
			width = cfg.Width
			height = cfg.Height
		}
		allowed := map[string]bool{
			"jpeg": true,
			"png":  true,
			"webp": true,
		}
		if !allowed[strings.ToLower(format)] {
			return fmt.Errorf("unsupported image format")
		}
		seeker.Seek(0, 0) // 重置指针
	}

	objectKey := fmt.Sprintf("note/%s/%s%s", time.Now().Format("2006/01/02"), uuid.NewString(),
		path.Ext(header.Filename),
	)

	if err := n.OssService.UploadReader(c.Request.Context(), file, objectKey); err != nil {
		return response.NewError(500, err.Error())
	}
	response.Success(c, types.UploadResponse{
		Key: fmt.Sprintf("https://%s.%s/%s",
			n.Config.Oss.Bucket, n.Config.Oss.Endpoint,
			objectKey),
		Width:  width,
		Height: height,
	})
	return nil
}

// CreateNote 创建笔记
func (n *Note) CreateNote(c *gin.Context) error {
	// 从 context 获取用户 ID
	userID, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}

	// 绑定请求参数
	var req types.CreateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, "参数格式错误: "+err.Error())
	}

	// 调用 Service 层创建笔记
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

	// 4. 调用 Service 层查询
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
	response.Success(c, types.GetMyNotesResponse{
		Notes: notes,
		Total: total,
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

	err = n.LikeService.Like(c.Request.Context(), uint64(userID), noteID)
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

	err = n.LikeService.Unlike(c.Request.Context(), uint64(userID), noteID)
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
