package auth

import (
	"errors"
	"net/http"

	apperrors "backend/internal/platform/errors"
)

var (
	ErrInvalidCredentials = apperrors.New("AUTH_INVALID_CREDENTIALS", "账号或密码错误", http.StatusUnauthorized, nil)
	ErrUnauthorized       = apperrors.New("AUTH_UNAUTHORIZED", "未登录", http.StatusUnauthorized, nil)
	ErrRefreshInvalid     = apperrors.New("AUTH_REFRESH_INVALID", "登录状态无效", http.StatusUnauthorized, nil)
	ErrRefreshReused      = apperrors.New("AUTH_REFRESH_REUSED", "登录状态无效", http.StatusUnauthorized, nil)
	ErrAccountDisabled    = apperrors.New("AUTH_ACCOUNT_DISABLED", "账号不可用", http.StatusForbidden, nil)
	ErrAccountLocked      = apperrors.New("AUTH_ACCOUNT_LOCKED", "账号已锁定", http.StatusLocked, nil)
	ErrUserNotFound       = errors.New("user not found")
)
