package withdrawal

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockWithdrawalRepo struct {
	withdrawal      *Withdrawal
	withdrawals     *WithdrawalPageResult
	adminWithdrawals *AdminWithdrawalPageResult
	stats           *WithdrawalStats
	incomeStats     *IncomeStats
	createErr       error
	getErr          error
	listErr         error
	updateErr       error
	cancelErr       error
}

func (m *mockWithdrawalRepo) CreateWithdrawal(ctx context.Context, withdrawal *Withdrawal) error {
	if m.createErr != nil {
		return m.createErr
	}
	withdrawal.ID = 1
	withdrawal.WithdrawalNo = "WD202401010000001"
	withdrawal.CreatedAt = time.Now()
	withdrawal.UpdatedAt = time.Now()
	return nil
}

func (m *mockWithdrawalRepo) GetWithdrawalByID(ctx context.Context, id int64) (*Withdrawal, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.withdrawal, nil
}

func (m *mockWithdrawalRepo) ListTeacherWithdrawals(ctx context.Context, teacherID int64, page, pageSize int) (*WithdrawalPageResult, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.withdrawals, nil
}

func (m *mockWithdrawalRepo) CancelWithdrawal(ctx context.Context, id int64) error {
	return m.cancelErr
}

func (m *mockWithdrawalRepo) ListAdminWithdrawals(ctx context.Context, query WithdrawalQuery) (*AdminWithdrawalPageResult, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.adminWithdrawals, nil
}

func (m *mockWithdrawalRepo) UpdateStatus(ctx context.Context, id int64, status, adminRemark string, processedBy int64) error {
	return m.updateErr
}

func (m *mockWithdrawalRepo) GetStats(ctx context.Context) (*WithdrawalStats, error) {
	return m.stats, nil
}

func (m *mockWithdrawalRepo) GetTeacherStats(ctx context.Context, teacherID int64) (*IncomeStats, error) {
	return m.incomeStats, nil
}

func TestService_CalculateWithdrawal(t *testing.T) {
	service := NewService(&mockWithdrawalRepo{})

	result, err := service.CalculateWithdrawal(context.Background(), CalculateRequest{Amount: 100})
	require.NoError(t, err)
	assert.Equal(t, 100.0, result.Amount)
	assert.Equal(t, 6.0, result.TaxAmount)
	assert.Equal(t, 94.0, result.ActualAmount)
	assert.Equal(t, 0.06, result.TaxRate)

	_, err = service.CalculateWithdrawal(context.Background(), CalculateRequest{Amount: 0})
	assert.Error(t, err)

	_, err = service.CalculateWithdrawal(context.Background(), CalculateRequest{Amount: -10})
	assert.Error(t, err)
}

func TestService_Apply(t *testing.T) {
	mockRepo := &mockWithdrawalRepo{
		incomeStats: &IncomeStats{TotalIncome: 1000, SettledIncome: 500, UnsettledIncome: 500, WithdrawnAmount: 0, PendingWithdrawal: 0},
	}
	service := NewService(mockRepo)

	withdrawal, err := service.Apply(context.Background(), 1, ApplyRequest{Amount: 100, BankAccount: "622202123456789", BankName: "ICBC", AccountName: "Test"})
	require.NoError(t, err)
	assert.Equal(t, int64(1), withdrawal.ID)
	assert.Equal(t, 100.0, withdrawal.Amount)
	assert.Equal(t, 6.0, withdrawal.TaxAmount)
	assert.Equal(t, 94.0, withdrawal.ActualAmount)
	assert.Equal(t, WithdrawalStatusPending, withdrawal.Status)

	// Invalid teacherID
	_, err = service.Apply(context.Background(), 0, ApplyRequest{Amount: 100, BankAccount: "622202123456789"})
	assert.Error(t, err)

	// Invalid amount
	_, err = service.Apply(context.Background(), 1, ApplyRequest{Amount: 0, BankAccount: "622202123456789"})
	assert.Error(t, err)

	// No payment method
	_, err = service.Apply(context.Background(), 1, ApplyRequest{Amount: 100})
	assert.Error(t, err)

	// Insufficient income
	_, err = service.Apply(context.Background(), 1, ApplyRequest{Amount: 600, BankAccount: "622202123456789"})
	assert.Error(t, err)
}

