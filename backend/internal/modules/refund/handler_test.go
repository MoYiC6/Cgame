package refund

import (
	"context"
	"fmt"
	"testing"
	"time"

	"backend/internal/platform/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRefundRepo struct {
	refund      *Refund
	refunds     *RefundPageResult
	adminRefunds *AdminRefundPageResult
	stats       *RefundStats
	createErr   error
	getErr      error
	listErr     error
	updateErr   error
}

func (m *mockRefundRepo) CreateRefund(ctx context.Context, refund *Refund) error {
	if m.createErr != nil {
		return m.createErr
	}
	refund.ID = 1
	refund.RefundNo = "RF202401010000001"
	refund.CreatedAt = time.Now()
	refund.UpdatedAt = time.Now()
	return nil
}

func (m *mockRefundRepo) ListUserRefunds(ctx context.Context, userID int64, page, pageSize int) (*RefundPageResult, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.refunds, nil
}

func (m *mockRefundRepo) GetRefundByID(ctx context.Context, id int64) (*Refund, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if m.refund != nil && m.refund.ID == id {
		return m.refund, nil
	}
	return m.refund, nil
}

func (m *mockRefundRepo) GetRefundByOrderID(ctx context.Context, orderID int64) (*Refund, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.refund, nil
}

func (m *mockRefundRepo) CancelRefund(ctx context.Context, id int64) error {
	return m.updateErr
}

func (m *mockRefundRepo) ListAdminRefunds(ctx context.Context, query RefundQuery) (*AdminRefundPageResult, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.adminRefunds, nil
}

func (m *mockRefundRepo) UpdateStatus(ctx context.Context, id int64, status, adminRemark string, processedBy int64) error {
	return m.updateErr
}

func (m *mockRefundRepo) GetStats(ctx context.Context) (*RefundStats, error) {
	return m.stats, nil
}

type mockOrderChecker struct {
	order  *OrderInfo
	getErr error
}

func (m *mockOrderChecker) GetOrderByID(ctx context.Context, id int64) (*OrderInfo, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.order, nil
}

type mockRefundTxManager struct{}

func (m mockRefundTxManager) WithinTx(ctx context.Context, fn func(context.Context) error, opts ...database.TxOption) error {
	return fn(ctx)
}

func TestService_Apply(t *testing.T) {
	mockRepo := &mockRefundRepo{}
	service := NewService(mockRepo, mockRefundTxManager{})
	checker := &mockOrderChecker{
		order: &OrderInfo{
			ID: 100, UserID: 1, Status: "completed", PayAmount: 100.0,
		},
	}
	service.SetOrderChecker(checker)

	id, err := service.Apply(context.Background(), 1, ApplyRequest{OrderID: 100, Amount: 50, Reason: "test"})
	require.NoError(t, err)
	assert.Equal(t, int64(1), id)

	// Invalid userID
	_, err = service.Apply(context.Background(), 0, ApplyRequest{OrderID: 100, Amount: 50, Reason: "test"})
	assert.Error(t, err)

	// Invalid orderID
	_, err = service.Apply(context.Background(), 1, ApplyRequest{OrderID: 0, Amount: 50, Reason: "test"})
	assert.Error(t, err)

	// Invalid amount
	_, err = service.Apply(context.Background(), 1, ApplyRequest{OrderID: 100, Amount: 0, Reason: "test"})
	assert.Error(t, err)

	// Empty reason
	_, err = service.Apply(context.Background(), 1, ApplyRequest{OrderID: 100, Amount: 50, Reason: ""})
	assert.Error(t, err)

	// Amount exceeds order pay amount
	_, err = service.Apply(context.Background(), 1, ApplyRequest{OrderID: 100, Amount: 150, Reason: "test"})
	assert.Error(t, err)
}

func TestService_ListMine(t *testing.T) {
	mockRepo := &mockRefundRepo{
		refunds: &RefundPageResult{
			Total: 1, PageNum: 1, PageSize: 10,
			Records: []RefundVO{{ID: 1, RefundNo: "RF001", Amount: 50}},
		},
	}
	service := NewService(mockRepo, mockRefundTxManager{})

	result, err := service.ListMine(context.Background(), 1, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.Total)

	_, err = service.ListMine(context.Background(), 0, 1, 10)
	assert.Error(t, err)
}

func TestService_GetMine(t *testing.T) {
	mockRepo := &mockRefundRepo{
		refund: &Refund{ID: 1, UserID: 1, RefundNo: "RF001", Amount: 50, Status: RefundStatusPending},
	}
	service := NewService(mockRepo, mockRefundTxManager{})

	result, err := service.GetMine(context.Background(), 1, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.ID)

	// Wrong user
	mockRepo.refund = &Refund{ID: 1, UserID: 2, RefundNo: "RF001", Amount: 50}
	_, err = service.GetMine(context.Background(), 1, 1)
	assert.Error(t, err)
}

