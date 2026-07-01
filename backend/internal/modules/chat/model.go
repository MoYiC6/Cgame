package chat

import "time"

type ChatSession struct {
	ID                 int64
	UserID             int64
	TeacherID          *int64
	TeacherUserID      *int64
	OrderID            *int64
	LastMessageID      *int64
	LastMessageContent *string
	LastMessageTime    *time.Time
	UserUnreadCount    int
	TeacherUnreadCount int
	Status             *int
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type ChatMessage struct {
	ID          int64
	SessionID   int64
	SenderID    int64
	SenderType  string
	ReceiverID  *int64
	Content     string
	MessageType *string
	ExtraData   map[string]any
	IsRead      bool
	ReadTime    *time.Time
	Status      *int
	CreatedAt   time.Time
}
