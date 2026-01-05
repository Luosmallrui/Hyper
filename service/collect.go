package service

import (
	"Hyper/dao"
	"context"
	"errors"
)

var _ ICollectService = (*CollectService)(nil)

type ICollectService interface {
	Collect(ctx context.Context, userID uint64, noteID uint64) error
	Uncollect(ctx context.Context, userID uint64, noteID uint64) error
	IsCollected(ctx context.Context, userID uint64, noteID uint64) (bool, error)
	GetCollectionCount(ctx context.Context, noteID uint64) (int64, error)
}

type CollectService struct {
	CollectionDAO *dao.NoteCollectionDAO
	StatsDAO      *dao.NoteStatsDAO
	NoteDAO       *dao.NoteDAO
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
