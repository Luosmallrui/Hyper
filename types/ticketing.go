package types

import (
	"Hyper/models"
	"time"
)

type PageResponse[T any] struct {
	List  []T   `json:"list"`
	Total int64 `json:"total"`
}

type OrganizerApplyRequest struct {
	Name     string `json:"name" binding:"required"`
	Type     string `json:"type" binding:"required,oneof=venue merchant"`
	Logo     string `json:"logo"`
	Province string `json:"province"`
	City     string `json:"city"`
	District string `json:"district"`
}

type OrganizerBasicRequest struct {
	Name     *string `json:"name"`
	Logo     *string `json:"logo"`
	Province *string `json:"province"`
	City     *string `json:"city"`
	District *string `json:"district"`
}

type OrganizerWithdrawRequest struct {
	BankAccountName string `json:"bank_account_name" binding:"required"`
	BankAccountNo   string `json:"bank_account_no" binding:"required"`
	BankName        string `json:"bank_name" binding:"required"`
}

type OrganizerInfoResponse struct {
	ID             int64   `json:"id"`
	Type           string  `json:"type"`
	Name           string  `json:"name"`
	Logo           string  `json:"logo"`
	Status         int8    `json:"status"`
	RejectReason   string  `json:"reject_reason,omitempty"`
	Level          string  `json:"level"`
	ServiceFeeRate float64 `json:"service_fee_rate"`
	JoinDays       int     `json:"join_days"`
	BasicInfo      struct {
		Province string `json:"province"`
		City     string `json:"city"`
		District string `json:"district"`
	} `json:"basic_info"`
	AccountInfo struct {
		BankAccountName string `json:"bank_account_name"`
		BankAccountNo   string `json:"bank_account_no"`
		BankName        string `json:"bank_name"`
	} `json:"account_info"`
}

type ActivityCreateRequest struct {
	ActivityID       int64                `json:"activity_id"`
	Step             int                  `json:"step" binding:"required,min=1,max=5"`
	Name             *string              `json:"name"`
	ShareTitle       *string              `json:"share_title"`
	StartTime        *string              `json:"start_time"`
	EndTime          *string              `json:"end_time"`
	RealNameMode     *int8                `json:"real_name_mode"`
	MinorCheck       *int8                `json:"minor_check"`
	Description      *string              `json:"description"`
	Province         *string              `json:"province"`
	City             *string              `json:"city"`
	District         *string              `json:"district"`
	Address          *string              `json:"address"`
	Latitude         *float64             `json:"latitude"`
	Longitude        *float64             `json:"longitude"`
	PosterDetail     *string              `json:"poster_detail"`
	PosterLong       *string              `json:"poster_long"`
	PosterList       *string              `json:"poster_list"`
	PosterWechat     *string              `json:"poster_wechat"`
	TicketSpecs      []TicketSpecSaveItem `json:"ticket_specs"`
	QualificationDoc *string              `json:"qualification_doc"`
}

type ActivityDetailResponse struct {
	models.Activity
	TicketSpecs []models.TicketSpec `json:"ticket_specs"`
	Organizer   *models.Organizer   `json:"organizer,omitempty"`
}

type ActivityListItem struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	PosterList string    `json:"poster_list"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	Status     int8      `json:"status"`
}

type TicketSpecSaveItem struct {
	ID            int64  `json:"id"`
	Name          string `json:"name" binding:"required"`
	IsEnabled     int8   `json:"is_enabled"`
	SaleStart     string `json:"sale_start"`
	SaleEnd       string `json:"sale_end"`
	Price         int64  `json:"price" binding:"required"` // 分
	Stock         int    `json:"stock" binding:"min=0"`
	PurchaseLimit int    `json:"purchase_limit"`
	MaxAttendees  int    `json:"max_attendees"`
}

type SaveTicketSpecsRequest struct {
	Specs []TicketSpecSaveItem `json:"specs" binding:"required"`
}

type CreateTicketOrderRequest struct {
	ActivityID   int64  `json:"activity_id" binding:"required"`
	TicketSpecID int64  `json:"ticket_spec_id" binding:"required"`
	Quantity     int    `json:"quantity" binding:"required,min=1"`
	BuyerName    string `json:"buyer_name"`
	BuyerIDCard  string `json:"buyer_id_card"`
}

type TicketOrderDetailResponse struct {
	OrderNo     string `json:"order_no"`
	Status      int8   `json:"status"`
	TotalPrice  int64  `json:"total_price"`
	ActualPrice int64  `json:"actual_price"`
	Quantity    int    `json:"quantity"`
	Activity    struct {
		ID         int64     `json:"id"`
		Name       string    `json:"name"`
		StartTime  time.Time `json:"start_time"`
		EndTime    time.Time `json:"end_time"`
		PosterList string    `json:"poster_list"`
	} `json:"activity"`
	TicketSpec struct {
		Name string `json:"name"`
	} `json:"ticket_spec"`
	BuyerName   string     `json:"buyer_name"`
	BuyerIDCard string     `json:"buyer_id_card"`
	PayMethod   string     `json:"pay_method"`
	PayTime     *time.Time `json:"pay_time"`
	CreatedAt   time.Time  `json:"created_at"`
	QRCode      string     `json:"qr_code"`
	ExpireTime  time.Time  `json:"expire_time"`
	RefundInfo  *struct {
		RefundAmount     int64  `json:"refund_amount"`
		Status           int8   `json:"status"`
		ExpectArriveDate string `json:"expect_arrive_date"`
	} `json:"refund_info,omitempty"`
}

type TicketOrderListItem struct {
	OrderNo     string `json:"order_no"`
	Status      int8   `json:"status"`
	TotalPrice  int64  `json:"total_price"`
	ActualPrice int64  `json:"actual_price"`
	Quantity    int    `json:"quantity"`
	Activity    struct {
		ID         int64     `json:"id"`
		Name       string    `json:"name"`
		StartTime  time.Time `json:"start_time"`
		EndTime    time.Time `json:"end_time"`
		PosterList string    `json:"poster_list"`
	} `json:"activity"`
	TicketSpec struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	} `json:"ticket_spec"`
	BuyerName   string     `json:"buyer_name"`
	BuyerIDCard string     `json:"buyer_id_card"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpireTime  time.Time  `json:"expire_time"`
	PayTime     *time.Time `json:"pay_time"`
}

