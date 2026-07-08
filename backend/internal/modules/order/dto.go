package order

import "time"

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

// ComplaintRequest 投诉请求
type ComplaintRequest struct {
	Reason string  `json:"reason"`
	Detail *string `json:"detail,omitempty"`
}

// ConfirmTeacherRequest 确认选手请求
type ConfirmTeacherRequest struct {
	TeacherID int64 `json:"teacherId"`
}

// ReviewRequest 评价请求
type ReviewRequest struct {
	Rating  int     `json:"rating"`
	Content string  `json:"content"`
	TeacherID *int64 `json:"teacherId,omitempty"`
}

// ReplyReviewRequest 回复评价请求
type ReplyReviewRequest struct {
	Reply string `json:"reply"`
}

// UpdateStatusRequest 更新订单状态请求
type UpdateStatusRequest struct {
	Status string `json:"status"`
}

// UpdateRemarkRequest 更新备注请求
type UpdateRemarkRequest struct {
	Remark string `json:"remark"`
}

// UpdateTeachersRequest 更新关联选手请求
type UpdateTeachersRequest struct {
	TeacherIDs []int64 `json:"teacherIds"`
}

// ManualOrderRequest 手动下单请求
type ManualOrderRequest struct {
	UserID    int64   `json:"userId"`
	GoodsID   *int64  `json:"goodsId,omitempty"`
	SKUName   string  `json:"skuName"`
	Quantity  int     `json:"quantity"`
	TotalAmount float64 `json:"totalAmount"`
	Remark    *string `json:"remark,omitempty"`
}

// TransferRequest 执行订单转移请求
type TransferRequest struct {
	TargetTeacherID int64 `json:"targetTeacherId"`
}

// OrderListResponse 订单列表响应
type OrderListResponse struct {
	List  []*Order `json:"list"`
	Total int      `json:"total"`
}

// ReviewListResponse 评价列表响应
type ReviewListResponse struct {
	List  []*OrderReview `json:"list"`
	Total int            `json:"total"`
}

// FinalReviewListResponse 终审列表响应
type FinalReviewListResponse struct {
	List  []*Order `json:"list"`
	Total int      `json:"total"`
}

// CreatePaymentRequest 创建支付请求
type CreatePaymentRequest struct {
	OrderID int64   `json:"orderId"`
	Channel string  `json:"channel"` // wxpay / alipay
	Amount  float64 `json:"amount"`
}

// CashierPayRequest 收银台支付请求
type CashierPayRequest struct {
	Token   string `json:"token"`
	Channel string `json:"channel"` // wxpay / alipay
}

// WxPayOrderResponse 微信支付订单响应
type WxPayOrderResponse struct {
	AppID     string `json:"appId"`
	PartnerID string `json:"partnerId"`
	PrepayID  string `json:"prepayId"`
	NonceStr  string `json:"nonceStr"`
	TimeStamp string `json:"timeStamp"`
	Package   string `json:"package"`
	Sign      string `json:"sign"`
}

// AlipayOrderResponse 支付宝订单响应
type AlipayOrderResponse struct {
	OrderStr string `json:"orderStr"`
}

// PaymentRecordResponse 支付记录响应
type PaymentRecordResponse struct {
	ID            int64      `json:"id"`
	OrderID       int64      `json:"orderId"`
	OutTradeNo    string     `json:"outTradeNo"`
	Channel       string     `json:"channel"`
	Amount        float64    `json:"amount"`
	Status        string     `json:"status"`
	PaidAt        *time.Time `json:"paidAt,omitempty"`
	TransactionID *string    `json:"transactionId,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
}

// PaymentListResponse 支付记录列表响应
type PaymentListResponse struct {
	List  []*PaymentRecordResponse `json:"list"`
	Total int                      `json:"total"`
}

// SyncPaymentRequest 手动同步支付请求
type SyncPaymentRequest struct {
	OutTradeNo string `json:"outTradeNo"`
	Channel    string `json:"channel"`
}

// BatchSyncPaymentRequest 批量同步支付请求
type BatchSyncPaymentRequest struct {
	IDs []int64 `json:"ids"`
}

// CashierOrderResponse 收银台订单响应
type CashierOrderResponse struct {
	Token     string     `json:"token"`
	OrderID   int64      `json:"orderId"`
	Amount    float64    `json:"amount"`
	Status    string     `json:"status"`
	ExpireAt  time.Time  `json:"expireAt"`
	PayURL    string     `json:"payUrl"`
	CreatedAt time.Time  `json:"createdAt"`
}
