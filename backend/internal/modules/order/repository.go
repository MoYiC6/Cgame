package order

import (
	"context"
	"fmt"
	"time"

	"backend/internal/platform/database"
)

type Repository interface {
	Ping(ctx context.Context) error
	CreateOrder(ctx context.Context, order *Order) error
	GetOrderByID(ctx context.Context, id int64) (*Order, error)
	GetOrderByNo(ctx context.Context, orderNo string) (*Order, error)
	ListOrders(ctx context.Context, userID int64, page, pageSize int) ([]*Order, int, error)
	UpdateOrderStatus(ctx context.Context, orderID int64, status string) error
	UpdateOrderRemark(ctx context.Context, orderID int64, remark string) error
	UpdateOrderTeachers(ctx context.Context, orderID int64, teacherIDs []int64) error
	CreateOrderItem(ctx context.Context, item *OrderItem) error
	GetOrderItems(ctx context.Context, orderID int64) ([]*OrderItem, error)

	// Admin
	AdminListOrders(ctx context.Context, query OrderQuery) ([]*Order, int, error)
	AdminGetOrderByID(ctx context.Context, id int64) (*Order, error)

	// Review
	CreateReview(ctx context.Context, review *OrderReview) error
	GetReviewByOrderID(ctx context.Context, orderID int64) (*OrderReview, error)
	ListReviews(ctx context.Context, query ReviewQuery) ([]*OrderReview, int, error)
	UpdateReviewStatus(ctx context.Context, reviewID int64, status string) error
	UpdateReviewReply(ctx context.Context, reviewID int64, reply string) error

	// Complaint
	CreateComplaint(ctx context.Context, complaint *OrderComplaint) error
	GetComplaintByOrderID(ctx context.Context, orderID int64) (*OrderComplaint, error)

	// Transfer config
	GetTransferConfig(ctx context.Context) (*OrderTransferConfig, error)

	// Stats
	GetOrderStats(ctx context.Context, start, end *string) (*OrderStats, error)

	// Final review
	ListFinalReviewOrders(ctx context.Context, query FinalReviewQuery) ([]*Order, int, error)

	// Payment
	CreatePaymentRecord(ctx context.Context, record *PaymentRecord) error
	GetPaymentByID(ctx context.Context, id int64) (*PaymentRecord, error)
	GetPaymentByOutTradeNo(ctx context.Context, outTradeNo string) (*PaymentRecord, error)
	ListPayments(ctx context.Context, query PaymentQuery) ([]*PaymentRecord, int, error)
	UpdatePaymentStatus(ctx context.Context, id int64, status string) error
	UpdatePaymentPaidAt(ctx context.Context, id int64, paidAt *time.Time, transactionID *string) error

	// Cashier
	CreateCashierOrder(ctx context.Context, co *CashierOrder) error
	GetCashierOrderByToken(ctx context.Context, token string) (*CashierOrder, error)
	UpdateCashierOrderStatus(ctx context.Context, id int64, status string, payChannel *string) error

	// Payment sync log
	CreatePaymentSyncLog(ctx context.Context, log *PaymentSyncLog) error
	ListPaymentSyncLogs(ctx context.Context, recordID int64) ([]*PaymentSyncLog, error)
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

func (r *repository) UpdateOrderRemark(ctx context.Context, orderID int64, remark string) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	_, err := exec.ExecContext(ctx,
		`UPDATE orders SET remark = $1, updated_at = NOW() WHERE id = $2`,
		remark, orderID,
	)
	if err != nil {
		return fmt.Errorf("update order remark: %w", err)
	}
	return nil
}

