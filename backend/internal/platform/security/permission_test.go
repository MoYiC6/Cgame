package security

import "testing"

func TestHasPermissionUsesStableMembership(t *testing.T) {
	p := &Principal{Permissions: []string{"inventory:read", "order:read"}}
	if !HasPermission(p, "order:read") {
		t.Fatal("expected permission match")
	}
	if HasPermission(p, "order:write") {
		t.Fatal("unexpected permission match")
	}
}
