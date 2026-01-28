package types

// PointRecord 每一条流水的细节
type PointRecord struct {
	ID          int    `json:"id"`          // 流水唯一ID
	Amount      int    `json:"amount"`      // 变动数值（如 +10, -50）
	Description string `json:"description"` // 详细描述（如：签到奖励、兑换商品）
	OrderType   string `json:"order_type"`  // 业务类型：INCOME(收入), EXPENSE(支出)
	Status      int    `json:"status"`      // 状态：0-待入账, 1-已入账
	CreatedAt   string `json:"created_at"`  // 发放/变动时间：格式化字符串
}

// ListPointsRecord 流水列表包装
type ListPointsRecord struct {
	Records    []PointRecord `json:"records"`     // 积分流水细节
	NextCursor int64         `json:"next_cursor"` // 游标：用于下一页请求
	HasMore    bool          `json:"has_more"`    // 标记是否还有更多数据
}

// PointsAccount 账户概览统计
type PointsAccount struct {
	Balance       int `json:"balance"`        // 当前可用积分余额
	TotalEarned   int `json:"total_earned"`   // 历史累计获得
	TotalUsed     int `json:"total_used"`     // 历史累计使用
	PendingCount  int `json:"pending_count"`  // 待入账订单数量
	PendingAmount int `json:"pending_amount"` // 待入账积分总额
}

// UserPointsResponse 用户点击“我的积分”后的总返回
type UserPointsResponse struct {
	Account PointsAccount    `json:"account"` // 顶部概览：余额、待入账等
	History ListPointsRecord `json:"history"` // 积分流水列表：支持分页和排序
}

// RewardPointsReq 奖励积分请求体
type RewardPointsReq struct {
	UserID     uint64 `json:"user_id" binding:"required"`     // 目标用户ID
	Amount     int64  `json:"amount" binding:"required,gt=0"` // 奖励数额
	ChangeType int    `json:"change_type" binding:"required"` // 变动类型 (1-签到, 2-返利等)
	SourceID   string `json:"source_id" binding:"required"`   // 唯一业务单号 (幂等关键)
	Remark     string `json:"remark" binding:"required"`      // 备注
	IsPending  bool   `json:"is_pending"`
}
type ConsumePointsReq struct {
	Amount     int64  `json:"amount" binding:"required,gt=0"` // 消费数额 (前端传正数)
	ChangeType int    `json:"change_type" binding:"required"` // 变动类型: 3-积分抵扣, 6-兑换商品等
	SourceID   string `json:"source_id" binding:"required"`   // 业务关联ID (订单号)
	Remark     string `json:"remark"`                         // 备注
}

type PointsAccountResp struct {
	Balance       int64 `json:"balance"`        // 当前可用余额
	TotalEarned   int64 `json:"total_earned"`   // 历史累计获得
	TotalUsed     int64 `json:"total_used"`     // 历史累计消耗
	PendingCount  int   `json:"pending_count"`  // 待入账笔数 (Status=0)
	PendingAmount int64 `json:"pending_amount"` // 待入账总额
}

// PointRecordItem 单条流水记录详情
type PointRecordItem struct {
	ID          uint64 `json:"id"`
	Amount      int64  `json:"amount"`      // 变动数值（正数为入账，负数为支出）
	Balance     int64  `json:"balance"`     // 变动后的余额快照
	Description string `json:"description"` // 对应 models.PointsLog 的 Remark
	OrderType   string `json:"order_type"`  // 界面显示类型: INCOME / EXPENSE
	Status      int    `json:"status"`      // 0: 待入账, 1: 已入账
	CreatedAt   string `json:"created_at"`  // 格式化时间: 2006-01-02 15:04:05
}

type ListPointRecordsReq struct {
	Action uint8 `form:"action" binding:"oneof=0 1 2"` // 0-全部, 1-仅收入, 2-仅支出
	Cursor int64 `form:"cursor"`                       // 分页游标 (ID)
	Limit  int   `form:"limit,default=10"`             // 每页数量
}
