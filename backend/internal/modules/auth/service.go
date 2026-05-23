package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"backend/internal/modules/user"
	"backend/internal/platform/database"
	"backend/internal/platform/observability"
	"backend/internal/platform/security"
)

type UserReader interface {
	GetByEmail(ctx context.Context, email string) (*user.User, error)
	GetByID(ctx context.Context, userID int64) (*user.User, error)
}

type Service interface {
	Login(ctx context.Context, req *LoginRequest) (*AuthResponse, *RefreshCookie, error)
	Refresh(ctx context.Context, req *RefreshRequest) (*AuthResponse, *RefreshCookie, error)
	Logout(ctx context.Context, req *LogoutRequest) error
	Me(ctx context.Context) (*MeResponse, error)
}

type service struct {
	userRepo        UserReader
	repo            Repository
	txManager       database.TxManager
	passwordHasher  security.PasswordHasher
	tokenManager    security.TokenManager
	randomGenerator security.RandomTokenGenerator
	config          ServiceConfig
	dummyHash       string
}

func NewService(userRepo UserReader, repo Repository, txManager database.TxManager, passwordHasher security.PasswordHasher, tokenManager security.TokenManager, randomGenerator security.RandomTokenGenerator, cfg ServiceConfig) Service {
	if txManager == nil {
		txManager = database.NoopTxManager{}
	}
	if randomGenerator == nil {
		randomGenerator = security.CryptoRandomTokenGenerator{}
	}
	var dummyHash string
	if passwordHasher != nil {
		dummyHash, _ = passwordHasher.Hash("dummy-password-for-timing")
	}
	return &service{userRepo: userRepo, repo: repo, txManager: txManager, passwordHasher: passwordHasher, tokenManager: tokenManager, randomGenerator: randomGenerator, config: cfg, dummyHash: dummyHash}
}

