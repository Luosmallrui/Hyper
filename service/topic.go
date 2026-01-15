package service

import (
	"Hyper/config"
	"Hyper/dao"
	"Hyper/models"
	"Hyper/pkg/snowflake"
	"Hyper/types"
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/cloudwego/kitex/tool/internal_pkg/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"regexp"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	maxTopicNum = 10 // 每个用户最多创建的话题数量
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
	ExtractTopicsFromContent(content string) ([]string, error)
	BatchCreateTopics(ctx context.Context, topicNames []string, userID uint64) (map[string]*models.Topic, error)
	CreateNewTopic(ctx context.Context, req *types.CreateTopicRequest, userID uint64) (*types.CreateTopicResponse, error)
	ExtractAndAssociateTopics(ctx context.Context, noteID uint64, content string, userID uint64) ([]*models.Topic, error)
	GetTopicNotes(ctx context.Context, topicID uint64, cursor int64, pageSize int) (*types.TopicNotesResponse, error)
}

func (ts *TopicService) CreateNewTopic(ctx context.Context, req *types.CreateTopicRequest, userID uint64) (*types.CreateTopicResponse, error) {
	if err := ts.validateTopicName(req.Name); err != nil {
		return nil, err
	}
	normalizeName := ts.normalizeTopicName(req.Name)

	extistopic, err := ts.TopicDAO.FindTopicByName(ctx, normalizeName)
	if err == nil && extistopic != nil {
		// 话题已存在，直接返回已存在的话题ID
		return &types.CreateTopicResponse{
			ID: extistopic.ID,
		}, nil
	}
	topicID := uint64(snowflake.GenID())

	now := time.Now()
	topic := &models.Topic{
		ID:          topicID,
		Name:        normalizeName,
		Description: req.Description,
		CoverURL:    req.CoverURL,
		CreatorID:   userID,
		CategoryID:  req.CategoryID,
		PostCount:   0,
		ViewCount:   0,
		FollowCount: 0,
		IsHot:       false,
		Status:      1,
		CreatedAt:   now,
		UpdatedAt:   now,
		LastPostAt:  now,
	}

	// 使用 ON DUPLICATE KEY UPDATE（实际上是为了防止并发创建同名话题，不太懂还有什么方式可以处理）
	result := ts.DB.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true, // 更新所有字段
	}).Create(topic)

	if result.Error != nil {
		return nil, result.Error
	}

	return &types.CreateTopicResponse{ID: topic.ID}, nil
}

func (ts *TopicService) BatchCreateTopics(ctx context.Context, topicNames []string, userID uint64) (map[string]*models.Topic, error) {
	//用户也可能不关联话题
	if len(topicNames) == 0 {
		return map[string]*models.Topic{}, nil
	}

	// ============ 1. 规范化所有话题名称 ============
	normalizedNames := make([]string, 0, len(topicNames))
	seen := make(map[string]bool)

	for _, name := range topicNames {
		normalized := ts.normalizeTopicName(name)
		if normalized != "" && !seen[normalized] {
			// 验证话题名称
			if err := ts.validateTopicName(normalized); err != nil {
				return nil, fmt.Errorf("话题名称 '%s' 无效: %w", name, err)
			}
			seen[normalized] = true
			normalizedNames = append(normalizedNames, normalized)
		}
	}

	existingTopics, err := ts.TopicDAO.BatchTopicFindByNames(ctx, normalizedNames)
	if err != nil {
		return nil, fmt.Errorf("批量查询话题失败: %w", err)
	}

	toCreate := make([]*models.Topic, 0)
	toUpdate := make([]uint64, 0)
	now := time.Now()

	for _, name := range normalizedNames {
		// 如果已经存在，记录
		if topic, exists := existingTopics[name]; exists {
			toUpdate = append(toUpdate, topic.ID)
			continue
		}

		topicID := uint64(snowflake.GenID())
		//不存在的时候才创建
		topic := &models.Topic{
			ID:          topicID,
			Name:        name,
			Description: "",
			CoverURL:    "",
			CreatorID:   userID,
			CategoryID:  0,
			PostCount:   0,
			ViewCount:   0,
			FollowCount: 0,
			IsHot:       false,
			SortWeight:  0,
			Status:      1,
			CreatedAt:   now,
			UpdatedAt:   now,
			LastPostAt:  now,
		}
		toCreate = append(toCreate, topic)
	}

	// ============ 4. 批量创建新话题 ============
	if len(toCreate) > 0 {
		// ✅ 使用 OnConflict 处理并发冲突，不返回错误(不太懂这个方法，有无其他代替）
		result := ts.DB.WithContext(ctx).Clauses(clause.OnConflict{
			DoNothing: true, // 冲突时什么都不做
		}).CreateInBatches(toCreate, 100)

		if result.Error != nil {
			return nil, fmt.Errorf("批量创建话题失败: %w", result.Error)
		}

		// 创建成功，添加到结果中
		for _, topic := range toCreate {
			existingTopics[topic.Name] = topic
		}
	}

	if len(toUpdate) > 0 {
		// 批量更新已存在话题的 UpdatedAt 字段
		err = ts.TopicDAO.BatchUpdateTopicsUpdatedAt(ctx, toUpdate, 1)
		if err != nil {
			return nil, fmt.Errorf("批量更新话题失败: %w", err)
		}
	}

	// ============ 5. 返回所有话题（已存在+新创建） ============
	return existingTopics, nil
}

