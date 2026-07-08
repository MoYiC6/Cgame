package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
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
	GetByIdentifier(ctx context.Context, identifier string) (*user.User, error)
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
	identifier := strings.TrimSpace(req.Username)
	clientIPHash, userAgentHash := requestHashes(ctx)
	if s.isLockedOut(ctx, FailedLoginAttemptFilter{IdentifierHash: sha256Hex(identifier), IPHash: clientIPHash, Since: time.Now().UTC().Add(-s.config.FailedWindow)}) {
		s.writeLoginAttempt(ctx, loginAttemptTarget{IdentifierHash: sha256Hex(identifier), IPHash: clientIPHash, UserAgentHash: userAgentHash}, false, "too_many_attempts")
		s.writeAuditLog(ctx, &AuditLog{EventType: "login_failed", Result: "failure", IPHash: clientIPHash, UserAgentHash: userAgentHash, MetadataJSON: map[string]any{"reason": "too_many_attempts"}, OccurredAt: time.Now().UTC()})
		return nil, nil, ErrAccountLocked
	}
	u, err := s.userRepo.GetByIdentifier(ctx, req.Username)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) && s.passwordHasher != nil && s.dummyHash != "" {
			_, _ = s.passwordHasher.Verify(req.Password, s.dummyHash)
			s.writeLoginAttempt(ctx, loginAttemptTarget{IdentifierHash: sha256Hex(identifier), IPHash: clientIPHash, UserAgentHash: userAgentHash}, false, "invalid_credentials")
			s.writeAuditLog(ctx, &AuditLog{EventType: "login_failed", Result: "failure", IPHash: clientIPHash, UserAgentHash: userAgentHash, MetadataJSON: map[string]any{"reason": "invalid_credentials"}, OccurredAt: time.Now().UTC()})
			return nil, nil, ErrInvalidCredentials
		}
		return nil, nil, err
	}
	if u == nil {
		if s.passwordHasher != nil && s.dummyHash != "" {
			_, _ = s.passwordHasher.Verify(req.Password, s.dummyHash)
		}
		s.writeLoginAttempt(ctx, loginAttemptTarget{IdentifierHash: sha256Hex(identifier), IPHash: clientIPHash, UserAgentHash: userAgentHash}, false, "invalid_credentials")
		s.writeAuditLog(ctx, &AuditLog{EventType: "login_failed", Result: "failure", IPHash: clientIPHash, UserAgentHash: userAgentHash, MetadataJSON: map[string]any{"reason": "invalid_credentials"}, OccurredAt: time.Now().UTC()})
		return nil, nil, ErrInvalidCredentials
	}
	if s.isLockedOut(ctx, FailedLoginAttemptFilter{UserID: ptrInt64(u.ID), IdentifierHash: sha256Hex(identifier), IPHash: clientIPHash, Since: time.Now().UTC().Add(-s.config.FailedWindow)}) {
		s.writeLoginAttempt(ctx, loginAttemptTarget{IdentifierHash: sha256Hex(identifier), UserID: ptrInt64(u.ID), IPHash: clientIPHash, UserAgentHash: userAgentHash}, false, "too_many_attempts")
		s.writeAuditLog(ctx, &AuditLog{EventType: "login_failed", Result: "failure", UserPublicID: u.PublicID, IPHash: clientIPHash, UserAgentHash: userAgentHash, MetadataJSON: map[string]any{"reason": "too_many_attempts"}, OccurredAt: time.Now().UTC()})
		return nil, nil, ErrAccountLocked
	}
	ok, err := s.passwordHasher.Verify(req.Password, u.PasswordHash)
	if errors.Is(err, security.ErrNoMatchingHasher) {
		return nil, nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, nil, err
	}
	if !ok {
		s.writeLoginAttempt(ctx, loginAttemptTarget{IdentifierHash: sha256Hex(identifier), UserID: ptrInt64(u.ID), IPHash: clientIPHash, UserAgentHash: userAgentHash}, false, "invalid_credentials")
		s.writeAuditLog(ctx, &AuditLog{EventType: "login_failed", Result: "failure", UserPublicID: u.PublicID, IPHash: clientIPHash, UserAgentHash: userAgentHash, MetadataJSON: map[string]any{"reason": "invalid_credentials"}, OccurredAt: time.Now().UTC()})
		return nil, nil, ErrInvalidCredentials
	}
	switch u.Status {
	case user.StatusDisabled:
		s.writeLoginAttempt(ctx, loginAttemptTarget{IdentifierHash: sha256Hex(identifier), UserID: ptrInt64(u.ID), IPHash: clientIPHash, UserAgentHash: userAgentHash}, false, "account_disabled")
		s.writeAuditLog(ctx, &AuditLog{EventType: "login_failed", Result: "failure", UserPublicID: u.PublicID, IPHash: clientIPHash, UserAgentHash: userAgentHash, MetadataJSON: map[string]any{"reason": "account_disabled"}, OccurredAt: time.Now().UTC()})
		return nil, nil, ErrAccountDisabled
	case user.StatusLocked:
		s.writeLoginAttempt(ctx, loginAttemptTarget{IdentifierHash: sha256Hex(identifier), UserID: ptrInt64(u.ID), IPHash: clientIPHash, UserAgentHash: userAgentHash}, false, "account_locked")
		s.writeAuditLog(ctx, &AuditLog{EventType: "login_failed", Result: "failure", UserPublicID: u.PublicID, IPHash: clientIPHash, UserAgentHash: userAgentHash, MetadataJSON: map[string]any{"reason": "account_locked"}, OccurredAt: time.Now().UTC()})
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
		principal := &security.Principal{UserID: fmt.Sprintf("%d", u.ID), PublicID: u.PublicID, SessionID: sessionID, Roles: security.NormalizeStrings(roles), Permissions: security.NormalizeStrings(permissions), Status: u.Status}
		if err := s.repo.CreateSession(txCtx, &AuthSession{ID: sessionID, UserID: u.ID, Status: "active", UserAgentHash: userAgentHash, IPHash: clientIPHash, CreatedAt: now, LastSeenAt: ptrTime(now), ExpiresAt: expiresAt}); err != nil {
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
		s.writeLoginAttempt(txCtx, loginAttemptTarget{IdentifierHash: sha256Hex(identifier), UserID: ptrInt64(u.ID), IPHash: clientIPHash, UserAgentHash: userAgentHash}, true, "success")
		s.writeAuditLog(txCtx, &AuditLog{EventType: "login_success", Result: "success", UserPublicID: u.PublicID, SessionID: sessionID, IPHash: clientIPHash, UserAgentHash: userAgentHash, OccurredAt: now})
		rolesStr := joinRoles(roles)
		response = &AuthResponse{
			AccessToken:     accessToken.Token,
			RefreshToken:    refreshValue,
			TokenType:       accessToken.TokenType,
			ExpiresIn:       accessToken.ExpiresIn,
			RefreshExpiresIn: int64(s.config.RefreshTokenTTL.Seconds()),
			UserID:          u.ID,
			Username:        u.Username,
			Nickname:        u.Nickname,
			Avatar:          u.Avatar,
			Roles:           rolesStr,
			Permissions:     security.NormalizeStrings(permissions),
		}
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
	var compromiseEvent *refreshCompromiseEvent
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
			compromiseEvent = &refreshCompromiseEvent{
				familyID:      stored.FamilyID,
				sessionID:     stored.SessionID,
				eventType:     "refresh_reuse_detected",
				result:        "failure",
				occurredAt:    now,
				ipHash:        ClientIPHashFromContext(txCtx),
				userAgentHash: UserAgentHashFromContext(txCtx),
			}
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
			compromiseEvent = &refreshCompromiseEvent{
				familyID:      stored.FamilyID,
				sessionID:     stored.SessionID,
				userPublicID:  u.PublicID,
				eventType:     "session_revoked",
				result:        "failure",
				occurredAt:    now,
				ipHash:        ClientIPHashFromContext(txCtx),
				userAgentHash: UserAgentHashFromContext(txCtx),
				metadata:      map[string]any{"reason": "password_changed"},
			}
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
			compromiseEvent = &refreshCompromiseEvent{
				familyID:      stored.FamilyID,
				sessionID:     stored.SessionID,
				eventType:     "refresh_reuse_detected",
				result:        "failure",
				occurredAt:    now,
				ipHash:        ClientIPHashFromContext(txCtx),
				userAgentHash: UserAgentHashFromContext(txCtx),
			}
			return ErrRefreshReused
		}
		principal := &security.Principal{UserID: fmt.Sprintf("%d", u.ID), PublicID: u.PublicID, SessionID: stored.SessionID, Roles: security.NormalizeStrings(roles), Permissions: security.NormalizeStrings(permissions), Status: u.Status}
		accessToken, err := s.tokenManager.IssueAccessToken(txCtx, principal)
		if err != nil {
			return err
		}
	s.writeAuditLog(txCtx, &AuditLog{EventType: "refresh_success", Result: "success", UserPublicID: u.PublicID, SessionID: stored.SessionID, IPHash: ClientIPHashFromContext(txCtx), UserAgentHash: UserAgentHashFromContext(txCtx), OccurredAt: now})
	rolesStr := joinRoles(roles)
	response = &AuthResponse{
		AccessToken:      accessToken.Token,
		RefreshToken:     refreshValue,
		TokenType:        accessToken.TokenType,
		ExpiresIn:        accessToken.ExpiresIn,
		RefreshExpiresIn: int64(s.config.RefreshTokenTTL.Seconds()),
		UserID:           u.ID,
		Username:         u.Username,
		Nickname:         u.Nickname,
		Avatar:           u.Avatar,
		Roles:            rolesStr,
		Permissions:      security.NormalizeStrings(permissions),
	}
	cookie = &RefreshCookie{Value: refreshValue, ExpiresAt: expiresAt}
	return nil
	})
	if compromiseEvent != nil {
		if revokeErr := s.persistRefreshCompromise(ctx, *compromiseEvent); revokeErr != nil {
			return nil, nil, revokeErr
		}
	}
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
		s.writeAuditLog(ctx, &AuditLog{EventType: "session_revoked", Result: "success", SessionID: preferredSessionID(cookieSessionID, principal), IPHash: ClientIPHashFromContext(ctx), UserAgentHash: UserAgentHashFromContext(ctx), MetadataJSON: cloneMetadata(metadata), OccurredAt: now})
	}
	s.writeAuditLog(ctx, &AuditLog{EventType: "logout_success", Result: "success", SessionID: preferredSessionID(cookieSessionID, principal), IPHash: ClientIPHashFromContext(ctx), UserAgentHash: UserAgentHashFromContext(ctx), MetadataJSON: metadata, OccurredAt: now})
	return nil
}

