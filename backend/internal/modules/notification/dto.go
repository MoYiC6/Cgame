package notification

type SendNotificationRequest struct {
	Channel string `json:"channel"`
	Target  string `json:"target"`
	Body    string `json:"body"`
}

type NotificationResponse struct {
	TaskNo  string `json:"task_no"`
	Status  string `json:"status"`
	Channel string `json:"channel"`
}

type PingResponse struct {
	Module  string `json:"module"`
	TraceID string `json:"trace_id,omitempty"`
}
