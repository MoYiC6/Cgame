package coupon

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
	CouponTypeFixed    = 1
	CouponTypeDiscount = 2
	CouponTypeExchange = 3

	CouponDistributionPrivate = 0
	CouponDistributionLevel   = 1
	CouponDistributionPublic  = 2

	UserCouponStatusAvailable = 0
	UserCouponStatusUsed      = 1
	UserCouponStatusExpired   = 2

	CouponStatusAvailable  = "available"
	CouponStatusExpired    = "expired"
	CouponStatusSoldOut    = "soldout"
	CouponStatusNotStarted = "notstarted"
	CouponStatusDisabled   = "disabled"
)

type Coupon struct {
	ID                    int64      `json:"id"`
	Name                  string     `json:"name"`
	Type                  int        `json:"type"`
	FaceValue             float64    `json:"faceValue"`
	MaxDiscountAmount     *float64   `json:"maxDiscountAmount,omitempty"`
	MinOrderAmount        float64    `json:"minOrderAmount"`
	TotalQuantity         int        `json:"totalQuantity"`
	ClaimedQuantity       int        `json:"claimedQuantity"`
	UsedQuantity          int        `json:"usedQuantity"`
	PerUserLimit          int        `json:"perUserLimit"`
	ValidDays             int        `json:"validDays"`
	StartTime             *time.Time `json:"startTime,omitempty"`
	EndTime               *time.Time `json:"endTime,omitempty"`
	ApplicableScope       string     `json:"applicableScope"`
	DistributionMode      int        `json:"distributionMode"`
	TargetLevelIDs        string     `json:"targetLevelIds"`
	Enabled               bool       `json:"enabled"`
	IsPermanent           bool       `json:"isPermanent"`
	RestrictedGoodsIDs    []int64    `json:"restrictedGoodsIds"`
	RestrictedCategoryIDs []int64    `json:"restrictedCategoryIds"`
	CreatedAt             time.Time  `json:"createdAt"`
	UpdatedAt             time.Time  `json:"updatedAt"`
}

type CouponVO struct {
	ID                int64      `json:"id"`
	Name              string     `json:"name"`
	Type              int        `json:"type"`
	TypeDesc          string     `json:"typeDesc"`
	FaceValue         float64    `json:"faceValue"`
	MinOrderAmount    float64    `json:"minOrderAmount"`
	TotalQuantity     int        `json:"totalQuantity"`
	ClaimedQuantity   int        `json:"claimedQuantity"`
	RemainingQuantity int        `json:"remainingQuantity"`
	PerUserLimit      int        `json:"perUserLimit"`
	ValidDays         int        `json:"validDays"`
	StartTime         *time.Time `json:"startTime,omitempty"`
	EndTime           *time.Time `json:"endTime,omitempty"`
	ApplicableScope   []string   `json:"applicableScope"`
	Claimed           bool       `json:"claimed"`
	Claimable         bool       `json:"claimable"`
	SoldOut           bool       `json:"soldOut"`
	ConditionDesc     string     `json:"conditionDesc"`
}

type UserCouponVO struct {
	ID                int64      `json:"id"`
	CouponID          int64      `json:"couponId"`
	Name              string     `json:"name"`
	Type              int        `json:"type"`
	TypeDesc          string     `json:"typeDesc"`
	FaceValue         float64    `json:"faceValue"`
	MaxDiscountAmount *float64   `json:"maxDiscountAmount,omitempty"`
	MinOrderAmount    float64    `json:"minOrderAmount"`
	ApplicableScope   []string   `json:"applicableScope"`
	Status            int        `json:"status"`
	StatusDesc        string     `json:"statusDesc"`
	Source            string     `json:"source"`
	SourceDesc        string     `json:"sourceDesc"`
	OrderID           *int64     `json:"orderId,omitempty"`
	ClaimedAt         *time.Time `json:"claimedAt,omitempty"`
	UsedAt            *time.Time `json:"usedAt,omitempty"`
	ExpireAt          *time.Time `json:"expireAt,omitempty"`
	ExpiringSoon      bool       `json:"expiringSoon"`
	ConditionDesc     string     `json:"conditionDesc"`
	UsableForOrder    bool       `json:"usableForOrder"`
	UnusableReason    string     `json:"unusableReason,omitempty"`
}

