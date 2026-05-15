package order

import (
	"net/http"

	apperrors "backend/internal/platform/errors"
	"backend/internal/platform/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(group *gin.RouterGroup) {
	group.GET("/order/ping", h.Ping)
}

func (h *Handler) Ping(c *gin.Context) {
	payload, err := h.service.Ping(c.Request.Context())
	if err != nil {
		response.Fail(c, apperrors.New("ORDER_PING_FAILED", "order ping failed", http.StatusInternalServerError, err))
		return
	}
	response.Success(c, payload)
}
