package refund

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
	client := group.Group("/client/refunds")
	if h.authMiddleware != nil {
		client.Use(h.authMiddleware)
	}
	{
		client.POST("/apply", h.Apply)
		client.GET("/list", h.ListMine)
		client.GET("/:id", h.GetMine)
		client.POST("/:id/cancel", h.Cancel)
		client.GET("/can-apply/:orderId", h.CanApply)
		client.GET("/by-order/:orderId", h.GetByOrder)
	}

	admin := group.Group("/admin/refunds")
	if h.authMiddleware != nil {
		admin.Use(h.authMiddleware)
	}
	{
		admin.GET("", h.ListAdmin)
		admin.GET("/:id", h.GetAdmin)
		admin.PUT("/:id/approve", h.Approve)
		admin.PUT("/:id/reject", h.Reject)
		admin.PUT("/:id/process", h.Process)
		admin.GET("/stats", h.Stats)
	}
}

func (h *Handler) Apply(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	var req ApplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	id, err := h.service.Apply(c.Request.Context(), userID, req)
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

func (h *Handler) Cancel(c *gin.Context) {
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
	if err := h.service.Cancel(c.Request.Context(), userID, id); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) CanApply(c *gin.Context) {
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
	result, err := h.service.CanApply(c.Request.Context(), userID, orderID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) GetByOrder(c *gin.Context) {
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
	result, err := h.service.GetByOrder(c.Request.Context(), userID, orderID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) ListAdmin(c *gin.Context) {
	query := RefundQuery{
		PageNum:  intQuery(c, "pageNum", 1),
		PageSize: intQuery(c, "pageSize", 10),
		Status:   c.Query("status"),
		RefundNo: c.Query("refundNo"),
	}
	if raw := c.Query("orderId"); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			response.Fail(c, err)
			return
		}
		query.OrderID = &value
	}
	if raw := c.Query("userId"); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			response.Fail(c, err)
			return
		}
		query.UserID = &value
	}
	query.CreateTimeStart = stringPtr(c.Query("createTimeStart"))
	query.CreateTimeEnd = stringPtr(c.Query("createTimeEnd"))

	result, err := h.service.ListAdmin(c.Request.Context(), query)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) GetAdmin(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	result, err := h.service.GetAdmin(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) Approve(c *gin.Context) {
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
	var req struct {
		AdminRemark string `json:"adminRemark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.Approve(c.Request.Context(), userID, id, req.AdminRemark); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) Reject(c *gin.Context) {
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
	var req struct {
		AdminRemark string `json:"adminRemark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.Reject(c.Request.Context(), userID, id, req.AdminRemark); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) Process(c *gin.Context) {
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
	var req struct {
		AdminRemark string `json:"adminRemark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.Process(c.Request.Context(), userID, id, req.AdminRemark); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) Stats(c *gin.Context) {
	stats, err := h.service.Stats(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, stats)
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

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