func (r *repository) UpdateOrderTeachers(ctx context.Context, orderID int64, teacherIDs []int64) error {
	_ = orderID
	_ = teacherIDs
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

func (r *repository) AdminListOrders(ctx context.Context, query OrderQuery) ([]*Order, int, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	where := "WHERE 1=1"
	args := []any{}
	argIdx := 1

	if query.OrderNo != "" {
		where += fmt.Sprintf(" AND order_no = $%d", argIdx)
		args = append(args, query.OrderNo)
		argIdx++
	}
	if query.UserID != nil {
		where += fmt.Sprintf(" AND user_id = $%d", argIdx)
		args = append(args, *query.UserID)
		argIdx++
	}
	if query.Status != "" {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, query.Status)
		argIdx++
	}
	if query.StartTime != nil {
		where += fmt.Sprintf(" AND created_at >= $%d", argIdx)
		args = append(args, *query.StartTime)
		argIdx++
	}
	if query.EndTime != nil {
		where += fmt.Sprintf(" AND created_at <= $%d", argIdx)
		args = append(args, *query.EndTime)
		argIdx++
	}

	countSQL := "SELECT COUNT(*) FROM orders " + where
	var total int
	if err := exec.QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count admin orders: %w", err)
	}

	listSQL := fmt.Sprintf(
		`SELECT id, order_no, user_id, status, total_amount, pay_amount, discount_amount, goods_id, sku_name, quantity, remark, pay_at, completed_at, cancelled_at, created_at, updated_at
		 FROM orders %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		where, argIdx, argIdx+1,
	)
	args = append(args, query.PageSize, (query.PageNum-1)*query.PageSize)

	rows, err := exec.QueryContext(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list admin orders: %w", err)
	}
	defer rows.Close()

	var orders []*Order
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.OrderNo, &o.UserID, &o.Status, &o.TotalAmount, &o.PayAmount, &o.DiscountAmount, &o.GoodsID, &o.SKUName, &o.Quantity, &o.Remark, &o.PayAt, &o.CompletedAt, &o.CancelledAt, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan admin order: %w", err)
		}
		orders = append(orders, &o)
	}
	return orders, total, nil
}

func (r *repository) AdminGetOrderByID(ctx context.Context, id int64) (*Order, error) {
	return r.GetOrderByID(ctx, id)
}

func (r *repository) CreateReview(ctx context.Context, review *OrderReview) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	return exec.QueryRowContext(ctx,
		`INSERT INTO order_reviews (order_id, user_id, teacher_id, rating, content, status)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		review.OrderID, review.UserID, review.TeacherID, review.Rating, review.Content, review.Status,
	).Scan(&review.ID)
}

func (r *repository) GetReviewByOrderID(ctx context.Context, orderID int64) (*OrderReview, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	row := exec.QueryRowContext(ctx,
		`SELECT id, order_id, user_id, teacher_id, rating, content, reply, status, created_at, updated_at
		 FROM order_reviews WHERE order_id = $1`,
		orderID,
	)
	var review OrderReview
	err := row.Scan(&review.ID, &review.OrderID, &review.UserID, &review.TeacherID, &review.Rating, &review.Content, &review.Reply, &review.Status, &review.CreatedAt, &review.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get review: %w", err)
	}
	return &review, nil
}

func (r *repository) ListReviews(ctx context.Context, query ReviewQuery) ([]*OrderReview, int, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	where := "WHERE 1=1"
	args := []any{}
	argIdx := 1

	if query.Status != "" {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, query.Status)
		argIdx++
	}
	if query.OrderID != nil {
		where += fmt.Sprintf(" AND order_id = $%d", argIdx)
		args = append(args, *query.OrderID)
		argIdx++
	}

	countSQL := "SELECT COUNT(*) FROM order_reviews " + where
	var total int
	if err := exec.QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count reviews: %w", err)
	}

	listSQL := fmt.Sprintf(
		`SELECT id, order_id, user_id, teacher_id, rating, content, reply, status, created_at, updated_at
		 FROM order_reviews %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		where, argIdx, argIdx+1,
	)
	args = append(args, query.PageSize, (query.PageNum-1)*query.PageSize)

	rows, err := exec.QueryContext(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list reviews: %w", err)
	}
	defer rows.Close()

	var reviews []*OrderReview
	for rows.Next() {
		var rv OrderReview
		if err := rows.Scan(&rv.ID, &rv.OrderID, &rv.UserID, &rv.TeacherID, &rv.Rating, &rv.Content, &rv.Reply, &rv.Status, &rv.CreatedAt, &rv.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan review: %w", err)
		}
		reviews = append(reviews, &rv)
	}
	return reviews, total, nil
}

func (r *repository) UpdateReviewStatus(ctx context.Context, reviewID int64, status string) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	_, err := exec.ExecContext(ctx,
		`UPDATE order_reviews SET status = $1, updated_at = NOW() WHERE id = $2`,
		status, reviewID,
	)
	if err != nil {
		return fmt.Errorf("update review status: %w", err)
	}
	return nil
}

func (r *repository) UpdateReviewReply(ctx context.Context, reviewID int64, reply string) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	_, err := exec.ExecContext(ctx,
		`UPDATE order_reviews SET reply = $1, updated_at = NOW() WHERE id = $2`,
		reply, reviewID,
	)
	if err != nil {
		return fmt.Errorf("update review reply: %w", err)
	}
	return nil
}

func (r *repository) CreateComplaint(ctx context.Context, complaint *OrderComplaint) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	return exec.QueryRowContext(ctx,
		`INSERT INTO order_complaints (order_id, user_id, reason, detail, status)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		complaint.OrderID, complaint.UserID, complaint.Reason, complaint.Detail, complaint.Status,
	).Scan(&complaint.ID)
}

