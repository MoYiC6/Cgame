package user

import (
	"strings"
	"time"
)

type User struct {
	ID                int64
	PublicID          string
	Username          string
	Email             string
	PasswordHash      string
	Nickname          string
	RealName          string
	Mobile            string
	MobileVerified    int16
	EmailVerified     int16
	Avatar            string
	Gender            int16
	Birthday          *time.Time
	Intro             string
	Province          string
	City              string
	District          string
	Wechat            string
	WechatUnionID     string
	WechatMPOpenID    string
	WechatH5OpenID    string
	AlipayOpenID      string
	Status            string
	IsTeacher         int16
	IDCard            string
	IDCardFront       string
	IDCardBack        string
	RealNameStatus    int16
	RealNameVerifyType string
	RealNameSubmitTime *time.Time
	RealNameVerifyTime *time.Time
	RealNameRejectReason string
	LoginFailedCount  int
	LastLoginTime     *time.Time
	LastLoginIP       string
	LastLoginPlatform string
	PasswordUpdatedAt *time.Time
	PasswordChangedAt *time.Time
	RegisterIP        string
	RegisterPlatform  string
	RegisterSource    string
	Balance           float64
	FrozenBalance     float64
	TotalRecharge     float64
	TotalConsumption  float64
	LevelID           *int64
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
