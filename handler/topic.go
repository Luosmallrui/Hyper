package handler

import (
	"Hyper/config"
	"Hyper/pkg/context"
	"Hyper/pkg/response"
	"Hyper/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
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
	//authorize := middleware.Auth([]byte(th.Config.Jwt.Secret))
	topics := r.Group("/v1/topics")
	topics.GET("/search", context.Wrap(th.SearchTopics))          // 搜索话题
	topics.GET("/:topicID/notes", context.Wrap(th.GetTopicNotes)) // 新增：获取话题的笔记列表
}

func (th *TopicHandler) SearchTopics(c *gin.Context) error {
	//var req types.SearchTopicsRequest
	//if err := c.ShouldBindJSON(&req); err != nil {
	//	return response.NewError(http.StatusBadRequest, err.Error())
	//}
	query := c.Query("query")
	//userIDval, err := context.GetUserID(c)
	//if err != nil {
	//	return response.NewError(http.StatusInternalServerError, err.Error())
	//}
	//userID := int64(userIDval)
	//userID := int64(1)
	topics, err := th.TopicService.SearchTopics(c.Request.Context(), query)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, "搜索话题失败: "+err.Error())
	}
	//没有查到就先提示没查到，不新建
	//if query != "" && len(topics) == 0 {
	//	//topicResp, err := th.TopicService.CreateTopicIfNotExists(c.Request.Context(), query, uint64(userID))
	//	if err != nil {
	//		return response.NewError(http.StatusInternalServerError, "话题为空"+err.Error())
	//	}
	//
	//	//topics = []types.CreateOrGetTopicResponse{
	//	//	{
	//	//		ID:        topicResp.ID,
	//	//		Name:      topicResp.Name,
	//	//		ViewCount: topicResp.ViewCount,
	//	//	},
	//	//}
	//}

	response.Success(c, topics)
	return nil
}

func (th *TopicHandler) GetTopicNotes(c *gin.Context) error {
	topicIDStr := c.Param("topicID")
	topicID, err := strconv.ParseUint(topicIDStr, 10, 64)
	if err != nil {
		return response.NewError(http.StatusBadRequest, "无效的话题ID")
	}

	cursorStr := c.DefaultQuery("cursor", "0")
	limitStr := c.DefaultQuery("limit", "10")

	cursor, err := strconv.ParseInt(cursorStr, 10, 64)
	if err != nil {
		cursor = 0
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	var currentUserID uint64 = 0
	userIDval, err := context.GetUserID(c)
	if err == nil {
		currentUserID = uint64(userIDval)
	}

	result, err := th.TopicService.GetNotesByTopic(c.Request.Context(), topicID, cursor, limit, currentUserID)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, "获取话题笔记失败: "+err.Error())
	}

	response.Success(c, result)
	return nil
}
