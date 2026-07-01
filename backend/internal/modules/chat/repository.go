package chat

import (
	"context"
	"encoding/json"
	"fmt"

	"backend/internal/platform/database"
)

type Repository struct {
	dbtx database.DBTX
}

func NewRepository(dbtx database.DBTX) *Repository {
	return &Repository{dbtx: dbtx}
}

func (r *Repository) GetSessionByUserTeacherOrder(ctx context.Context, userID, teacherID, orderID int64) (*ChatSession, error) {
	row := r.dbtx.QueryRowContext(ctx,
		`SELECT id, user_id, teacher_id, teacher_user_id, order_id, last_message_id, last_message_content, last_message_time, user_unread_count, teacher_unread_count, status, created_at, updated_at
		 FROM chat_sessions WHERE user_id = $1 AND teacher_id = $2 AND order_id = $3`,
		userID, teacherID, orderID,
	)
	var s ChatSession
	err := row.Scan(&s.ID, &s.UserID, &s.TeacherID, &s.TeacherUserID, &s.OrderID, &s.LastMessageID, &s.LastMessageContent, &s.LastMessageTime, &s.UserUnreadCount, &s.TeacherUnreadCount, &s.Status, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}
	return &s, nil
}

func (r *Repository) CreateSession(ctx context.Context, s *ChatSession) error {
	return r.dbtx.QueryRowContext(ctx,
		`INSERT INTO chat_sessions (user_id, teacher_id, teacher_user_id, order_id) VALUES ($1, $2, $3, $4) RETURNING id`,
		s.UserID, s.TeacherID, s.TeacherUserID, s.OrderID,
	).Scan(&s.ID)
}

func (r *Repository) GetUserSessions(ctx context.Context, userID int64) ([]*ChatSession, error) {
	rows, err := r.dbtx.QueryContext(ctx,
		`SELECT id, user_id, teacher_id, teacher_user_id, order_id, last_message_id, last_message_content, last_message_time, user_unread_count, teacher_unread_count, status, created_at, updated_at
		 FROM chat_sessions WHERE user_id = $1 AND status = 0 ORDER BY last_message_time DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("get user sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*ChatSession
	for rows.Next() {
		var s ChatSession
		if err := rows.Scan(&s.ID, &s.UserID, &s.TeacherID, &s.TeacherUserID, &s.OrderID, &s.LastMessageID, &s.LastMessageContent, &s.LastMessageTime, &s.UserUnreadCount, &s.TeacherUnreadCount, &s.Status, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan session: %w", err)
		}
		sessions = append(sessions, &s)
	}
	return sessions, nil
}

func (r *Repository) GetTeacherSessions(ctx context.Context, teacherUserID int64) ([]*ChatSession, error) {
	rows, err := r.dbtx.QueryContext(ctx,
		`SELECT id, user_id, teacher_id, teacher_user_id, order_id, last_message_id, last_message_content, last_message_time, user_unread_count, teacher_unread_count, status, created_at, updated_at
		 FROM chat_sessions WHERE teacher_user_id = $1 AND status = 0 ORDER BY last_message_time DESC`,
		teacherUserID,
	)
	if err != nil {
		return nil, fmt.Errorf("get teacher sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*ChatSession
	for rows.Next() {
		var s ChatSession
		if err := rows.Scan(&s.ID, &s.UserID, &s.TeacherID, &s.TeacherUserID, &s.OrderID, &s.LastMessageID, &s.LastMessageContent, &s.LastMessageTime, &s.UserUnreadCount, &s.TeacherUnreadCount, &s.Status, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan session: %w", err)
		}
		sessions = append(sessions, &s)
	}
	return sessions, nil
}

func (r *Repository) CreateMessage(ctx context.Context, m *ChatMessage) error {
	extraDataJSON, _ := json.Marshal(m.ExtraData)
	return r.dbtx.QueryRowContext(ctx,
		`INSERT INTO chat_messages (session_id, sender_id, sender_type, receiver_id, content, message_type, extra_data)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		m.SessionID, m.SenderID, m.SenderType, m.ReceiverID, m.Content, m.MessageType, string(extraDataJSON),
	).Scan(&m.ID)
}

