package notification

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

func (s *Service) CreateNotification(ctx context.Context, n *Notification) error {
	if n.Title == "" || n.Type == "" {
		return fmt.Errorf("title and type are required")
	}
	return s.repo.CreateNotification(ctx, n)
}

func (s *Service) GetUserNotifications(ctx context.Context, userID int64, page, pageSize int) ([]*Notification, int, error) {
	return s.repo.GetUserNotifications(ctx, userID, page, pageSize)
}

func (s *Service) MarkAsRead(ctx context.Context, userID, notificationID int64) error {
	return s.repo.MarkAsRead(ctx, userID, notificationID)
}

func (s *Service) MarkAllAsRead(ctx context.Context, userID int64) error {
	return s.repo.MarkAllAsRead(ctx, userID)
}

func (s *Service) GetUnreadCount(ctx context.Context, userID int64) (int, error) {
	return s.repo.GetUnreadCount(ctx, userID)
}

func (s *Service) ListInboxNotifications(ctx context.Context, userID int64, page, pageSize int, notificationType string, unreadOnly *bool) (*NotificationInboxList, error) {
	if userID == 0 {
		return nil, fmt.Errorf("user id is required")
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return s.repo.ListInboxNotifications(ctx, userID, page, pageSize, notificationType, unreadOnly)
}

func (s *Service) MarkInboxAsRead(ctx context.Context, userID, inboxID int64) error {
	if userID == 0 || inboxID == 0 {
		return fmt.Errorf("user id and inbox id are required")
	}
	return s.repo.MarkInboxAsRead(ctx, userID, inboxID)
}

func (s *Service) MarkAllInboxAsRead(ctx context.Context, userID int64, notificationType string) error {
	if userID == 0 {
		return fmt.Errorf("user id is required")
	}
	return s.repo.MarkAllInboxAsRead(ctx, userID, notificationType)
}

func (s *Service) CreateTodo(ctx context.Context, t *SystemTodo) (int64, error) {
	if t.Title == "" {
		return 0, fmt.Errorf("title is required")
	}
	if t.SortOrder == nil {
		zero := 0
		t.SortOrder = &zero
	}
	if err := s.repo.CreateTodo(ctx, t); err != nil {
		return 0, err
	}
	return t.ID, nil
}

func (s *Service) GetTodos(ctx context.Context, completed *bool) ([]*SystemTodo, error) {
	return s.repo.GetTodos(ctx, completed)
}

func (s *Service) ToggleTodo(ctx context.Context, id int64, completed bool, operator string) error {
	return s.repo.ToggleTodo(ctx, id, completed, operator)
}

func (s *Service) DeleteTodos(ctx context.Context, ids []int64) error {
	return s.repo.DeleteTodos(ctx, ids)
}
