package order

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
	group.GET("/order/ping", h.Ping)
	client := group.Group("/client/order")
	{
		client.POST("/create", h.CreateOrder)
		client.GET("/:orderId", h.GetOrder)
		client.GET("/list", h.ListOrders)
		client.POST("/:orderId/pay", h.PayOrder)
		client.POST("/:orderId/complete", h.CompleteOrder)
		client.POST("/:orderId/cancel", h.CancelOrder)
	}
}

func (h *Handler) Ping(c *gin.Context) {
	response.Success(c, gin.H{"module": "order"})
}

func (h *Handler) CreateOrder(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.GetString("userID"), 10, 64)
	var req struct {
		SKUName string  `json:"skuName"`
		Quantity int    `json:"quantity"`
		Remark  *string `json:"remark"`
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
	orderID, _ := strconv.ParseInt(c.Param("orderId"), 10, 64)
	order, err := h.service.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, order)
}

func (h *Handler) ListOrders(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.GetString("userID"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	orders, total, err := h.service.ListOrders(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"list": orders, "total": total})
}

func (h *Handler) PayOrder(c *gin.Context) {
	orderID, _ := strconv.ParseInt(c.Param("orderId"), 10, 64)
	if err := h.service.PayOrder(c.Request.Context(), orderID); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) CompleteOrder(c *gin.Context) {
	orderID, _ := strconv.ParseInt(c.Param("orderId"), 10, 64)
	if err := h.service.CompleteOrder(c.Request.Context(), orderID); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) CancelOrder(c *gin.Context) {
	orderID, _ := strconv.ParseInt(c.Param("orderId"), 10, 64)
	if err := h.service.CancelOrder(c.Request.Context(), orderID); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}
