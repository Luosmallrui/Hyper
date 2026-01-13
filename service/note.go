package service

import (
	"Hyper/dao"
	"Hyper/models"
	"Hyper/pkg/snowflake"
	"Hyper/types"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cloudwego/kitex/tool/internal_pkg/log"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"sync"
	"time"
)

var _ INoteService = (*NoteService)(nil)

type INoteService interface {
	CreateNote(ctx context.Context, userID uint64, req *types.CreateNoteRequest) (uint64, error)
	GetUserNotes(ctx context.Context, userID uint64, status int, limit, offset int) ([]*models.Note, error)
	UpdateNoteStatus(ctx context.Context, noteID uint64, status int) error
	ListNote(ctx context.Context, cursor int64, pageSize int, userID uint64) (types.ListNotesRep, error)
	GetNoteDetail(ctx context.Context, noteID uint64, currentUserID uint64) (*types.NoteDetail, error)
}
type NoteService struct {
	NoteDAO        *dao.NoteDAO
	UserService    IUserService
	LikeService    ILikeService
	RedisClient    *redis.Client
	StatsDAO       *dao.NoteStatsDAO
	FollowService  IFollowService
	CollectService ICollectService
}

func (s *NoteService) ListNote(ctx context.Context, cursor int64, pageSize int, userID uint64) (types.ListNotesRep, error) {
	limit := pageSize + 1
	nodes, err := s.NoteDAO.ListNode(ctx, cursor, limit)
	if err != nil {
		return types.ListNotesRep{}, err
	}

	rep := types.ListNotesRep{
		Notes:   make([]*types.Notes, 0),
		HasMore: false,
	}

	if len(nodes) == 0 {
		return rep, nil
	}

	displayCount := len(nodes)
	if displayCount > pageSize {
		rep.HasMore = true
		displayCount = pageSize
	}

	// 收集ID
	userIds := make([]uint64, 0, displayCount)
	noteIds := make([]uint64, 0, displayCount)
	for i := 0; i < displayCount; i++ {
		userIds = append(userIds, nodes[i].UserID)
		noteIds = append(noteIds, nodes[i].ID)
	}

	// 并发获取关联数据
	var (
		userMap       map[uint64]types.UserProfile
		statsMap      map[uint64]*types.NoteStats
		likeStatusMap map[uint64]bool
		wg            sync.WaitGroup
		mu            sync.Mutex
	)

	wg.Add(3)

	// 获取用户信息
	go func() {
		defer wg.Done()
		userMap = s.UserService.BatchGetUserInfo(ctx, userIds)
	}()

	// 获取统计数据
	go func() {
		defer wg.Done()
		stats, err := s.LikeService.BatchGetNoteStats(ctx, noteIds)
		mu.Lock()
		statsMap = stats
		mu.Unlock()
		if err != nil {
			log.Error("批量获取统计数据失败", "error", err)
		}
	}()

	// 获取点赞状态
	go func() {
		defer wg.Done()
		if userID > 0 {
			status, err := s.LikeService.BatchCheckLikeStatus(ctx, userID, noteIds)
			mu.Lock()
			likeStatusMap = status
			mu.Unlock()
			if err != nil {
				log.Error("批量获取点赞状态失败", "error", err)
			}
		}
	}()

	wg.Wait()

	// 组装数据
	for i := 0; i < displayCount; i++ {
		note := nodes[i]
		stats := statsMap[note.ID]
		if stats == nil {
			stats = &types.NoteStats{} // 默认值
		}

		dto := &types.Notes{
			ID:           int64(note.ID),
			UserID:       int64(note.UserID),
			Title:        note.Title,
			Content:      note.Content,
			Type:         note.Type,
			Status:       note.Status,
			VisibleConf:  note.VisibleConf,
			LikeCount:    stats.LikeCount,
			CollCount:    stats.CollCount,
			ShareCount:   stats.ShareCount,
			CommentCount: stats.CommentCount,
			ViewCount:    stats.ViewCount,
			IsLiked:      likeStatusMap[note.ID],
			CreatedAt:    note.CreatedAt,
			UpdatedAt:    note.UpdatedAt,
		}

		if user, ok := userMap[note.UserID]; ok {
			dto.Avatar = user.Avatar
			dto.Nickname = user.Nickname
		}

		// 处理其他字段
		if err := json.Unmarshal([]byte(note.TopicIDs), &dto.TopicIDs); err != nil {
			dto.TopicIDs = make([]int64, 0)
		}
		if err := json.Unmarshal([]byte(note.Location), &dto.Location); err != nil {
			dto.Location = types.Location{}
		}

		var noteMedia []types.NoteMedia
		if err := json.Unmarshal([]byte(note.MediaData), &noteMedia); err == nil && len(noteMedia) > 0 {
			dto.MediaData = noteMedia[0]
		} else {
			dto.MediaData = types.NoteMedia{}
		}

		rep.Notes = append(rep.Notes, dto)
	}

	if displayCount > 0 {
		rep.NextCursor = nodes[displayCount-1].CreatedAt.UnixNano()
	}

	return rep, nil
}

