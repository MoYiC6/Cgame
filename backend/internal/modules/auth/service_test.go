package auth

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
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

func TestServiceLoginDisabledUserReturnsForbiddenAndWritesAudit(t *testing.T) {
	ctx := context.Background()
	hasher := security.NewArgon2idHasher(19456, 2, 1, "pepper")
	hash, err := hasher.Hash("secret-password")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	userRepo := &stubUserReader{user: &user.User{ID: 1, PublicID: "usr_disabled", Email: "admin@example.com", PasswordHash: hash, Status: user.StatusDisabled}}
	authRepo := &stubAuthRepository{}
	tokenManager := security.NewHMACTokenManager(security.HMACTokenConfig{Issuer: "backend", Audience: "admin-api", KeyID: "test-key", Secret: []byte("01234567890123456789012345678901"), AccessTokenTTL: 15 * time.Minute, ClockSkew: 30 * time.Second})

	svc := NewService(userRepo, authRepo, database.NoopTxManager{}, hasher, tokenManager, security.CryptoRandomTokenGenerator{}, ServiceConfig{RefreshTokenTTL: 24 * time.Hour, RefreshCookieName: "refresh_token"})
	_, _, err = svc.Login(ctx, &LoginRequest{Identifier: "admin@example.com", Password: "secret-password"})
	if !errors.Is(err, ErrAccountDisabled) {
		t.Fatalf("expected ErrAccountDisabled, got %v", err)
	}
	if len(authRepo.loginAttempts) != 1 || authRepo.loginAttempts[0].Reason != "account_disabled" {
		t.Fatalf("expected account_disabled login attempt, got %#v", authRepo.loginAttempts)
	}
	if len(authRepo.auditLogs) != 1 || authRepo.auditLogs[0].EventType != "login_failed" {
		t.Fatalf("expected login_failed audit log, got %#v", authRepo.auditLogs)
	}
}

func TestServiceLoginLockedUserReturnsLockedAndWritesAudit(t *testing.T) {
	ctx := context.Background()
	hasher := security.NewArgon2idHasher(19456, 2, 1, "pepper")
	hash, err := hasher.Hash("secret-password")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	userRepo := &stubUserReader{user: &user.User{ID: 1, PublicID: "usr_locked", Email: "admin@example.com", PasswordHash: hash, Status: user.StatusLocked}}
	authRepo := &stubAuthRepository{}
	tokenManager := security.NewHMACTokenManager(security.HMACTokenConfig{Issuer: "backend", Audience: "admin-api", KeyID: "test-key", Secret: []byte("01234567890123456789012345678901"), AccessTokenTTL: 15 * time.Minute, ClockSkew: 30 * time.Second})

	svc := NewService(userRepo, authRepo, database.NoopTxManager{}, hasher, tokenManager, security.CryptoRandomTokenGenerator{}, ServiceConfig{RefreshTokenTTL: 24 * time.Hour, RefreshCookieName: "refresh_token"})
	_, _, err = svc.Login(ctx, &LoginRequest{Identifier: "admin@example.com", Password: "secret-password"})
	if !errors.Is(err, ErrAccountLocked) {
		t.Fatalf("expected ErrAccountLocked, got %v", err)
	}
	if len(authRepo.loginAttempts) != 1 || authRepo.loginAttempts[0].Reason != "account_locked" {
		t.Fatalf("expected account_locked login attempt, got %#v", authRepo.loginAttempts)
	}
	if len(authRepo.auditLogs) != 1 || authRepo.auditLogs[0].EventType != "login_failed" {
		t.Fatalf("expected login_failed audit log, got %#v", authRepo.auditLogs)
	}
}

