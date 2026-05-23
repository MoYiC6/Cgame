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

	p := &Principal{PublicID: "usr_123", SessionID: "ses_123", Roles: []string{"admin"}, Permissions: []string{"order:read", "order:read"}}
	tok, err := mgr.IssueAccessToken(context.Background(), p)
	if err != nil {
		t.Fatalf("IssueAccessToken() error = %v", err)
	}

	principal, claims, err := mgr.VerifyAccessToken(context.Background(), tok.Token)
	if err != nil {
		t.Fatalf("VerifyAccessToken() error = %v", err)
	}
	if principal.PublicID != "usr_123" || claims.Subject != "usr_123" {
		t.Fatalf("unexpected principal/claims: %+v %+v", principal, claims)
	}
	if !reflect.DeepEqual(principal.Permissions, []string{"order:read"}) {
		t.Fatalf("expected deduped sorted permissions, got %#v", principal.Permissions)
	}
}

func TestTokenManagerRejectsTokenWithoutExp(t *testing.T) {
	mgr := newTestTokenManager()
	raw := signTestToken(t, mgr.config.Secret, mgr.config.KeyID, customClaims{
		SessionID: "ses_123",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    mgr.config.Issuer,
			Subject:   "usr_123",
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
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    mgr.config.Issuer,
			Subject:   "usr_123",
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
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    mgr.config.Issuer,
			Subject:   "usr_123",
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
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    mgr.config.Issuer,
			Subject:   "usr_123",
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
