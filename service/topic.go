package service

import (
	"Hyper/config"
	"Hyper/dao"
	"Hyper/models"
	"Hyper/types"
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"sync"
	"time"
)

const (
	maxSearchResults = 10 // 搜索结果的最大数量
	maxTopicNum      = 10 // 每个用户最多创建的话题数量

	// Redis 缓存相关常量
	hotTopicsRedisKey = "topics:hot:list" // 热门话题列表的 Redis Key
	topicsCacheTTL    = 24 * time.Hour    // 热门话题缓存过期时间：24小时
)

var _ ITopicService = (*TopicService)(nil)

type TopicService struct {
	Config      *config.Config
	DB          *gorm.DB
	TopicDAO    *dao.Topic
	NoteDAO     *dao.NoteDAO
	UserService IUserService
	LikeService ILikeService
	Redis       *redis.Client
}

type ITopicService interface {
	SearchTopics(ctx context.Context, query string) ([]types.CreateOrGetTopicResponse, error)
	CreateTopicIfNotExists(ctx context.Context, name string, creatorID uint64) (*types.CreateOrGetTopicResponse, error)
	GetNotesByTopic(ctx context.Context, topicID uint64, cursor int64, limit int, currentUserID uint64) (*types.TopicNotesResponse, error)
}

func (ts *TopicService) LoadOrCacheHotTopics(ctx context.Context, limit int) ([]*models.Topic, error) {
	// 尝试从 Redis 获取缓存
	cachedJSON, err := ts.Redis.Get(ctx, hotTopicsRedisKey).Result()
	if err == nil && cachedJSON != "" {
		// 缓存命中，反序列化返回
		var topics []*models.Topic
		if err := json.Unmarshal([]byte(cachedJSON), &topics); err == nil {
			return topics, nil
		}
	}

	// 缓存未命中，从数据库加载热门话题
	topics, err := ts.TopicDAO.GetHotTopics(ctx, limit)
	if err != nil {
		return nil, err
	}

	// 更新 Redis 缓存
	if len(topics) > 0 {
		if data, err := json.Marshal(topics); err == nil {
			ts.Redis.Set(ctx, hotTopicsRedisKey, data, topicsCacheTTL)
		}
	}

	return topics, nil
}

// SearchTopics 根据话题已有的字词进行搜索
func (ts *TopicService) SearchTopics(ctx context.Context, query string) ([]types.CreateOrGetTopicResponse, error) {
	limit := maxSearchResults
	var topics []*models.Topic
	var err error
	//空的就返回热门话题
	if query == "" {
		topics, err = ts.LoadOrCacheHotTopics(ctx, limit)
	} else {
		//否则就模糊搜索
		topics, err = ts.TopicDAO.FindTopicsByName(ctx, query, limit)
	}

	if err != nil {
		return nil, err
	}
	//转换为响应结构
	results := make([]types.CreateOrGetTopicResponse, 0)
	for _, topic := range topics {
		results = append(results, types.CreateOrGetTopicResponse{
			ID:        topic.ID,
			Name:      topic.Name,
			ViewCount: topic.ViewCount,
		})
	}
	return results, nil

}

// CreateTopicIfNotExists 创建话题（如果不存在的话）
func (ts *TopicService) CreateTopicIfNotExists(ctx context.Context, name string, creatorID uint64) (*types.CreateOrGetTopicResponse, error) {

	existing, err := ts.TopicDAO.FindTopicByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return &types.CreateOrGetTopicResponse{
			ID:        existing.ID,
			Name:      existing.Name,
			ViewCount: existing.ViewCount,
		}, nil
	}

	newTopic := &models.Topic{
		Name:        name,
		CreatorID:   creatorID,
		Status:      1,
		IsHot:       false,
		PostCount:   0,
		ViewCount:   0,
		FollowCount: 0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		LastPostAt:  time.Now(),
	}
	err = ts.TopicDAO.CreateTopic(ctx, newTopic)
	if err != nil {
		return nil, err
	}
	return &types.CreateOrGetTopicResponse{
		ID:        newTopic.ID,
		Name:      newTopic.Name,
		ViewCount: newTopic.ViewCount,
	}, nil
}

