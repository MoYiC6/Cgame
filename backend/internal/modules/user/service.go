package user

import (
	"context"
	"fmt"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
	if req.Username == "" || req.Email == "" || req.Password == "" {
		return nil, fmt.Errorf("username, email and password are required")
	}
	user := &User{
		Username: req.Username,
		Email:    NormalizeEmail(req.Email),
		Nickname: req.Nickname,
		RealName: req.RealName,
		Mobile:   req.Mobile,
		Status:   "active",
		IsTeacher: req.IsTeacher,
	}
	if req.Status != 0 {
		user.Status = mapStatusInt(req.Status)
	}
	// TODO: hash password properly using bcrypt
	user.PasswordHash = req.Password
	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) DeleteUser(ctx context.Context, userID int64) error {
	return s.repo.DeleteUser(ctx, userID)
}

func mapStatusInt(status int16) string {
	switch status {
	case 1:
		return "active"
	case 2:
		return "locked"
	default:
		return "disabled"
	}
}

func (s *Service) GetUser(ctx context.Context, userID int64) (*User, error) {
	return s.repo.GetByID(ctx, userID)
}

func (s *Service) CreateBalanceLog(ctx context.Context, log *UserBalanceLog) error {
	if log.UserID == 0 || log.ChangeType == "" {
		return fmt.Errorf("user_id and change_type are required")
	}
	return s.repo.CreateBalanceLog(ctx, log)
}

func (s *Service) GetBalanceLogs(ctx context.Context, userID int64, page, pageSize int) ([]*UserBalanceLog, int, error) {
	return s.repo.GetUserBalanceLogs(ctx, userID, page, pageSize)
}

func (s *Service) GetUserLevel(ctx context.Context, userID int64) (*UserLevel, error) {
	return s.repo.GetUserLevel(ctx, userID)
}

func (s *Service) CreateUserLevelLog(ctx context.Context, log *UserLevelLog) error {
	if log.UserID == 0 || log.NewLevelID == nil {
		return fmt.Errorf("user_id and new_level_id are required")
	}
	return s.repo.CreateUserLevelLog(ctx, log)
}

func (s *Service) CreatePurchaseRecord(ctx context.Context, record *UserPurchaseRecord) error {
	if record.UserID == 0 || record.GoodsID == nil {
		return fmt.Errorf("user_id and goods_id are required")
	}
	return s.repo.CreatePurchaseRecord(ctx, record)
}

func (s *Service) GetPurchaseCount(ctx context.Context, userID, goodsID int64) (int, error) {
	return s.repo.GetUserPurchaseCount(ctx, userID, goodsID)
}

func (s *Service) GetUserCenterInfo(ctx context.Context, userID int64) (*UserCenterInfo, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}
	info := &UserCenterInfo{
		ID:               user.ID,
		Username:         user.Username,
		Nickname:         user.Nickname,
		Avatar:           user.Avatar,
		Mobile:           user.Mobile,
		Email:            user.Email,
		Gender:           user.Gender,
		Birthday:         user.Birthday,
		Province:         user.Province,
		City:             user.City,
		District:         user.District,
		Intro:            user.Intro,
		Balance:          user.Balance,
		FrozenBalance:    user.FrozenBalance,
		TotalRecharge:    user.TotalRecharge,
		TotalConsumption: user.TotalConsumption,
		LevelID:          user.LevelID,
		IsTeacher:        user.IsTeacher,
		RealNameStatus:   user.RealNameStatus,
		Status:           user.Status,
	}
	if user.LevelID != nil {
		level, err := s.repo.GetUserLevelByID(ctx, *user.LevelID)
		if err == nil && level != nil {
			info.LevelName = level.Name
		}
	}
	return info, nil
}

func (s *Service) UpdateProfile(ctx context.Context, userID int64, req *UpdateProfileRequest) error {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}
	user.Nickname = req.Nickname
	user.Avatar = req.Avatar
	user.Gender = req.Gender
	user.Birthday = req.Birthday
	user.Province = req.Province
	user.City = req.City
	user.District = req.District
	user.Intro = req.Intro
	return s.repo.UpdateUser(ctx, user)
}

func (s *Service) UpdateUserStatus(ctx context.Context, userID int64, status int16) error {
	return s.repo.UpdateUserStatus(ctx, userID, status)
}

func (s *Service) ListUsers(ctx context.Context, query UserQuery) ([]*User, int, error) {
	if query.PageNum <= 0 {
		query.PageNum = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	return s.repo.ListUsers(ctx, query)
}

func (s *Service) GetUserByID(ctx context.Context, userID int64) (*User, error) {
	return s.repo.GetByID(ctx, userID)
}

func (s *Service) GetUserLevels(ctx context.Context) ([]*UserLevel, error) {
	return s.repo.GetUserLevels(ctx)
}

func (s *Service) GetRecentBalanceLogs(ctx context.Context, userID int64, limit int) ([]*UserBalanceLog, error) {
	return s.repo.GetRecentBalanceLogs(ctx, userID, limit)
}

func (s *Service) GetConsumptionRanking(ctx context.Context, limit int) ([]*ConsumptionRankingItem, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.repo.GetConsumptionRanking(ctx, limit)
}

func (s *Service) ListUserSelectors(ctx context.Context, keyword string, limit int) ([]*UserSelectorItem, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.repo.ListUserSelectors(ctx, keyword, limit)
}

func (s *Service) ListUserLoginLogs(ctx context.Context, userID *int64, page, pageSize int) ([]*UserLoginLog, int, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	return s.repo.ListUserLoginLogs(ctx, userID, page, pageSize)
}

func (s *Service) DeleteUserLoginLogs(ctx context.Context, ids []int64) error {
	return s.repo.DeleteUserLoginLogs(ctx, ids)
}
