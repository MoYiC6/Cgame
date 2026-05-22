package security

import (
	"context"
	"reflect"
	"testing"
	"time"
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
