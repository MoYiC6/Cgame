package response

import (
	stderrors "errors"
	"net/http"

	apperrors "backend/internal/platform/errors"
	"backend/internal/platform/observability"
	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Data      any    `json:"data,omitempty"`
	RequestID string `json:"request_id,omitempty"`
	TraceID   string `json:"trace_id,omitempty"`
}

func Success(c *gin.Context, data any) {
	requestID, _ := observability.RequestIDFromContext(c.Request.Context())
	traceID, _ := observability.TraceIDFromContext(c.Request.Context())
	c.JSON(http.StatusOK, APIResponse{Code: apperrors.CodeOK, Message: "success", Data: data, RequestID: requestID, TraceID: traceID})
}

func Fail(c *gin.Context, err error) {
	if err == nil {
		return
	}
	requestID, _ := observability.RequestIDFromContext(c.Request.Context())
	traceID, _ := observability.TraceIDFromContext(c.Request.Context())

	var appErr *apperrors.AppError
	if stderrors.As(err, &appErr) {
		c.JSON(apperrors.Status(err), APIResponse{Code: apperrors.Code(err), Message: apperrors.SafeMessage(err), RequestID: requestID, TraceID: traceID})
		return
	}

	c.JSON(http.StatusInternalServerError, APIResponse{Code: apperrors.CodeInternal, Message: "internal error", RequestID: requestID, TraceID: traceID})
}
