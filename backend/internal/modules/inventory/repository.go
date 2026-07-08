package inventory

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"backend/internal/platform/database"
)

type Repository interface {
	Ping(ctx context.Context) error

	// Categories
	ListCategories(ctx context.Context) ([]*GoodsCategory, error)
	ListAllCategories(ctx context.Context) ([]*GoodsCategory, error)
	GetCategory(ctx context.Context, id int64) (*GoodsCategory, error)
	CreateCategory(ctx context.Context, c *GoodsCategory) error
	UpdateCategory(ctx context.Context, c *GoodsCategory) error
	DeleteCategory(ctx context.Context, id int64) error

	// Goods
	CreateGoods(ctx context.Context, g *Goods) error
	GetGoods(ctx context.Context, id int64) (*Goods, error)
	ListGoods(ctx context.Context, page, pageSize int) ([]*Goods, int, error)
	UpdateGoods(ctx context.Context, g *Goods) error
	DeleteGoods(ctx context.Context, id int64) error
	UpdateGoodsStatus(ctx context.Context, id int64, status int) error
	GetGoodsStats(ctx context.Context) (*GoodsStats, error)

	// SKU
	CreateSKU(ctx context.Context, sku *GoodsSKU) error
	GetSKU(ctx context.Context, id int64) (*GoodsSKU, error)
	GetSKUsByGoodsID(ctx context.Context, goodsID int64) ([]*GoodsSKU, error)
	ListSKUs(ctx context.Context, goodsID int64, page, pageSize int) ([]*GoodsSKU, int, error)
	UpdateSKU(ctx context.Context, sku *GoodsSKU) error
	DeleteSKU(ctx context.Context, id int64) error
	DecreaseStock(ctx context.Context, skuID int64, quantity int) error
	IncreaseStock(ctx context.Context, skuID int64, quantity int) error
	CreateStockLog(ctx context.Context, log *GoodsSKUStockLog) error

	// Purchase Limit Rules
	CreatePurchaseLimitRule(ctx context.Context, rule *PurchaseLimitRule) error
	GetPurchaseLimitRuleByID(ctx context.Context, id int64) (*PurchaseLimitRule, error)
	UpdatePurchaseLimitRule(ctx context.Context, rule *PurchaseLimitRule) error
	DeletePurchaseLimitRule(ctx context.Context, id int64) error
	ListPurchaseLimitRules(ctx context.Context, page, pageSize int) ([]*PurchaseLimitRule, int, error)
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

// Categories

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

func (r *repository) ListAllCategories(ctx context.Context) ([]*GoodsCategory, error) {
	rows, err := r.dbtx.QueryContext(ctx,
		"SELECT id, parent_id, name, icon, sort, status, created_at, updated_at FROM goods_categories ORDER BY sort ASC, id ASC",
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

func (r *repository) UpdateCategory(ctx context.Context, c *GoodsCategory) error {
	_, err := r.dbtx.ExecContext(ctx,
		"UPDATE goods_categories SET parent_id = $1, name = $2, icon = $3, sort = $4, status = $5, updated_at = NOW() WHERE id = $6",
		c.ParentID, c.Name, c.Icon, c.Sort, c.Status, c.ID,
	)
	return err
}

func (r *repository) DeleteCategory(ctx context.Context, id int64) error {
	_, err := r.dbtx.ExecContext(ctx,
		"UPDATE goods_categories SET status = 0, updated_at = NOW() WHERE id = $1",
		id,
	)
	return err
}

// Goods

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

func (r *repository) UpdateGoods(ctx context.Context, g *Goods) error {
	_, err := r.dbtx.ExecContext(ctx,
		`UPDATE goods SET category_id = $1, platform = $2, name = $3, description = $4, cover_image = $5, billing_mode = $6, status = $7, is_visible = $8, commission_type = $9, commission_rate = $10, min_teacher_level = $11, map_select_enabled = $12, version = version + 1, updated_at = NOW() WHERE id = $13`,
		g.CategoryID, g.Platform, g.Name, g.Description, g.CoverImage, g.BillingMode, g.Status, g.IsVisible, g.CommissionType, g.CommissionRate, g.MinTeacherLevel, g.MapSelectEnabled, g.ID,
	)
	return err
}

func (r *repository) DeleteGoods(ctx context.Context, id int64) error {
	_, err := r.dbtx.ExecContext(ctx,
		"UPDATE goods SET status = 0, is_visible = FALSE, updated_at = NOW() WHERE id = $1",
		id,
	)
	return err
}

func (r *repository) UpdateGoodsStatus(ctx context.Context, id int64, status int) error {
	_, err := r.dbtx.ExecContext(ctx,
		"UPDATE goods SET status = $1, updated_at = NOW() WHERE id = $2",
		status, id,
	)
	return err
}

func (r *repository) GetGoodsStats(ctx context.Context) (*GoodsStats, error) {
	var stats GoodsStats
	err := r.dbtx.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM goods`,
	).Scan(&stats.TotalGoods)
	if err != nil {
		return nil, err
	}

	err = r.dbtx.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM goods_skus`,
	).Scan(&stats.TotalSKUs)
	if err != nil {
		return nil, err
	}

	err = r.dbtx.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(stock), 0) FROM goods_skus WHERE status = 1`,
	).Scan(&stats.TotalStock)
	if err != nil {
		return nil, err
	}

	err = r.dbtx.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM goods WHERE status = 1`,
	).Scan(&stats.ActiveGoods)
	if err != nil {
		return nil, err
	}

	err = r.dbtx.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM goods WHERE status = 0`,
	).Scan(&stats.InactiveGoods)
	if err != nil {
		return nil, err
	}

	return &stats, nil
}

// SKU

func (r *repository) CreateSKU(ctx context.Context, sku *GoodsSKU) error {
	snapshotJSON, _ := json.Marshal(sku.SKUSnapshot)
	return r.dbtx.QueryRowContext(ctx,
		`INSERT INTO goods_skus (goods_id, sku_name, sku_snapshot, price, stock, sort, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		sku.GoodsID, sku.SKUName, string(snapshotJSON), sku.Price, sku.Stock, sku.Sort, sku.Status,
	).Scan(&sku.ID)
}

