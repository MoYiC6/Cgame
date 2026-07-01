-- name: CreateSession :one
INSERT INTO chat_sessions (user_id, teacher_id, teacher_user_id, order_id, last_message_id, last_message_content, last_message_time, user_unread_count, teacher_unread_count, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
RETURNING id;

-- name: GetSessionByID :one
SELECT id, user_id, teacher_id, teacher_user_id, order_id, last_message_id, last_message_content, last_message_time, user_unread_count, teacher_unread_count, status, created_at, updated_at
FROM chat_sessions
WHERE id = $1;

-- name: ListUserSessions :many
SELECT id, user_id, teacher_id, teacher_user_id, order_id, last_message_id, last_message_content, last_message_time, user_unread_count, teacher_unread_count, status, created_at, updated_at
FROM chat_sessions
WHERE user_id = $1
ORDER BY last_message_time DESC
LIMIT $2 OFFSET $3;

-- name: CountUserSessions :one
SELECT COUNT(*) FROM chat_sessions
WHERE user_id = $1;

-- name: CreateMessage :one
INSERT INTO chat_messages (session_id, sender_id, sender_type, receiver_id, content, message_type, extra_data, is_read, read_time, status, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
RETURNING id;

-- name: GetMessagesBySessionID :many
SELECT id, session_id, sender_id, sender_type, receiver_id, content, message_type, extra_data, is_read, read_time, status, created_at
FROM chat_messages
WHERE session_id = $1
ORDER BY created_at ASC
LIMIT $2 OFFSET $3;

-- name: MarkSessionRead :exec
UPDATE chat_sessions
SET user_unread_count = 0, updated_at = NOW()
WHERE id = $1 AND user_id = $2;