func (s *service) Login(ctx context.Context, req *LoginRequest) (*AuthResponse, *RefreshCookie, error) {
	if req == nil {
		return nil, nil, ErrInvalidCredentials
	}
	email := user.NormalizeEmail(req.Identifier)
	u, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) && s.passwordHasher != nil && s.dummyHash != "" {
			_, _ = s.passwordHasher.Verify(req.Password, s.dummyHash)
			s.writeLoginAttempt(ctx, email, false, "invalid_credentials")
			s.writeAuditLog(ctx, &AuditLog{EventType: "login_failed", Result: "failure", MetadataJSON: map[string]any{"reason": "invalid_credentials"}, OccurredAt: time.Now().UTC()})
			return nil, nil, ErrInvalidCredentials
		}
		return nil, nil, err
	}
	if u == nil {
		if s.passwordHasher != nil && s.dummyHash != "" {
			_, _ = s.passwordHasher.Verify(req.Password, s.dummyHash)
		}
		s.writeLoginAttempt(ctx, email, false, "invalid_credentials")
		s.writeAuditLog(ctx, &AuditLog{EventType: "login_failed", Result: "failure", MetadataJSON: map[string]any{"reason": "invalid_credentials"}, OccurredAt: time.Now().UTC()})
		return nil, nil, ErrInvalidCredentials
	}
	ok, err := s.passwordHasher.Verify(req.Password, u.PasswordHash)
	if err != nil {
		return nil, nil, err
	}
	if !ok {
		s.writeLoginAttempt(ctx, email, false, "invalid_credentials")
		s.writeAuditLog(ctx, &AuditLog{EventType: "login_failed", Result: "failure", UserPublicID: u.PublicID, MetadataJSON: map[string]any{"reason": "invalid_credentials"}, OccurredAt: time.Now().UTC()})
		return nil, nil, ErrInvalidCredentials
	}
	switch u.Status {
	case user.StatusDisabled:
		s.writeLoginAttempt(ctx, email, false, "account_disabled")
		s.writeAuditLog(ctx, &AuditLog{EventType: "login_failed", Result: "failure", UserPublicID: u.PublicID, MetadataJSON: map[string]any{"reason": "account_disabled"}, OccurredAt: time.Now().UTC()})
		return nil, nil, ErrAccountDisabled
	case user.StatusLocked:
		s.writeLoginAttempt(ctx, email, false, "account_locked")
		s.writeAuditLog(ctx, &AuditLog{EventType: "login_failed", Result: "failure", UserPublicID: u.PublicID, MetadataJSON: map[string]any{"reason": "account_locked"}, OccurredAt: time.Now().UTC()})
		return nil, nil, ErrAccountLocked
	}
	roles, err := s.repo.ListUserRoles(ctx, u.ID)
	if err != nil {
		return nil, nil, err
	}
	permissions, err := s.repo.ListUserPermissions(ctx, u.ID)
	if err != nil {
		return nil, nil, err
	}
	var response *AuthResponse
	var cookie *RefreshCookie
	err = s.txManager.WithinTx(ctx, func(txCtx context.Context) error {
		now := time.Now().UTC()
		sessionID, err := s.randomGenerator.GenerateURLSafe(24)
		if err != nil {
			return err
		}
		familyID, err := s.randomGenerator.GenerateURLSafe(24)
		if err != nil {
			return err
		}
		refreshValue, err := s.randomGenerator.GenerateURLSafe(32)
		if err != nil {
			return err
		}
		refreshHash := sha256Hex(refreshValue)
		expiresAt := now.Add(s.config.RefreshTokenTTL)
		principal := &security.Principal{PublicID: u.PublicID, SessionID: sessionID, Roles: security.NormalizeStrings(roles), Permissions: security.NormalizeStrings(permissions)}
		if err := s.repo.CreateSession(txCtx, &AuthSession{ID: sessionID, UserID: u.ID, Status: "active", CreatedAt: now, LastSeenAt: ptrTime(now), ExpiresAt: expiresAt}); err != nil {
			return err
		}
		if err := s.repo.CreateRefreshToken(txCtx, &RefreshToken{UserID: u.ID, SessionID: sessionID, TokenHash: refreshHash, FamilyID: familyID, ExpiresAt: expiresAt, CreatedAt: now}); err != nil {
			return err
		}
		accessToken, err := s.tokenManager.IssueAccessToken(txCtx, principal)
		if err != nil {
			return err
		}
		if err := s.repo.UpdateLastLoginAt(txCtx, u.ID, now); err != nil {
			return err
		}
		s.writeLoginAttempt(txCtx, email, true, "success")
		s.writeAuditLog(txCtx, &AuditLog{EventType: "login_success", Result: "success", UserPublicID: u.PublicID, SessionID: sessionID, OccurredAt: now})
		response = &AuthResponse{AccessToken: accessToken.Token, TokenType: accessToken.TokenType, ExpiresIn: accessToken.ExpiresIn, User: &AuthUser{ID: u.PublicID, Roles: principal.Roles, Permissions: principal.Permissions}}
		cookie = &RefreshCookie{Value: refreshValue, ExpiresAt: expiresAt}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return response, cookie, nil
}

func (s *service) Refresh(ctx context.Context, req *RefreshRequest) (*AuthResponse, *RefreshCookie, error) {
	if req == nil || strings.TrimSpace(req.RefreshToken) == "" {
		return nil, s.clearRefreshCookie(), ErrRefreshInvalid
	}
	tokenHash := sha256Hex(req.RefreshToken)
	var response *AuthResponse
	var cookie *RefreshCookie
	err := s.txManager.WithinTx(ctx, func(txCtx context.Context) error {
		stored, err := s.repo.GetRefreshTokenByHashForUpdate(txCtx, tokenHash)
		if err != nil {
			return err
		}
		if stored == nil {
			return ErrRefreshInvalid
		}
		now := time.Now().UTC()
		if stored.RevokedAt != nil || now.After(stored.ExpiresAt) {
			return ErrRefreshInvalid
		}
		if stored.UsedAt != nil {
			_ = s.repo.RevokeRefreshTokenFamily(txCtx, stored.FamilyID)
			_ = s.repo.RevokeSession(txCtx, stored.SessionID)
			s.writeAuditLog(txCtx, &AuditLog{EventType: "refresh_reuse_detected", Result: "failure", SessionID: stored.SessionID, OccurredAt: now})
			return ErrRefreshReused
		}
		session, err := s.repo.GetSessionByID(txCtx, stored.SessionID)
		if err != nil {
			return err
		}
		if session == nil || session.RevokedAt != nil || strings.TrimSpace(session.Status) != "active" || now.After(session.ExpiresAt) {
			return ErrRefreshInvalid
		}
		u, err := s.userRepo.GetByID(txCtx, stored.UserID)
		if err != nil {
			return err
		}
		if u == nil || u.Status != user.StatusActive {
			return ErrRefreshInvalid
		}
		if passwordChangedAfterSession(u, session, stored) {
			_ = s.repo.RevokeRefreshTokenFamily(txCtx, stored.FamilyID)
			_ = s.repo.RevokeSession(txCtx, stored.SessionID)
			s.writeAuditLog(txCtx, &AuditLog{EventType: "session_revoked", Result: "failure", UserPublicID: u.PublicID, SessionID: stored.SessionID, MetadataJSON: map[string]any{"reason": "password_changed"}, OccurredAt: now})
			return ErrRefreshInvalid
		}
		roles, err := s.repo.ListUserRoles(txCtx, u.ID)
		if err != nil {
			return err
		}
		permissions, err := s.repo.ListUserPermissions(txCtx, u.ID)
		if err != nil {
			return err
		}
		refreshValue, err := s.randomGenerator.GenerateURLSafe(32)
		if err != nil {
			return err
		}
		newTokenHash := sha256Hex(refreshValue)
		expiresAt := now.Add(s.config.RefreshTokenTTL)
		newToken := &RefreshToken{UserID: u.ID, SessionID: stored.SessionID, TokenHash: newTokenHash, FamilyID: stored.FamilyID, ExpiresAt: expiresAt, CreatedAt: now}
		if err := s.repo.CreateRefreshToken(txCtx, newToken); err != nil {
			return err
		}
		used, err := s.repo.MarkRefreshTokenUsed(txCtx, stored.ID, newToken.ID)
		if err != nil {
			return err
		}
		if !used {
			return ErrRefreshReused
		}
		principal := &security.Principal{PublicID: u.PublicID, SessionID: stored.SessionID, Roles: security.NormalizeStrings(roles), Permissions: security.NormalizeStrings(permissions)}
		accessToken, err := s.tokenManager.IssueAccessToken(txCtx, principal)
		if err != nil {
			return err
		}
		s.writeAuditLog(txCtx, &AuditLog{EventType: "refresh_success", Result: "success", UserPublicID: u.PublicID, SessionID: stored.SessionID, OccurredAt: now})
		response = &AuthResponse{AccessToken: accessToken.Token, TokenType: accessToken.TokenType, ExpiresIn: accessToken.ExpiresIn, User: &AuthUser{ID: u.PublicID, Roles: principal.Roles, Permissions: principal.Permissions}}
		cookie = &RefreshCookie{Value: refreshValue, ExpiresAt: expiresAt}
		return nil
	})
	if err != nil {
		if errors.Is(err, ErrRefreshInvalid) || errors.Is(err, ErrRefreshReused) {
			return nil, s.clearRefreshCookie(), err
		}
		return nil, nil, err
	}
	return response, cookie, nil
}

func (s *service) Logout(ctx context.Context, req *LogoutRequest) error {
	if req == nil {
		req = &LogoutRequest{}
	}
	principal, _ := security.PrincipalFromContext(ctx)
	var cookieSessionID string
	var familyID string
	if strings.TrimSpace(req.RefreshToken) != "" {
		stored, err := s.repo.GetRefreshTokenByHashForUpdate(ctx, sha256Hex(req.RefreshToken))
		if err != nil {
			return err
		}
		if stored != nil {
			cookieSessionID = stored.SessionID
			familyID = stored.FamilyID
		}
	}
	revokedAny := false
	revokedSessions := make(map[string]struct{})
	metadata := map[string]any{}
	if cookieSessionID != "" {
		if familyID != "" {
			if err := s.repo.RevokeRefreshTokenFamily(ctx, familyID); err != nil {
				return err
			}
		}
		if err := s.revokeSessionIfNeeded(ctx, revokedSessions, cookieSessionID); err != nil {
			return err
		}
		revokedAny = true
	}
	if principal != nil && strings.TrimSpace(principal.SessionID) != "" {
		if cookieSessionID != "" && cookieSessionID != principal.SessionID {
			metadata["session_mismatch"] = true
		}
		if err := s.revokeSessionIfNeeded(ctx, revokedSessions, principal.SessionID); err != nil {
			return err
		}
		revokedAny = true
	}
	now := time.Now().UTC()
	if revokedAny {
		s.writeAuditLog(ctx, &AuditLog{EventType: "session_revoked", Result: "success", SessionID: preferredSessionID(cookieSessionID, principal), MetadataJSON: cloneMetadata(metadata), OccurredAt: now})
	}
	s.writeAuditLog(ctx, &AuditLog{EventType: "logout_success", Result: "success", SessionID: preferredSessionID(cookieSessionID, principal), MetadataJSON: metadata, OccurredAt: now})
	return nil
}

func (s *service) Me(ctx context.Context) (*MeResponse, error) {
	p, ok := security.PrincipalFromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}
	return &MeResponse{User: AuthUser{ID: p.PublicID, Roles: append([]string(nil), p.Roles...), Permissions: append([]string(nil), p.Permissions...)}, SessionID: p.SessionID}, nil
}

