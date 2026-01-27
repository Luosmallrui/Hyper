package service

import (
	"Hyper/config"
	"Hyper/dao"
	"Hyper/models"
	"Hyper/types"
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type ProductService struct {
	Config     *config.Config
	DB         *gorm.DB
	Redis      *redis.Client
	ProductDAO *dao.Product
}

var _ IProductService = (*ProductService)(nil)

type IProductService interface {
	CreateProduct(ctx context.Context, req *types.CreateProductRequest) error
	// 定义商品相关的方法接口
}

func (p *ProductService) CreateProduct(ctx context.Context, req *types.CreateProductRequest) error {
	if req.Price < 10 {
		return errors.New("商品价格不能低于10分")
	}
	var count int64
	p.DB.Model(&models.Product{}).Where("mproduct_name = ?", req.ProductName).Count(&count)
	if count > 0 {
		return errors.New("该商家已存在同名商品，请勿重复创建")
	}
	//要不要加验证上传者的身份是不是商家
	product := &models.Product{
		ProductName: req.ProductName,
		Price:       req.Price,
		Stock:       req.Stock,
		Description: req.Description,
		CoverImage:  req.CoverImage,
		Status:      req.Status,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err := p.ProductDAO.Create(ctx, product)
	if err != nil {
		return errors.New("创建商品失败: " + err.Error())
	}
	return nil
}
