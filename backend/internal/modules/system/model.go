package system

import "time"

type SystemSetting struct {
	ID          int64
	Key         string
	Value       *string
	Type        *string
	Category    *string
	Description *string
	IsPublic    *int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type PartnerConfig struct {
	ID          int64
	Key         string
	Value       *string
	Description *string
	Status      *int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type FaceIdConfig struct {
	ID            int64
	SecretID      string
	SecretKey     string
	RuleID        string
	Region        *string
	RedirectURL   *string
	IsEnabled     *int
	Remark        *string
	ManualEnabled *int
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Deleted       *int
}

type RealNameVerifyLog struct {
	ID              int64
	UserID          int64
	EventType       string
	OperatorID      *int64
	OperatorType    *string
	Detail          *string
	IPAddress       *string
	SubmittedName   *string
	SubmittedIDCard *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Deleted         *int
}
