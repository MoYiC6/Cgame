package user

import (
	"strings"
	"time"
)

type User struct {
	ID                int64
	PublicID          string
	Email             string
	PasswordHash      string
	Status            string
	PasswordChangedAt *time.Time
	LastLoginAt       *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func NormalizeEmail(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}
