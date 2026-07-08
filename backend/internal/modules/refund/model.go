package refund

import "time"

const (
	RefundStatusPending   = "pending"
	RefundStatusApproved  = "approved"
	RefundStatusRejected  = "rejected"
	RefundStatusProcessed = "processed"
	RefundStatusCancelled = "cancelled"
)

// Refund represents a refund request in the database.
type Refund struct {
	ID           int64      `json:"id"`
	RefundNo     string     `json:"refundNo"`
	OrderID      int64      `json:"orderId"`
	UserID       int64      `json:"userId"`
	Amount       float64    `json:"amount"`
	Reason       string     `json:"reason"`
	Status       string     `json:"status"`
	AdminRemark  string     `json:"adminRemark,omitempty"`
	ProcessedBy  *int64     `json:"processedBy,omitempty"`
	ProcessedAt  *time.Time `json:"processedAt,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

// RefundVO is the client view of a refund.
type RefundVO struct {
	ID          int64      `json:"id"`
	RefundNo    string     `json:"refundNo"`
	OrderID     int64      `json:"orderId"`
	Amount      float64    `json:"amount"`
	Reason      string     `json:"reason"`
	Status      string     `json:"status"`
	StatusDesc  string     `json:"statusDesc"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

// AdminRefundVO is the admin view of a refund.
type AdminRefundVO struct {
	ID           int64      `json:"id"`
	RefundNo     string     `json:"refundNo"`
	OrderID      int64      `json:"orderId"`
	UserID       int64      `json:"userId"`
	Amount       float64    `json:"amount"`
	Reason       string     `json:"reason"`
	Status       string     `json:"status"`
	StatusDesc   string     `json:"statusDesc"`
	AdminRemark  string     `json:"adminRemark,omitempty"`
	ProcessedBy  *int64     `json:"processedBy,omitempty"`
	ProcessedAt  *time.Time `json:"processedAt,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

// RefundPageResult wraps paginated refunds.
type RefundPageResult struct {
	Total    int64         `json:"total"`
	PageNum  int           `json:"pageNum"`
	PageSize int           `json:"pageSize"`
	Records  []RefundVO    `json:"records"`
}

// AdminRefundPageResult wraps paginated admin refunds.
type AdminRefundPageResult struct {
	Total    int64           `json:"total"`
	PageNum  int             `json:"pageNum"`
	PageSize int             `json:"pageSize"`
	Records  []AdminRefundVO `json:"records"`
}

// RefundStats holds refund statistics.
type RefundStats struct {
	TotalRefunds    int     `json:"totalRefunds"`
	PendingRefunds  int     `json:"pendingRefunds"`
	ApprovedRefunds int     `json:"approvedRefunds"`
	RejectedRefunds int     `json:"rejectedRefunds"`
	ProcessedRefunds int    `json:"processedRefunds"`
	TotalAmount     float64 `json:"totalAmount"`
}

// ApplyRequest is the client request to apply for a refund.
type ApplyRequest struct {
	OrderID int64   `json:"orderId"`
	Amount  float64 `json:"amount"`
	Reason  string  `json:"reason"`
}

// RefundQuery is the admin query for listing refunds.
type RefundQuery struct {
	PageNum     int     `json:"pageNum"`
	PageSize    int     `json:"pageSize"`
	Status      string  `json:"status"`
	OrderID     *int64  `json:"orderId"`
	UserID      *int64  `json:"userId"`
	RefundNo    string  `json:"refundNo"`
	CreateTimeStart *string `json:"createTimeStart"`
	CreateTimeEnd   *string `json:"createTimeEnd"`
}

// CanApplyResult tells whether a refund can be applied for an order.
type CanApplyResult struct {
	CanApply bool   `json:"canApply"`
	Reason   string `json:"reason,omitempty"`
}

func statusDesc(status string) string {
	switch status {
	case RefundStatusPending:
		return "待处理"
	case RefundStatusApproved:
		return "已批准"
	case RefundStatusRejected:
		return "已拒绝"
	case RefundStatusProcessed:
		return "已处理"
	case RefundStatusCancelled:
		return "已取消"
	default:
		return "未知状态"
	}
}

func toRefundVO(r Refund) RefundVO {
	return RefundVO{
		ID:         r.ID,
		RefundNo:   r.RefundNo,
		OrderID:    r.OrderID,
		Amount:     r.Amount,
		Reason:     r.Reason,
		Status:     r.Status,
		StatusDesc: statusDesc(r.Status),
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}
}

func toAdminRefundVO(r Refund) AdminRefundVO {
	return AdminRefundVO{
		ID:          r.ID,
		RefundNo:    r.RefundNo,
		OrderID:     r.OrderID,
		UserID:      r.UserID,
		Amount:      r.Amount,
		Reason:      r.Reason,
		Status:      r.Status,
		StatusDesc:  statusDesc(r.Status),
		AdminRemark: r.AdminRemark,
		ProcessedBy: r.ProcessedBy,
		ProcessedAt: r.ProcessedAt,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}
