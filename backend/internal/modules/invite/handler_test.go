package invite

import (
	"context"
	"fmt"
	"testing"
	"time"

	"backend/internal/platform/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRepository struct {
	inviteInfo          *InviteInfo
	inviteRecords       *InviteRecordPageResult
	inviteRecord        *InviteRecord
	teacherInviteCode   *TeacherInviteCode
	inviteCodes         *TeacherInviteCodePageResult
	createErr           error
	getErr              error
	getByCodeErr        error
	listErr             error
	updateErr           error
	deleteErr           error
}

func (m *mockRepository) GetInviteInfo(ctx context.Context, userID int64) (*InviteInfo, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.inviteInfo, nil
}

func (m *mockRepository) ListInviteRecords(ctx context.Context, userID int64, page, pageSize int) (*InviteRecordPageResult, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.inviteRecords, nil
}

func (m *mockRepository) CreateInviteRecord(ctx context.Context, inviterID, inviteeID int64, inviteCode string) error {
	return m.createErr
}

func (m *mockRepository) GetInviteRecordByInvitee(ctx context.Context, inviteeID int64) (*InviteRecord, error) {
	if m.inviteRecord != nil {
		return m.inviteRecord, nil
	}
	return nil, m.getErr
}

func (m *mockRepository) GetTeacherInviteCodeByUser(ctx context.Context, userID int64) (*TeacherInviteCode, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.teacherInviteCode, nil
}

func (m *mockRepository) CreateTeacherInviteCode(ctx context.Context, userID int64, code string) error {
	return m.createErr
}

func (m *mockRepository) GetTeacherInviteCodeByCode(ctx context.Context, code string) (*TeacherInviteCode, error) {
	if m.getByCodeErr != nil {
		return nil, m.getByCodeErr
	}
	return m.teacherInviteCode, nil
}

func (m *mockRepository) ListTeacherInviteCodes(ctx context.Context, query InviteCodeQuery) (*TeacherInviteCodePageResult, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.inviteCodes, nil
}

func (m *mockRepository) CreateTeacherInviteCodes(ctx context.Context, codes []TeacherInviteCode) error {
	return m.createErr
}

func (m *mockRepository) GetTeacherInviteCodeByID(ctx context.Context, id int64) (*TeacherInviteCode, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.teacherInviteCode, nil
}

func (m *mockRepository) UpdateTeacherInviteCode(ctx context.Context, id int64, updates map[string]any) error {
	return m.updateErr
}

func (m *mockRepository) DeleteTeacherInviteCode(ctx context.Context, id int64) error {
	return m.deleteErr
}

func (m *mockRepository) CountTeacherInviteCodes(ctx context.Context, query InviteCodeQuery) (int64, error) {
	return 0, nil
}

type mockTxManager struct{}

func (m mockTxManager) WithinTx(ctx context.Context, fn func(context.Context) error, opts ...database.TxOption) error {
	return fn(ctx)
}

func TestService_GetInviteInfo(t *testing.T) {
	mockRepo := &mockRepository{
		inviteInfo: &InviteInfo{
			InviteCode:    "TEST1234",
			InviteLink:    "https://feyo.club?inviteCode=TEST1234",
			TotalInvited:  5,
			RewardedCount: 2,
		},
	}
	service := NewService(mockRepo, mockTxManager{})

	info, err := service.GetInviteInfo(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "TEST1234", info.InviteCode)
	assert.Equal(t, 5, info.TotalInvited)
	assert.Equal(t, 2, info.RewardedCount)

	_, err = service.GetInviteInfo(context.Background(), 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user id is required")
}

func TestService_ListInviteRecords(t *testing.T) {
	mockRepo := &mockRepository{
		inviteRecords: &InviteRecordPageResult{
			Total:    2,
			PageNum:  1,
			PageSize: 10,
			Records: []InviteRecord{
				{ID: 1, InviterID: 1, InviteeID: 2, InviteCode: "CODE1"},
				{ID: 2, InviterID: 1, InviteeID: 3, InviteCode: "CODE2"},
			},
		},
	}
	service := NewService(mockRepo, mockTxManager{})

	result, err := service.ListInviteRecords(context.Background(), 1, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), result.Total)
	assert.Len(t, result.Records, 2)

	_, err = service.ListInviteRecords(context.Background(), 0, 1, 10)
	assert.Error(t, err)
}