func (r *repository) GetComplaintByOrderID(ctx context.Context, orderID int64) (*OrderComplaint, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	row := exec.QueryRowContext(ctx,
		`SELECT id, order_id, user_id, reason, detail, status, created_at, updated_at
		 FROM order_complaints WHERE order_id = $1`,
		orderID,
	)
	var c OrderComplaint
	err := row.Scan(&c.ID, &c.OrderID, &c.UserID, &c.Reason, &c.Detail, &c.Status, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get complaint: %w", err)
	}
	return &c, nil
}

func (r *repository) GetTransferConfig(ctx context.Context) (*OrderTransferConfig, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	row := exec.QueryRowContext(ctx,
		`SELECT id, enabled, max_times, timeout, created_at, updated_at
		 FROM order_transfer_configs LIMIT 1`,
	)
	var cfg OrderTransferConfig
	err := row.Scan(&cfg.ID, &cfg.Enabled, &cfg.MaxTimes, &cfg.Timeout, &cfg.CreatedAt, &cfg.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get transfer config: %w", err)
	}
	return &cfg, nil
}

func (r *repository) GetOrderStats(ctx context.Context, start, end *string) (*OrderStats, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	where := "WHERE 1=1"
	args := []any{}
	argIdx := 1
	if start != nil {
		where += fmt.Sprintf(" AND created_at >= $%d", argIdx)
		args = append(args, *start)
		argIdx++
	}
	if end != nil {
		where += fmt.Sprintf(" AND created_at <= $%d", argIdx)
		args = append(args, *end)
		argIdx++
	}

	var stats OrderStats
	countSQL := "SELECT COUNT(*), COALESCE(SUM(total_amount),0) FROM orders " + where
	if err := exec.QueryRowContext(ctx, countSQL, args...).Scan(&stats.TotalOrders, &stats.TotalAmount); err != nil {
		return nil, fmt.Errorf("get total stats: %w", err)
	}

	paidSQL := "SELECT COUNT(*), COALESCE(SUM(total_amount),0) FROM orders " + where + fmt.Sprintf(" AND status = $%d", argIdx)
	paidArgs := append(args, OrderStatusPaid)
	if err := exec.QueryRowContext(ctx, paidSQL, paidArgs...).Scan(&stats.PaidOrders, &stats.PaidAmount); err != nil {
		return nil, fmt.Errorf("get paid stats: %w", err)
	}

	pendingSQL := "SELECT COUNT(*) FROM orders " + where + fmt.Sprintf(" AND status = $%d", argIdx)
	pendingArgs := append(args, OrderStatusPending)
	if err := exec.QueryRowContext(ctx, pendingSQL, pendingArgs...).Scan(&stats.PendingOrders); err != nil {
		return nil, fmt.Errorf("get pending stats: %w", err)
	}

	cancelledSQL := "SELECT COUNT(*) FROM orders " + where + fmt.Sprintf(" AND status = $%d", argIdx)
	cancelledArgs := append(args, OrderStatusCancelled)
	if err := exec.QueryRowContext(ctx, cancelledSQL, cancelledArgs...).Scan(&stats.CancelledOrders); err != nil {
		return nil, fmt.Errorf("get cancelled stats: %w", err)
	}

	return &stats, nil
}

