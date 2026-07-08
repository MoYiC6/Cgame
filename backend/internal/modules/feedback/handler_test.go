package feedback

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/platform/database"
	"backend/internal/platform/response"
	"backend/internal/platform/security"

	"github.com/gin-gonic/gin"
)

func TestClientSubmitFeedbackUsesPrincipalAndReturnsID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &stubFeedbackRepository{createdID: 77}
	engine := newFeedbackHandlerEngine(NewHandler(NewService(repo, database.NoopTxManager{}), nil))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/client/feedback/submit", bytes.NewBufferString(`{"content":"希望增加更多支付方式","images":["https://cdn.example/a.png"]}`))
	request.Header.Set("Content-Type", "application/json")
	request = request.WithContext(security.WithPrincipal(request.Context(), &security.Principal{UserID: "42"}))
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if repo.created == nil || repo.created.UserID != 42 || repo.created.Content != "希望增加更多支付方式" {
		t.Fatalf("expected submitted feedback for user 42, got %#v", repo.created)
	}

	var body response.APIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	if body.Data != float64(77) {
		t.Fatalf("expected response data id 77, got %#v", body.Data)
	}
}

func TestClientFeedbackListUsesJavaPagingParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &stubFeedbackRepository{
		listResult: &FeedbackPage{Total: 1, Records: []FeedbackVO{{ID: 7, TicketNo: "FB202607080001", Content: "希望增加更多支付方式", Status: FeedbackStatusPending, StatusText: "待处理"}}},
	}
	engine := newFeedbackHandlerEngine(NewHandler(NewService(repo, database.NoopTxManager{}), nil))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/client/feedback/list?pageNum=2&pageSize=3", nil)
	request = request.WithContext(security.WithPrincipal(request.Context(), &security.Principal{UserID: "42"}))
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if repo.lastListUserID != 42 || repo.lastListPage != 2 || repo.lastListPageSize != 3 {
		t.Fatalf("expected user/page/pageSize 42/2/3, got %d/%d/%d", repo.lastListUserID, repo.lastListPage, repo.lastListPageSize)
	}
}

func TestAdminFeedbackRoutesSupportJavaAndDocumentPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &stubFeedbackRepository{
		adminListResult: &FeedbackPage{Total: 1, Records: []FeedbackVO{{ID: 8, TicketNo: "FB202607080002", Content: "页面异常", Status: FeedbackStatusProcessing, StatusText: "处理中"}}},
		detail:          &FeedbackDetailVO{ID: 8, TicketNo: "FB202607080002", Content: "页面异常", Status: FeedbackStatusProcessing, StatusText: "处理中"},
		replyID:         9,
	}
	engine := newFeedbackHandlerEngine(NewHandler(NewService(repo, database.NoopTxManager{}), nil))

	requests := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodGet, "/api/admin/feedback/list?pageNum=2&pageSize=3&status=1&keyword=FB", ""},
		{http.MethodGet, "/api/admin/feedback?pageNum=2&pageSize=3&status=1&keyword=FB", ""},
		{http.MethodGet, "/api/admin/feedback/8", ""},
		{http.MethodPost, "/api/admin/feedback/reply", `{"feedbackId":8,"content":"已收到"}`},
		{http.MethodPost, "/api/admin/feedback/8/reply", `{"content":"继续补充"}`},
		{http.MethodPut, "/api/admin/feedback/status", `{"feedbackId":8,"status":2}`},
		{http.MethodPut, "/api/admin/feedback/8/status", `{"status":2}`},
		{http.MethodDelete, "/api/admin/feedback/8", ""},
	}

	for _, tt := range requests {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
		request.Header.Set("Content-Type", "application/json")
		request = request.WithContext(security.WithPrincipal(request.Context(), &security.Principal{UserID: "99"}))
		engine.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Fatalf("%s %s expected status 200, got %d body=%s", tt.method, tt.path, recorder.Code, recorder.Body.String())
		}
	}

	if repo.lastAdminListPage != 2 || repo.lastAdminListPageSize != 3 || repo.lastAdminStatus == nil || *repo.lastAdminStatus != 1 || repo.lastAdminKeyword != "FB" {
		t.Fatalf("expected admin filters page=2 size=3 status=1 keyword=FB, got page=%d size=%d status=%#v keyword=%q", repo.lastAdminListPage, repo.lastAdminListPageSize, repo.lastAdminStatus, repo.lastAdminKeyword)
	}
	if repo.lastReplyAdminID != 99 || repo.lastReplyFeedbackID != 8 {
		t.Fatalf("expected reply admin/id 99/8, got %d/%d", repo.lastReplyAdminID, repo.lastReplyFeedbackID)
	}
	if repo.lastStatusFeedbackID != 8 || repo.lastStatus != FeedbackStatusResolved {
		t.Fatalf("expected status update 8/resolved, got %d/%d", repo.lastStatusFeedbackID, repo.lastStatus)
	}
	if repo.lastDeleteID != 8 {
		t.Fatalf("expected delete id 8, got %d", repo.lastDeleteID)
	}
}

func newFeedbackHandlerEngine(h *Handler) *gin.Engine {
	engine := gin.New()
	api := engine.Group("/api")
	h.RegisterRoutes(api)
	return engine
}

type stubFeedbackRepository struct {
	createdID int64
	created   *Feedback

	listResult            *FeedbackPage
	lastListUserID        int64
	lastListPage          int
	lastListPageSize      int
	adminListResult       *FeedbackPage
	lastAdminListPage     int
	lastAdminListPageSize int
	lastAdminStatus       *int
	lastAdminKeyword      string
	detail                *FeedbackDetailVO
	replyID               int64
	lastReplyAdminID      int64
	lastReplyFeedbackID   int64
	lastReplyContent      string
	lastStatusFeedbackID  int64
	lastStatus            int
	lastDeleteID          int64
}

func (s *stubFeedbackRepository) CreateFeedback(ctx context.Context, feedback *Feedback) error {
	s.created = feedback
	feedback.ID = s.createdID
	return nil
}

func (s *stubFeedbackRepository) ListUserFeedback(ctx context.Context, userID int64, page, pageSize int) (*FeedbackPage, error) {
	s.lastListUserID = userID
	s.lastListPage = page
	s.lastListPageSize = pageSize
	return s.listResult, nil
}

func (s *stubFeedbackRepository) GetFeedbackDetail(ctx context.Context, userID *int64, id int64) (*FeedbackDetailVO, error) {
	return s.detail, nil
}

func (s *stubFeedbackRepository) ListAdminFeedback(ctx context.Context, page, pageSize int, status *int, keyword string) (*FeedbackPage, error) {
	s.lastAdminListPage = page
	s.lastAdminListPageSize = pageSize
	s.lastAdminStatus = status
	s.lastAdminKeyword = keyword
	return s.adminListResult, nil
}

func (s *stubFeedbackRepository) CreateReply(ctx context.Context, reply *FeedbackReply) error {
	s.lastReplyAdminID = reply.ReplyUserID
	s.lastReplyFeedbackID = reply.FeedbackID
	s.lastReplyContent = reply.Content
	reply.ID = s.replyID
	return nil
}

func (s *stubFeedbackRepository) UpdateStatus(ctx context.Context, id int64, status int) error {
	s.lastStatusFeedbackID = id
	s.lastStatus = status
	return nil
}

func (s *stubFeedbackRepository) DeleteFeedback(ctx context.Context, id int64) error {
	s.lastDeleteID = id
	return nil
}
