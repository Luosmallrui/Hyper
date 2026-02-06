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
	GetOrderList(ctx context.Context, UserId int, cursor int64, pageSize int) ([]*types.Order, int64, bool, error)
}

func (f *OrderService) GetOrderList(ctx context.Context, UserId int, cursor int64, pageSize int) ([]*types.Order, int64, bool, error) {
	if pageSize <= 0 {
		pageSize = 10 // 默认每页10条
	}

	var orders []*models.Order
	query := f.DB.WithContext(ctx).Where("user_id = ?", UserId)

	// 1. Cursor 分页核心逻辑
	if cursor > 0 {
		// 假设按 ID 倒序排列，下一页的数据 ID 必须小于当前的游标
		query = query.Where("id < ?", cursor)
	}

	// 多查一条（pageSize + 1）用来判断是否还有下一页
	err := query.Order("id desc").Limit(pageSize + 1).Find(&orders).Error
	if err != nil {
		return nil, 0, false, err
	}

	// 2. 判断 hasMore
	hasMore := false
	if len(orders) > pageSize {
		hasMore = true
		orders = orders[:pageSize] // 截掉最后一条多查的
	}

	if len(orders) == 0 {
		return make([]*types.Order, 0), 0, false, nil
	}

	nextCursor := orders[len(orders)-1].CreatedAt.UnixNano()

	// --- 下面是之前的详情填充逻辑 ---
	resp := make([]*types.Order, len(orders))
	orderSns := make([]string, len(orders))
	for i, order := range orders {
		resp[i] = &types.Order{
			Id:      order.ID,
			Created: order.CreatedAt,
			Status:  order.Status,
			Price:   int(order.TotalAmount),
		}
		orderSns[i] = order.OrderSn
	}
	// 2. 获取订单详情 (OrderItem)
	var orderItems []*models.OrderItem
	err = f.DB.WithContext(ctx).Where("order_sn IN ?", orderSns).Find(&orderItems).Error
	if err != nil {
		return resp, nextCursor, hasMore, err
	}

	// 使用 Map 存储详情，并使用 set 思想收集去重后的 SellerID
	orderItemMap := make(map[string]*models.OrderItem)
	sellerIDMap := make(map[int]struct{}) // 用于去重

	for _, item := range orderItems {
		orderItemMap[item.OrderSn] = item
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
			for i, _ := range resp {
				if orderItem, ok := orderItemMap[orders[i].OrderSn]; ok && orderItem != nil {
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

	return resp, nextCursor, hasMore, err
}