func TestServiceLoginAppliesFailureWindowLockoutByIdentifierAndIP(t *testing.T) {
	ctx := context.Background()
	ctx = WithClientMetadata(ctx, "203.0.113.10", "Mozilla/5.0")
	hasher := security.NewArgon2idHasher(19456, 2, 1, "pepper")
	tokenManager := security.NewHMACTokenManager(security.HMACTokenConfig{Issuer: "backend", Audience: "admin-api", KeyID: "test-key", Secret: []byte("01234567890123456789012345678901"), AccessTokenTTL: 15 * time.Minute, ClockSkew: 30 * time.Second})
	now := time.Now().UTC()
	identifierHash := sha256Hex("missing@example.com")
	ipHash := sha256Hex("203.0.113.10")

	userRepo := &stubUserReader{err: ErrUserNotFound}
	authRepo := &stubAuthRepository{
		recentFailedAttemptTimesByIdentifier: map[string][]time.Time{
			identifierHash: {now.Add(-2 * time.Minute), now.Add(-4 * time.Minute), now.Add(-6 * time.Minute), now.Add(-8 * time.Minute), now.Add(-10 * time.Minute)},
		},
		recentFailedAttemptTimesByIP: map[string][]time.Time{
			ipHash: {now.Add(-3 * time.Minute), now.Add(-5 * time.Minute), now.Add(-7 * time.Minute), now.Add(-9 * time.Minute), now.Add(-11 * time.Minute)},
		},
	}
	svc := NewService(userRepo, authRepo, database.NoopTxManager{}, hasher, tokenManager, security.CryptoRandomTokenGenerator{}, ServiceConfig{
		RefreshTokenTTL:   24 * time.Hour,
		RefreshCookieName: "refresh_token",
		MaxFailedAttempts: 5,
		FailedWindow:      15 * time.Minute,
		LockDuration:      30 * time.Minute,
	})

	_, _, err := svc.Login(ctx, &LoginRequest{Identifier: "missing@example.com", Password: "secret-password"})
	if !errors.Is(err, ErrAccountLocked) {
		t.Fatalf("expected ErrAccountLocked, got %v", err)
	}
	if len(authRepo.loginAttempts) != 1 || authRepo.loginAttempts[0].Reason != "too_many_attempts" {
		t.Fatalf("expected too_many_attempts login attempt, got %#v", authRepo.loginAttempts)
	}
	if len(authRepo.auditLogs) != 1 || authRepo.auditLogs[0].EventType != "login_failed" {
		t.Fatalf("expected login_failed audit log, got %#v", authRepo.auditLogs)
	}
	if got := authRepo.lastRecentFailedAttemptsInput.IdentifierHash; got == "" {
		t.Fatal("expected CountRecentFailedAttempts to receive identifier hash")
	}
	if got := authRepo.lastRecentFailedAttemptsInput.IPHash; got == "" {
		t.Fatal("expected CountRecentFailedAttempts to receive ip hash")
	}
}

func TestServiceLoginAppliesFailureWindowLockoutByUserID(t *testing.T) {
	ctx := WithClientMetadata(context.Background(), "203.0.113.11", "Mozilla/5.0")
	hasher := security.NewArgon2idHasher(19456, 2, 1, "pepper")
	hash, err := hasher.Hash("secret-password")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}
	tokenManager := security.NewHMACTokenManager(security.HMACTokenConfig{Issuer: "backend", Audience: "admin-api", KeyID: "test-key", Secret: []byte("01234567890123456789012345678901"), AccessTokenTTL: 15 * time.Minute, ClockSkew: 30 * time.Second})
	now := time.Now().UTC()

	userRepo := &stubUserReader{user: &user.User{ID: 9, PublicID: "usr_9", Email: "admin@example.com", PasswordHash: hash, Status: user.StatusActive}}
	authRepo := &stubAuthRepository{
		recentFailedAttemptTimesByUserID: map[int64][]time.Time{
			9: {now.Add(-time.Minute), now.Add(-2 * time.Minute), now.Add(-3 * time.Minute), now.Add(-4 * time.Minute), now.Add(-5 * time.Minute)},
		},
	}
	svc := NewService(userRepo, authRepo, database.NoopTxManager{}, hasher, tokenManager, security.CryptoRandomTokenGenerator{}, ServiceConfig{
		RefreshTokenTTL:   24 * time.Hour,
		RefreshCookieName: "refresh_token",
		MaxFailedAttempts: 5,
		FailedWindow:      15 * time.Minute,
		LockDuration:      30 * time.Minute,
	})

	_, _, err = svc.Login(ctx, &LoginRequest{Identifier: "admin@example.com", Password: "secret-password"})
	if !errors.Is(err, ErrAccountLocked) {
		t.Fatalf("expected ErrAccountLocked, got %v", err)
	}
	if authRepo.lastRecentFailedAttemptsInput.UserID == nil || *authRepo.lastRecentFailedAttemptsInput.UserID != 9 {
		t.Fatalf("expected user id lockout filter, got %#v", authRepo.lastRecentFailedAttemptsInput)
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

func TestServiceRefreshSuccessRotatesTokenAndWritesAudit(t *testing.T) {
	ctx := context.Background()
	authRepo := &stubAuthRepository{
		markUsedOK: true,
		refreshTokenByHash: map[string]*RefreshToken{
			sha256Hex("refresh-ok"): {
				ID:        40,
				UserID:    7,
				SessionID: "ses_ok",
				FamilyID:  "fam_ok",
				TokenHash: sha256Hex("refresh-ok"),
				ExpiresAt: time.Now().UTC().Add(time.Hour),
				CreatedAt: time.Now().UTC().Add(-time.Hour),
			},
		},
		sessionsByID: map[string]*AuthSession{
			"ses_ok": {
				ID:         "ses_ok",
				UserID:     7,
				Status:     "active",
				CreatedAt:  time.Now().UTC().Add(-time.Hour),
				LastSeenAt: ptrTime(time.Now().UTC().Add(-10 * time.Minute)),
				ExpiresAt:  time.Now().UTC().Add(time.Hour),
			},
		},
		roles:       []string{"admin"},
		permissions: []string{"order:read"},
	}
	userRepo := &stubUserReader{userByID: map[int64]*user.User{
		7: {ID: 7, PublicID: "usr_7", Email: "admin@example.com", Status: user.StatusActive},
	}}
	tokenManager := security.NewHMACTokenManager(security.HMACTokenConfig{Issuer: "backend", Audience: "admin-api", KeyID: "test-key", Secret: []byte("01234567890123456789012345678901"), AccessTokenTTL: 15 * time.Minute, ClockSkew: 30 * time.Second})

	svc := NewService(userRepo, authRepo, database.NoopTxManager{}, nil, tokenManager, stubRandomTokenGenerator{values: []string{"refresh-rotated"}}, ServiceConfig{RefreshTokenTTL: 24 * time.Hour, RefreshCookieName: "refresh_token"})

	resp, cookie, err := svc.Refresh(ctx, &RefreshRequest{RefreshToken: "refresh-ok"})
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}
	if resp == nil || resp.AccessToken == "" {
		t.Fatalf("expected access token, got %+v", resp)
	}
	if cookie == nil || cookie.Value != "refresh-rotated" {
		t.Fatalf("expected rotated refresh cookie, got %+v", cookie)
	}
	if len(authRepo.refreshTokens) != 1 {
		t.Fatalf("expected one new refresh token, got %#v", authRepo.refreshTokens)
	}
	if len(authRepo.usedTokenPairs) != 1 || authRepo.usedTokenPairs[0][0] != 40 {
		t.Fatalf("expected token usage to be recorded, got %#v", authRepo.usedTokenPairs)
	}
	if len(authRepo.auditLogs) != 1 || authRepo.auditLogs[0].EventType != "refresh_success" {
		t.Fatalf("expected refresh_success audit log, got %#v", authRepo.auditLogs)
	}
}