// 获取话题的笔记列表（参考 CommentsService.GetComments）
func (ts *TopicService) GetNotesByTopic(ctx context.Context, topicID uint64, cursor int64, limit int, currentUserID uint64) (*types.TopicNotesResponse, error) {
	// 1. 参数校验
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	// 2. 多查一条判断是否还有更多
	queryLimit := limit + 1

	// 3. 获取话题笔记列表（使用游标分页）
	notes, err := ts.TopicDAO.GetNotesByTopic(ctx, topicID, queryLimit, int(cursor))
	if err != nil {
		return nil, err
	}

	// 4. 判断是否还有更多
	hasMore := false
	displayCount := len(notes)
	if displayCount > limit {
		hasMore = true
		displayCount = limit
		notes = notes[:displayCount]
	}

	if len(notes) == 0 {
		return &types.TopicNotesResponse{
			Topic:      nil,
			Notes:      make([]*types.NoteWithStats, 0),
			HasMore:    false,
			NextCursor: 0,
		}, nil
	}

	// 5. 收集需要查询的ID
	noteIDs := make([]uint64, 0, displayCount)
	userIDs := make([]uint64, 0, displayCount)

	for i := 0; i < displayCount; i++ {
		noteIDs = append(noteIDs, notes[i].ID)
		userIDs = append(userIDs, notes[i].UserID)
	}

	// 6. 并发获取关联数据
	var (
		userMap       map[uint64]types.UserProfile
		likeStatusMap map[uint64]bool
		statsMap      map[uint64]*types.NoteStats
		topic         *models.Topic
		wg            sync.WaitGroup
		mu            sync.Mutex
	)

	wg.Add(4)

	// 6.1 获取用户信息
	go func() {
		defer wg.Done()
		userMap = ts.UserService.BatchGetUserInfo(ctx, userIDs)
	}()

	// 6.2 获取点赞状态（当前用户是否点赞过这些笔记）
	go func() {
		defer wg.Done()
		if currentUserID > 0 {
			likes, err := ts.LikeService.BatchCheckLikeStatus(ctx, currentUserID, noteIDs)
			if err == nil {
				mu.Lock()
				likeStatusMap = likes
				mu.Unlock()
			}
		}
	}()

	// 6.3 获取笔记统计信息（点赞数、评论数等）
	go func() {
		defer wg.Done()
		stats, err := ts.LikeService.BatchGetNoteStats(ctx, noteIDs)
		if err == nil {
			mu.Lock()
			statsMap = stats
			mu.Unlock()
		}
	}()

	// 6.4 获取话题信息
	go func() {
		defer wg.Done()
		var err error
		topic, err = ts.TopicDAO.FindById(ctx, topicID)
		if err != nil {
			// 记录日志但不影响其他数据获取
		}
	}()

	wg.Wait()

	// 7. 初始化 nil maps，避免 panic
	if likeStatusMap == nil {
		likeStatusMap = make(map[uint64]bool)
	}
	if statsMap == nil {
		statsMap = make(map[uint64]*types.NoteStats)
	}
	if userMap == nil {
		userMap = make(map[uint64]types.UserProfile)
	}

	// 8. 组装响应
	result := make([]*types.NoteWithStats, 0, displayCount)

	for i := 0; i < displayCount; i++ {
		note := notes[i]
		stat := statsMap[note.ID]

		resp := &types.NoteWithStats{
			ID:           int64(note.ID),
			UserID:       int64(note.UserID),
			Title:        note.Title,
			Content:      note.Content,
			Type:         note.Type,
			Nickname:     userMap[note.UserID].Nickname,
			Avatar:       userMap[note.UserID].Avatar,
			LikeCount:    0,
			CommentCount: 0,
			ViewCount:    0,
			CreatedAt:    note.CreatedAt.Unix(),
		}

		// 从统计数据中获取各项计数
		if stat != nil {
			resp.LikeCount = stat.LikeCount
			resp.CommentCount = stat.CommentCount
			resp.ViewCount = stat.ViewCount
		}

		result = append(result, resp)
	}

	// 9. 组装话题信息
	var topicInfo *types.TopicInfo
	if topic != nil {
		topicInfo = &types.TopicInfo{
			ID:          topic.ID,
			Name:        topic.Name,
			PostCount:   topic.PostCount,
			ViewCount:   topic.ViewCount,
			FollowCount: topic.FollowCount,
			IsHot:       topic.IsHot,
		}
	}

	// 10. 计算下一个游标
	nextCursor := int64(0)
	if displayCount > 0 {
		nextCursor = notes[displayCount-1].CreatedAt.UnixNano()
	}

	return &types.TopicNotesResponse{
		Topic:      topicInfo,
		Notes:      result,
		HasMore:    hasMore,
		NextCursor: nextCursor,
	}, nil
}
