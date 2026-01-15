package handler

import (
	"Hyper/config"
	"Hyper/middleware"
	"Hyper/pkg/context"
	"Hyper/pkg/response"
	"Hyper/service"
	"Hyper/types"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type TopicHandler struct {
	Config       *config.Config
	TopicService service.ITopicService
}

func NewTopicHandler(config *config.Config, topicService service.ITopicService) *TopicHandler {
	return &TopicHandler{
		TopicService: topicService,
		Config:       config,
	}
}

func (th *TopicHandler) RegisterRouter(r gin.IRouter) {
	authorize := middleware.Auth([]byte(th.Config.Jwt.Secret))
	topics := r.Group("topics") // ✅ 改为相对路径，不要加 /v1/
	topics.POST("/create", authorize, context.Wrap(th.CreateTopic))
	//测试成功
	topics.POST("/test/batch-create", context.Wrap(th.TestBatchCreateTopics))
	topics.POST("/test/extract", context.Wrap(th.TestExtractAndAssociate))
	topics.GET("/:topicId/notes", context.Wrap(th.GetTopicNotes)) // 👈 添加这一行
}

func (th *TopicHandler) CreateTopic(c *gin.Context) error {
	var req types.CreateTopicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, err.Error())
	}
	userIDval, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(http.StatusUnauthorized, "未登录")
	}
	userID := uint64(userIDval)
	if userID == 0 {
		return response.NewError(http.StatusUnauthorized, "用户ID无效")
	}
	topic, err := th.TopicService.CreateNewTopic(c, &req, userID)
	if err != nil {
		return response.NewError(http.StatusBadRequest, "创建话题失败: "+err.Error())
	}
	response.Success(c, topic)
	return nil
}

// ============ 测试接口 ============

// TestBatchCreateTopics 测试批量创建话题
// POST /v1/topics/test/batch-create
// 请求体：
//
//	{
//	  "topic_names": ["旅行", "美食", "摄影"]
//	}
func (th *TopicHandler) TestBatchCreateTopics(c *gin.Context) error {
	var req struct {
		TopicNames []string `json:"topic_names" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, "请求参数错误: "+err.Error())
	}

	userIDval, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(http.StatusUnauthorized, "未登录")
	}
	//userID := uint64(1)
	userID := uint64(userIDval)
	topicMap, err := th.TopicService.BatchCreateTopics(c.Request.Context(), req.TopicNames, userID)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, "批量创建话题失败: "+err.Error())
	}

	// ============ 构建响应 ============
	type TopicInfo struct {
		ID          uint64 `json:"id"`
		Name        string `json:"name"`
		PostCount   uint32 `json:"post_count"`
		ViewCount   uint32 `json:"view_count"`
		FollowCount uint32 `json:"follow_count"`
		IsHot       bool   `json:"is_hot"`
	}

	var topics []TopicInfo
	for name, topic := range topicMap {
		topics = append(topics, TopicInfo{
			ID:          topic.ID,
			Name:        name,
			PostCount:   topic.PostCount,
			ViewCount:   topic.ViewCount,
			FollowCount: topic.FollowCount,
			IsHot:       topic.IsHot,
		})
	}

	response.Success(c, map[string]interface{}{
		"count":  len(topics),
		"topics": topics,
	})
	return nil
}

// TestExtractAndAssociate 测试提取话题并关联
// POST /v1/topics/test/extract
// 请求体：
//
//	{
//	  "note_id": 123,
//	  "content": "今天去旅行 #旅行 #美食"
//	}
func (th *TopicHandler) TestExtractAndAssociate(c *gin.Context) error {
	var req struct {
		NoteID  uint64 `json:"note_id" binding:"required"`
		Content string `json:"content" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, "请求参数错误: "+err.Error())
	}

	//userIDval, err := context.GetUserID(c)
	//if err != nil {
	//	return response.NewError(http.StatusUnauthorized, "未登录")
	//}
	//userID := uint64(userIDval)
	userID := uint64(1) // 测试用户ID
	// ============ 调用 TopicService 提取并关联 ============
	topics, err := th.TopicService.ExtractAndAssociateTopics(c, req.NoteID, req.Content, userID)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, "提取并关联话题失败: "+err.Error())
	}

	// ============ 构建响应 ============
	response.Success(c, map[string]interface{}{
		"note_id": req.NoteID,
		"count":   len(topics),
		"topics":  topics,
	})
	return nil
}

// GetTopicNotes 获取话题下的笔记列表
func (th *TopicHandler) GetTopicNotes(c *gin.Context) error {
	// 1. 解析路径参数 topicId
	topicIDStr := c.Param("topicId")
	topicID, err := strconv.ParseUint(topicIDStr, 10, 64)
	if err != nil {
		return response.NewError(http.StatusBadRequest, "话题ID无效")
	}

	// 2. 解析查询参数
	// cursor 是分页游标，表示上一次查询的最后一条记录的时间戳
	cursorStr := c.DefaultQuery("cursor", "0")
	cursor, err := strconv.ParseInt(cursorStr, 10, 64)
	if err != nil {
		cursor = 0
	}

	// page_size 是每页的笔记数，默认 10，最大 100
	pageSize := 10
	pageSizeStr := c.Query("page_size")
	if pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	// 3. 调用 Service 层获取话题笔记
	resp, err := th.TopicService.GetTopicNotes(c.Request.Context(), topicID, cursor, pageSize)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, "获取话题笔记失败: "+err.Error())
	}

	// 4. 返回成功响应
	response.Success(c, resp)
	return nil
}
