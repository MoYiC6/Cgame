package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password string, encodedHash string) (bool, error)
}

type Argon2idHasher struct {
	memoryKiB   uint32
	iterations  uint32
	parallelism uint8
	pepper      string
}

func NewArgon2idHasher(memoryKiB, iterations, parallelism int, pepper string) *Argon2idHasher {
	return &Argon2idHasher{
		memoryKiB:   uint32(memoryKiB),
		iterations:  uint32(iterations),
		parallelism: uint8(parallelism),
		pepper:      pepper,
	}
}

func (h *Argon2idHasher) Hash(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	derived := argon2.IDKey([]byte(password+h.pepper), salt, h.iterations, h.memoryKiB, h.parallelism, 32)
	return fmt.Sprintf(
		"argon2id$%d$%d$%d$%s$%s",
		h.memoryKiB,
		h.iterations,
		h.parallelism,
		base64.RawURLEncoding.EncodeToString(salt),
		base64.RawURLEncoding.EncodeToString(derived),
	), nil
}

func (h *Argon2idHasher) Verify(password string, encodedHash string) (bool, error) {
	memoryKiB, iterations, parallelism, salt, want, err := parseArgon2idHash(encodedHash)
	if err != nil {
		return false, nil
	}
	got := argon2.IDKey([]byte(password+h.pepper), salt, iterations, memoryKiB, parallelism, uint32(len(want)))
	return subtle.ConstantTimeCompare(got, want) == 1, nil
}

func (h *Argon2idHasher) Supports(encodedHash string) bool {
	return strings.HasPrefix(encodedHash, "argon2id$")
}

func parseArgon2idHash(encodedHash string) (uint32, uint32, uint8, []byte, []byte, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 || parts[0] != "argon2id" {
		return 0, 0, 0, nil, nil, fmt.Errorf("invalid argon2id hash format")
	}
	memoryKiB, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		return 0, 0, 0, nil, nil, fmt.Errorf("invalid argon2id memory: %w", err)
	}
	iterations, err := strconv.ParseUint(parts[2], 10, 32)
	if err != nil {
		return 0, 0, 0, nil, nil, fmt.Errorf("invalid argon2id iterations: %w", err)
	}
	parallelism, err := strconv.ParseUint(parts[3], 10, 8)
	if err != nil {
		return 0, 0, 0, nil, nil, fmt.Errorf("invalid argon2id parallelism: %w", err)
	}
	salt, err := base64.RawURLEncoding.DecodeString(parts[4])
	if err != nil {
		return 0, 0, 0, nil, nil, err
	}
	want, err := base64.RawURLEncoding.DecodeString(parts[5])
	if err != nil {
		return 0, 0, 0, nil, nil, err
	}
	return uint32(memoryKiB), uint32(iterations), uint8(parallelism), salt, want, nil
}
