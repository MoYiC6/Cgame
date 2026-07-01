-- name: CreateNotification :one
INSERT INTO notifications (user_id, title, content, type, sub_type, target_type, target_id, related_id, related_type, extra_data, priority, is_read, expire_time, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW(), NOW())
RETURNING id;

-- name: GetUserNotifications :many
SELECT n.id, n.user_id, n.title, n.content, n.type, n.sub_type, n.target_type, n.target_id, n.related_id, n.related_type, n.extra_data, n.priority, n.is_read, n.expire_time, n.read_time, n.created_at, n.updated_at
FROM notifications n
WHERE n.user_id = $1
ORDER BY n.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountUserNotifications :one
SELECT COUNT(*) FROM notifications
WHERE user_id = $1;

-- name: MarkAsRead :exec
UPDATE user_notifications
SET is_read = 1, read_time = NOW()
WHERE user_id = $1 AND notification_id = $2;

-- name: MarkAllAsRead :exec
UPDATE user_notifications
SET is_read = 1, read_time = NOW()
WHERE user_id = $1 AND is_read = 0;

-- name: GetUnreadCount :one
SELECT COUNT(*) FROM user_notifications
WHERE user_id = $1 AND is_read = 0;

-- name: CreateTodo :one
INSERT INTO system_todos (title, completed, sort_order, created_by, completed_by, completed_time, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
RETURNING id;

-- name: GetTodos :many
SELECT id, title, completed, sort_order, created_by, completed_by, completed_time, created_at, updated_at
FROM system_todos
ORDER BY sort_order ASC, id ASC;

-- name: ToggleTodo :exec
UPDATE system_todos
SET completed = $2, completed_by = $3, completed_time = $4, updated_at = NOW()
WHERE id = $1;

-- name: DeleteTodos :exec
DELETE FROM system_todos
WHERE id = ANY($1::bigint[]);
