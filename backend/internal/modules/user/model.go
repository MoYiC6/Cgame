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
	Balance           float64
	LevelID           *int64
	PasswordChangedAt *time.Time
	LastLoginAt       *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func NormalizeEmail(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

type UserBalanceLog struct {
	ID           int64
	UserID       int64
	ChangeType   string
	Amount       float64
	BalanceAfter float64
	RelatedID    *int64
	RelatedNo    *string
	Description  *string
	CreatedAt    time.Time
}

type UserLevel struct {
	ID             int64
	Name           string
	MinConsumption float64
	DiscountRate   float64
	Benefits       map[string]any
	Status         *int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type UserLevelLog struct {
	ID          int64
	UserID      int64
	OldLevelID  *int64
	NewLevelID  *int64
	ChangeReason *string
	CreatedAt   time.Time
}

type UserPurchaseRecord struct {
	ID           int64
	UserID       int64
	GoodsID      *int64
	OrderID      *int64
	Quantity     int
	PurchaseTime time.Time
}
