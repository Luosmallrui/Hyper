package service

import (
	"Hyper/dao"
	"Hyper/models"
	"Hyper/pkg/log"
	"Hyper/pkg/snowflake"
	"Hyper/types"
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"gorm.io/gorm"
)

const (
	CommentLikeCountKey  = "comment:like:count:%d"  // 评论点赞数
	UserLikedCommentsKey = "user:liked:comments:%d" // 用户点赞的评论集合
)

var _ ICommentsService = (*CommentsService)(nil)

type CommentsService struct {
	DB             *gorm.DB
	CommentDAO     *dao.Comment
	CommentLikeDAO *dao.CommentLike
	UserService    IUserService
	Redis          *redis.Client
}

type ICommentsService interface {
	CreateComment(ctx context.Context, req *types.CreateCommentRequest, userID uint64) (*types.CommentResponse, error)
	GetComments(ctx context.Context, noteID uint64, cursor int64, pageSize int, currentUserID uint64) (*types.CommentsListResponse, error)
	GetReplies(ctx context.Context, rootID uint64, cursor int64, pageSize int, currentUserID uint64) (*types.RepliesListResponse, error)
	DeleteComment(ctx context.Context, commentID, userID uint64) error
	LikeComment(ctx context.Context, commentID, userID uint64) error
	UnlikeComment(ctx context.Context, commentID, userID uint64) error
	GetTopComments(ctx context.Context, noteID uint64, limit int, currentUserID uint64) ([]*types.CommentResponse, error)
}

func (s *CommentsService) GetTopComments(ctx context.Context, noteID uint64, limit int, currentUserID uint64) ([]*types.CommentResponse, error) {
	// 1. 获取前N条一级评论
	comments, err := s.CommentDAO.GetRootCommentsByCursor(ctx, noteID, 0, limit)
	if err != nil {
		return nil, err
	}

	if len(comments) == 0 {
		return make([]*types.CommentResponse, 0), nil
	}

	// 2. 收集用户ID和评论ID
	commentIDs := make([]uint64, 0, len(comments))
	userIDs := make([]uint64, 0, len(comments))

	for _, comment := range comments {
		commentIDs = append(commentIDs, comment.ID)
		userIDs = append(userIDs, comment.UserID)
	}

	// 3. 并发获取关联数据
	var (
		userMap       map[uint64]types.UserProfile
		likeStatusMap map[uint64]bool
		wg            sync.WaitGroup
		mu            sync.Mutex
	)

	wg.Add(2)

	// 获取用户信息
	go func() {
		defer wg.Done()
		userMap = s.UserService.BatchGetUserInfo(ctx, userIDs)
	}()

	// 获取点赞状态
	go func() {
		defer wg.Done()
		if currentUserID > 0 {
			likes, err := s.CommentLikeDAO.BatchCheckExists(ctx, commentIDs, currentUserID)
			if err == nil {
				mu.Lock()
				likeStatusMap = likes
				mu.Unlock()
			}
		}
	}()

	wg.Wait()

	// 4. 组装响应(不包含回复,笔记详情页只展示评论本身)
	result := make([]*types.CommentResponse, 0, len(comments))

	for _, comment := range comments {
		resp := &types.CommentResponse{
			ID:            comment.ID,
			NoteID:        comment.NoteID,
			UserID:        comment.UserID,
			Content:       comment.Content,
			LikeCount:     comment.LikeCount,
			ReplyCount:    comment.ReplyCount,
			IPLocation:    comment.IPLocation,
			IsLiked:       likeStatusMap[comment.ID],
			CreatedAt:     comment.CreatedAt,
			User:          userMap[comment.UserID],
			LatestReplies: make([]*types.ReplyResponse, 0), // 详情页不展示回复
		}

		result = append(result, resp)
	}

	return result, nil
}