type CancelOrderRequest struct {
	ReasonID int64 `json:"reason_id" binding:"required"`
}

type ApplyRefundRequest struct {
	OrderNo  string `json:"order_no" binding:"required"`
	ReasonID int64  `json:"reason_id" binding:"required"`
}

type RejectRefundRequest struct {
	RejectReason string `json:"reject_reason" binding:"required"`
}

type OrganizerRefundListItem struct {
	RefundNo         string    `json:"refund_no"`
	Status           int8      `json:"status"`
	RefundAmount     int64     `json:"refund_amount"`
	DeductAmount     int64     `json:"deduct_amount"`
	Reason           string    `json:"reason"`
	RejectReason     string    `json:"reject_reason"`
	ExpectArriveDate string    `json:"expect_arrive_date"`
	WechatRefundID   string    `json:"wechat_refund_id"`
	WechatStatus     string    `json:"wechat_status"`
	OrderNo          string    `json:"order_no"`
	UserID           int64     `json:"user_id"`
	BuyerName        string    `json:"buyer_name"`
	BuyerIDCard      string    `json:"buyer_id_card"`
	ActivityID       int64     `json:"activity_id"`
	ActivityName     string    `json:"activity_name"`
	TicketSpecID     int64     `json:"ticket_spec_id"`
	TicketSpecName   string    `json:"ticket_spec_name"`
	Quantity         int       `json:"quantity"`
	CreatedAt        time.Time `json:"created_at"`
}

type RefundDetailResponse struct {
	RefundNo         string `json:"refund_no"`
	Status           int8   `json:"status"`
	RefundAmount     int64  `json:"refund_amount"`
	DeductAmount     int64  `json:"deduct_amount"`
	ExpectArriveDate string `json:"expect_arrive_date"`
	Order            struct {
		OrderNo     string `json:"order_no"`
		TotalPrice  int64  `json:"total_price"`
		ActualPrice int64  `json:"actual_price"`
		Quantity    int    `json:"quantity"`
	} `json:"order"`
	Activity struct {
		Name      string    `json:"name"`
		StartTime time.Time `json:"start_time"`
		EndTime   time.Time `json:"end_time"`
	} `json:"activity"`
	Progress []models.RefundLog `json:"progress"`
}

type VerifierRequest struct {
	Name  string `json:"name" binding:"required"`
	Phone string `json:"phone" binding:"required"`
}

type ActivateVerifierRequest struct {
	Phone   string `json:"phone" binding:"required"`
	Channel string `json:"channel" binding:"required"`
}

type ScanOrderRequest struct {
	QRCode     string `json:"qr_code" binding:"required"`
	ActivityID int64  `json:"activity_id" binding:"required"`
}

type ScanOrderResponse struct {
	Success bool `json:"success"`
	Order   *struct {
		ActivityName      string `json:"activity_name"`
		TicketSpecName    string `json:"ticket_spec_name"`
		Quantity          int    `json:"quantity"`
		BuyerNameMasked   string `json:"buyer_name_masked"`
		BuyerIDCardMasked string `json:"buyer_id_card_masked"`
	} `json:"order,omitempty"`
	ErrorCode string `json:"error_code,omitempty"`
}

type ConfirmVerifyRequest struct {
	OrderNo string `json:"order_no" binding:"required"`
}

type VerifiedListItem struct {
	ActivityName      string    `json:"activity_name"`
	TicketSpecName    string    `json:"ticket_spec_name"`
	Quantity          int       `json:"quantity"`
	BuyerNameMasked   string    `json:"buyer_name_masked"`
	BuyerIDCardMasked string    `json:"buyer_id_card_masked"`
	VerifiedAt        time.Time `json:"verified_at"`
}

type ViewerItem struct {
	ID        int64     `json:"id"`
	RealName  string    `json:"real_name"`
	IDCard    string    `json:"id_card"`
	Phone     string    `json:"phone"`
	Type      int8      `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
