package invite

import (
	"fmt"
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
	// Client routes
	client := group.Group("/client/invite")
	if h.authMiddleware != nil {
		client.Use(h.authMiddleware)
	}
	{
		client.GET("/info", h.GetInviteInfo)
		client.GET("/records", h.ListInviteRecords)
		client.POST("/bindInviter", h.BindInviter)
		client.GET("/validate", h.ValidateInviteCode)
	}

	// Teacher invite code routes
	teacher := group.Group("/client/teacher")
	if h.authMiddleware != nil {
		teacher.Use(h.authMiddleware)
	}
	{
		teacher.GET("/invite-code", h.GetMyTeacherInviteCode)
		teacher.POST("/invite-code", h.GenerateTeacherInviteCode)
	}

	// Admin routes
	admin := group.Group("/admin/teacher/invite-code")
	if h.authMiddleware != nil {
		admin.Use(h.authMiddleware)
	}
	{
		admin.GET("", h.ListAdminInviteCodes)
		admin.POST("", h.CreateAdminInviteCodes)
		admin.PUT("/:id", h.UpdateAdminInviteCode)
		admin.DELETE("/:id", h.DeleteAdminInviteCode)
		admin.POST("/:id/revoke", h.RevokeAdminInviteCode)
	}
}

// Client handlers

func (h *Handler) GetInviteInfo(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	info, err := h.service.GetInviteInfo(c.Request.Context(), userID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, info)
}

func (h *Handler) ListInviteRecords(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	result, err := h.service.ListInviteRecords(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) BindInviter(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	var req BindRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.BindInviter(c.Request.Context(), userID, req); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) ValidateInviteCode(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		response.Fail(c, fmt.Errorf("invite code is required"))
		return
	}
	result, err := h.service.ValidateInviteCode(c.Request.Context(), code)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, result)
}

// Teacher invite code handlers

func (h *Handler) GetMyTeacherInviteCode(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	code, err := h.service.GetMyTeacherInviteCode(c.Request.Context(), userID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, code)
}

func (h *Handler) GenerateTeacherInviteCode(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	code, err := h.service.GenerateTeacherInviteCode(c.Request.Context(), userID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, code)
}

// Admin invite code handlers

func (h *Handler) ListAdminInviteCodes(c *gin.Context) {
	query := InviteCodeQuery{
		PageNum:  intQuery(c, "pageNum", 1),
		PageSize: intQuery(c, "pageSize", 10),
		Code:     c.Query("code"),
		Status:   c.Query("status"),
	}
	if raw := c.Query("usedBy"); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			response.Fail(c, err)
			return
		}
		query.UsedBy = &value
	}
	if raw := c.Query("createdBy"); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			response.Fail(c, err)
			return
		}
		query.CreatedBy = &value
	}
	query.CreateTimeStart = stringPtr(c.Query("createTimeStart"))
	query.CreateTimeEnd = stringPtr(c.Query("createTimeEnd"))

	result, err := h.service.ListTeacherInviteCodes(c.Request.Context(), query)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) CreateAdminInviteCodes(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.GenerateAdminInviteCodes(c.Request.Context(), userID, req); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) UpdateAdminInviteCode(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	var req struct {
		Remark string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.UpdateTeacherInviteCode(c.Request.Context(), id, req.Remark); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) DeleteAdminInviteCode(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.DeleteTeacherInviteCode(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) RevokeAdminInviteCode(c *gin.Context) {
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
	if err := h.service.RevokeTeacherInviteCode(c.Request.Context(), id, userID); err != nil {
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
