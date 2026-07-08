package withdrawal

import "time"

const (
	WithdrawalStatusPending   = "pending"
	WithdrawalStatusApproved  = "approved"
	WithdrawalStatusRejected  = "rejected"
	WithdrawalStatusPaid      = "paid"
	WithdrawalStatusCancelled = "cancelled"
)

// Withdrawal represents a withdrawal request in the database.
type Withdrawal struct {
	ID            int64      `json:"id"`
	WithdrawalNo  string     `json:"withdrawalNo"`
	TeacherID     int64      `json:"teacherId"`
	Amount        float64    `json:"amount"`
	TaxAmount     float64    `json:"taxAmount"`
	ActualAmount  float64    `json:"actualAmount"`
	Status        string     `json:"status"`
	BankName      string     `json:"bankName,omitempty"`
	BankAccount   string     `json:"bankAccount,omitempty"`
	AccountName   string     `json:"accountName,omitempty"`
	AlipayAccount string     `json:"alipayAccount,omitempty"`
	Remark        string     `json:"remark,omitempty"`
	AdminRemark   string     `json:"adminRemark,omitempty"`
	ProcessedBy   *int64     `json:"processedBy,omitempty"`
	ProcessedAt   *time.Time `json:"processedAt,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

// WithdrawalVO is the client view of a withdrawal.
type WithdrawalVO struct {
	ID            int64      `json:"id"`
	WithdrawalNo  string     `json:"withdrawalNo"`
	Amount        float64    `json:"amount"`
	TaxAmount     float64    `json:"taxAmount"`
	ActualAmount  float64    `json:"actualAmount"`
	Status        string     `json:"status"`
	StatusDesc    string     `json:"statusDesc"`
	BankName      string     `json:"bankName,omitempty"`
	BankAccount   string     `json:"bankAccount,omitempty"`
	AccountName   string     `json:"accountName,omitempty"`
	AlipayAccount string     `json:"alipayAccount,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
}

// AdminWithdrawalVO is the admin view of a withdrawal.
type AdminWithdrawalVO struct {
	ID            int64      `json:"id"`
	WithdrawalNo  string     `json:"withdrawalNo"`
	TeacherID     int64      `json:"teacherId"`
	Amount        float64    `json:"amount"`
	TaxAmount     float64    `json:"taxAmount"`
	ActualAmount  float64    `json:"actualAmount"`
	Status        string     `json:"status"`
	StatusDesc    string     `json:"statusDesc"`
	BankName      string     `json:"bankName,omitempty"`
	BankAccount   string     `json:"bankAccount,omitempty"`
	AccountName   string     `json:"accountName,omitempty"`
	AlipayAccount string     `json:"alipayAccount,omitempty"`
	Remark        string     `json:"remark,omitempty"`
	AdminRemark   string     `json:"adminRemark,omitempty"`
	ProcessedBy   *int64     `json:"processedBy,omitempty"`
	ProcessedAt   *time.Time `json:"processedAt,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

// WithdrawalPageResult wraps paginated withdrawals.
type WithdrawalPageResult struct {
	Total    int64          `json:"total"`
	PageNum  int            `json:"pageNum"`
	PageSize int            `json:"pageSize"`
	Records  []WithdrawalVO `json:"records"`
}

// AdminWithdrawalPageResult wraps paginated admin withdrawals.
type AdminWithdrawalPageResult struct {
	Total    int64               `json:"total"`
	PageNum  int                 `json:"pageNum"`
	PageSize int                 `json:"pageSize"`
	Records  []AdminWithdrawalVO `json:"records"`
}

// WithdrawalStats holds withdrawal statistics.
type WithdrawalStats struct {
	TotalWithdrawals   int     `json:"totalWithdrawals"`
	PendingWithdrawals int     `json:"pendingWithdrawals"`
	ApprovedWithdrawals int    `json:"approvedWithdrawals"`
	PaidWithdrawals    int     `json:"paidWithdrawals"`
	TotalAmount        float64 `json:"totalAmount"`
}

// IncomeStats holds teacher income statistics.
type IncomeStats struct {
	TotalIncome      float64 `json:"totalIncome"`
	SettledIncome    float64 `json:"settledIncome"`
	UnsettledIncome  float64 `json:"unsettledIncome"`
	WithdrawnAmount  float64 `json:"withdrawnAmount"`
	PendingWithdrawal float64 `json:"pendingWithdrawal"`
}

// ApplyRequest is the client request to apply for withdrawal.
type ApplyRequest struct {
	Amount        float64 `json:"amount"`
	BankName      string  `json:"bankName"`
	BankAccount   string  `json:"bankAccount"`
	AccountName   string  `json:"accountName"`
	AlipayAccount string  `json:"alipayAccount"`
	Remark        string  `json:"remark"`
}

// CalculateRequest is the request to calculate withdrawal amount.
type CalculateRequest struct {
	Amount float64 `json:"amount"`
}

// CalculateResult is the result of withdrawal calculation.
type CalculateResult struct {
	Amount       float64 `json:"amount"`
	TaxAmount    float64 `json:"taxAmount"`
	ActualAmount float64 `json:"actualAmount"`
	TaxRate      float64 `json:"taxRate"`
}

// WithdrawalQuery is the admin query for listing withdrawals.
type WithdrawalQuery struct {
	PageNum     int     `json:"pageNum"`
	PageSize    int     `json:"pageSize"`
	Status      string  `json:"status"`
	TeacherID   *int64  `json:"teacherId"`
	WithdrawalNo string `json:"withdrawalNo"`
	CreateTimeStart *string `json:"createTimeStart"`
	CreateTimeEnd   *string `json:"createTimeEnd"`
}

func statusDesc(status string) string {
	switch status {
	case WithdrawalStatusPending:
		return "待处理"
	case WithdrawalStatusApproved:
		return "已批准"
	case WithdrawalStatusRejected:
		return "已拒绝"
	case WithdrawalStatusPaid:
		return "已打款"
	case WithdrawalStatusCancelled:
		return "已取消"
	default:
		return "未知状态"
	}
}

func toWithdrawalVO(w Withdrawal) WithdrawalVO {
	return WithdrawalVO{
		ID:            w.ID,
		WithdrawalNo:  w.WithdrawalNo,
		Amount:        w.Amount,
		TaxAmount:     w.TaxAmount,
		ActualAmount:  w.ActualAmount,
		Status:        w.Status,
		StatusDesc:    statusDesc(w.Status),
		BankName:      w.BankName,
		BankAccount:   w.BankAccount,
		AccountName:   w.AccountName,
		AlipayAccount: w.AlipayAccount,
		CreatedAt:     w.CreatedAt,
	}
}

func toAdminWithdrawalVO(w Withdrawal) AdminWithdrawalVO {
	return AdminWithdrawalVO{
		ID:            w.ID,
		WithdrawalNo:  w.WithdrawalNo,
		TeacherID:     w.TeacherID,
		Amount:        w.Amount,
		TaxAmount:     w.TaxAmount,
		ActualAmount:  w.ActualAmount,
		Status:        w.Status,
		StatusDesc:    statusDesc(w.Status),
		BankName:      w.BankName,
		BankAccount:   w.BankAccount,
		AccountName:   w.AccountName,
		AlipayAccount: w.AlipayAccount,
		Remark:        w.Remark,
		AdminRemark:   w.AdminRemark,
		ProcessedBy:   w.ProcessedBy,
		ProcessedAt:   w.ProcessedAt,
		CreatedAt:     w.CreatedAt,
		UpdatedAt:     w.UpdatedAt,
	}
}
