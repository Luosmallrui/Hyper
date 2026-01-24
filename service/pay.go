package service

import (
	"gorm.io/gorm"
)

var _ IPayService = (*PayService)(nil)

type PayService struct {
	DB *gorm.DB
}

type IPayService interface {
}
