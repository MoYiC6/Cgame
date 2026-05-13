package inventory

type InventoryResponse struct {
	SKU       string `json:"sku"`
	Available int    `json:"available"`
}

type PingResponse struct {
	Module  string `json:"module"`
	TraceID string `json:"trace_id,omitempty"`
}