func (r *Repository) GetSessionMessages(ctx context.Context, sessionID int64, page, pageSize int) ([]*ChatMessage, int, error) {
	countQuery := "SELECT COUNT(*) FROM chat_messages WHERE session_id = $1 AND status = 0"
	var total int
	if err := r.dbtx.QueryRowContext(ctx, countQuery, sessionID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count messages: %w", err)
	}

	query := `SELECT id, session_id, sender_id, sender_type, receiver_id, content, message_type, extra_data, is_read, read_time, status, created_at
			  FROM chat_messages WHERE session_id = $1 AND status = 0 ORDER BY created_at ASC LIMIT $2 OFFSET $3`
	rows, err := r.dbtx.QueryContext(ctx, query, sessionID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("get messages: %w", err)
	}
	defer rows.Close()

	var messages []*ChatMessage
	for rows.Next() {
		var m ChatMessage
		var extraDataJSON []byte
		err := rows.Scan(&m.ID, &m.SessionID, &m.SenderID, &m.SenderType, &m.ReceiverID, &m.Content, &m.MessageType, &extraDataJSON, &m.IsRead, &m.ReadTime, &m.Status, &m.CreatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("scan message: %w", err)
		}
		if len(extraDataJSON) > 0 {
			json.Unmarshal(extraDataJSON, &m.ExtraData)
		}
		messages = append(messages, &m)
	}
	return messages, total, nil
}

func (r *Repository) MarkMessagesAsRead(ctx context.Context, sessionID, receiverID int64) error {
	_, err := r.dbtx.ExecContext(ctx,
		`UPDATE chat_messages SET is_read = 1, read_time = NOW() WHERE session_id = $1 AND receiver_id = $2 AND is_read = 0 AND status = 0`,
		sessionID, receiverID,
	)
	if err != nil {
		return fmt.Errorf("mark as read: %w", err)
	}
	return nil
}

func (r *Repository) GetUnreadCount(ctx context.Context, userID int64, isTeacher bool) (int, error) {
	var count int
	if isTeacher {
		err := r.dbtx.QueryRowContext(ctx,
			"SELECT COALESCE(SUM(teacher_unread_count), 0) FROM chat_sessions WHERE teacher_user_id = $1 AND status = 0",
			userID,
		).Scan(&count)
		if err != nil {
			return 0, fmt.Errorf("get teacher unread count: %w", err)
		}
	} else {
		err := r.dbtx.QueryRowContext(ctx,
			"SELECT COALESCE(SUM(user_unread_count), 0) FROM chat_sessions WHERE user_id = $1 AND status = 0",
			userID,
		).Scan(&count)
		if err != nil {
			return 0, fmt.Errorf("get user unread count: %w", err)
		}
	}
	return count, nil
}

func (r *Repository) IncrementUserUnread(ctx context.Context, sessionID int64) error {
	_, err := r.dbtx.ExecContext(ctx,
		`UPDATE chat_sessions SET user_unread_count = user_unread_count + 1, updated_at = NOW() WHERE id = $1`,
		sessionID,
	)
	return err
}

func (r *Repository) IncrementTeacherUnread(ctx context.Context, sessionID int64) error {
	_, err := r.dbtx.ExecContext(ctx,
		`UPDATE chat_sessions SET teacher_unread_count = teacher_unread_count + 1, updated_at = NOW() WHERE id = $1`,
		sessionID,
	)
	return err
}

func (r *Repository) UpdateSessionLastMessage(ctx context.Context, sessionID int64, content string) error {
	truncated := content
	if len(truncated) > 50 {
		truncated = truncated[:50]
	}
	_, err := r.dbtx.ExecContext(ctx,
		`UPDATE chat_sessions SET last_message_content = $1, last_message_time = NOW(), updated_at = NOW() WHERE id = $2`,
		truncated, sessionID,
	)
	return err
}