func (s *CommentsService) LikeComment(ctx context.Context, commentID, userID uint64) error {
	// 1. 分布式锁,防止重复点赞
	lockKey := fmt.Sprintf("lock:comment:like:%d:%d", userID, commentID)
	lock, err := s.Redis.SetNX(ctx, lockKey, 1, 5*time.Second).Result()
	if err != nil || !lock {
		return errors.New("操作太频繁,请稍后重试")
	}
	defer s.Redis.Del(ctx, lockKey)

	// 2. 检查是否已点赞(先查 Redis,再查 DB)
	isLiked, err := s.checkCommentLikeStatus(ctx, userID, commentID)
	if err != nil {
		return err
	}
	if isLiked {
		return errors.New("已经点赞过了")
	}

	// 3. 先写数据库(保证数据一致性)
	if err := s.createCommentLikeRecord(ctx, userID, commentID); err != nil {
		return err
	}

	// 4. 更新 Redis 缓存(即使失败也不影响)
	s.updateRedisAfterCommentLike(ctx, userID, commentID)

	return nil
}

func (s *CommentsService) UnlikeComment(ctx context.Context, commentID, userID uint64) error {
	lockKey := fmt.Sprintf("lock:comment:like:%d:%d", userID, commentID)
	lock, err := s.Redis.SetNX(ctx, lockKey, 1, 5*time.Second).Result()
	if err != nil || !lock {
		return errors.New("操作太频繁,请稍后重试")
	}
	defer s.Redis.Del(ctx, lockKey)

	// 检查是否已点赞
	isLiked, err := s.checkCommentLikeStatus(ctx, userID, commentID)
	if err != nil {
		return err
	}
	if !isLiked {
		return errors.New("还未点赞")
	}

	// 先写数据库
	if err := s.deleteCommentLikeRecord(ctx, userID, commentID); err != nil {
		return err
	}

	// 更新 Redis
	s.updateRedisAfterCommentUnlike(ctx, userID, commentID)

	return nil
}

// 检查点赞状态
func (s *CommentsService) checkCommentLikeStatus(ctx context.Context, userID, commentID uint64) (bool, error) {
	// 1. 先查 Redis Set
	userLikedKey := fmt.Sprintf(UserLikedCommentsKey, userID)
	exists, err := s.Redis.SIsMember(ctx, userLikedKey, commentID).Result()
	if err == nil {
		return exists, nil
	}

	// 2. Redis 没有,查数据库
	exists, err = s.CommentLikeDAO.CheckExists(ctx, commentID, userID)
	if err != nil {
		return false, err
	}

	// 3. 回写 Redis
	if exists {
		s.Redis.SAdd(ctx, userLikedKey, commentID)
		s.Redis.Expire(ctx, userLikedKey, CacheTTL)
	}

	return exists, nil
}

// 写数据库:创建点赞记录 + 更新计数
func (s *CommentsService) createCommentLikeRecord(ctx context.Context, userID, commentID uint64) error {
	return s.CommentDAO.Transaction(ctx, func(tx *gorm.DB) error {
		// 1. 插入点赞记录
		like := &models.CommentLike{
			CommentID: commentID,
			UserID:    userID,
			CreatedAt: time.Now(),
		}
		if err := tx.Create(like).Error; err != nil {
			if strings.Contains(err.Error(), "Duplicate entry") {
				return errors.New("已经点赞过了")
			}
			return err
		}

		// 2. 更新点赞数
		return tx.Model(&models.Comment{}).
			Where("id = ?", commentID).
			UpdateColumn("like_count", gorm.Expr("like_count + 1")).
			Error
	})
}

// 写数据库:删除点赞记录 + 更新计数
func (s *CommentsService) deleteCommentLikeRecord(ctx context.Context, userID, commentID uint64) error {
	return s.CommentDAO.Transaction(ctx, func(tx *gorm.DB) error {
		// 1. 删除点赞记录
		result := tx.Where("comment_id = ? AND user_id = ?", commentID, userID).
			Delete(&models.CommentLike{})

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return errors.New("点赞记录不存在")
		}

		// 2. 更新点赞数
		return tx.Model(&models.Comment{}).
			Where("id = ?", commentID).
			UpdateColumn("like_count", gorm.Expr("like_count - 1")).
			Error
	})
}

