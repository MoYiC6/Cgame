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
