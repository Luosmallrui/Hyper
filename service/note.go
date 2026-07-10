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
	"sync"
	"time"

	"github.com/cloudwego/kitex/tool/internal_pkg/log"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var _ INoteService = (*NoteService)(nil)

type INoteService interface {
	CreateNote(ctx context.Context, userID uint64, req *types.CreateNoteRequest) (uint64, error)
	GetUserNotes(ctx context.Context, userID uint64, status int, limit, offset int) ([]*models.Note, error)
	UpdateNoteStatus(ctx context.Context, noteID uint64, status int) error
	UpdateNoteRelation(ctx context.Context, userID uint64, noteID uint64, req types.UpdateNoteRelationRequest) error
	ListNote(ctx context.Context, cursor int64, pageSize int, userID uint64) (types.ListNotesRep, error)
	GetNoteDetail(ctx context.Context, noteID uint64, currentUserID uint64) (*types.NoteDetail, error)
	GetMyNotesFeed(ctx context.Context, userID int, cursor int64, pageSize int) ([]*models.Note, int64, bool, error)
	ListNoteByUser(ctx context.Context, cursor int64, pageSize int, userID int, TargetUser int) (types.ListNotesBriefRep, error)
	GetFollowedPosts(ctx context.Context, userId int, cursor int64, pageSize int) (types.ListNotesRep, error)
	GetALlNote(ctx context.Context) ([]*models.Note, error)
	GetNoteByChannelID(ctx context.Context, userId int, cursor int64, pageSize int, channelId int) (types.ListNotesRep, error)
	GetRelatedNotes(ctx context.Context, req types.ListRelatedNotesReq, currentUserID uint64) (types.ListNotesRep, error)
	RecordShare(ctx context.Context, userID, noteID uint64, channel string) error
}

func (s *NoteService) RecordShare(ctx context.Context, userID, noteID uint64, channel string) error {
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var note models.Note
		if err := tx.Where("id = ?", noteID).First(&note).Error; err != nil {
			return err
		}
		if note.Status == -1 {
			return errors.New("笔记已删除")
		}
		if err := tx.Create(&models.NoteShare{NoteID: noteID, UserID: userID, Channel: channel}).Error; err != nil {
			return err
		}
		return tx.Exec(
			"INSERT INTO note_stats (note_id, share_count, updated_at) VALUES (?, 1, NOW()) "+
				"ON DUPLICATE KEY UPDATE share_count = share_count + 1, updated_at = NOW()",
			noteID,
		).Error
	})
}

type NoteService struct {
	NoteDAO         *dao.NoteDAO
	CommentDAO      *dao.Comment
	UserService     IUserService
	LikeService     ILikeService
	RedisClient     *redis.Client
	StatsDAO        *dao.NoteStatsDAO
	FollowService   IFollowService
	MerchantService IMerchantService
	CollectService  ICollectService
	CommentService  ICommentsService
	TopicService    ITopicService
	DB              *gorm.DB
}

// ============================================================================
// 通用数据加载器 — 消除所有列表方法中的重复并发逻辑
// ============================================================================

// noteEnrichment 包含批量加载的关联数据
type noteEnrichment struct {
	Users      map[uint64]types.UserProfile
	Stats      map[uint64]*types.NoteStats
	LikeStatus map[uint64]bool
}

