package bootstrap

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"backend/internal/platform/config"
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

func CORSMiddleware(cfg config.CORSConfig) gin.HandlerFunc {
	allowedMethods := strings.Join(cfg.AllowedMethods, ", ")
	allowedHeaders := strings.Join(cfg.AllowedHeaders, ", ")

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && isAllowedOrigin(origin, cfg.AllowedOrigins) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			if allowedMethods != "" {
				c.Header("Access-Control-Allow-Methods", allowedMethods)
			}
			if allowedHeaders != "" {
				c.Header("Access-Control-Allow-Headers", allowedHeaders)
			}
			if cfg.AllowCredentials {
				c.Header("Access-Control-Allow-Credentials", "true")
			}
			if cfg.MaxAgeSecs > 0 {
				c.Header("Access-Control-Max-Age", strconv.Itoa(cfg.MaxAgeSecs))
			}
		}
		if isCORSPreflightRequest(c.Request) {
			c.Status(http.StatusNoContent)
			c.Abort()
			return
		}
		c.Next()
	}
}

func SecurityHeadersMiddleware(cfg config.SecurityHeadersConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cfg.FrameOptions != "" {
			c.Header("X-Frame-Options", cfg.FrameOptions)
		}
		if cfg.ContentTypeOptions {
			c.Header("X-Content-Type-Options", "nosniff")
		}
		if cfg.ReferrerPolicy != "" {
			c.Header("Referrer-Policy", cfg.ReferrerPolicy)
		}
		c.Next()
	}
}

type rateLimiter struct {
	mu      sync.Mutex
	entries map[string]*rateLimitEntry
}

type rateLimitEntry struct {
	windowStarted time.Time
	count         int
}

func RateLimitMiddleware(cfg config.RateLimitConfig) gin.HandlerFunc {
	limiter := &rateLimiter{entries: make(map[string]*rateLimitEntry)}
	window := time.Duration(cfg.WindowSecs) * time.Second

	return func(c *gin.Context) {
		if cfg.Requests <= 0 || cfg.WindowSecs <= 0 {
			c.Next()
			return
		}

		if limiter.allow(clientAddress(c.Request), time.Now(), cfg.Requests, window) {
			c.Next()
			return
		}

		response.Fail(c, apperrors.New("RATE_LIMITED", "rate limit exceeded", http.StatusTooManyRequests, nil))
		c.Abort()
	}
}

func (l *rateLimiter) allow(key string, now time.Time, limit int, window time.Duration) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.evictExpired(now, window)

	entry, ok := l.entries[key]
	if !ok || now.Sub(entry.windowStarted) >= window {
		l.entries[key] = &rateLimitEntry{windowStarted: now, count: 1}
		return true
	}
	if entry.count >= limit {
		return false
	}
	entry.count++
	return true
}

func (l *rateLimiter) evictExpired(now time.Time, window time.Duration) {
	for key, entry := range l.entries {
		if now.Sub(entry.windowStarted) >= window {
			delete(l.entries, key)
		}
	}
}

func clientAddress(request *http.Request) string {
	if request == nil {
		return "unknown"
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(request.RemoteAddr))
	if err == nil && host != "" {
		return host
	}
	if strings.TrimSpace(request.RemoteAddr) != "" {
		return strings.TrimSpace(request.RemoteAddr)
	}
	return "unknown"
}

func isCORSPreflightRequest(request *http.Request) bool {
	if request == nil {
		return false
	}
	if request.Method != http.MethodOptions {
		return false
	}
	if strings.TrimSpace(request.Header.Get("Origin")) == "" {
		return false
	}
	return strings.TrimSpace(request.Header.Get("Access-Control-Request-Method")) != ""
}

func isAllowedOrigin(origin string, allowedOrigins []string) bool {
	if origin == "" {
		return false
	}
	for _, allowedOrigin := range allowedOrigins {
		if allowedOrigin == origin {
			return true
		}
	}
	return false
}

func RecoveryMiddleware(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				if log != nil {
					log.Error("panic recovered", "panic", recovered, "path", c.Request.URL.Path)
				}
				response.Fail(c, apperrors.NewAppError("INTERNAL_ERROR", "internal error", http.StatusInternalServerError).WithCause(fmt.Errorf("panic: %v", recovered)))
				c.Abort()
			}
		}()

		c.Next()
	}
}
