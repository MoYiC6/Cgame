package visitor

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/platform/config"
	"backend/internal/platform/database"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *database.SQLDB {
	t.Helper()
	ctx := context.Background()
	db, err := database.NewPgxPool(ctx, config.DBConfig{DSN: "postgres://backend:backend@localhost:25432/Cgame?sslmode=disable"})
	require.NoError(t, err)
	sqlDB, err := db.SQLDB()
	require.NoError(t, err)
	return sqlDB
}

func TestTrackHandler(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/api/common/visitor/track", bytes.NewBufferString(`{"visitorId":"v1","sessionId":"s1","page":"/home","timestamp":1672531200000}`))
	c.Request.Header.Set("Content-Type", "application/json")

	repo := NewRepository(sqlDB)
	service := NewService(repo)
	handler := NewHandler(service, nil)

	handler.Track(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTrackHandlerInvalidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/api/common/visitor/track", bytes.NewBufferString(`{"visitorId":"","sessionId":"","page":"","timestamp":0}`))
	c.Request.Header.Set("Content-Type", "application/json")

	repo := NewRepository(&database.SQLDB{})
	service := NewService(repo)
	handler := NewHandler(service, nil)

	handler.Track(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBatchTrackHandler(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/api/common/visitor/batch", bytes.NewBufferString(`[{"visitorId":"v1","sessionId":"s1","page":"/home","timestamp":1672531200000}]`))
	c.Request.Header.Set("Content-Type", "application/json")

	repo := NewRepository(sqlDB)
	service := NewService(repo)
	handler := NewHandler(service, nil)

	handler.BatchTrack(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDashboardHandler(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/admin/visitor-stats/dashboard", nil)

	repo := NewRepository(sqlDB)
	service := NewService(repo)
	handler := NewHandler(service, nil)

	handler.Dashboard(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTrendHandler(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/admin/visitor-stats/trend?days=7", nil)

	repo := NewRepository(sqlDB)
	service := NewService(repo)
	handler := NewHandler(service, nil)

	handler.Trend(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTrendHandlerInvalidDays(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/admin/visitor-stats/trend?days=abc", nil)

	repo := NewRepository(sqlDB)
	service := NewService(repo)
	handler := NewHandler(service, nil)

	handler.Trend(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestServiceTrackVisitorValidation(t *testing.T) {
	repo := NewRepository(&database.SQLDB{})
	service := NewService(repo)
	ctx := context.Background()

	resp, err := service.TrackVisitor(ctx, &TrackVisitorRequest{})
	require.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Contains(t, resp.Message, "missing required fields")
}

func TestServiceBatchTrackEmpty(t *testing.T) {
	repo := NewRepository(&database.SQLDB{})
	service := NewService(repo)
	ctx := context.Background()

	resp, err := service.BatchTrack(ctx, nil)
	require.NoError(t, err)
	assert.True(t, resp.Success)
}

func TestServiceTrackVisitorTimestampTooOld(t *testing.T) {
	repo := NewRepository(&database.SQLDB{})
	service := NewService(repo)
	ctx := context.Background()

	oldTimestamp := int64(1672531200000 - (25 * 60 * 60 * 1000))
	resp, err := service.TrackVisitor(ctx, &TrackVisitorRequest{
		VisitorID: "v1", SessionID: "s1", Page: "/home", Timestamp: oldTimestamp,
	})
	require.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Contains(t, resp.Message, "timestamp too old")
}
