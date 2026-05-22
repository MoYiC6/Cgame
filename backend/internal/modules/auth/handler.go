package auth

import (
	"net/http"
	"time"

	"backend/internal/platform/config"
	apperrors "backend/internal/platform/errors"
	"backend/internal/platform/response"
	"backend/internal/platform/security"
	"github.com/gin-gonic/gin"
)

type CookieConfig struct {
	Name     string
	Domain   string
	Path     string
	Secure   bool
	HTTPOnly bool
	SameSite string
}

type HandlerConfig struct {
	Cookie CookieConfig
}

type Handler struct {
	service Service
	config  HandlerConfig
}

func NewHandler(service Service, cfg HandlerConfig) *Handler {
	return &Handler{service: service, config: cfg}
}

func NewHandlerConfigFromAuth(cfg config.AuthConfig) HandlerConfig {
	return HandlerConfig{Cookie: CookieConfig{Name: cfg.Cookie.Name, Domain: cfg.Cookie.Domain, Path: cfg.Cookie.Path, Secure: cfg.Cookie.Secure, HTTPOnly: cfg.Cookie.HTTPOnly, SameSite: cfg.Cookie.SameSite}}
}

func (h *Handler) RegisterRoutes(group *gin.RouterGroup) {
	authGroup := group.Group("/auth")
	authGroup.POST("/login", h.Login)
	authGroup.POST("/refresh", h.Refresh)
	authGroup.POST("/logout", h.Logout)
	authGroup.GET("/me", h.Me)
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Identifier == "" || req.Password == "" {
		response.Fail(c, apperrors.New("INVALID_ARGUMENT", "invalid input", http.StatusBadRequest, err))
		return
	}
	resp, cookie, err := h.service.Login(c.Request.Context(), &req)
	if err != nil {
		response.Fail(c, toAppError(err))
		return
	}
	h.writeRefreshCookie(c, cookie)
	response.Success(c, resp)
}

func (h *Handler) Refresh(c *gin.Context) {
	refreshToken := ""
	if cookie, err := c.Cookie(h.config.Cookie.Name); err == nil {
		refreshToken = cookie
	} else {
		h.writeRefreshCookie(c, &RefreshCookie{Clear: true})
	}
	resp, cookie, err := h.service.Refresh(c.Request.Context(), &RefreshRequest{RefreshToken: refreshToken})
	if cookie != nil {
		h.writeRefreshCookie(c, cookie)
	}
	if err != nil {
		response.Fail(c, toAppError(err))
		return
	}
	response.Success(c, resp)
}

func (h *Handler) Logout(c *gin.Context) {
	refreshToken := ""
	if cookie, err := c.Cookie(h.config.Cookie.Name); err == nil {
		refreshToken = cookie
	}
	principal, _ := security.PrincipalFromContext(c.Request.Context())
	var sessionID, publicID string
	if principal != nil {
		sessionID = principal.SessionID
		publicID = principal.PublicID
	}
	if err := h.service.Logout(c.Request.Context(), &LogoutRequest{RefreshToken: refreshToken, SessionID: sessionID, PublicID: publicID}); err != nil {
		response.Fail(c, toAppError(err))
		return
	}
	h.writeRefreshCookie(c, &RefreshCookie{Clear: true})
	response.Success(c, gin.H{"success": true})
}

func (h *Handler) Me(c *gin.Context) {
	resp, err := h.service.Me(c.Request.Context())
	if err != nil {
		response.Fail(c, toAppError(err))
		return
	}
	response.Success(c, resp)
}

func (h *Handler) writeRefreshCookie(c *gin.Context, cookie *RefreshCookie) {
	if cookie == nil {
		return
	}
	value := cookie.Value
	maxAge := 0
	expires := cookie.ExpiresAt
	if cookie.Clear {
		value = ""
		maxAge = -1
		expires = time.Unix(0, 0).UTC()
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     h.config.Cookie.Name,
		Value:    value,
		Path:     h.config.Cookie.Path,
		Domain:   h.config.Cookie.Domain,
		HttpOnly: h.config.Cookie.HTTPOnly,
		Secure:   h.config.Cookie.Secure,
		SameSite: mapSameSite(h.config.Cookie.SameSite),
		Expires:  expires,
		MaxAge:   maxAge,
	})
}

func mapSameSite(raw string) http.SameSite {
	switch raw {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	case "lax":
		fallthrough
	default:
		return http.SameSiteLaxMode
	}
}

func toAppError(err error) *apperrors.AppError {
	if appErr, ok := err.(*apperrors.AppError); ok {
		return appErr
	}
	return apperrors.New("INTERNAL_ERROR", "internal error", http.StatusInternalServerError, err)
}