func TestService_Cancel(t *testing.T) {
	mockRepo := &mockWithdrawalRepo{
		withdrawal: &Withdrawal{ID: 1, TeacherID: 1, Status: WithdrawalStatusPending},
	}
	service := NewService(mockRepo)

	err := service.Cancel(context.Background(), 1, 1)
	require.NoError(t, err)

	// Wrong teacher
	mockRepo.withdrawal.TeacherID = 2
	err = service.Cancel(context.Background(), 1, 1)
	assert.Error(t, err)

	// Already approved
	mockRepo.withdrawal.TeacherID = 1
	mockRepo.withdrawal.Status = WithdrawalStatusApproved
	err = service.Cancel(context.Background(), 1, 1)
	assert.Error(t, err)
}

func TestService_ListMine(t *testing.T) {
	mockRepo := &mockWithdrawalRepo{
		withdrawals: &WithdrawalPageResult{
			Total: 1, PageNum: 1, PageSize: 10,
			Records: []WithdrawalVO{{ID: 1, WithdrawalNo: "WD001", Amount: 100, Status: WithdrawalStatusPending}},
		},
	}
	service := NewService(mockRepo)

	result, err := service.ListMine(context.Background(), 1, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.Total)

	_, err = service.ListMine(context.Background(), 0, 1, 10)
	assert.Error(t, err)
}

func TestService_GetMine(t *testing.T) {
	mockRepo := &mockWithdrawalRepo{
		withdrawal: &Withdrawal{ID: 1, TeacherID: 1, WithdrawalNo: "WD001", Amount: 100, Status: WithdrawalStatusPending},
	}
	service := NewService(mockRepo)

	result, err := service.GetMine(context.Background(), 1, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.ID)

	// Wrong teacher
	mockRepo.withdrawal.TeacherID = 2
	_, err = service.GetMine(context.Background(), 1, 1)
	assert.Error(t, err)
}

func TestService_GetIncomeStats(t *testing.T) {
	mockRepo := &mockWithdrawalRepo{
		incomeStats: &IncomeStats{TotalIncome: 1000, SettledIncome: 500, UnsettledIncome: 300, WithdrawnAmount: 500, PendingWithdrawal: 200},
	}
	service := NewService(mockRepo)

	stats, err := service.GetIncomeStats(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, 1000.0, stats.TotalIncome)
	assert.Equal(t, 500.0, stats.SettledIncome)
	assert.Equal(t, 300.0, stats.UnsettledIncome)

	_, err = service.GetIncomeStats(context.Background(), 0)
	assert.Error(t, err)
}