func sha256Hex(value string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(value)))
	return hex.EncodeToString(sum[:])
}

func (s *service) clearRefreshCookie() *RefreshCookie {
	return &RefreshCookie{Clear: true}
}

func (s *service) writeLoginAttempt(ctx context.Context, identifier string, success bool, reason string) {
	if s.repo == nil {
		return
	}
	now := time.Now().UTC()
	requestID, _ := observability.RequestIDFromContext(ctx)
	traceID, _ := observability.TraceIDFromContext(ctx)
	_ = s.repo.CreateLoginAttempt(ctx, &LoginAttempt{IdentifierHash: sha256Hex(identifier), Success: success, Reason: reason, RequestID: requestID, TraceID: traceID, CreatedAt: now})
}

func (s *service) writeAuditLog(ctx context.Context, log *AuditLog) {
	if s.repo == nil || log == nil {
		return
	}
	if log.MetadataJSON == nil {
		log.MetadataJSON = map[string]any{}
	}
	if log.OccurredAt.IsZero() {
		log.OccurredAt = time.Now().UTC()
	}
	if requestID, ok := observability.RequestIDFromContext(ctx); ok {
		log.RequestID = requestID
	}
	if traceID, ok := observability.TraceIDFromContext(ctx); ok {
		log.TraceID = traceID
	}
	_ = s.repo.CreateAuditLog(ctx, log)
}

