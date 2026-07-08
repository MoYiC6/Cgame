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

type NotificationInboxList struct {
	Total       int64                   `json:"total"`
	UnreadCount int64                   `json:"unreadCount"`
	Rows        []NotificationInboxItem `json:"rows"`
}

type NotificationInboxItem struct {
	ID             int64      `json:"id"`
	NotificationID int64      `json:"notificationId"`
	Title          string     `json:"title"`
	Content        *string    `json:"content"`
	Type           string     `json:"type"`
	TargetType     *string    `json:"targetType"`
	TargetID       *string    `json:"targetId"`
	IsRead         int16      `json:"isRead"`
	ReadTime       *time.Time `json:"readTime"`
	SendTime       *time.Time `json:"sendTime"`
	CreateTime     time.Time  `json:"createTime"`
}

type NotificationStats struct {
	TotalNotifications int64 `json:"totalNotifications"`
	UnreadCount        int64 `json:"unreadCount"`
	TodayCount         int64 `json:"todayCount"`
}

type SubscribeTemplate struct {
	ID          int64  `json:"id"`
	TemplateID  string `json:"templateId"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Enabled     bool   `json:"enabled"`
}

type SubscribeStatus struct {
	TemplateID string `json:"templateId"`
	Subscribed bool   `json:"subscribed"`
	SubscribedAt *time.Time `json:"subscribedAt,omitempty"`
}

type SystemNotification struct {
	ID        int64      `json:"id"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	Type      string     `json:"type"`
	Priority  int        `json:"priority"`
	CreatedAt time.Time  `json:"createdAt"`
}

type AdminNotificationQuery struct {
	PageNum  int    `json:"pageNum"`
	PageSize int    `json:"pageSize"`
	Type     string `json:"type"`
}

type SubscribeRecordRequest struct {
	TemplateID string `json:"templateId"`
}

type SystemTodo struct {
	ID            int64
	Title         string
	Completed     bool
	SortOrder     *int
	CreatedBy     *string
	CompletedBy   *string
	CompletedTime *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
