package handler

import (
	"Hyper/config"
	"Hyper/middleware"
	"Hyper/pkg/context"
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

func NewCommentsHandler(config *config.Config, commentsService service.ICommentsService) *CommentsHandler {
	return &CommentsHandler{
		CommentsService: commentsService,
		Config:          config,
	}
}

func (ch *CommentsHandler) RegisterRouter(r gin.IRouter) {
	authorize := middleware.Auth([]byte(ch.Config.Jwt.Secret))
	comments := r.Group("/comments")
	comments.POST("/create", authorize, context.Wrap(ch.CreateComment)) //创建评论
	comments.GET("/list/:postId", context.Wrap(ch.GetComments))
	comments.GET("/replies/:rootId", context.Wrap(ch.GetReplyComments))
	comments.POST("/delete", authorize, context.Wrap(ch.DeleteComment))
	comments.POST("/like", authorize, context.Wrap(ch.LikeComment)) //点赞评论
	comments.POST("/unlike", authorize, context.Wrap(ch.UnlikeComment))
}

// CreateComment 创建评论
func (ch *CommentsHandler) CreateComment(c *gin.Context) error {
	var req types.CreateCommentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数失败" + err.Error(),
		})
		return err
	}

	userIDval, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未授权,请登录",
		})
		return nil
	}
	userID := uint64(userIDval.(int))
	//userID := uint64(1)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "用户ID无效",
		})
		return nil
	}
	comment, err := ch.CommentsService.CreateComment(c, &req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "创建评论失败: " + err.Error(),
		})
		return nil
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "评论创建成功",
		"data":    comment,
	})
	return nil
}

// GetComments 获取评论列表
func (ch *CommentsHandler) GetComments(c *gin.Context) error {
	postIdStr := c.Param("postId")
	postID, err := strconv.ParseUint(postIdStr, 10, 64)
	if err != nil || postID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "post_id参数错误",
		})
		return err
	}

	page := 1
	pageSize := 20

	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}

	if ps := c.Query("pageSize"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
			pageSize = v
		}
	}
	comments, total, err := ch.CommentsService.GetComments(c, postID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "获取评论失败: " + err.Error(),
		})
		return nil
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取评论成功",
		"data": gin.H{
			"comments":  comments,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
	return nil
}

func (ch *CommentsHandler) GetReplyComments(c *gin.Context) error {
	//获取评论的rootid
	rootId := c.Param("rootId")
	rootID, err := strconv.ParseUint(rootId, 10, 64)
	if err != nil || rootID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "root_id参数错误",
		})
		return err
	}

	page := 1
	pageSize := 20

	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if ps := c.Query("pageSize"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
			pageSize = v
		}
	}

	replies, total, err := ch.CommentsService.GetReplies(c, rootID, page, pageSize)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "获取回复评论失败: " + err.Error(),
		})
		return nil
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取回复评论成功",
		"data": gin.H{
			"replies":   replies,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
	return nil
}

func (ch *CommentsHandler) DeleteComment(c *gin.Context) error {
	var req types.DeleteCommentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数失败" + err.Error(),
		})
		return err
	}
	userIDval, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未授权,请登录",
		})
		return nil
	}
	userID := uint64(userIDval.(int))
	//userID := uint64(1)

	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "用户ID无效",
		})
		return nil
	}
	err := ch.CommentsService.DeleteComment(c, req.CommentID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "删除评论失败: " + err.Error(),
		})
		return nil
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "评论删除成功",
	})
	return nil
}

func (ch *CommentsHandler) LikeComment(c *gin.Context) error {
	var req types.LikeCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数失败" + err.Error(),
		})
		return err
	}
	userIDval, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未授权,请登录",
		})
		return nil
	}
	userID := uint64(userIDval.(int))
	//userID := uint64(1)

	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "用户ID无效",
		})
		return nil
	}
	err := ch.CommentsService.LikeComment(c, req.CommentID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "点赞评论失败: " + err.Error(),
		})
		return nil
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "评论点赞成功",
	})
	return nil
}

func (ch *CommentsHandler) UnlikeComment(c *gin.Context) error {

	var req types.UnlikeCommentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数失败" + err.Error(),
		})
		return err
	}
	userIDval, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未授权,请登录",
		})
		return nil
	}
	userID := uint64(userIDval.(int))
	//userID := uint64(1)

	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "用户ID无效",
		})
		return nil
	}

	err := ch.CommentsService.UnlikeComment(c, req.CommentID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "取消点赞评论失败: " + err.Error(),
		})
		return nil
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "评论取消点赞成功",
	})
	return nil
}
