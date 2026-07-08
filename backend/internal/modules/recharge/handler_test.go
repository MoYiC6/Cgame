package recharge

import (
	"context"
	"testing"
	"time"

	"backend/internal/platform/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRechargeRepo struct {
	record      *RechargeRecord
	records     *RechargeRecordPageResult
	stats       *RechargeStats
	recent      []RechargeRecordVO
	rule        *RechargeRebateRule
	rules       []RechargeRebateRule
	rulePage    *RebateRulePageResult
	createErr   error
	getErr      error
	listErr     error
	updateErr   error
	deleteErr   error
}

func (m *mockRechargeRepo) CreateRechargeRecord(ctx context.Context, record *RechargeRecord) error {
	if m.createErr != nil {
		return m.createErr
	}
	record.ID = 1
	record.RechargeNo = "RC202401010000001"
	record.CreatedAt = time.Now()
	record.UpdatedAt = time.Now()
	return nil
}

func (m *mockRechargeRepo) GetRechargeByID(ctx context.Context, id int64) (*RechargeRecord, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.record, nil
}

func (m *mockRechargeRepo) GetRechargeByNo(ctx context.Context, rechargeNo string) (*RechargeRecord, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.record, nil
}

func (m *mockRechargeRepo) ListUserRecharges(ctx context.Context, userID int64, page, pageSize int) (*RechargeRecordPageResult, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.records, nil
}

func (m *mockRechargeRepo) ListRecharges(ctx context.Context, query RechargeQuery) (*RechargeRecordPageResult, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.records, nil
}

func (m *mockRechargeRepo) UpdateStatus(ctx context.Context, rechargeNo, status string, payChannel string, payTime, callbackTime *time.Time) error {
	return m.updateErr
}

func (m *mockRechargeRepo) CancelRecharge(ctx context.Context, rechargeNo string) error {
	return m.updateErr
}

func (m *mockRechargeRepo) GetStats(ctx context.Context) (*RechargeStats, error) {
	return m.stats, nil
}

func (m *mockRechargeRepo) GetRecentRecharges(ctx context.Context, userID int64, limit int) ([]RechargeRecordVO, error) {
	return m.recent, nil
}

func (m *mockRechargeRepo) CreateRebateRule(ctx context.Context, rule *RechargeRebateRule) error {
	if m.createErr != nil {
		return m.createErr
	}
	rule.ID = 1
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()
	return nil
}

func (m *mockRechargeRepo) GetRebateRuleByID(ctx context.Context, id int64) (*RechargeRebateRule, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.rule, nil
}

func (m *mockRechargeRepo) ListRebateRules(ctx context.Context, page, pageSize int) (*RebateRulePageResult, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.rulePage, nil
}

func (m *mockRechargeRepo) UpdateRebateRule(ctx context.Context, id int64, updates map[string]any) error {
	return m.updateErr
}

func (m *mockRechargeRepo) DeleteRebateRule(ctx context.Context, id int64) error {
	return m.deleteErr
}

func (m *mockRechargeRepo) ListEnabledRebateRules(ctx context.Context) ([]RechargeRebateRule, error) {
	return m.rules, nil
}

type mockRechargeTxManager struct{}

func (m mockRechargeTxManager) WithinTx(ctx context.Context, fn func(context.Context) error, opts ...database.TxOption) error {
	return fn(ctx)
}

func TestService_CreateRecharge(t *testing.T) {
	mockRepo := &mockRechargeRepo{
		rules: []RechargeRebateRule{
			{ID: 1, Name: "满100送10", MinAmount: 100, GiftAmount: 10, GiftRate: 0, Enabled: true, Priority: 1},
		},
	}
	service := NewService(mockRepo, mockRechargeTxManager{})

	record, err := service.CreateRecharge(context.Background(), 1, CreateRechargeRequest{Amount: 200})
	require.NoError(t, err)
	assert.Equal(t, int64(1), record.ID)
	assert.Equal(t, 200.0, record.Amount)
	assert.Equal(t, 10.0, record.GiftAmount)
	assert.Equal(t, 210.0, record.TotalAmount)

	// Invalid userID
	_, err = service.CreateRecharge(context.Background(), 0, CreateRechargeRequest{Amount: 100})
	assert.Error(t, err)

	// Invalid amount
	_, err = service.CreateRecharge(context.Background(), 1, CreateRechargeRequest{Amount: 0})
	assert.Error(t, err)
}

