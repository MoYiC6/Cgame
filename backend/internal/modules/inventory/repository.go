package inventory

import (
	"context"
	"encoding/json"

	"backend/internal/platform/database"
)

type Repository interface {
	Ping(ctx context.Context) error
	ListCategories(ctx context.Context) ([]*GoodsCategory, error)
	GetCategory(ctx context.Context, id int64) (*GoodsCategory, error)
	CreateCategory(ctx context.Context, c *GoodsCategory) error
	CreateGoods(ctx context.Context, g *Goods) error
	GetGoods(ctx context.Context, id int64) (*Goods, error)
	ListGoods(ctx context.Context, page, pageSize int) ([]*Goods, int, error)
	CreateSKU(ctx context.Context, sku *GoodsSKU) error
	GetSKUsByGoodsID(ctx context.Context, goodsID int64) ([]*GoodsSKU, error)
	DecreaseStock(ctx context.Context, skuID int64, quantity int) error
	IncreaseStock(ctx context.Context, skuID int64, quantity int) error
	CreateStockLog(ctx context.Context, log *GoodsSKUStockLog) error
}

type repository struct {
	dbtx database.DBTX
}

func NewRepository(dbtx database.DBTX) Repository {
	return &repository{dbtx: dbtx}
}

func (r *repository) Ping(ctx context.Context) error {
	return r.dbtx.QueryRowContext(ctx, "SELECT 1").Err()
}

