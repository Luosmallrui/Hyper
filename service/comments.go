package service

import (
	"Hyper/models"
	"Hyper/types"
	"context"
	"errors"

	"gorm.io/gorm"
)

var _ ICommentsService = (*CommentsService)(nil)

type ICommentsService interface {
	GetCommentByID(ctx context.Context, commentID uint64) (*models.Comments, error)
	CreateComment(ctx context.Context, req *types.CreateCommentRequest, userID uint64) (*models.Comments, error)
	GetComments(ctx context.Context, postId uint64, page, pageSize int) ([]*models.Comments, int64, error)
	GetReplies(ctx context.Context, rootID uint64, page, pageSize int) ([]*models.Comments, int64, error)
}

type CommentsService struct {
	DB *gorm.DB
}

func NewCommentsService(db *gorm.DB) *CommentsService {
	return &CommentsService{
		DB: db,
	}
}

// 测试评论接口
func (cs *CommentsService) TestComments() error {
	return nil
}

func (cs *CommentsService) GetCommentByID(ctx context.Context, commentID uint64) (*models.Comments, error) {
	var comment models.Comments
	if err := cs.DB.WithContext(ctx).
		Where("id = ? AND status = 1", commentID).
		First(&comment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("评论不存在: " + err.Error())
		}
		return nil, errors.New("查询失败" + err.Error())
	}
	return &comment, nil
}

// 创建评论
func (cs *CommentsService) CreateComment(ctx context.Context, req *types.CreateCommentRequest, userID uint64) (*models.Comments, error) {
	if req.PostID == 0 {
		return nil, errors.New("PostID不能为空")
	}

	if req.Content == "" {
		return nil, errors.New("Content不能为空")
	}
	//内容长度限制
	if len(req.Content) > 1000 {
		return nil, errors.New("Content长度不能超过1000字符")
	}

	//如果是直接评论帖子，则RootID和ParentID都为0，ReplyToUserID为空（顶级评论）
	if req.ParentID == 0 {
		req.RootID = 0
		req.ReplyToUserID = nil
	} else {
		//如果是回复评论，则RootID和ParentID必须大于0
		parentComment, err := cs.GetCommentByID(ctx, req.ParentID)
		if err != nil {
			return nil, errors.New("回复的评论不存在: " + err.Error())
		}
		if parentComment.Status != 1 {
			return nil, errors.New("回复的评论已经删除" + err.Error())
		}
		if parentComment.PostID != req.PostID {
			return nil, errors.New("回复的评论不属于该帖子")
		}
		//如果回复的是直接回复帖子的回复（也就是一级评论），则RootID为该评论ID
		if parentComment.RootID == 0 {
			// 被回复的评论的 RootID = 0
			// 说明：被回复的评论就是一级评论（顶级）
			// 结论：新回复应该属于这个一级评论
			// 所以：新回复的 RootID = 这个一级评论的 ID
			// 而这个 ID 就是 req.ParentID
			req.RootID = req.ParentID
		} else {
			// 被回复的评论的 RootID ≠ 0
			// 说明：被回复的评论不是一级评论
			// 结论：新回复应该属于同一个顶级评论
			// 所以：新回复的 RootID = 被回复评论的 RootID
			req.RootID = parentComment.RootID
		}
		// 设置被回复人的用户ID
		if req.ReplyToUserID == nil || *req.ReplyToUserID == 0 {
			req.ReplyToUserID = &parentComment.UserID
		}
	}
	comment := &models.Comments{
		PostID:        req.PostID,
		UserID:        userID,
		RootID:        req.RootID,
		ParentID:      req.ParentID,
		ReplyToUserID: req.ReplyToUserID,
		Content:       req.Content,
		Status:        1,
		LikeCount:     0,
	}

	if err := cs.DB.WithContext(ctx).Create(comment).Error; err != nil {
		return nil, errors.New("创建评论失败: " + err.Error())
	}

	return comment, nil
}

// 获取评论详情(一级评论)，分页，每页显示多少评论（一次加载所有评论非常慢）
func (cs *CommentsService) GetComments(ctx context.Context, postId uint64, page, pageSize int) ([]*models.Comments, int64, error) {
	var comments []*models.Comments
	var total int64

	if postId == 0 {
		return nil, 0, errors.New("帖子的Id不存在")
	}

	if page <= 0 {
		page = 1
	}

	if pageSize <= 0 {
		pageSize = 20
	}

	// 获取评论总数
	if err := cs.DB.WithContext(ctx).
		Where("post_id = ? AND root_id = 0 AND status = 1", postId).
		Count(&total).Error; err != nil {
		return nil, 0, errors.New("获取评论总数失败: " + err.Error())
	}
	// 分页获取评论
	if err := cs.DB.WithContext(ctx).
		Where("post_id = ? AND root_id = 0 AND status = 1", postId).
		Order("created_at DESC"). // 最新评论在前
		Offset((page - 1) * pageSize). //跳过前面的评论
		Limit(pageSize). //限制取到的评论数（下面的Find是把获取的评论传给comments）
		Find(&comments).Error; err != nil {
		return nil, 0, errors.New("获取评论失败: " + err.Error())
	}

	return comments, total, nil
}

// 获取二级评论（也就是评论的回复）
func (cs *CommentsService) GetReplies(ctx context.Context, rootID uint64, page, pageSize int) ([]*models.Comments, int64, error) {
	var replies []*models.Comments
	var total int64

	if rootID == 0 {
		return nil, 0, errors.New("该评论不存在")
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	_, err := cs.GetCommentByID(ctx, rootID)
	if err != nil {
		return nil, 0, errors.New("该评论不存在: " + err.Error())
	}
	if err := cs.DB.WithContext(ctx).
		Where("root_id = ? AND status = 1", rootID).
		Count(&total).Error; err != nil {
		return nil, 0, errors.New("获取回复总数失败: " + err.Error())
	}
	//从早到晚
	if err := cs.DB.WithContext(ctx).
		Where("root_id = ? AND status = 1", rootID).
		Order("created_at ASC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&replies).Error; err != nil {
		return nil, 0, errors.New("获取回复失败: " + err.Error())
	}
	return replies, total, nil
}