// enrichNotes 并发批量加载用户信息、统计数据、点赞状态
func (s *NoteService) enrichNotes(ctx context.Context, userIDs, noteIDs []uint64, currentUserID uint64) *noteEnrichment {
	result := &noteEnrichment{
		Users:      make(map[uint64]types.UserProfile),
		Stats:      make(map[uint64]*types.NoteStats),
		LikeStatus: make(map[uint64]bool),
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	wg.Add(3)

	go func() {
		defer wg.Done()
		users := s.UserService.BatchGetUserInfo(ctx, userIDs)
		mu.Lock()
		result.Users = users
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		stats, err := s.LikeService.BatchGetNoteStats(ctx, noteIDs)
		if err != nil {
			log.Error("批量获取统计数据失败", "error", err)
			return
		}
		mu.Lock()
		result.Stats = stats
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		if currentUserID == 0 {
			return
		}
		status, err := s.LikeService.BatchCheckLikeStatus(ctx, currentUserID, noteIDs)
		if err != nil {
			log.Error("批量获取点赞状态失败", "error", err)
			return
		}
		mu.Lock()
		result.LikeStatus = status
		mu.Unlock()
	}()

	wg.Wait()
	return result
}

// getStatsOrDefault 安全地获取统计数据，不存在时返回零值
func (e *noteEnrichment) getStats(noteID uint64) *types.NoteStats {
	if s, ok := e.Stats[noteID]; ok && s != nil {
		return s
	}
	return &types.NoteStats{}
}

// ============================================================================
// 通用 DTO 转换 — 消除重复的 JSON 解析和字段映射
// ============================================================================
func (s *NoteService) getTopicsByIDs(ctx context.Context, ids []int64) ([]types.Topic, error) {
	if len(ids) == 0 {
		return []types.Topic{}, nil
	}

	topics := make([]types.Topic, 0, len(ids))
	missIDs := make([]int64, 0)

	// 先批量查缓存
	for _, id := range ids {
		key := fmt.Sprintf("topic:%d", id)
		val, err := s.RedisClient.Get(ctx, key).Result()
		if err == nil {
			var t types.Topic
			if json.Unmarshal([]byte(val), &t) == nil {
				topics = append(topics, t)
				continue
			}
		}
		missIDs = append(missIDs, id)
	}

	// 缓存未命中的回源数据库
	if len(missIDs) > 0 {
		var dbTopics []types.Topic
		err := s.DB.WithContext(ctx).
			Where("id IN ?", missIDs).
			Find(&dbTopics).Error
		if err != nil {
			return nil, err
		}

		// 回写缓存
		for _, t := range dbTopics {
			data, _ := json.Marshal(t)
			key := fmt.Sprintf("topic:%d", t.ID)
			s.RedisClient.Set(ctx, key, string(data), 30*time.Minute)
		}
		topics = append(topics, dbTopics...)
	}
	return topics, nil
}

// noteToDTO 将 models.Note 转为 types.Notes，附带关联数据
func (s *NoteService) noteToDTO(ctx context.Context, note *models.Note, enrich *noteEnrichment, currentUserID int) *types.Notes {
	stats := enrich.getStats(note.ID)

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
		IsLiked:      enrich.LikeStatus[note.ID],
		CreatedAt:    note.CreatedAt,
		UpdatedAt:    note.UpdatedAt,
		ActivityID:   note.ActivityID,
		StoreID:      note.StoreID,
	}

	if user, ok := enrich.Users[note.UserID]; ok {
		dto.Avatar = user.Avatar
		dto.Nickname = user.Nickname
	}

	var topicIDs []int64
	_ = jsonUnmarshalOr(note.TopicIDs, &topicIDs, make([]int64, 0))
	topics, _ := s.getTopicsByIDs(ctx, topicIDs)
	_ = jsonUnmarshalOr(note.Location, &dto.Location, types.Location{})
	dto.TopicIDs = topics
	dto.MediaData = parseFirstMedia(note.MediaData)
	return dto
}

// noteToBrief 将 models.Note 转为 types.NoteBrief
func noteToBrief(note *models.Note, enrich *noteEnrichment) *types.NoteBrief {
	stats := enrich.getStats(note.ID)

	dto := &types.NoteBrief{
		ID:           int64(note.ID),
		UserID:       int64(note.UserID),
		Title:        note.Title,
		Type:         note.Type,
		LikeCount:    stats.LikeCount,
		CollCount:    stats.CollCount,
		ShareCount:   stats.ShareCount,
		CommentCount: stats.CommentCount,
		ViewCount:    stats.ViewCount,
		IsLiked:      enrich.LikeStatus[note.ID],
		CreatedAt:    note.CreatedAt,
		UpdatedAt:    note.UpdatedAt,
		MediaData:    parseFirstMedia(note.MediaData),
	}

	return dto
}

// ============================================================================
// JSON 工具函数
// ============================================================================

// jsonUnmarshalOr 尝试解析 JSON，失败时设置默认值
func jsonUnmarshalOr[T any](raw string, dest *T, fallback T) error {
	if err := json.Unmarshal([]byte(raw), dest); err != nil {
		*dest = fallback
		return err
	}
	return nil
}

// parseFirstMedia 从 MediaData JSON 数组中提取第一个元素
func parseFirstMedia(raw string) types.NoteMedia {
	var media []types.NoteMedia
	if err := json.Unmarshal([]byte(raw), &media); err == nil && len(media) > 0 {
		return media[0]
	}
	return types.NoteMedia{}
}

