package visitor

import (
	"context"
	"testing"
	"time"

	"backend/internal/platform/config"
	"backend/internal/platform/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupVisitorDB(t *testing.T) *database.SQLDB {
	t.Helper()
	ctx := context.Background()
	db, err := database.NewPgxPool(ctx, config.DBConfig{DSN: "postgres://djclub:djclub_local_pass@localhost:15432/cgame?sslmode=disable"})
	require.NoError(t, err)
	sqlDB, err := db.SQLDB()
	require.NoError(t, err)
	return sqlDB
}

func TestUpsertSession(t *testing.T) {
	sqlDB := setupVisitorDB(t)
	defer sqlDB.Close()

	repo := NewRepository(sqlDB)
	ctx := context.Background()

	session := &VisitorSession{
		VisitorID:  "visitor-1",
		SessionID:  "session-1",
		Platform:   strPtr("web"),
		Version:    strPtr("1.0.0"),
		UserAgent:  strPtr("Mozilla/5.0"),
	}

	err := repo.UpsertSession(ctx, session)
	require.NoError(t, err)

	session2 := &VisitorSession{
		VisitorID:  "visitor-1",
		SessionID:  "session-1",
		Platform:   strPtr("web"),
		Version:    strPtr("1.0.0"),
		UserAgent:  strPtr("Mozilla/5.0"),
	}
	err = repo.UpsertSession(ctx, session2)
	require.NoError(t, err)
}

func TestBatchInsertPageViews(t *testing.T) {
	sqlDB := setupVisitorDB(t)
	defer sqlDB.Close()

	repo := NewRepository(sqlDB)
	ctx := context.Background()

	views := []*VisitorPageView{
		{VisitorID: "v1", SessionID: "s1", PagePath: "/home", PageTitle: strPtr("Home"), VisitTime: time.Now(), Duration: intPtr(30), Referrer: strPtr("")},
		{VisitorID: "v1", SessionID: "s1", PagePath: "/about", PageTitle: strPtr("About"), VisitTime: time.Now(), Duration: intPtr(15), Referrer: strPtr("/home")},
	}

	err := repo.BatchInsertPageViews(ctx, views)
	require.NoError(t, err)
}

func TestGetDailyStats(t *testing.T) {
	sqlDB := setupVisitorDB(t)
	defer sqlDB.Close()

	repo := NewRepository(sqlDB)
	ctx := context.Background()

	stats, err := repo.GetDailyStats(ctx, time.Now())
	require.NoError(t, err)
	assert.Nil(t, stats)
}

func strPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}