func (ts *TopicService) ExtractAndAssociateTopics(ctx context.Context, noteID uint64, content string, userID uint64) ([]*models.Topic, error) {
	topicNames, err := ts.ExtractTopicsFromContent(content)
	if err != nil {
		return nil, errors.New("提取话题失败: %w")
	}
	if len(topicNames) == 0 {
		return []*models.Topic{}, nil
	}

	topicMap, err := ts.BatchCreateTopics(ctx, topicNames, userID)
	if err != nil {
		return nil, errors.New("批量创建话题失败: %w")
	}

	// 提取话题ID列表
	topicIDs := make([]uint64, 0, len(topicMap))
	topics := make([]*models.Topic, 0, len(topicMap))
	for _, topic := range topicMap {
		topicIDs = append(topicIDs, topic.ID)
		topics = append(topics, topic)
	}
	if err := ts.TopicDAO.BatchCreateNoteTopicAssociations(ctx, noteID, topicIDs); err != nil {
		return nil, errors.New("批量创建笔记与话题关联失败: %w")
	}
	return topics, nil
}

func (ts *TopicService) GetTopicNotes(ctx context.Context, topicID uint64, cursor int64, pageSize int) (*types.TopicNotesResponse, error) {
	// 1. 查询话题信息
	topic, err := ts.TopicDAO.FindTopicByID(ctx, topicID)
	if err != nil {
		return nil, fmt.Errorf("话题不存在: %w", err)
	}
	// 2. 按 Cursor 查询该话题下的笔记ID
	noteIDs, nextCursor, hasMore, err := ts.TopicDAO.GetNoteIDsByTopicWithCursor(ctx, topicID, cursor, pageSize)
	if err != nil {
		return nil, fmt.Errorf("查询话题笔记失败: %w", err)
	}

	// 3. 如果没有笔记，直接返回空列表
	if len(noteIDs) == 0 {
		return &types.TopicNotesResponse{
			Topic: &types.TopicInfo{
				ID:          topic.ID,
				Name:        topic.Name,
				PostCount:   topic.PostCount,
				ViewCount:   topic.ViewCount,
				FollowCount: topic.FollowCount,
				IsHot:       topic.IsHot,
			},
			Notes:      make([]*types.NoteWithStats, 0),
			HasMore:    false,
			NextCursor: 0,
		}, nil
	}

	// 4. 并发获取关联数据
	var (
		noteModels []*models.Note
		statsMap   map[uint64]*types.NoteStats
		userMap    map[uint64]types.UserProfile
		wg         sync.WaitGroup
		mu         sync.Mutex
	)

	wg.Add(3)

	// 获取笔记详情
	go func() {
		defer wg.Done()
		notes, err := ts.NoteDAO.FindByIDs(ctx, noteIDs)
		if err != nil {
			log.Error("批量获取笔记失败: %v", err)
		} else {
			mu.Lock()
			noteModels = notes
			mu.Unlock()
		}
	}()
	// 获取统计数据
	go func() {
		defer wg.Done()
		stats, err := ts.LikeService.BatchGetNoteStats(ctx, noteIDs)
		if err != nil {
			log.Error("查询统计失败", "error", err)
		} else {
			mu.Lock()
			statsMap = stats
			mu.Unlock()
		}
	}()
	// 4.3 获取用户信息
	go func() {
		defer wg.Done()
		userIDs := make([]uint64, 0)
		for _, note := range noteModels {
			userIDs = append(userIDs, note.UserID)
		}
		if len(userIDs) > 0 {
			users := ts.UserService.BatchGetUserInfo(ctx, userIDs)
			if users != nil {
				mu.Lock()
				userMap = users
				mu.Unlock()
			}
		}
	}()
	wg.Wait()
	// 5. 组装返回结果
	notes := make([]*types.NoteWithStats, 0, len(noteModels))
	for _, note := range noteModels {
		stats := statsMap[note.ID]
		if stats == nil {
			stats = &types.NoteStats{}
		}

		user := userMap[note.UserID]
		notes = append(notes, &types.NoteWithStats{
			ID:           int64(note.ID),
			UserID:       int64(note.UserID),
			Title:        note.Title,
			Content:      note.Content,
			Type:         note.Type,
			Nickname:     user.Nickname,
			Avatar:       user.Avatar,
			LikeCount:    stats.LikeCount,
			CommentCount: stats.CommentCount,
			ViewCount:    stats.ViewCount,
			CreatedAt:    note.CreatedAt.Unix(),
		})
	}

	return &types.TopicNotesResponse{
		Topic: &types.TopicInfo{
			ID:          topic.ID,
			Name:        topic.Name,
			PostCount:   topic.PostCount,
			ViewCount:   topic.ViewCount,
			FollowCount: topic.FollowCount,
			IsHot:       topic.IsHot,
		},
		Notes:      notes,
		HasMore:    hasMore,
		NextCursor: nextCursor,
	}, nil
}

