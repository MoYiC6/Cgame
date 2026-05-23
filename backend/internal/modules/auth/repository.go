package auth

import (
	"context"
	"database/sql"
	"encoding/json"
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
	ListRecentFailedAttemptTimes(ctx context.Context, filter FailedLoginAttemptFilter) ([]time.Time, error)
	UpdateLastLoginAt(ctx context.Context, userID int64, at time.Time) error
	CreateLoginAttempt(ctx context.Context, attempt *LoginAttempt) error
	CreateAuditLog(ctx context.Context, log *AuditLog) error
}

type repository struct {
	dbtx database.DBTX
}

func NewRepository(dbtx database.DBTX) Repository {
	return &repository{dbtx: dbtx}
}

func (r *repository) ListUserRoles(ctx context.Context, userID int64) ([]string, error) {
	rows, err := r.executor(ctx).QueryContext(ctx, `
		SELECT r.code
		FROM roles r
		JOIN user_roles ur ON ur.role_id = r.id
		WHERE ur.user_id = $1
		ORDER BY r.code
	`, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var roles []string
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, err
		}
		roles = append(roles, code)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *repository) ListUserPermissions(ctx context.Context, userID int64) ([]string, error) {
	rows, err := r.executor(ctx).QueryContext(ctx, `
		SELECT DISTINCT p.code
		FROM permissions p
		JOIN role_permissions rp ON rp.permission_id = p.id
		JOIN user_roles ur ON ur.role_id = rp.role_id
		WHERE ur.user_id = $1
		ORDER BY p.code
	`, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var permissions []string
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, err
		}
		permissions = append(permissions, code)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return permissions, nil
}

