package invite

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

const (
	InviteCodeStatusUnused  = "unused"
	InviteCodeStatusUsed    = "used"
	InviteCodeStatusRevoked = "revoked"
	InviteCodeStatusExpired = "expired"
)

const inviteCodeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

// InviteInfo represents a user's invite summary for client display.
type InviteInfo struct {
	InviteCode              string `json:"inviteCode"`
	InviteLink              string `json:"inviteLink"`
	TotalInvited            int    `json:"totalInvited"`
	RewardedCount           int    `json:"rewardedCount"`
	InviterRewardCouponName string `json:"inviterRewardCouponName"`
	InviteeRewardCouponName string `json:"inviteeRewardCouponName"`
}

// InviteRecord represents a single invite relationship.
type InviteRecord struct {
	ID                    int64      `json:"id"`
	InviterID            int64      `json:"inviterId"`
	InviteeID            int64      `json:"inviteeId"`
	InviteCode           string     `json:"inviteCode"`
	InviterRewardCouponID *int64    `json:"inviterRewardCouponId,omitempty"`
	InviteeRewardCouponID *int64    `json:"inviteeRewardCouponId,omitempty"`
	CreateTime           time.Time  `json:"createTime"`
}

// TeacherInviteCode represents a teacher invite code in the database.
type TeacherInviteCode struct {
	ID          int64      `json:"id"`
	Code        string     `json:"code"`
	Status      string     `json:"status"`
	Remark      string     `json:"remark"`
	ExpireTime  *time.Time `json:"expireTime,omitempty"`
	CreatedBy   int64      `json:"createdBy"`
	CreateTime  time.Time  `json:"createTime"`
	UsedBy      *int64     `json:"usedBy,omitempty"`
	UsedTime    *time.Time `json:"usedTime,omitempty"`
	TeacherID   *int64     `json:"teacherId,omitempty"`
	RevokedBy   *int64     `json:"revokedBy,omitempty"`
	RevokedTime *time.Time `json:"revokedTime,omitempty"`
	UpdateTime  time.Time  `json:"updateTime"`
}

// GenerateRequest is the admin request to generate invite codes.
type GenerateRequest struct {
	Count     int    `json:"count"`
	ValidDays int    `json:"validDays"`
	Remark    string `json:"remark"`
}

// ApplyRequest is the client request to apply with a teacher invite code.
type ApplyRequest struct {
	InviteCode         string `json:"inviteCode"`
	AssessmentLevel    string `json:"assessmentLevel"`
	OrderName          string `json:"orderName"`
	AssessmentExaminer string `json:"assessmentExaminer"`
	AssessmentRecord   []any  `json:"assessmentRecord"`
}

// BindRequest is the client request to bind an inviter by invite code.
type BindRequest struct {
	InviteCode string `json:"inviteCode"`
}

// InviteCodeQuery is the admin query for listing teacher invite codes.
type InviteCodeQuery struct {
	PageNum         int     `json:"pageNum"`
	PageSize        int     `json:"pageSize"`
	Code            string  `json:"code"`
	Status          string  `json:"status"`
	CreateTimeStart *string `json:"createTimeStart"`
	CreateTimeEnd   *string `json:"createTimeEnd"`
	UsedBy          *int64  `json:"usedBy"`
	CreatedBy       *int64  `json:"createdBy"`
}

// TeacherInviteCodeVO is the list-item VO for admin invite codes.
type TeacherInviteCodeVO struct {
	ID          int64      `json:"id"`
	Code        string     `json:"code"`
	Status      string     `json:"status"`
	StatusDesc  string     `json:"statusDesc"`
	Remark      string     `json:"remark"`
	ExpireTime  *time.Time `json:"expireTime,omitempty"`
	CreatedBy   int64      `json:"createdBy"`
	CreateTime  time.Time  `json:"createTime"`
	UsedBy      *int64     `json:"usedBy,omitempty"`
	UsedTime    *time.Time `json:"usedTime,omitempty"`
	TeacherID   *int64     `json:"teacherId,omitempty"`
	RevokedBy   *int64     `json:"revokedBy,omitempty"`
	RevokedTime *time.Time `json:"revokedTime,omitempty"`
}

// TeacherInviteCodeDetailVO is the detailed VO for a single invite code.
type TeacherInviteCodeDetailVO struct {
	ID          int64      `json:"id"`
	Code        string     `json:"code"`
	Status      string     `json:"status"`
	StatusDesc  string     `json:"statusDesc"`
	Remark      string     `json:"remark"`
	ExpireTime  *time.Time `json:"expireTime,omitempty"`
	CreatedBy   int64      `json:"createdBy"`
	CreateTime  time.Time  `json:"createTime"`
	UsedBy      *int64     `json:"usedBy,omitempty"`
	UsedTime    *time.Time `json:"usedTime,omitempty"`
	TeacherID   *int64     `json:"teacherId,omitempty"`
	RevokedBy   *int64     `json:"revokedBy,omitempty"`
	RevokedTime *time.Time `json:"revokedTime,omitempty"`
	UpdateTime  time.Time  `json:"updateTime"`
}

// InviteCodeValidationResult is returned when verifying a teacher invite code.
type InviteCodeValidationResult struct {
	Valid  bool   `json:"valid"`
	Code   string `json:"code,omitempty"`
	Remark string `json:"remark,omitempty"`
}

// PageResult is the generic paginated result.
type PageResult struct {
	Total    int64 `json:"total"`
	PageNum  int   `json:"pageNum"`
	PageSize int   `json:"pageSize"`
	Records  any   `json:"records"`
}

// TeacherInviteCodePageResult wraps paginated teacher invite codes.
type TeacherInviteCodePageResult struct {
	Total    int64               `json:"total"`
	PageNum  int                 `json:"pageNum"`
	PageSize int                 `json:"pageSize"`
	Records  []TeacherInviteCodeVO `json:"records"`
}

// generateRandomInviteCode generates a random n-character alphanumeric code.
func generateRandomInviteCode(length int) string {
	var sb strings.Builder
	alphabetLen := len(inviteCodeAlphabet)
	for i := 0; i < length; i++ {
		sb.WriteByte(inviteCodeAlphabet[rand.Intn(alphabetLen)])
	}
	return sb.String()
}

func statusDesc(status string) string {
	switch status {
	case InviteCodeStatusUnused:
		return "未使用"
	case InviteCodeStatusUsed:
		return "已使用"
	case InviteCodeStatusRevoked:
		return "已撤销"
	case InviteCodeStatusExpired:
		return "已过期"
	default:
		return "未知状态"
	}
}

func resolveStatus(code TeacherInviteCode, now time.Time) string {
	if code.Status == InviteCodeStatusRevoked {
		return InviteCodeStatusRevoked
	}
	if code.Status == InviteCodeStatusUsed {
		return InviteCodeStatusUsed
	}
	if code.ExpireTime != nil && now.After(*code.ExpireTime) {
		return InviteCodeStatusExpired
	}
	return InviteCodeStatusUnused
}

func buildInviteLink(baseURL, inviteCode string) string {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = "https://feyo.club"
	}
	return fmt.Sprintf("%s?inviteCode=%s", strings.TrimRight(baseURL, "/"), inviteCode)
}
