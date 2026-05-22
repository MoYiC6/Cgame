package auth

import (
	"context"
	"time"
)

type Repository interface {
	ListUserRoles(ctx context.Context, userID int64) ([]string, error)
	ListUserPermissions(ctx context.Context, userID int64) ([]string, error)
	CreateSession(ctx context.Context, session *AuthSession) error
	GetSessionByID(ctx context.Context, sessionID string) (*AuthSession, error)
	RevokeSession(ctx context.Context, sessionID string) error
	CreateRefreshToken(ctx context.Context, token *RefreshToken) error
	GetRefreshTokenByHashForUpdate(ctx context.Context, tokenHash string) (*RefreshToken, error)
	MarkRefreshTokenUsed(ctx context.Context, tokenID int64, replacedByTokenID int64) (bool, error)
	RevokeRefreshTokenFamily(ctx context.Context, familyID string) error
	RevokeRefreshTokensBySessionID(ctx context.Context, sessionID string) error
	UpdateLastLoginAt(ctx context.Context, userID int64, at time.Time) error
	CreateLoginAttempt(ctx context.Context, attempt *LoginAttempt) error
	CreateAuditLog(ctx context.Context, log *AuditLog) error
}
