package coupon

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend/internal/platform/database"
	"backend/internal/platform/response"
	"backend/internal/platform/security"

	"github.com/gin-gonic/gin"
)

func TestClientCouponRoutesUsePrincipal(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &stubCouponRepository{
		availableCoupons: []CouponVO{{ID: 7, Name: "新人券", Type: CouponTypeFixed, FaceValue: 20, MinOrderAmount: 100, Claimable: true}},
		myCoupons:        []UserCouponVO{{ID: 11, CouponID: 7, Name: "新人券", Status: UserCouponStatusAvailable}},
		claimID:          11,
	}
	engine := newCouponHandlerEngine(NewHandler(NewService(repo, database.NoopTxManager{}), nil))

	requests := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/client/coupon/available"},
		{http.MethodGet, "/api/client/coupon/my?status=0"},
		{http.MethodPost, "/api/client/coupon/claim/7"},
	}

	for _, tt := range requests {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(tt.method, tt.path, nil)
		request = request.WithContext(security.WithPrincipal(request.Context(), &security.Principal{UserID: "42"}))
		engine.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Fatalf("%s %s expected status 200, got %d body=%s", tt.method, tt.path, recorder.Code, recorder.Body.String())
		}
	}

	if repo.lastAvailableUserID != 42 {
		t.Fatalf("expected available coupons user 42, got %d", repo.lastAvailableUserID)
	}
	if repo.lastMyUserID != 42 || repo.lastMyStatus == nil || *repo.lastMyStatus != UserCouponStatusAvailable {
		t.Fatalf("expected my coupons user/status 42/0, got user=%d status=%#v", repo.lastMyUserID, repo.lastMyStatus)
	}
	if repo.lastClaimUserID != 42 || repo.lastClaimCouponID != 7 {
		t.Fatalf("expected claim user/coupon 42/7, got %d/%d", repo.lastClaimUserID, repo.lastClaimCouponID)
	}
}

func TestClientClaimCouponReturnsUserCouponID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &stubCouponRepository{claimID: 123}
	engine := newCouponHandlerEngine(NewHandler(NewService(repo, database.NoopTxManager{}), nil))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/client/coupon/claim/7", nil)
	request = request.WithContext(security.WithPrincipal(request.Context(), &security.Principal{UserID: "42"}))
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	var body response.APIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	if body.Data != float64(123) {
		t.Fatalf("expected response data id 123, got %#v", body.Data)
	}
}

func TestAdminCouponRoutesUseJavaQueryAndPathParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Date(2026, 7, 8, 12, 0, 0, 0, time.UTC)
	repo := &stubCouponRepository{
		adminPage: &CouponPage{Total: 1, Records: []AdminCouponVO{{ID: 7, Name: "夏日券", Type: CouponTypeFixed, Status: CouponStatusAvailable, CreatedAt: now}}},
		createID:  7,
		stats:     &CouponStats{TotalCoupons: 3, EnabledCoupons: 2, ClaimedCoupons: 9, UsedCoupons: 4},
	}
	engine := newCouponHandlerEngine(NewHandler(NewService(repo, database.NoopTxManager{}), nil))

	requests := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodGet, "/api/admin/coupon?pageNum=2&pageSize=3&name=%E5%A4%8F&type=1&enabled=true&status=available&isPermanent=false", ""},
		{http.MethodPost, "/api/admin/coupon", `{"name":"夏日券","type":1,"faceValue":20,"minOrderAmount":100,"totalQuantity":50,"perUserLimit":1,"validDays":7,"startTime":"2026-07-08T00:00:00Z","endTime":"2026-08-08T00:00:00Z","applicableScope":"[\"all\"]","distributionMode":2,"enabled":true,"isPermanent":false}`},
		{http.MethodPut, "/api/admin/coupon/7", `{"name":"夏日券-更新","enabled":false}`},
		{http.MethodDelete, "/api/admin/coupon/7", ""},
		{http.MethodGet, "/api/admin/coupon/stats", ""},
	}

	for _, tt := range requests {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
		request.Header.Set("Content-Type", "application/json")
		engine.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Fatalf("%s %s expected status 200, got %d body=%s", tt.method, tt.path, recorder.Code, recorder.Body.String())
		}
	}

	if repo.lastQuery.PageNum != 2 || repo.lastQuery.PageSize != 3 || repo.lastQuery.Name != "夏" || repo.lastQuery.Type == nil || *repo.lastQuery.Type != CouponTypeFixed {
		t.Fatalf("unexpected admin query: %#v", repo.lastQuery)
	}
	if repo.lastQuery.Enabled == nil || *repo.lastQuery.Enabled != true || repo.lastQuery.Status != CouponStatusAvailable || repo.lastQuery.IsPermanent == nil || *repo.lastQuery.IsPermanent != false {
		t.Fatalf("unexpected admin bool/status query: %#v", repo.lastQuery)
	}
	if repo.created == nil || repo.created.Name != "夏日券" || repo.created.Type != CouponTypeFixed || repo.created.FaceValue != 20 || repo.created.DistributionMode != CouponDistributionPublic {
		t.Fatalf("unexpected created coupon: %#v", repo.created)
	}
	if repo.lastUpdateID != 7 || repo.updated == nil || repo.updated.Name == nil || *repo.updated.Name != "夏日券-更新" || repo.updated.Enabled == nil || *repo.updated.Enabled != false {
		t.Fatalf("unexpected update id/request: id=%d req=%#v", repo.lastUpdateID, repo.updated)
	}
	if repo.lastDeleteID != 7 {
		t.Fatalf("expected delete id 7, got %d", repo.lastDeleteID)
	}
}