// ============================================================================
// 通用分页处理
// ============================================================================

// collectIDs 从笔记列表中收集去重的 userID 和 noteID
func collectIDs(notes []*models.Note, count int) (userIDs, noteIDs []uint64) {
	userSet := make(map[uint64]struct{}, count)
	noteSet := make(map[uint64]struct{}, count)
	userIDs = make([]uint64, 0, count)
	noteIDs = make([]uint64, 0, count)

	for i := 0; i < count; i++ {
		if _, dup := userSet[notes[i].UserID]; !dup {
			userSet[notes[i].UserID] = struct{}{}
			userIDs = append(userIDs, notes[i].UserID)
		}
		if _, dup := noteSet[notes[i].ID]; !dup {
			noteSet[notes[i].ID] = struct{}{}
			noteIDs = append(noteIDs, notes[i].ID)
		}
	}
	return
}

// paginateNotes 对查询到的 notes 做分页截断，返回实际展示数量和是否有更多
func paginateNotes(notes []*models.Note, pageSize int) (displayCount int, hasMore bool) {
	if len(notes) > pageSize {
		return pageSize, true
	}
	return len(notes), false
}

// buildListNotesRep 通用列表构建：分页 + 加载关联数据 + 组装 DTO
func (s *NoteService) buildListNotesRep(ctx context.Context, notes []*models.Note, pageSize int, currentUserID uint64) types.ListNotesRep {
	rep := types.ListNotesRep{Notes: make([]*types.Notes, 0)}

	if len(notes) == 0 {
		return rep
	}

	displayCount, hasMore := paginateNotes(notes, pageSize)
	rep.HasMore = hasMore

	userIDs, noteIDs := collectIDs(notes, displayCount)
	enrich := s.enrichNotes(ctx, userIDs, noteIDs, currentUserID)

	for i := 0; i < displayCount; i++ {
		rep.Notes = append(rep.Notes, s.noteToDTO(ctx, notes[i], enrich, int(currentUserID)))
	}

	if displayCount > 0 {
		rep.NextCursor = notes[displayCount-1].CreatedAt.UnixNano()
	}

	return rep
}

func (s *NoteService) GetALlNote(ctx context.Context) ([]*models.Note, error) {
	return s.NoteDAO.ListAllNote(ctx)
}

func (s *NoteService) GetNoteByChannelID(ctx context.Context, userId int, cursor int64, pageSize int, channelId int) (types.ListNotesRep, error) {
	notes, err := s.NoteDAO.ListNodeByChannel(ctx, cursor, pageSize+1, channelId)
	if err != nil {
		return types.ListNotesRep{}, err
	}
	return s.buildListNotesRep(ctx, notes, pageSize, uint64(userId)), nil
}

func (s *NoteService) GetFollowedPosts(ctx context.Context, userId int, cursor int64, pageSize int) (types.ListNotesRep, error) {
	followingIDs, err := s.FollowService.GetFollowingIDs(ctx, userId)
	if err != nil {
		return types.ListNotesRep{}, err
	}
	if len(followingIDs) == 0 {
		return types.ListNotesRep{Notes: make([]*types.Notes, 0)}, nil
	}

	notes, err := s.NoteDAO.ListNodeByUserIDs(ctx, followingIDs, cursor, pageSize+1)
	if err != nil {
		return types.ListNotesRep{}, err
	}
	return s.buildListNotesRep(ctx, notes, pageSize, uint64(userId)), nil
}

func (s *NoteService) ListNote(ctx context.Context, cursor int64, pageSize int, userID uint64) (types.ListNotesRep, error) {
	notes, err := s.NoteDAO.ListNode(ctx, cursor, pageSize+1)
	if err != nil {
		return types.ListNotesRep{}, err
	}
	return s.buildListNotesRep(ctx, notes, pageSize, userID), nil
}

func (s *NoteService) GetRelatedNotes(ctx context.Context, req types.ListRelatedNotesReq, currentUserID uint64) (types.ListNotesRep, error) {
	if req.PageSize <= 0 {
		req.PageSize = types.DefaultPageSize
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}
	if req.StoreID <= 0 && req.ActivityID <= 0 {
		return types.ListNotesRep{Notes: make([]*types.Notes, 0)}, nil
	}
	notes, err := s.NoteDAO.ListRelatedNotes(ctx, req.StoreID, req.ActivityID, req.Cursor, req.PageSize+1)
	if err != nil {
		return types.ListNotesRep{}, err
	}
	return s.buildListNotesRep(ctx, notes, req.PageSize, currentUserID), nil
}

