package service

import (
	"Hyper/dao"
	"context"
	"errors"
)

var _ ILikeService = (*LikeService)(nil)

type ILikeService interface {
	Like(ctx context.Context, userID uint64, noteID uint64) error
	Unlike(ctx context.Context, userID uint64, noteID uint64) error
	IsLiked(ctx context.Context, userID uint64, noteID uint64) (bool, error)
	GetLikeCount(ctx context.Context, noteID uint64) (int64, error)
	GetUserTotalLikes(ctx context.Context, userID uint64) (int64, error)
}

type LikeService struct {
	LikeDAO  *dao.NoteLikeDAO
	StatsDAO *dao.NoteStatsDAO
	NoteDAO  *dao.NoteDAO
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
