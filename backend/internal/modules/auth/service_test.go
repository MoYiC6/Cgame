package auth

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"backend/internal/modules/user"
	"backend/internal/platform/database"
	apperrors "backend/internal/platform/errors"
	"backend/internal/platform/observability"
	"backend/internal/platform/security"
)

func TestServiceLoginSuccess(t *testing.T) {
	ctx := context.Background()
	hasher := security.NewArgon2idHasher(19456, 2, 1, "pepper")
	hash, err := hasher.Hash("secret-password")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	userRepo := &stubUserReader{user: &user.User{ID: 1, PublicID: "usr_123", Email: "admin@example.com", PasswordHash: hash, Status: user.StatusActive}}
	authRepo := &stubAuthRepository{roles: []string{"admin"}, permissions: []string{"order:read", "order:read"}}
	tokenManager := security.NewHMACTokenManager(security.HMACTokenConfig{
		Issuer:         "backend",
		Audience:       "admin-api",
		KeyID:          "test-key",
		Secret:         []byte("01234567890123456789012345678901"),
		AccessTokenTTL: 15 * time.Minute,
		ClockSkew:      30 * time.Second,
	})

	svc := NewService(userRepo, authRepo, database.NoopTxManager{}, hasher, tokenManager, security.CryptoRandomTokenGenerator{}, ServiceConfig{RefreshTokenTTL: 24 * time.Hour, RefreshCookieName: "refresh_token"})
	resp, cookie, err := svc.Login(ctx, &LoginRequest{Identifier: "admin@example.com", Password: "secret-password"})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if resp.AccessToken == "" || cookie == nil || cookie.Value == "" {
		t.Fatalf("expected access token and refresh cookie, got resp=%+v cookie=%+v", resp, cookie)
	}
	if len(resp.User.Permissions) != 1 || resp.User.Permissions[0] != "order:read" {
		t.Fatalf("expected deduped permissions, got %#v", resp.User.Permissions)
	}
}

func TestServiceLoginInvalidCredentialsWritesAttemptAndAudit(t *testing.T) {
	ctx := context.Background()
	hasher := security.NewArgon2idHasher(19456, 2, 1, "pepper")
	tokenManager := security.NewHMACTokenManager(security.HMACTokenConfig{Issuer: "backend", Audience: "admin-api", KeyID: "test-key", Secret: []byte("01234567890123456789012345678901"), AccessTokenTTL: 15 * time.Minute, ClockSkew: 30 * time.Second})

	userRepo := &stubUserReader{err: ErrUserNotFound}
	authRepo := &stubAuthRepository{}
	svc := NewService(userRepo, authRepo, database.NoopTxManager{}, hasher, tokenManager, security.CryptoRandomTokenGenerator{}, ServiceConfig{RefreshTokenTTL: 24 * time.Hour, RefreshCookieName: "refresh_token"})

	_, _, err := svc.Login(ctx, &LoginRequest{Identifier: "missing@example.com", Password: "secret-password"})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
	if len(authRepo.loginAttempts) != 1 {
		t.Fatalf("expected one login attempt, got %d", len(authRepo.loginAttempts))
	}
	if len(authRepo.auditLogs) != 1 || authRepo.auditLogs[0].EventType != "login_failed" {
		t.Fatalf("expected login_failed audit log, got %#v", authRepo.auditLogs)
	}
}

func TestServiceRefreshReuseReturnsAuthRefreshReused(t *testing.T) {
	ctx := observability.WithRequestID(context.Background(), "req-refresh-reuse")
	ctx = observability.WithTraceID(ctx, "trace-refresh-reuse")

	authRepo := &stubAuthRepository{
		refreshTokenByHash: map[string]*RefreshToken{
			sha256Hex("refresh-reused"): {
				ID:        10,
				UserID:    1,
				SessionID: "ses_reused",
				FamilyID:  "fam_reused",
				TokenHash: sha256Hex("refresh-reused"),
				UsedAt:    ptrTime(time.Now().UTC().Add(-time.Minute)),
				ExpiresAt: time.Now().UTC().Add(time.Hour),
				CreatedAt: time.Now().UTC().Add(-2 * time.Hour),
			},
		},
	}

	svc := NewService(&stubUserReader{}, authRepo, database.NoopTxManager{}, nil, nil, nil, ServiceConfig{RefreshTokenTTL: 24 * time.Hour, RefreshCookieName: "refresh_token"})

	resp, cookie, err := svc.Refresh(ctx, &RefreshRequest{RefreshToken: "refresh-reused"})
	if !errors.Is(err, ErrRefreshReused) {
		t.Fatalf("expected ErrRefreshReused, got %v", err)
	}
	if resp != nil {
		t.Fatalf("expected nil response, got %+v", resp)
	}
	if cookie == nil || !cookie.Clear {
		t.Fatalf("expected clear cookie, got %+v", cookie)
	}
	if !reflect.DeepEqual(authRepo.revokedFamilies, []string{"fam_reused"}) {
		t.Fatalf("expected family revoke, got %#v", authRepo.revokedFamilies)
	}
	if !reflect.DeepEqual(authRepo.revokedSessions, []string{"ses_reused"}) {
		t.Fatalf("expected session revoke, got %#v", authRepo.revokedSessions)
	}
	if len(authRepo.auditLogs) != 1 || authRepo.auditLogs[0].EventType != "refresh_reuse_detected" {
		t.Fatalf("expected refresh_reuse_detected audit, got %#v", authRepo.auditLogs)
	}
}

