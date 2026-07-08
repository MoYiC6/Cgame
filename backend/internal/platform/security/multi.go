package security

import "errors"

var ErrNoMatchingHasher = errors.New("no matching hasher for password format")

type MultiHasher struct {
	hashers   []PasswordHasher
	primary   PasswordHasher
}

func NewMultiHasher(primary PasswordHasher, fallbacks ...PasswordHasher) *MultiHasher {
	hashers := append([]PasswordHasher{primary}, fallbacks...)
	return &MultiHasher{hashers: hashers, primary: primary}
}

func (m *MultiHasher) Hash(password string) (string, error) {
	return m.primary.Hash(password)
}

func (m *MultiHasher) Verify(password string, encodedHash string) (bool, error) {
	for _, h := range m.hashers {
		if s, ok := h.(interface{ Supports(string) bool }); ok && !s.Supports(encodedHash) {
			continue
		}
		ok, err := h.Verify(password, encodedHash)
		if err != nil {
			continue
		}
		return ok, nil
	}
	return false, ErrNoMatchingHasher
}
