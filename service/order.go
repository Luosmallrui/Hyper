package service

import (
	"Hyper/models"
	"Hyper/types"
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

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
	AddViewers(ctx context.Context, UserId int, req types.CreateViewerReq) error
	DeleteViewer(ctx context.Context, UserId int, req types.DeleteViewerReq) error
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

func (f *OrderService) AddViewers(ctx context.Context, UserId int, req types.CreateViewerReq) error {
	return f.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 身份证合法性验证
		viewerType, err := GetAgeTypeByIDCard(req.IDCard)
		if err != nil {
			return err // 已经包含 http code 前缀
		}

		// 2. 数量上限校验
		var count int64
		tx.Model(&models.Viewer{}).Where("user_id = ?", UserId).Count(&count)
		if count >= 5 {
			return errors.New(fmt.Sprintf("%d:常用观影人已达上限", http.StatusBadRequest))
		}

		// 3. 查重 (修正了你代码中的 SQL 笔误: AND user_id 后缺少 = ?)
		var existViewer models.Viewer
		if err := tx.Where("id_card = ? AND user_id = ?", req.IDCard, UserId).First(&existViewer).Error; err == nil {
			return errors.New(fmt.Sprintf("%d:该观影人已存在", http.StatusBadRequest))
		}

		// 4. 执行创建
		viewer := models.Viewer{
			UserID:   UserId,
			RealName: req.RealName,
			IDCard:   req.IDCard,
			Phone:    req.Phone,
			Type:     viewerType,
		}
		if err := tx.Create(&viewer).Error; err != nil {
			return errors.New(fmt.Sprintf("%d:添加失败", http.StatusInternalServerError))
		}
		return nil
	})
}

func GetAgeTypeByIDCard(idCard string) (int8, error) {
	// 调用上面的校验函数
	if !IsValidIDCard(idCard) {
		return 0, errors.New(fmt.Sprintf("%d:身份证号格式或校验码错误", http.StatusBadRequest))
	}

	yearStr := idCard[6:10]
	birthYear, _ := strconv.Atoi(yearStr) // 校验过正则，这里不会报错

	currentYear := time.Now().Year()
	age := currentYear - birthYear

	if age < 18 {
		return 1, nil // 未成年
	} else if age >= 18 && age < 60 {
		return 2, nil // 成年
	} else {
		return 3, nil // 老年
	}
}
func (f *OrderService) DeleteViewer(ctx context.Context, UserId int, req types.DeleteViewerReq) error {
	return f.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var viewer models.Viewer
		err := tx.Where("id = ? AND user_id = ?", req.ID, UserId).First(&viewer).Error
		if err != nil {
			return errors.New("查询观影人失败: " + err.Error())
		}
		//这里要考虑添加的观影人是否存在订单
		if err := tx.Delete(&viewer).Error; err != nil {
			return errors.New("删除观影人失败: " + err.Error())
		}
		return nil
	})
}

// IsValidIDCard 验证身份证号是否合法 (GB 11643-1999)
func IsValidIDCard(idCard string) bool {
	// 1. 长度及基本格式校验
	reg := regexp.MustCompile(`^[1-9]\d{5}(18|19|20)\d{2}((0[1-9])|(1[0-2]))(([0-2][1-9])|10|20|30|31)\d{3}[0-9Xx]$`)
	if !reg.MatchString(idCard) {
		return false
	}

	// 2. 校验码计算 (加权因子)
	weights := []int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
	checkCodes := []byte{'1', '0', 'X', '9', '8', '7', '6', '5', '4', '3', '2'}

	sum := 0
	for i := 0; i < 17; i++ {
		n, _ := strconv.Atoi(string(idCard[i]))
		sum += n * weights[i]
	}

	// 取模结果对应的校验码
	actualCheckCode := idCard[17]
	if actualCheckCode == 'x' {
		actualCheckCode = 'X'
	}

	return checkCodes[sum%11] == actualCheckCode
}
