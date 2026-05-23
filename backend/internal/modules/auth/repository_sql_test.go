package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestRepositoryListUserRolesAndPermissions(t *testing.T) {
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

	userID := h.insertUser(ctx, t, "usr_roles", "roles@example.com")
	adminRoleID := h.lookupRoleID(ctx, t, "admin")
	systemRoleID := h.lookupRoleID(ctx, t, "system")
	permissionID := h.lookupPermissionID(ctx, t, "admin:user:disable")
	h.linkUserRole(ctx, t, userID, adminRoleID)
	h.linkUserRole(ctx, t, userID, systemRoleID)
	h.linkRolePermission(ctx, t, adminRoleID, permissionID)

	repo := NewRepository(h.db)
	roles, err := repo.ListUserRoles(ctx, userID)
	if err != nil {
		t.Fatalf("ListUserRoles() error = %v", err)
	}
	if len(roles) < 2 || roles[0] != "admin" || roles[1] != "system" {
		t.Fatalf("unexpected roles: %#v", roles)
	}

	permissions, err := repo.ListUserPermissions(ctx, userID)
	if err != nil {
		t.Fatalf("ListUserPermissions() error = %v", err)
	}
	if len(permissions) == 0 {
		t.Fatal("expected permissions, got none")
	}
	if permissions[0] != "admin:user:disable" {
		t.Fatalf("expected sorted permissions, got %#v", permissions)
	}
}

