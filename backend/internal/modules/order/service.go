package order

import (
	"context"
	"fmt"
	"time"

	"backend/internal/platform/database"
)

type Service interface {
	Ping(ctx context.Context) (PingResponse, error)
	CreateOrder(ctx context.Context, userID int64, skuName string, quantity int, remark *string) (*Order, error)
	GetOrder(ctx context.Context, orderID int64) (*Order, error)
	ListOrders(ctx context.Context, userID int64, page, pageSize int) ([]*Order, int, error)
	PayOrder(ctx context.Context, orderID int64) error
	CompleteOrder(ctx context.Context, orderID int64) error
	CancelOrder(ctx context.Context, orderID int64) error

	// Client extended
	ComplaintOrder(ctx context.Context, orderID, userID int64, reason string, detail *string) error
	ConfirmTeacher(ctx context.Context, orderID int64, teacherID int64) error
	GetOrderStatistics(ctx context.Context, userID int64) (*OrderStats, error)
	CreateReview(ctx context.Context, orderID, userID int64, req ReviewRequest) error
	GetReviewByOrder(ctx context.Context, orderID int64) (*OrderReview, error)
	ListReviews(ctx context.Context, page, pageSize int) ([]*OrderReview, int, error)

	// Admin
	AdminListOrders(ctx context.Context, query OrderQuery) ([]*Order, int, error)
	AdminGetOrder(ctx context.Context, orderID int64) (*Order, error)
	AdminUpdateOrderStatus(ctx context.Context, orderID int64, status string) error
	AdminRefundOrder(ctx context.Context, orderID int64) error
	AdminManualComplete(ctx context.Context, orderID int64) error
	AdminUpdateRemark(ctx context.Context, orderID int64, remark string) error
	AdminUpdateTeachers(ctx context.Context, orderID int64, teacherIDs []int64) error
	AdminManualCreateOrder(ctx context.Context, req ManualOrderRequest) (*Order, error)
	AdminGetOrderStats(ctx context.Context, start, end *string) (*OrderStats, error)
	AdminListReviews(ctx context.Context, query ReviewQuery) ([]*OrderReview, int, error)
	AdminUpdateReviewStatus(ctx context.Context, reviewID int64, status string) error
	AdminReplyReview(ctx context.Context, reviewID int64, reply string) error
	AdminListFinalReview(ctx context.Context, query FinalReviewQuery) ([]*Order, int, error)
	AdminApproveFinalReview(ctx context.Context, orderID int64) error
	AdminRejectFinalReview(ctx context.Context, orderID int64) error

	// Transfer
	GetTransferConfig(ctx context.Context) (*OrderTransferConfig, error)
	TransferOrder(ctx context.Context, orderID int64, targetTeacherID int64) error
}

type service struct {
	repo      Repository
	txManager database.TxManager
}

func NewService(repo Repository, txManager database.TxManager) Service {
	if txManager == nil {
		txManager = database.NoopTxManager{}
	}
	return &service{repo: repo, txManager: txManager}
}

func (s *service) Ping(ctx context.Context) (PingResponse, error) {
	if err := s.repo.Ping(ctx); err != nil {
		return PingResponse{}, err
	}
	return PingResponse{Module: "order"}, nil
}

func (s *service) CreateOrder(ctx context.Context, userID int64, skuName string, quantity int, remark *string) (*Order, error) {
	if userID == 0 {
		return nil, fmt.Errorf("user_id is required")
	}
	if quantity <= 0 {
		return nil, fmt.Errorf("quantity must be > 0")
	}

	orderNo := fmt.Sprintf("ORD%d%d", time.Now().UnixNano(), userID)
	order := &Order{
		OrderNo:  orderNo,
		UserID:   userID,
		Status:   OrderStatusPending,
		SKUName:  skuName,
		Quantity: quantity,
		Remark:   remark,
	}

	if err := s.repo.CreateOrder(ctx, order); err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	return order, nil
}

func (s *service) GetOrder(ctx context.Context, orderID int64) (*Order, error) {
	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("get order: %w", err)
	}
	return order, nil
}

