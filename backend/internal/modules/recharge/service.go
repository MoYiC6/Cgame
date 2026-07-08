package recharge

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

// Client recharge operations

func (s *Service) CreateRecharge(ctx context.Context, userID int64, req CreateRechargeRequest) (*RechargeRecord, error) {
	if userID == 0 {
		return nil, fmt.Errorf("user id is required")
	}
	if req.Amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than 0")
	}

	// Calculate gift amount based on enabled rebate rules
	giftAmount := s.calculateGiftAmount(ctx, req.Amount)
	totalAmount := req.Amount + giftAmount

	record := &RechargeRecord{
		UserID:      userID,
		Amount:      req.Amount,
		GiftAmount:  giftAmount,
		TotalAmount: totalAmount,
		PayAmount:   0,
		Status:      RechargeStatusPending,
	}
	if err := s.repo.CreateRechargeRecord(ctx, record); err != nil {
		return nil, fmt.Errorf("create recharge record: %w", err)
	}
	return record, nil
}

func (s *Service) ManualRecharge(ctx context.Context, adminUserID int64, req ManualRechargeRequest) (int64, error) {
	if adminUserID == 0 {
		return 0, fmt.Errorf("admin user id is required")
	}
	if req.UserID == 0 {
		return 0, fmt.Errorf("user id is required")
	}
	if req.Amount <= 0 {
		return 0, fmt.Errorf("amount must be greater than 0")
	}

	giftAmount := s.calculateGiftAmount(ctx, req.Amount)
	totalAmount := req.Amount + giftAmount

	record := &RechargeRecord{
		UserID:      req.UserID,
		Amount:      req.Amount,
		GiftAmount:  giftAmount,
		TotalAmount: totalAmount,
		PayAmount:   req.Amount,
		Status:      RechargeStatusPaid,
		Remark:      strings.TrimSpace(req.Remark),
	}
	if err := s.repo.CreateRechargeRecord(ctx, record); err != nil {
		return 0, fmt.Errorf("create recharge record: %w", err)
	}
	return record.ID, nil
}

func (s *Service) ListMine(ctx context.Context, userID int64, page, pageSize int) (*RechargeRecordPageResult, error) {
	if userID == 0 {
		return nil, fmt.Errorf("user id is required")
	}
	return s.repo.ListUserRecharges(ctx, userID, normalizePage(page), normalizePageSize(pageSize))
}

func (s *Service) GetMine(ctx context.Context, userID, id int64) (*RechargeRecordVO, error) {
	if userID == 0 || id == 0 {
		return nil, fmt.Errorf("user id and recharge id are required")
	}
	record, err := s.repo.GetRechargeByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("recharge record not found")
	}
	if record.UserID != userID {
		return nil, fmt.Errorf("recharge record does not belong to user")
	}
	vo := toRechargeRecordVO(*record)
	return &vo, nil
}

func (s *Service) GetDetailByNo(ctx context.Context, userID int64, rechargeNo string) (*RechargeRecordVO, error) {
	if rechargeNo == "" {
		return nil, fmt.Errorf("recharge no is required")
	}
	record, err := s.repo.GetRechargeByNo(ctx, rechargeNo)
	if err != nil {
		return nil, fmt.Errorf("recharge record not found")
	}
	if userID != 0 && record.UserID != userID {
		return nil, fmt.Errorf("recharge record does not belong to user")
	}
	vo := toRechargeRecordVO(*record)
	return &vo, nil
}

func (s *Service) Cancel(ctx context.Context, userID int64, rechargeNo string) error {
	if userID == 0 || rechargeNo == "" {
		return fmt.Errorf("user id and recharge no are required")
	}
	record, err := s.repo.GetRechargeByNo(ctx, rechargeNo)
	if err != nil {
		return fmt.Errorf("recharge record not found")
	}
	if record.UserID != userID {
		return fmt.Errorf("recharge record does not belong to user")
	}
	if record.Status != RechargeStatusPending {
		return fmt.Errorf("only pending recharges can be cancelled")
	}
	return s.repo.CancelRecharge(ctx, rechargeNo)
}

func (s *Service) ContinuePay(ctx context.Context, userID int64, rechargeNo string) (*RechargeRecordVO, error) {
	if userID == 0 || rechargeNo == "" {
		return nil, fmt.Errorf("user id and recharge no are required")
	}
	record, err := s.repo.GetRechargeByNo(ctx, rechargeNo)
	if err != nil {
		return nil, fmt.Errorf("recharge record not found")
	}
	if record.UserID != userID {
		return nil, fmt.Errorf("recharge record does not belong to user")
	}
	if record.Status != RechargeStatusPending {
		return nil, fmt.Errorf("only pending recharges can be continued")
	}
	vo := toRechargeRecordVO(*record)
	return &vo, nil
}

func (s *Service) VerifyPayment(ctx context.Context, userID int64, rechargeNo string) (*RechargeRecordVO, error) {
	if userID == 0 || rechargeNo == "" {
		return nil, fmt.Errorf("user id and recharge no are required")
	}
	record, err := s.repo.GetRechargeByNo(ctx, rechargeNo)
	if err != nil {
		return nil, fmt.Errorf("recharge record not found")
	}
	if record.UserID != userID {
		return nil, fmt.Errorf("recharge record does not belong to user")
	}
	vo := toRechargeRecordVO(*record)
	return &vo, nil
}

func (s *Service) GetRecent(ctx context.Context, userID int64, limit int) ([]RechargeRecordVO, error) {
	if userID == 0 {
		return nil, fmt.Errorf("user id is required")
	}
	if limit <= 0 {
		limit = 5
	}
	if limit > 20 {
		limit = 20
	}
	return s.repo.GetRecentRecharges(ctx, userID, limit)
}

