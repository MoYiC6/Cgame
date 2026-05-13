package response

import (
	"net/http"

	apperrors "backend/internal/platform/errors"
	"backend/internal/platform/observability"
	"github.com/gin-gonic/gin"
)

type APIResponse[T any] struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Data      T      `json:"data,omitempty"`
	RequestID string `json:"request_id"`
}

func Success(c *gin.Context, data any) {
	requestID, _ := observability.RequestIDFromContext(c.Request.Context())
	c.JSON(http.StatusOK, APIResponse[any]{Code: "OK", Message: "success", Data: data, RequestID: requestID})
}

func Fail(c *gin.Context, err error) {
	requestID, _ := observability.RequestIDFromContext(c.Request.Context())
	c.JSON(apperrors.Status(err), APIResponse[any]{Code: apperrors.Code(err), Message: apperrors.SafeMessage(err), RequestID: requestID})
}
