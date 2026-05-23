package security

import "testing"

func TestArgon2idHasherHashAndVerify(t *testing.T) {
	h := NewArgon2idHasher(19456, 2, 1, "pepper")
	hash, err := h.Hash("correct horse battery staple")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}
	ok, err := h.Verify("correct horse battery staple", hash)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if !ok {
		t.Fatal("expected verify success")
	}
}
