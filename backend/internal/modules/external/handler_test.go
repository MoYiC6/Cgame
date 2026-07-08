package external

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/platform/response"

	"github.com/gin-gonic/gin"
)

func TestWxPayConfigDetailUsesPathID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &stubExternalRepository{
		configByID: &WxPayConfig{ID: 7, ConfigType: "miniapp", AppID: "wx-app", MchID: "mch-1", Status: 1},
	}
	engine := newExternalHandlerEngine(NewHandler(NewService(repo), nil))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/admin/wxpay/config/7", nil)
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if repo.lastConfigID != 7 {
		t.Fatalf("expected config id 7, got %d", repo.lastConfigID)
	}

	var body response.APIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	data, ok := body.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected response data map, got %#v", body.Data)
	}
	if data["ID"] != float64(7) {
		t.Fatalf("expected ID 7 in response, got %#v", data["ID"])
	}
}

func TestWxPayConfigStatusUsesJavaEnabledQueryParam(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &stubExternalRepository{}
	engine := newExternalHandlerEngine(NewHandler(NewService(repo), nil))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, "/api/admin/wxpay/config/7/status?enabled=1", nil)
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if repo.lastStatusUpdateID != 7 || repo.lastStatusUpdate != 1 {
		t.Fatalf("expected status update id/status 7/1, got %d/%d", repo.lastStatusUpdateID, repo.lastStatusUpdate)
	}
	if repo.usedFullConfigUpdate {
		t.Fatal("expected status endpoint to use a dedicated status update instead of the full config update")
	}
}

func TestWxPayConfigListUsesJavaPagingParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &stubExternalRepository{
		configs: []*WxPayConfig{{ID: 5, ConfigType: "miniapp", AppID: "wx-app", MchID: "mch-1", Status: 1}},
		total:   8,
	}
	engine := newExternalHandlerEngine(NewHandler(NewService(repo), nil))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/admin/wxpay/config/page?pageIndex=2&pageSize=3&configType=miniapp", nil)
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if repo.lastListPage != 2 || repo.lastListPageSize != 3 {
		t.Fatalf("expected list paging 2/3, got %d/%d", repo.lastListPage, repo.lastListPageSize)
	}
	if repo.lastListConfigType == nil || *repo.lastListConfigType != "miniapp" {
		t.Fatalf("expected configType miniapp, got %#v", repo.lastListConfigType)
	}

	var body response.APIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	data, ok := body.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected response data map, got %#v", body.Data)
	}
	if data["total"] != float64(8) {
		t.Fatalf("expected total 8, got %#v", data["total"])
	}
	rows, ok := data["rows"].([]any)
	if !ok || len(rows) != 1 {
		t.Fatalf("expected one row, got %#v", data["rows"])
	}
}

func newExternalHandlerEngine(h *Handler) *gin.Engine {
	engine := gin.New()
	api := engine.Group("/api")
	h.RegisterRoutes(api)
	return engine
}

type stubExternalRepository struct {
	configByID   *WxPayConfig
	lastConfigID int64

	lastStatusUpdateID   int64
	lastStatusUpdate     int
	usedFullConfigUpdate bool

	configs            []*WxPayConfig
	total              int
	lastListPage       int
	lastListPageSize   int
	lastListConfigType *string
}

func (s *stubExternalRepository) GetUserOAuth(ctx context.Context, platform, openID string) (*UserOAuth, error) {
	return nil, nil
}

func (s *stubExternalRepository) GetUserOAuthByUserID(ctx context.Context, userID int64, platform string) (*UserOAuth, error) {
	return nil, nil
}

func (s *stubExternalRepository) CreateUserOAuth(ctx context.Context, oauth *UserOAuth) error {
	return nil
}

func (s *stubExternalRepository) UpdateUserOAuth(ctx context.Context, oauth *UserOAuth) error {
	return nil
}

func (s *stubExternalRepository) DeleteUserOAuth(ctx context.Context, userID int64, platform string) error {
	return nil
}

func (s *stubExternalRepository) CreateUserToken(ctx context.Context, token *UserToken) error {
	return nil
}

func (s *stubExternalRepository) GetUserToken(ctx context.Context, accessToken string) (*UserToken, error) {
	return nil, nil
}

func (s *stubExternalRepository) CreateScanLoginSession(ctx context.Context, session *ScanLoginSession) error {
	return nil
}

func (s *stubExternalRepository) GetScanLoginSession(ctx context.Context, loginKey string) (*ScanLoginSession, error) {
	return nil, nil
}

func (s *stubExternalRepository) UpdateScanLoginSession(ctx context.Context, session *ScanLoginSession) error {
	return nil
}

func (s *stubExternalRepository) CreateWxPayConfig(ctx context.Context, config *WxPayConfig) error {
	return nil
}

func (s *stubExternalRepository) GetWxPayConfig(ctx context.Context, configType string) (*WxPayConfig, error) {
	return nil, nil
}

func (s *stubExternalRepository) GetWxPayConfigByID(ctx context.Context, id int64) (*WxPayConfig, error) {
	s.lastConfigID = id
	return s.configByID, nil
}

func (s *stubExternalRepository) ListWxPayConfigs(ctx context.Context, page, pageSize int, configType *string) ([]*WxPayConfig, int, error) {
	s.lastListPage = page
	s.lastListPageSize = pageSize
	s.lastListConfigType = configType
	return s.configs, s.total, nil
}

func (s *stubExternalRepository) UpdateWxPayConfig(ctx context.Context, config *WxPayConfig) error {
	s.usedFullConfigUpdate = true
	return nil
}

func (s *stubExternalRepository) UpdateWxPayConfigStatus(ctx context.Context, id int64, status int) error {
	s.lastStatusUpdateID = id
	s.lastStatusUpdate = status
	return nil
}

func (s *stubExternalRepository) DeleteWxPayConfig(ctx context.Context, id int64) error {
	return nil
}

func (s *stubExternalRepository) CreateKookBinding(ctx context.Context, binding *KookBinding) error {
	return nil
}

func (s *stubExternalRepository) GetKookBindingByUserID(ctx context.Context, userID int64) (*KookBinding, error) {
	return nil, nil
}

func (s *stubExternalRepository) DeleteKookBinding(ctx context.Context, userID int64) error {
	return nil
}
