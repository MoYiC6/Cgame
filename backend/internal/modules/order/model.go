package order

import "time"

type Order struct {
	ID             int64
	OrderNo        string
	UserID         int64
	Status         string
	TotalAmount    float64
	PayAmount      float64
	DiscountAmount float64
	GoodsID        *int64
	SKUName        string
	Quantity       int
	Remark         *string
	PayAt          *time.Time
	CompletedAt    *time.Time
	CancelledAt    *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type OrderItem struct {
	ID        int64
	OrderID   int64
	GoodsID   int64
	SKUName   string
	Price     float64
	Quantity  int
	Subtotal  float64
	CreatedAt time.Time
}

// OrderReview 订单评价
type OrderReview struct {
	ID        int64
	OrderID   int64
	UserID    int64
	TeacherID *int64
	Rating    int
	Content   string
	Reply     *string
	Status    string // pending / approved / rejected
	CreatedAt time.Time
	UpdatedAt time.Time
}

// OrderComplaint 订单投诉
type OrderComplaint struct {
	ID        int64
	OrderID   int64
	UserID    int64
	Reason    string
	Detail    *string
	Status    string // pending / resolved / rejected
	CreatedAt time.Time
	UpdatedAt time.Time
}

// OrderTransferConfig 订单转移配置
type OrderTransferConfig struct {
	ID        int64
	Enabled   bool
	MaxTimes  int
	Timeout   int // minutes
	CreatedAt time.Time
	UpdatedAt time.Time
}

// OrderQuery 管理端订单查询参数
type OrderQuery struct {
	PageNum   int
	PageSize  int
	OrderNo   string
	UserID    *int64
	Status    string
	StartTime *time.Time
	EndTime   *time.Time
}

// OrderStats 订单统计
type OrderStats struct {
	TotalOrders   int
	TotalAmount   float64
	PaidOrders    int
	PaidAmount    float64
	PendingOrders int
	CancelledOrders int
}

// ReviewQuery 管理端评价查询参数
type ReviewQuery struct {
	PageNum  int
	PageSize int
	Status   string
	OrderID  *int64
}

// FinalReviewQuery 终审列表查询参数
type FinalReviewQuery struct {
	PageNum  int
	PageSize int
	Status   string
}

// PaymentRecord 支付记录
type PaymentRecord struct {
	ID          int64
	OrderID     int64
	UserID      int64
	OutTradeNo  string
	Channel     string // wxpay / alipay
	Amount      float64
	Status      string // pending / paid / failed / refunded
	PaidAt      *time.Time
	TransactionID *string
	NotifyRaw   *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CashierOrder 收银台订单
type CashierOrder struct {
	ID         int64
	Token      string
	OrderID    int64
	UserID     int64
	Amount     float64
	Status     string // pending / paid / expired / cancelled
	ExpireAt   time.Time
	PaidAt     *time.Time
	PayChannel *string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// PaymentSyncLog 支付同步日志
type PaymentSyncLog struct {
	ID        int64
	RecordID  int64
	Channel   string
	OutTradeNo string
	Action    string // query / notify / manual
	Request   *string
	Response  *string
	Success   bool
	ErrorMsg  *string
	CreatedAt time.Time
}

// PaymentQuery 支付记录查询参数
type PaymentQuery struct {
	PageNum    int
	PageSize   int
	OrderID    *int64
	UserID     *int64
	OutTradeNo string
	Status     string
	Channel    string
	StartTime  *time.Time
	EndTime    *time.Time
}

// WxPayConfig 微信支付配置
type WxPayConfig struct {
	AppID       string
	MchID       string
	APIKey      string
	NotifyURL   string
	AppSecret   string
}

// AlipayConfig 支付宝配置
type AlipayConfig struct {
	AppID      string
	PrivateKey string
	PublicKey  string
	NotifyURL  string
	ReturnURL  string
}
