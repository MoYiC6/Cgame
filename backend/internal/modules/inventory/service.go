package inventory

import (
	"context"
	"fmt"
	"net/http"

	apperrors "backend/internal/platform/errors"
	"backend/internal/platform/database"
)

type Service struct {
	repo      Repository
	txManager database.TxManager
}

func NewService(repo Repository, txManager database.TxManager) *Service {
	s := &Service{repo: repo}
	if txManager != nil {
		s.txManager = txManager
	} else {
		s.txManager = database.NoopTxManager{}
	}
	return s
}

func (s *Service) ListCategories(ctx context.Context) ([]*GoodsCategory, error) {
	return s.repo.ListCategories(ctx)
}

func (s *Service) ListAllCategories(ctx context.Context) ([]*GoodsCategory, error) {
	return s.repo.ListAllCategories(ctx)
}

func (s *Service) GetCategory(ctx context.Context, id int64) (*GoodsCategory, error) {
	return s.repo.GetCategory(ctx, id)
}

func (s *Service) CreateCategory(ctx context.Context, c *GoodsCategory) error {
	if c.Name == "" {
		return fmt.Errorf("name is required")
	}
	return s.repo.CreateCategory(ctx, c)
}

func (s *Service) UpdateCategory(ctx context.Context, c *GoodsCategory) error {
	if c.ID == 0 {
		return apperrors.New(apperrors.CodeInvalidArgument, "category id is required", http.StatusBadRequest, nil)
	}
	if c.Name == "" {
		return apperrors.New(apperrors.CodeInvalidArgument, "name is required", http.StatusBadRequest, nil)
	}
	return s.repo.UpdateCategory(ctx, c)
}

func (s *Service) DeleteCategory(ctx context.Context, id int64) error {
	if id == 0 {
		return apperrors.New(apperrors.CodeInvalidArgument, "category id is required", http.StatusBadRequest, nil)
	}
	return s.repo.DeleteCategory(ctx, id)
}

func (s *Service) CreateGoods(ctx context.Context, g *Goods) (int64, error) {
	if g.Name == "" {
		return 0, fmt.Errorf("name is required")
	}
	if err := s.repo.CreateGoods(ctx, g); err != nil {
		return 0, err
	}
	return g.ID, nil
}

func (s *Service) GetGoods(ctx context.Context, id int64) (*Goods, error) {
	return s.repo.GetGoods(ctx, id)
}

func (s *Service) ListGoods(ctx context.Context, page, pageSize int) ([]*Goods, int, error) {
	return s.repo.ListGoods(ctx, page, pageSize)
}

func (s *Service) UpdateGoods(ctx context.Context, g *Goods) error {
	if g.ID == 0 {
		return apperrors.New(apperrors.CodeInvalidArgument, "goods id is required", http.StatusBadRequest, nil)
	}
	if g.Name == "" {
		return apperrors.New(apperrors.CodeInvalidArgument, "name is required", http.StatusBadRequest, nil)
	}
	return s.repo.UpdateGoods(ctx, g)
}

func (s *Service) DeleteGoods(ctx context.Context, id int64) error {
	if id == 0 {
		return apperrors.New(apperrors.CodeInvalidArgument, "goods id is required", http.StatusBadRequest, nil)
	}
	return s.repo.DeleteGoods(ctx, id)
}

func (s *Service) UpdateGoodsStatus(ctx context.Context, id int64, status int) error {
	if id == 0 {
		return apperrors.New(apperrors.CodeInvalidArgument, "goods id is required", http.StatusBadRequest, nil)
	}
	return s.repo.UpdateGoodsStatus(ctx, id, status)
}

func (s *Service) GetGoodsStats(ctx context.Context) (*GoodsStats, error) {
	return s.repo.GetGoodsStats(ctx)
}

func (s *Service) CreateSKU(ctx context.Context, sku *GoodsSKU) (int64, error) {
	if sku.GoodsID == 0 || sku.SKUName == "" {
		return 0, fmt.Errorf("goods_id and sku_name are required")
	}
	if err := s.repo.CreateSKU(ctx, sku); err != nil {
		return 0, err
	}
	return sku.ID, nil
}

