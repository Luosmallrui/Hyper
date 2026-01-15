package service

import (
	"Hyper/dao"
	"Hyper/models"
	"Hyper/types"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
)

var _ ICollectService = (*CollectService)(nil)

type ICollectService interface {
	Collect(ctx context.Context, userID uint64, noteID uint64) error
	Uncollect(ctx context.Context, userID uint64, noteID uint64) error
	IsCollected(ctx context.Context, userID uint64, noteID uint64) (bool, error)
	GetCollectionCount(ctx context.Context, noteID uint64) (int64, error)
	GetUserCollections(ctx context.Context, userID uint64, limit, offset int) ([]*types.Note, int64, error)
	GetUserTotalCollects(ctx context.Context, userID uint64) (int64, error)
	CheckCollectStatus(ctx context.Context, userID, noteID uint64) (bool, error)
}

type CollectService struct {
	CollectionDAO *dao.NoteCollectionDAO
	StatsDAO      *dao.NoteStatsDAO
	NoteDAO       *dao.NoteDAO
	Redis         *redis.Client
}

func (s *CollectService) CheckCollectStatus(ctx context.Context, userID, noteID uint64) (bool, error) {
	if userID == 0 {
		return false, nil
	}

	// 类似点赞的逻辑,先查 Redis,再查数据库
	key := fmt.Sprintf("user:collected:notes:%d", userID)
	exists, err := s.Redis.SIsMember(ctx, key, noteID).Result()
	if err == nil {
		return exists, nil
	}

	return s.CollectionDAO.CheckExists(ctx, userID, noteID)
}
func (s *CollectService) Collect(ctx context.Context, userID uint64, noteID uint64) error {
	exist, err := s.NoteDAO.IsExist(ctx, "id = ?", noteID)
	if err != nil {
		return err
	}
	if !exist {
		return errors.New("笔记不存在")
	}

	isCollected, err := s.CollectionDAO.IsCollected(ctx, noteID, userID)
	if err != nil {
		return err
	}
	if isCollected {
		return nil
	}

	if err := s.CollectionDAO.SetStatus(ctx, noteID, userID, 1); err != nil {
		return err
	}
	if err := s.StatsDAO.IncrCollCount(ctx, noteID, 1); err != nil {
		return err
	}
	return nil
}

func (s *CollectService) Uncollect(ctx context.Context, userID uint64, noteID uint64) error {
	exist, err := s.NoteDAO.IsExist(ctx, "id = ?", noteID)
	if err != nil {
		return err
	}
	if !exist {
		return errors.New("笔记不存在")
	}

	isCollected, err := s.CollectionDAO.IsCollected(ctx, noteID, userID)
	if err != nil {
		return err
	}
	if !isCollected {
		return nil
	}

	if err := s.CollectionDAO.SetStatus(ctx, noteID, userID, 0); err != nil {
		return err
	}
	if err := s.StatsDAO.IncrCollCount(ctx, noteID, -1); err != nil {
		return err
	}
	return nil
}

func (s *CollectService) IsCollected(ctx context.Context, userID uint64, noteID uint64) (bool, error) {
	return s.CollectionDAO.IsCollected(ctx, noteID, userID)
}

func (s *CollectService) GetCollectionCount(ctx context.Context, noteID uint64) (int64, error) {
	stat, err := s.StatsDAO.GetByNoteID(ctx, noteID)
	if err != nil {
		return 0, err
	}
	if stat == nil {
		return 0, errors.New("stat not found")
	}
	return int64(stat.CollCount), nil
}

// GetUserCollections 查询用户收藏的笔记列表（分页）
func (s *CollectService) GetUserCollections(ctx context.Context, userID uint64, limit, offset int) ([]*types.Note, int64, error) {
	ids, total, err := s.CollectionDAO.ListNoteIDsByUser(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	if len(ids) == 0 {
		return []*types.Note{}, total, nil
	}
	notes, err := s.NoteDAO.FindByIDs(ctx, ids)
	if err != nil {
		return nil, 0, err
	}
	// 按收藏时间顺序（ListNoteIDsByUser 已按 created_at DESC）恢复顺序
	noteMap := make(map[uint64]*models.Note, len(notes))
	for _, note := range notes {
		noteMap[note.ID] = note
	}
	ordered := make([]*models.Note, 0, len(ids))
	for _, id := range ids {
		if n, ok := noteMap[id]; ok {
			ordered = append(ordered, n)
		}
	}

	result := make([]*types.Note, 0, len(ordered))
	for _, note := range ordered {
		k := &types.Note{
			ID:          int64(note.ID),
			UserID:      int64(note.UserID),
			Title:       note.Title,
			Content:     note.Content,
			Type:        int(note.Type),
			Status:      int(note.Status),
			VisibleConf: int(note.VisibleConf),
			CreatedAt:   note.CreatedAt,
			UpdatedAt:   note.UpdatedAt,
		}
		_ = json.Unmarshal([]byte(note.TopicIDs), &k.TopicIDs)
		_ = json.Unmarshal([]byte(note.Location), &k.Location)
		_ = json.Unmarshal([]byte(note.MediaData), &k.MediaData)
		result = append(result, k)
	}
	return result, total, nil
}

func (s *CollectService) GetUserTotalCollects(ctx context.Context, userID uint64) (int64, error) {
	return s.StatsDAO.GetUserTotalCollects(ctx, userID)
}
