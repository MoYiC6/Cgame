package partner

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockPartnerRepo struct {
	config        *PartnerConfig
	configs       *PartnerConfigPageResult
	teacherPartner *TeacherPartner
	teacherPartners *TeacherPartnerPageResult
	createErr     error
	getErr        error
	listErr       error
	updateErr     error
	deleteErr     error
}

func (m *mockPartnerRepo) CreatePartnerConfig(ctx context.Context, config *PartnerConfig) error {
	if m.createErr != nil {
		return m.createErr
	}
	config.ID = 1
	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()
	return nil
}

func (m *mockPartnerRepo) GetPartnerConfigByID(ctx context.Context, id int64) (*PartnerConfig, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.config, nil
}

func (m *mockPartnerRepo) ListPartnerConfigs(ctx context.Context, query PartnerConfigQuery) (*PartnerConfigPageResult, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.configs, nil
}

func (m *mockPartnerRepo) UpdatePartnerConfig(ctx context.Context, id int64, updates map[string]any) error {
	return m.updateErr
}

func (m *mockPartnerRepo) DeletePartnerConfig(ctx context.Context, id int64) error {
	return m.deleteErr
}

func (m *mockPartnerRepo) CreateTeacherPartner(ctx context.Context, tp *TeacherPartner) error {
	if m.createErr != nil {
		return m.createErr
	}
	tp.ID = 1
	tp.CreatedAt = time.Now()
	tp.UpdatedAt = time.Now()
	return nil
}

func (m *mockPartnerRepo) GetTeacherPartnerByID(ctx context.Context, id int64) (*TeacherPartner, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.teacherPartner, nil
}

func (m *mockPartnerRepo) ListTeacherPartners(ctx context.Context, query TeacherPartnerQuery) (*TeacherPartnerPageResult, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.teacherPartners, nil
}

func (m *mockPartnerRepo) UpdateTeacherPartner(ctx context.Context, id int64, updates map[string]any) error {
	return m.updateErr
}

func (m *mockPartnerRepo) DeleteTeacherPartner(ctx context.Context, id int64) error {
	return m.deleteErr
}

func (m *mockPartnerRepo) ListPartneredTeachers(ctx context.Context, page, pageSize int) (*TeacherPartnerPageResult, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.teacherPartners, nil
}

func TestService_CreatePartnerConfig(t *testing.T) {
	mockRepo := &mockPartnerRepo{}
	service := NewService(mockRepo)

	id, err := service.CreatePartnerConfig(context.Background(), PartnerConfigCreateRequest{Name: "Test Partner", PartnerType: PartnerTypeAgency, CommissionRate: 10, FixedFee: 100})
	require.NoError(t, err)
	assert.Equal(t, int64(1), id)

	// Empty name
	_, err = service.CreatePartnerConfig(context.Background(), PartnerConfigCreateRequest{Name: "", PartnerType: PartnerTypeAgency})
	assert.Error(t, err)

	// Negative commission rate
	_, err = service.CreatePartnerConfig(context.Background(), PartnerConfigCreateRequest{Name: "Test", CommissionRate: -1})
	assert.Error(t, err)

	// Negative fixed fee
	_, err = service.CreatePartnerConfig(context.Background(), PartnerConfigCreateRequest{Name: "Test", FixedFee: -1})
	assert.Error(t, err)
}

