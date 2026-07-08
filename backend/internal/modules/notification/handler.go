package notification

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
	group.GET("/notification/ping", h.Ping)

	client := group.Group("/client/notifications")
	if h.authMiddleware != nil {
		client.Use(h.authMiddleware)
	}
	{
		client.GET("", h.List)
		client.PUT("/:id/read", h.MarkRead)
		client.PUT("/read-all", h.MarkAllRead)
		client.GET("/system", h.GetSystemNotifications)
		client.GET("/system/unread-count", h.GetUnreadCount)
	}

	admin := group.Group("/admin/notifications")
	if h.authMiddleware != nil {
		admin.Use(h.authMiddleware)
	}
	{
		admin.POST("", h.AdminCreate)
		admin.GET("", h.AdminList)
		admin.GET("/stats", h.AdminStats)
		admin.DELETE("/:id", h.AdminDelete)
	}

	subscribe := group.Group("/client/subscribe-message")
	if h.authMiddleware != nil {
		subscribe.Use(h.authMiddleware)
	}
	{
		subscribe.GET("/templates", h.SubscribeTemplates)
		subscribe.POST("/record", h.RecordSubscribe)
		subscribe.GET("/status", h.SubscribeStatus)
	}

	adminInbox := group.Group("/admin/notification-inbox")
	if h.authMiddleware != nil {
		adminInbox.Use(h.authMiddleware)
	}
	{
		adminInbox.GET("", h.AdminInboxList)
		adminInbox.PUT("/:id/read", h.AdminInboxMarkRead)
		adminInbox.PUT("/read-all", h.AdminInboxMarkAllRead)
	}

	adminTodo := group.Group("/admin/system-todos")
	if h.authMiddleware != nil {
		adminTodo.Use(h.authMiddleware)
	}
	{
		adminTodo.POST("", h.CreateTodo)
		adminTodo.GET("", h.ListTodos)
		adminTodo.PUT("/:id/toggle", h.ToggleTodo)
		adminTodo.DELETE("", h.DeleteTodos)
	}
}

func (h *Handler) Ping(c *gin.Context) {
	response.Success(c, gin.H{"module": "notification", "status": "ok"})
}

// Client
func (h *Handler) List(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.GetString("userID"), 10, 64)
	if userID == 0 {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	notifications, total, err := h.service.GetUserNotifications(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"list": notifications, "total": total})
}

func (h *Handler) MarkRead(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.GetString("userID"), 10, 64)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.service.MarkAsRead(c.Request.Context(), userID, id); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) MarkAllRead(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.GetString("userID"), 10, 64)
	if err := h.service.MarkAllAsRead(c.Request.Context(), userID); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) GetSystemNotifications(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.GetString("userID"), 10, 64)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id > 0 {
		notification, err := h.service.GetSystemNotification(c.Request.Context(), id)
		if err != nil {
			response.Fail(c, err)
			return
		}
		response.Success(c, notification)
		return
	}
	notifications, total, err := h.service.GetUserNotifications(c.Request.Context(), userID, 1, 20)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"list": notifications, "total": total})
}

func (h *Handler) GetUnreadCount(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.GetString("userID"), 10, 64)
	count, err := h.service.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"count": count})
}

// Admin
func (h *Handler) AdminCreate(c *gin.Context) {
	var n Notification
	if err := c.ShouldBindJSON(&n); err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.CreateNotification(c.Request.Context(), &n); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) AdminList(c *gin.Context) {
	query := AdminNotificationQuery{
		PageNum:  intQuery(c, "pageNum", 1),
		PageSize: intQuery(c, "pageSize", 10),
		Type:     c.Query("type"),
	}
	notifications, total, err := h.service.ListAdminNotifications(c.Request.Context(), query)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"list": notifications, "total": total})
}

func (h *Handler) AdminDelete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id == 0 {
		var ids []int64
		if err := c.ShouldBindJSON(&ids); err != nil {
			response.Fail(c, err)
			return
		}
		if err := h.service.DeleteTodos(c.Request.Context(), ids); err != nil {
			response.Fail(c, err)
			return
		}
		response.Success(c, nil)
		return
	}
	if err := h.service.DeleteNotification(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) AdminStats(c *gin.Context) {
	stats, err := h.service.GetNotificationStats(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, stats)
}

// Subscribe message
func (h *Handler) SubscribeTemplates(c *gin.Context) {
	templates, err := h.service.GetSubscribeTemplates(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, templates)
}

func (h *Handler) RecordSubscribe(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.GetString("userID"), 10, 64)
	var req SubscribeRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.RecordSubscribe(c.Request.Context(), userID, req); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) SubscribeStatus(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.GetString("userID"), 10, 64)
	templateID := c.Query("templateId")
	if templateID == "" {
		response.Fail(c, fmt.Errorf("templateId is required"))
		return
	}
	status, err := h.service.GetSubscribeStatus(c.Request.Context(), userID, templateID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, status)
}

func intQuery(c *gin.Context, key string, fallback int) int {
	value, err := strconv.Atoi(c.DefaultQuery(key, strconv.Itoa(fallback)))
	if err != nil {
		return fallback
	}
	return value
}

func (h *Handler) AdminInboxList(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("pageIndex", c.DefaultQuery("page", "1")))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	var unreadOnly *bool
	if raw := c.Query("unreadOnly"); raw != "" {
		value, err := strconv.ParseBool(raw)
		if err != nil {
			response.Fail(c, err)
			return
		}
		unreadOnly = &value
	}

	result, err := h.service.ListInboxNotifications(c.Request.Context(), userID, page, pageSize, c.Query("type"), unreadOnly)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) AdminInboxMarkRead(c *gin.Context) {
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
	if err := h.service.MarkInboxAsRead(c.Request.Context(), userID, id); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) AdminInboxMarkAllRead(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	if err := h.service.MarkAllInboxAsRead(c.Request.Context(), userID, c.Query("type")); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

// Todos
func (h *Handler) CreateTodo(c *gin.Context) {
	response.Success(c, gin.H{"message": "create todo endpoint"})
}

func (h *Handler) ListTodos(c *gin.Context) {
	completedStr := c.Query("completed")
	var completed *bool
	if completedStr != "" {
		val := completedStr == "true"
		completed = &val
	}
	todos, err := h.service.GetTodos(c.Request.Context(), completed)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"list": todos})
}

func (h *Handler) ToggleTodo(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	completed := c.Query("completed") == "true"
	if err := h.service.ToggleTodo(c.Request.Context(), id, completed, "admin"); err != nil {
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

func (h *Handler) DeleteTodos(c *gin.Context) {
	var ids []int64
	if err := c.ShouldBindJSON(&ids); err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.DeleteTodos(c.Request.Context(), ids); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}