func TestService_ManualRecharge(t *testing.T) {
	mockRepo := &mockRechargeRepo{}
	service := NewService(mockRepo, mockRechargeTxManager{})

	id, err := service.ManualRecharge(context.Background(), 1, ManualRechargeRequest{UserID: 2, Amount: 100, Remark: "test"})
	require.NoError(t, err)
	assert.Equal(t, int64(1), id)

	// Invalid admin userID
	_, err = service.ManualRecharge(context.Background(), 0, ManualRechargeRequest{UserID: 2, Amount: 100})
	assert.Error(t, err)

	// Invalid target userID
	_, err = service.ManualRecharge(context.Background(), 1, ManualRechargeRequest{UserID: 0, Amount: 100})
	assert.Error(t, err)

	// Invalid amount
	_, err = service.ManualRecharge(context.Background(), 1, ManualRechargeRequest{UserID: 2, Amount: 0})
	assert.Error(t, err)
}

func TestService_ListMine(t *testing.T) {
	mockRepo := &mockRechargeRepo{
		records: &RechargeRecordPageResult{
			Total: 1, PageNum: 1, PageSize: 10,
			Records: []RechargeRecordVO{{ID: 1, RechargeNo: "RC001", Amount: 100}},
		},
	}
	service := NewService(mockRepo, mockRechargeTxManager{})

	result, err := service.ListMine(context.Background(), 1, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.Total)

	_, err = service.ListMine(context.Background(), 0, 1, 10)
	assert.Error(t, err)
}

func TestService_GetMine(t *testing.T) {
	mockRepo := &mockRechargeRepo{
		record: &RechargeRecord{ID: 1, UserID: 1, RechargeNo: "RC001", Amount: 100, Status: RechargeStatusPending},
	}
	service := NewService(mockRepo, mockRechargeTxManager{})

	result, err := service.GetMine(context.Background(), 1, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.ID)

	// Wrong user
	mockRepo.record.UserID = 2
	_, err = service.GetMine(context.Background(), 1, 1)
	assert.Error(t, err)
}

func TestService_Cancel(t *testing.T) {
	mockRepo := &mockRechargeRepo{
		record: &RechargeRecord{ID: 1, UserID: 1, RechargeNo: "RC001", Status: RechargeStatusPending},
	}
	service := NewService(mockRepo, mockRechargeTxManager{})

	err := service.Cancel(context.Background(), 1, "RC001")
	require.NoError(t, err)

	// Already paid
	mockRepo.record.Status = RechargeStatusPaid
	err = service.Cancel(context.Background(), 1, "RC001")
	assert.Error(t, err)
}

func TestService_Callback(t *testing.T) {
	mockRepo := &mockRechargeRepo{
		record: &RechargeRecord{ID: 1, RechargeNo: "RC001", Status: RechargeStatusPending},
	}
	service := NewService(mockRepo, mockRechargeTxManager{})

	err := service.Callback(context.Background(), "RC001", "wxpay", 100)
	require.NoError(t, err)

	// Already paid
	mockRepo.record.Status = RechargeStatusPaid
	err = service.Callback(context.Background(), "RC001", "wxpay", 100)
	assert.Error(t, err)
}

