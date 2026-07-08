package feedback

import "time"

const (
	FeedbackStatusPending    = 0
	FeedbackStatusProcessing = 1
	FeedbackStatusResolved   = 2
	ReplyTypeUser            = 0
	ReplyTypeAdmin           = 1
)

type Feedback struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"userId"`
	TicketNo  string    `json:"ticketNo"`
	Content   string    `json:"content"`
	Images    []string  `json:"images"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type FeedbackReply struct {
	ID          int64     `json:"id"`
	FeedbackID  int64     `json:"feedbackId"`
	ReplyUserID int64     `json:"replyUserId"`
	ReplyType   int       `json:"replyType"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"createdAt"`
}

type FeedbackVO struct {
	ID         int64     `json:"id"`
	TicketNo   string    `json:"ticketNo"`
	UserID     int64     `json:"userId,omitempty"`
	Content    string    `json:"content"`
	Images     []string  `json:"images"`
	Status     int       `json:"status"`
	StatusText string    `json:"statusText"`
	ReplyCount int       `json:"replyCount"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type FeedbackDetailVO struct {
	ID         int64             `json:"id"`
	TicketNo   string            `json:"ticketNo"`
	Content    string            `json:"content"`
	Images     []string          `json:"images"`
	Status     int               `json:"status"`
	StatusText string            `json:"statusText"`
	CreatedAt  time.Time         `json:"createdAt"`
	UpdatedAt  time.Time         `json:"updatedAt"`
	Replies    []FeedbackReplyVO `json:"replies"`
}

type FeedbackReplyVO struct {
	ID            int64     `json:"id"`
	ReplyType     int       `json:"replyType"`
	ReplyTypeText string    `json:"replyTypeText"`
	Content       string    `json:"content"`
	ReplyUserName string    `json:"replyUserName,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
}

type FeedbackPage struct {
	Records []FeedbackVO `json:"records"`
	Total   int64        `json:"total"`
}

type SubmitRequest struct {
	Content string   `json:"content"`
	Images  []string `json:"images"`
}

func statusText(status int) string {
	switch status {
	case FeedbackStatusPending:
		return "待处理"
	case FeedbackStatusProcessing:
		return "处理中"
	case FeedbackStatusResolved:
		return "已解决"
	default:
		return "未知状态"
	}
}

func replyTypeText(replyType int) string {
	if replyType == ReplyTypeAdmin {
		return "管理员"
	}
	return "用户"
}
