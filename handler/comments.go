package handler

import (
	"Hyper/config"
	"Hyper/middleware"
	"Hyper/pkg/context"
	"Hyper/pkg/response"
	"Hyper/service"
	"Hyper/types"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CommentsHandler struct {
	Config          *config.Config
	CommentsService service.ICommentsService
}

func (ch *CommentsHandler) RegisterRouter(r gin.IRouter) {
	authorize := middleware.Auth([]byte(ch.Config.Jwt.Secret))
	comments := r.Group("/v1/comments")
	comments.POST("/create", authorize, context.Wrap(ch.CreateComment)) //创建评论
	comments.GET("/list/:note_id", authorize, context.Wrap(ch.GetComments))
	comments.GET("/replies/:rootId", authorize, context.Wrap(ch.GetReplyComments))
	comments.POST("/delete", authorize, context.Wrap(ch.DeleteComment))
	comments.POST("/like", authorize, context.Wrap(ch.LikeComment)) //点赞评论
	comments.POST("/unlike", authorize, context.Wrap(ch.UnlikeComment))
}

// CreateComment 创建评论
func (ch *CommentsHandler) CreateComment(c *gin.Context) error {
	var req types.CreateCommentRequest

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
	comment, err := ch.CommentsService.CreateComment(c, &req, userID)
	if err != nil {
		return response.NewError(http.StatusBadRequest, "创建评论失败: "+err.Error())
	}

	response.Success(c, comment)
	return nil
}

// handler/comment_handler.go

// GetComments 获取评论列表(游标分页)
func (ch *CommentsHandler) GetComments(c *gin.Context) error {
	noteIDStr := c.Param("note_id")
	noteID, err := strconv.ParseUint(noteIDStr, 10, 64)
	if err != nil || noteID == 0 {
		return response.NewError(http.StatusBadRequest, "note_id参数错误")
	}

	// 游标(可选)
	cursor := int64(0)
	if cursorStr := c.Query("cursor"); cursorStr != "" {
		if v, err := strconv.ParseInt(cursorStr, 10, 64); err == nil {
			cursor = v
		}
	}

	// 每页数量
	pageSize := 20
	if ps := c.Query("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
			pageSize = v
		}
	}

	// 获取当前用户ID(可能未登录)
	currentUserID := uint64(0)
	if userIDval, err := context.GetUserID(c); err == nil {
		currentUserID = uint64(userIDval)
	}

	// 调用 Service
	result, err := ch.CommentsService.GetComments(c.Request.Context(), noteID, cursor, pageSize, currentUserID)
	if err != nil {
		return response.NewError(http.StatusBadRequest, "获取评论失败: "+err.Error())
	}

	response.Success(c, result)
	return nil
}

// GetReplyComments 获取回复列表(游标分页)
func (ch *CommentsHandler) GetReplyComments(c *gin.Context) error {
	rootIDStr := c.Param("rootId")
	rootID, err := strconv.ParseUint(rootIDStr, 10, 64)
	if err != nil || rootID == 0 {
		return response.NewError(http.StatusBadRequest, "rootId参数错误")
	}

	// 游标(可选)
	cursor := int64(0)
	if cursorStr := c.Query("cursor"); cursorStr != "" {
		if v, err := strconv.ParseInt(cursorStr, 10, 64); err == nil {
			cursor = v
		}
	}

	// 每页数量
	pageSize := 20
	if ps := c.Query("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
			pageSize = v
		}
	}

	// 获取当前用户ID
	currentUserID := uint64(0)
	if userIDval, err := context.GetUserID(c); err == nil {
		currentUserID = uint64(userIDval)
	}

	// 调用 Service
	result, err := ch.CommentsService.GetReplies(c.Request.Context(), rootID, cursor, pageSize, currentUserID)
	if err != nil {
		return response.NewError(http.StatusBadRequest, "获取回复失败: "+err.Error())
	}

	response.Success(c, result)
	return nil
}
func (ch *CommentsHandler) DeleteComment(c *gin.Context) error {
	var req types.DeleteCommentRequest

	// 1. 绑定参数
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, "请求参数失败"+err.Error())
	}
	userIDval, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(http.StatusUnauthorized, "未登录")
	}

	userID := uint64(userIDval)
	if userID == 0 {
		return response.NewError(http.StatusUnauthorized, "用户ID无效")
	}

	// 4. 执行业务逻辑
	if err := ch.CommentsService.DeleteComment(c, req.CommentID, userID); err != nil {
		return response.NewError(http.StatusBadRequest, "删除评论失败: "+err.Error())
	}

	response.Success(c, "删除评论成功")

	return nil
}

func (ch *CommentsHandler) LikeComment(c *gin.Context) error {
	var req types.LikeCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, "请求参数失败"+err.Error())
	}
	userIDval, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(http.StatusUnauthorized, "未登录")
	}

	userID := uint64(userIDval)
	if userID == 0 {
		return response.NewError(http.StatusUnauthorized, "用户ID无效")
	}
	if err := ch.CommentsService.LikeComment(c, req.CommentID, userID); err != nil {
		return response.NewError(http.StatusBadRequest, "点赞评论失败: "+err.Error())
	}
	response.Success(c, "ok")
	return nil
}

func (ch *CommentsHandler) UnlikeComment(c *gin.Context) error {

	var req types.UnlikeCommentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, "请求参数失败"+err.Error())
	}
	userIDval, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(http.StatusUnauthorized, "未登录")
	}

	userID := uint64(userIDval)
	if userID == 0 {
		return response.NewError(http.StatusUnauthorized, "用户ID无效")
	}
	if err := ch.CommentsService.UnlikeComment(c, req.CommentID, userID); err != nil {
		return response.NewError(http.StatusBadRequest, "取消点赞评论失败: "+err.Error())
	}
	response.Success(c, "评论取消点赞成功")
	return nil
}
