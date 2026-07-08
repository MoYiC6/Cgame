package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"backend/internal/platform/database"
)

type Repository interface {
	CreateNotification(ctx context.Context, n *Notification) error
	GetNotificationByID(ctx context.Context, id int64) (*Notification, error)
	GetUserNotifications(ctx context.Context, userID int64, page, pageSize int) ([]*Notification, int, error)
	MarkAsRead(ctx context.Context, userID, notificationID int64) error
	MarkAllAsRead(ctx context.Context, userID int64) error
	GetUnreadCount(ctx context.Context, userID int64) (int, error)
	ListAdminNotifications(ctx context.Context, page, pageSize int) ([]*Notification, int, error)
	DeleteNotification(ctx context.Context, id int64) error
	GetNotificationStats(ctx context.Context) (*NotificationStats, error)
	ListInboxNotifications(ctx context.Context, userID int64, page, pageSize int, notificationType string, unreadOnly *bool) (*NotificationInboxList, error)
	MarkInboxAsRead(ctx context.Context, userID, inboxID int64) error
	MarkAllInboxAsRead(ctx context.Context, userID int64, notificationType string) error
	CreateTodo(ctx context.Context, t *SystemTodo) error
	GetTodos(ctx context.Context, completed *bool) ([]*SystemTodo, error)
	ToggleTodo(ctx context.Context, id int64, completed bool, operator string) error
	DeleteTodos(ctx context.Context, ids []int64) error

	// Subscribe message
	GetSubscribeTemplates(ctx context.Context) ([]*SubscribeTemplate, error)
	RecordSubscribe(ctx context.Context, userID int64, templateID string) error
	GetSubscribeStatus(ctx context.Context, userID int64, templateID string) (*SubscribeStatus, error)

	// Admin notification management
	UpdateNotification(ctx context.Context, n *Notification) error
	MarkAdminNotificationAsRead(ctx context.Context, userID, notificationID int64) error
	MarkAllAdminNotificationsAsRead(ctx context.Context, userID int64) error
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

func (r *repository) ListInboxNotifications(ctx context.Context, userID int64, page, pageSize int, notificationType string, unreadOnly *bool) (*NotificationInboxList, error) {
	where, args := inboxWhereClause(userID, notificationType, unreadOnly)

	var total int64
	if err := r.dbtx.QueryRowContext(ctx, "SELECT COUNT(*) FROM user_notifications un JOIN notifications n ON n.id = un.notification_id "+where, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count inbox notifications: %w", err)
	}

	var unreadCount int64
	if err := r.dbtx.QueryRowContext(ctx, `SELECT COUNT(*) FROM user_notifications WHERE user_id = $1 AND is_read = 0`, userID).Scan(&unreadCount); err != nil {
		return nil, fmt.Errorf("count unread inbox notifications: %w", err)
	}

	query := `SELECT un.id, un.notification_id, n.title, n.content, n.type, n.target_type,
	                 n.target_id::text, un.is_read, un.read_time, n.created_at, un.created_at
	          FROM user_notifications un
	          JOIN notifications n ON n.id = un.notification_id ` + where + `
	          ORDER BY un.created_at DESC
	          LIMIT $` + strconv.Itoa(len(args)+1) + ` OFFSET $` + strconv.Itoa(len(args)+2)
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := r.dbtx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list inbox notifications: %w", err)
	}
	defer rows.Close()

	result := &NotificationInboxList{Total: total, UnreadCount: unreadCount, Rows: []NotificationInboxItem{}}
	for rows.Next() {
		var item NotificationInboxItem
		if err := rows.Scan(&item.ID, &item.NotificationID, &item.Title, &item.Content, &item.Type, &item.TargetType, &item.TargetID, &item.IsRead, &item.ReadTime, &item.SendTime, &item.CreateTime); err != nil {
			return nil, fmt.Errorf("scan inbox notification: %w", err)
		}
		result.Rows = append(result.Rows, item)
	}
	return result, nil
}

func (r *repository) MarkInboxAsRead(ctx context.Context, userID, inboxID int64) error {
	_, err := r.dbtx.ExecContext(ctx,
		`UPDATE user_notifications SET is_read = 1, read_time = NOW() WHERE id = $1 AND user_id = $2`,
		inboxID, userID,
	)
	if err != nil {
		return fmt.Errorf("mark inbox as read: %w", err)
	}
	return nil
}

func (r *repository) MarkAllInboxAsRead(ctx context.Context, userID int64, notificationType string) error {
	query := `UPDATE user_notifications un SET is_read = 1, read_time = NOW()
	          FROM notifications n
	          WHERE n.id = un.notification_id AND un.user_id = $1 AND un.is_read = 0`
	args := []interface{}{userID}
	if notificationType != "" {
		query += " AND n.type = $2"
		args = append(args, notificationType)
	}
	_, err := r.dbtx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("mark all inbox as read: %w", err)
	}
	return nil
}