func (s *Service) GetSKU(ctx context.Context, id int64) (*GoodsSKU, error) {
	return s.repo.GetSKU(ctx, id)
}

func (s *Service) GetSKUsByGoodsID(ctx context.Context, goodsID int64) ([]*GoodsSKU, error) {
	return s.repo.GetSKUsByGoodsID(ctx, goodsID)
}

func (s *Service) ListSKUs(ctx context.Context, goodsID int64, page, pageSize int) ([]*GoodsSKU, int, error) {
	return s.repo.ListSKUs(ctx, goodsID, page, pageSize)
}

func (s *Service) UpdateSKU(ctx context.Context, sku *GoodsSKU) error {
	if sku.ID == 0 {
		return apperrors.New(apperrors.CodeInvalidArgument, "sku id is required", http.StatusBadRequest, nil)
	}
	if sku.SKUName == "" {
		return apperrors.New(apperrors.CodeInvalidArgument, "sku_name is required", http.StatusBadRequest, nil)
	}
	return s.repo.UpdateSKU(ctx, sku)
}

func (s *Service) DeleteSKU(ctx context.Context, id int64) error {
	if id == 0 {
		return apperrors.New(apperrors.CodeInvalidArgument, "sku id is required", http.StatusBadRequest, nil)
	}
	return s.repo.DeleteSKU(ctx, id)
}

func (s *Service) DecreaseStock(ctx context.Context, skuID int64, quantity int) error {
	if quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}
	return s.repo.DecreaseStock(ctx, skuID, quantity)
}

func (s *Service) IncreaseStock(ctx context.Context, skuID int64, quantity int) error {
	if quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}
	return s.repo.IncreaseStock(ctx, skuID, quantity)
}

func (s *Service) CreateStockLog(ctx context.Context, log *GoodsSKUStockLog) error {
	if log.SKUID == 0 {
		return fmt.Errorf("sku_id is required")
	}
	return s.repo.CreateStockLog(ctx, log)
}

func (s *Service) CreatePurchaseLimitRule(ctx context.Context, rule *PurchaseLimitRule) error {
	if rule.LimitType == "" {
		return apperrors.New(apperrors.CodeInvalidArgument, "limit_type is required", http.StatusBadRequest, nil)
	}
	if rule.LimitCount <= 0 {
		return apperrors.New(apperrors.CodeInvalidArgument, "limit_count must be positive", http.StatusBadRequest, nil)
	}
	return s.repo.CreatePurchaseLimitRule(ctx, rule)
}

func (s *Service) GetPurchaseLimitRuleByID(ctx context.Context, id int64) (*PurchaseLimitRule, error) {
	return s.repo.GetPurchaseLimitRuleByID(ctx, id)
}

func (s *Service) UpdatePurchaseLimitRule(ctx context.Context, rule *PurchaseLimitRule) error {
	if rule.ID == 0 {
		return apperrors.New(apperrors.CodeInvalidArgument, "rule id is required", http.StatusBadRequest, nil)
	}
	if rule.LimitType == "" {
		return apperrors.New(apperrors.CodeInvalidArgument, "limit_type is required", http.StatusBadRequest, nil)
	}
	if rule.LimitCount <= 0 {
		return apperrors.New(apperrors.CodeInvalidArgument, "limit_count must be positive", http.StatusBadRequest, nil)
	}
	return s.repo.UpdatePurchaseLimitRule(ctx, rule)
}

func (s *Service) DeletePurchaseLimitRule(ctx context.Context, id int64) error {
	if id == 0 {
		return apperrors.New(apperrors.CodeInvalidArgument, "rule id is required", http.StatusBadRequest, nil)
	}
	return s.repo.DeletePurchaseLimitRule(ctx, id)
}

func (s *Service) ListPurchaseLimitRules(ctx context.Context, page, pageSize int) ([]*PurchaseLimitRule, int, error) {
	return s.repo.ListPurchaseLimitRules(ctx, page, pageSize)
}
