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
	ID            int64
	LevelID       int64
	GoodsID       int64
	CommissionRate float64
	CreatedAt     time.Time
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
