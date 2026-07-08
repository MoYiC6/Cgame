package payment

import "time"

type PaymentRecord struct {
	ID        int64
	PaymentNo string
	OrderNo   string
	UserID    int64
	Amount    float64
	Status    string
	PayMethod string
	PaidAt    *time.Time
	RefundAt  *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

type PaymentStats struct {
	TotalAmount      float64            `json:"totalAmount"`
	TodayAmount      float64            `json:"todayAmount"`
	MonthAmount      float64            `json:"monthAmount"`
	PaidCount        int                `json:"paidCount"`
	PendingCount     int                `json:"pendingCount"`
	RefundedCount    int                `json:"refundedCount"`
	PayMethodAmounts map[string]float64 `json:"payMethodAmounts"`
}