func TestRepositoryCreateSessionAndGetSessionByID(t *testing.T) {
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

	userID := h.insertUser(ctx, t, "usr_session", "session@example.com")
	now := time.Now().UTC().Truncate(time.Second)
	repo := NewRepository(h.db)
	err = repo.CreateSession(ctx, &AuthSession{
		ID:         "ses_123",
		UserID:     userID,
		Status:     "active",
		CreatedAt:  now,
		LastSeenAt: ptrTime(now),
		ExpiresAt:  now.Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	session, err := repo.GetSessionByID(ctx, "ses_123")
	if err != nil {
		t.Fatalf("GetSessionByID() error = %v", err)
	}
	if session == nil {
		t.Fatal("expected session, got nil")
	}
	if session.UserID != userID || session.Status != "active" {
		t.Fatalf("unexpected session: %+v", session)
	}
}

func TestRepositoryCreateRefreshTokenAndGetByHash(t *testing.T) {
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

	userID := h.insertUser(ctx, t, "usr_refresh", "refresh@example.com")
	h.insertSession(ctx, t, "ses_refresh", userID)
	repo := NewRepository(h.db)
	now := time.Now().UTC().Truncate(time.Second)
	err = repo.CreateRefreshToken(ctx, &RefreshToken{
		UserID:    userID,
		SessionID: "ses_refresh",
		TokenHash: "hash_refresh",
		FamilyID:  "fam_refresh",
		ExpiresAt: now.Add(24 * time.Hour),
		CreatedAt: now,
	})
	if err != nil {
		t.Fatalf("CreateRefreshToken() error = %v", err)
	}

	token, err := repo.GetRefreshTokenByHashForUpdate(ctx, "hash_refresh")
	if err != nil {
		t.Fatalf("GetRefreshTokenByHashForUpdate() error = %v", err)
	}
	if token == nil {
		t.Fatal("expected refresh token, got nil")
	}
	if token.SessionID != "ses_refresh" || token.FamilyID != "fam_refresh" {
		t.Fatalf("unexpected refresh token: %+v", token)
	}
}

func TestRepositoryMarkRefreshTokenUsed(t *testing.T) {
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

	userID := h.insertUser(ctx, t, "usr_mark", "mark@example.com")
	h.insertSession(ctx, t, "ses_mark", userID)
	oldID := h.insertRefreshToken(ctx, t, userID, "ses_mark", "hash_old", "fam_mark")
	newID := h.insertRefreshToken(ctx, t, userID, "ses_mark", "hash_new", "fam_mark")
	repo := NewRepository(h.db)

	used, err := repo.MarkRefreshTokenUsed(ctx, oldID, newID)
	if err != nil {
		t.Fatalf("MarkRefreshTokenUsed() error = %v", err)
	}
	if !used {
		t.Fatal("expected used=true")
	}

	var usedAt sql.NullTime
	var replacedBy sql.NullInt64
	err = h.db.QueryRowContext(ctx, `SELECT used_at, replaced_by_token_id FROM refresh_tokens WHERE id = $1`, oldID).Scan(&usedAt, &replacedBy)
	if err != nil {
		t.Fatalf("query marked token error = %v", err)
	}
	if !usedAt.Valid || !replacedBy.Valid || replacedBy.Int64 != newID {
		t.Fatalf("expected token used and replaced_by=%d, got used_at=%v replaced_by=%v", newID, usedAt, replacedBy)
	}
}

func TestRepositoryRevokeRefreshTokenFamily(t *testing.T) {
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

	userID := h.insertUser(ctx, t, "usr_family", "family@example.com")
	h.insertSession(ctx, t, "ses_family", userID)
	h.insertRefreshToken(ctx, t, userID, "ses_family", "hash_1", "fam_revoke")
	h.insertRefreshToken(ctx, t, userID, "ses_family", "hash_2", "fam_revoke")
	repo := NewRepository(h.db)

	err = repo.RevokeRefreshTokenFamily(ctx, "fam_revoke")
	if err != nil {
		t.Fatalf("RevokeRefreshTokenFamily() error = %v", err)
	}

	var revokedCount int
	err = h.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM refresh_tokens WHERE family_id = $1 AND revoked_at IS NOT NULL`, "fam_revoke").Scan(&revokedCount)
	if err != nil {
		t.Fatalf("count revoked family tokens error = %v", err)
	}
	if revokedCount != 2 {
		t.Fatalf("expected 2 revoked family tokens, got %d", revokedCount)
	}
}

func TestRepositoryWriteLoginAttemptAndAuditLog(t *testing.T) {
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

	repo := NewRepository(h.db)
	now := time.Now().UTC().Truncate(time.Second)
	err = repo.CreateLoginAttempt(ctx, &LoginAttempt{
		IdentifierHash: "ident_hash",
		UserID:         ptrInt64(77),
		Success:        false,
		Reason:         "invalid_credentials",
		IPHash:         "ip_hash",
		UserAgentHash:  "ua_hash",
		RequestID:      "req_auth",
		TraceID:        "trace_auth",
		CreatedAt:      now,
	})
	if err != nil {
		t.Fatalf("CreateLoginAttempt() error = %v", err)
	}
	err = repo.CreateAuditLog(ctx, &AuditLog{
		EventType:    "login_failed",
		Result:       "failure",
		UserPublicID: "usr_audit",
		SessionID:    "ses_audit",
		RequestID:    "req_auth",
		TraceID:      "trace_auth",
		MetadataJSON: map[string]any{"reason": "invalid_credentials"},
		OccurredAt:   now,
	})
	if err != nil {
		t.Fatalf("CreateAuditLog() error = %v", err)
	}

	var attempts int
	err = h.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM login_attempts WHERE identifier_hash = $1`, "ident_hash").Scan(&attempts)
	if err != nil {
		t.Fatalf("count login_attempts error = %v", err)
	}
	if attempts != 1 {
		t.Fatalf("expected 1 login_attempt, got %d", attempts)
	}

	var storedUserID sql.NullInt64
	var storedIPHash sql.NullString
	var storedUserAgentHash sql.NullString
	err = h.db.QueryRowContext(ctx, `SELECT user_id, ip_hash, user_agent_hash FROM login_attempts WHERE identifier_hash = $1`, "ident_hash").Scan(&storedUserID, &storedIPHash, &storedUserAgentHash)
	if err != nil {
		t.Fatalf("query login_attempts detail error = %v", err)
	}
	if !storedUserID.Valid || storedUserID.Int64 != 77 {
		t.Fatalf("expected user_id 77, got %#v", storedUserID)
	}
	if storedIPHash.String != "ip_hash" || storedUserAgentHash.String != "ua_hash" {
		t.Fatalf("expected ip/user_agent hashes persisted, got ip=%#v ua=%#v", storedIPHash, storedUserAgentHash)
	}

	var metadataRaw []byte
	err = h.db.QueryRowContext(ctx, `SELECT metadata_json FROM audit_logs WHERE event_type = $1`, "login_failed").Scan(&metadataRaw)
	if err != nil {
		t.Fatalf("query audit_logs metadata error = %v", err)
	}
	var metadata map[string]any
	if err := json.Unmarshal(metadataRaw, &metadata); err != nil {
		t.Fatalf("json.Unmarshal metadata error = %v", err)
	}
	if metadata["reason"] != "invalid_credentials" {
		t.Fatalf("expected metadata reason invalid_credentials, got %#v", metadata)
	}
}

func TestRepositoryListRecentFailedAttemptTimesByUserID(t *testing.T) {
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

	userID := h.insertUser(ctx, t, "usr_recent", "recent@example.com")
	now := time.Now().UTC().Truncate(time.Second)
	_, err = h.db.ExecContext(ctx, `
		INSERT INTO login_attempts (identifier_hash, user_id, success, reason, created_at)
		VALUES ($1, $2, false, $3, $4), ($1, $2, false, $3, $5), ($1, $2, true, $6, $7)
	`, "ident_hash_recent", userID, "invalid_credentials", now.Add(-2*time.Minute), now.Add(-4*time.Minute), "success", now)
	if err != nil {
		t.Fatalf("insert login attempts error = %v", err)
	}

	repo := NewRepository(h.db)
	recent, err := repo.ListRecentFailedAttemptTimes(ctx, FailedLoginAttemptFilter{UserID: ptrInt64(userID), Since: now.Add(-10 * time.Minute)})
	if err != nil {
		t.Fatalf("ListRecentFailedAttemptTimes() error = %v", err)
	}
	if len(recent) != 2 {
		t.Fatalf("expected 2 recent failed attempts, got %#v", recent)
	}
}

type authRepositoryHarness struct {
	db        *sql.DB
	dbDSN     string
	container *postgres.PostgresContainer
}

func newAuthRepositoryHarness(ctx context.Context, t *testing.T) (*authRepositoryHarness, error) {
	t.Helper()
	configureAuthRepositoryDockerEnv(t)

	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("backend_test"),
		postgres.WithUsername("backend"),
		postgres.WithPassword("backend"),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2)),
	)
	if err != nil {
		return nil, fmt.Errorf("start postgres container: %w", err)
	}

	connString, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("postgres connection string: %w", err)
	}

	db, err := sql.Open("pgx", connString)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("sql open: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &authRepositoryHarness{db: db, dbDSN: connString, container: container}, nil
}

