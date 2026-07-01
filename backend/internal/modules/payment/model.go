package payment

import "time"

type PaymentRecord struct {
	ID          int64
	PaymentNo   string
	OrderNo     string
	UserID      int64
	Amount      float64
	Status      string
	PayMethod   string
	PaidAt      *time.Time
	RefundAt    *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
