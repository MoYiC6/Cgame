package notification

import (
	"context"
	"encoding/json"
	"fmt"

	"backend/internal/platform/database"
)

type Repository interface {
	CreateNotification(ctx context.Context, n *Notification) error
	GetUserNotifications(ctx context.Context, userID int64, page, pageSize int) ([]*Notification, int, error)
	MarkAsRead(ctx context.Context, userID, notificationID int64) error
	MarkAllAsRead(ctx context.Context, userID int64) error
	GetUnreadCount(ctx context.Context, userID int64) (int, error)
	CreateTodo(ctx context.Context, t *SystemTodo) error
	GetTodos(ctx context.Context, completed *bool) ([]*SystemTodo, error)
	ToggleTodo(ctx context.Context, id int64, completed bool, operator string) error
	DeleteTodos(ctx context.Context, ids []int64) error
}

type repository struct {
	dbtx database.DBTX
}

func NewRepository(dbtx database.DBTX) Repository {
	return &repository{dbtx: dbtx}
}

func (r *repository) CreateNotification(ctx context.Context, n *Notification) error {
	extraDataJSON, _ := json.Marshal(n.ExtraData)
	_, err := r.dbtx.ExecContext(ctx,
		`INSERT INTO notifications (user_id, title, content, type, sub_type, target_type, target_id, related_id, related_type, extra_data, priority, expire_time)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		n.UserID, n.Title, n.Content, n.Type, n.SubType, n.TargetType, n.TargetID, n.RelatedID, n.RelatedType, string(extraDataJSON), n.Priority, n.ExpireTime,
	)
	if err != nil {
		return fmt.Errorf("create notification: %w", err)
	}
	return nil
}

func (r *repository) GetUserNotifications(ctx context.Context, userID int64, page, pageSize int) ([]*Notification, int, error) {
	countQuery := "SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND (expire_time IS NULL OR expire_time > NOW())"
	var total int
	if err := r.dbtx.QueryRowContext(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count notifications: %w", err)
	}

	query := `SELECT id, user_id, title, content, type, sub_type, target_type, target_id, related_id, related_type, extra_data, priority, is_read, expire_time, read_time, created_at, updated_at
			  FROM notifications WHERE user_id = $1 AND (expire_time IS NULL OR expire_time > NOW())
			  ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.dbtx.QueryContext(ctx, query, userID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()

	notifications := make([]*Notification, 0)
	for rows.Next() {
		var n Notification
		var extraDataJSON []byte
		err := rows.Scan(&n.ID, &n.UserID, &n.Title, &n.Content, &n.Type, &n.SubType, &n.TargetType, &n.TargetID, &n.RelatedID, &n.RelatedType, &extraDataJSON, &n.Priority, &n.IsRead, &n.ExpireTime, &n.ReadTime, &n.CreatedAt, &n.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("scan notification: %w", err)
		}
		if len(extraDataJSON) > 0 {
			json.Unmarshal(extraDataJSON, &n.ExtraData)
		}
		notifications = append(notifications, &n)
	}
	return notifications, total, nil
}

func (r *repository) MarkAsRead(ctx context.Context, userID, notificationID int64) error {
	_, err := r.dbtx.ExecContext(ctx,
		`UPDATE notifications SET is_read = TRUE, read_time = NOW(), updated_at = NOW()
		 WHERE id = $1 AND user_id = $2`,
		notificationID, userID,
	)
	if err != nil {
		return fmt.Errorf("mark as read: %w", err)
	}
	return nil
}

func (r *repository) MarkAllAsRead(ctx context.Context, userID int64) error {
	_, err := r.dbtx.ExecContext(ctx,
		`UPDATE notifications SET is_read = TRUE, read_time = NOW(), updated_at = NOW()
		 WHERE user_id = $1 AND is_read = FALSE`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("mark all as read: %w", err)
	}
	return nil
}

func (r *repository) GetUnreadCount(ctx context.Context, userID int64) (int, error) {
	var count int
	err := r.dbtx.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = FALSE AND (expire_time IS NULL OR expire_time > NOW())",
		userID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("get unread count: %w", err)
	}
	return count, nil
}

func (r *repository) CreateTodo(ctx context.Context, t *SystemTodo) error {
	return r.dbtx.QueryRowContext(ctx,
		`INSERT INTO system_todos (title, sort_order, created_by) VALUES ($1, $2, $3) RETURNING id`,
		t.Title, t.SortOrder, t.CreatedBy,
	).Scan(&t.ID)
}

func (r *repository) GetTodos(ctx context.Context, completed *bool) ([]*SystemTodo, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	if completed != nil {
		where += fmt.Sprintf(" AND completed = $%d", len(args)+1)
		args = append(args, *completed)
	}
	rows, err := r.dbtx.QueryContext(ctx,
		"SELECT id, title, completed, sort_order, created_by, completed_by, completed_time, created_at, updated_at FROM system_todos "+where+" ORDER BY completed ASC, sort_order ASC, created_at DESC",
		args...,
	)
	if err != nil {
		return nil, fmt.Errorf("get todos: %w", err)
	}
	defer rows.Close()

	var todos []*SystemTodo
	for rows.Next() {
		var t SystemTodo
		if err := rows.Scan(&t.ID, &t.Title, &t.Completed, &t.SortOrder, &t.CreatedBy, &t.CompletedBy, &t.CompletedTime, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan todo: %w", err)
		}
		todos = append(todos, &t)
	}
	return todos, nil
}

func (r *repository) ToggleTodo(ctx context.Context, id int64, completed bool, operator string) error {
	if completed {
		return r.dbtx.QueryRowContext(ctx,
			"UPDATE system_todos SET completed = TRUE, completed_by = $1, completed_time = NOW(), updated_at = NOW() WHERE id = $2 RETURNING id",
			operator, id,
		).Scan(&id)
	}
	return r.dbtx.QueryRowContext(ctx,
		"UPDATE system_todos SET completed = FALSE, completed_by = NULL, completed_time = NULL, updated_at = NOW() WHERE id = $1 RETURNING id",
		id,
	).Scan(&id)
}

func (r *repository) DeleteTodos(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	query := "DELETE FROM system_todos WHERE id = ANY($1::bigint[])"
	_, err := r.dbtx.ExecContext(ctx, query, ids)
	if err != nil {
		return fmt.Errorf("delete todos: %w", err)
	}
	return nil
}
