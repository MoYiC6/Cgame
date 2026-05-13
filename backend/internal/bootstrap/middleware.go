package bootstrap

import (
	"fmt"
	"net/http"
	"time"

	apperrors "backend/internal/platform/errors"
	"backend/internal/platform/logger"
	"backend/internal/platform/observability"
	"backend/internal/platform/response"
	"github.com/gin-gonic/gin"
)

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = fmt.Sprintf("req-%d", time.Now().UnixNano())
		}

		ctx := observability.WithRequestID(c.Request.Context(), requestID)
		c.Request = c.Request.WithContext(ctx)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

func TraceContextMiddleware(propagator observability.Propagator) gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = fmt.Sprintf("trace-%d", time.Now().UnixNano())
		}

		carrier := observability.MapCarrier{"X-Trace-ID": traceID}
		ctx := propagator.Extract(c.Request.Context(), carrier)
		ctx = observability.WithTraceID(ctx, traceID)
		c.Request = c.Request.WithContext(ctx)
		c.Header("X-Trace-ID", traceID)
		c.Next()
	}
}

func RecoveryMiddleware(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				if log != nil {
					log.Error("panic recovered", "panic", recovered, "path", c.Request.URL.Path)
				}
				response.Fail(c, apperrors.Wrap(
					apperrors.NewAppError("INTERNAL_ERROR", "internal error", http.StatusInternalServerError),
					fmt.Errorf("panic: %v", recovered),
				))
				c.Abort()
			}
		}()

		c.Next()
	}
}