// 更新 Redis:点赞后
func (s *CommentsService) updateRedisAfterCommentLike(ctx context.Context, userID, commentID uint64) {
	pipe := s.Redis.Pipeline()

	// 1. 点赞数 +1
	likeCountKey := fmt.Sprintf(CommentLikeCountKey, commentID)
	pipe.Incr(ctx, likeCountKey)
	pipe.Expire(ctx, likeCountKey, CacheTTL)

	// 2. 用户点赞集合添加
	userLikedKey := fmt.Sprintf(UserLikedCommentsKey, userID)
	pipe.SAdd(ctx, userLikedKey, commentID)
	pipe.Expire(ctx, userLikedKey, CacheTTL)

	// 执行 Pipeline (即使失败也不影响业务)
	if _, err := pipe.Exec(ctx); err != nil {
		// 记录日志
		log.L.Error("更新Redis缓存失败", zap.Error(err))
	}
}

// 更新 Redis:取消点赞后
func (s *CommentsService) updateRedisAfterCommentUnlike(ctx context.Context, userID, commentID uint64) {
	pipe := s.Redis.Pipeline()

	// 1. 点赞数 -1
	likeCountKey := fmt.Sprintf(CommentLikeCountKey, commentID)
	pipe.Decr(ctx, likeCountKey)

	// 2. 用户点赞集合移除
	userLikedKey := fmt.Sprintf(UserLikedCommentsKey, userID)
	pipe.SRem(ctx, userLikedKey, commentID)

	pipe.Exec(ctx)
}

// 批量获取评论点赞数
func (s *CommentsService) BatchGetCommentLikeCount(ctx context.Context, commentIDs []uint64) (map[uint64]int, error) {
	result := make(map[uint64]int)
	if len(commentIDs) == 0 {
		return result, nil
	}

	missedIDs := make([]uint64, 0)

	// 1. 批量从 Redis 获取
	pipe := s.Redis.Pipeline()
	cmds := make(map[uint64]*redis.StringCmd)

	for _, commentID := range commentIDs {
		key := fmt.Sprintf(CommentLikeCountKey, commentID)
		cmds[commentID] = pipe.Get(ctx, key)
	}

	pipe.Exec(ctx)

	// 2. 收集 Redis 结果
	for commentID, cmd := range cmds {
		count, err := cmd.Int()
		if err == nil {
			result[commentID] = count
		} else {
			missedIDs = append(missedIDs, commentID)
		}
	}

	// 3. 从数据库加载缺失的(直接读 comments 表的 like_count 字段)
	if len(missedIDs) > 0 {
		var comments []*models.Comment
		err := s.CommentDAO.Db.WithContext(ctx).
			Select("id, like_count").
			Where("id IN ?", missedIDs).
			Find(&comments).Error

		if err != nil {
			return result, err
		}

		// 回写 Redis
		pipe2 := s.Redis.Pipeline()
		for _, comment := range comments {
			result[comment.ID] = comment.LikeCount
			key := fmt.Sprintf(CommentLikeCountKey, comment.ID)
			pipe2.Set(ctx, key, comment.LikeCount, CacheTTL)
		}
		pipe2.Exec(ctx)
	}

	return result, nil
}

// validateCreateComment 验证创建评论请求
func (s *CommentsService) validateCreateComment(req *types.CreateCommentRequest) error {
	if req.NoteID == 0 {
		return errors.New("笔记ID不能为空")
	}

	content := strings.TrimSpace(req.Content)
	if content == "" {
		return errors.New("评论内容不能为空")
	}

	if len(content) > 1000 {
		return errors.New("评论内容不能超过1000个字符")
	}

	// 如果是回复评论
	if req.RootID > 0 {
		if req.ParentID == 0 {
			return errors.New("回复评论时必须指定父评论ID")
		}
	}

	return nil
}

