package teacher

import "time"

type Teacher struct {
	ID                int64
	UserID            int64
	Name              *string
	Mobile            *string
	Avatar            *string
	Status            *int
	Rating            *float64
	OrderCount        int
	Deposit           *float64
	Balance           *float64
	Platforms         map[string]any
	Tags              []string
	GoodsIDs          []int64
	AutoStatusEnabled bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type TeacherLevel struct {
	ID             int64
	Name           string
	MinOrders      int
	CommissionRate float64
	Priority       int
	Status         *int
	Description    *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type TeacherLevelGoods struct {
	ID             int64
	LevelID        int64
	GoodsID        int64
	CommissionRate float64
	CreatedAt      time.Time
}

type TeacherStatusLog struct {
	ID         int64
	TeacherID  int64
	OldStatus  *int
	NewStatus  *int
	Reason     *string
	OperatorID *int64
	CreatedAt  time.Time
}

type TeacherBalanceLog struct {
	ID           int64
	TeacherID    int64
	ChangeType   string
	Amount       float64
	BalanceAfter float64
	RelatedID    *int64
	RelatedNo    *string
	Description  *string
	CreatedAt    time.Time
}

// TeacherApplication represents an application to become a teacher
type TeacherApplication struct {
	ID          int64
	UserID      int64
	Name        string
	Mobile      *string
	Avatar      *string
	Platforms   map[string]any
	Tags        []string
	Intro       *string
	Status      int // 0=pending, 1=approved, 2=rejected
	Reason      *string
	OperatorID  *int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TeacherRanking represents a teacher ranking entry
type TeacherRanking struct {
	TeacherID  int64
	Name       string
	Avatar     *string
	Rating     float64
	OrderCount int
	Balance    float64
	Rank       int
}

// TeacherDashboardStats represents teacher dashboard statistics
type TeacherDashboardStats struct {
	TodayIncome     float64
	WeekIncome      float64
	MonthIncome     float64
	TotalIncome     float64
	TodayOrders     int
	WeekOrders      int
	MonthOrders     int
	TotalOrders     int
	PendingOrders   int
	Rating          float64
	CompletionRate  float64
}

// TeacherOnlineStatus represents a teacher's online status update
type TeacherOnlineStatus struct {
	TeacherID int64
	Status    int // 1=online, 2=offline, 3=busy
}

// TeacherIntro represents a teacher's introduction
type TeacherIntro struct {
	TeacherID int64
	Intro     string
	Tags      []string
}

// TeacherPaymentInfo represents a teacher's payment information
type TeacherPaymentInfo struct {
	TeacherID     int64
	AlipayAccount *string
	BankName      *string
	BankAccount   *string
	RealName      *string
}

// TeacherAutoStatus represents auto status configuration
type TeacherAutoStatus struct {
	TeacherID     int64
	Enabled       bool
	OnlineTime    *string
	OfflineTime   *string
}

// TeacherHeartbeat represents a heartbeat from a teacher
type TeacherHeartbeat struct {
	TeacherID int64
	Timestamp time.Time
}
