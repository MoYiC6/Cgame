package user

import (
	"strconv"

	"backend/internal/platform/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service  *Service
	authMiddleware gin.HandlerFunc
}

func NewHandler(service *Service, authMiddleware gin.HandlerFunc) *Handler {
	return &Handler{service: service, authMiddleware: authMiddleware}
}

func (h *Handler) RegisterRoutes(group *gin.RouterGroup) {
	user := group.Group("/user")
	if h.authMiddleware != nil {
		user.Use(h.authMiddleware)
	}
	{
		user.GET("/balance", h.GetBalance)
		user.GET("/balance/logs", h.GetBalanceLogs)
		user.GET("/level", h.GetLevel)
	}
}

func (h *Handler) GetBalance(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.GetString("userID"), 10, 64)
	user, err := h.service.GetUser(c.Request.Context(), userID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"balance": user.Balance})
}

func (h *Handler) GetBalanceLogs(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.GetString("userID"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	logs, total, err := h.service.GetBalanceLogs(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"list": logs, "total": total})
}

func (h *Handler) GetLevel(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.GetString("userID"), 10, 64)
	level, err := h.service.GetUserLevel(c.Request.Context(), userID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if level == nil {
		response.Success(c, gin.H{"level": nil})
		return
	}
	response.Success(c, level)
}