func (r *repository) GetSKU(ctx context.Context, id int64) (*GoodsSKU, error) {
	row := r.dbtx.QueryRowContext(ctx,
		"SELECT id, goods_id, sku_name, sku_snapshot, price, stock, sort, status, version, created_at, updated_at FROM goods_skus WHERE id = $1",
		id,
	)
	var s GoodsSKU
	var snapshotJSON []byte
	err := row.Scan(&s.ID, &s.GoodsID, &s.SKUName, &snapshotJSON, &s.Price, &s.Stock, &s.Sort, &s.Status, &s.Version, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if len(snapshotJSON) > 0 {
		json.Unmarshal(snapshotJSON, &s.SKUSnapshot)
	}
	return &s, nil
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

func (r *repository) ListSKUs(ctx context.Context, goodsID int64, page, pageSize int) ([]*GoodsSKU, int, error) {
	var args []any
	var conditions []string
	if goodsID > 0 {
		conditions = append(conditions, "goods_id = $1")
		args = append(args, goodsID)
	}
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM goods_skus %s", whereClause)
	var total int
	if err := r.dbtx.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(
		`SELECT id, goods_id, sku_name, sku_snapshot, price, stock, sort, status, version, created_at, updated_at
		 FROM goods_skus %s ORDER BY sort ASC, id ASC LIMIT $%d OFFSET $%d`,
		whereClause, len(args)+1, len(args)+2,
	)
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := r.dbtx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var skus []*GoodsSKU
	for rows.Next() {
		var s GoodsSKU
		var snapshotJSON []byte
		if err := rows.Scan(&s.ID, &s.GoodsID, &s.SKUName, &snapshotJSON, &s.Price, &s.Stock, &s.Sort, &s.Status, &s.Version, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, 0, err
		}
		if len(snapshotJSON) > 0 {
			json.Unmarshal(snapshotJSON, &s.SKUSnapshot)
		}
		skus = append(skus, &s)
	}
	return skus, total, nil
}

func (r *repository) UpdateSKU(ctx context.Context, sku *GoodsSKU) error {
	snapshotJSON, _ := json.Marshal(sku.SKUSnapshot)
	_, err := r.dbtx.ExecContext(ctx,
		"UPDATE goods_skus SET sku_name = $1, sku_snapshot = $2, price = $3, stock = $4, sort = $5, status = $6, version = version + 1, updated_at = NOW() WHERE id = $7",
		sku.SKUName, string(snapshotJSON), sku.Price, sku.Stock, sku.Sort, sku.Status, sku.ID,
	)
	return err
}

func (r *repository) DeleteSKU(ctx context.Context, id int64) error {
	_, err := r.dbtx.ExecContext(ctx,
		"UPDATE goods_skus SET status = 0, updated_at = NOW() WHERE id = $1",
		id,
	)
	return err
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

// Purchase Limit Rules

func (r *repository) CreatePurchaseLimitRule(ctx context.Context, rule *PurchaseLimitRule) error {
	return r.dbtx.QueryRowContext(ctx,
		`INSERT INTO purchase_limit_rules (goods_id, limit_type, limit_count, limit_period, start_time, end_time, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		rule.GoodsID, rule.LimitType, rule.LimitCount, rule.LimitPeriod, rule.StartTime, rule.EndTime, rule.Status,
	).Scan(&rule.ID)
}

func (r *repository) GetPurchaseLimitRuleByID(ctx context.Context, id int64) (*PurchaseLimitRule, error) {
	row := r.dbtx.QueryRowContext(ctx,
		`SELECT id, goods_id, limit_type, limit_count, limit_period, start_time, end_time, status, created_at, updated_at FROM purchase_limit_rules WHERE id = $1`,
		id,
	)
	var rule PurchaseLimitRule
	err := row.Scan(&rule.ID, &rule.GoodsID, &rule.LimitType, &rule.LimitCount, &rule.LimitPeriod, &rule.StartTime, &rule.EndTime, &rule.Status, &rule.CreatedAt, &rule.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *repository) UpdatePurchaseLimitRule(ctx context.Context, rule *PurchaseLimitRule) error {
	_, err := r.dbtx.ExecContext(ctx,
		`UPDATE purchase_limit_rules SET goods_id = $1, limit_type = $2, limit_count = $3, limit_period = $4, start_time = $5, end_time = $6, status = $7, updated_at = NOW() WHERE id = $8`,
		rule.GoodsID, rule.LimitType, rule.LimitCount, rule.LimitPeriod, rule.StartTime, rule.EndTime, rule.Status, rule.ID,
	)
	return err
}

func (r *repository) DeletePurchaseLimitRule(ctx context.Context, id int64) error {
	_, err := r.dbtx.ExecContext(ctx,
		"UPDATE purchase_limit_rules SET status = 0, updated_at = NOW() WHERE id = $1",
		id,
	)
	return err
}

func (r *repository) ListPurchaseLimitRules(ctx context.Context, page, pageSize int) ([]*PurchaseLimitRule, int, error) {
	var total int
	if err := r.dbtx.QueryRowContext(ctx, "SELECT COUNT(*) FROM purchase_limit_rules WHERE status = 1").Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.dbtx.QueryContext(ctx,
		`SELECT id, goods_id, limit_type, limit_count, limit_period, start_time, end_time, status, created_at, updated_at
		 FROM purchase_limit_rules WHERE status = 1 ORDER BY id DESC LIMIT $1 OFFSET $2`,
		pageSize, (page-1)*pageSize,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var rules []*PurchaseLimitRule
	for rows.Next() {
		var rule PurchaseLimitRule
		if err := rows.Scan(&rule.ID, &rule.GoodsID, &rule.LimitType, &rule.LimitCount, &rule.LimitPeriod, &rule.StartTime, &rule.EndTime, &rule.Status, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
			return nil, 0, err
		}
		rules = append(rules, &rule)
	}
	return rules, total, nil
}
