package order

type CreateOrderRequest struct {
	UserID int64  `json:"user_id"`
	SKU    string `json:"sku"`
	Qty    int    `json:"qty"`
}

type OrderResponse struct {
	OrderNo string `json:"order_no"`
	Status  string `json:"status"`
}

type PingResponse struct {
	Module  string `json:"module"`
	TraceID string `json:"trace_id,omitempty"`
}
