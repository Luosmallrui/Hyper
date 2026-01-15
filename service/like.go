package service

import (
	"Hyper/dao"
	"Hyper/models"
	"Hyper/types"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cloudwego/kitex/tool/internal_pkg/log"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"strings"
	"time"
)

const (
	// 统计数据缓存: note:stats:{note_id} (包含所有计数)
	NoteStatsKey = "note:stats:%d"

	// 点赞数单独缓存: note:like:count:{note_id}
	NoteLikeCountKey = "note:like:count:%d"

	// 用户点赞集合: user:liked:notes:{user_id}
	UserLikedNotesKey = "user:liked:notes:%d"

	CacheTTL = 24 * time.Hour
)

var _ ILikeService = (*LikeService)(nil)

type ILikeService interface {
	Like(ctx context.Context, userID uint64, noteID uint64) error
	Unlike(ctx context.Context, userID uint64, noteID uint64) error
	IsLiked(ctx context.Context, userID uint64, noteID uint64) (bool, error)
	GetLikeCount(ctx context.Context, noteID uint64) (int64, error)
	GetUserTotalLikes(ctx context.Context, userID uint64) (int64, error)
	BatchGetNoteStats(ctx context.Context, noteIDs []uint64) (map[uint64]*types.NoteStats, error)
	BatchCheckLikeStatus(ctx context.Context, userID uint64, noteIDs []uint64) (map[uint64]bool, error)
	LikeNote(ctx context.Context, userID, noteID uint64) error
	UnlikeNote(ctx context.Context, userID, noteID uint64) error
	checkLikeStatus(ctx context.Context, userID, noteID uint64) (bool, error)
}

type LikeService struct {
	LikeDAO  *dao.NoteLikeDAO
	StatsDAO *dao.NoteStatsDAO
	NoteDAO  *dao.NoteDAO
	Redis    *redis.Client
}

func (s *LikeService) Like(ctx context.Context, userID uint64, noteID uint64) error {
	// 校验笔记存在
	exist, err := s.NoteDAO.IsExist(ctx, "id = ?", noteID)
	if err != nil {
		return err
	}
	if !exist {
		return errors.New("笔记不存在")
	}

	// 检查用户是否已经点赞过
	isLiked, err := s.LikeDAO.IsLiked(ctx, noteID, userID)
	if err != nil {
		return err
	}
	if isLiked {
		// 已经点赞过，不做任何操作
		return nil
	}

	// 设置点赞状态为已点赞
	if err := s.LikeDAO.SetStatus(ctx, noteID, userID, 1); err != nil {
		return err
	}
	// 计数 +1（只有在之前未点赞时才增加）
	if err := s.StatsDAO.IncrLikeCount(ctx, noteID, 1); err != nil {
		return err
	}
	return nil
}

func (s *LikeService) Unlike(ctx context.Context, userID uint64, noteID uint64) error {
	// 校验笔记存在
	exist, err := s.NoteDAO.IsExist(ctx, "id = ?", noteID)
	if err != nil {
		return err
	}
	if !exist {
		return errors.New("笔记不存在")
	}

	// 检查用户是否已经点赞
	isLiked, err := s.LikeDAO.IsLiked(ctx, noteID, userID)
	if err != nil {
		return err
	}
	if !isLiked {
		// 没有点赞过，不做任何操作
		return nil
	}

	// 设置点赞状态为未点赞
	if err := s.LikeDAO.SetStatus(ctx, noteID, userID, 0); err != nil {
		return err
	}
	// 计数 -1（只有在之前已点赞时才减少）
	if err := s.StatsDAO.IncrLikeCount(ctx, noteID, -1); err != nil {
		return err
	}
	return nil
}

func (s *LikeService) IsLiked(ctx context.Context, userID uint64, noteID uint64) (bool, error) {
	return s.LikeDAO.IsLiked(ctx, noteID, userID)
}

