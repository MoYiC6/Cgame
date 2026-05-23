package security

import (
	"context"
	"testing"
)

func TestWithPrincipalRoundTrip(t *testing.T) {
	ctx := context.Background()
	want := &Principal{PublicID: "usr_123", SessionID: "ses_123", Roles: []string{"admin"}, Permissions: []string{"order:read"}}

	ctx = WithPrincipal(ctx, want)
	got, ok := PrincipalFromContext(ctx)
	if !ok {
		t.Fatal("expected principal in context")
	}
	if got.PublicID != want.PublicID || got.SessionID != want.SessionID {
		t.Fatalf("unexpected principal: %+v", got)
	}
}
