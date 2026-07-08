package feedback

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
	client := group.Group("/client/feedback")
	if h.authMiddleware != nil {
		client.Use(h.authMiddleware)
	}
	{
		client.POST("/submit", h.Submit)
		client.GET("/list", h.ListMine)
		client.GET("/:id", h.GetMine)
	}

	admin := group.Group("/admin/feedback")
	if h.authMiddleware != nil {
		admin.Use(h.authMiddleware)
	}
	{
		admin.GET("", h.ListAdmin)
		admin.GET("/list", h.ListAdmin)
		admin.GET("/:id", h.GetAdmin)
		admin.POST("/reply", h.Reply)
		admin.POST("/:id/reply", h.ReplyByPath)
		admin.PUT("/status", h.UpdateStatus)
		admin.PUT("/:id/status", h.UpdateStatusByPath)
		admin.DELETE("/:id", h.Delete)
	}
}

func (h *Handler) Submit(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	var req SubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	id, err := h.service.Submit(c.Request.Context(), userID, req)
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
	page, _ := strconv.Atoi(c.DefaultQuery("pageNum", c.DefaultQuery("page", "1")))
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

func (h *Handler) ListAdmin(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("pageNum", c.DefaultQuery("page", "1")))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	var status *int
	if raw := c.Query("status"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil {
			response.Fail(c, err)
			return
		}
		status = &value
	}
	result, err := h.service.ListAdmin(c.Request.Context(), page, pageSize, status, c.Query("keyword"))
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

func (h *Handler) Reply(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	var req struct {
		FeedbackID int64  `json:"feedbackId"`
		Content    string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	replyID, err := h.service.Reply(c.Request.Context(), userID, req.FeedbackID, req.Content)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, replyID)
}

func (h *Handler) ReplyByPath(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	feedbackID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	var req struct {
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	replyID, err := h.service.Reply(c.Request.Context(), userID, feedbackID, req.Content)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, replyID)
}

func (h *Handler) UpdateStatus(c *gin.Context) {
	var req struct {
		FeedbackID int64 `json:"feedbackId"`
		Status     int   `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.UpdateStatus(c.Request.Context(), req.FeedbackID, req.Status); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) UpdateStatusByPath(c *gin.Context) {
	feedbackID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	var req struct {
		Status int `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.UpdateStatus(c.Request.Context(), feedbackID, req.Status); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
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