func inboxWhereClause(userID int64, notificationType string, unreadOnly *bool) (string, []interface{}) {
	where := "WHERE un.user_id = $1"
	args := []interface{}{userID}
	if notificationType != "" {
		where += fmt.Sprintf(" AND n.type = $%d", len(args)+1)
		args = append(args, notificationType)
	}
	if unreadOnly != nil && *unreadOnly {
		where += " AND un.is_read = 0"
	}
	return where, args
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

func (r *repository) GetNotificationByID(ctx context.Context, id int64) (*Notification, error) {
	var n Notification
	var extraDataJSON []byte
	err := r.dbtx.QueryRowContext(ctx,
		`SELECT id, user_id, title, content, type, sub_type, target_type, target_id, related_id, related_type, extra_data, priority, is_read, expire_time, read_time, created_at, updated_at
		 FROM notifications WHERE id = $1`,
		id,
	).Scan(&n.ID, &n.UserID, &n.Title, &n.Content, &n.Type, &n.SubType, &n.TargetType, &n.TargetID, &n.RelatedID, &n.RelatedType, &extraDataJSON, &n.Priority, &n.IsRead, &n.ExpireTime, &n.ReadTime, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get notification by id: %w", err)
	}
	if len(extraDataJSON) > 0 {
		json.Unmarshal(extraDataJSON, &n.ExtraData)
	}
	return &n, nil
}

func (r *repository) ListAdminNotifications(ctx context.Context, page, pageSize int) ([]*Notification, int, error) {
	var total int
	if err := r.dbtx.QueryRowContext(ctx, `SELECT COUNT(*) FROM notifications`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count admin notifications: %w", err)
	}

	rows, err := r.dbtx.QueryContext(ctx,
		`SELECT id, user_id, title, content, type, sub_type, target_type, target_id, related_id, related_type, extra_data, priority, is_read, expire_time, read_time, created_at, updated_at
		 FROM notifications ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		pageSize, (page-1)*pageSize,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list admin notifications: %w", err)
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

func (r *repository) DeleteNotification(ctx context.Context, id int64) error {
	_, err := r.dbtx.ExecContext(ctx, `DELETE FROM notifications WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete notification: %w", err)
	}
	return nil
}

func (r *repository) GetNotificationStats(ctx context.Context) (*NotificationStats, error) {
	stats := &NotificationStats{}
	if err := r.dbtx.QueryRowContext(ctx, `SELECT COUNT(*) FROM notifications`).Scan(&stats.TotalNotifications); err != nil {
		return nil, fmt.Errorf("get total notifications: %w", err)
	}
	if err := r.dbtx.QueryRowContext(ctx, `SELECT COUNT(*) FROM notifications WHERE is_read = FALSE`).Scan(&stats.UnreadCount); err != nil {
		return nil, fmt.Errorf("get unread count: %w", err)
	}
	if err := r.dbtx.QueryRowContext(ctx, `SELECT COUNT(*) FROM notifications WHERE created_at >= CURRENT_DATE`).Scan(&stats.TodayCount); err != nil {
		return nil, fmt.Errorf("get today count: %w", err)
	}
	return stats, nil
}

func (r *repository) GetSubscribeTemplates(ctx context.Context) ([]*SubscribeTemplate, error) {
	return []*SubscribeTemplate{
		{ID: 1, TemplateID: "order_complete", Name: "订单完成通知", Enabled: true},
		{ID: 2, TemplateID: "withdrawal_approved", Name: "提现审批通知", Enabled: true},
		{ID: 3, TemplateID: "new_message", Name: "新消息通知", Enabled: true},
	}, nil
}

func (r *repository) RecordSubscribe(ctx context.Context, userID int64, templateID string) error {
	_, err := r.dbtx.ExecContext(ctx,
		`INSERT INTO subscribe_records (user_id, template_id, subscribed_at) VALUES ($1, $2, NOW())
		 ON CONFLICT (user_id, template_id) DO UPDATE SET subscribed_at = NOW()`,
		userID, templateID,
	)
	if err != nil {
		return fmt.Errorf("record subscribe: %w", err)
	}
	return nil
}

func (r *repository) GetSubscribeStatus(ctx context.Context, userID int64, templateID string) (*SubscribeStatus, error) {
	var subscribedAt *time.Time
	err := r.dbtx.QueryRowContext(ctx,
		`SELECT subscribed_at FROM subscribe_records WHERE user_id = $1 AND template_id = $2`,
		userID, templateID,
	).Scan(&subscribedAt)
	if err != nil {
		return &SubscribeStatus{TemplateID: templateID, Subscribed: false}, nil
	}
	return &SubscribeStatus{TemplateID: templateID, Subscribed: true, SubscribedAt: subscribedAt}, nil
}

func (r *repository) UpdateNotification(ctx context.Context, n *Notification) error {
	extraDataJSON, _ := json.Marshal(n.ExtraData)
	_, err := r.dbtx.ExecContext(ctx,
		`UPDATE notifications SET title = $1, content = $2, type = $3, sub_type = $4, target_type = $5, target_id = $6, related_id = $7, related_type = $8, extra_data = $9, priority = $10, expire_time = $11, updated_at = NOW() WHERE id = $12`,
		n.Title, n.Content, n.Type, n.SubType, n.TargetType, n.TargetID, n.RelatedID, n.RelatedType, string(extraDataJSON), n.Priority, n.ExpireTime, n.ID,
	)
	if err != nil {
		return fmt.Errorf("update notification: %w", err)
	}
	return nil
}

func (r *repository) MarkAdminNotificationAsRead(ctx context.Context, userID, notificationID int64) error {
	_, err := r.dbtx.ExecContext(ctx,
		`INSERT INTO notification_reads (user_id, notification_id, read_at) VALUES ($1, $2, NOW())
		 ON CONFLICT (user_id, notification_id) DO NOTHING`,
		userID, notificationID,
	)
	if err != nil {
		return fmt.Errorf("mark admin notification as read: %w", err)
	}
	return nil
}

func (r *repository) MarkAllAdminNotificationsAsRead(ctx context.Context, userID int64) error {
	_, err := r.dbtx.ExecContext(ctx,
		`INSERT INTO notification_reads (user_id, notification_id, read_at)
		 SELECT $1, id, NOW() FROM notifications
		 WHERE id NOT IN (SELECT notification_id FROM notification_reads WHERE user_id = $1)
		 ON CONFLICT (user_id, notification_id) DO NOTHING`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("mark all admin notifications as read: %w", err)
	}
	return nil
}
