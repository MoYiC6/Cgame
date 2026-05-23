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

func TestArgon2idHasherVerifyUsesEncodedParameters(t *testing.T) {
	hasher := NewArgon2idHasher(19456, 2, 1, "pepper")
	hash, err := hasher.Hash("correct horse battery staple")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	verifier := NewArgon2idHasher(65536, 4, 2, "pepper")
	ok, err := verifier.Verify("correct horse battery staple", hash)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if !ok {
		t.Fatal("expected verify success when encoded hash parameters differ from current config")
	}
}
