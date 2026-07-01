package finance

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
