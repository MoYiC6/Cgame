package auth

import (
	"context"
	"time"

	"backend/internal/platform/database"
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

type noopRepository struct {
	dbtx database.DBTX
}

func NewRepository() Repository {
	return &noopRepository{}
}

func (r *noopRepository) ListUserRoles(ctx context.Context, userID int64) ([]string, error) {
	_ = r.executor(ctx)
	return nil, nil
}

func (r *noopRepository) ListUserPermissions(ctx context.Context, userID int64) ([]string, error) {
	_ = r.executor(ctx)
	return nil, nil
}

func (r *noopRepository) CreateSession(ctx context.Context, session *AuthSession) error {
	_ = r.executor(ctx)
	return nil
}

func (r *noopRepository) GetSessionByID(ctx context.Context, sessionID string) (*AuthSession, error) {
	_ = r.executor(ctx)
	return nil, nil
}

func (r *noopRepository) RevokeSession(ctx context.Context, sessionID string) error {
	_ = r.executor(ctx)
	return nil
}

func (r *noopRepository) CreateRefreshToken(ctx context.Context, token *RefreshToken) error {
	_ = r.executor(ctx)
	return nil
}

func (r *noopRepository) GetRefreshTokenByHashForUpdate(ctx context.Context, tokenHash string) (*RefreshToken, error) {
	_ = r.executor(ctx)
	return nil, nil
}

func (r *noopRepository) MarkRefreshTokenUsed(ctx context.Context, tokenID int64, replacedByTokenID int64) (bool, error) {
	_ = r.executor(ctx)
	return false, nil
}

func (r *noopRepository) RevokeRefreshTokenFamily(ctx context.Context, familyID string) error {
	_ = r.executor(ctx)
	return nil
}

func (r *noopRepository) RevokeRefreshTokensBySessionID(ctx context.Context, sessionID string) error {
	_ = r.executor(ctx)
	return nil
}

func (r *noopRepository) UpdateLastLoginAt(ctx context.Context, userID int64, at time.Time) error {
	_ = r.executor(ctx)
	return nil
}

func (r *noopRepository) CreateLoginAttempt(ctx context.Context, attempt *LoginAttempt) error {
	_ = r.executor(ctx)
	return nil
}

func (r *noopRepository) CreateAuditLog(ctx context.Context, log *AuditLog) error {
	_ = r.executor(ctx)
	return nil
}

func (r *noopRepository) executor(ctx context.Context) database.DBTX {
	return database.ExecutorFromContext(ctx, r.dbtx)
}
