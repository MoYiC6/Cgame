package partner

import "time"

const (
	PartnerConfigStatusEnabled  = "enabled"
	PartnerConfigStatusDisabled = "disabled"

	PartnerTypeAgency   = "agency"
	PartnerTypePersonal = "personal"
	PartnerTypeOther    = "other"

	CooperationStatusActive   = "active"
	CooperationStatusInactive = "inactive"
	CooperationStatusExpired  = "expired"

	CooperationTypeExclusive = "exclusive"
	CooperationTypeNonExclusive = "non_exclusive"
)

// PartnerConfig represents a partner configuration in the database.
type PartnerConfig struct {
	ID             int64      `json:"id"`
	Name           string     `json:"name"`
	PartnerType    string     `json:"partnerType"`
	CommissionRate float64    `json:"commissionRate"`
	FixedFee       float64    `json:"fixedFee"`
	Description    string     `json:"description,omitempty"`
	ContactName    string     `json:"contactName,omitempty"`
	ContactPhone   string     `json:"contactPhone,omitempty"`
	ContactEmail   string     `json:"contactEmail,omitempty"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

// PartnerConfigVO is the view object for partner config.
type PartnerConfigVO struct {
	ID             int64      `json:"id"`
	Name           string     `json:"name"`
	PartnerType    string     `json:"partnerType"`
	PartnerTypeDesc string    `json:"partnerTypeDesc"`
	CommissionRate float64    `json:"commissionRate"`
	FixedFee       float64    `json:"fixedFee"`
	Description    string     `json:"description,omitempty"`
	ContactName    string     `json:"contactName,omitempty"`
	ContactPhone   string     `json:"contactPhone,omitempty"`
	ContactEmail   string     `json:"contactEmail,omitempty"`
	Status         string     `json:"status"`
	StatusDesc     string     `json:"statusDesc"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

// TeacherPartner represents a cooperation relationship between teacher and partner.
type TeacherPartner struct {
	ID              int64      `json:"id"`
	TeacherID       int64      `json:"teacherId"`
	PartnerID       int64      `json:"partnerId"`
	PartnerConfigID *int64     `json:"partnerConfigId,omitempty"`
	CooperationType string     `json:"cooperationType"`
	CommissionRate  float64    `json:"commissionRate"`
	StartDate       *time.Time `json:"startDate,omitempty"`
	EndDate         *time.Time `json:"endDate,omitempty"`
	Status          string     `json:"status"`
	Remark          string     `json:"remark,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

// TeacherPartnerVO is the view object for teacher partner.
type TeacherPartnerVO struct {
	ID              int64      `json:"id"`
	TeacherID       int64      `json:"teacherId"`
	PartnerID       int64      `json:"partnerId"`
	PartnerName     string     `json:"partnerName,omitempty"`
	PartnerConfigID *int64     `json:"partnerConfigId,omitempty"`
	CooperationType string     `json:"cooperationType"`
	CooperationTypeDesc string `json:"cooperationTypeDesc"`
	CommissionRate  float64    `json:"commissionRate"`
	StartDate       *time.Time `json:"startDate,omitempty"`
	EndDate         *time.Time `json:"endDate,omitempty"`
	Status          string     `json:"status"`
	StatusDesc      string     `json:"statusDesc"`
	Remark          string     `json:"remark,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
}

// PartnerConfigPageResult wraps paginated partner configs.
type PartnerConfigPageResult struct {
	Total    int64             `json:"total"`
	PageNum  int               `json:"pageNum"`
	PageSize int               `json:"pageSize"`
	Records  []PartnerConfigVO `json:"records"`
}

// TeacherPartnerPageResult wraps paginated teacher partners.
type TeacherPartnerPageResult struct {
	Total    int64              `json:"total"`
	PageNum  int                `json:"pageNum"`
	PageSize int                `json:"pageSize"`
	Records  []TeacherPartnerVO `json:"records"`
}

// PartnerConfigCreateRequest is the admin request to create partner config.
type PartnerConfigCreateRequest struct {
	Name           string  `json:"name"`
	PartnerType    string  `json:"partnerType"`
	CommissionRate float64 `json:"commissionRate"`
	FixedFee       float64 `json:"fixedFee"`
	Description    string  `json:"description"`
	ContactName    string  `json:"contactName"`
	ContactPhone   string  `json:"contactPhone"`
	ContactEmail   string  `json:"contactEmail"`
}

// PartnerConfigUpdateRequest is the admin request to update partner config.
type PartnerConfigUpdateRequest struct {
	Name           *string  `json:"name,omitempty"`
	PartnerType    *string  `json:"partnerType,omitempty"`
	CommissionRate *float64 `json:"commissionRate,omitempty"`
	FixedFee       *float64 `json:"fixedFee,omitempty"`
	Description    *string  `json:"description,omitempty"`
	ContactName    *string  `json:"contactName,omitempty"`
	ContactPhone   *string  `json:"contactPhone,omitempty"`
	ContactEmail   *string  `json:"contactEmail,omitempty"`
	Status         *string  `json:"status,omitempty"`
}

// TeacherPartnerCreateRequest is the admin request to create teacher partner.
type TeacherPartnerCreateRequest struct {
	TeacherID       int64      `json:"teacherId"`
	PartnerID       int64      `json:"partnerId"`
	PartnerConfigID *int64     `json:"partnerConfigId,omitempty"`
	CooperationType string     `json:"cooperationType"`
	CommissionRate  float64    `json:"commissionRate"`
	StartDate       *time.Time `json:"startDate,omitempty"`
	EndDate         *time.Time `json:"endDate,omitempty"`
	Remark          string     `json:"remark"`
}

// TeacherPartnerUpdateRequest is the admin request to update teacher partner.
type TeacherPartnerUpdateRequest struct {
	CooperationType *string    `json:"cooperationType,omitempty"`
	CommissionRate  *float64   `json:"commissionRate,omitempty"`
	StartDate       *time.Time `json:"startDate,omitempty"`
	EndDate         *time.Time `json:"endDate,omitempty"`
	Status          *string    `json:"status,omitempty"`
	Remark          *string    `json:"remark,omitempty"`
}

// PartnerConfigQuery is the admin query for listing partner configs.
type PartnerConfigQuery struct {
	PageNum  int    `json:"pageNum"`
	PageSize int    `json:"pageSize"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	PartnerType string `json:"partnerType"`
}

// TeacherPartnerQuery is the admin query for listing teacher partners.
type TeacherPartnerQuery struct {
	PageNum   int    `json:"pageNum"`
	PageSize  int    `json:"pageSize"`
	TeacherID int64  `json:"teacherId"`
	PartnerID int64  `json:"partnerId"`
	Status    string `json:"status"`
}

func statusDesc(status string) string {
	switch status {
	case PartnerConfigStatusEnabled:
		return "启用"
	case PartnerConfigStatusDisabled:
		return "禁用"
	case CooperationStatusActive:
		return "生效中"
	case CooperationStatusInactive:
		return "已停用"
	case CooperationStatusExpired:
		return "已过期"
	default:
		return "未知状态"
	}
}

func partnerTypeDesc(pt string) string {
	switch pt {
	case PartnerTypeAgency:
		return "代理机构"
	case PartnerTypePersonal:
		return "个人"
	case PartnerTypeOther:
		return "其他"
	default:
		return "未知类型"
	}
}

func cooperationTypeDesc(ct string) string {
	switch ct {
	case CooperationTypeExclusive:
		return "独家"
	case CooperationTypeNonExclusive:
		return "非独家"
	default:
		return "未知类型"
	}
}

func toPartnerConfigVO(p PartnerConfig) PartnerConfigVO {
	return PartnerConfigVO{
		ID:              p.ID,
		Name:            p.Name,
		PartnerType:     p.PartnerType,
		PartnerTypeDesc: partnerTypeDesc(p.PartnerType),
		CommissionRate:  p.CommissionRate,
		FixedFee:        p.FixedFee,
		Description:     p.Description,
		ContactName:     p.ContactName,
		ContactPhone:    p.ContactPhone,
		ContactEmail:    p.ContactEmail,
		Status:          p.Status,
		StatusDesc:      statusDesc(p.Status),
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
	}
}

func toTeacherPartnerVO(tp TeacherPartner, partnerName string) TeacherPartnerVO {
	return TeacherPartnerVO{
		ID:                  tp.ID,
		TeacherID:           tp.TeacherID,
		PartnerID:           tp.PartnerID,
		PartnerName:         partnerName,
		PartnerConfigID:     tp.PartnerConfigID,
		CooperationType:     tp.CooperationType,
		CooperationTypeDesc: cooperationTypeDesc(tp.CooperationType),
		CommissionRate:      tp.CommissionRate,
		StartDate:           tp.StartDate,
		EndDate:             tp.EndDate,
		Status:              tp.Status,
		StatusDesc:          statusDesc(tp.Status),
		Remark:              tp.Remark,
		CreatedAt:           tp.CreatedAt,
	}
}