func (s *LikeService) GetLikeCount(ctx context.Context, noteID uint64) (int64, error) {
	stat, err := s.StatsDAO.GetByNoteID(ctx, noteID)
	if err != nil {
		return 0, err
	}
	if stat == nil {
		return 0, errors.New("stat not found")
	}
	return int64(stat.LikeCount), nil
}

func (s *LikeService) GetUserTotalLikes(ctx context.Context, userID uint64) (int64, error) {
	return s.StatsDAO.GetUserTotalLikes(ctx, userID)
}

func (s *LikeService) LikeNote(ctx context.Context, userID, noteID uint64) error {
	// 1. 分布式锁,防止重复点赞
	lockKey := fmt.Sprintf("lock:like:%d:%d", userID, noteID)
	lock, err := s.Redis.SetNX(ctx, lockKey, 1, 5*time.Second).Result()
	if err != nil || !lock {
		return errors.New("操作太频繁,请稍后重试")
	}
	defer s.Redis.Del(ctx, lockKey)

	// 2. 检查是否已点赞
	isLiked, err := s.checkLikeStatus(ctx, userID, noteID)
	if err != nil {
		return err
	}
	if isLiked {
		return errors.New("已经点赞过了")
	}

	// 3. 先写数据库(保证数据一致性)
	if err := s.createLikeRecord(ctx, userID, noteID); err != nil {
		return err
	}

	// 4. 更新 Redis 缓存(即使失败也不影响)
	s.updateRedisAfterLike(ctx, userID, noteID)

	return nil
}

// 取消点赞
func (s *LikeService) UnlikeNote(ctx context.Context, userID, noteID uint64) error {
	lockKey := fmt.Sprintf("lock:like:%d:%d", userID, noteID)
	lock, err := s.Redis.SetNX(ctx, lockKey, 1, 5*time.Second).Result()
	if err != nil || !lock {
		return errors.New("操作太频繁,请稍后重试")
	}
	defer s.Redis.Del(ctx, lockKey)

	// 检查是否已点赞
	isLiked, err := s.checkLikeStatus(ctx, userID, noteID)
	if err != nil {
		return err
	}
	if !isLiked {
		return errors.New("还未点赞")
	}

	// 先写数据库
	if err := s.deleteLikeRecord(ctx, userID, noteID); err != nil {
		return err
	}

	// 更新 Redis
	s.updateRedisAfterUnlike(ctx, userID, noteID)

	return nil
}

// 写数据库:创建点赞记录 + 更新统计
func (s *LikeService) createLikeRecord(ctx context.Context, userID, noteID uint64) error {
	return s.LikeDAO.Transaction(ctx, func(tx *gorm.DB) error {
		// 1. 插入点赞记录
		like := &models.NoteLike{
			UserID:    int(userID),
			NoteID:    noteID,
			CreatedAt: time.Now(),
		}
		if err := tx.Create(like).Error; err != nil {
			// 唯一键冲突说明已经点赞了
			if strings.Contains(err.Error(), "Duplicate entry") {
				return errors.New("已经点赞过了")
			}
			return err
		}

		// 2. 更新统计表
		result := tx.Model(&models.NoteStats{}).
			Where("note_id = ?", noteID).
			UpdateColumn("like_count", gorm.Expr("like_count + 1"))

		if result.Error != nil {
			return result.Error
		}

		// 如果统计记录不存在,创建一条
		if result.RowsAffected == 0 {
			stats := &models.NoteStats{
				NoteID:    noteID,
				LikeCount: 1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			return tx.Create(stats).Error
		}

		return nil
	})
}

// 写数据库:删除点赞记录 + 更新统计
func (s *LikeService) deleteLikeRecord(ctx context.Context, userID, noteID uint64) error {
	return s.LikeDAO.Transaction(ctx, func(tx *gorm.DB) error {
		// 1. 删除点赞记录
		result := tx.Where("user_id = ? AND note_id = ?", userID, noteID).
			Delete(&models.NoteLike{})

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return errors.New("点赞记录不存在")
		}

		// 2. 更新统计表
		return tx.Model(&models.NoteStats{}).
			Where("note_id = ?", noteID).
			UpdateColumn("like_count", gorm.Expr("like_count - 1")).
			Error
	})
}

