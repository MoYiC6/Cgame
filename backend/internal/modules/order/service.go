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
