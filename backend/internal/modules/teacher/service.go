package teacher

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

func (s *Service) CreateTeacher(ctx context.Context, t *Teacher) (int64, error) {
	if t.UserID == 0 {
		return 0, fmt.Errorf("user_id is required")
	}
	if err := s.repo.CreateTeacher(ctx, t); err != nil {
		return 0, err
	}
	return t.ID, nil
}

func (s *Service) GetTeacher(ctx context.Context, userID int64) (*Teacher, error) {
	return s.repo.GetTeacherByUserID(ctx, userID)
}

func (s *Service) ListTeachers(ctx context.Context, page, pageSize int) ([]*Teacher, int, error) {
	return s.repo.ListTeachers(ctx, page, pageSize)
}

func (s *Service) CreateTeacherLevel(ctx context.Context, level *TeacherLevel) (int64, error) {
	if level.Name == "" {
		return 0, fmt.Errorf("name is required")
	}
	if level.Status == nil {
		one := 1
		level.Status = &one
	}
	if err := s.repo.CreateTeacherLevel(ctx, level); err != nil {
		return 0, err
	}
	return level.ID, nil
}

func (s *Service) GetTeacherLevels(ctx context.Context) ([]*TeacherLevel, error) {
	return s.repo.GetTeacherLevels(ctx)
}

func (s *Service) LogStatusChange(ctx context.Context, log *TeacherStatusLog) error {
	if log.TeacherID == 0 || log.NewStatus == nil {
		return fmt.Errorf("teacher_id and new_status are required")
	}
	return s.repo.CreateTeacherStatusLog(ctx, log)
}

func (s *Service) CreateBalanceLog(ctx context.Context, log *TeacherBalanceLog) error {
	if log.TeacherID == 0 || log.ChangeType == "" {
		return fmt.Errorf("teacher_id and change_type are required")
	}
	return s.repo.CreateTeacherBalanceLog(ctx, log)
}