func (s *service) ListOrders(ctx context.Context, userID int64, page, pageSize int) ([]*Order, int, error) {
	return s.repo.ListOrders(ctx, userID, page, pageSize)
}

func (s *service) PayOrder(ctx context.Context, orderID int64) error {
	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("order not found: %w", err)
	}
	if order.Status != OrderStatusPending {
		return fmt.Errorf("order cannot be paid")
	}
	if err := s.repo.UpdateOrderStatus(ctx, orderID, OrderStatusPaid); err != nil {
		return fmt.Errorf("pay order: %w", err)
	}
	return nil
}

func (s *service) CompleteOrder(ctx context.Context, orderID int64) error {
	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("order not found: %w", err)
	}
	if order.Status != OrderStatusPaid {
		return fmt.Errorf("order cannot be completed")
	}
	if err := s.repo.UpdateOrderStatus(ctx, orderID, OrderStatusCompleted); err != nil {
		return fmt.Errorf("complete order: %w", err)
	}
	return nil
}

func (s *service) CancelOrder(ctx context.Context, orderID int64) error {
	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("order not found: %w", err)
	}
	if order.Status != OrderStatusPending {
		return fmt.Errorf("order cannot be cancelled")
	}
	if err := s.repo.UpdateOrderStatus(ctx, orderID, OrderStatusCancelled); err != nil {
		return fmt.Errorf("cancel order: %w", err)
	}
	return nil
}

func (s *service) ComplaintOrder(ctx context.Context, orderID, userID int64, reason string, detail *string) error {
	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("order not found: %w", err)
	}
	if order.UserID != userID {
		return fmt.Errorf("not your order")
	}
	complaint := &OrderComplaint{
		OrderID: orderID,
		UserID:  userID,
		Reason:  reason,
		Detail:  detail,
		Status:  string(ComplaintStatusPending),
	}
	return s.repo.CreateComplaint(ctx, complaint)
}

func (s *service) ConfirmTeacher(ctx context.Context, orderID int64, teacherID int64) error {
	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("order not found: %w", err)
	}
	if order.Status != OrderStatusPaid && order.Status != OrderStatusPending {
		return fmt.Errorf("order cannot confirm teacher")
	}
	return s.repo.UpdateOrderTeachers(ctx, orderID, []int64{teacherID})
}

func (s *service) GetOrderStatistics(ctx context.Context, userID int64) (*OrderStats, error) {
	// For client statistics, we can reuse the admin stats without date filter or add user filter later.
	// Here we return overall stats as a simplified implementation.
	return s.repo.GetOrderStats(ctx, nil, nil)
}

func (s *service) CreateReview(ctx context.Context, orderID, userID int64, req ReviewRequest) error {
	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("order not found: %w", err)
	}
	if order.UserID != userID {
		return fmt.Errorf("not your order")
	}
	if order.Status != OrderStatusCompleted {
		return fmt.Errorf("order not completed")
	}
	review := &OrderReview{
		OrderID:   orderID,
		UserID:    userID,
		TeacherID: req.TeacherID,
		Rating:    req.Rating,
		Content:   req.Content,
		Status:    string(ReviewStatusPending),
	}
	return s.repo.CreateReview(ctx, review)
}

func (s *service) GetReviewByOrder(ctx context.Context, orderID int64) (*OrderReview, error) {
	return s.repo.GetReviewByOrderID(ctx, orderID)
}

func (s *service) ListReviews(ctx context.Context, page, pageSize int) ([]*OrderReview, int, error) {
	return s.repo.ListReviews(ctx, ReviewQuery{PageNum: page, PageSize: pageSize})
}

func (s *service) AdminListOrders(ctx context.Context, query OrderQuery) ([]*Order, int, error) {
	return s.repo.AdminListOrders(ctx, query)
}

func (s *service) AdminGetOrder(ctx context.Context, orderID int64) (*Order, error) {
	return s.repo.AdminGetOrderByID(ctx, orderID)
}

func (s *service) AdminUpdateOrderStatus(ctx context.Context, orderID int64, status string) error {
	return s.repo.UpdateOrderStatus(ctx, orderID, status)
}