func (s *NoteService) ListNoteByUser(ctx context.Context, cursor int64, pageSize int, userID int, targetUser int) (types.ListNotesBriefRep, error) {
	notes, err := s.NoteDAO.ListNodeByUser(ctx, cursor, pageSize+1, targetUser)
	if err != nil {
		return types.ListNotesBriefRep{}, err
	}

	rep := types.ListNotesBriefRep{Notes: make([]*types.NoteBrief, 0)}

	if len(notes) == 0 {
		return rep, nil
	}

	displayCount, hasMore := paginateNotes(notes, pageSize)
	rep.HasMore = hasMore

	_, noteIDs := collectIDs(notes, displayCount)
	// ListNoteByUser 不需要用户信息，传空 userIDs 即可
	enrich := s.enrichNotes(ctx, nil, noteIDs, uint64(userID))

	for i := 0; i < displayCount; i++ {
		rep.Notes = append(rep.Notes, noteToBrief(notes[i], enrich))
	}

	if displayCount > 0 {
		rep.NextCursor = notes[displayCount-1].CreatedAt.UnixNano()
	}

	return rep, nil
}

func (s *NoteService) GetMyNotesFeed(ctx context.Context, userID int, cursor int64, pageSize int) ([]*models.Note, int64, bool, error) {
	query := s.DB.WithContext(ctx).
		Model(&models.Note{}).
		Where("user_id = ? AND status = 0", userID).
		Order("created_at DESC")

	if cursor > 0 {
		query = query.Where("created_at < ?", time.Unix(cursor, 0))
	}

	var notes []*models.Note
	if err := query.Limit(pageSize + 1).Find(&notes).Error; err != nil {
		return nil, 0, false, err
	}

	hasMore := len(notes) > pageSize
	if hasMore {
		notes = notes[:pageSize]
	}

	var nextCursor int64
	if hasMore && len(notes) > 0 {
		nextCursor = notes[len(notes)-1].CreatedAt.Unix()
	}

	return notes, nextCursor, hasMore, nil
}

// ============================================================================
// 创建笔记
// ============================================================================

func (s *NoteService) CreateNote(ctx context.Context, userID uint64, req *types.CreateNoteRequest) (uint64, error) {
	if req.Title == "" {
		return 0, errors.New("标题不能为空")
	}

	noteID := uint64(snowflake.GenUserID())

	// 规范化空集合，避免存 null
	if req.TopicIDs == nil {
		req.TopicIDs = make([]int64, 0)
	}
	if req.MediaData == nil {
		req.MediaData = make([]types.NoteMedia, 0)
	}

	topicJSON, _ := json.Marshal(req.TopicIDs)
	mediaJSON, _ := json.Marshal(req.MediaData)

	locationJSON := "{}"
	if req.Location != nil {
		if loc, err := json.Marshal(req.Location); err == nil {
			locationJSON = string(loc)
		}
	}

	now := time.Now()
	note := &models.Note{
		ID:          noteID,
		UserID:      userID,
		Title:       req.Title,
		Content:     req.Content,
		TopicIDs:    string(topicJSON),
		Location:    locationJSON,
		MediaData:   string(mediaJSON),
		Type:        req.Type,
		Status:      0,
		VisibleConf: cond(req.VisibleConf != 0, req.VisibleConf, types.VisibleConfPublic),
		CreatedAt:   now,
		UpdatedAt:   now,
		ActivityID:  req.ActivityID,
		StoreID:     req.StoreID,
	}

	err := s.NoteDAO.Transaction(ctx, func(tx *gorm.DB) error {
		if err := tx.Create(note).Error; err != nil {
			return err
		}

		stats := &models.NoteStats{
			NoteID:    noteID,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := tx.Create(stats).Error; err != nil {
			return err
		}

		if len(req.TopicIDs) > 0 {
			noteTopics := make([]*models.NoteTopic, 0, len(req.TopicIDs))
			for _, tid := range req.TopicIDs {
				noteTopics = append(noteTopics, &models.NoteTopic{
					NoteID:    noteID,
					TopicID:   uint64(tid),
					CreatedAt: now,
				})
			}
			return tx.CreateInBatches(noteTopics, 100).Error
		}
		return nil
	})

	if err != nil {
		return 0, err
	}
	return noteID, nil
}

// ============================================================================
// 获取用户笔记 / 更新状态
// ============================================================================

func (s *NoteService) GetUserNotes(ctx context.Context, userID uint64, status int, limit, offset int) ([]*models.Note, error) {
	return s.NoteDAO.FindByUserID(ctx, userID, status, limit, offset)
}

func (s *NoteService) UpdateNoteStatus(ctx context.Context, noteID uint64, status int) error {
	return s.NoteDAO.UpdateStatus(ctx, noteID, status)
}

func (s *NoteService) UpdateNoteRelation(ctx context.Context, userID uint64, noteID uint64, req types.UpdateNoteRelationRequest) error {
	updates := make(map[string]any)
	if req.StoreID != nil {
		if *req.StoreID < 0 {
			return errors.New("store_id不能小于0")
		}
		updates["store_id"] = *req.StoreID
	}
	if req.ActivityID != nil {
		if *req.ActivityID < 0 {
			return errors.New("activity_id不能小于0")
		}
		updates["activity_id"] = *req.ActivityID
	}
	if len(updates) == 0 {
		return errors.New("请至少提交store_id或activity_id")
	}
	updates["updated_at"] = time.Now()

	result := s.NoteDAO.Db.WithContext(ctx).
		Model(&models.Note{}).
		Where("id = ? AND user_id = ? AND status <> ?", noteID, userID, -1).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("动态不存在或无权限修改")
	}
	return nil
}

