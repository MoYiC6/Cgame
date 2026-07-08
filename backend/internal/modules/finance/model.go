package finance

import "time"

type OperatorCommission struct {
	ID         int64      `json:"id"`
	OperatorID int64      `json:"operatorId"`
	OrderID    *int64     `json:"orderId,omitempty"`
	Amount     float64    `json:"amount"`
	Balance    float64    `json:"balance"`
	Status     string     `json:"status"`
	Remark     string     `json:"remark,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}

type OperatorWithdrawal struct {
	ID          int64      `json:"id"`
	OperatorID  int64      `json:"operatorId"`
	Amount      float64    `json:"amount"`
	Status      string     `json:"status"`
	AdminRemark string     `json:"adminRemark,omitempty"`
	ProcessedBy *int64     `json:"processedBy,omitempty"`
	ProcessedAt *time.Time `json:"processedAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type BalanceDetail struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"userId"`
	ChangeType   string    `json:"changeType"`
	Amount       float64   `json:"amount"`
	Balance      float64   `json:"balance"`
	Remark       string    `json:"remark,omitempty"`
	RelatedID    *int64    `json:"relatedId,omitempty"`
	RelatedType  string    `json:"relatedType,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
}

type MonthlyReport struct {
	Month        string  `json:"month"`
	TotalRevenue float64 `json:"totalRevenue"`
	TotalOrders  int     `json:"totalOrders"`
	Commission   float64 `json:"commission"`
}

type FinanceStats struct {
	TotalRevenue      float64        `json:"totalRevenue"`
	TodayRevenue      float64        `json:"todayRevenue"`
	MonthRevenue      float64        `json:"monthRevenue"`
	YearRevenue       float64        `json:"yearRevenue"`
	TeacherCommission float64        `json:"teacherCommission"`
	PlatformRevenue   float64        `json:"platformRevenue"`
	PendingSettlement float64        `json:"pendingSettlement"`
	MonthlyTrend      []MonthlyTrend `json:"monthlyTrend"`
}

type MonthlyTrend struct {
	Month   string  `json:"month"`
	Revenue float64 `json:"revenue"`
}
