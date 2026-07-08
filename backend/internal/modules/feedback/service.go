package feedback

import (
	"context"
	"fmt"
	"strings"
	"time"

	"backend/internal/platform/database"
)

type Repository interface {
	CreateFeedback(ctx context.Context, feedback *Feedback) error
	ListUserFeedback(ctx context.Context, userID int64, page, pageSize int) (*FeedbackPage, error)
	GetFeedbackDetail(ctx context.Context, userID *int64, id int64) (*FeedbackDetailVO, error)
	ListAdminFeedback(ctx context.Context, page, pageSize int, status *int, keyword string) (*FeedbackPage, error)
	CreateReply(ctx context.Context, reply *FeedbackReply) error
	UpdateStatus(ctx context.Context, id int64, status int) error
	DeleteFeedback(ctx context.Context, id int64) error
}

type Service struct {
	repo      Repository
	txManager database.TxManager
}

func NewService(repo Repository, txManager database.TxManager) *Service {
	if txManager == nil {
		txManager = database.NoopTxManager{}
	}
	return &Service{repo: repo, txManager: txManager}
}

func (s *Service) Submit(ctx context.Context, userID int64, req SubmitRequest) (int64, error) {
	content := strings.TrimSpace(req.Content)
	if userID == 0 {
		return 0, fmt.Errorf("user id is required")
	}
	if len([]rune(content)) < 10 || len([]rune(content)) > 500 {
		return 0, fmt.Errorf("feedback content length must be 10-500 characters")
	}
	if len(req.Images) > 3 {
		return 0, fmt.Errorf("feedback images cannot exceed 3")
	}
	feedback := &Feedback{
		UserID:   userID,
		TicketNo: fmt.Sprintf("FB%s%06d", time.Now().Format("20060102150405"), userID%1000000),
		Content:  content,
		Images:   req.Images,
		Status:   FeedbackStatusPending,
	}
	if err := s.repo.CreateFeedback(ctx, feedback); err != nil {
		return 0, fmt.Errorf("create feedback: %w", err)
	}
	return feedback.ID, nil
}

func (s *Service) ListMine(ctx context.Context, userID int64, page, pageSize int) (*FeedbackPage, error) {
	if userID == 0 {
		return nil, fmt.Errorf("user id is required")
	}
	return s.repo.ListUserFeedback(ctx, userID, normalizePage(page), normalizePageSize(pageSize))
}

func (s *Service) GetMine(ctx context.Context, userID, id int64) (*FeedbackDetailVO, error) {
	if userID == 0 || id == 0 {
		return nil, fmt.Errorf("user id and feedback id are required")
	}
	return s.repo.GetFeedbackDetail(ctx, &userID, id)
}

func (s *Service) ListAdmin(ctx context.Context, page, pageSize int, status *int, keyword string) (*FeedbackPage, error) {
	return s.repo.ListAdminFeedback(ctx, normalizePage(page), normalizePageSize(pageSize), status, strings.TrimSpace(keyword))
}

func (s *Service) GetAdmin(ctx context.Context, id int64) (*FeedbackDetailVO, error) {
	if id == 0 {
		return nil, fmt.Errorf("feedback id is required")
	}
	return s.repo.GetFeedbackDetail(ctx, nil, id)
}

func (s *Service) Reply(ctx context.Context, adminUserID, feedbackID int64, content string) (int64, error) {
	content = strings.TrimSpace(content)
	if adminUserID == 0 || feedbackID == 0 {
		return 0, fmt.Errorf("admin user id and feedback id are required")
	}
	if content == "" || len([]rune(content)) > 500 {
		return 0, fmt.Errorf("reply content length must be 1-500 characters")
	}
	reply := &FeedbackReply{FeedbackID: feedbackID, ReplyUserID: adminUserID, ReplyType: ReplyTypeAdmin, Content: content}
	if err := s.repo.CreateReply(ctx, reply); err != nil {
		return 0, fmt.Errorf("create feedback reply: %w", err)
	}
	return reply.ID, nil
}

func (s *Service) UpdateStatus(ctx context.Context, id int64, status int) error {
	if id == 0 {
		return fmt.Errorf("feedback id is required")
	}
	if status < FeedbackStatusPending || status > FeedbackStatusResolved {
		return fmt.Errorf("invalid feedback status")
	}
	return s.repo.UpdateStatus(ctx, id, status)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id == 0 {
		return fmt.Errorf("feedback id is required")
	}
	return s.repo.DeleteFeedback(ctx, id)
}

func normalizePage(page int) int {
	if page <= 0 {
		return 1
	}
	return page
}

func normalizePageSize(pageSize int) int {
	if pageSize <= 0 {
		return 10
	}
	return pageSize
}
