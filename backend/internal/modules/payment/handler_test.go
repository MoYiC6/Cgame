package payment

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/platform/response"

	"github.com/gin-gonic/gin"
)

func TestAdminListPaymentsUsesJavaPagingParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &stubPaymentService{
		adminPayments: []*PaymentRecord{{PaymentNo: "PAY-1", OrderNo: "ORD-1", UserID: 7, Amount: 12.5, Status: "paid", PayMethod: "wxpay"}},
		adminTotal:    9,
	}
	engine := newPaymentHandlerEngine(NewHandler(service))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/admin/payments?pageIndex=2&pageSize=3", nil)
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if service.lastAdminPage != 2 || service.lastAdminPageSize != 3 {
		t.Fatalf("expected admin paging 2/3, got %d/%d", service.lastAdminPage, service.lastAdminPageSize)
	}

	var body response.APIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	data, ok := body.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected response data map, got %#v", body.Data)
	}
	if data["total"] != float64(9) {
		t.Fatalf("expected total 9, got %#v", data["total"])
	}
	list, ok := data["list"].([]any)
	if !ok || len(list) != 1 {
		t.Fatalf("expected one payment in list, got %#v", data["list"])
	}
}

func TestAdminPaymentStatsReturnsAggregateFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &stubPaymentService{
		stats: &PaymentStats{
			TotalAmount:      100,
			TodayAmount:      10,
			MonthAmount:      50,
			PaidCount:        4,
			PendingCount:     2,
			RefundedCount:    1,
			PayMethodAmounts: map[string]float64{"wxpay": 80, "alipay": 20},
		},
	}
	engine := newPaymentHandlerEngine(NewHandler(service))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/admin/payments/stats", nil)
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	var body response.APIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	data, ok := body.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected response data map, got %#v", body.Data)
	}
	if data["totalAmount"] != float64(100) {
		t.Fatalf("expected totalAmount 100, got %#v", data["totalAmount"])
	}
	if data["paidCount"] != float64(4) {
		t.Fatalf("expected paidCount 4, got %#v", data["paidCount"])
	}
}

func newPaymentHandlerEngine(h *Handler) *gin.Engine {
	engine := gin.New()
	api := engine.Group("/api")
	h.RegisterRoutes(api)
	return engine
}

type stubPaymentService struct {
	adminPayments []*PaymentRecord
	adminTotal    int
	stats         *PaymentStats

	lastAdminPage     int
	lastAdminPageSize int
}

func (s *stubPaymentService) Ping(ctx context.Context) (PingResponse, error) {
	return PingResponse{Module: "payment"}, nil
}

func (s *stubPaymentService) CreatePayment(ctx context.Context, userID int64, orderNo string, amount float64, payMethod string) (*PaymentRecord, error) {
	return nil, nil
}

func (s *stubPaymentService) ConfirmPayment(ctx context.Context, paymentNo string) error {
	return nil
}

func (s *stubPaymentService) GetPayment(ctx context.Context, paymentNo string) (*PaymentRecord, error) {
	return nil, nil
}

func (s *stubPaymentService) ListPayments(ctx context.Context, userID int64, page, pageSize int) ([]*PaymentRecord, int, error) {
	return nil, 0, nil
}

func (s *stubPaymentService) ListAdminPayments(ctx context.Context, page, pageSize int) ([]*PaymentRecord, int, error) {
	s.lastAdminPage = page
	s.lastAdminPageSize = pageSize
	return s.adminPayments, s.adminTotal, nil
}

func (s *stubPaymentService) GetPaymentStats(ctx context.Context) (*PaymentStats, error) {
	return s.stats, nil
}
