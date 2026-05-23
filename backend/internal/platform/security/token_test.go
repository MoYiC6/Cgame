package security

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

func TestTokenManagerIssueAndVerifyRoundTrip(t *testing.T) {
	mgr := NewHMACTokenManager(HMACTokenConfig{
		Issuer:         "backend",
		Audience:       "admin-api",
		KeyID:          "test-key",
		Secret:         []byte("01234567890123456789012345678901"),
		AccessTokenTTL: 15 * time.Minute,
		ClockSkew:      30 * time.Second,
	})

	p := &Principal{UserID: "42", PublicID: "usr_123", SessionID: "ses_123", Roles: []string{"admin"}, Permissions: []string{"order:read", "order:read"}, Status: "active"}
	tok, err := mgr.IssueAccessToken(context.Background(), p)
	if err != nil {
		t.Fatalf("IssueAccessToken() error = %v", err)
	}

	principal, claims, err := mgr.VerifyAccessToken(context.Background(), tok.Token)
	if err != nil {
		t.Fatalf("VerifyAccessToken() error = %v", err)
	}
	if principal.UserID != "42" || claims.Subject != "42" {
		t.Fatalf("expected user id subject round trip, principal=%+v claims=%+v", principal, claims)
	}
	if principal.PublicID != "usr_123" {
		t.Fatalf("expected public id to round trip via explicit claim, got %+v", principal)
	}
	if principal.Status != "active" {
		t.Fatalf("expected status active, got %+v", principal)
	}
	if !reflect.DeepEqual(principal.Permissions, []string{"order:read"}) {
		t.Fatalf("expected deduped sorted permissions, got %#v", principal.Permissions)
	}
}

func TestTokenManagerRejectsTokenWithoutExp(t *testing.T) {
	mgr := newTestTokenManager()
	raw := signTestToken(t, mgr.config.Secret, mgr.config.KeyID, customClaims{
		SessionID: "ses_123",
		Status:    "active",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    mgr.config.Issuer,
			Subject:   "42",
			Audience:  jwt.ClaimStrings{mgr.config.Audience},
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			NotBefore: jwt.NewNumericDate(time.Now().UTC()),
		},
	})

	_, _, err := mgr.VerifyAccessToken(context.Background(), raw)
	if !errors.Is(err, ErrTokenInvalid) {
		t.Fatalf("expected ErrTokenInvalid, got %v", err)
	}
}

func TestTokenManagerRejectsTokenWithoutNbf(t *testing.T) {
	mgr := newTestTokenManager()
	now := time.Now().UTC()
	raw := signTestToken(t, mgr.config.Secret, mgr.config.KeyID, customClaims{
		SessionID: "ses_123",
		Status:    "active",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    mgr.config.Issuer,
			Subject:   "42",
			Audience:  jwt.ClaimStrings{mgr.config.Audience},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
		},
	})

	_, _, err := mgr.VerifyAccessToken(context.Background(), raw)
	if !errors.Is(err, ErrTokenInvalid) {
		t.Fatalf("expected ErrTokenInvalid, got %v", err)
	}
}

func TestTokenManagerRejectsTokenWithoutIat(t *testing.T) {
	mgr := newTestTokenManager()
	now := time.Now().UTC()
	raw := signTestToken(t, mgr.config.Secret, mgr.config.KeyID, customClaims{
		SessionID: "ses_123",
		Status:    "active",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    mgr.config.Issuer,
			Subject:   "42",
			Audience:  jwt.ClaimStrings{mgr.config.Audience},
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
		},
	})

	_, _, err := mgr.VerifyAccessToken(context.Background(), raw)
	if !errors.Is(err, ErrTokenInvalid) {
		t.Fatalf("expected ErrTokenInvalid, got %v", err)
	}
}

func TestTokenManagerRejectsMismatchedKid(t *testing.T) {
	mgr := newTestTokenManager()
	now := time.Now().UTC()
	raw := signTestToken(t, mgr.config.Secret, "other-key", customClaims{
		SessionID: "ses_123",
		Status:    "active",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    mgr.config.Issuer,
			Subject:   "42",
			Audience:  jwt.ClaimStrings{mgr.config.Audience},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
		},
	})

	_, _, err := mgr.VerifyAccessToken(context.Background(), raw)
	if !errors.Is(err, ErrTokenInvalid) {
		t.Fatalf("expected ErrTokenInvalid, got %v", err)
	}
}

func TestNewHMACTokenManagerRejectsWeakSecret(t *testing.T) {
	defer func() {
		if recovered := recover(); recovered == nil {
			t.Fatal("expected panic for weak secret")
		}
	}()

	_ = NewHMACTokenManager(HMACTokenConfig{
		Issuer:         "backend",
		Audience:       "admin-api",
		KeyID:          "test-key",
		Secret:         []byte("short-secret"),
		AccessTokenTTL: 15 * time.Minute,
		ClockSkew:      30 * time.Second,
	})
}

func newTestTokenManager() *HMACTokenManager {
	return NewHMACTokenManager(HMACTokenConfig{
		Issuer:         "backend",
		Audience:       "admin-api",
		KeyID:          "test-key",
		Secret:         []byte("01234567890123456789012345678901"),
		AccessTokenTTL: 15 * time.Minute,
		ClockSkew:      30 * time.Second,
	})
}

func signTestToken(t *testing.T, secret []byte, kid string, claims customClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["kid"] = kid
	raw, err := token.SignedString(secret)
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}
	return raw
}

func TestTokenManagerVerifyAccessTokenFallsBackToSubjectForLegacyPublicID(t *testing.T) {
	mgr := newTestTokenManager()
	now := time.Now().UTC()
	raw := signTestToken(t, mgr.config.Secret, mgr.config.KeyID, customClaims{
		SessionID: "ses_legacy",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    mgr.config.Issuer,
			Subject:   "usr_legacy",
			Audience:  jwt.ClaimStrings{mgr.config.Audience},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
		},
	})

	principal, claims, err := mgr.VerifyAccessToken(context.Background(), raw)
	if err != nil {
		t.Fatalf("VerifyAccessToken() error = %v", err)
	}
	if principal.UserID != "" {
		t.Fatalf("expected legacy token to avoid fabricating user id, got %+v", principal)
	}
	if principal.PublicID != "usr_legacy" {
		t.Fatalf("expected legacy subject fallback into public id, got %+v", principal)
	}
	if claims.PublicID != "usr_legacy" {
		t.Fatalf("expected legacy subject fallback into claims public id, got %+v", claims)
	}
}
