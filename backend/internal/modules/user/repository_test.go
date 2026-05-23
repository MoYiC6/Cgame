package user

import "testing"

func TestNormalizeEmail(t *testing.T) {
	got := NormalizeEmail("  Admin@Example.COM ")
	if got != "admin@example.com" {
		t.Fatalf("expected normalized email, got %q", got)
	}
}