func TestService_ListAdmin(t *testing.T) {
	mockRepo := &mockWithdrawalRepo{
		adminWithdrawals: &AdminWithdrawalPageResult{
			Total: 1, PageNum: 1, PageSize: 10,
			Records: []AdminWithdrawalVO{{ID: 1, WithdrawalNo: "WD001", Amount: 100, Status: WithdrawalStatusPending}},
		},
	}
	service := NewService(mockRepo)

	result, err := service.ListAdmin(context.Background(), WithdrawalQuery{PageNum: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.Total)
}

func TestService_GetAdmin(t *testing.T) {
	mockRepo := &mockWithdrawalRepo{
		withdrawal: &Withdrawal{ID: 1, TeacherID: 1, WithdrawalNo: "WD001", Amount: 100, Status: WithdrawalStatusPending},
	}
	service := NewService(mockRepo)

	result, err := service.GetAdmin(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.ID)

	_, err = service.GetAdmin(context.Background(), 0)
	assert.Error(t, err)
}

func TestService_Approve(t *testing.T) {
	mockRepo := &mockWithdrawalRepo{
		withdrawal: &Withdrawal{ID: 1, TeacherID: 1, Status: WithdrawalStatusPending},
	}
	service := NewService(mockRepo)

	err := service.Approve(context.Background(), 1, 1, "approved")
	require.NoError(t, err)

	// Already approved
	mockRepo.withdrawal.Status = WithdrawalStatusApproved
	err = service.Approve(context.Background(), 1, 1, "approved")
	assert.Error(t, err)
}

func TestService_Reject(t *testing.T) {
	mockRepo := &mockWithdrawalRepo{
		withdrawal: &Withdrawal{ID: 1, TeacherID: 1, Status: WithdrawalStatusPending},
	}
	service := NewService(mockRepo)

	err := service.Reject(context.Background(), 1, 1, "insufficient info")
	require.NoError(t, err)

	// Already paid
	mockRepo.withdrawal.Status = WithdrawalStatusPaid
	err = service.Reject(context.Background(), 1, 1, "insufficient info")
	assert.Error(t, err)
}

func TestService_Pay(t *testing.T) {
	mockRepo := &mockWithdrawalRepo{
		withdrawal: &Withdrawal{ID: 1, TeacherID: 1, Status: WithdrawalStatusApproved},
	}
	service := NewService(mockRepo)

	err := service.Pay(context.Background(), 1, 1)
	require.NoError(t, err)

	// Not approved
	mockRepo.withdrawal.Status = WithdrawalStatusPending
	err = service.Pay(context.Background(), 1, 1)
	assert.Error(t, err)
}

func TestService_Stats(t *testing.T) {
	mockRepo := &mockWithdrawalRepo{
		stats: &WithdrawalStats{TotalWithdrawals: 10, PendingWithdrawals: 3, ApprovedWithdrawals: 2, PaidWithdrawals: 5, TotalAmount: 1000},
	}
	service := NewService(mockRepo)

	stats, err := service.Stats(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 10, stats.TotalWithdrawals)
	assert.Equal(t, 3, stats.PendingWithdrawals)
	assert.Equal(t, 5, stats.PaidWithdrawals)
	assert.Equal(t, 1000.0, stats.TotalAmount)
}

func TestStatusDesc(t *testing.T) {
	assert.Equal(t, "待处理", statusDesc(WithdrawalStatusPending))
	assert.Equal(t, "已批准", statusDesc(WithdrawalStatusApproved))
	assert.Equal(t, "已拒绝", statusDesc(WithdrawalStatusRejected))
	assert.Equal(t, "已打款", statusDesc(WithdrawalStatusPaid))
	assert.Equal(t, "已取消", statusDesc(WithdrawalStatusCancelled))
	assert.Equal(t, "未知状态", statusDesc("unknown"))
}

func TestToWithdrawalVO(t *testing.T) {
	w := Withdrawal{ID: 1, WithdrawalNo: "WD001", Amount: 100, TaxAmount: 6, ActualAmount: 94, Status: WithdrawalStatusPending, CreatedAt: time.Now()}
	vo := toWithdrawalVO(w)
	assert.Equal(t, int64(1), vo.ID)
	assert.Equal(t, "待处理", vo.StatusDesc)
}

func TestToAdminWithdrawalVO(t *testing.T) {
	w := Withdrawal{ID: 1, WithdrawalNo: "WD001", TeacherID: 1, Amount: 100, TaxAmount: 6, ActualAmount: 94, Status: WithdrawalStatusPending, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	vo := toAdminWithdrawalVO(w)
	assert.Equal(t, int64(1), vo.ID)
	assert.Equal(t, "待处理", vo.StatusDesc)
	assert.Equal(t, int64(1), vo.TeacherID)
}

func TestNormalizePage(t *testing.T) {
	assert.Equal(t, 1, normalizePage(0))
	assert.Equal(t, 1, normalizePage(-1))
	assert.Equal(t, 5, normalizePage(5))
}

func TestNormalizePageSize(t *testing.T) {
	assert.Equal(t, 10, normalizePageSize(0))
	assert.Equal(t, 10, normalizePageSize(-1))
	assert.Equal(t, 20, normalizePageSize(20))
	assert.Equal(t, 100, normalizePageSize(200))
}