// ============================================================================
// 笔记详情
// ============================================================================

func (s *NoteService) GetNoteDetail(ctx context.Context, noteID uint64, currentUserID uint64) (*types.NoteDetail, error) {
	note, err := s.NoteDAO.GetByID(ctx, noteID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("笔记不存在")
		}
		return nil, err
	}

	var (
		userInfo       *types.UserProfile
		stats          *types.NoteStats
		commentPreview *types.CommentPreview
		isLiked        bool
		isCollected    bool
		isFollowed     bool
		wg             sync.WaitGroup
		mu             sync.Mutex
	)

	wg.Add(6)

	go func() {
		defer wg.Done()
		users := s.UserService.BatchGetUserInfo(ctx, []uint64{note.UserID})
		if u, ok := users[note.UserID]; ok {
			mu.Lock()
			userInfo = &types.UserProfile{Avatar: u.Avatar, Nickname: u.Nickname}
			mu.Unlock()
		}
	}()

	go func() {
		defer wg.Done()
		if p, err := s.getCommentPreview(ctx, noteID, currentUserID); err == nil {
			mu.Lock()
			commentPreview = p
			mu.Unlock()
		}
	}()

	go func() {
		defer wg.Done()
		if m, err := s.LikeService.BatchGetNoteStats(ctx, []uint64{noteID}); err == nil {
			if st, ok := m[noteID]; ok {
				mu.Lock()
				stats = st
				mu.Unlock()
			}
		}
	}()

	go func() {
		defer wg.Done()
		if currentUserID > 0 {
			if v, err := s.LikeService.checkLikeStatus(ctx, currentUserID, noteID); err == nil {
				mu.Lock()
				isLiked = v
				mu.Unlock()
			}
		}
	}()

	go func() {
		defer wg.Done()
		if currentUserID > 0 {
			if v, err := s.CollectService.CheckCollectStatus(ctx, currentUserID, noteID); err == nil {
				mu.Lock()
				isCollected = v
				mu.Unlock()
			}
		}
	}()

	go func() {
		defer wg.Done()
		if currentUserID > 0 && currentUserID != note.UserID {
			if v, err := s.FollowService.CheckFollowStatus(ctx, currentUserID, note.UserID); err == nil {
				mu.Lock()
				isFollowed = v
				mu.Unlock()
			}
		}
	}()

	wg.Wait()

	// 异步增加浏览数
	go func() {
		_ = s.incrementViewCount(context.Background(), noteID)
	}()

	return s.buildNoteDetail(ctx, note, userInfo, stats, isLiked, isCollected, isFollowed, commentPreview, currentUserID), nil
}

func (s *NoteService) getCommentPreview(ctx context.Context, noteID uint64, currentUserID uint64) (*types.CommentPreview, error) {
	comments, err := s.CommentService.GetTopComments(ctx, noteID, 3, currentUserID)
	if err != nil {
		return nil, err
	}

	totalCount, err := s.CommentDAO.GetRootCommentCount(ctx, noteID)
	if err != nil {
		totalCount = 0
	}

	return &types.CommentPreview{
		TotalCount: totalCount,
		Comments:   comments,
	}, nil
}