func TestService_PreviewRebate(t *testing.T) {
	mockRepo := &mockRechargeRepo{
		rules: []RechargeRebateRule{
			{ID: 1, Name: "满100送10", MinAmount: 100, GiftAmount: 10, GiftRate: 0, Enabled: true, Priority: 1},
			{ID: 2, Name: "满50送5", MinAmount: 50, GiftAmount: 5, GiftRate: 0, Enabled: true, Priority: 2},
		},
	}
	service := NewService(mockRepo, mockRechargeTxManager{})

	result, err := service.PreviewRebate(context.Background(), 200)
	require.NoError(t, err)
	assert.Equal(t, 200.0, result.Amount)
	assert.Equal(t, 10.0, result.GiftAmount)
	assert.Equal(t, 210.0, result.TotalAmount)
	assert.Equal(t, "满100送10", result.RuleName)

	// Invalid amount
	_, err = service.PreviewRebate(context.Background(), 0)
	assert.Error(t, err)
}

func TestService_RebateRules(t *testing.T) {
	mockRepo := &mockRechargeRepo{
		rulePage: &RebateRulePageResult{
			Total: 1, PageNum: 1, PageSize: 10,
			Records: []RechargeRebateRuleVO{{ID: 1, Name: "满100送10"}},
		},
	}
	service := NewService(mockRepo, mockRechargeTxManager{})

	// List
	result, err := service.ListRebateRules(context.Background(), 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.Total)

	// Create
	id, err := service.CreateRebateRule(context.Background(), RebateRuleCreateRequest{Name: "满200送20", MinAmount: 200, GiftAmount: 20})
	require.NoError(t, err)
	assert.Equal(t, int64(1), id)

	// Create empty name
	_, err = service.CreateRebateRule(context.Background(), RebateRuleCreateRequest{Name: "", MinAmount: 200})
	assert.Error(t, err)

	// Create negative min amount
	_, err = service.CreateRebateRule(context.Background(), RebateRuleCreateRequest{Name: "test", MinAmount: -1})
	assert.Error(t, err)

	// Update
	err = service.UpdateRebateRule(context.Background(), 1, RebateRuleUpdateRequest{Name: strPtr("updated")})
	require.NoError(t, err)

	// Update empty name
	err = service.UpdateRebateRule(context.Background(), 1, RebateRuleUpdateRequest{Name: strPtr("")})
	assert.Error(t, err)

	// Delete
	err = service.DeleteRebateRule(context.Background(), 1)
	require.NoError(t, err)

	// Delete invalid id
	err = service.DeleteRebateRule(context.Background(), 0)
	assert.Error(t, err)
}

func TestService_Stats(t *testing.T) {
	mockRepo := &mockRechargeRepo{
		stats: &RechargeStats{TotalRecords: 10, PaidRecords: 5, TotalAmount: 1000},
	}
	service := NewService(mockRepo, mockRechargeTxManager{})

	stats, err := service.Stats(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 10, stats.TotalRecords)
	assert.Equal(t, 5, stats.PaidRecords)
	assert.Equal(t, 1000.0, stats.TotalAmount)
}

func TestStatusDesc(t *testing.T) {
	assert.Equal(t, "待支付", statusDesc(RechargeStatusPending))
	assert.Equal(t, "已支付", statusDesc(RechargeStatusPaid))
	assert.Equal(t, "已取消", statusDesc(RechargeStatusCancelled))
	assert.Equal(t, "失败", statusDesc(RechargeStatusFailed))
	assert.Equal(t, "未知状态", statusDesc("unknown"))
}

func TestToRechargeRecordVO(t *testing.T) {
	r := RechargeRecord{ID: 1, RechargeNo: "RC001", Amount: 100, GiftAmount: 10, TotalAmount: 110, Status: RechargeStatusPending, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	vo := toRechargeRecordVO(r)
	assert.Equal(t, int64(1), vo.ID)
	assert.Equal(t, "待支付", vo.StatusDesc)
}

func TestToRebateRuleVO(t *testing.T) {
	r := RechargeRebateRule{ID: 1, Name: "满100送10", MinAmount: 100, GiftAmount: 10, Enabled: true, Priority: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	vo := toRebateRuleVO(r)
	assert.Equal(t, int64(1), vo.ID)
	assert.Equal(t, "满100送10", vo.Name)
}

func strPtr(s string) *string {
	return &s
}
