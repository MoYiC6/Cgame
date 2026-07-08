package recharge

import (
	"net/http"
	"strconv"

	apperrors "backend/internal/platform/errors"
	"backend/internal/platform/response"
	"backend/internal/platform/security"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service        *Service
	authMiddleware gin.HandlerFunc
}

func NewHandler(service *Service, authMiddleware gin.HandlerFunc) *Handler {
	return &Handler{service: service, authMiddleware: authMiddleware}
}

func (h *Handler) RegisterRoutes(group *gin.RouterGroup) {
	// Client recharge routes
	client := group.Group("/recharge")
	if h.authMiddleware != nil {
		client.Use(h.authMiddleware)
	}
	{
		client.POST("/create", h.Create)
		client.GET("/my-records", h.ListMine)
		client.GET("/detail/:id", h.GetMine)
		client.GET("/statistics", h.Stats)
		client.GET("/recent/:userId", h.GetRecent)
		client.POST("/cancel/:rechargeNo", h.Cancel)
		client.POST("/continue-pay/:rechargeNo", h.ContinuePay)
		client.POST("/verify-payment/:rechargeNo", h.VerifyPayment)
		client.POST("/callback", h.Callback)
	}

	// Manual recharge (admin)
	group.POST("/recharge/manual", h.authMiddleware, h.ManualRecharge)

	// Rebate rules (client)
	rebateClient := group.Group("/client/recharge-rebate")
	if h.authMiddleware != nil {
		rebateClient.Use(h.authMiddleware)
	}
	{
		rebateClient.GET("/available-rules", h.ListAvailableRules)
		rebateClient.GET("/preview", h.PreviewRebate)
	}

	// Rebate rules (admin)
	rebateAdmin := group.Group("/admin/recharge-rebate")
	if h.authMiddleware != nil {
		rebateAdmin.Use(h.authMiddleware)
	}
	{
		rebateAdmin.GET("", h.ListRebateRules)
		rebateAdmin.POST("", h.CreateRebateRule)
		rebateAdmin.PUT("/:id", h.UpdateRebateRule)
		rebateAdmin.DELETE("/:id", h.DeleteRebateRule)
	}
}

func (h *Handler) Create(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	var req CreateRechargeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	record, err := h.service.CreateRecharge(c.Request.Context(), userID, req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, record)
}

func (h *Handler) ManualRecharge(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	var req ManualRechargeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	id, err := h.service.ManualRecharge(c.Request.Context(), userID, req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, id)
}

func (h *Handler) ListMine(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	result, err := h.service.ListMine(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) GetMine(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	result, err := h.service.GetMine(c.Request.Context(), userID, id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) Stats(c *gin.Context) {
	stats, err := h.service.Stats(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, stats)
}

func (h *Handler) GetRecent(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))
	result, err := h.service.GetRecent(c.Request.Context(), userID, limit)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) Cancel(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	rechargeNo := c.Param("rechargeNo")
	if err := h.service.Cancel(c.Request.Context(), userID, rechargeNo); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) ContinuePay(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	rechargeNo := c.Param("rechargeNo")
	result, err := h.service.ContinuePay(c.Request.Context(), userID, rechargeNo)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) VerifyPayment(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	rechargeNo := c.Param("rechargeNo")
	result, err := h.service.VerifyPayment(c.Request.Context(), userID, rechargeNo)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) Callback(c *gin.Context) {
	var req struct {
		RechargeNo string  `json:"rechargeNo"`
		PayChannel string  `json:"payChannel"`
		PayAmount  float64 `json:"payAmount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.Callback(c.Request.Context(), req.RechargeNo, req.PayChannel, req.PayAmount); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

// Rebate rule handlers

func (h *Handler) ListAvailableRules(c *gin.Context) {
	result, err := h.service.ListAvailableRebateRules(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) PreviewRebate(c *gin.Context) {
	amount, err := strconv.ParseFloat(c.Query("amount"), 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	result, err := h.service.PreviewRebate(c.Request.Context(), amount)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) ListRebateRules(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	result, err := h.service.ListRebateRules(c.Request.Context(), page, pageSize)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) CreateRebateRule(c *gin.Context) {
	var req RebateRuleCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	id, err := h.service.CreateRebateRule(c.Request.Context(), req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, id)
}

func (h *Handler) UpdateRebateRule(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	var req RebateRuleUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.UpdateRebateRule(c.Request.Context(), id, req); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) DeleteRebateRule(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.DeleteRebateRule(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
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