func (s *CommentsService) DeleteComment(ctx context.Context, commentID, userID uint64) error {
	// 1. 查询评论
	comment, err := s.CommentDAO.GetByID(ctx, commentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("评论不存在")
		}
		return err
	}

	// 2. 权限检查(只能删除自己的评论)
	if comment.UserID != userID {
		return errors.New("无权删除该评论")
	}

	// 3. 使用事务删除
	return s.CommentDAO.Transaction(ctx, func(tx *gorm.DB) error {
		// 3.1 删除评论
		if err := tx.Model(&models.Comment{}).
			Where("id = ?", commentID).
			Update("status", 0).Error; err != nil {
			return err
		}

		// 3.2 如果是回复,更新一级评论的回复数
		if comment.RootID > 0 {
			if err := tx.Model(&models.Comment{}).
				Where("id = ?", comment.RootID).
				UpdateColumn("reply_count", gorm.Expr("reply_count - 1")).
				Error; err != nil {
				return err
			}
		}

		// 3.3 更新笔记统计表的评论数
		if err := tx.Model(&models.NoteStats{}).
			Where("note_id = ?", comment.NoteID).
			UpdateColumn("comment_count", gorm.Expr("comment_count - 1")).
			Error; err != nil {
			return err
		}

		return nil
	})
}