// 更新 Redis:点赞后
func (s *LikeService) updateRedisAfterLike(ctx context.Context, userID, noteID uint64) {
	pipe := s.Redis.Pipeline()

	// 1. 点赞数 +1
	likeCountKey := fmt.Sprintf(NoteLikeCountKey, noteID)
	pipe.Incr(ctx, likeCountKey)
	pipe.Expire(ctx, likeCountKey, CacheTTL)

	// 2. 用户点赞集合添加
	userLikedKey := fmt.Sprintf(UserLikedNotesKey, userID)
	pipe.SAdd(ctx, userLikedKey, noteID)
	pipe.Expire(ctx, userLikedKey, CacheTTL)

	// 3. 统计数据缓存失效(下次查询重新加载)
	statsKey := fmt.Sprintf(NoteStatsKey, noteID)
	pipe.Del(ctx, statsKey)

	// 执行 Pipeline (即使失败也不影响业务,只是缓存不一致)
	if _, err := pipe.Exec(ctx); err != nil {
		log.Error("更新Redis缓存失败", "error", err, "userID", userID, "noteID", noteID)
	}
}

// 更新 Redis:取消点赞后
func (s *LikeService) updateRedisAfterUnlike(ctx context.Context, userID, noteID uint64) {
	pipe := s.Redis.Pipeline()

	// 1. 点赞数 -1
	likeCountKey := fmt.Sprintf(NoteLikeCountKey, noteID)
	pipe.Decr(ctx, likeCountKey)

	// 2. 用户点赞集合移除
	userLikedKey := fmt.Sprintf(UserLikedNotesKey, userID)
	pipe.SRem(ctx, userLikedKey, noteID)

	// 3. 统计数据缓存失效
	statsKey := fmt.Sprintf(NoteStatsKey, noteID)
	pipe.Del(ctx, statsKey)

	pipe.Exec(ctx)
}

// 检查点赞状态
func (s *LikeService) checkLikeStatus(ctx context.Context, userID, noteID uint64) (bool, error) {
	// 1. 先查 Redis
	userLikedKey := fmt.Sprintf(UserLikedNotesKey, userID)
	exists, err := s.Redis.SIsMember(ctx, userLikedKey, noteID).Result()
	if err == nil {
		return exists, nil
	}

	// 2. Redis 查不到,查数据库
	exists, err = s.LikeDAO.CheckExists(ctx, userID, noteID)
	if err != nil {
		return false, err
	}

	// 3. 回写 Redis
	if exists {
		s.Redis.SAdd(ctx, userLikedKey, noteID)
		s.Redis.Expire(ctx, userLikedKey, CacheTTL)
	}

	return exists, nil
}

// 批量获取点赞数
func (s *LikeService) BatchGetLikeCount(ctx context.Context, noteIDs []uint64) (map[uint64]int64, error) {
	result := make(map[uint64]int64)
	if len(noteIDs) == 0 {
		return result, nil
	}

	missedIDs := make([]uint64, 0)

	// 1. 批量从 Redis 获取
	pipe := s.Redis.Pipeline()
	cmds := make(map[uint64]*redis.StringCmd)

	for _, noteID := range noteIDs {
		key := fmt.Sprintf(NoteLikeCountKey, noteID)
		cmds[noteID] = pipe.Get(ctx, key)
	}

	pipe.Exec(ctx)

	// 2. 收集 Redis 结果
	for noteID, cmd := range cmds {
		count, err := cmd.Int64()
		if err == nil {
			result[noteID] = count
		} else {
			missedIDs = append(missedIDs, noteID)
		}
	}

	// 3. 从数据库加载缺失的
	if len(missedIDs) > 0 {
		statsList, err := s.StatsDAO.BatchGetByNoteIDs(ctx, missedIDs)
		if err != nil {
			return result, err
		}

		// 回写 Redis
		pipe2 := s.Redis.Pipeline()
		for _, stats := range statsList {
			result[stats.NoteID] = int64(stats.LikeCount)
			key := fmt.Sprintf(NoteLikeCountKey, stats.NoteID)
			pipe2.Set(ctx, key, stats.LikeCount, CacheTTL)
		}
		pipe2.Exec(ctx)
	}

	return result, nil
}

