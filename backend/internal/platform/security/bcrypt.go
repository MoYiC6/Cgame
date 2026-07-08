package security

import (
	"golang.org/x/crypto/bcrypt"
)

const bcryptDefaultCost = bcrypt.DefaultCost

type BCryptHasher struct {
	cost int
}

func NewBCryptHasher(cost int) *BCryptHasher {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}
	return &BCryptHasher{cost: cost}
}

func (h *BCryptHasher) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (h *BCryptHasher) Verify(password string, encodedHash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(encodedHash), []byte(password))
	if err == nil {
		return true, nil
	}
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return false, nil
	}
	return false, err
}

func (h *BCryptHasher) Supports(encodedHash string) bool {
	return len(encodedHash) >= 4 && (encodedHash[0:2] == "$2")
}