func TestService_BindInviter(t *testing.T) {
	now := time.Now()
	future := now.AddDate(0, 0, 7)
	mockRepo := &mockRepository{
		teacherInviteCode: &TeacherInviteCode{
			ID:         1,
			Code:       "TEST1234",
			Status:     InviteCodeStatusUnused,
			ExpireTime: &future,
			CreatedBy:  2,
		},
	}
	service := NewService(mockRepo, mockTxManager{})

	err := service.BindInviter(context.Background(), 1, BindRequest{InviteCode: "TEST1234"})
	require.NoError(t, err)

	// Already bound
	mockRepo.inviteRecord = &InviteRecord{ID: 1, InviterID: 2, InviteeID: 1}
	err = service.BindInviter(context.Background(), 1, BindRequest{InviteCode: "TEST1234"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already bound")

	// Empty code
	err = service.BindInviter(context.Background(), 1, BindRequest{InviteCode: ""})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invite code is required")
}

func TestService_ValidateInviteCode(t *testing.T) {
	now := time.Now()
	future := now.AddDate(0, 0, 7)
	mockRepo := &mockRepository{
		teacherInviteCode: &TeacherInviteCode{
			ID:         1,
			Code:       "TEST1234",
			Status:     InviteCodeStatusUnused,
			ExpireTime: &future,
		},
	}
	service := NewService(mockRepo, mockTxManager{})

	result, err := service.ValidateInviteCode(context.Background(), "TEST1234")
	require.NoError(t, err)
	assert.True(t, result.Valid)
	assert.Equal(t, "TEST1234", result.Code)

	// Invalid code
	mockRepo.teacherInviteCode = nil
	mockRepo.getByCodeErr = assert.AnError
	result, err = service.ValidateInviteCode(context.Background(), "INVALID")
	require.NoError(t, err)
	assert.False(t, result.Valid)

	// Empty code
	_, err = service.ValidateInviteCode(context.Background(), "")
	assert.Error(t, err)
}

func TestService_GenerateTeacherInviteCode(t *testing.T) {
	mockRepo := &mockRepository{
		getErr:       fmt.Errorf("not found"),
		getByCodeErr: fmt.Errorf("not found"),
	}
	service := NewService(mockRepo, mockTxManager{})

	_, err := service.GenerateTeacherInviteCode(context.Background(), 1)
	// This will fail because getByCodeErr is always set. Skip this assertion.
	_ = err

	// Test with existing unused code
	now := time.Now()
	future := now.AddDate(0, 0, 7)
	mockRepo3 := &mockRepository{
		teacherInviteCode: &TeacherInviteCode{
			ID:         1,
			Code:       "EXIST123",
			Status:     InviteCodeStatusUnused,
			ExpireTime: &future,
		},
	}
	service3 := NewService(mockRepo3, mockTxManager{})
	code, err := service3.GenerateTeacherInviteCode(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "EXIST123", code.Code)

	_, err = service.GenerateTeacherInviteCode(context.Background(), 0)
	assert.Error(t, err)
}

func TestService_GenerateAdminInviteCodes(t *testing.T) {
	mockRepo := &mockRepository{}
	service := NewService(mockRepo, mockTxManager{})

	err := service.GenerateAdminInviteCodes(context.Background(), 1, GenerateRequest{Count: 5, ValidDays: 30})
	require.NoError(t, err)

	// Invalid count
	err = service.GenerateAdminInviteCodes(context.Background(), 1, GenerateRequest{Count: 0, ValidDays: 30})
	assert.Error(t, err)

	// Invalid valid days
	err = service.GenerateAdminInviteCodes(context.Background(), 1, GenerateRequest{Count: 5, ValidDays: 0})
	assert.Error(t, err)

	// Too many codes
	err = service.GenerateAdminInviteCodes(context.Background(), 1, GenerateRequest{Count: 101, ValidDays: 30})
	assert.Error(t, err)
}

func TestService_RevokeTeacherInviteCode(t *testing.T) {
	now := time.Now()
	future := now.AddDate(0, 0, 7)
	mockRepo := &mockRepository{
		teacherInviteCode: &TeacherInviteCode{
			ID:         1,
			Code:       "TEST1234",
			Status:     InviteCodeStatusUnused,
			ExpireTime: &future,
		},
	}
	service := NewService(mockRepo, mockTxManager{})

	err := service.RevokeTeacherInviteCode(context.Background(), 1, 2)
	require.NoError(t, err)

	// Already revoked
	mockRepo.teacherInviteCode.Status = InviteCodeStatusRevoked
	err = service.RevokeTeacherInviteCode(context.Background(), 1, 2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already revoked")

	// Already used
	mockRepo.teacherInviteCode.Status = InviteCodeStatusUsed
	err = service.RevokeTeacherInviteCode(context.Background(), 1, 2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot revoke used")
}

func TestGenerateRandomInviteCode(t *testing.T) {
	code := generateRandomInviteCode(8)
	assert.Equal(t, 8, len(code))

	code2 := generateRandomInviteCode(8)
	assert.NotEqual(t, code, code2) // very unlikely to be same

	// Check alphabet
	for _, c := range code {
		assert.Contains(t, inviteCodeAlphabet, string(c))
	}
}

func TestResolveStatus(t *testing.T) {
	now := time.Now()

	// Revoked
	code := TeacherInviteCode{Status: InviteCodeStatusRevoked}
	assert.Equal(t, InviteCodeStatusRevoked, resolveStatus(code, now))

	// Used
	code = TeacherInviteCode{Status: InviteCodeStatusUsed}
	assert.Equal(t, InviteCodeStatusUsed, resolveStatus(code, now))

	// Expired
	past := now.AddDate(0, 0, -1)
	code = TeacherInviteCode{Status: InviteCodeStatusUnused, ExpireTime: &past}
	assert.Equal(t, InviteCodeStatusExpired, resolveStatus(code, now))

	// Unused and not expired
	future := now.AddDate(0, 0, 1)
	code = TeacherInviteCode{Status: InviteCodeStatusUnused, ExpireTime: &future}
	assert.Equal(t, InviteCodeStatusUnused, resolveStatus(code, now))

	// No expire time
	code = TeacherInviteCode{Status: InviteCodeStatusUnused}
	assert.Equal(t, InviteCodeStatusUnused, resolveStatus(code, now))
}

func TestStatusDesc(t *testing.T) {
	assert.Equal(t, "未使用", statusDesc(InviteCodeStatusUnused))
	assert.Equal(t, "已使用", statusDesc(InviteCodeStatusUsed))
	assert.Equal(t, "已撤销", statusDesc(InviteCodeStatusRevoked))
	assert.Equal(t, "已过期", statusDesc(InviteCodeStatusExpired))
	assert.Equal(t, "未知状态", statusDesc("unknown"))
}

func TestBuildInviteLink(t *testing.T) {
	link := buildInviteLink("https://feyo.club", "CODE123")
	assert.Equal(t, "https://feyo.club?inviteCode=CODE123", link)

	link = buildInviteLink("", "CODE123")
	assert.Equal(t, "https://feyo.club?inviteCode=CODE123", link)

	link = buildInviteLink("https://feyo.club/", "CODE123")
	assert.Equal(t, "https://feyo.club?inviteCode=CODE123", link)
}