func TestServiceRefreshExpiredTokenReturnsInvalidAndClearsCookie(t *testing.T) {
	authRepo := &stubAuthRepository{
		refreshTokenByHash: map[string]*RefreshToken{
			sha256Hex("refresh-expired"): {
				ID:        41,
				UserID:    7,
				SessionID: "ses_expired",
				FamilyID:  "fam_expired",
				TokenHash: sha256Hex("refresh-expired"),
				ExpiresAt: time.Now().UTC().Add(-time.Minute),
				CreatedAt: time.Now().UTC().Add(-time.Hour),
			},
		},
	}

	svc := NewService(&stubUserReader{}, authRepo, database.NoopTxManager{}, nil, nil, nil, ServiceConfig{RefreshTokenTTL: 24 * time.Hour, RefreshCookieName: "refresh_token"})

	resp, cookie, err := svc.Refresh(context.Background(), &RefreshRequest{RefreshToken: "refresh-expired"})
	if !errors.Is(err, ErrRefreshInvalid) {
		t.Fatalf("expected ErrRefreshInvalid, got %v", err)
	}
	if resp != nil {
		t.Fatalf("expected nil response, got %+v", resp)
	}
	if cookie == nil || !cookie.Clear {
		t.Fatalf("expected clear cookie, got %+v", cookie)
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

func TestServiceLoginSuccessStoresClientHashes(t *testing.T) {
	ctx := WithClientMetadata(context.Background(), "203.0.113.10", "Mozilla/5.0")
	hasher := security.NewArgon2idHasher(19456, 2, 1, "pepper")
	hash, err := hasher.Hash("secret-password")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	userRepo := &stubUserReader{user: &user.User{ID: 1, PublicID: "usr_123", Email: "admin@example.com", PasswordHash: hash, Status: user.StatusActive}}
	authRepo := &stubAuthRepository{roles: []string{"admin"}, permissions: []string{"order:read"}}
	tokenManager := security.NewHMACTokenManager(security.HMACTokenConfig{Issuer: "backend", Audience: "admin-api", KeyID: "test-key", Secret: []byte("01234567890123456789012345678901"), AccessTokenTTL: 15 * time.Minute, ClockSkew: 30 * time.Second})

	svc := NewService(userRepo, authRepo, database.NoopTxManager{}, hasher, tokenManager, security.CryptoRandomTokenGenerator{}, ServiceConfig{RefreshTokenTTL: 24 * time.Hour, RefreshCookieName: "refresh_token"})
	_, _, err = svc.Login(ctx, &LoginRequest{Identifier: "admin@example.com", Password: "secret-password"})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if len(authRepo.sessions) != 1 {
		t.Fatalf("expected one session, got %#v", authRepo.sessions)
	}
	wantIPHash := sha256Hex("203.0.113.10")
	wantUAHash := sha256Hex("Mozilla/5.0")
	if authRepo.sessions[0].IPHash != wantIPHash || authRepo.sessions[0].UserAgentHash != wantUAHash {
		t.Fatalf("expected session hashes ip=%q ua=%q, got %+v", wantIPHash, wantUAHash, authRepo.sessions[0])
	}
	if len(authRepo.loginAttempts) != 1 || authRepo.loginAttempts[0].IPHash != wantIPHash || authRepo.loginAttempts[0].UserAgentHash != wantUAHash {
		t.Fatalf("expected login attempt hashes, got %#v", authRepo.loginAttempts)
	}
	if len(authRepo.auditLogs) != 1 || authRepo.auditLogs[0].IPHash != wantIPHash || authRepo.auditLogs[0].UserAgentHash != wantUAHash {
		t.Fatalf("expected audit log hashes, got %#v", authRepo.auditLogs)
	}
}

func TestProductionCodeDoesNotLogSensitiveAuthFields(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	for _, path := range []string{
		filepath.Join(root, "cmd", "api", "main.go"),
		filepath.Join(root, "cmd", "worker", "main.go"),
		filepath.Join(root, "internal", "modules", "auth", "handler.go"),
		filepath.Join(root, "internal", "modules", "auth", "service.go"),
	} {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile(%s) error = %v", path, err)
		}
		text := strings.ToLower(string(content))
		for _, forbidden := range []string{"authorization", "cookie", "password", "access_token", "refresh_token"} {
			if strings.Contains(text, fmt.Sprintf("\"%s\"", forbidden)) {
				t.Fatalf("expected no explicit logging of sensitive field %q in %s", forbidden, path)
			}
		}
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
	roles                                []string
	permissions                          []string
	loginAttempts                        []*LoginAttempt
	auditLogs                            []*AuditLog
	sessions                             []*AuthSession
	refreshTokens                        []*RefreshToken
	refreshTokenByHash                   map[string]*RefreshToken
	sessionsByID                         map[string]*AuthSession
	revokedFamilies                      []string
	revokedSessions                      []string
	revokedSessionTokens                 []string
	recentFailedAttemptTimesByIdentifier map[string][]time.Time
	recentFailedAttemptTimesByUserID     map[int64][]time.Time
	recentFailedAttemptTimesByIP         map[string][]time.Time
	lastRecentFailedAttemptsInput        FailedLoginAttemptFilter
	markUsedOK                           bool
	usedTokenPairs                       [][2]int64
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
func (s *stubAuthRepository) ListRecentFailedAttemptTimes(ctx context.Context, filter FailedLoginAttemptFilter) ([]time.Time, error) {
	s.lastRecentFailedAttemptsInput = filter
	if filter.UserID != nil {
		return append([]time.Time(nil), s.recentFailedAttemptTimesByUserID[*filter.UserID]...), nil
	}
	if filter.IPHash != "" {
		return append([]time.Time(nil), s.recentFailedAttemptTimesByIP[filter.IPHash]...), nil
	}
	return append([]time.Time(nil), s.recentFailedAttemptTimesByIdentifier[filter.IdentifierHash]...), nil
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

type stubRandomTokenGenerator struct {
	values []string
	index  int
}

func (s stubRandomTokenGenerator) GenerateURLSafe(n int) (string, error) {
	if s.index >= len(s.values) {
		return "", fmt.Errorf("unexpected GenerateURLSafe(%d)", n)
	}
	value := s.values[s.index]
	s.index++
	return value, nil
}

func TestAuthErrorCodesExposeExpectedValues(t *testing.T) {
	if got := apperrors.Code(ErrRefreshInvalid); got != "AUTH_REFRESH_INVALID" {
		t.Fatalf("expected AUTH_REFRESH_INVALID, got %q", got)
	}
	if got := apperrors.Code(ErrRefreshReused); got != "AUTH_REFRESH_REUSED" {
		t.Fatalf("expected AUTH_REFRESH_REUSED, got %q", got)
	}
}
