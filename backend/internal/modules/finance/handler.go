package finance

import (
	"net/http"
	"strconv"
	"time"

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
	admin := group.Group("/admin/finance")
	if h.authMiddleware != nil {
		admin.Use(h.authMiddleware)
	}
	{
		admin.GET("/stats", h.GetFinanceStats)
		admin.GET("/operator-commissions/me", h.GetMyCommissions)
		admin.GET("/operator-commissions/me/balance", h.GetMyCommissionBalance)
		admin.GET("/operator-commissions/withdrawals", h.ListOperatorWithdrawals)
		admin.GET("/operator-commissions/withdrawals/me", h.ListMyWithdrawals)
		admin.POST("/operator-commissions/withdrawals", h.ApplyWithdrawal)
		admin.PUT("/operator-commissions/withdrawals/:id/approve", h.ApproveWithdrawal)
		admin.PUT("/operator-commissions/withdrawals/:id/reject", h.RejectWithdrawal)
		admin.PUT("/operator-commissions/withdrawals/:id/pay", h.PayWithdrawal)
		admin.PUT("/operator-commissions/withdrawals/:id/cancel", h.CancelWithdrawal)
		admin.GET("/balance/details", h.ListBalanceDetails)
		admin.GET("/user-monthly-report", h.GetMonthlyReport)
	}
}

func (h *Handler) GetFinanceStats(c *gin.Context) {
	stats, err := h.service.GetFinanceStats(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, stats)
}

func (h *Handler) GetMyCommissions(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	commissions, total, err := h.service.GetMyCommissions(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"list": commissions, "total": total})
}

func (h *Handler) GetMyCommissionBalance(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	balance, err := h.service.GetMyCommissionBalance(c.Request.Context(), userID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"balance": balance})
}

func (h *Handler) ListOperatorWithdrawals(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	operatorID, _ := strconv.ParseInt(c.Query("operatorId"), 10, 64)
	withdrawals, total, err := h.service.ListMyWithdrawals(c.Request.Context(), operatorID, page, pageSize)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"list": withdrawals, "total": total})
}

func (h *Handler) ListMyWithdrawals(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	withdrawals, total, err := h.service.ListMyWithdrawals(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"list": withdrawals, "total": total})
}

func (h *Handler) ApplyWithdrawal(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	var req struct {
		Amount float64 `json:"amount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	id, err := h.service.ApplyWithdrawal(c.Request.Context(), userID, req.Amount)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"id": id})
}

func (h *Handler) ApproveWithdrawal(c *gin.Context) {
	adminUserID, ok := currentUserID(c)
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
	if err := h.service.ApproveWithdrawal(c.Request.Context(), adminUserID, id, req.AdminRemark); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) RejectWithdrawal(c *gin.Context) {
	adminUserID, ok := currentUserID(c)
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
	if err := h.service.RejectWithdrawal(c.Request.Context(), adminUserID, id, req.AdminRemark); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) PayWithdrawal(c *gin.Context) {
	adminUserID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.PayWithdrawal(c.Request.Context(), adminUserID, id); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) CancelWithdrawal(c *gin.Context) {
	adminUserID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.CancelWithdrawal(c.Request.Context(), adminUserID, id); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) ListBalanceDetails(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.Query("userId"), 10, 64)
	if userID == 0 {
		userID, _ = strconv.ParseInt(c.GetString("userID"), 10, 64)
	}
	page, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	details, total, err := h.service.ListBalanceDetails(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"list": details, "total": total})
}

func (h *Handler) GetMonthlyReport(c *gin.Context) {
	month := c.Query("month")
	if month == "" {
		month = time.Now().Format("2006-01")
	}
	report, err := h.service.GetMonthlyReport(c.Request.Context(), month)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, report)
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
