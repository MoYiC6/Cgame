package payment

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
	group.GET("/payment/ping", h.Ping)
}

func (h *Handler) Ping(c *gin.Context) {
	payload, err := h.service.Ping(c.Request.Context())
	if err != nil {
		response.Fail(c, apperrors.New("PAYMENT_PING_FAILED", "payment ping failed", http.StatusInternalServerError, err))
		return
	}
	response.Success(c, payload)
}
