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

type UserLoginLog struct {
	ID          int64
	UserID      *int64
	LoginType   string
	IPAddress   string
	UserAgent   string
	LoginStatus string
	FailReason  string
	CreatedAt   time.Time
}

type UserQuery struct {
	PageNum    int
	PageSize   int
	Username   string
	Nickname   string
	Mobile     string
	Email      string
	Status     *int16
	IsTeacher  *int16
	LevelID    *int64
	StartTime  *time.Time
	EndTime    *time.Time
}

type UserCenterInfo struct {
	ID               int64      `json:"id"`
	Username         string     `json:"username"`
	Nickname         string     `json:"nickname"`
	Avatar           string     `json:"avatar"`
	Mobile           string     `json:"mobile"`
	Email            string     `json:"email"`
	Gender           int16      `json:"gender"`
	Birthday         *time.Time `json:"birthday,omitempty"`
	Province         string     `json:"province"`
	City             string     `json:"city"`
	District         string     `json:"district"`
	Intro            string     `json:"intro"`
	Balance          float64    `json:"balance"`
	FrozenBalance    float64    `json:"frozenBalance"`
	TotalRecharge    float64    `json:"totalRecharge"`
	TotalConsumption float64    `json:"totalConsumption"`
	LevelID          *int64     `json:"levelId,omitempty"`
	LevelName        string     `json:"levelName"`
	IsTeacher        int16      `json:"isTeacher"`
	RealNameStatus   int16      `json:"realNameStatus"`
	Status           string     `json:"status"`
}

type UpdateProfileRequest struct {
	Nickname string     `json:"nickname"`
	Avatar   string     `json:"avatar"`
	Gender   int16      `json:"gender"`
	Birthday *time.Time `json:"birthday,omitempty"`
	Province string     `json:"province"`
	City     string     `json:"city"`
	District string     `json:"district"`
	Intro    string     `json:"intro"`
}

type UpdateUserStatusRequest struct {
	Status int16 `json:"status"`
}

type CreateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
	RealName string `json:"realName"`
	Mobile   string `json:"mobile"`
	Status   int16  `json:"status"`
	IsTeacher int16 `json:"isTeacher"`
}

type ConsumptionRankingItem struct {
	UserID       int64   `json:"userId"`
	Username     string  `json:"username"`
	Nickname     string  `json:"nickname"`
	Avatar       string  `json:"avatar"`
	Consumption  float64 `json:"consumption"`
	Rank         int     `json:"rank"`
}

type UserSelectorItem struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Mobile   string `json:"mobile"`
	Avatar   string `json:"avatar"`
}
