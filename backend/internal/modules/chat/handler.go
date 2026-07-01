package chat

import (
	"strconv"

	"backend/internal/platform/response"

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
	client := group.Group("/client/chat")
	if h.authMiddleware != nil {
		client.Use(h.authMiddleware)
	}
	{
		client.GET("/sessions", h.GetUserSessions)
		client.POST("/sessions", h.CreateSession)
		client.GET("/sessions/:sessionId/messages", h.GetMessages)
		client.POST("/sessions/:sessionId/messages", h.SendMessage)
		client.PUT("/sessions/:sessionId/read", h.MarkRead)
		client.GET("/unread-count", h.GetUnreadCount)
	}
}

func (h *Handler) GetUserSessions(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.GetString("userID"), 10, 64)
	sessions, err := h.service.GetUserSessions(c.Request.Context(), userID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, sessions)
}

func (h *Handler) CreateSession(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.GetString("userID"), 10, 64)
	var req struct {
		TeacherID int64 `json:"teacherId"`
		OrderID   int64 `json:"orderId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	session, err := h.service.GetOrCreateSession(c.Request.Context(), userID, req.TeacherID, req.OrderID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, session)
}

func (h *Handler) GetMessages(c *gin.Context) {
	sessionID, _ := strconv.ParseInt(c.Param("sessionId"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	messages, total, err := h.service.GetSessionMessages(c.Request.Context(), sessionID, page, pageSize)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"list": messages, "total": total})
}

func (h *Handler) SendMessage(c *gin.Context) {
	sessionID, _ := strconv.ParseInt(c.Param("sessionId"), 10, 64)
	userID, _ := strconv.ParseInt(c.GetString("userID"), 10, 64)
	var req struct {
		Content     string  `json:"content"`
		MessageType *string `json:"messageType"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	msg, err := h.service.SendMessage(c.Request.Context(), sessionID, userID, "user", req.Content, req.MessageType, nil)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, msg)
}

func (h *Handler) MarkRead(c *gin.Context) {
	sessionID, _ := strconv.ParseInt(c.Param("sessionId"), 10, 64)
	userID, _ := strconv.ParseInt(c.GetString("userID"), 10, 64)
	if err := h.service.MarkSessionAsRead(c.Request.Context(), sessionID, userID, false); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) GetUnreadCount(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.GetString("userID"), 10, 64)
	count, err := h.service.GetUnreadCount(c.Request.Context(), userID, false)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"count": count})
}