func (s *service) AdminRefundOrder(ctx context.Context, orderID int64) error {
	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("order not found: %w", err)
	}
	if order.Status != OrderStatusPaid && order.Status != OrderStatusCompleted {
		return fmt.Errorf("order cannot be refunded")
	}
	return s.repo.UpdateOrderStatus(ctx, orderID, OrderStatusRefunded)
}

func (s *service) AdminManualComplete(ctx context.Context, orderID int64) error {
	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("order not found: %w", err)
	}
	if order.Status != OrderStatusPaid && order.Status != OrderStatusPending {
		return fmt.Errorf("order cannot be manually completed")
	}
	return s.repo.UpdateOrderStatus(ctx, orderID, OrderStatusCompleted)
}

func (s *service) AdminUpdateRemark(ctx context.Context, orderID int64, remark string) error {
	return s.repo.UpdateOrderRemark(ctx, orderID, remark)
}

func (s *service) AdminUpdateTeachers(ctx context.Context, orderID int64, teacherIDs []int64) error {
	return s.repo.UpdateOrderTeachers(ctx, orderID, teacherIDs)
}

func (s *service) AdminManualCreateOrder(ctx context.Context, req ManualOrderRequest) (*Order, error) {
	orderNo := fmt.Sprintf("ORD%d%d", time.Now().UnixNano(), req.UserID)
	order := &Order{
		OrderNo:     orderNo,
		UserID:      req.UserID,
		Status:      OrderStatusPending,
		GoodsID:     req.GoodsID,
		SKUName:     req.SKUName,
		Quantity:    req.Quantity,
		TotalAmount: req.TotalAmount,
		Remark:      req.Remark,
	}
	if err := s.repo.CreateOrder(ctx, order); err != nil {
		return nil, fmt.Errorf("manual create order: %w", err)
	}
	return order, nil
}

func (s *service) AdminGetOrderStats(ctx context.Context, start, end *string) (*OrderStats, error) {
	return s.repo.GetOrderStats(ctx, start, end)
}

func (s *service) AdminListReviews(ctx context.Context, query ReviewQuery) ([]*OrderReview, int, error) {
	return s.repo.ListReviews(ctx, query)
}

func (s *service) AdminUpdateReviewStatus(ctx context.Context, reviewID int64, status string) error {
	return s.repo.UpdateReviewStatus(ctx, reviewID, status)
}

func (s *service) AdminReplyReview(ctx context.Context, reviewID int64, reply string) error {
	return s.repo.UpdateReviewReply(ctx, reviewID, reply)
}

func (s *service) AdminListFinalReview(ctx context.Context, query FinalReviewQuery) ([]*Order, int, error) {
	return s.repo.ListFinalReviewOrders(ctx, query)
}

func (s *service) AdminApproveFinalReview(ctx context.Context, orderID int64) error {
	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("order not found: %w", err)
	}
	if order.Status != OrderStatusFinalReview {
		return fmt.Errorf("order not in final review")
	}
	return s.repo.UpdateOrderStatus(ctx, orderID, OrderStatusCompleted)
}

func (s *service) AdminRejectFinalReview(ctx context.Context, orderID int64) error {
	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("order not found: %w", err)
	}
	if order.Status != OrderStatusFinalReview {
		return fmt.Errorf("order not in final review")
	}
	return s.repo.UpdateOrderStatus(ctx, orderID, OrderStatusCancelled)
}

func (s *service) GetTransferConfig(ctx context.Context) (*OrderTransferConfig, error) {
	return s.repo.GetTransferConfig(ctx)
}

func (s *service) TransferOrder(ctx context.Context, orderID int64, targetTeacherID int64) error {
	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("order not found: %w", err)
	}
	if order.Status != OrderStatusPaid && order.Status != OrderStatusInTransfer {
		return fmt.Errorf("order cannot be transferred")
	}
	if err := s.repo.UpdateOrderStatus(ctx, orderID, OrderStatusInTransfer); err != nil {
		return fmt.Errorf("transfer order: %w", err)
	}
	return s.repo.UpdateOrderTeachers(ctx, orderID, []int64{targetTeacherID})
}