type AdminCouponVO struct {
	ID                    int64      `json:"id"`
	Name                  string     `json:"name"`
	Type                  int        `json:"type"`
	TypeDesc              string     `json:"typeDesc"`
	FaceValue             float64    `json:"faceValue"`
	FaceValueDesc         string     `json:"faceValueDesc"`
	MaxDiscountAmount     *float64   `json:"maxDiscountAmount,omitempty"`
	MinOrderAmount        float64    `json:"minOrderAmount"`
	TotalQuantity         int        `json:"totalQuantity"`
	ClaimedQuantity       int        `json:"claimedQuantity"`
	RemainingQuantity     int        `json:"remainingQuantity"`
	UsedQuantity          int        `json:"usedQuantity"`
	PerUserLimit          int        `json:"perUserLimit"`
	ValidDays             int        `json:"validDays"`
	StartTime             *time.Time `json:"startTime,omitempty"`
	EndTime               *time.Time `json:"endTime,omitempty"`
	ApplicableScope       string     `json:"applicableScope"`
	DistributionMode      int        `json:"distributionMode"`
	TargetLevelIDs        string     `json:"targetLevelIds"`
	Enabled               bool       `json:"enabled"`
	IsPermanent           bool       `json:"isPermanent"`
	RestrictedGoodsIDs    []int64    `json:"restrictedGoodsIds"`
	RestrictedCategoryIDs []int64    `json:"restrictedCategoryIds"`
	RestrictionType       string     `json:"restrictionType"`
	RestrictionTypeDesc   string     `json:"restrictionTypeDesc"`
	Status                string     `json:"status"`
	StatusDesc            string     `json:"statusDesc"`
	CreatedAt             time.Time  `json:"createdAt"`
	UpdatedAt             time.Time  `json:"updatedAt"`
}

type CouponPage struct {
	Records []AdminCouponVO `json:"records"`
	Total   int64           `json:"total"`
	Current int             `json:"current"`
	Size    int             `json:"size"`
}

type CouponPageResult = *CouponPage

type CouponStats struct {
	TotalCoupons   int `json:"totalCoupons"`
	EnabledCoupons int `json:"enabledCoupons"`
	ClaimedCoupons int `json:"claimedCoupons"`
	UsedCoupons    int `json:"usedCoupons"`
}

type CouponQuery struct {
	PageNum     int
	PageSize    int
	Name        string
	Type        *int
	Enabled     *bool
	Status      string
	IsPermanent *bool
}

type CouponCreateRequest struct {
	Name                  string     `json:"name"`
	Type                  int        `json:"type"`
	FaceValue             float64    `json:"faceValue"`
	MaxDiscountAmount     *float64   `json:"maxDiscountAmount"`
	MinOrderAmount        float64    `json:"minOrderAmount"`
	TotalQuantity         int        `json:"totalQuantity"`
	PerUserLimit          int        `json:"perUserLimit"`
	ValidDays             int        `json:"validDays"`
	StartTime             *time.Time `json:"startTime"`
	EndTime               *time.Time `json:"endTime"`
	ApplicableScope       string     `json:"applicableScope"`
	DistributionMode      int        `json:"distributionMode"`
	DistributionModeSet   bool       `json:"-"`
	TargetLevelIDs        []int64    `json:"targetLevelIds"`
	Enabled               *bool      `json:"enabled"`
	IsPermanent           *bool      `json:"isPermanent"`
	RestrictedGoodsIDs    []int64    `json:"restrictedGoodsIds"`
	RestrictedCategoryIDs []int64    `json:"restrictedCategoryIds"`
}

