package invite

import (
	"context"
	"fmt"
	"strings"
	"time"

	"backend/internal/platform/database"
)

type Service struct {
	repo      Repository
	txManager database.TxManager
}

func NewService(repo Repository, txManager database.TxManager) *Service {
	if txManager == nil {
		txManager = database.NoopTxManager{}
	}
	return &Service{repo: repo, txManager: txManager}
}

// Client invite operations

func (s *Service) GetInviteInfo(ctx context.Context, userID int64) (*InviteInfo, error) {
	if userID == 0 {
		return nil, fmt.Errorf("user id is required")
	}
	return s.repo.GetInviteInfo(ctx, userID)
}

func (s *Service) ListInviteRecords(ctx context.Context, userID int64, page, pageSize int) (*InviteRecordPageResult, error) {
	if userID == 0 {
		return nil, fmt.Errorf("user id is required")
	}
	return s.repo.ListInviteRecords(ctx, userID, normalizePage(page), normalizePageSize(pageSize))
}

func (s *Service) BindInviter(ctx context.Context, inviteeID int64, req BindRequest) error {
	if inviteeID == 0 {
		return fmt.Errorf("user id is required")
	}
	code := strings.TrimSpace(req.InviteCode)
	if code == "" {
		return fmt.Errorf("invite code is required")
	}

	// Check if user already has an inviter
	existing, err := s.repo.GetInviteRecordByInvitee(ctx, inviteeID)
	if err == nil && existing != nil {
		return fmt.Errorf("already bound to an inviter")
	}

	// Validate invite code
	inviteCode, err := s.repo.GetTeacherInviteCodeByCode(ctx, code)
	if err != nil {
		return fmt.Errorf("invalid invite code")
	}
	now := time.Now()
	if inviteCode.Status == InviteCodeStatusUsed {
		return fmt.Errorf("invite code already used")
	}
	if inviteCode.Status == InviteCodeStatusRevoked {
		return fmt.Errorf("invite code revoked")
	}
	if inviteCode.ExpireTime != nil && now.After(*inviteCode.ExpireTime) {
		return fmt.Errorf("invite code expired")
	}

	return s.txManager.WithinTx(ctx, func(ctx context.Context) error {
		if err := s.repo.CreateInviteRecord(ctx, inviteCode.CreatedBy, inviteeID, code); err != nil {
			return err
		}
		// Mark invite code as used
		updates := map[string]any{
			"status":   InviteCodeStatusUsed,
			"used_by":  inviteeID,
			"used_time": now,
		}
		return s.repo.UpdateTeacherInviteCode(ctx, inviteCode.ID, updates)
	})
}

func (s *Service) ValidateInviteCode(ctx context.Context, code string) (*InviteCodeValidationResult, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, fmt.Errorf("invite code is required")
	}
	inviteCode, err := s.repo.GetTeacherInviteCodeByCode(ctx, code)
	if err != nil {
		return &InviteCodeValidationResult{Valid: false}, nil
	}
	now := time.Now()
	status := resolveStatus(*inviteCode, now)
	if status == InviteCodeStatusUsed || status == InviteCodeStatusRevoked || status == InviteCodeStatusExpired {
		return &InviteCodeValidationResult{Valid: false}, nil
	}
	return &InviteCodeValidationResult{Valid: true, Code: code, Remark: inviteCode.Remark}, nil
}

// Teacher invite code operations

func (s *Service) GetMyTeacherInviteCode(ctx context.Context, userID int64) (*TeacherInviteCode, error) {
	if userID == 0 {
		return nil, fmt.Errorf("user id is required")
	}
	return s.repo.GetTeacherInviteCodeByUser(ctx, userID)
}

