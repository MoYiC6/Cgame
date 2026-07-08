package partner

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Partner config operations

func (s *Service) CreatePartnerConfig(ctx context.Context, req PartnerConfigCreateRequest) (int64, error) {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return 0, fmt.Errorf("name is required")
	}
	if req.PartnerType == "" {
		req.PartnerType = PartnerTypeAgency
	}
	if req.CommissionRate < 0 {
		return 0, fmt.Errorf("commission rate cannot be negative")
	}
	if req.FixedFee < 0 {
		return 0, fmt.Errorf("fixed fee cannot be negative")
	}
	config := &PartnerConfig{
		Name:           req.Name,
		PartnerType:    req.PartnerType,
		CommissionRate: req.CommissionRate,
		FixedFee:       req.FixedFee,
		Description:    req.Description,
		ContactName:    req.ContactName,
		ContactPhone:   req.ContactPhone,
		ContactEmail:   req.ContactEmail,
		Status:         PartnerConfigStatusEnabled,
	}
	if err := s.repo.CreatePartnerConfig(ctx, config); err != nil {
		return 0, fmt.Errorf("create partner config: %w", err)
	}
	return config.ID, nil
}

func (s *Service) ListPartnerConfigs(ctx context.Context, query PartnerConfigQuery) (*PartnerConfigPageResult, error) {
	query.PageNum = normalizePage(query.PageNum)
	query.PageSize = normalizePageSize(query.PageSize)
	return s.repo.ListPartnerConfigs(ctx, query)
}

func (s *Service) GetPartnerConfig(ctx context.Context, id int64) (*PartnerConfigVO, error) {
	if id == 0 {
		return nil, fmt.Errorf("partner config id is required")
	}
	config, err := s.repo.GetPartnerConfigByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("partner config not found")
	}
	vo := toPartnerConfigVO(*config)
	return &vo, nil
}

func (s *Service) UpdatePartnerConfig(ctx context.Context, id int64, req PartnerConfigUpdateRequest) error {
	if id == 0 {
		return fmt.Errorf("partner config id is required")
	}
	updates := map[string]any{}
	if req.Name != nil {
		value := strings.TrimSpace(*req.Name)
		if value == "" {
			return fmt.Errorf("name cannot be empty")
		}
		updates["name"] = value
	}
	if req.PartnerType != nil {
		updates["partner_type"] = *req.PartnerType
	}
	if req.CommissionRate != nil {
		if *req.CommissionRate < 0 {
			return fmt.Errorf("commission rate cannot be negative")
		}
		updates["commission_rate"] = *req.CommissionRate
	}
	if req.FixedFee != nil {
		if *req.FixedFee < 0 {
			return fmt.Errorf("fixed fee cannot be negative")
		}
		updates["fixed_fee"] = *req.FixedFee
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.ContactName != nil {
		updates["contact_name"] = *req.ContactName
	}
	if req.ContactPhone != nil {
		updates["contact_phone"] = *req.ContactPhone
	}
	if req.ContactEmail != nil {
		updates["contact_email"] = *req.ContactEmail
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	return s.repo.UpdatePartnerConfig(ctx, id, updates)
}

func (s *Service) DeletePartnerConfig(ctx context.Context, id int64) error {
	if id == 0 {
		return fmt.Errorf("partner config id is required")
	}
	return s.repo.DeletePartnerConfig(ctx, id)
}

// Teacher partner operations

func (s *Service) CreateTeacherPartner(ctx context.Context, req TeacherPartnerCreateRequest) (int64, error) {
	if req.TeacherID == 0 {
		return 0, fmt.Errorf("teacher id is required")
	}
	if req.PartnerID == 0 {
		return 0, fmt.Errorf("partner id is required")
	}
	if req.CooperationType == "" {
		req.CooperationType = CooperationTypeExclusive
	}
	if req.CommissionRate < 0 {
		return 0, fmt.Errorf("commission rate cannot be negative")
	}
	now := time.Now()
	if req.StartDate == nil {
		req.StartDate = &now
	}
	tp := &TeacherPartner{
		TeacherID:       req.TeacherID,
		PartnerID:       req.PartnerID,
		PartnerConfigID: req.PartnerConfigID,
		CooperationType: req.CooperationType,
		CommissionRate:  req.CommissionRate,
		StartDate:       req.StartDate,
		EndDate:         req.EndDate,
		Status:          CooperationStatusActive,
		Remark:          req.Remark,
	}
	if err := s.repo.CreateTeacherPartner(ctx, tp); err != nil {
		return 0, fmt.Errorf("create teacher partner: %w", err)
	}
	return tp.ID, nil
}

func (s *Service) ListTeacherPartners(ctx context.Context, query TeacherPartnerQuery) (*TeacherPartnerPageResult, error) {
	query.PageNum = normalizePage(query.PageNum)
	query.PageSize = normalizePageSize(query.PageSize)
	return s.repo.ListTeacherPartners(ctx, query)
}

func (s *Service) GetTeacherPartner(ctx context.Context, id int64) (*TeacherPartnerVO, error) {
	if id == 0 {
		return nil, fmt.Errorf("teacher partner id is required")
	}
	tp, err := s.repo.GetTeacherPartnerByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("teacher partner not found")
	}
	vo := toTeacherPartnerVO(*tp, "")
	return &vo, nil
}

func (s *Service) UpdateTeacherPartner(ctx context.Context, id int64, req TeacherPartnerUpdateRequest) error {
	if id == 0 {
		return fmt.Errorf("teacher partner id is required")
	}
	updates := map[string]any{}
	if req.CooperationType != nil {
		updates["cooperation_type"] = *req.CooperationType
	}
	if req.CommissionRate != nil {
		if *req.CommissionRate < 0 {
			return fmt.Errorf("commission rate cannot be negative")
		}
		updates["commission_rate"] = *req.CommissionRate
	}
	if req.StartDate != nil {
		updates["start_date"] = *req.StartDate
	}
	if req.EndDate != nil {
		updates["end_date"] = *req.EndDate
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Remark != nil {
		updates["remark"] = *req.Remark
	}
	return s.repo.UpdateTeacherPartner(ctx, id, updates)
}

func (s *Service) DeleteTeacherPartner(ctx context.Context, id int64) error {
	if id == 0 {
		return fmt.Errorf("teacher partner id is required")
	}
	return s.repo.DeleteTeacherPartner(ctx, id)
}

func (s *Service) ListPartneredTeachers(ctx context.Context, page, pageSize int) (*TeacherPartnerPageResult, error) {
	return s.repo.ListPartneredTeachers(ctx, normalizePage(page), normalizePageSize(pageSize))
}

func normalizePage(page int) int {
	if page <= 0 {
		return 1
	}
	return page
}

func normalizePageSize(pageSize int) int {
	if pageSize <= 0 {
		return 10
	}
	if pageSize > 100 {
		return 100
	}
	return pageSize
}