// CreateNote 创建笔记
func (s *NoteService) CreateNote(ctx context.Context, userID uint64, req *types.CreateNoteRequest) (uint64, error) {
	// 参数验证
	if req.Title == "" {
		return 0, errors.New("标题不能为空")
	}

	// 生成笔记ID
	noteID := uint64(snowflake.GenUserID())
	if len(req.TopicIDs) == 0 {
		req.TopicIDs = make([]int64, 0)
	}
	if len(req.MediaData) == 0 {
		req.MediaData = make([]types.NoteMedia, 0)
	}

	// 序列化 JSON 字段
	topicIDsJSON, err := json.Marshal(req.TopicIDs)
	if err != nil {
		return 0, err
	}

	// 修改这里: Location 为 nil 时使用 "{}" 或 "null"
	locationJSON := "{}" // 或者用 "null"
	if req.Location != nil {
		locBytes, err := json.Marshal(req.Location)
		if err != nil {
			return 0, err
		}
		locationJSON = string(locBytes)
	}

	mediaDataJSON, err := json.Marshal(req.MediaData)
	if err != nil {
		return 0, err
	}

	// 构建笔记对象
	note := &models.Note{
		ID:          noteID,
		UserID:      userID,
		Title:       req.Title,
		Content:     req.Content,
		TopicIDs:    string(topicIDsJSON),
		Location:    locationJSON,
		MediaData:   string(mediaDataJSON),
		Type:        req.Type,
		Status:      0,
		VisibleConf: req.VisibleConf,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if note.VisibleConf == 0 {
		note.VisibleConf = types.VisibleConfPublic
	}

	// 保存到数据库
	if err := s.NoteDAO.Create(ctx, note); err != nil {
		return 0, err
	}

	return noteID, nil
}

// GetUserNotes 获取用户的笔记列表
func (s *NoteService) GetUserNotes(ctx context.Context, userID uint64, status int, limit, offset int) ([]*models.Note, error) {
	return s.NoteDAO.FindByUserID(ctx, userID, status, limit, offset)
}

// UpdateNoteStatus 更新笔记状态
func (s *NoteService) UpdateNoteStatus(ctx context.Context, noteID uint64, status int) error {
	return s.NoteDAO.UpdateStatus(ctx, noteID, status)
}

// service/note_service.go

// 获取笔记详情
func (s *NoteService) GetNoteDetail(ctx context.Context, noteID uint64, currentUserID uint64) (*types.NoteDetail, error) {
	// 1. 查询笔记基本信息
	note, err := s.NoteDAO.GetByID(ctx, noteID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("笔记不存在")
		}
		return nil, err
	}

	// 2. 检查笔记可见性
	if err := s.checkNoteVisible(ctx, note, currentUserID); err != nil {
		return nil, err
	}

	// 3. 并发获取关联数据
	var (
		userInfo    *types.UserProfile
		stats       *types.NoteStats
		isLiked     bool
		isCollected bool
		isFollowed  bool
		wg          sync.WaitGroup
		mu          sync.Mutex
	)

	wg.Add(5)

	// 获取作者信息
	go func() {
		defer wg.Done()
		users := s.UserService.BatchGetUserInfo(ctx, []uint64{note.UserID})
		if user, ok := users[note.UserID]; ok {
			userInfo = &types.UserProfile{
				Avatar:   user.Avatar,
				Nickname: user.Nickname,
			}
		}
	}()

	// 获取统计数据
	go func() {
		defer wg.Done()
		statsMap, err := s.LikeService.BatchGetNoteStats(ctx, []uint64{noteID})
		if err == nil {
			if s, ok := statsMap[noteID]; ok {
				mu.Lock()
				stats = s
				mu.Unlock()
			}
		}
	}()

	// 获取点赞状态
	go func() {
		defer wg.Done()
		if currentUserID > 0 {
			liked, err := s.LikeService.checkLikeStatus(ctx, currentUserID, noteID)
			if err == nil {
				mu.Lock()
				isLiked = liked
				mu.Unlock()
			}
		}
	}()

	// 获取收藏状态
	go func() {
		defer wg.Done()
		if currentUserID > 0 {
			collected, err := s.CollectService.CheckCollectStatus(ctx, currentUserID, noteID)
			if err == nil {
				mu.Lock()
				isCollected = collected
				mu.Unlock()
			}
		}
	}()

	// 获取关注状态
	go func() {
		defer wg.Done()
		if currentUserID > 0 && currentUserID != note.UserID {
			followed, err := s.FollowService.CheckFollowStatus(ctx, currentUserID, note.UserID)
			if err == nil {
				mu.Lock()
				isFollowed = followed
				mu.Unlock()
			}
		}
	}()

	wg.Wait()

	// 4. 异步增加浏览数(不阻塞响应)
	go func() {
		_ = s.incrementViewCount(context.Background(), noteID)
	}()

	// 5. 组装返回数据
	detail := s.buildNoteDetail(note, userInfo, stats, isLiked, isCollected, isFollowed)

	return detail, nil
}

