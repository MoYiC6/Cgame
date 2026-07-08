package security

import (
	"testing"
)

func TestMultiHasher_BCryptThenArgon2id(t *testing.T) {
	bcryptHasher := NewBCryptHasher(0)
	argon2Hasher := NewArgon2idHasher(19456, 2, 1, "pepper")
	multi := NewMultiHasher(argon2Hasher, bcryptHasher)

	// Java BCrypt
	javaHash, err := bcryptHasher.Hash("admin123")
	if err != nil {
		t.Fatal(err)
	}

	// Go Argon2id
	goHash, err := argon2Hasher.Hash("newpassword")
	if err != nil {
		t.Fatal(err)
	}

	// verify BCrypt
	ok, err := multi.Verify("admin123", javaHash)
	if err != nil || !ok {
		t.Fatalf("verify bcrypt failed: ok=%v err=%v", ok, err)
	}

	// verify Argon2id
	ok, err = multi.Verify("newpassword", goHash)
	if err != nil || !ok {
		t.Fatalf("verify argon2id failed: ok=%v err=%v", ok, err)
	}

	// wrong password
	ok, err = multi.Verify("wrong", javaHash)
	if err != nil || ok {
		t.Fatalf("wrong password should fail: ok=%v err=%v", ok, err)
	}

	// new hash should be argon2id (primary)
	newHash, err := multi.Hash("fresh")
	if err != nil {
		t.Fatal(err)
	}
	if len(newHash) < 20 || newHash[:9] != "argon2id$" {
		t.Fatalf("expected argon2id hash, got: %s", newHash)
	}
}