func TestService_Cancel(t *testing.T) {
	mockRepo := &mockRefundRepo{
		refund: &Refund{ID: 1, UserID: 1, Status: RefundStatusPending},
	}
	service := NewService(mockRepo, mockRefundTxManager{})

	err := service.Cancel(context.Background(), 1, 1)
	require.NoError(t, err)

	// Already processed
	mockRepo.refund.Status = RefundStatusProcessed
	err = service.Cancel(context.Background(), 1, 1)
	assert.Error(t, err)
}

func TestService_CanApply(t *testing.T) {
	mockRepo := &mockRefundRepo{}
	service := NewService(mockRepo, mockRefundTxManager{})
	checker := &mockOrderChecker{
		order: &OrderInfo{
			ID: 100, UserID: 1, Status: "completed", PayAmount: 100.0,
		},
	}
	service.SetOrderChecker(checker)

	result, err := service.CanApply(context.Background(), 1, 100)
	require.NoError(t, err)
	assert.True(t, result.CanApply)

	// Order not found
	checker.getErr = fmt.Errorf("not found")
	result, err = service.CanApply(context.Background(), 1, 100)
	require.NoError(t, err)
	assert.False(t, result.CanApply)
}

func TestService_GetByOrder(t *testing.T) {
	mockRepo := &mockRefundRepo{
		refund: &Refund{ID: 1, UserID: 1, OrderID: 100, RefundNo: "RF001"},
	}
	service := NewService(mockRepo, mockRefundTxManager{})

	result, err := service.GetByOrder(context.Background(), 1, 100)
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.ID)

	// Wrong user
	mockRepo.refund.UserID = 2
	_, err = service.GetByOrder(context.Background(), 1, 100)
	assert.Error(t, err)
}

func TestService_AdminOperations(t *testing.T) {
	mockRepo := &mockRefundRepo{
		refund: &Refund{ID: 1, UserID: 1, Status: RefundStatusPending, RefundNo: "RF001"},
		adminRefunds: &AdminRefundPageResult{
			Total: 1, PageNum: 1, PageSize: 10,
			Records: []AdminRefundVO{{ID: 1, RefundNo: "RF001"}},
		},
		stats: &RefundStats{TotalRefunds: 5, PendingRefunds: 2},
	}
	service := NewService(mockRepo, mockRefundTxManager{})

	// ListAdmin
	result, err := service.ListAdmin(context.Background(), RefundQuery{PageNum: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.Total)

	// GetAdmin
	adminResult, err := service.GetAdmin(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), adminResult.ID)

	// Approve
	err = service.Approve(context.Background(), 2, 1, "approved")
	require.NoError(t, err)

	// Reject
	mockRepo.refund.Status = RefundStatusPending
	err = service.Reject(context.Background(), 2, 1, "rejected")
	require.NoError(t, err)

	// Process
	mockRepo.refund.Status = RefundStatusApproved
	err = service.Process(context.Background(), 2, 1, "processed")
	require.NoError(t, err)

	// Stats
	stats, err := service.Stats(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 5, stats.TotalRefunds)

	// Invalid status transitions
	mockRepo.refund.Status = RefundStatusProcessed
	err = service.Approve(context.Background(), 2, 1, "")
	assert.Error(t, err)

	mockRepo.refund.Status = RefundStatusRejected
	err = service.Reject(context.Background(), 2, 1, "")
	assert.Error(t, err)

	mockRepo.refund.Status = RefundStatusPending
	err = service.Process(context.Background(), 2, 1, "")
	assert.Error(t, err)
}

func TestStatusDesc(t *testing.T) {
	assert.Equal(t, "待处理", statusDesc(RefundStatusPending))
	assert.Equal(t, "已批准", statusDesc(RefundStatusApproved))
	assert.Equal(t, "已拒绝", statusDesc(RefundStatusRejected))
	assert.Equal(t, "已处理", statusDesc(RefundStatusProcessed))
	assert.Equal(t, "已取消", statusDesc(RefundStatusCancelled))
	assert.Equal(t, "未知状态", statusDesc("unknown"))
}

func TestToRefundVO(t *testing.T) {
	r := Refund{ID: 1, RefundNo: "RF001", OrderID: 100, Amount: 50, Status: RefundStatusPending, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	vo := toRefundVO(r)
	assert.Equal(t, int64(1), vo.ID)
	assert.Equal(t, "待处理", vo.StatusDesc)
}

func TestToAdminRefundVO(t *testing.T) {
	now := time.Now()
	adminRemark := "test"
	processedBy := int64(2)
	r := Refund{ID: 1, RefundNo: "RF001", OrderID: 100, UserID: 1, Amount: 50, Status: RefundStatusApproved, AdminRemark: adminRemark, ProcessedBy: &processedBy, ProcessedAt: &now, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	vo := toAdminRefundVO(r)
	assert.Equal(t, int64(1), vo.ID)
	assert.Equal(t, "已批准", vo.StatusDesc)
	assert.Equal(t, adminRemark, vo.AdminRemark)
	assert.Equal(t, &processedBy, vo.ProcessedBy)
}