func TestServiceRefreshPasswordChangedRejectsAndClearsFamily(t *testing.T) {
	passwordChangedAt := time.Now().UTC().Add(-30 * time.Minute)
	issuedAt := passwordChangedAt.Add(-time.Hour)

	authRepo := &stubAuthRepository{
		refreshTokenByHash: map[string]*RefreshToken{
			sha256Hex("refresh-old-password"): {
				ID:        20,
				UserID:    7,
				SessionID: "ses_password",
				FamilyID:  "fam_password",
				TokenHash: sha256Hex("refresh-old-password"),
				ExpiresAt: time.Now().UTC().Add(time.Hour),
				CreatedAt: issuedAt,
			},
		},
		sessionsByID: map[string]*AuthSession{
			"ses_password": {
				ID:        "ses_password",
				UserID:    7,
				Status:    "active",
				CreatedAt: issuedAt,
				ExpiresAt: time.Now().UTC().Add(time.Hour),
			},
		},
	}
	userRepo := &stubUserReader{userByID: map[int64]*user.User{
		7: {ID: 7, PublicID: "usr_7", Email: "admin@example.com", Status: user.StatusActive, PasswordChangedAt: &passwordChangedAt},
	}}

	svc := NewService(userRepo, authRepo, database.NoopTxManager{}, nil, nil, nil, ServiceConfig{RefreshTokenTTL: 24 * time.Hour, RefreshCookieName: "refresh_token"})

	resp, cookie, err := svc.Refresh(context.Background(), &RefreshRequest{RefreshToken: "refresh-old-password"})
	if !errors.Is(err, ErrRefreshInvalid) {
		t.Fatalf("expected ErrRefreshInvalid, got %v", err)
	}
	if resp != nil {
		t.Fatalf("expected nil response, got %+v", resp)
	}
	if cookie == nil || !cookie.Clear {
		t.Fatalf("expected clear cookie, got %+v", cookie)
	}
	if !reflect.DeepEqual(authRepo.revokedFamilies, []string{"fam_password"}) {
		t.Fatalf("expected family revoke, got %#v", authRepo.revokedFamilies)
	}
	if !reflect.DeepEqual(authRepo.revokedSessions, []string{"ses_password"}) {
		t.Fatalf("expected session revoke, got %#v", authRepo.revokedSessions)
	}
	if len(authRepo.auditLogs) != 1 || authRepo.auditLogs[0].EventType != "session_revoked" {
		t.Fatalf("expected session_revoked audit, got %#v", authRepo.auditLogs)
	}
}

func TestServiceLogoutWithSessionMismatchRevokesBoth(t *testing.T) {
	ctx := security.WithPrincipal(context.Background(), &security.Principal{PublicID: "usr_123", SessionID: "ses_access", Roles: []string{"admin"}, Permissions: []string{"order:read"}})

	authRepo := &stubAuthRepository{
		refreshTokenByHash: map[string]*RefreshToken{
			sha256Hex("refresh-logout"): {
				ID:        30,
				UserID:    9,
				SessionID: "ses_cookie",
				FamilyID:  "fam_logout",
				TokenHash: sha256Hex("refresh-logout"),
				ExpiresAt: time.Now().UTC().Add(time.Hour),
				CreatedAt: time.Now().UTC().Add(-time.Hour),
			},
		},
	}
	svc := NewService(&stubUserReader{}, authRepo, database.NoopTxManager{}, nil, nil, nil, ServiceConfig{RefreshTokenTTL: 24 * time.Hour, RefreshCookieName: "refresh_token"})

	err := svc.Logout(ctx, &LogoutRequest{RefreshToken: "refresh-logout"})
	if err != nil {
		t.Fatalf("Logout() error = %v", err)
	}
	if !reflect.DeepEqual(authRepo.revokedFamilies, []string{"fam_logout"}) {
		t.Fatalf("expected family revoke, got %#v", authRepo.revokedFamilies)
	}
	if !reflect.DeepEqual(authRepo.revokedSessions, []string{"ses_cookie", "ses_access"}) {
		t.Fatalf("expected both sessions revoked, got %#v", authRepo.revokedSessions)
	}
	if len(authRepo.auditLogs) != 2 {
		t.Fatalf("expected two audit logs, got %#v", authRepo.auditLogs)
	}
	if authRepo.auditLogs[0].EventType != "session_revoked" {
		t.Fatalf("expected first audit event session_revoked, got %#v", authRepo.auditLogs)
	}
	if authRepo.auditLogs[1].EventType != "logout_success" {
		t.Fatalf("expected second audit event logout_success, got %#v", authRepo.auditLogs)
	}
	if mismatch, _ := authRepo.auditLogs[1].MetadataJSON["session_mismatch"].(bool); !mismatch {
		t.Fatalf("expected logout_success metadata to include session_mismatch=true, got %#v", authRepo.auditLogs[1].MetadataJSON)
	}
}

