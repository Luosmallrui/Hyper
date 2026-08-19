package service

import (
	"Hyper/dao"
	"Hyper/models"
	"Hyper/types"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

	// 收藏关系以 note_collections.status 为准。此前这里读取 Redis Set，
	// 但收藏/取消收藏流程并没有维护该 Set；Redis 对不存在的 Set 会返回
	// false 而不会报错，导致详情页始终把已收藏显示成未收藏。
	return s.CollectionDAO.IsCollected(ctx, noteID, userID)
}
func (s *CollectService) Collect(ctx context.Context, userID uint64, noteID uint64) error {
	exist, err := s.NoteDAO.IsExist(ctx, "id = ?", noteID)
	if err != nil {
		return err
	}
	if !exist {
		return errors.New("笔记不存在")
	}

	// Redis 分布式锁（与点赞 LikeNote 的模式对齐），防止并发双击计数翻倍
	lockKey := fmt.Sprintf("lock:collect:%d:%d", userID, noteID)
	lock, err := s.Redis.SetNX(ctx, lockKey, 1, 5*time.Second).Result()
	if err != nil || !lock {
		return errors.New("操作太频繁,请稍后重试")
	}
	defer s.Redis.Del(ctx, lockKey)

	// 状态迁移与计数更新放进同一事务（行锁保证并发下只迁移一次）
	err = s.CollectionDAO.Db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var item models.NoteCollection
		findErr := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("note_id = ? AND user_id = ?", noteID, int(userID)).
			First(&item).Error
		switch {
		case errors.Is(findErr, gorm.ErrRecordNotFound):
			item = models.NoteCollection{NoteID: noteID, UserID: int(userID), Status: 1, CreatedAt: time.Now()}
			if err := tx.Create(&item).Error; err != nil {
				return err
			}
		case findErr != nil:
			return findErr
		case item.Status == 1:
			// 已收藏，幂等返回，不重复增加计数
			return nil
		default:
			if err := tx.Model(&models.NoteCollection{}).Where("id = ?", item.ID).Update("status", 1).Error; err != nil {
				return err
			}
		}

		result := tx.Model(&models.NoteStats{}).
			Where("note_id = ?", noteID).
			UpdateColumn("coll_count", gorm.Expr("coll_count + 1"))
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return tx.Create(&models.NoteStats{
				NoteID:    noteID,
				CollCount: 1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}).Error
		}
		return nil
	})
	if err != nil {
		return err
	}
	s.invalidateStatsCache(ctx, noteID)
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

	// Redis 分布式锁（与点赞 LikeNote 的模式对齐），防止并发双击计数翻倍
	lockKey := fmt.Sprintf("lock:collect:%d:%d", userID, noteID)
	lock, err := s.Redis.SetNX(ctx, lockKey, 1, 5*time.Second).Result()
	if err != nil || !lock {
		return errors.New("操作太频繁,请稍后重试")
	}
	defer s.Redis.Del(ctx, lockKey)

	// 状态迁移与计数更新放进同一事务（行锁保证并发下只迁移一次）
	err = s.CollectionDAO.Db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var item models.NoteCollection
		findErr := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("note_id = ? AND user_id = ?", noteID, int(userID)).
			First(&item).Error
		switch {
		case errors.Is(findErr, gorm.ErrRecordNotFound):
			// 没有收藏记录，幂等返回
			return nil
		case findErr != nil:
			return findErr
		case item.Status != 1:
			// 已是未收藏状态，幂等返回，不重复扣减计数
			return nil
		default:
			if err := tx.Model(&models.NoteCollection{}).Where("id = ?", item.ID).Update("status", 0).Error; err != nil {
				return err
			}
		}

		return tx.Model(&models.NoteStats{}).
			Where("note_id = ? AND coll_count > 0", noteID).
			UpdateColumn("coll_count", gorm.Expr("coll_count - 1")).
			Error
	})
	if err != nil {
		return err
	}
	s.invalidateStatsCache(ctx, noteID)
	return nil
}

// invalidateStatsCache 使 note:stats:{note_id} 缓存失效，下次读取会回源到数据库最新值。
// 与点赞流程（service/like.go 的 updateRedisAfterLike/Unlike）保持一致，避免收藏数长期读到旧缓存。
func (s *CollectService) invalidateStatsCache(ctx context.Context, noteID uint64) {
	statsKey := fmt.Sprintf(NoteStatsKey, noteID)
	s.Redis.Del(ctx, statsKey)
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
