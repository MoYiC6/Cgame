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
