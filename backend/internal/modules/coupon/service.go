package coupon

import (
	"context"
	"fmt"
	"strings"

	"backend/internal/platform/database"
)

type Repository interface {
	ListAvailableCoupons(ctx context.Context, userID int64) ([]CouponVO, error)
	ListUserCoupons(ctx context.Context, userID int64, status *int) ([]UserCouponVO, error)
	ClaimCoupon(ctx context.Context, userID, couponID int64) (int64, error)
	ListAdminCoupons(ctx context.Context, query CouponQuery) (CouponPageResult, error)
	CreateCoupon(ctx context.Context, req CouponCreateRequest) (int64, error)
	UpdateCoupon(ctx context.Context, id int64, req CouponUpdateRequest) error
	DeleteCoupon(ctx context.Context, id int64) error
	GetStats(ctx context.Context) (*CouponStats, error)
}

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

func (s *Service) ListAvailable(ctx context.Context, userID int64) ([]CouponVO, error) {
	if userID == 0 {
		return nil, fmt.Errorf("user id is required")
	}
	return s.repo.ListAvailableCoupons(ctx, userID)
}

func (s *Service) ListMine(ctx context.Context, userID int64, status *int) ([]UserCouponVO, error) {
	if userID == 0 {
		return nil, fmt.Errorf("user id is required")
	}
	if status != nil && (*status < UserCouponStatusAvailable || *status > UserCouponStatusExpired) {
		return nil, fmt.Errorf("invalid coupon status")
	}
	return s.repo.ListUserCoupons(ctx, userID, status)
}

func (s *Service) Claim(ctx context.Context, userID, couponID int64) (int64, error) {
	if userID == 0 || couponID == 0 {
		return 0, fmt.Errorf("user id and coupon id are required")
	}
	var userCouponID int64
	err := s.txManager.WithinTx(ctx, func(ctx context.Context) error {
		id, err := s.repo.ClaimCoupon(ctx, userID, couponID)
		if err != nil {
			return err
		}
		userCouponID = id
		return nil
	})
	return userCouponID, err
}

func (s *Service) ListAdmin(ctx context.Context, query CouponQuery) (CouponPageResult, error) {
	query.PageNum = normalizePage(query.PageNum)
	query.PageSize = normalizePageSize(query.PageSize)
	query.Name = strings.TrimSpace(query.Name)
	return s.repo.ListAdminCoupons(ctx, query)
}

func (s *Service) Create(ctx context.Context, req CouponCreateRequest) (int64, error) {
	if err := validateCreateRequest(&req); err != nil {
		return 0, err
	}
	return s.repo.CreateCoupon(ctx, req)
}

func (s *Service) Update(ctx context.Context, id int64, req CouponUpdateRequest) error {
	if id == 0 {
		return fmt.Errorf("coupon id is required")
	}
	if err := validateUpdateRequest(&req); err != nil {
		return err
	}
	return s.repo.UpdateCoupon(ctx, id, req)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id == 0 {
		return fmt.Errorf("coupon id is required")
	}
	return s.repo.DeleteCoupon(ctx, id)
}

func (s *Service) Stats(ctx context.Context) (*CouponStats, error) {
	return s.repo.GetStats(ctx)
}

func validateCreateRequest(req *CouponCreateRequest) error {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return fmt.Errorf("coupon name is required")
	}
	if !validCouponType(req.Type) {
		return fmt.Errorf("invalid coupon type")
	}
	if req.FaceValue <= 0 {
		return fmt.Errorf("face value must be greater than 0")
	}
	if req.MinOrderAmount < 0 {
		return fmt.Errorf("min order amount cannot be negative")
	}
	if req.TotalQuantity <= 0 {
		return fmt.Errorf("total quantity must be greater than 0")
	}
	if req.PerUserLimit <= 0 {
		return fmt.Errorf("per user limit must be greater than 0")
	}
	if req.ValidDays <= 0 {
		return fmt.Errorf("valid days must be greater than 0")
	}
	if req.StartTime == nil || req.EndTime == nil {
		return fmt.Errorf("start time and end time are required")
	}
	if !req.DistributionModeSet {
		req.DistributionMode = CouponDistributionPublic
	}
	if !validDistributionMode(req.DistributionMode) {
		return fmt.Errorf("invalid distribution mode")
	}
	defaultEnabled := true
	if req.Enabled == nil {
		req.Enabled = &defaultEnabled
	}
	defaultPermanent := false
	if req.IsPermanent == nil {
		req.IsPermanent = &defaultPermanent
	}
	if strings.TrimSpace(req.ApplicableScope) == "" {
		req.ApplicableScope = `["all"]`
	}
	return nil
}

func validateUpdateRequest(req *CouponUpdateRequest) error {
	if req.Name != nil {
		value := strings.TrimSpace(*req.Name)
		if value == "" {
			return fmt.Errorf("coupon name cannot be empty")
		}
		req.Name = &value
	}
	if req.Type != nil && !validCouponType(*req.Type) {
		return fmt.Errorf("invalid coupon type")
	}
	if req.FaceValue != nil && *req.FaceValue <= 0 {
		return fmt.Errorf("face value must be greater than 0")
	}
	if req.MinOrderAmount != nil && *req.MinOrderAmount < 0 {
		return fmt.Errorf("min order amount cannot be negative")
	}
	if req.TotalQuantity != nil && *req.TotalQuantity <= 0 {
		return fmt.Errorf("total quantity must be greater than 0")
	}
	if req.PerUserLimit != nil && *req.PerUserLimit <= 0 {
		return fmt.Errorf("per user limit must be greater than 0")
	}
	if req.ValidDays != nil && *req.ValidDays <= 0 {
		return fmt.Errorf("valid days must be greater than 0")
	}
	if req.DistributionMode != nil && !validDistributionMode(*req.DistributionMode) {
		return fmt.Errorf("invalid distribution mode")
	}
	return nil
}

func validCouponType(couponType int) bool {
	return couponType >= CouponTypeFixed && couponType <= CouponTypeExchange
}

func validDistributionMode(mode int) bool {
	return mode >= CouponDistributionPrivate && mode <= CouponDistributionPublic
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
	return pageSize
}