func (r *repository) ListCategories(ctx context.Context) ([]*GoodsCategory, error) {
	rows, err := r.dbtx.QueryContext(ctx,
		"SELECT id, parent_id, name, icon, sort, status, created_at, updated_at FROM goods_categories WHERE status = 1 ORDER BY sort ASC, id ASC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*GoodsCategory
	for rows.Next() {
		var c GoodsCategory
		if err := rows.Scan(&c.ID, &c.ParentID, &c.Name, &c.Icon, &c.Sort, &c.Status, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, &c)
	}
	return categories, nil
}

func (r *repository) GetCategory(ctx context.Context, id int64) (*GoodsCategory, error) {
	row := r.dbtx.QueryRowContext(ctx,
		"SELECT id, parent_id, name, icon, sort, status, created_at, updated_at FROM goods_categories WHERE id = $1",
		id,
	)
	var c GoodsCategory
	err := row.Scan(&c.ID, &c.ParentID, &c.Name, &c.Icon, &c.Sort, &c.Status, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *repository) CreateCategory(ctx context.Context, c *GoodsCategory) error {
	return r.dbtx.QueryRowContext(ctx,
		"INSERT INTO goods_categories (parent_id, name, icon, sort) VALUES ($1, $2, $3, $4) RETURNING id",
		c.ParentID, c.Name, c.Icon, c.Sort,
	).Scan(&c.ID)
}

func (r *repository) CreateGoods(ctx context.Context, g *Goods) error {
	return r.dbtx.QueryRowContext(ctx,
		`INSERT INTO goods (category_id, platform, name, description, cover_image, billing_mode, status, is_visible, commission_type, commission_rate, min_teacher_level, map_select_enabled)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING id`,
		g.CategoryID, g.Platform, g.Name, g.Description, g.CoverImage, g.BillingMode, g.Status, g.IsVisible, g.CommissionType, g.CommissionRate, g.MinTeacherLevel, g.MapSelectEnabled,
	).Scan(&g.ID)
}

func (r *repository) GetGoods(ctx context.Context, id int64) (*Goods, error) {
	row := r.dbtx.QueryRowContext(ctx,
		`SELECT id, category_id, platform, name, description, cover_image, billing_mode, status, is_visible, commission_type, commission_rate, min_teacher_level, map_select_enabled, version, created_at, updated_at
		 FROM goods WHERE id = $1`,
		id,
	)
	var g Goods
	err := row.Scan(&g.ID, &g.CategoryID, &g.Platform, &g.Name, &g.Description, &g.CoverImage, &g.BillingMode, &g.Status, &g.IsVisible, &g.CommissionType, &g.CommissionRate, &g.MinTeacherLevel, &g.MapSelectEnabled, &g.Version, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *repository) ListGoods(ctx context.Context, page, pageSize int) ([]*Goods, int, error) {
	countQuery := "SELECT COUNT(*) FROM goods WHERE status = 1 AND is_visible = TRUE"
	var total int
	if err := r.dbtx.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.dbtx.QueryContext(ctx,
		`SELECT id, category_id, platform, name, description, cover_image, billing_mode, status, is_visible, commission_type, commission_rate, min_teacher_level, map_select_enabled, version, created_at, updated_at
		 FROM goods WHERE status = 1 AND is_visible = TRUE ORDER BY id DESC LIMIT $1 OFFSET $2`,
		pageSize, (page-1)*pageSize,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var goods []*Goods
	for rows.Next() {
		var g Goods
		if err := rows.Scan(&g.ID, &g.CategoryID, &g.Platform, &g.Name, &g.Description, &g.CoverImage, &g.BillingMode, &g.Status, &g.IsVisible, &g.CommissionType, &g.CommissionRate, &g.MinTeacherLevel, &g.MapSelectEnabled, &g.Version, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, 0, err
		}
		goods = append(goods, &g)
	}
	return goods, total, nil
}

func (r *repository) CreateSKU(ctx context.Context, sku *GoodsSKU) error {
	snapshotJSON, _ := json.Marshal(sku.SKUSnapshot)
	return r.dbtx.QueryRowContext(ctx,
		`INSERT INTO goods_skus (goods_id, sku_name, sku_snapshot, price, stock, sort, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		sku.GoodsID, sku.SKUName, string(snapshotJSON), sku.Price, sku.Stock, sku.Sort, sku.Status,
	).Scan(&sku.ID)
}

func (r *repository) GetSKUsByGoodsID(ctx context.Context, goodsID int64) ([]*GoodsSKU, error) {
	rows, err := r.dbtx.QueryContext(ctx,
		"SELECT id, goods_id, sku_name, sku_snapshot, price, stock, sort, status, version, created_at, updated_at FROM goods_skus WHERE goods_id = $1 AND status = 1",
		goodsID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var skus []*GoodsSKU
	for rows.Next() {
		var s GoodsSKU
		var snapshotJSON []byte
		if err := rows.Scan(&s.ID, &s.GoodsID, &s.SKUName, &snapshotJSON, &s.Price, &s.Stock, &s.Sort, &s.Status, &s.Version, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		if len(snapshotJSON) > 0 {
			json.Unmarshal(snapshotJSON, &s.SKUSnapshot)
		}
		skus = append(skus, &s)
	}
	return skus, nil
}

func (r *repository) DecreaseStock(ctx context.Context, skuID int64, quantity int) error {
	_, err := r.dbtx.ExecContext(ctx,
		"UPDATE goods_skus SET stock = stock - $1, version = version + 1 WHERE id = $2 AND stock >= $1",
		quantity, skuID,
	)
	return err
}

func (r *repository) IncreaseStock(ctx context.Context, skuID int64, quantity int) error {
	_, err := r.dbtx.ExecContext(ctx,
		"UPDATE goods_skus SET stock = stock + $1, version = version + 1 WHERE id = $2",
		quantity, skuID,
	)
	return err
}

func (r *repository) CreateStockLog(ctx context.Context, log *GoodsSKUStockLog) error {
	return r.dbtx.QueryRowContext(ctx,
		`INSERT INTO goods_sku_stock_logs (sku_id, old_stock, new_stock, change_type, order_id, operator_id) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		log.SKUID, log.OldStock, log.NewStock, log.ChangeType, log.OrderID, log.OperatorID,
	).Scan(&log.ID)
}
