package teacher

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
	client := group.Group("/client/teacher")
	if h.authMiddleware != nil {
		client.Use(h.authMiddleware)
	}
	{
		client.GET("/my-status", h.GetMyStatus)
		client.GET("/levels", h.GetLevels)
	}

	admin := group.Group("/admin/teachers")
	if h.authMiddleware != nil {
		admin.Use(h.authMiddleware)
	}
	{
		admin.GET("", h.ListTeachers)
	}
}

func (h *Handler) GetMyStatus(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.GetString("userID"), 10, 64)
	teacher, err := h.service.GetTeacher(c.Request.Context(), userID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if teacher == nil {
		response.Success(c, gin.H{"isTeacher": false})
		return
	}
	response.Success(c, teacher)
}

func (h *Handler) GetLevels(c *gin.Context) {
	levels, err := h.service.GetTeacherLevels(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, levels)
}

func (h *Handler) ListTeachers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	teachers, total, err := h.service.ListTeachers(c.Request.Context(), page, pageSize)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"list": teachers, "total": total})
}
