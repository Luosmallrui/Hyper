package handler

import (
	"Hyper/config"
	"Hyper/middleware"
	"Hyper/pkg/context"
	"Hyper/pkg/response"
	"Hyper/pkg/snowflake"
	"Hyper/service"
	"Hyper/types"
	"fmt"
	"image"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Event struct {
	Config       *config.Config
	EventService service.IEventService
	OssService   service.IOssService
}

func (e *Event) RegisterRouter(r gin.IRouter) {
	authorize := middleware.Auth([]byte(e.Config.Jwt.Secret))
	g := r.Group("/v1/event")
	g.Use(authorize)
	{
		g.POST("/create", context.Wrap(e.CreateEvent))
		g.POST("/upload", context.Wrap(e.UploadPoster))
		g.GET("/list", context.Wrap(e.GetEventList))
		g.GET("/my", context.Wrap(e.GetMyEvents))
		g.GET("/:id", context.Wrap(e.GetEventDetail))
	}
}

// UploadPoster 上传活动海报
// POST /api/v1/event/upload  form-data: image=<file>
func (e *Event) UploadPoster(c *gin.Context) error {
	header, err := c.FormFile("image")
	if err != nil {
		return response.NewError(400, "missing image")
	}

	const maxSize int64 = 2 << 20 // 2MB
	if header.Size <= 0 || header.Size > maxSize {
		return response.NewError(400, "图片大小不能超过 2MB")
	}

	f, err := header.Open()
	if err != nil {
		return response.NewError(500, "打开文件失败")
	}
	defer f.Close()

	seeker, ok := f.(io.ReadSeeker)
	if !ok {
		return response.NewError(500, "文件不可读")
	}

	head := make([]byte, 512)
	n, _ := seeker.Read(head)
	contentType := http.DetectContentType(head[:n])
	if !strings.HasPrefix(contentType, "image/") {
		return response.NewError(400, "仅支持图片格式")
	}
	_, _ = seeker.Seek(0, io.SeekStart)

	cfg, format, err := image.DecodeConfig(seeker)
	if err != nil {
		return response.NewError(400, "无效的图片")
	}
	_, _ = seeker.Seek(0, io.SeekStart)

	ext := "." + strings.ToLower(format)
	if ext == ".jpeg" {
		ext = ".jpg"
	}
	objectKey := fmt.Sprintf("event/%s/%d%s",
		time.Now().Format("2006/01/02"),
		snowflake.GenID(),
		ext,
	)

	if err := e.OssService.UploadRaw(c.Request.Context(), seeker, objectKey); err != nil {
		return response.NewError(500, "上传失败: "+err.Error())
	}

	url := "https://cdn.hypercn.cn/" + objectKey
	response.Success(c, gin.H{
		"url":    url,
		"width":  cfg.Width,
		"height": cfg.Height,
	})
	return nil
}

// CreateEvent 创建活动
// POST /api/v1/event/create
func (e *Event) CreateEvent(c *gin.Context) error {
	var req types.CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, "参数格式错误: "+err.Error())
	}

	userID := c.GetInt("user_id")
	if userID == 0 {
		return response.NewError(http.StatusUnauthorized, "请先登录")
	}

	eventID, err := e.EventService.CreateEvent(c.Request.Context(), userID, req)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, "创建活动失败: "+err.Error())
	}

	response.Success(c, gin.H{"id": eventID})
	return nil
}

// GetEventList 获取活动列表
// GET /api/v1/event/list
func (e *Event) GetEventList(c *gin.Context) error {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	list, total, err := e.EventService.GetEventList(c.Request.Context(), page, pageSize)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, "查询失败")
	}

	response.Success(c, types.EventListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
	return nil
}

// GetMyEvents 获取我的活动列表
// GET /api/v1/event/my
func (e *Event) GetMyEvents(c *gin.Context) error {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	userID := c.GetInt("user_id")

	list, total, err := e.EventService.GetMyEvents(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, "查询失败")
	}

	response.Success(c, types.EventListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
	return nil
}

// GetEventDetail 获取活动详情
// GET /api/v1/event/:id
func (e *Event) GetEventDetail(c *gin.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return response.NewError(http.StatusBadRequest, "无效的活动ID")
	}

	detail, err := e.EventService.GetEventDetail(c.Request.Context(), id)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, "查询失败: "+err.Error())
	}

	response.Success(c, detail)
	return nil
}