func (r *repository) ListFinalReviewOrders(ctx context.Context, query FinalReviewQuery) ([]*Order, int, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	where := "WHERE status = $1"
	args := []any{OrderStatusFinalReview}
	argIdx := 2

	if query.Status != "" && query.Status != string(OrderStatusFinalReview) {
		where = fmt.Sprintf("WHERE status = $%d", argIdx)
		args = append(args, query.Status)
		argIdx++
	}

	countSQL := "SELECT COUNT(*) FROM orders " + where
	var total int
	if err := exec.QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count final review orders: %w", err)
	}

	listSQL := fmt.Sprintf(
		`SELECT id, order_no, user_id, status, total_amount, pay_amount, discount_amount, goods_id, sku_name, quantity, remark, pay_at, completed_at, cancelled_at, created_at, updated_at
		 FROM orders %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		where, argIdx, argIdx+1,
	)
	args = append(args, query.PageSize, (query.PageNum-1)*query.PageSize)

	rows, err := exec.QueryContext(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list final review orders: %w", err)
	}
	defer rows.Close()

	var orders []*Order
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.OrderNo, &o.UserID, &o.Status, &o.TotalAmount, &o.PayAmount, &o.DiscountAmount, &o.GoodsID, &o.SKUName, &o.Quantity, &o.Remark, &o.PayAt, &o.CompletedAt, &o.CancelledAt, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan final review order: %w", err)
		}
		orders = append(orders, &o)
	}
	return orders, total, nil
}

func (r *repository) CreatePaymentRecord(ctx context.Context, record *PaymentRecord) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	return exec.QueryRowContext(ctx,
		`INSERT INTO payment_records (order_id, user_id, out_trade_no, channel, amount, status, transaction_id, notify_raw)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		record.OrderID, record.UserID, record.OutTradeNo, record.Channel, record.Amount, record.Status, record.TransactionID, record.NotifyRaw,
	).Scan(&record.ID)
}

func (r *repository) GetPaymentByID(ctx context.Context, id int64) (*PaymentRecord, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	row := exec.QueryRowContext(ctx,
		`SELECT id, order_id, user_id, out_trade_no, channel, amount, status, paid_at, transaction_id, notify_raw, created_at, updated_at
		 FROM payment_records WHERE id = $1`, id,
	)
	var rec PaymentRecord
	err := row.Scan(&rec.ID, &rec.OrderID, &rec.UserID, &rec.OutTradeNo, &rec.Channel, &rec.Amount, &rec.Status, &rec.PaidAt, &rec.TransactionID, &rec.NotifyRaw, &rec.CreatedAt, &rec.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get payment by id: %w", err)
	}
	return &rec, nil
}

func (r *repository) GetPaymentByOutTradeNo(ctx context.Context, outTradeNo string) (*PaymentRecord, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	row := exec.QueryRowContext(ctx,
		`SELECT id, order_id, user_id, out_trade_no, channel, amount, status, paid_at, transaction_id, notify_raw, created_at, updated_at
		 FROM payment_records WHERE out_trade_no = $1`, outTradeNo,
	)
	var rec PaymentRecord
	err := row.Scan(&rec.ID, &rec.OrderID, &rec.UserID, &rec.OutTradeNo, &rec.Channel, &rec.Amount, &rec.Status, &rec.PaidAt, &rec.TransactionID, &rec.NotifyRaw, &rec.CreatedAt, &rec.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get payment by out trade no: %w", err)
	}
	return &rec, nil
}

