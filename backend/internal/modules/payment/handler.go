package payment

import (
	"strconv"

	"backend/internal/platform/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(group *gin.RouterGroup) {
	group.GET("/payment/ping", h.Ping)

	client := group.Group("/client/payment")
	{
		client.POST("/create", h.CreatePayment)
		client.POST("/confirm", h.ConfirmPayment)
		client.GET("/:paymentNo", h.GetPayment)
		client.GET("/list", h.ListPayments)
	}

	compatibleClient := group.Group("/client/payments")
	{
		compatibleClient.POST("", h.CreatePayment)
		compatibleClient.POST("/confirm", h.ConfirmPayment)
		compatibleClient.GET("/status", h.GetPaymentStatus)
	}

	admin := group.Group("/admin/payments")
	{
		admin.GET("", h.ListAdminPayments)
		admin.GET("/stats", h.GetPaymentStats)
	}
}

func (h *Handler) Ping(c *gin.Context) {
	response.Success(c, gin.H{"module": "payment"})
}

func (h *Handler) CreatePayment(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.GetString("userID"), 10, 64)
	var req struct {
		OrderNo   string  `json:"orderNo"`
		Amount    float64 `json:"amount"`
		PayMethod string  `json:"payMethod"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	payment, err := h.service.CreatePayment(c.Request.Context(), userID, req.OrderNo, req.Amount, req.PayMethod)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, payment)
}

func (h *Handler) ConfirmPayment(c *gin.Context) {
	var req struct {
		PaymentNo string `json:"paymentNo"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.ConfirmPayment(c.Request.Context(), req.PaymentNo); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) GetPayment(c *gin.Context) {
	paymentNo := c.Param("paymentNo")
	payment, err := h.service.GetPayment(c.Request.Context(), paymentNo)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, payment)
}

func (h *Handler) GetPaymentStatus(c *gin.Context) {
	paymentNo := c.Query("paymentNo")
	if paymentNo == "" {
		paymentNo = c.Query("outTradeNo")
	}
	if paymentNo == "" {
		response.Success(c, gin.H{"status": "unknown"})
		return
	}

	payment, err := h.service.GetPayment(c.Request.Context(), paymentNo)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"paymentNo": payment.PaymentNo, "status": payment.Status})
}

func (h *Handler) ListPayments(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.GetString("userID"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	payments, total, err := h.service.ListPayments(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"list": payments, "total": total})
}

func (h *Handler) ListAdminPayments(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("pageIndex", c.DefaultQuery("page", "1")))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	payments, total, err := h.service.ListAdminPayments(c.Request.Context(), page, pageSize)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"list": payments, "total": total})
}

func (h *Handler) GetPaymentStats(c *gin.Context) {
	stats, err := h.service.GetPaymentStats(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, stats)
}