// 检查笔记可见性
func (s *NoteService) checkNoteVisible(ctx context.Context, note *models.Note, currentUserID uint64) error {
	// 审核状态检查
	if note.Status != 1 { // 1-审核通过
		// 只有作者本人可以看未通过审核的笔记
		if currentUserID != note.UserID {
			return errors.New("笔记审核中或已下架")
		}
	}

	// 可见性检查
	switch note.VisibleConf {
	case 1: // 公开
		return nil
	case 2: // 粉丝可见
		if currentUserID == 0 {
			return errors.New("请先登录")
		}
		if currentUserID == note.UserID {
			return nil
		}
		// 检查是否是粉丝
		isFollower, err := s.FollowService.CheckFollowStatus(ctx, currentUserID, note.UserID)
		if err != nil || !isFollower {
			return errors.New("仅粉丝可见")
		}
	case 3: // 自己可见
		if currentUserID != note.UserID {
			return errors.New("仅作者本人可见")
		}
	}

	return nil
}

// 组装笔记详情
func (s *NoteService) buildNoteDetail(note *models.Note, userInfo *types.UserProfile, stats *types.NoteStats, isLiked, isCollected, isFollowed bool) *types.NoteDetail {

	detail := &types.NoteDetail{
		ID:          int64(note.ID),
		UserID:      int64(note.UserID),
		Title:       note.Title,
		Content:     note.Content,
		Type:        note.Type,
		Status:      note.Status,
		VisibleConf: note.VisibleConf,
		CreatedAt:   note.CreatedAt,
		UpdatedAt:   note.UpdatedAt,
		Nickname:    userInfo.Nickname,
		Avatar:      userInfo.Avatar,
	}

	// 统计数据
	if stats != nil {
		detail.LikeCount = stats.LikeCount
		detail.CollCount = stats.CollCount
		detail.ShareCount = stats.ShareCount
		detail.CommentCount = stats.CommentCount
	}

	// 用户交互状态
	detail.IsLiked = isLiked
	detail.IsCollected = isCollected
	detail.IsFollowed = isFollowed

	// 解析 JSON 字段
	if err := json.Unmarshal([]byte(note.TopicIDs), &detail.TopicIDs); err != nil {
		detail.TopicIDs = make([]int64, 0)
	}

	if err := json.Unmarshal([]byte(note.Location), &detail.Location); err != nil {
		detail.Location = types.Location{}
	}

	if err := json.Unmarshal([]byte(note.MediaData), &detail.MediaData); err != nil {
		detail.MediaData = make([]types.NoteMedia, 0)
	}

	return detail
}

// 增加浏览数(异步)
func (s *NoteService) incrementViewCount(ctx context.Context, noteID uint64) error {
	// 先更新 Redis
	key := fmt.Sprintf("note:view:count:%d", noteID)
	s.RedisClient.Incr(ctx, key)
	s.RedisClient.Expire(ctx, key, 24*time.Hour)

	// 定期批量刷新到数据库(或者用消息队列)
	// 这里简化为直接更新
	return s.StatsDAO.IncrementViewCount(ctx, noteID, 1)
}
