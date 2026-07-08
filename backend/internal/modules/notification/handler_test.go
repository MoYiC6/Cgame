package notification

import (
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

func TestAdminNotificationInboxListUsesJavaParamsAndPrincipal(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &stubNotificationRepository{
		inboxList: &NotificationInboxList{
			Total:       2,
			UnreadCount: 1,
			Rows: []NotificationInboxItem{
				{ID: 9, NotificationID: 7, Title: "系统通知", Content: ptrString("请处理"), Type: "system", IsRead: 0},
			},
		},
	}
	engine := newNotificationHandlerEngine(NewHandler(NewService(repo, database.NoopTxManager{}), nil))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/admin/notification-inbox?pageIndex=2&pageSize=3&type=system&unreadOnly=true", nil)
	request = request.WithContext(security.WithPrincipal(request.Context(), &security.Principal{UserID: "42"}))
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if repo.lastInboxUserID != 42 || repo.lastInboxPage != 2 || repo.lastInboxPageSize != 3 {
		t.Fatalf("expected user/page/pageSize 42/2/3, got %d/%d/%d", repo.lastInboxUserID, repo.lastInboxPage, repo.lastInboxPageSize)
	}
	if repo.lastInboxType != "system" || repo.lastInboxUnreadOnly == nil || !*repo.lastInboxUnreadOnly {
		t.Fatalf("expected type system and unreadOnly=true, got type=%q unreadOnly=%#v", repo.lastInboxType, repo.lastInboxUnreadOnly)
	}

	var body response.APIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	data, ok := body.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected response data map, got %#v", body.Data)
	}
	if data["total"] != float64(2) || data["unreadCount"] != float64(1) {
		t.Fatalf("expected total/unreadCount 2/1, got %#v", data)
	}
	rows, ok := data["rows"].([]any)
	if !ok || len(rows) != 1 {
		t.Fatalf("expected one Java-compatible row, got %#v", data["rows"])
	}
}

func TestAdminNotificationInboxMarkReadUsesInboxIDAndPrincipal(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &stubNotificationRepository{}
	engine := newNotificationHandlerEngine(NewHandler(NewService(repo, database.NoopTxManager{}), nil))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, "/api/admin/notification-inbox/9/read", nil)
	request = request.WithContext(security.WithPrincipal(request.Context(), &security.Principal{UserID: "42"}))
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if repo.lastMarkInboxUserID != 42 || repo.lastMarkInboxID != 9 {
		t.Fatalf("expected mark read user/id 42/9, got %d/%d", repo.lastMarkInboxUserID, repo.lastMarkInboxID)
	}
}

func TestAdminNotificationInboxMarkAllReadUsesTypeFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &stubNotificationRepository{}
	engine := newNotificationHandlerEngine(NewHandler(NewService(repo, database.NoopTxManager{}), nil))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, "/api/admin/notification-inbox/read-all?type=system", nil)
	request = request.WithContext(security.WithPrincipal(request.Context(), &security.Principal{UserID: "42"}))
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if repo.lastMarkAllInboxUserID != 42 || repo.lastMarkAllInboxType != "system" {
		t.Fatalf("expected mark all user/type 42/system, got %d/%q", repo.lastMarkAllInboxUserID, repo.lastMarkAllInboxType)
	}
}

func newNotificationHandlerEngine(h *Handler) *gin.Engine {
	engine := gin.New()
	api := engine.Group("/api")
	h.RegisterRoutes(api)
	return engine
}

func ptrString(value string) *string {
	return &value
}

type stubNotificationRepository struct {
	inboxList *NotificationInboxList

	lastInboxUserID     int64
	lastInboxPage       int
	lastInboxPageSize   int
	lastInboxType       string
	lastInboxUnreadOnly *bool

	lastMarkInboxUserID    int64
	lastMarkInboxID        int64
	lastMarkAllInboxUserID int64
	lastMarkAllInboxType   string
}

func (s *stubNotificationRepository) CreateNotification(ctx context.Context, n *Notification) error {
	return nil
}

func (s *stubNotificationRepository) GetUserNotifications(ctx context.Context, userID int64, page, pageSize int) ([]*Notification, int, error) {
	return nil, 0, nil
}

func (s *stubNotificationRepository) MarkAsRead(ctx context.Context, userID, notificationID int64) error {
	return nil
}

func (s *stubNotificationRepository) MarkAllAsRead(ctx context.Context, userID int64) error {
	return nil
}

func (s *stubNotificationRepository) GetUnreadCount(ctx context.Context, userID int64) (int, error) {
	return 0, nil
}

func (s *stubNotificationRepository) CreateTodo(ctx context.Context, t *SystemTodo) error {
	return nil
}

func (s *stubNotificationRepository) GetTodos(ctx context.Context, completed *bool) ([]*SystemTodo, error) {
	return nil, nil
}

func (s *stubNotificationRepository) ToggleTodo(ctx context.Context, id int64, completed bool, operator string) error {
	return nil
}

func (s *stubNotificationRepository) DeleteTodos(ctx context.Context, ids []int64) error {
	return nil
}

func (s *stubNotificationRepository) ListInboxNotifications(ctx context.Context, userID int64, page, pageSize int, notificationType string, unreadOnly *bool) (*NotificationInboxList, error) {
	s.lastInboxUserID = userID
	s.lastInboxPage = page
	s.lastInboxPageSize = pageSize
	s.lastInboxType = notificationType
	s.lastInboxUnreadOnly = unreadOnly
	return s.inboxList, nil
}

func (s *stubNotificationRepository) MarkInboxAsRead(ctx context.Context, userID, inboxID int64) error {
	s.lastMarkInboxUserID = userID
	s.lastMarkInboxID = inboxID
	return nil
}

func (s *stubNotificationRepository) MarkAllInboxAsRead(ctx context.Context, userID int64, notificationType string) error {
	s.lastMarkAllInboxUserID = userID
	s.lastMarkAllInboxType = notificationType
	return nil
}

func (s *stubNotificationRepository) GetNotificationByID(ctx context.Context, id int64) (*Notification, error) {
	return nil, nil
}

func (s *stubNotificationRepository) ListAdminNotifications(ctx context.Context, page, pageSize int) ([]*Notification, int, error) {
	return nil, 0, nil
}

func (s *stubNotificationRepository) DeleteNotification(ctx context.Context, id int64) error {
	return nil
}

func (s *stubNotificationRepository) GetNotificationStats(ctx context.Context) (*NotificationStats, error) {
	return nil, nil
}

func (s *stubNotificationRepository) GetSubscribeTemplates(ctx context.Context) ([]*SubscribeTemplate, error) {
	return nil, nil
}

func (s *stubNotificationRepository) RecordSubscribe(ctx context.Context, userID int64, templateID string) error {
	return nil
}

func (s *stubNotificationRepository) GetSubscribeStatus(ctx context.Context, userID int64, templateID string) (*SubscribeStatus, error) {
	return nil, nil
}
