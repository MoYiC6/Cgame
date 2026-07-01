package chat

import (
	"context"
	"fmt"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetOrCreateSession(ctx context.Context, userID, teacherID, orderID int64) (*ChatSession, error) {
	session, err := s.repo.GetSessionByUserTeacherOrder(ctx, userID, teacherID, orderID)
	if err == nil {
		return session, nil
	}
	if err.Error() == "get session: sql: no rows in result set" {
		session = &ChatSession{UserID: userID, TeacherID: &teacherID, OrderID: &orderID}
		if err := s.repo.CreateSession(ctx, session); err != nil {
			return nil, fmt.Errorf("create session: %w", err)
		}
		return session, nil
	}
	return nil, err
}

func (s *Service) GetUserSessions(ctx context.Context, userID int64) ([]*ChatSession, error) {
	return s.repo.GetUserSessions(ctx, userID)
}

func (s *Service) GetTeacherSessions(ctx context.Context, teacherUserID int64) ([]*ChatSession, error) {
	return s.repo.GetTeacherSessions(ctx, teacherUserID)
}

func (s *Service) SendMessage(ctx context.Context, sessionID, senderID int64, senderType string, content string, messageType *string, receiverID *int64) (*ChatMessage, error) {
	msg := &ChatMessage{
		SessionID:  sessionID,
		SenderID:   senderID,
		SenderType: senderType,
		ReceiverID: receiverID,
		Content:    content,
		MessageType: messageType,
	}
	if err := s.repo.CreateMessage(ctx, msg); err != nil {
		return nil, fmt.Errorf("send message: %w", err)
	}
	if senderType == "user" {
		_ = s.repo.IncrementTeacherUnread(ctx, sessionID)
	} else {
		_ = s.repo.IncrementUserUnread(ctx, sessionID)
	}
	_ = s.repo.UpdateSessionLastMessage(ctx, sessionID, content)
	return msg, nil
}

func (s *Service) GetSessionMessages(ctx context.Context, sessionID int64, page, pageSize int) ([]*ChatMessage, int, error) {
	return s.repo.GetSessionMessages(ctx, sessionID, page, pageSize)
}

func (s *Service) MarkSessionAsRead(ctx context.Context, sessionID, userID int64, isTeacher bool) error {
	if isTeacher {
		_, _ = s.repo.GetTeacherSessions(ctx, userID)
		_ = s.repo.MarkMessagesAsRead(ctx, sessionID, userID)
	} else {
		_ = s.repo.MarkMessagesAsRead(ctx, sessionID, userID)
	}
	return nil
}

func (s *Service) GetUnreadCount(ctx context.Context, userID int64, isTeacher bool) (int, error) {
	return s.repo.GetUnreadCount(ctx, userID, isTeacher)
}

func (s *Service) HasAccessToSession(ctx context.Context, sessionID, userID int64) (bool, error) {
	_, err := s.repo.GetSessionByUserTeacherOrder(ctx, userID, 0, 0)
	if err == nil {
		return true, nil
	}
	_, err = s.repo.GetTeacherSessions(ctx, userID)
	if err == nil {
		return true, nil
	}
	return false, nil
}