func TestServiceMeReturnsSnapshot(t *testing.T) {
	p := &security.Principal{PublicID: "usr_123", SessionID: "ses_123", Roles: []string{"admin"}, Permissions: []string{"order:read"}}
	ctx := security.WithPrincipal(context.Background(), p)
	svc := NewService(nil, &stubAuthRepository{}, database.NoopTxManager{}, nil, nil, nil, ServiceConfig{})
	resp, err := svc.Me(ctx)
	if err != nil {
		t.Fatalf("Me() error = %v", err)
	}
	if resp.User.ID != "usr_123" || resp.SessionID != "ses_123" {
		t.Fatalf("unexpected me response: %+v", resp)
	}
}

type stubUserReader struct {
	user     *user.User
	userByID map[int64]*user.User
	err      error
}

func (s *stubUserReader) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	return s.user, s.err
}
func (s *stubUserReader) GetByID(ctx context.Context, userID int64) (*user.User, error) {
	if s.userByID == nil {
		return nil, nil
	}
	return s.userByID[userID], nil
}

type stubAuthRepository struct {
	roles                []string
	permissions          []string
	loginAttempts        []*LoginAttempt
	auditLogs            []*AuditLog
	sessions             []*AuthSession
	refreshTokens        []*RefreshToken
	refreshTokenByHash   map[string]*RefreshToken
	sessionsByID         map[string]*AuthSession
	revokedFamilies      []string
	revokedSessions      []string
	revokedSessionTokens []string
	markUsedOK           bool
	usedTokenPairs       [][2]int64
}

func (s *stubAuthRepository) ListUserRoles(ctx context.Context, userID int64) ([]string, error) {
	return s.roles, nil
}
func (s *stubAuthRepository) ListUserPermissions(ctx context.Context, userID int64) ([]string, error) {
	return s.permissions, nil
}
func (s *stubAuthRepository) CreateSession(ctx context.Context, session *AuthSession) error {
	s.sessions = append(s.sessions, session)
	return nil
}
func (s *stubAuthRepository) GetSessionByID(ctx context.Context, sessionID string) (*AuthSession, error) {
	if s.sessionsByID == nil {
		return nil, nil
	}
	return s.sessionsByID[sessionID], nil
}
func (s *stubAuthRepository) RevokeSession(ctx context.Context, sessionID string) error {
	s.revokedSessions = append(s.revokedSessions, sessionID)
	return nil
}
func (s *stubAuthRepository) CreateRefreshToken(ctx context.Context, token *RefreshToken) error {
	s.refreshTokens = append(s.refreshTokens, token)
	return nil
}
func (s *stubAuthRepository) GetRefreshTokenByHashForUpdate(ctx context.Context, tokenHash string) (*RefreshToken, error) {
	if s.refreshTokenByHash == nil {
		return nil, nil
	}
	return s.refreshTokenByHash[tokenHash], nil
}
func (s *stubAuthRepository) MarkRefreshTokenUsed(ctx context.Context, tokenID int64, replacedByTokenID int64) (bool, error) {
	s.usedTokenPairs = append(s.usedTokenPairs, [2]int64{tokenID, replacedByTokenID})
	if !s.markUsedOK {
		return false, nil
	}
	return true, nil
}
func (s *stubAuthRepository) RevokeRefreshTokenFamily(ctx context.Context, familyID string) error {
	s.revokedFamilies = append(s.revokedFamilies, familyID)
	return nil
}
func (s *stubAuthRepository) RevokeRefreshTokensBySessionID(ctx context.Context, sessionID string) error {
	s.revokedSessionTokens = append(s.revokedSessionTokens, sessionID)
	return nil
}
func (s *stubAuthRepository) UpdateLastLoginAt(ctx context.Context, userID int64, at time.Time) error {
	return nil
}
func (s *stubAuthRepository) CreateLoginAttempt(ctx context.Context, attempt *LoginAttempt) error {
	s.loginAttempts = append(s.loginAttempts, attempt)
	return nil
}
func (s *stubAuthRepository) CreateAuditLog(ctx context.Context, log *AuditLog) error {
	s.auditLogs = append(s.auditLogs, log)
	return nil
}

func TestAuthErrorCodesExposeExpectedValues(t *testing.T) {
	if got := apperrors.Code(ErrRefreshInvalid); got != "AUTH_REFRESH_INVALID" {
		t.Fatalf("expected AUTH_REFRESH_INVALID, got %q", got)
	}
	if got := apperrors.Code(ErrRefreshReused); got != "AUTH_REFRESH_REUSED" {
		t.Fatalf("expected AUTH_REFRESH_REUSED, got %q", got)
	}
}
