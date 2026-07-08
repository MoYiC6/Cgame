package order

import (
	"net/http"
	"strconv"

	apperrors "backend/internal/platform/errors"
	"backend/internal/platform/response"
	"backend/internal/platform/security"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service        Service
	authMiddleware gin.HandlerFunc
}

func NewHandler(service Service, authMiddleware gin.HandlerFunc) *Handler {
	return &Handler{service: service, authMiddleware: authMiddleware}
}

func (h *Handler) RegisterRoutes(group *gin.RouterGroup) {
	group.GET("/order/ping", h.Ping)

	client := group.Group("/client/order")
	if h.authMiddleware != nil {
		client.Use(h.authMiddleware)
	}
	{
		client.POST("/create", h.CreateOrder)
		client.GET("/:orderId", h.GetOrder)
		client.GET("/list", h.ListOrders)
		client.POST("/:orderId/pay", h.PayOrder)
		client.POST("/:orderId/complete", h.CompleteOrder)
		client.POST("/:orderId/cancel", h.CancelOrder)
		client.POST("/:orderId/complaint", h.ComplaintOrder)
		client.POST("/:orderId/confirm-teacher", h.ConfirmTeacher)
		client.GET("/statistics", h.OrderStatistics)
	}

	compatibleClient := group.Group("/client/orders")
	if h.authMiddleware != nil {
		compatibleClient.Use(h.authMiddleware)
	}
	{
		compatibleClient.POST("", h.CreateOrder)
		compatibleClient.GET("", h.ListOrders)
		compatibleClient.GET("/:orderId", h.GetOrder)
		compatibleClient.POST("/:orderId/cancel", h.CancelOrder)
		compatibleClient.POST("/:orderId/confirm", h.CompleteOrder)
		compatibleClient.POST("/:orderId/complaint", h.ComplaintOrder)
		compatibleClient.POST("/:orderId/confirm-teacher", h.ConfirmTeacher)
	}

	// Reviews client
	clientReviews := group.Group("/client/reviews/orders")
	if h.authMiddleware != nil {
		clientReviews.Use(h.authMiddleware)
	}
	{
		clientReviews.GET("", h.ListReviews)
		clientReviews.GET("/:orderId", h.GetReviewByOrder)
		clientReviews.POST("/:orderId", h.CreateReview)
	}

	// Transfer
	transfer := group.Group("/order-transfer")
	if h.authMiddleware != nil {
		transfer.Use(h.authMiddleware)
	}
	{
		transfer.GET("", h.GetTransferConfig)
		transfer.POST("/transfer", h.TransferOrder)
	}

	// Admin
	admin := group.Group("/admin/orders")
	if h.authMiddleware != nil {
		admin.Use(h.authMiddleware)
	}
	{
		admin.GET("", h.AdminListOrders)
		admin.GET("/:id", h.AdminGetOrder)
		admin.PUT("/:id/status", h.AdminUpdateStatus)
		admin.POST("/:id/refund", h.AdminRefund)
		admin.POST("/:id/manual-complete", h.AdminManualComplete)
		admin.PUT("/:id/remark", h.AdminUpdateRemark)
		admin.PUT("/:id/teachers", h.AdminUpdateTeachers)
		admin.POST("/manual", h.AdminManualCreate)
		admin.GET("/stats", h.AdminStats)
	}

	adminReviews := group.Group("/admin/reviews")
	if h.authMiddleware != nil {
		adminReviews.Use(h.authMiddleware)
	}
	{
		adminReviews.GET("", h.AdminListReviews)
		adminReviews.PUT("/:id/status", h.AdminUpdateReviewStatus)
		adminReviews.POST("/:id/reply", h.AdminReplyReview)
	}

	adminFinal := group.Group("/admin/orders/final-review")
	if h.authMiddleware != nil {
		adminFinal.Use(h.authMiddleware)
	}
	{
		adminFinal.GET("", h.AdminListFinalReview)
		adminFinal.POST("/:id/approve", h.AdminApproveFinalReview)
		adminFinal.POST("/:id/reject", h.AdminRejectFinalReview)
	}

	// Payment / Cashier
	cashier := group.Group("/cashier")
	if h.authMiddleware != nil {
		cashier.Use(h.authMiddleware)
	}
	{
		cashier.POST("/create", h.CreateCashierOrder)
		cashier.GET("/:token", h.GetCashierOrder)
		cashier.POST("/:token/pay", h.CashierPay)
	}

	payment := group.Group("/payment")
	if h.authMiddleware != nil {
		payment.Use(h.authMiddleware)
	}
	{
		payment.POST("/wxpay/create", h.CreateWxPayOrder)
		payment.POST("/wxpay/notify", h.WxPayNotify)
		payment.GET("/wxpay/query", h.QueryWxPay)
		payment.POST("/alipay/create", h.CreateAlipayOrder)
		payment.POST("/alipay/notify", h.AlipayNotify)
		payment.GET("/alipay/query", h.QueryAlipay)
		payment.POST("/sync", h.ManualSyncPayment)
		payment.POST("/batch-sync", h.BatchSyncPayments)
		payment.POST("/sync-overdue", h.SyncOverduePayments)
	}
}

