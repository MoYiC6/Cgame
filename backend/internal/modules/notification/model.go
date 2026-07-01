package notification

import "time"

type Notification struct {
	ID          int64
	UserID      *int64
	Title       string
	Content     *string
	Type        string
	SubType     *string
	TargetType  *string
	TargetID    *int64
	RelatedID   *int64
	RelatedType *string
	ExtraData   map[string]any
	Priority    *int
	IsRead      bool
	ExpireTime  *time.Time
	ReadTime    *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type UserNotification struct {
	ID             int64
	UserID         int64
	NotificationID int64
	IsRead         bool
	ReadTime       *time.Time
	CreatedAt      time.Time
}

type SystemTodo struct {
	ID          int64
	Title       string
	Completed   bool
	SortOrder   *int
	CreatedBy   *string
	CompletedBy *string
	CompletedTime *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