func (s *Service) GenerateTeacherInviteCode(ctx context.Context, userID int64) (*TeacherInviteCode, error) {
	if userID == 0 {
		return nil, fmt.Errorf("user id is required")
	}
	// Check if user already has an unused code
	existing, err := s.repo.GetTeacherInviteCodeByUser(ctx, userID)
	if err == nil && existing != nil {
		now := time.Now()
		status := resolveStatus(*existing, now)
		if status == InviteCodeStatusUnused {
			return existing, nil
		}
	}

	// Generate new code
	code := generateRandomInviteCode(8)
	// Ensure uniqueness
	for i := 0; i < 10; i++ {
		_, err := s.repo.GetTeacherInviteCodeByCode(ctx, code)
		if err != nil {
			break
		}
		code = generateRandomInviteCode(8)
	}

	if err := s.repo.CreateTeacherInviteCode(ctx, userID, code); err != nil {
		return nil, fmt.Errorf("create teacher invite code: %w", err)
	}
	return s.repo.GetTeacherInviteCodeByCode(ctx, code)
}

// Admin invite code operations

func (s *Service) ListTeacherInviteCodes(ctx context.Context, query InviteCodeQuery) (*TeacherInviteCodePageResult, error) {
	query.PageNum = normalizePage(query.PageNum)
	query.PageSize = normalizePageSize(query.PageSize)
	return s.repo.ListTeacherInviteCodes(ctx, query)
}

func (s *Service) GenerateAdminInviteCodes(ctx context.Context, adminUserID int64, req GenerateRequest) error {
	if adminUserID == 0 {
		return fmt.Errorf("admin user id is required")
	}
	if req.Count <= 0 || req.Count > 100 {
		return fmt.Errorf("count must be between 1 and 100")
	}
	if req.ValidDays <= 0 {
		return fmt.Errorf("valid days must be greater than 0")
	}

	now := time.Now()
	expireTime := now.AddDate(0, 0, req.ValidDays)
	codes := make([]TeacherInviteCode, 0, req.Count)

	for i := 0; i < req.Count; i++ {
		code := generateRandomInviteCode(8)
		codes = append(codes, TeacherInviteCode{
			Code:       code,
			Status:     InviteCodeStatusUnused,
			Remark:     req.Remark,
			ExpireTime: &expireTime,
			CreatedBy:  adminUserID,
			CreateTime: now,
		})
	}

	return s.repo.CreateTeacherInviteCodes(ctx, codes)
}

func (s *Service) UpdateTeacherInviteCode(ctx context.Context, id int64, remark string) error {
	if id == 0 {
		return fmt.Errorf("invite code id is required")
	}
	updates := map[string]any{}
	if remark != "" {
		updates["remark"] = strings.TrimSpace(remark)
	}
	if len(updates) == 0 {
		return nil
	}
	return s.repo.UpdateTeacherInviteCode(ctx, id, updates)
}

func (s *Service) DeleteTeacherInviteCode(ctx context.Context, id int64) error {
	if id == 0 {
		return fmt.Errorf("invite code id is required")
	}
	return s.repo.DeleteTeacherInviteCode(ctx, id)
}

func (s *Service) RevokeTeacherInviteCode(ctx context.Context, id int64, revokedBy int64) error {
	if id == 0 {
		return fmt.Errorf("invite code id is required")
	}
	if revokedBy == 0 {
		return fmt.Errorf("revoked by user id is required")
	}
	code, err := s.repo.GetTeacherInviteCodeByID(ctx, id)
	if err != nil {
		return fmt.Errorf("invite code not found")
	}
	if code.Status == InviteCodeStatusUsed {
		return fmt.Errorf("cannot revoke used invite code")
	}
	if code.Status == InviteCodeStatusRevoked {
		return fmt.Errorf("invite code already revoked")
	}
	now := time.Now()
	updates := map[string]any{
		"status":       InviteCodeStatusRevoked,
		"revoked_by":   revokedBy,
		"revoked_time": now,
	}
	return s.repo.UpdateTeacherInviteCode(ctx, id, updates)
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