func (s *NoteService) checkNoteVisible(ctx context.Context, note *models.Note, currentUserID uint64) error {
	if note.Status != 1 && currentUserID != note.UserID {
		return errors.New("笔记审核中或已下架")
	}

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

func (s *NoteService) buildNoteDetail(
	ctx context.Context,
	note *models.Note,
	userInfo *types.UserProfile,
	stats *types.NoteStats,
	isLiked, isCollected, isFollowed bool,
	commentPreview *types.CommentPreview,
	currentUserID uint64,
) *types.NoteDetail {
	detail := &types.NoteDetail{
		ID:          int64(note.ID),
		UserID:      int64(note.UserID),
		Title:       note.Title,
		Content:     note.Content,
		ActivityID:  note.ActivityID,
		Type:        note.Type,
		Status:      note.Status,
		VisibleConf: note.VisibleConf,
		CreatedAt:   note.CreatedAt,
		UpdatedAt:   note.UpdatedAt,
		IsLiked:     isLiked,
		IsCollected: isCollected,
		IsFollowed:  isFollowed,
	}

	if userInfo != nil {
		detail.Nickname = userInfo.Nickname
		detail.Avatar = userInfo.Avatar
	}

	if stats != nil {
		detail.LikeCount = stats.LikeCount
		detail.CollCount = stats.CollCount
		detail.ShareCount = stats.ShareCount
	}

	if commentPreview != nil {
		detail.CommentCount = commentPreview.TotalCount
		detail.CommentPreview = commentPreview
	}
	var topicIDs []int64
	_ = jsonUnmarshalOr(note.TopicIDs, &topicIDs, make([]int64, 0))
	topics, _ := s.getTopicsByIDs(context.Background(), topicIDs)
	detail.TopicIDs = topics
	_ = jsonUnmarshalOr(note.Location, &detail.Location, types.Location{})
	_ = jsonUnmarshalOr(note.MediaData, &detail.MediaData, make([]types.NoteMedia, 0))

	if note.ActivityID > 0 {
		detail.Activity = s.buildNoteActivity(ctx, int64(note.ActivityID), int64(currentUserID))
	}
	return detail
}

func (s *NoteService) buildNoteActivity(ctx context.Context, activityID int64, currentUserID int64) *types.NoteActivity {
	var activity models.Activity
	if err := s.DB.WithContext(ctx).First(&activity, activityID).Error; err != nil {
		return nil
	}
	resp := &types.NoteActivity{
		ID:           activity.ID,
		Name:         activity.Name,
		ShareTitle:   activity.ShareTitle,
		Description:  activity.Description,
		StartTime:    activity.StartTime,
		EndTime:      activity.EndTime,
		Address:      activity.Address,
		Latitude:     activity.Latitude,
		Longitude:    activity.Longitude,
		PosterList:   activity.PosterList,
		PosterDetail: activity.PosterDetail,
		Status:       activity.Status,
		DetailURL:    fmt.Sprintf("/api/v1/activity/%d", activity.ID),
		OrganizerID:  activity.OrganizerID,
	}
	var organizer models.Organizer
	if err := s.DB.WithContext(ctx).First(&organizer, activity.OrganizerID).Error; err == nil {
		resp.OrganizerName = organizer.Name
		resp.OrganizerLogo = organizer.Logo
	}
	if currentUserID > 0 {
		var count int64
		_ = s.DB.WithContext(ctx).Model(&models.ActivitySubscription{}).
			Where("activity_id = ? AND user_id = ?", activityID, currentUserID).
			Count(&count).Error
		resp.IsSubscribe = count > 0
	}
	return resp
}

func (s *NoteService) incrementViewCount(ctx context.Context, noteID uint64) error {
	key := fmt.Sprintf("note:view:count:%d", noteID)
	s.RedisClient.Incr(ctx, key)
	s.RedisClient.Expire(ctx, key, 24*time.Hour)
	return nil
}

// ============================================================================
// 工具函数
// ============================================================================

// cond 是一个简单的三元表达式替代
func cond[T any](predicate bool, trueVal, falseVal T) T {
	if predicate {
		return trueVal
	}
	return falseVal
}