func (h *authRepositoryHarness) ApplyMigrations(ctx context.Context) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}
	migrationsDir, err := authMigrationsDir()
	if err != nil {
		return err
	}
	return goose.UpContext(ctx, h.db, migrationsDir)
}

func (h *authRepositoryHarness) Close(ctx context.Context) error {
	var errs []string
	if h.db != nil {
		if err := h.db.Close(); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if h.container != nil {
		if err := h.container.Terminate(ctx); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("close harness: %s", strings.Join(errs, "; "))
	}
	return nil
}

func (h *authRepositoryHarness) insertUser(ctx context.Context, t *testing.T, publicID, email string) int64 {
	t.Helper()
	var userID int64
	now := time.Now().UTC().Truncate(time.Second)
	err := h.db.QueryRowContext(ctx, `
		INSERT INTO users (public_id, email, password_hash, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, publicID, email, "hash", "active", now, now).Scan(&userID)
	if err != nil {
		t.Fatalf("insert user error = %v", err)
	}
	return userID
}

func (h *authRepositoryHarness) insertSession(ctx context.Context, t *testing.T, sessionID string, userID int64) {
	t.Helper()
	now := time.Now().UTC().Truncate(time.Second)
	_, err := h.db.ExecContext(ctx, `
		INSERT INTO auth_sessions (id, user_id, status, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5)
	`, sessionID, userID, "active", now, now.Add(24*time.Hour))
	if err != nil {
		t.Fatalf("insert session error = %v", err)
	}
}

func (h *authRepositoryHarness) insertRefreshToken(ctx context.Context, t *testing.T, userID int64, sessionID, tokenHash, familyID string) int64 {
	t.Helper()
	var tokenID int64
	now := time.Now().UTC().Truncate(time.Second)
	err := h.db.QueryRowContext(ctx, `
		INSERT INTO refresh_tokens (user_id, session_id, token_hash, family_id, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, userID, sessionID, tokenHash, familyID, now.Add(24*time.Hour), now).Scan(&tokenID)
	if err != nil {
		t.Fatalf("insert refresh token error = %v", err)
	}
	return tokenID
}

func (h *authRepositoryHarness) lookupRoleID(ctx context.Context, t *testing.T, code string) int64 {
	t.Helper()
	var roleID int64
	err := h.db.QueryRowContext(ctx, `SELECT id FROM roles WHERE code = $1`, code).Scan(&roleID)
	if err != nil {
		t.Fatalf("lookup role id error = %v", err)
	}
	return roleID
}

func (h *authRepositoryHarness) linkUserRole(ctx context.Context, t *testing.T, userID, roleID int64) {
	t.Helper()
	_, err := h.db.ExecContext(ctx, `
		INSERT INTO user_roles (user_id, role_id, created_at)
		VALUES ($1, $2, $3)
	`, userID, roleID, time.Now().UTC().Truncate(time.Second))
	if err != nil {
		t.Fatalf("link user role error = %v", err)
	}
}

func (h *authRepositoryHarness) lookupPermissionID(ctx context.Context, t *testing.T, code string) int64 {
	t.Helper()
	var permissionID int64
	err := h.db.QueryRowContext(ctx, `SELECT id FROM permissions WHERE code = $1`, code).Scan(&permissionID)
	if err != nil {
		t.Fatalf("lookup permission id error = %v", err)
	}
	return permissionID
}

func (h *authRepositoryHarness) linkRolePermission(ctx context.Context, t *testing.T, roleID, permissionID int64) {
	t.Helper()
	_, err := h.db.ExecContext(ctx, `
		INSERT INTO role_permissions (role_id, permission_id, created_at)
		VALUES ($1, $2, $3)
	`, roleID, permissionID, time.Now().UTC().Truncate(time.Second))
	if err != nil {
		t.Fatalf("link role permission error = %v", err)
	}
}

func configureAuthRepositoryDockerEnv(t *testing.T) {
	t.Helper()
	if endpoint := strings.TrimSpace(os.Getenv("DOCKER_HOST")); endpoint != "" {
		t.Setenv("DOCKER_HOST", endpoint)
	}
	if socketOverride := strings.TrimSpace(os.Getenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE")); socketOverride != "" {
		t.Setenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE", socketOverride)
	}
	colimaSocket := filepath.Join(os.Getenv("HOME"), ".colima", "default", "docker.sock")
	if _, err := os.Stat(colimaSocket); err == nil {
		t.Setenv("DOCKER_HOST", "unix://"+colimaSocket)
		t.Setenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE", "/var/run/docker.sock")
	}
}

func authMigrationsDir() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("resolve runtime caller for migrations dir")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", "..", "migrations")), nil
}