func TestService_ListPartnerConfigs(t *testing.T) {
	mockRepo := &mockPartnerRepo{
		configs: &PartnerConfigPageResult{
			Total: 1, PageNum: 1, PageSize: 10,
			Records: []PartnerConfigVO{{ID: 1, Name: "Test Partner", Status: PartnerConfigStatusEnabled}},
		},
	}
	service := NewService(mockRepo)

	result, err := service.ListPartnerConfigs(context.Background(), PartnerConfigQuery{PageNum: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.Total)
}

func TestService_GetPartnerConfig(t *testing.T) {
	mockRepo := &mockPartnerRepo{
		config: &PartnerConfig{ID: 1, Name: "Test Partner", Status: PartnerConfigStatusEnabled},
	}
	service := NewService(mockRepo)

	result, err := service.GetPartnerConfig(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.ID)

	_, err = service.GetPartnerConfig(context.Background(), 0)
	assert.Error(t, err)
}

func TestService_UpdatePartnerConfig(t *testing.T) {
	mockRepo := &mockPartnerRepo{}
	service := NewService(mockRepo)

	name := "Updated"
	err := service.UpdatePartnerConfig(context.Background(), 1, PartnerConfigUpdateRequest{Name: &name})
	require.NoError(t, err)

	// Empty name
	emptyName := ""
	err = service.UpdatePartnerConfig(context.Background(), 1, PartnerConfigUpdateRequest{Name: &emptyName})
	assert.Error(t, err)

	// Negative commission rate
	negativeRate := -1.0
	err = service.UpdatePartnerConfig(context.Background(), 1, PartnerConfigUpdateRequest{CommissionRate: &negativeRate})
	assert.Error(t, err)

	// Invalid id
	err = service.UpdatePartnerConfig(context.Background(), 0, PartnerConfigUpdateRequest{})
	assert.Error(t, err)
}

func TestService_DeletePartnerConfig(t *testing.T) {
	mockRepo := &mockPartnerRepo{}
	service := NewService(mockRepo)

	err := service.DeletePartnerConfig(context.Background(), 1)
	require.NoError(t, err)

	err = service.DeletePartnerConfig(context.Background(), 0)
	assert.Error(t, err)
}

func TestService_CreateTeacherPartner(t *testing.T) {
	mockRepo := &mockPartnerRepo{}
	service := NewService(mockRepo)

	id, err := service.CreateTeacherPartner(context.Background(), TeacherPartnerCreateRequest{TeacherID: 1, PartnerID: 1, CommissionRate: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(1), id)

	// Missing teacherID
	_, err = service.CreateTeacherPartner(context.Background(), TeacherPartnerCreateRequest{PartnerID: 1})
	assert.Error(t, err)

	// Missing partnerID
	_, err = service.CreateTeacherPartner(context.Background(), TeacherPartnerCreateRequest{TeacherID: 1})
	assert.Error(t, err)

	// Negative commission rate
	_, err = service.CreateTeacherPartner(context.Background(), TeacherPartnerCreateRequest{TeacherID: 1, PartnerID: 1, CommissionRate: -1})
	assert.Error(t, err)
}

func TestService_ListTeacherPartners(t *testing.T) {
	mockRepo := &mockPartnerRepo{
		teacherPartners: &TeacherPartnerPageResult{
			Total: 1, PageNum: 1, PageSize: 10,
			Records: []TeacherPartnerVO{{ID: 1, TeacherID: 1, PartnerID: 1, Status: CooperationStatusActive}},
		},
	}
	service := NewService(mockRepo)

	result, err := service.ListTeacherPartners(context.Background(), TeacherPartnerQuery{PageNum: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.Total)
}

func TestService_GetTeacherPartner(t *testing.T) {
	mockRepo := &mockPartnerRepo{
		teacherPartner: &TeacherPartner{ID: 1, TeacherID: 1, PartnerID: 1, Status: CooperationStatusActive},
	}
	service := NewService(mockRepo)

	result, err := service.GetTeacherPartner(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.ID)

	_, err = service.GetTeacherPartner(context.Background(), 0)
	assert.Error(t, err)
}

func TestService_UpdateTeacherPartner(t *testing.T) {
	mockRepo := &mockPartnerRepo{}
	service := NewService(mockRepo)

	status := CooperationStatusInactive
	err := service.UpdateTeacherPartner(context.Background(), 1, TeacherPartnerUpdateRequest{Status: &status})
	require.NoError(t, err)

	// Negative commission rate
	negativeRate := -1.0
	err = service.UpdateTeacherPartner(context.Background(), 1, TeacherPartnerUpdateRequest{CommissionRate: &negativeRate})
	assert.Error(t, err)

	// Invalid id
	err = service.UpdateTeacherPartner(context.Background(), 0, TeacherPartnerUpdateRequest{})
	assert.Error(t, err)
}

func TestService_DeleteTeacherPartner(t *testing.T) {
	mockRepo := &mockPartnerRepo{}
	service := NewService(mockRepo)

	err := service.DeleteTeacherPartner(context.Background(), 1)
	require.NoError(t, err)

	err = service.DeleteTeacherPartner(context.Background(), 0)
	assert.Error(t, err)
}

func TestService_ListPartneredTeachers(t *testing.T) {
	mockRepo := &mockPartnerRepo{
		teacherPartners: &TeacherPartnerPageResult{
			Total: 2, PageNum: 1, PageSize: 10,
			Records: []TeacherPartnerVO{
				{ID: 1, TeacherID: 1, PartnerID: 1, Status: CooperationStatusActive},
				{ID: 2, TeacherID: 2, PartnerID: 1, Status: CooperationStatusActive},
			},
		},
	}
	service := NewService(mockRepo)

	result, err := service.ListPartneredTeachers(context.Background(), 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), result.Total)
}

func TestStatusDesc(t *testing.T) {
	assert.Equal(t, "启用", statusDesc(PartnerConfigStatusEnabled))
	assert.Equal(t, "禁用", statusDesc(PartnerConfigStatusDisabled))
	assert.Equal(t, "生效中", statusDesc(CooperationStatusActive))
	assert.Equal(t, "已停用", statusDesc(CooperationStatusInactive))
	assert.Equal(t, "已过期", statusDesc(CooperationStatusExpired))
	assert.Equal(t, "未知状态", statusDesc("unknown"))
}

func TestPartnerTypeDesc(t *testing.T) {
	assert.Equal(t, "代理机构", partnerTypeDesc(PartnerTypeAgency))
	assert.Equal(t, "个人", partnerTypeDesc(PartnerTypePersonal))
	assert.Equal(t, "其他", partnerTypeDesc(PartnerTypeOther))
	assert.Equal(t, "未知类型", partnerTypeDesc("unknown"))
}

func TestCooperationTypeDesc(t *testing.T) {
	assert.Equal(t, "独家", cooperationTypeDesc(CooperationTypeExclusive))
	assert.Equal(t, "非独家", cooperationTypeDesc(CooperationTypeNonExclusive))
	assert.Equal(t, "未知类型", cooperationTypeDesc("unknown"))
}

func TestToPartnerConfigVO(t *testing.T) {
	p := PartnerConfig{ID: 1, Name: "Test", PartnerType: PartnerTypeAgency, Status: PartnerConfigStatusEnabled, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	vo := toPartnerConfigVO(p)
	assert.Equal(t, int64(1), vo.ID)
	assert.Equal(t, "启用", vo.StatusDesc)
	assert.Equal(t, "代理机构", vo.PartnerTypeDesc)
}

func TestToTeacherPartnerVO(t *testing.T) {
	tp := TeacherPartner{ID: 1, TeacherID: 1, PartnerID: 1, CooperationType: CooperationTypeExclusive, Status: CooperationStatusActive, CreatedAt: time.Now()}
	vo := toTeacherPartnerVO(tp, "Partner A")
	assert.Equal(t, int64(1), vo.ID)
	assert.Equal(t, "独家", vo.CooperationTypeDesc)
	assert.Equal(t, "生效中", vo.StatusDesc)
	assert.Equal(t, "Partner A", vo.PartnerName)
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