func (s *service) Me(ctx context.Context) (*MeResponse, error) {
	p, ok := security.PrincipalFromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}
	if s.userRepo == nil {
		return nil, ErrUserNotFound
	}
	userID, _ := strconv.ParseInt(p.UserID, 10, 64)
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrUserNotFound
	}
	return &MeResponse{
		UserID:      u.ID,
		Username:    u.Username,
		Nickname:    u.Nickname,
		Email:       u.Email,
		Avatar:      u.Avatar,
		Gender:      u.Gender,
		Mobile:      u.Mobile,
		IsTeacher:   u.IsTeacher,
		TeacherID:   nil,
		Status:      u.Status,
		Roles:       p.Roles,
		Permissions: p.Permissions,
		Menus:       []any{},
		Buttons:     []string{},
		SessionID:   p.SessionID,
	}, nil
}

func sha256Hex(value string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(value)))
	return hex.EncodeToString(sum[:])
}

func (s *service) clearRefreshCookie() *RefreshCookie {
	return &RefreshCookie{Clear: true}
}

type refreshCompromiseEvent struct {
	familyID      string
	sessionID     string
	userPublicID  string
	eventType     string
	result        string
	occurredAt    time.Time
	ipHash        string
	userAgentHash string
	metadata      map[string]any
}