func currentUserID(c *gin.Context) (int64, bool) {
	if principal, ok := security.PrincipalFromContext(c.Request.Context()); ok {
		userID, err := strconv.ParseInt(principal.UserID, 10, 64)
		if err == nil && userID != 0 {
			return userID, true
		}
	}
	userID, err := strconv.ParseInt(c.GetString("userID"), 10, 64)
	if err != nil || userID == 0 {
		return 0, false
	}
	return userID, true
}

func intQuery(c *gin.Context, key string, fallback int) int {
	value, err := strconv.Atoi(c.DefaultQuery(key, strconv.Itoa(fallback)))
	if err != nil {
		return fallback
	}
	return value
}

func (h *Handler) Ping(c *gin.Context) {
	response.Success(c, gin.H{"module": "order"})
}

func (h *Handler) CreateOrder(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	var req struct {
		SKUName  string  `json:"skuName"`
		Quantity int     `json:"quantity"`
		Remark   *string `json:"remark,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	order, err := h.service.CreateOrder(c.Request.Context(), userID, req.SKUName, req.Quantity, req.Remark)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, order)
}

func (h *Handler) GetOrder(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("orderId"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	order, err := h.service.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, order)
}

func (h *Handler) ListOrders(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	page := intQuery(c, "page", 1)
	pageSize := intQuery(c, "pageSize", 20)
	orders, total, err := h.service.ListOrders(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"list": orders, "total": total})
}

func (h *Handler) PayOrder(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("orderId"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.PayOrder(c.Request.Context(), orderID); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) CompleteOrder(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("orderId"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.CompleteOrder(c.Request.Context(), orderID); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) CancelOrder(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("orderId"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.CancelOrder(c.Request.Context(), orderID); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) ComplaintOrder(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	orderID, err := strconv.ParseInt(c.Param("orderId"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	var req ComplaintRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.ComplaintOrder(c.Request.Context(), orderID, userID, req.Reason, req.Detail); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) ConfirmTeacher(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("orderId"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	var req ConfirmTeacherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.ConfirmTeacher(c.Request.Context(), orderID, req.TeacherID); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) OrderStatistics(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	stats, err := h.service.GetOrderStatistics(c.Request.Context(), userID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, stats)
}

func (h *Handler) CreateReview(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	orderID, err := strconv.ParseInt(c.Param("orderId"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	var req ReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.CreateReview(c.Request.Context(), orderID, userID, req); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) GetReviewByOrder(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("orderId"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	review, err := h.service.GetReviewByOrder(c.Request.Context(), orderID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, review)
}

func (h *Handler) ListReviews(c *gin.Context) {
	page := intQuery(c, "page", 1)
	pageSize := intQuery(c, "pageSize", 20)
	reviews, total, err := h.service.ListReviews(c.Request.Context(), page, pageSize)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"list": reviews, "total": total})
}

func (h *Handler) GetTransferConfig(c *gin.Context) {
	cfg, err := h.service.GetTransferConfig(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, cfg)
}

func (h *Handler) TransferOrder(c *gin.Context) {
	var req TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	orderID, err := strconv.ParseInt(c.Query("orderId"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.TransferOrder(c.Request.Context(), orderID, req.TargetTeacherID); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) AdminListOrders(c *gin.Context) {
	query := OrderQuery{
		PageNum:  intQuery(c, "pageNum", 1),
		PageSize: intQuery(c, "pageSize", 10),
		OrderNo:  c.Query("orderNo"),
		Status:   c.Query("status"),
	}
	if raw := c.Query("userId"); raw != "" {
		val, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			response.Fail(c, err)
			return
		}
		query.UserID = &val
	}
	orders, total, err := h.service.AdminListOrders(c.Request.Context(), query)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"list": orders, "total": total})
}

func (h *Handler) AdminGetOrder(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	order, err := h.service.AdminGetOrder(c.Request.Context(), orderID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, order)
}

func (h *Handler) AdminUpdateStatus(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.AdminUpdateOrderStatus(c.Request.Context(), orderID, req.Status); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) AdminRefund(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.AdminRefundOrder(c.Request.Context(), orderID); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) AdminManualComplete(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.AdminManualComplete(c.Request.Context(), orderID); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) AdminUpdateRemark(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	var req UpdateRemarkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.AdminUpdateRemark(c.Request.Context(), orderID, req.Remark); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) AdminUpdateTeachers(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	var req UpdateTeachersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.AdminUpdateTeachers(c.Request.Context(), orderID, req.TeacherIDs); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) AdminManualCreate(c *gin.Context) {
	var req ManualOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	order, err := h.service.AdminManualCreateOrder(c.Request.Context(), req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, order)
}

func (h *Handler) AdminStats(c *gin.Context) {
	var start, end *string
	if s := c.Query("start"); s != "" {
		start = &s
	}
	if e := c.Query("end"); e != "" {
		end = &e
	}
	stats, err := h.service.AdminGetOrderStats(c.Request.Context(), start, end)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, stats)
}

func (h *Handler) AdminListReviews(c *gin.Context) {
	query := ReviewQuery{
		PageNum:  intQuery(c, "pageNum", 1),
		PageSize: intQuery(c, "pageSize", 10),
		Status:   c.Query("status"),
	}
	if raw := c.Query("orderId"); raw != "" {
		val, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			response.Fail(c, err)
			return
		}
		query.OrderID = &val
	}
	reviews, total, err := h.service.AdminListReviews(c.Request.Context(), query)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"list": reviews, "total": total})
}

func (h *Handler) AdminUpdateReviewStatus(c *gin.Context) {
	reviewID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.AdminUpdateReviewStatus(c.Request.Context(), reviewID, req.Status); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) AdminReplyReview(c *gin.Context) {
	reviewID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	var req ReplyReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.AdminReplyReview(c.Request.Context(), reviewID, req.Reply); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) AdminListFinalReview(c *gin.Context) {
	query := FinalReviewQuery{
		PageNum:  intQuery(c, "pageNum", 1),
		PageSize: intQuery(c, "pageSize", 10),
		Status:   c.Query("status"),
	}
	orders, total, err := h.service.AdminListFinalReview(c.Request.Context(), query)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"list": orders, "total": total})
}

func (h *Handler) AdminApproveFinalReview(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.AdminApproveFinalReview(c.Request.Context(), orderID); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) AdminRejectFinalReview(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.AdminRejectFinalReview(c.Request.Context(), orderID); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) CreateCashierOrder(c *gin.Context) {
	var req CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	co, err := h.service.CreateCashierOrder(c.Request.Context(), req.OrderID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, co)
}

func (h *Handler) GetCashierOrder(c *gin.Context) {
	token := c.Param("token")
	co, err := h.service.GetCashierOrder(c.Request.Context(), token)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, co)
}

func (h *Handler) CashierPay(c *gin.Context) {
	token := c.Param("token")
	var req CashierPayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	record, err := h.service.CashierPay(c.Request.Context(), token, req.Channel)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, record)
}

func (h *Handler) CreateWxPayOrder(c *gin.Context) {
	var req CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	resp, err := h.service.CreateWxPayOrder(c.Request.Context(), req.OrderID, req.Amount)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *Handler) WxPayNotify(c *gin.Context) {
	var req SyncPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.WxPayNotify(c.Request.Context(), req.OutTradeNo); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) QueryWxPay(c *gin.Context) {
	outTradeNo := c.Query("outTradeNo")
	record, err := h.service.QueryWxPay(c.Request.Context(), outTradeNo)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, record)
}

func (h *Handler) CreateAlipayOrder(c *gin.Context) {
	var req CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	resp, err := h.service.CreateAlipayOrder(c.Request.Context(), req.OrderID, req.Amount)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *Handler) AlipayNotify(c *gin.Context) {
	var req SyncPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.AlipayNotify(c.Request.Context(), req.OutTradeNo); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) QueryAlipay(c *gin.Context) {
	outTradeNo := c.Query("outTradeNo")
	record, err := h.service.QueryAlipay(c.Request.Context(), outTradeNo)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, record)
}

func (h *Handler) ManualSyncPayment(c *gin.Context) {
	var req SyncPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.ManualSyncPayment(c.Request.Context(), req.OutTradeNo, req.Channel); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) BatchSyncPayments(c *gin.Context) {
	var req BatchSyncPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.BatchSyncPayments(c.Request.Context(), req.IDs); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) SyncOverduePayments(c *gin.Context) {
	if err := h.service.SyncOverduePayments(c.Request.Context()); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}
