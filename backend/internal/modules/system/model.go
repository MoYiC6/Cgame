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

// RBAC Models

type SystemMenu struct {
	ID             int64
	Name           string
	Path           *string
	Component      *string
	Icon           *string
	Sort           *int
	ParentID       *int64
	Status         *int
	MenuType       *string
	PermissionCode *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Children       []*SystemMenu `json:"children,omitempty"`
}

type SystemPermission struct {
	ID          int64
	Code        string
	Name        string
	Description *string
	Status      *int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type SystemRole struct {
	ID          int64
	Name        string
	Code        string
	Description *string
	Status      *int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type RolePermission struct {
	RoleID       int64
	PermissionID int64
	CreatedAt    time.Time
}

type RoleMenu struct {
	RoleID    int64
	MenuID    int64
	CreatedAt time.Time
}