func (r *repository) ListPayments(ctx context.Context, query PaymentQuery) ([]*PaymentRecord, int, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	where := "WHERE 1=1"
	args := []any{}
	argIdx := 1

	if query.OrderID != nil {
		where += fmt.Sprintf(" AND order_id = $%d", argIdx)
		args = append(args, *query.OrderID)
		argIdx++
	}
	if query.UserID != nil {
		where += fmt.Sprintf(" AND user_id = $%d", argIdx)
		args = append(args, *query.UserID)
		argIdx++
	}
	if query.OutTradeNo != "" {
		where += fmt.Sprintf(" AND out_trade_no = $%d", argIdx)
		args = append(args, query.OutTradeNo)
		argIdx++
	}
	if query.Status != "" {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, query.Status)
		argIdx++
	}
	if query.Channel != "" {
		where += fmt.Sprintf(" AND channel = $%d", argIdx)
		args = append(args, query.Channel)
		argIdx++
	}
	if query.StartTime != nil {
		where += fmt.Sprintf(" AND created_at >= $%d", argIdx)
		args = append(args, *query.StartTime)
		argIdx++
	}
	if query.EndTime != nil {
		where += fmt.Sprintf(" AND created_at <= $%d", argIdx)
		args = append(args, *query.EndTime)
		argIdx++
	}

	countSQL := "SELECT COUNT(*) FROM payment_records " + where
	var total int
	if err := exec.QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count payments: %w", err)
	}

	listSQL := fmt.Sprintf(
		`SELECT id, order_id, user_id, out_trade_no, channel, amount, status, paid_at, transaction_id, notify_raw, created_at, updated_at
		 FROM payment_records %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		where, argIdx, argIdx+1,
	)
	args = append(args, query.PageSize, (query.PageNum-1)*query.PageSize)

	rows, err := exec.QueryContext(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list payments: %w", err)
	}
	defer rows.Close()

	var records []*PaymentRecord
	for rows.Next() {
		var rec PaymentRecord
		if err := rows.Scan(&rec.ID, &rec.OrderID, &rec.UserID, &rec.OutTradeNo, &rec.Channel, &rec.Amount, &rec.Status, &rec.PaidAt, &rec.TransactionID, &rec.NotifyRaw, &rec.CreatedAt, &rec.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan payment: %w", err)
		}
		records = append(records, &rec)
	}
	return records, total, nil
}

func (r *repository) UpdatePaymentStatus(ctx context.Context, id int64, status string) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	_, err := exec.ExecContext(ctx,
		`UPDATE payment_records SET status = $1, updated_at = NOW() WHERE id = $2`,
		status, id,
	)
	if err != nil {
		return fmt.Errorf("update payment status: %w", err)
	}
	return nil
}

func (r *repository) UpdatePaymentPaidAt(ctx context.Context, id int64, paidAt *time.Time, transactionID *string) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	_, err := exec.ExecContext(ctx,
		`UPDATE payment_records SET paid_at = $1, transaction_id = $2, status = $3, updated_at = NOW() WHERE id = $4`,
		paidAt, transactionID, PaymentStatusPaid, id,
	)
	if err != nil {
		return fmt.Errorf("update payment paid at: %w", err)
	}
	return nil
}

func (r *repository) CreateCashierOrder(ctx context.Context, co *CashierOrder) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	return exec.QueryRowContext(ctx,
		`INSERT INTO cashier_orders (token, order_id, user_id, amount, status, expire_at)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		co.Token, co.OrderID, co.UserID, co.Amount, co.Status, co.ExpireAt,
	).Scan(&co.ID)
}

func (r *repository) GetCashierOrderByToken(ctx context.Context, token string) (*CashierOrder, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	row := exec.QueryRowContext(ctx,
		`SELECT id, token, order_id, user_id, amount, status, expire_at, paid_at, pay_channel, created_at, updated_at
		 FROM cashier_orders WHERE token = $1`, token,
	)
	var co CashierOrder
	err := row.Scan(&co.ID, &co.Token, &co.OrderID, &co.UserID, &co.Amount, &co.Status, &co.ExpireAt, &co.PaidAt, &co.PayChannel, &co.CreatedAt, &co.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get cashier order by token: %w", err)
	}
	return &co, nil
}

func (r *repository) UpdateCashierOrderStatus(ctx context.Context, id int64, status string, payChannel *string) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	_, err := exec.ExecContext(ctx,
		`UPDATE cashier_orders SET status = $1, pay_channel = $2, updated_at = NOW() WHERE id = $3`,
		status, payChannel, id,
	)
	if err != nil {
		return fmt.Errorf("update cashier order status: %w", err)
	}
	return nil
}

func (r *repository) CreatePaymentSyncLog(ctx context.Context, log *PaymentSyncLog) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	return exec.QueryRowContext(ctx,
		`INSERT INTO payment_sync_logs (record_id, channel, out_trade_no, action, request, response, success, error_msg)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		log.RecordID, log.Channel, log.OutTradeNo, log.Action, log.Request, log.Response, log.Success, log.ErrorMsg,
	).Scan(&log.ID)
}

func (r *repository) ListPaymentSyncLogs(ctx context.Context, recordID int64) ([]*PaymentSyncLog, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	rows, err := exec.QueryContext(ctx,
		`SELECT id, record_id, channel, out_trade_no, action, request, response, success, error_msg, created_at
		 FROM payment_sync_logs WHERE record_id = $1 ORDER BY created_at DESC`,
		recordID,
	)
	if err != nil {
		return nil, fmt.Errorf("list payment sync logs: %w", err)
	}
	defer rows.Close()

	var logs []*PaymentSyncLog
	for rows.Next() {
		var l PaymentSyncLog
		if err := rows.Scan(&l.ID, &l.RecordID, &l.Channel, &l.OutTradeNo, &l.Action, &l.Request, &l.Response, &l.Success, &l.ErrorMsg, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan payment sync log: %w", err)
		}
		logs = append(logs, &l)
	}
	return logs, nil
}