// 支持格式：#旅行 #美食 等
// 返回：去重后的话题名称列表
func (ts *TopicService) ExtractTopicsFromContent(content string) ([]string, error) {
	// 正则表达式：匹配 #开头，后跟1-20个汉字、英文、数字、下划线、连字符
	re := regexp.MustCompile(`#([\p{Han}a-zA-Z0-9_-]{1,20})`)
	matches := re.FindAllStringSubmatch(content, -1)

	seen := make(map[string]bool)
	var topics []string

	for _, match := range matches {
		if len(match) > 1 {
			topic := strings.TrimSpace(match[1])
			// 如果不为空且未见过，添加到列表
			if topic != "" && !seen[topic] {
				seen[topic] = true
				topics = append(topics, topic)
			}
		}
	}
	return topics, nil
}
func (ts *TopicService) validateTopicName(name string) error {
	name = strings.TrimSpace(name)

	if len(name) == 0 || len(name) > 20 {
		return errors.New("话题不能为空或者超过20个字符")
	}

	validPattern := regexp.MustCompile(`^[\p{Han}a-zA-Z0-9_-]+$`)
	if !validPattern.MatchString(name) {
		return errors.New("话题名称只能包含中文、英文、数字、下划线和连字符")
	}
	return nil
}

func (ts *TopicService) normalizeTopicName(name string) string {
	// 去除前后空格和特殊字符
	name = strings.TrimSpace(name)
	// 去除前缀的特殊字符（如#）
	name = strings.TrimPrefix(name, "#")
	return name
}