// Admin recharge operations

func (s *Service) ListAdmin(ctx context.Context, query RechargeQuery) (*RechargeRecordPageResult, error) {
	query.PageNum = normalizePage(query.PageNum)
	query.PageSize = normalizePageSize(query.PageSize)
	return s.repo.ListRecharges(ctx, query)
}

func (s *Service) GetAdmin(ctx context.Context, id int64) (*RechargeRecordVO, error) {
	if id == 0 {
		return nil, fmt.Errorf("recharge id is required")
	}
	record, err := s.repo.GetRechargeByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("recharge record not found")
	}
	vo := toRechargeRecordVO(*record)
	return &vo, nil
}

func (s *Service) Stats(ctx context.Context) (*RechargeStats, error) {
	return s.repo.GetStats(ctx)
}

// Callback for payment gateway
func (s *Service) Callback(ctx context.Context, rechargeNo string, payChannel string, payAmount float64) error {
	if rechargeNo == "" {
		return fmt.Errorf("recharge no is required")
	}
	record, err := s.repo.GetRechargeByNo(ctx, rechargeNo)
	if err != nil {
		return fmt.Errorf("recharge record not found")
	}
	if record.Status != RechargeStatusPending {
		return fmt.Errorf("recharge is not pending")
	}
	now := s.now()
	return s.repo.UpdateStatus(ctx, rechargeNo, RechargeStatusPaid, payChannel, &now, &now)
}

// Rebate rule operations

func (s *Service) ListAvailableRebateRules(ctx context.Context) ([]RechargeRebateRuleVO, error) {
	rules, err := s.repo.ListEnabledRebateRules(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]RechargeRebateRuleVO, 0, len(rules))
	for _, rule := range rules {
		result = append(result, toRebateRuleVO(rule))
	}
	return result, nil
}

func (s *Service) PreviewRebate(ctx context.Context, amount float64) (*RebatePreviewResult, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than 0")
	}
	giftAmount := s.calculateGiftAmount(ctx, amount)
	var ruleName string
	rules, _ := s.repo.ListEnabledRebateRules(ctx)
	for _, rule := range rules {
		if amount >= rule.MinAmount {
			ruleName = rule.Name
			break
		}
	}
	return &RebatePreviewResult{
		Amount:      amount,
		GiftAmount:  giftAmount,
		TotalAmount: amount + giftAmount,
		RuleName:    ruleName,
	}, nil
}

func (s *Service) ListRebateRules(ctx context.Context, page, pageSize int) (*RebateRulePageResult, error) {
	return s.repo.ListRebateRules(ctx, normalizePage(page), normalizePageSize(pageSize))
}

func (s *Service) CreateRebateRule(ctx context.Context, req RebateRuleCreateRequest) (int64, error) {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return 0, fmt.Errorf("rule name is required")
	}
	if req.MinAmount < 0 {
		return 0, fmt.Errorf("min amount cannot be negative")
	}
	if req.GiftRate < 0 {
		return 0, fmt.Errorf("gift rate cannot be negative")
	}
	if req.GiftAmount < 0 {
		return 0, fmt.Errorf("gift amount cannot be negative")
	}
	rule := &RechargeRebateRule{
		Name:       req.Name,
		MinAmount:  req.MinAmount,
		GiftRate:   req.GiftRate,
		GiftAmount: req.GiftAmount,
		Enabled:    req.Enabled,
		Priority:   req.Priority,
	}
	if err := s.repo.CreateRebateRule(ctx, rule); err != nil {
		return 0, fmt.Errorf("create rebate rule: %w", err)
	}
	return rule.ID, nil
}

func (s *Service) UpdateRebateRule(ctx context.Context, id int64, req RebateRuleUpdateRequest) error {
	if id == 0 {
		return fmt.Errorf("rule id is required")
	}
	updates := map[string]any{}
	if req.Name != nil {
		value := strings.TrimSpace(*req.Name)
		if value == "" {
			return fmt.Errorf("rule name cannot be empty")
		}
		updates["name"] = value
	}
	if req.MinAmount != nil {
		if *req.MinAmount < 0 {
			return fmt.Errorf("min amount cannot be negative")
		}
		updates["min_amount"] = *req.MinAmount
	}
	if req.GiftRate != nil {
		if *req.GiftRate < 0 {
			return fmt.Errorf("gift rate cannot be negative")
		}
		updates["gift_rate"] = *req.GiftRate
	}
	if req.GiftAmount != nil {
		if *req.GiftAmount < 0 {
			return fmt.Errorf("gift amount cannot be negative")
		}
		updates["gift_amount"] = *req.GiftAmount
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}
	return s.repo.UpdateRebateRule(ctx, id, updates)
}

func (s *Service) DeleteRebateRule(ctx context.Context, id int64) error {
	if id == 0 {
		return fmt.Errorf("rule id is required")
	}
	return s.repo.DeleteRebateRule(ctx, id)
}

func (s *Service) calculateGiftAmount(ctx context.Context, amount float64) float64 {
	rules, err := s.repo.ListEnabledRebateRules(ctx)
	if err != nil {
		return 0
	}
	for _, rule := range rules {
		if amount >= rule.MinAmount {
			gift := rule.GiftAmount
			if rule.GiftRate > 0 {
				gift += amount * rule.GiftRate / 100
			}
			return gift
		}
	}
	return 0
}

func (s *Service) now() time.Time {
	return time.Now()
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
