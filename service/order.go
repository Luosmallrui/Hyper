package service

import (
	"Hyper/models"
	"Hyper/types"
	"context"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type OrderService struct {
	Redis *redis.Client
	DB    *gorm.DB
}

var _ IOrderService = (*OrderService)(nil)

type IOrderService interface {
	GetOrderList(ctx context.Context, UserId int) ([]*types.Order, error)
}

func (f *OrderService) GetOrderList(ctx context.Context, UserId int) ([]*types.Order, error) {
	var orders []*models.Order
	// 1. 获取订单主表
	err := f.DB.WithContext(ctx).Where("user_id = ?", UserId).Order("id desc").Find(&orders).Error
	if err != nil {
		return nil, err
	}
	if len(orders) == 0 {
		return make([]*types.Order, 0), nil
	}

	resp := make([]*types.Order, len(orders))
	orderIds := make([]int, len(orders))
	for i, order := range orders {
		resp[i] = &types.Order{
			Id:      order.ID,
			Created: order.CreatedAt,
			PaidAt:  order.UpdatedAt,
			Status:  order.Status,
			Price:   int(order.TotalAmount),
			UserId:  order.UserID,
		}
		orderIds[i] = order.ID
	}

	// 2. 获取订单详情 (OrderItem)
	var orderItems []*models.OrderItem
	err = f.DB.WithContext(ctx).Where("order_id IN ?", orderIds).Find(&orderItems).Error
	if err != nil {
		return resp, err
	}

	// 使用 Map 存储详情，并使用 set 思想收集去重后的 SellerID
	orderItemMap := make(map[int]*models.OrderItem)
	sellerIDMap := make(map[int]struct{}) // 用于去重
	for _, item := range orderItems {
		orderItemMap[int(item.OrderID)] = item
		sellerIDMap[item.SellerID] = struct{}{}
	}

	// 3. 批量获取卖家信息 (Parties)
	if len(sellerIDMap) > 0 {
		// 提取去重后的 ID 列表
		uniqueSellerIDs := make([]int, 0, len(sellerIDMap))
		for sid := range sellerIDMap {
			uniqueSellerIDs = append(uniqueSellerIDs, sid)
		}

		var sellerItems []*struct {
			Id    int    `gorm:"column:id"`
			Title string `gorm:"column:title"`
		}
		// 显式指定映射，避免 types.Seller 结构体不匹配
		err = f.DB.WithContext(ctx).Table("parties").
			Select("id, title").
			Where("id IN ?", uniqueSellerIDs).
			Find(&sellerItems).Error

		if err == nil {
			sellerMap := make(map[int]string)
			for _, s := range sellerItems {
				sellerMap[s.Id] = s.Title
			}

			// 4. 最后一次性回填数据
			for i, item := range resp {
				if orderItem, ok := orderItemMap[item.Id]; ok && orderItem != nil {
					resp[i].Name = orderItem.ProductName
					resp[i].ImageUrl = orderItem.CoverImage
					resp[i].Quantity = int(orderItem.Quantity)
					resp[i].Type = orderItem.ConsumeType

					// 回填卖家名称
					if title, exist := sellerMap[orderItem.SellerID]; exist {
						resp[i].SellerName = title
					}
				} else {
					resp[i].Name = "未知商品"
				}
			}
		}
	}

	return resp, nil
}