type CouponUpdateRequest struct {
	Name                  *string    `json:"name"`
	Type                  *int       `json:"type"`
	FaceValue             *float64   `json:"faceValue"`
	MaxDiscountAmount     *float64   `json:"maxDiscountAmount"`
	MinOrderAmount        *float64   `json:"minOrderAmount"`
	TotalQuantity         *int       `json:"totalQuantity"`
	PerUserLimit          *int       `json:"perUserLimit"`
	ValidDays             *int       `json:"validDays"`
	StartTime             *time.Time `json:"startTime"`
	EndTime               *time.Time `json:"endTime"`
	ApplicableScope       *string    `json:"applicableScope"`
	DistributionMode      *int       `json:"distributionMode"`
	TargetLevelIDs        []int64    `json:"targetLevelIds"`
	TargetLevelIDsSet     bool       `json:"-"`
	Enabled               *bool      `json:"enabled"`
	IsPermanent           *bool      `json:"isPermanent"`
	RestrictedGoodsIDs    []int64    `json:"restrictedGoodsIds"`
	RestrictedGoodsIDsSet bool       `json:"-"`
	RestrictedCategoryIDs []int64    `json:"restrictedCategoryIds"`
	RestrictedCategorySet bool       `json:"-"`
}

func typeDesc(couponType int) string {
	switch couponType {
	case CouponTypeFixed:
		return "满减券"
	case CouponTypeDiscount:
		return "折扣券"
	case CouponTypeExchange:
		return "兑换券"
	default:
		return "未知类型"
	}
}

func userCouponStatusDesc(status int) string {
	switch status {
	case UserCouponStatusAvailable:
		return "可用"
	case UserCouponStatusUsed:
		return "已使用"
	case UserCouponStatusExpired:
		return "已过期"
	default:
		return "未知状态"
	}
}

func sourceDesc(source string) string {
	switch source {
	case "claim":
		return "领取"
	case "invite_reward":
		return "邀请奖励"
	case "new_user":
		return "新用户奖励"
	case "admin_grant":
		return "管理员赠送"
	default:
		if source == "" {
			return ""
		}
		return source
	}
}

func couponStatus(c Coupon, now time.Time) (string, string) {
	if !c.Enabled {
		return CouponStatusDisabled, "已禁用"
	}
	if c.TotalQuantity > 0 && c.ClaimedQuantity >= c.TotalQuantity {
		return CouponStatusSoldOut, "已抢光"
	}
	if c.StartTime != nil && now.Before(*c.StartTime) {
		return CouponStatusNotStarted, "未开始"
	}
	if c.EndTime != nil && now.After(*c.EndTime) {
		return CouponStatusExpired, "已过期"
	}
	return CouponStatusAvailable, "可领取"
}

func buildConditionDesc(minOrderAmount float64) string {
	if minOrderAmount <= 0 {
		return "无门槛"
	}
	return fmt.Sprintf("满%.2f元可用", minOrderAmount)
}

func buildFaceValueDesc(couponType int, faceValue float64, minOrderAmount float64) string {
	switch couponType {
	case CouponTypeFixed:
		return fmt.Sprintf("满%.2f减%.2f", minOrderAmount, faceValue)
	case CouponTypeDiscount:
		return fmt.Sprintf("%.0f折", faceValue/10)
	case CouponTypeExchange:
		return "兑换券"
	default:
		return fmt.Sprintf("%.2f", faceValue)
	}
}

func restrictionType(goodsIDs, categoryIDs []int64) (string, string) {
	if len(goodsIDs) > 0 {
		return "goods", "指定商品"
	}
	if len(categoryIDs) > 0 {
		return "category", "指定分类"
	}
	return "all", "全部商品"
}

func stringListFromJSON(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []string{}
	}
	var values []string
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return []string{}
	}
	return values
}

func int64ListToJSON(values []int64) string {
	if values == nil {
		return ""
	}
	raw, err := json.Marshal(values)
	if err != nil {
		return "[]"
	}
	return string(raw)
}

func int64ListFromJSON(raw string) []int64 {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []int64{}
	}
	var values []int64
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return []int64{}
	}
	return values
}
