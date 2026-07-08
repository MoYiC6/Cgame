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
