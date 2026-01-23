package service

import (
	"context"
	"gorm.io/gorm"
)

var _ IPayService = (*PayService)(nil)

type PayService struct {
	DB *gorm.DB
}

type IPayService interface {
	CheckCollectStatus(ctx context.Context, userID, noteID uint64) error
}

func (p *PayService) CheckCollectStatus(ctx context.Context, userID, noteID uint64) error {

	return nil
}
