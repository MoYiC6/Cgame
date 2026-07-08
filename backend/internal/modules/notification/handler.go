package notification

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
		admin.DELETE("", h.AdminDelete)
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
	response.Success(c, gin.H{"message": "system notifications endpoint"})
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
	response.Success(c, gin.H{"message": "admin create notification endpoint"})
}

func (h *Handler) AdminList(c *gin.Context) {
	response.Success(c, gin.H{"message": "admin list notifications endpoint"})
}

func (h *Handler) AdminDelete(c *gin.Context) {
	response.Success(c, gin.H{"message": "admin delete notifications endpoint"})
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
