package service

import (
	"Hyper/config"
	"Hyper/dao"
	"Hyper/models"
	"Hyper/types"
	"context"
	"errors"

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
	CreateProduct(ctx context.Context, partyId int, req *types.CreateProductRequest) error
	BatchGetProducts(ctx context.Context, partyId int64, cursor int64, limit int) (*types.BatchGetProductsResponse, error)
	GetDetailProduct(ctx context.Context, productID uint64, partyId int64) (*types.GetDetailProductResponse, error)
	// 定义商品相关的方法接口
}

func (p *ProductService) CreateProduct(ctx context.Context, partyId int, req *types.CreateProductRequest) error {
	// 1. 价格基本校验
	if req.Price < 10 && req.ParentId == 0 {
		return errors.New("主商品价格不得低于10分")
	}

	// 2. 商家合法性与权限校验
	var party models.Merchant
	if err := p.DB.WithContext(ctx).Where("id = ?", partyId).First(&party).Error; err != nil {
		return errors.New("目标商家不存在，无法发布商品")
	}
	//套餐和单品都是商品
	//// 3. 套餐绑定逻辑校验
	//if req.ParentId > 0 {
	//	var parent models.Product
	//	// 校验父级套餐是否存在，且必须属于同一个商家
	//	err := p.DB.WithContext(ctx).Where("id = ? AND party_id = ?", req.ParentId, partyId).First(&parent).Error
	//	if err != nil {
	//		return errors.New("指定的父级套餐不存在或不属于当前商家")
	//	}
	//	// 限制：只有 ParentId 为 0 的才能作为父级（禁止二级嵌套）
	//	if parent.ParentId != 0 {
	//		return errors.New("不能在子商品下继续添加商品")
	//	}
	//}

	var count int64
	err := p.DB.WithContext(ctx).Model(&models.Product{}).
		Where("party_id = ? AND product_name = ? AND parent_id = ? AND deleted_at IS NULL",
			partyId, req.ProductName, req.ParentId).
		Count(&count).Error
	if err != nil {
		return errors.New("系统繁忙，请稍后再试")
	}
	if count > 0 {
		return errors.New("已存在同名商品，请更换名称")
	}

	// 5. 构造并写入模型
	product := &models.Product{
		PartyId:       uint64(partyId),
		ProductName:   req.ProductName,
		Price:         req.Price,
		OriginalPrice: req.OriginalPrice,
		Stock:         req.Stock,
		Description:   req.Description,
		CoverImage:    req.CoverImage,
		Status:        req.Status,
	}

	if err := p.ProductDAO.Create(ctx, product); err != nil {
		return errors.New("数据库写入失败: " + err.Error())
	}
	return nil
}

func (p *ProductService) BatchGetProducts(ctx context.Context, partyId int64, cursor int64, limit int) (*types.BatchGetProductsResponse, error) {
	if limit <= 0 || limit > 6 {
		limit = 6
	}
	queryLimit := limit + 1
	products, err := p.ProductDAO.GetProductsByCursor(ctx, uint64(partyId), queryLimit, int(cursor))
	if err != nil {
		return nil, err
	}
	hasmore := false
	displayCount := len(products)
	if displayCount > limit {
		hasmore = true
		displayCount = limit
		products = products[:displayCount]
	}
	if displayCount == 0 {
		return &types.BatchGetProductsResponse{
			Products:   []*models.Product{},
			HasMore:    false,
			NextCursor: 0,
		}, nil
	}

	nextCursor := int64(0)
	if displayCount > 0 {
		nextCursor = products[displayCount-1].CreatedAt.UnixNano()
	}
	return &types.BatchGetProductsResponse{
		Products:   products,
		HasMore:    hasmore,
		NextCursor: nextCursor,
	}, nil
}

func (p *ProductService) GetDetailProduct(ctx context.Context, productID uint64, partyId int64) (*types.GetDetailProductResponse, error) {
	product, err := p.ProductDAO.FindProductById(ctx, productID, partyId)

	if err != nil {
		return nil, errors.New("查询商品失败: " + err.Error())
	}

	res := types.GetDetailProductResponse{
		ID:            product.ID,
		ProductName:   product.ProductName,
		Price:         product.Price,
		OriginalPrice: product.OriginalPrice,
		CoverImage:    product.CoverImage,
		Description:   product.Description,
		Stock:         product.Stock,
		Status:        product.Status,
		SalesVolume:   product.SalesVolume,
	}

	return &res, nil
}
