package withdrawal

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
	// Client teacher withdrawal routes
	client := group.Group("/client/teacher/withdrawal")
	if h.authMiddleware != nil {
		client.Use(h.authMiddleware)
	}
	{
		client.GET("/income-stats", h.GetIncomeStats)
		client.POST("/calculate", h.Calculate)
		client.POST("/apply", h.Apply)
		client.PUT("/:id/cancel", h.Cancel)
		client.GET("/records", h.ListMine)
		client.GET("/:id", h.GetMine)
	}

	// Admin withdrawal routes
	admin := group.Group("/admin/withdrawal")
	if h.authMiddleware != nil {
		admin.Use(h.authMiddleware)
	}
	{
		admin.GET("/list", h.ListAdmin)
		admin.GET("/:id", h.GetAdmin)
		admin.PUT("/:id/approve", h.Approve)
		admin.PUT("/:id/reject", h.Reject)
		admin.PUT("/:id/pay", h.Pay)
		admin.GET("/stats", h.Stats)
	}
}

func (h *Handler) GetIncomeStats(c *gin.Context) {
	teacherID, ok := currentTeacherID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	stats, err := h.service.GetIncomeStats(c.Request.Context(), teacherID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, stats)
}

func (h *Handler) Calculate(c *gin.Context) {
	var req CalculateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	result, err := h.service.CalculateWithdrawal(c.Request.Context(), req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) Apply(c *gin.Context) {
	teacherID, ok := currentTeacherID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	var req ApplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	withdrawal, err := h.service.Apply(c.Request.Context(), teacherID, req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, withdrawal)
}

func (h *Handler) Cancel(c *gin.Context) {
	teacherID, ok := currentTeacherID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.Cancel(c.Request.Context(), teacherID, id); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) ListMine(c *gin.Context) {
	teacherID, ok := currentTeacherID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	result, err := h.service.ListMine(c.Request.Context(), teacherID, page, pageSize)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) GetMine(c *gin.Context) {
	teacherID, ok := currentTeacherID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	result, err := h.service.GetMine(c.Request.Context(), teacherID, id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, result)
}

// Admin handlers

func (h *Handler) ListAdmin(c *gin.Context) {
	var query WithdrawalQuery
	query.PageNum, _ = strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	query.PageSize, _ = strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	query.Status = c.Query("status")
	if teacherIDStr := c.Query("teacherId"); teacherIDStr != "" {
		if teacherID, err := strconv.ParseInt(teacherIDStr, 10, 64); err == nil {
			query.TeacherID = &teacherID
		}
	}
	query.WithdrawalNo = c.Query("withdrawalNo")
	createTimeStart := c.Query("createTimeStart")
	if createTimeStart != "" {
		query.CreateTimeStart = &createTimeStart
	}
	createTimeEnd := c.Query("createTimeEnd")
	if createTimeEnd != "" {
		query.CreateTimeEnd = &createTimeEnd
	}
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
	if err := h.service.Approve(c.Request.Context(), adminUserID, id, req.AdminRemark); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) Reject(c *gin.Context) {
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
	if err := h.service.Reject(c.Request.Context(), adminUserID, id, req.AdminRemark); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) Pay(c *gin.Context) {
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
	if err := h.service.Pay(c.Request.Context(), adminUserID, id); err != nil {
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

func currentTeacherID(c *gin.Context) (int64, bool) {
	// Teacher ID is derived from userID for now; in production this may be
	// resolved via teacher profile lookup.
	return currentUserID(c)
}