func (s *service) persistRefreshCompromise(ctx context.Context, event refreshCompromiseEvent) error {
	if strings.TrimSpace(event.familyID) == "" && strings.TrimSpace(event.sessionID) == "" {
		return nil
	}
	baseCtx := context.Background()
	if requestID, ok := observability.RequestIDFromContext(ctx); ok {
		baseCtx = observability.WithRequestID(baseCtx, requestID)
	}
	if traceID, ok := observability.TraceIDFromContext(ctx); ok {
		baseCtx = observability.WithTraceID(baseCtx, traceID)
	}
	return s.txManager.WithinTx(baseCtx, func(txCtx context.Context) error {
		if strings.TrimSpace(event.familyID) != "" {
			if err := s.repo.RevokeRefreshTokenFamily(txCtx, event.familyID); err != nil {
				return err
			}
		}
		if strings.TrimSpace(event.sessionID) != "" {
			if err := s.repo.RevokeSession(txCtx, event.sessionID); err != nil {
				return err
			}
		}
		s.writeAuditLog(txCtx, &AuditLog{
			EventType:     event.eventType,
			Result:        event.result,
			UserPublicID:  event.userPublicID,
			SessionID:     event.sessionID,
			IPHash:        event.ipHash,
			UserAgentHash: event.userAgentHash,
			MetadataJSON:  cloneMetadata(event.metadata),
			OccurredAt:    event.occurredAt,
		})
		return nil
	})
}