func TestAdminCreateCouponKeepsPrivateDistributionMode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &stubCouponRepository{createID: 7}
	engine := newCouponHandlerEngine(NewHandler(NewService(repo, database.NoopTxManager{}), nil))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/admin/coupon", bytes.NewBufferString(`{"name":"定向券","type":1,"faceValue":20,"minOrderAmount":100,"totalQuantity":50,"perUserLimit":1,"validDays":7,"startTime":"2026-07-08T00:00:00","endTime":"2026-08-08T00:00:00","distributionMode":0}`))
	request.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if repo.created == nil || repo.created.DistributionMode != CouponDistributionPrivate {
		t.Fatalf("expected private distribution mode to be preserved, got %#v", repo.created)
	}
}

func newCouponHandlerEngine(h *Handler) *gin.Engine {
	engine := gin.New()
	api := engine.Group("/api")
	h.RegisterRoutes(api)
	return engine
}

type stubCouponRepository struct {
	availableCoupons    []CouponVO
	lastAvailableUserID int64

	myCoupons    []UserCouponVO
	lastMyUserID int64
	lastMyStatus *int

	claimID           int64
	lastClaimUserID   int64
	lastClaimCouponID int64

	adminPage CouponPageResult
	lastQuery CouponQuery

	createID int64
	created  *CouponCreateRequest

	lastUpdateID int64
	updated      *CouponUpdateRequest

	lastDeleteID int64
	stats        *CouponStats
}

func (s *stubCouponRepository) ListAvailableCoupons(ctx context.Context, userID int64) ([]CouponVO, error) {
	s.lastAvailableUserID = userID
	return s.availableCoupons, nil
}

func (s *stubCouponRepository) ListUserCoupons(ctx context.Context, userID int64, status *int) ([]UserCouponVO, error) {
	s.lastMyUserID = userID
	s.lastMyStatus = status
	return s.myCoupons, nil
}

func (s *stubCouponRepository) ClaimCoupon(ctx context.Context, userID, couponID int64) (int64, error) {
	s.lastClaimUserID = userID
	s.lastClaimCouponID = couponID
	return s.claimID, nil
}

func (s *stubCouponRepository) ListAdminCoupons(ctx context.Context, query CouponQuery) (CouponPageResult, error) {
	s.lastQuery = query
	return s.adminPage, nil
}

func (s *stubCouponRepository) CreateCoupon(ctx context.Context, req CouponCreateRequest) (int64, error) {
	s.created = &req
	return s.createID, nil
}

func (s *stubCouponRepository) UpdateCoupon(ctx context.Context, id int64, req CouponUpdateRequest) error {
	s.lastUpdateID = id
	s.updated = &req
	return nil
}

func (s *stubCouponRepository) DeleteCoupon(ctx context.Context, id int64) error {
	s.lastDeleteID = id
	return nil
}

func (s *stubCouponRepository) GetStats(ctx context.Context) (*CouponStats, error) {
	return s.stats, nil
}
