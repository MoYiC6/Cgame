package security

import (
	"crypto/rand"
	"encoding/base64"
)

type RandomTokenGenerator interface {
	GenerateURLSafe(n int) (string, error)
}

type CryptoRandomTokenGenerator struct{}

func (g CryptoRandomTokenGenerator) GenerateURLSafe(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
