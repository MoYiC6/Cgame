package order

type OrderStatus string

const (
	OrderStatusPending    = "pending"
	OrderStatusPaid       = "paid"
	OrderStatusCompleted  = "completed"
	OrderStatusCancelled  = "cancelled"
	OrderStatusRefunding  = "refunding"
	OrderStatusRefunded   = "refunded"
	OrderStatusInTransfer = "in_transfer"
	OrderStatusFinalReview = "final_review"
)

type ReviewStatus string

const (
	ReviewStatusPending  = "pending"
	ReviewStatusApproved = "approved"
	ReviewStatusRejected = "rejected"
)

type ComplaintStatus string

const (
	ComplaintStatusPending  = "pending"
	ComplaintStatusResolved = "resolved"
	ComplaintStatusRejected = "rejected"
)