type loginAttemptTarget struct {
	IdentifierHash string
	UserID         *int64
	IPHash         string
	UserAgentHash  string
}

func (s *service) writeLoginAttempt(ctx context.Context, target loginAttemptTarget, success bool, reason string) {
	if s.repo == nil {
		return
	}
	now := time.Now().UTC()
	requestID, _ := observability.RequestIDFromContext(ctx)
	traceID, _ := observability.TraceIDFromContext(ctx)
	_ = s.repo.CreateLoginAttempt(ctx, &LoginAttempt{IdentifierHash: target.IdentifierHash, UserID: target.UserID, Success: success, Reason: reason, IPHash: target.IPHash, UserAgentHash: target.UserAgentHash, RequestID: requestID, TraceID: traceID, CreatedAt: now})
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

func (s *service) isLockedOut(ctx context.Context, filter FailedLoginAttemptFilter) bool {
	if s.repo == nil || s.config.MaxFailedAttempts <= 0 || s.config.FailedWindow <= 0 || s.config.LockDuration <= 0 {
		return false
	}
	recent, err := s.repo.ListRecentFailedAttemptTimes(ctx, filter)
	if err != nil || len(recent) < s.config.MaxFailedAttempts {
		return false
	}
	cutoff := time.Now().UTC().Add(-s.config.LockDuration)
	count := 0
	for _, occurredAt := range recent {
		if occurredAt.Before(filter.Since) {
			continue
		}
		if occurredAt.Before(cutoff) {
			continue
		}
		count++
		if count >= s.config.MaxFailedAttempts {
			return true
		}
	}
	return false
}

func requestHashes(ctx context.Context) (string, string) {
	return ClientIPHashFromContext(ctx), UserAgentHashFromContext(ctx)
}

func ClientIPHashFromContext(ctx context.Context) string {
	return hashClientMetadata(ClientIPFromContext(ctx))
}

func UserAgentHashFromContext(ctx context.Context) string {
	return hashClientMetadata(UserAgentFromContext(ctx))
}

func hashClientMetadata(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		trimmed = "unknown"
	}
	return sha256Hex(trimmed)
}

func ptrInt64(value int64) *int64 {
	return &value
}

func joinRoles(roles []string) string {
	if len(roles) == 0 {
		return ""
	}
	return strings.Join(roles, ",")
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
