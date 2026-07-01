package finance

import (
	"backend/internal/platform/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(group *gin.RouterGroup) {
	admin := group.Group("/admin/finance")
	{
		admin.GET("/stats", h.GetFinanceStats)
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
