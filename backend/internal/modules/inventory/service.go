package inventory

import (
	"context"
	"fmt"

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

func (s *Service) CreateCategory(ctx context.Context, c *GoodsCategory) error {
	if c.Name == "" {
		return fmt.Errorf("name is required")
	}
	return s.repo.CreateCategory(ctx, c)
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

func (s *Service) CreateSKU(ctx context.Context, sku *GoodsSKU) (int64, error) {
	if sku.GoodsID == 0 || sku.SKUName == "" {
		return 0, fmt.Errorf("goods_id and sku_name are required")
	}
	if err := s.repo.CreateSKU(ctx, sku); err != nil {
		return 0, err
	}
	return sku.ID, nil
}

func (s *Service) GetSKUsByGoodsID(ctx context.Context, goodsID int64) ([]*GoodsSKU, error) {
	return s.repo.GetSKUsByGoodsID(ctx, goodsID)
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
