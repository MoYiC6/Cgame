package order

type OrderStatus string

const (
	OrderStatusPending   = "pending"
	OrderStatusPaid      = "paid"
	OrderStatusCompleted = "completed"
	OrderStatusCancelled = "cancelled"
)