// CreateComment 创建评论
func (s *CommentsService) CreateComment(ctx context.Context, req *types.CreateCommentRequest, userID uint64) (*types.CommentResponse, error) {
	// 1. 参数验证
	//if err := s.validateCreateComment(req); err != nil {
	//	return nil, err
	//}

	// 2. 生成评论ID
	commentID := uint64(snowflake.GenUserID())

	// 3. 构建评论对象
	now := time.Now()
	comment := &models.Comment{
		ID:            commentID,
		NoteID:        req.NoteID,
		UserID:        userID,
		RootID:        req.RootID,
		ParentID:      req.ParentID,
		ReplyToUserID: uint64(req.ReplyToUserID),
		Content:       strings.TrimSpace(req.Content),
		LikeCount:     0,
		ReplyCount:    0,
		IPLocation:    "", // TODO: 从请求中获取IP归属地
		Status:        1,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// 4. 使用事务保存
	err := s.CommentDAO.Transaction(ctx, func(tx *gorm.DB) error {
		// 4.1 创建评论
		if err := tx.Create(comment).Error; err != nil {
			return err
		}

		// 4.2 如果是回复(二级评论),更新一级评论的回复数
		if req.RootID > 0 {
			if err := tx.Model(&models.Comment{}).
				Where("id = ?", req.RootID).
				UpdateColumn("reply_count", gorm.Expr("reply_count + 1")).
				Error; err != nil {
				return err
			}
		}

		// 4.3 更新笔记统计表的评论数
		if err := tx.Model(&models.NoteStats{}).
			Where("note_id = ?", req.NoteID).
			UpdateColumn("comment_count", gorm.Expr("comment_count + 1")).
			Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 5. 组装返回数据
	users := s.UserService.BatchGetUserInfo(ctx, []uint64{userID})
	user := users[userID]

	resp := &types.CommentResponse{
		ID:            comment.ID,
		NoteID:        comment.NoteID,
		UserID:        comment.UserID,
		Content:       comment.Content,
		LikeCount:     comment.LikeCount,
		ReplyCount:    comment.ReplyCount,
		IPLocation:    comment.IPLocation,
		IsLiked:       false,
		CreatedAt:     comment.CreatedAt,
		User:          user,
		LatestReplies: make([]*types.ReplyResponse, 0),
	}

	return resp, nil
}

// GetComments 获取一级评论列表(游标分页)
func (s *CommentsService) GetComments(ctx context.Context, noteID uint64, cursor int64, pageSize int, currentUserID uint64) (*types.CommentsListResponse, error) {
	// 1. 参数校验
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	// 2. 多查一条判断是否还有更多
	limit := pageSize + 1

	// 3. 获取评论列表
	comments, err := s.CommentDAO.GetRootCommentsByCursor(ctx, noteID, cursor, limit)
	if err != nil {
		return nil, err
	}

	// 4. 判断是否还有更多
	hasMore := false
	displayCount := len(comments)
	if displayCount > pageSize {
		hasMore = true
		displayCount = pageSize
		comments = comments[:displayCount] // 只保留 pageSize 条
	}

	if len(comments) == 0 {
		return &types.CommentsListResponse{
			Comments:   make([]*types.CommentResponse, 0),
			NextCursor: 0,
			HasMore:    false,
		}, nil
	}

	// 5. 收集需要查询的ID
	commentIDs := make([]uint64, 0, displayCount)
	userIDs := make([]uint64, 0, displayCount)

	for i := 0; i < displayCount; i++ {
		commentIDs = append(commentIDs, comments[i].ID)
		userIDs = append(userIDs, comments[i].UserID)
	}

	// 6. 并发获取关联数据
	var (
		userMap       map[uint64]types.UserProfile
		likeStatusMap map[uint64]bool
		repliesMap    map[uint64][]*models.Comment
		wg            sync.WaitGroup
		mu            sync.Mutex
	)

	wg.Add(3)

	// 6.1 获取用户信息
	go func() {
		defer wg.Done()
		userMap = s.UserService.BatchGetUserInfo(ctx, userIDs)
	}()

	// 6.2 获取点赞状态
	go func() {
		defer wg.Done()
		if currentUserID > 0 {
			likes, err := s.CommentLikeDAO.BatchCheckExists(ctx, commentIDs, currentUserID)
			if err == nil {
				mu.Lock()
				likeStatusMap = likes
				mu.Unlock()
			}
		}
	}()

	// 6.3 批量获取最新3条回复
	go func() {
		defer wg.Done()
		replies, err := s.CommentDAO.BatchGetLatestReplies(ctx, commentIDs, 3)
		if err == nil {
			mu.Lock()
			repliesMap = replies
			mu.Unlock()
		}
	}()

	wg.Wait()

	// 7. 收集回复中的用户ID
	replyUserIDs := make([]uint64, 0)
	replyCommentIDs := make([]uint64, 0)

	for _, replies := range repliesMap {
		for _, reply := range replies {
			replyCommentIDs = append(replyCommentIDs, reply.ID)
			replyUserIDs = append(replyUserIDs, reply.UserID)
			if reply.ReplyToUserID > 0 {
				replyUserIDs = append(replyUserIDs, reply.ReplyToUserID)
			}
		}
	}

	// 7.1 获取回复用户信息
	if len(replyUserIDs) > 0 {
		replyUsers := s.UserService.BatchGetUserInfo(ctx, replyUserIDs)
		for k, v := range replyUsers {
			userMap[k] = v
		}
	}

	// 7.2 获取回复的点赞状态
	replyLikeStatusMap := make(map[uint64]bool)
	if currentUserID > 0 && len(replyCommentIDs) > 0 {
		replyLikes, _ := s.CommentLikeDAO.BatchCheckExists(ctx, replyCommentIDs, currentUserID)
		replyLikeStatusMap = replyLikes
	}

	// 8. 组装响应
	result := make([]*types.CommentResponse, 0, displayCount)

	for i := 0; i < displayCount; i++ {
		comment := comments[i]

		resp := &types.CommentResponse{
			ID:         comment.ID,
			NoteID:     comment.NoteID,
			UserID:     comment.UserID,
			Content:    comment.Content,
			LikeCount:  comment.LikeCount,
			ReplyCount: comment.ReplyCount,
			IPLocation: comment.IPLocation,
			IsLiked:    likeStatusMap[comment.ID],
			CreatedAt:  comment.CreatedAt,
			User:       userMap[comment.UserID],
		}

		// 添加最新回复
		if replies, ok := repliesMap[comment.ID]; ok {
			resp.LatestReplies = make([]*types.ReplyResponse, 0, len(replies))
			for _, reply := range replies {
				replyResp := &types.ReplyResponse{
					ID:         reply.ID,
					RootID:     reply.RootID,
					ParentID:   reply.ParentID,
					Content:    reply.Content,
					LikeCount:  reply.LikeCount,
					IsLiked:    replyLikeStatusMap[reply.ID],
					IPLocation: reply.IPLocation,
					CreatedAt:  reply.CreatedAt,
					User:       userMap[reply.UserID],
				}

				if reply.ReplyToUserID > 0 {
					replyResp.ReplyToUser = userMap[reply.ReplyToUserID]
				}

				resp.LatestReplies = append(resp.LatestReplies, replyResp)
			}
		} else {
			resp.LatestReplies = make([]*types.ReplyResponse, 0)
		}

		result = append(result, resp)
	}

	// 9. 计算下一个游标(取最后一条的时间戳)
	nextCursor := int64(0)
	if displayCount > 0 {
		nextCursor = comments[displayCount-1].CreatedAt.UnixNano()
	}

	return &types.CommentsListResponse{
		Comments:   result,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

// GetReplies 获取某条评论的所有回复(游标分页)
func (s *CommentsService) GetReplies(ctx context.Context, rootID uint64, cursor int64, pageSize int, currentUserID uint64) (*types.RepliesListResponse, error) {
	// 1. 参数校验
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	// 2. 多查一条
	limit := pageSize + 1

	// 3. 获取回复列表
	replies, err := s.CommentDAO.GetRepliesByCursor(ctx, rootID, cursor, limit)
	if err != nil {
		return nil, err
	}

	// 4. 判断是否还有更多
	hasMore := false
	displayCount := len(replies)
	if displayCount > pageSize {
		hasMore = true
		displayCount = pageSize
		replies = replies[:displayCount]
	}

	if len(replies) == 0 {
		return &types.RepliesListResponse{
			Replies:    make([]*types.ReplyResponse, 0),
			NextCursor: 0,
			HasMore:    false,
		}, nil
	}

	// 5. 收集用户ID
	userIDs := make([]uint64, 0, displayCount*2)
	replyIDs := make([]uint64, 0, displayCount)

	for i := 0; i < displayCount; i++ {
		replyIDs = append(replyIDs, replies[i].ID)
		userIDs = append(userIDs, replies[i].UserID)
		if replies[i].ReplyToUserID > 0 {
			userIDs = append(userIDs, replies[i].ReplyToUserID)
		}
	}

	// 6. 并发获取关联数据
	var (
		userMap       map[uint64]types.UserProfile
		likeStatusMap map[uint64]bool
		wg            sync.WaitGroup
		mu            sync.Mutex
	)

	wg.Add(2)

	go func() {
		defer wg.Done()
		userMap = s.UserService.BatchGetUserInfo(ctx, userIDs)
	}()

	go func() {
		defer wg.Done()
		if currentUserID > 0 {
			likes, err := s.CommentLikeDAO.BatchCheckExists(ctx, replyIDs, currentUserID)
			if err == nil {
				mu.Lock()
				likeStatusMap = likes
				mu.Unlock()
			}
		}
	}()

	wg.Wait()

	// 7. 组装响应
	result := make([]*types.ReplyResponse, 0, displayCount)

	for i := 0; i < displayCount; i++ {
		reply := replies[i]

		resp := &types.ReplyResponse{
			ID:         reply.ID,
			RootID:     reply.RootID,
			ParentID:   reply.ParentID,
			Content:    reply.Content,
			LikeCount:  reply.LikeCount,
			IsLiked:    likeStatusMap[reply.ID],
			IPLocation: reply.IPLocation,
			CreatedAt:  reply.CreatedAt,
			User:       userMap[reply.UserID],
		}

		if reply.ReplyToUserID > 0 {
			resp.ReplyToUser = userMap[reply.ReplyToUserID]
		}

		result = append(result, resp)
	}

	// 8. 计算下一个游标
	nextCursor := int64(0)
	if displayCount > 0 {
		nextCursor = replies[displayCount-1].CreatedAt.UnixNano()
	}

	return &types.RepliesListResponse{
		Replies:    result,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}
