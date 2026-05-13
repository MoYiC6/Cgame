package payment

type CreatePaymentRequest struct {
	OrderNo string `json:"order_no"`
	Amount  int64  `json:"amount"`
	Channel string `json:"channel"`
}

type PaymentResponse struct {
	PaymentNo string `json:"payment_no"`
	Status    string `json:"status"`
}

type PingResponse struct {
	Module  string `json:"module"`
	TraceID string `json:"trace_id,omitempty"`
}