func passwordChangedAfterSession(u *user.User, session *AuthSession, token *RefreshToken) bool {
	if u == nil || u.PasswordChangedAt == nil {
		return false
	}
	changedAt := u.PasswordChangedAt.UTC()
	if session != nil && session.CreatedAt.Before(changedAt) {
		return true
	}
	if token != nil && token.CreatedAt.Before(changedAt) {
		return true
	}
	return false
}

func (s *service) revokeSessionIfNeeded(ctx context.Context, seen map[string]struct{}, sessionID string) error {
	if strings.TrimSpace(sessionID) == "" {
		return nil
	}
	if _, ok := seen[sessionID]; ok {
		return nil
	}
	seen[sessionID] = struct{}{}
	return s.repo.RevokeSession(ctx, sessionID)
}

func preferredSessionID(cookieSessionID string, principal *security.Principal) string {
	if strings.TrimSpace(cookieSessionID) != "" {
		return cookieSessionID
	}
	if principal == nil {
		return ""
	}
	return principal.SessionID
}

func cloneMetadata(metadata map[string]any) map[string]any {
	if len(metadata) == 0 {
		return nil
	}
	cloned := make(map[string]any, len(metadata))
	for k, v := range metadata {
		cloned[k] = v
	}
	return cloned
}

func ptrTime(value time.Time) *time.Time {
	return &value
}