// 批量获取完整统计数据
func (s *LikeService) BatchGetNoteStats(ctx context.Context, noteIDs []uint64) (map[uint64]*types.NoteStats, error) {
	result := make(map[uint64]*types.NoteStats)
	if len(noteIDs) == 0 {
		return result, nil
	}

	missedIDs := make([]uint64, 0)

	// 1. 从 Redis 获取
	pipe := s.Redis.Pipeline()
	cmds := make(map[uint64]*redis.StringCmd)

	for _, noteID := range noteIDs {
		key := fmt.Sprintf(NoteStatsKey, noteID)
		cmds[noteID] = pipe.Get(ctx, key)
	}

	pipe.Exec(ctx)

	// 2. 解析结果
	for noteID, cmd := range cmds {
		statsJSON, err := cmd.Result()
		if err == nil {
			var stats types.NoteStats
			if json.Unmarshal([]byte(statsJSON), &stats) == nil {
				result[noteID] = &stats
				continue
			}
		}
		missedIDs = append(missedIDs, noteID)
	}

	// 3. 从数据库加载
	if len(missedIDs) > 0 {
		statsList, err := s.StatsDAO.BatchGetByNoteIDs(ctx, missedIDs)
		if err != nil {
			return result, err
		}

		// 回写 Redis
		pipe2 := s.Redis.Pipeline()
		for _, stats := range statsList {
			noteStats := &types.NoteStats{
				NoteID:       stats.NoteID,
				LikeCount:    stats.LikeCount,
				CollCount:    stats.CollCount,
				ShareCount:   stats.ShareCount,
				CommentCount: stats.CommentCount,
			}
			result[stats.NoteID] = noteStats

			statsJSON, _ := json.Marshal(noteStats)
			key := fmt.Sprintf(NoteStatsKey, stats.NoteID)
			pipe2.Set(ctx, key, statsJSON, CacheTTL)
		}
		pipe2.Exec(ctx)
	}

	return result, nil
}

// 批量检查点赞状态
func (s *LikeService) BatchCheckLikeStatus(ctx context.Context, userID uint64, noteIDs []uint64) (map[uint64]bool, error) {
	result := make(map[uint64]bool)
	if userID == 0 || len(noteIDs) == 0 {
		return result, nil
	}

	// 1. 从 Redis 批量检查
	userLikedKey := fmt.Sprintf(UserLikedNotesKey, userID)

	pipe := s.Redis.Pipeline()
	cmds := make(map[uint64]*redis.BoolCmd)

	for _, noteID := range noteIDs {
		cmds[noteID] = pipe.SIsMember(ctx, userLikedKey, noteID)
	}

	pipe.Exec(ctx)

	// 2. 收集结果
	needDBCheck := make([]uint64, 0)
	for noteID, cmd := range cmds {
		exists, err := cmd.Result()
		if err == nil && exists {
			result[noteID] = true
		} else if err != nil {
			// Redis 查不到,需要查数据库
			needDBCheck = append(needDBCheck, noteID)
		}
	}

	// 3. 查数据库
	if len(needDBCheck) > 0 {
		likes, err := s.LikeDAO.BatchGetByUserAndNotes(ctx, userID, needDBCheck)
		if err == nil && len(likes) > 0 {
			// 回写 Redis
			pipe2 := s.Redis.Pipeline()
			for _, like := range likes {
				result[like.NoteID] = true
				pipe2.SAdd(ctx, userLikedKey, like.NoteID)
			}
			pipe2.Expire(ctx, userLikedKey, CacheTTL)
			pipe2.Exec(ctx)
		}
	}

	return result, nil
}