func (r *repository) CreateSession(ctx context.Context, session *AuthSession) error {
	if session == nil {
		return nil
	}
	_, err := r.executor(ctx).ExecContext(ctx, `
		INSERT INTO auth_sessions (id, user_id, status, user_agent_hash, ip_hash, created_at, last_seen_at, revoked_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, session.ID, session.UserID, session.Status, nullString(session.UserAgentHash), nullString(session.IPHash), session.CreatedAt, session.LastSeenAt, session.RevokedAt, session.ExpiresAt)
	return err
}

func (r *repository) GetSessionByID(ctx context.Context, sessionID string) (*AuthSession, error) {
	row := r.executor(ctx).QueryRowContext(ctx, `
		SELECT id, user_id, status, user_agent_hash, ip_hash, created_at, last_seen_at, revoked_at, expires_at
		FROM auth_sessions
		WHERE id = $1
	`, sessionID)
	return scanAuthSession(row)
}

func (r *repository) RevokeSession(ctx context.Context, sessionID string) error {
	_, err := r.executor(ctx).ExecContext(ctx, `
		UPDATE auth_sessions
		SET status = 'revoked', revoked_at = NOW()
		WHERE id = $1 AND revoked_at IS NULL
	`, sessionID)
	return err
}

func (r *repository) CreateRefreshToken(ctx context.Context, token *RefreshToken) error {
	if token == nil {
		return nil
	}
	return r.executor(ctx).QueryRowContext(ctx, `
		INSERT INTO refresh_tokens (user_id, session_id, token_hash, family_id, replaced_by_token_id, revoked_at, used_at, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`, token.UserID, token.SessionID, token.TokenHash, token.FamilyID, token.ReplacedByTokenID, token.RevokedAt, token.UsedAt, token.ExpiresAt, token.CreatedAt).Scan(&token.ID)
}

func (r *repository) GetRefreshTokenByHashForUpdate(ctx context.Context, tokenHash string) (*RefreshToken, error) {
	row := r.executor(ctx).QueryRowContext(ctx, `
		SELECT id, user_id, session_id, token_hash, family_id, replaced_by_token_id, revoked_at, used_at, expires_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1
		FOR UPDATE
	`, tokenHash)
	return scanRefreshToken(row)
}

func (r *repository) MarkRefreshTokenUsed(ctx context.Context, tokenID int64, replacedByTokenID int64) (bool, error) {
	result, err := r.executor(ctx).ExecContext(ctx, `
		UPDATE refresh_tokens
		SET used_at = NOW(), replaced_by_token_id = $2
		WHERE id = $1 AND used_at IS NULL AND revoked_at IS NULL
	`, tokenID, replacedByTokenID)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected == 1, nil
}

func (r *repository) RevokeRefreshTokenFamily(ctx context.Context, familyID string) error {
	_, err := r.executor(ctx).ExecContext(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE family_id = $1 AND revoked_at IS NULL
	`, familyID)
	return err
}

func (r *repository) RevokeRefreshTokensBySessionID(ctx context.Context, sessionID string) error {
	_, err := r.executor(ctx).ExecContext(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE session_id = $1 AND revoked_at IS NULL
	`, sessionID)
	return err
}

func (r *repository) ListRecentFailedAttemptTimes(ctx context.Context, filter FailedLoginAttemptFilter) ([]time.Time, error) {
	query := `
		SELECT created_at
		FROM login_attempts
		WHERE success = false AND created_at >= $1
	`
	args := []any{filter.Since}
	if filter.UserID != nil {
		query += ` AND user_id = $2`
		args = append(args, *filter.UserID)
	} else if filter.IPHash != "" {
		query += ` AND ip_hash = $2`
		args = append(args, filter.IPHash)
	} else {
		query += ` AND identifier_hash = $2`
		args = append(args, filter.IdentifierHash)
	}
	query += ` ORDER BY created_at DESC`

	rows, err := r.executor(ctx).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	result := make([]time.Time, 0)
	for rows.Next() {
		var createdAt time.Time
		if err := rows.Scan(&createdAt); err != nil {
			return nil, err
		}
		result = append(result, createdAt)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *repository) UpdateLastLoginAt(ctx context.Context, userID int64, at time.Time) error {
	_, err := r.executor(ctx).ExecContext(ctx, `
		UPDATE users
		SET last_login_at = $2, updated_at = $2
		WHERE id = $1
	`, userID, at)
	return err
}

func (r *repository) CreateLoginAttempt(ctx context.Context, attempt *LoginAttempt) error {
	if attempt == nil {
		return nil
	}
	_, err := r.executor(ctx).ExecContext(ctx, `
		INSERT INTO login_attempts (identifier_hash, user_id, success, reason, ip_hash, user_agent_hash, request_id, trace_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, attempt.IdentifierHash, attempt.UserID, attempt.Success, attempt.Reason, nullString(attempt.IPHash), nullString(attempt.UserAgentHash), nullString(attempt.RequestID), nullString(attempt.TraceID), attempt.CreatedAt)
	return err
}

func (r *repository) CreateAuditLog(ctx context.Context, log *AuditLog) error {
	if log == nil {
		return nil
	}
	metadata, err := json.Marshal(log.MetadataJSON)
	if err != nil {
		return err
	}
	_, err = r.executor(ctx).ExecContext(ctx, `
		INSERT INTO audit_logs (event_type, result, user_public_id, session_id, request_id, trace_id, ip_hash, user_agent_hash, metadata_json, occurred_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, log.EventType, log.Result, nullString(log.UserPublicID), nullString(log.SessionID), nullString(log.RequestID), nullString(log.TraceID), nullString(log.IPHash), nullString(log.UserAgentHash), metadata, log.OccurredAt)
	return err
}

func (r *repository) executor(ctx context.Context) database.DBTX {
	return database.ExecutorFromContext(ctx, r.dbtx)
}

func scanAuthSession(row rowScanner) (*AuthSession, error) {
	if row == nil {
		return nil, nil
	}
	var session AuthSession
	var userAgentHash sql.NullString
	var ipHash sql.NullString
	err := row.Scan(
		&session.ID,
		&session.UserID,
		&session.Status,
		&userAgentHash,
		&ipHash,
		&session.CreatedAt,
		&session.LastSeenAt,
		&session.RevokedAt,
		&session.ExpiresAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	session.UserAgentHash = userAgentHash.String
	session.IPHash = ipHash.String
	return &session, nil
}

func scanRefreshToken(row rowScanner) (*RefreshToken, error) {
	if row == nil {
		return nil, nil
	}
	var token RefreshToken
	err := row.Scan(
		&token.ID,
		&token.UserID,
		&token.SessionID,
		&token.TokenHash,
		&token.FamilyID,
		&token.ReplacedByTokenID,
		&token.RevokedAt,
		&token.UsedAt,
		&token.ExpiresAt,
		&token.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &token, nil
}

func nullString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

type rowScanner interface {
	Scan(dest ...any) error
}
