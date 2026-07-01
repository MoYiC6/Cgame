package order

import (
	"context"
	"fmt"

	"backend/internal/platform/database"
)

type Repository interface {
	Ping(ctx context.Context) error
	CreateOrder(ctx context.Context, order *Order) error
	GetOrderByID(ctx context.Context, id int64) (*Order, error)
	GetOrderByNo(ctx context.Context, orderNo string) (*Order, error)
	ListOrders(ctx context.Context, userID int64, page, pageSize int) ([]*Order, int, error)
	UpdateOrderStatus(ctx context.Context, orderID int64, status string) error
	CreateOrderItem(ctx context.Context, item *OrderItem) error
	GetOrderItems(ctx context.Context, orderID int64) ([]*OrderItem, error)
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

func (r *repository) CreateOrder(ctx context.Context, order *Order) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	return exec.QueryRowContext(ctx,
		`INSERT INTO orders (order_no, user_id, status, total_amount, pay_amount, discount_amount, goods_id, sku_name, quantity, remark)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`,
		order.OrderNo, order.UserID, order.Status, order.TotalAmount, order.PayAmount, order.DiscountAmount, order.GoodsID, order.SKUName, order.Quantity, order.Remark,
	).Scan(&order.ID)
}

func (r *repository) GetOrderByID(ctx context.Context, id int64) (*Order, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	row := exec.QueryRowContext(ctx,
		`SELECT id, order_no, user_id, status, total_amount, pay_amount, discount_amount, goods_id, sku_name, quantity, remark, pay_at, completed_at, cancelled_at, created_at, updated_at
		 FROM orders WHERE id = $1`,
		id,
	)
	var order Order
	err := row.Scan(&order.ID, &order.OrderNo, &order.UserID, &order.Status, &order.TotalAmount, &order.PayAmount, &order.DiscountAmount, &order.GoodsID, &order.SKUName, &order.Quantity, &order.Remark, &order.PayAt, &order.CompletedAt, &order.CancelledAt, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get order: %w", err)
	}
	return &order, nil
}

func (r *repository) GetOrderByNo(ctx context.Context, orderNo string) (*Order, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	row := exec.QueryRowContext(ctx,
		`SELECT id, order_no, user_id, status, total_amount, pay_amount, discount_amount, goods_id, sku_name, quantity, remark, pay_at, completed_at, cancelled_at, created_at, updated_at
		 FROM orders WHERE order_no = $1`,
		orderNo,
	)
	var order Order
	err := row.Scan(&order.ID, &order.OrderNo, &order.UserID, &order.Status, &order.TotalAmount, &order.PayAmount, &order.DiscountAmount, &order.GoodsID, &order.SKUName, &order.Quantity, &order.Remark, &order.PayAt, &order.CompletedAt, &order.CancelledAt, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get order: %w", err)
	}
	return &order, nil
}

func (r *repository) ListOrders(ctx context.Context, userID int64, page, pageSize int) ([]*Order, int, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	var total int
	if err := exec.QueryRowContext(ctx, "SELECT COUNT(*) FROM orders WHERE user_id = $1", userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count orders: %w", err)
	}

	rows, err := exec.QueryContext(ctx,
		`SELECT id, order_no, user_id, status, total_amount, pay_amount, discount_amount, goods_id, sku_name, quantity, remark, pay_at, completed_at, cancelled_at, created_at, updated_at
		 FROM orders WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, pageSize, (page-1)*pageSize,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list orders: %w", err)
	}
	defer rows.Close()

	var orders []*Order
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.OrderNo, &o.UserID, &o.Status, &o.TotalAmount, &o.PayAmount, &o.DiscountAmount, &o.GoodsID, &o.SKUName, &o.Quantity, &o.Remark, &o.PayAt, &o.CompletedAt, &o.CancelledAt, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan order: %w", err)
		}
		orders = append(orders, &o)
	}
	return orders, total, nil
}

func (r *repository) UpdateOrderStatus(ctx context.Context, orderID int64, status string) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	_, err := exec.ExecContext(ctx,
		`UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`,
		status, orderID,
	)
	if err != nil {
		return fmt.Errorf("update order status: %w", err)
	}
	return nil
}

func (r *repository) CreateOrderItem(ctx context.Context, item *OrderItem) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	return exec.QueryRowContext(ctx,
		`INSERT INTO order_items (order_id, goods_id, sku_name, price, quantity, subtotal) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		item.OrderID, item.GoodsID, item.SKUName, item.Price, item.Quantity, item.Subtotal,
	).Scan(&item.ID)
}

func (r *repository) GetOrderItems(ctx context.Context, orderID int64) ([]*OrderItem, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	rows, err := exec.QueryContext(ctx,
		`SELECT id, order_id, goods_id, sku_name, price, quantity, subtotal, created_at FROM order_items WHERE order_id = $1`,
		orderID,
	)
	if err != nil {
		return nil, fmt.Errorf("get order items: %w", err)
	}
	defer rows.Close()

	var items []*OrderItem
	for rows.Next() {
		var item OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.GoodsID, &item.SKUName, &item.Price, &item.Quantity, &item.Subtotal, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan order item: %w", err)
		}
		items = append(items, &item)
	}
	return items, nil
}
