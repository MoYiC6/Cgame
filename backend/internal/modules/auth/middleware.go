package auth

import (
	"errors"
	"net/http"
	"strings"

	apperrors "backend/internal/platform/errors"
	"backend/internal/platform/response"
	"backend/internal/platform/security"
	"github.com/gin-gonic/gin"
)

var (
	ErrTokenMissing = apperrors.New("AUTH_TOKEN_MISSING", "未登录", http.StatusUnauthorized, nil)
	ErrTokenInvalid = apperrors.New("AUTH_TOKEN_INVALID", "登录状态无效", http.StatusUnauthorized, nil)
	ErrTokenExpired = apperrors.New("AUTH_TOKEN_EXPIRED", "登录状态已过期", http.StatusUnauthorized, nil)
	ErrForbidden    = apperrors.New("AUTH_FORBIDDEN", "无权限", http.StatusForbidden, nil)
)

func AuthMiddleware(tokenManager security.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := extractBearerToken(c.GetHeader("Authorization"))
		if raw == "" {
			response.Fail(c, ErrTokenMissing)
			c.Abort()
			return
		}
		principal, _, err := tokenManager.VerifyAccessToken(c.Request.Context(), raw)
		if err != nil {
			response.Fail(c, mapTokenError(err))
			c.Abort()
			return
		}
		ctx := security.WithPrincipal(c.Request.Context(), principal)
		ctx = security.WithSessionID(ctx, principal.SessionID)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, ok := security.PrincipalFromContext(c.Request.Context())
		if !ok {
			response.Fail(c, ErrUnauthorized)
			c.Abort()
			return
		}
		if !security.HasPermission(principal, permission) {
			response.Fail(c, ErrForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}

func extractBearerToken(header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func mapTokenError(err error) *apperrors.AppError {
	if errors.Is(err, security.ErrTokenExpired) {
		return ErrTokenExpired
	}
	return ErrTokenInvalid
}
