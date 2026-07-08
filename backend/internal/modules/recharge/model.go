package recharge

import "time"

const (
	RechargeStatusPending   = "pending"
	RechargeStatusPaid      = "paid"
	RechargeStatusCancelled = "cancelled"
	RechargeStatusFailed    = "failed"
)

// RechargeRecord represents a recharge order in the database.
type RechargeRecord struct {
	ID           int64      `json:"id"`
	RechargeNo   string     `json:"rechargeNo"`
	UserID       int64      `json:"userId"`
	Amount       float64    `json:"amount"`
	GiftAmount   float64    `json:"giftAmount"`
	TotalAmount  float64    `json:"totalAmount"`
	PayAmount    float64    `json:"payAmount"`
	Status       string     `json:"status"`
	PayChannel   string     `json:"payChannel,omitempty"`
	PayTime      *time.Time `json:"payTime,omitempty"`
	CallbackTime *time.Time `json:"callbackTime,omitempty"`
	Remark       string     `json:"remark,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

// RechargeRecordVO is the view object for recharge records.
type RechargeRecordVO struct {
	ID          int64      `json:"id"`
	RechargeNo  string     `json:"rechargeNo"`
	Amount      float64    `json:"amount"`
	GiftAmount  float64    `json:"giftAmount"`
	TotalAmount float64    `json:"totalAmount"`
	Status      string     `json:"status"`
	StatusDesc  string     `json:"statusDesc"`
	PayChannel  string     `json:"payChannel,omitempty"`
	PayTime     *time.Time `json:"payTime,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
}

// RechargeRecordPageResult wraps paginated recharge records.
type RechargeRecordPageResult struct {
	Total    int64              `json:"total"`
	PageNum  int                `json:"pageNum"`
	PageSize int                `json:"pageSize"`
	Records  []RechargeRecordVO `json:"records"`
}

// RechargeStats holds recharge statistics.
type RechargeStats struct {
	TotalRecords   int     `json:"totalRecords"`
	PendingRecords int     `json:"pendingRecords"`
	PaidRecords    int     `json:"paidRecords"`
	TotalAmount    float64 `json:"totalAmount"`
	TotalGiftAmount float64 `json:"totalGiftAmount"`
}

// RechargeRebateRule represents a rebate rule for recharges.
type RechargeRebateRule struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	MinAmount  float64   `json:"minAmount"`
	GiftRate   float64   `json:"giftRate"`
	GiftAmount float64   `json:"giftAmount"`
	Enabled    bool      `json:"enabled"`
	Priority   int       `json:"priority"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// RechargeRebateRuleVO is the view object for rebate rules.
type RechargeRebateRuleVO struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	MinAmount  float64   `json:"minAmount"`
	GiftRate   float64   `json:"giftRate"`
	GiftAmount float64   `json:"giftAmount"`
	Enabled    bool      `json:"enabled"`
	Priority   int       `json:"priority"`
	CreatedAt  time.Time `json:"createdAt"`
}

// ManualRechargeRequest is the admin request for manual recharge.
type ManualRechargeRequest struct {
	UserID   int64   `json:"userId"`
	Amount   float64 `json:"amount"`
	GiftAmount float64 `json:"giftAmount"`
	Remark   string  `json:"remark"`
}

// CreateRechargeRequest is the client request to create a recharge order.
type CreateRechargeRequest struct {
	Amount float64 `json:"amount"`
}

// RechargeQuery is the admin query for listing recharge records.
type RechargeQuery struct {
	PageNum     int     `json:"pageNum"`
	PageSize    int     `json:"pageSize"`
	Status      string  `json:"status"`
	UserID      *int64  `json:"userId"`
	RechargeNo  string  `json:"rechargeNo"`
	CreateTimeStart *string `json:"createTimeStart"`
	CreateTimeEnd   *string `json:"createTimeEnd"`
}

// RebateRuleCreateRequest is the admin request to create a rebate rule.
type RebateRuleCreateRequest struct {
	Name       string  `json:"name"`
	MinAmount  float64 `json:"minAmount"`
	GiftRate   float64 `json:"giftRate"`
	GiftAmount float64 `json:"giftAmount"`
	Enabled    bool    `json:"enabled"`
	Priority   int     `json:"priority"`
}

// RebateRuleUpdateRequest is the admin request to update a rebate rule.
type RebateRuleUpdateRequest struct {
	Name       *string  `json:"name"`
	MinAmount  *float64 `json:"minAmount"`
	GiftRate   *float64 `json:"giftRate"`
	GiftAmount *float64 `json:"giftAmount"`
	Enabled    *bool    `json:"enabled"`
	Priority   *int     `json:"priority"`
}

// RebatePreviewResult is the preview of rebate for a given amount.
type RebatePreviewResult struct {
	Amount      float64 `json:"amount"`
	GiftAmount  float64 `json:"giftAmount"`
	TotalAmount float64 `json:"totalAmount"`
	RuleName    string  `json:"ruleName,omitempty"`
}

func statusDesc(status string) string {
	switch status {
	case RechargeStatusPending:
		return "待支付"
	case RechargeStatusPaid:
		return "已支付"
	case RechargeStatusCancelled:
		return "已取消"
	case RechargeStatusFailed:
		return "失败"
	default:
		return "未知状态"
	}
}

func toRechargeRecordVO(r RechargeRecord) RechargeRecordVO {
	return RechargeRecordVO{
		ID:          r.ID,
		RechargeNo:  r.RechargeNo,
		Amount:      r.Amount,
		GiftAmount:  r.GiftAmount,
		TotalAmount: r.TotalAmount,
		Status:      r.Status,
		StatusDesc:  statusDesc(r.Status),
		PayChannel:  r.PayChannel,
		PayTime:     r.PayTime,
		CreatedAt:   r.CreatedAt,
	}
}

func toRebateRuleVO(r RechargeRebateRule) RechargeRebateRuleVO {
	return RechargeRebateRuleVO{
		ID:         r.ID,
		Name:       r.Name,
		MinAmount:  r.MinAmount,
		GiftRate:   r.GiftRate,
		GiftAmount: r.GiftAmount,
		Enabled:    r.Enabled,
		Priority:   r.Priority,
		CreatedAt:  r.CreatedAt,
	}
}
