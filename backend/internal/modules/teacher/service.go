package teacher

import (
	"context"
	"fmt"
	"net/http"

	apperrors "backend/internal/platform/errors"
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

func (s *Service) GetTeacherByID(ctx context.Context, id int64) (*Teacher, error) {
	return s.repo.GetTeacherByID(ctx, id)
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

// UpdateOnlineStatus updates a teacher's online status
func (s *Service) UpdateOnlineStatus(ctx context.Context, userID int64, status int) error {
	if userID == 0 {
		return apperrors.New(apperrors.CodeInvalidArgument, "user_id is required", http.StatusBadRequest, nil)
	}
	teacher, err := s.repo.GetTeacherByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if teacher == nil {
		return apperrors.New(apperrors.CodeNotFound, "teacher not found", http.StatusNotFound, nil)
	}
	return s.repo.UpdateTeacherStatus(ctx, teacher.ID, status)
}

// GetCurrentStatus returns the current teacher status
func (s *Service) GetCurrentStatus(ctx context.Context, userID int64) (int, error) {
	if userID == 0 {
		return 0, apperrors.New(apperrors.CodeInvalidArgument, "user_id is required", http.StatusBadRequest, nil)
	}
	teacher, err := s.repo.GetTeacherByUserID(ctx, userID)
	if err != nil {
		return 0, err
	}
	if teacher == nil {
		return 0, apperrors.New(apperrors.CodeNotFound, "teacher not found", http.StatusNotFound, nil)
	}
	if teacher.Status == nil {
		return 0, nil
	}
	return *teacher.Status, nil
}

// TeacherHeartbeat records a heartbeat from a teacher
func (s *Service) TeacherHeartbeat(ctx context.Context, userID int64) error {
	if userID == 0 {
		return apperrors.New(apperrors.CodeInvalidArgument, "user_id is required", http.StatusBadRequest, nil)
	}
	teacher, err := s.repo.GetTeacherByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if teacher == nil {
		return apperrors.New(apperrors.CodeNotFound, "teacher not found", http.StatusNotFound, nil)
	}
	// Update heartbeat timestamp - stored in a simple way via updated_at
	return s.repo.UpdateTeacherHeartbeat(ctx, teacher.ID)
}

// SetAutoStatus sets auto status configuration for a teacher
func (s *Service) SetAutoStatus(ctx context.Context, userID int64, enabled bool, onlineTime, offlineTime *string) error {
	if userID == 0 {
		return apperrors.New(apperrors.CodeInvalidArgument, "user_id is required", http.StatusBadRequest, nil)
	}
	teacher, err := s.repo.GetTeacherByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if teacher == nil {
		return apperrors.New(apperrors.CodeNotFound, "teacher not found", http.StatusNotFound, nil)
	}
	return s.repo.UpdateTeacherAutoStatus(ctx, teacher.ID, enabled, onlineTime, offlineTime)
}

// GetAutoStatus gets auto status configuration for a teacher
func (s *Service) GetAutoStatus(ctx context.Context, userID int64) (*TeacherAutoStatus, error) {
	if userID == 0 {
		return nil, apperrors.New(apperrors.CodeInvalidArgument, "user_id is required", http.StatusBadRequest, nil)
	}
	teacher, err := s.repo.GetTeacherByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if teacher == nil {
		return nil, apperrors.New(apperrors.CodeNotFound, "teacher not found", http.StatusNotFound, nil)
	}
	return &TeacherAutoStatus{
		TeacherID: teacher.ID,
		Enabled:   teacher.AutoStatusEnabled,
	}, nil
}

// ApplyTeacher creates a teacher application
func (s *Service) ApplyTeacher(ctx context.Context, app *TeacherApplication) (int64, error) {
	if app.UserID == 0 {
		return 0, apperrors.New(apperrors.CodeInvalidArgument, "user_id is required", http.StatusBadRequest, nil)
	}
	if app.Name == "" {
		return 0, apperrors.New(apperrors.CodeInvalidArgument, "name is required", http.StatusBadRequest, nil)
	}
	// Check if already a teacher
	existing, err := s.repo.GetTeacherByUserID(ctx, app.UserID)
	if err != nil {
		return 0, err
	}
	if existing != nil {
		return 0, apperrors.New(apperrors.CodeConflict, "already a teacher", http.StatusConflict, nil)
	}
	return s.repo.CreateTeacherApplication(ctx, app)
}

// GetTeacherRanking returns teacher ranking
func (s *Service) GetTeacherRanking(ctx context.Context, limit int) ([]*TeacherRanking, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.repo.GetTeacherRanking(ctx, limit)
}

// GetTeacherDashboardStats returns dashboard stats for a teacher
func (s *Service) GetTeacherDashboardStats(ctx context.Context, userID int64) (*TeacherDashboardStats, error) {
	if userID == 0 {
		return nil, apperrors.New(apperrors.CodeInvalidArgument, "user_id is required", http.StatusBadRequest, nil)
	}
	teacher, err := s.repo.GetTeacherByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if teacher == nil {
		return nil, apperrors.New(apperrors.CodeNotFound, "teacher not found", http.StatusNotFound, nil)
	}
	return s.repo.GetTeacherDashboardStats(ctx, teacher.ID)
}

// UpdateTeacherStatus updates a teacher's status (admin)
func (s *Service) UpdateTeacherStatus(ctx context.Context, teacherID int64, newStatus int, reason *string, operatorID int64) error {
	if teacherID == 0 {
		return apperrors.New(apperrors.CodeInvalidArgument, "teacher_id is required", http.StatusBadRequest, nil)
	}
	teacher, err := s.repo.GetTeacherByID(ctx, teacherID)
	if err != nil {
		return err
	}
	if teacher == nil {
		return apperrors.New(apperrors.CodeNotFound, "teacher not found", http.StatusNotFound, nil)
	}
	oldStatus := 0
	if teacher.Status != nil {
		oldStatus = *teacher.Status
	}
	if err := s.repo.UpdateTeacherStatus(ctx, teacherID, newStatus); err != nil {
		return err
	}
	// Log status change
	log := &TeacherStatusLog{
		TeacherID:  teacherID,
		OldStatus:  &oldStatus,
		NewStatus:  &newStatus,
		Reason:     reason,
		OperatorID: &operatorID,
	}
	return s.repo.CreateTeacherStatusLog(ctx, log)
}

// BatchUpdateTeacherStatus batch updates teacher statuses
func (s *Service) BatchUpdateTeacherStatus(ctx context.Context, ids []int64, newStatus int, reason *string, operatorID int64) error {
	if len(ids) == 0 {
		return apperrors.New(apperrors.CodeInvalidArgument, "ids are required", http.StatusBadRequest, nil)
	}
	for _, id := range ids {
		if err := s.UpdateTeacherStatus(ctx, id, newStatus, reason, operatorID); err != nil {
			return err
		}
	}
	return nil
}

// GetTeacherStatusLogs returns status logs for a teacher
func (s *Service) GetTeacherStatusLogs(ctx context.Context, teacherID int64, page, pageSize int) ([]*TeacherStatusLog, int, error) {
	if teacherID == 0 {
		return nil, 0, apperrors.New(apperrors.CodeInvalidArgument, "teacher_id is required", http.StatusBadRequest, nil)
	}
	return s.repo.GetTeacherStatusLogs(ctx, teacherID, page, pageSize)
}

// VerifyTeacher verifies a teacher
func (s *Service) VerifyTeacher(ctx context.Context, teacherID int64) error {
	if teacherID == 0 {
		return apperrors.New(apperrors.CodeInvalidArgument, "teacher_id is required", http.StatusBadRequest, nil)
	}
	teacher, err := s.repo.GetTeacherByID(ctx, teacherID)
	if err != nil {
		return err
	}
	if teacher == nil {
		return apperrors.New(apperrors.CodeNotFound, "teacher not found", http.StatusNotFound, nil)
	}
	status := 1 // verified/active
	return s.repo.UpdateTeacherStatus(ctx, teacherID, status)
}

// ListApplications returns teacher applications
func (s *Service) ListApplications(ctx context.Context, status *int, page, pageSize int) ([]*TeacherApplication, int, error) {
	return s.repo.ListTeacherApplications(ctx, status, page, pageSize)
}

// ApproveApplication approves a teacher application
func (s *Service) ApproveApplication(ctx context.Context, appID int64, operatorID int64) error {
	if appID == 0 {
		return apperrors.New(apperrors.CodeInvalidArgument, "application_id is required", http.StatusBadRequest, nil)
	}
	app, err := s.repo.GetTeacherApplicationByID(ctx, appID)
	if err != nil {
		return err
	}
	if app == nil {
		return apperrors.New(apperrors.CodeNotFound, "application not found", http.StatusNotFound, nil)
	}
	if app.Status != 0 {
		return apperrors.New(apperrors.CodePreconditionFailed, "application already processed", http.StatusPreconditionFailed, nil)
	}
	// Update application status
	if err := s.repo.UpdateApplicationStatus(ctx, appID, 1, nil, &operatorID); err != nil {
		return err
	}
	// Create teacher from application
	status := 1
	teacher := &Teacher{
		UserID:    app.UserID,
		Name:      &app.Name,
		Mobile:    app.Mobile,
		Avatar:    app.Avatar,
		Status:    &status,
		Platforms: app.Platforms,
		Tags:      app.Tags,
	}
	err = s.repo.CreateTeacher(ctx, teacher)
	return err
}

// RejectApplication rejects a teacher application
func (s *Service) RejectApplication(ctx context.Context, appID int64, reason string, operatorID int64) error {
	if appID == 0 {
		return apperrors.New(apperrors.CodeInvalidArgument, "application_id is required", http.StatusBadRequest, nil)
	}
	app, err := s.repo.GetTeacherApplicationByID(ctx, appID)
	if err != nil {
		return err
	}
	if app == nil {
		return apperrors.New(apperrors.CodeNotFound, "application not found", http.StatusNotFound, nil)
	}
	if app.Status != 0 {
		return apperrors.New(apperrors.CodePreconditionFailed, "application already processed", http.StatusPreconditionFailed, nil)
	}
	return s.repo.UpdateApplicationStatus(ctx, appID, 2, &reason, &operatorID)
}

// UpdateTeacherLevel updates a teacher level
func (s *Service) UpdateTeacherLevel(ctx context.Context, level *TeacherLevel) error {
	if level.ID == 0 {
		return apperrors.New(apperrors.CodeInvalidArgument, "level_id is required", http.StatusBadRequest, nil)
	}
	if level.Name == "" {
		return apperrors.New(apperrors.CodeInvalidArgument, "name is required", http.StatusBadRequest, nil)
	}
	return s.repo.UpdateTeacherLevel(ctx, level)
}

// DeleteTeacherLevel deletes a teacher level
func (s *Service) DeleteTeacherLevel(ctx context.Context, id int64) error {
	if id == 0 {
		return apperrors.New(apperrors.CodeInvalidArgument, "level_id is required", http.StatusBadRequest, nil)
	}
	return s.repo.DeleteTeacherLevel(ctx, id)
}

// GetLevelGoods returns goods for a level
func (s *Service) GetLevelGoods(ctx context.Context, levelID int64) ([]*TeacherLevelGoods, error) {
	if levelID == 0 {
		return nil, apperrors.New(apperrors.CodeInvalidArgument, "level_id is required", http.StatusBadRequest, nil)
	}
	return s.repo.GetTeacherLevelGoods(ctx, levelID)
}

// UpdateLevelGoods updates goods for a level
func (s *Service) UpdateLevelGoods(ctx context.Context, levelID int64, goods []*TeacherLevelGoods) error {
	if levelID == 0 {
		return apperrors.New(apperrors.CodeInvalidArgument, "level_id is required", http.StatusBadRequest, nil)
	}
	return s.repo.UpdateTeacherLevelGoods(ctx, levelID, goods)
}
