package auth

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"backend/internal/modules/user"
	"backend/internal/platform/config"
	"backend/internal/platform/database"
)

func TestServiceRefreshReusePersistsRevocationAcrossRollbackBoundary(t *testing.T) {
	ctx := context.Background()
	h, err := newAuthRepositoryHarness(ctx, t)
	if err != nil {
		t.Fatalf("newAuthRepositoryHarness() error = %v", err)
	}
	defer func() {
		if err := h.Close(ctx); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	}()

	if err := h.ApplyMigrations(ctx); err != nil {
		t.Fatalf("ApplyMigrations() error = %v", err)
	}

	userID := h.insertUser(ctx, t, "usr_refresh_reuse", "refresh-reuse@example.com")
	h.insertSession(ctx, t, "ses_refresh_reuse", userID)
	refreshTokenID := h.insertRefreshToken(ctx, t, userID, "ses_refresh_reuse", sha256Hex("refresh-reused-db"), "fam_refresh_reuse")
	usedAt := time.Now().UTC().Add(-time.Minute).Truncate(time.Second)
	_, err = h.db.ExecContext(ctx, `UPDATE refresh_tokens SET used_at = $2 WHERE id = $1`, refreshTokenID, usedAt)
	if err != nil {
		t.Fatalf("mark refresh token used error = %v", err)
	}

	sqlDB, err := database.NewSQLDB(config.DBConfig{DSN: h.dbDSN})
	if err != nil {
		t.Fatalf("NewSQLDB() error = %v", err)
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			t.Fatalf("Close() sqlDB error = %v", err)
		}
	}()

	repo := NewRepository(sqlDB)
	userRepo := user.NewRepository(sqlDB)
	txManager := database.NewSQLTxManager(sqlDB)
	svc := NewService(userRepo, repo, txManager, nil, nil, nil, ServiceConfig{RefreshTokenTTL: 24 * time.Hour, RefreshCookieName: "refresh_token"})

	resp, cookie, err := svc.Refresh(ctx, &RefreshRequest{RefreshToken: "refresh-reused-db"})
	if !errors.Is(err, ErrRefreshReused) {
		t.Fatalf("expected ErrRefreshReused, got %v", err)
	}
	if resp != nil {
		t.Fatalf("expected nil response, got %+v", resp)
	}
	if cookie == nil || !cookie.Clear {
		t.Fatalf("expected clear cookie, got %+v", cookie)
	}

	assertRevokedSessionAndFamily(t, ctx, h.db, "ses_refresh_reuse", "fam_refresh_reuse")
}

func TestServiceRefreshPasswordChangedPersistsRevocationAcrossRollbackBoundary(t *testing.T) {
	ctx := context.Background()
	h, err := newAuthRepositoryHarness(ctx, t)
	if err != nil {
		t.Fatalf("newAuthRepositoryHarness() error = %v", err)
	}
	defer func() {
		if err := h.Close(ctx); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	}()

	if err := h.ApplyMigrations(ctx); err != nil {
		t.Fatalf("ApplyMigrations() error = %v", err)
	}

	userID := h.insertUser(ctx, t, "usr_password_changed", "password-changed@example.com")
	sessionCreatedAt := time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Second)
	passwordChangedAt := sessionCreatedAt.Add(time.Hour)
	_, err = h.db.ExecContext(ctx, `UPDATE users SET password_changed_at = $2 WHERE id = $1`, userID, passwordChangedAt)
	if err != nil {
		t.Fatalf("update user password_changed_at error = %v", err)
	}
	_, err = h.db.ExecContext(ctx, `
		INSERT INTO auth_sessions (id, user_id, status, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5)
	`, "ses_password_changed", userID, "active", sessionCreatedAt, time.Now().UTC().Add(24*time.Hour).Truncate(time.Second))
	if err != nil {
		t.Fatalf("insert auth session error = %v", err)
	}
	_, err = h.db.ExecContext(ctx, `
		INSERT INTO refresh_tokens (user_id, session_id, token_hash, family_id, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, userID, "ses_password_changed", sha256Hex("refresh-password-changed"), "fam_password_changed", time.Now().UTC().Add(24*time.Hour).Truncate(time.Second), sessionCreatedAt)
	if err != nil {
		t.Fatalf("insert refresh token error = %v", err)
	}

	sqlDB, err := database.NewSQLDB(config.DBConfig{DSN: h.dbDSN})
	if err != nil {
		t.Fatalf("NewSQLDB() error = %v", err)
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			t.Fatalf("Close() sqlDB error = %v", err)
		}
	}()

	repo := NewRepository(sqlDB)
	userRepo := user.NewRepository(sqlDB)
	txManager := database.NewSQLTxManager(sqlDB)
	svc := NewService(userRepo, repo, txManager, nil, nil, nil, ServiceConfig{RefreshTokenTTL: 24 * time.Hour, RefreshCookieName: "refresh_token"})

	resp, cookie, err := svc.Refresh(ctx, &RefreshRequest{RefreshToken: "refresh-password-changed"})
	if !errors.Is(err, ErrRefreshInvalid) {
		t.Fatalf("expected ErrRefreshInvalid, got %v", err)
	}
	if resp != nil {
		t.Fatalf("expected nil response, got %+v", resp)
	}
	if cookie == nil || !cookie.Clear {
		t.Fatalf("expected clear cookie, got %+v", cookie)
	}

	assertRevokedSessionAndFamily(t, ctx, h.db, "ses_password_changed", "fam_password_changed")
}

func assertRevokedSessionAndFamily(t *testing.T, ctx context.Context, db *sql.DB, sessionID string, familyID string) {
	t.Helper()

	var sessionStatus string
	var sessionRevokedAt sql.NullTime
	err := db.QueryRowContext(ctx, `SELECT status, revoked_at FROM auth_sessions WHERE id = $1`, sessionID).Scan(&sessionStatus, &sessionRevokedAt)
	if err != nil {
		t.Fatalf("query session revoke state error = %v", err)
	}
	if sessionStatus != "revoked" || !sessionRevokedAt.Valid {
		t.Fatalf("expected session to stay revoked, got status=%q revoked_at=%v", sessionStatus, sessionRevokedAt)
	}

	var revokedCount int
	err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM refresh_tokens WHERE family_id = $1 AND revoked_at IS NOT NULL`, familyID).Scan(&revokedCount)
	if err != nil {
		t.Fatalf("count revoked family tokens error = %v", err)
	}
	if revokedCount == 0 {
		t.Fatal("expected refresh token family revoke to persist")
	}
}
