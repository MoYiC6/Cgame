package chat

import (
	"context"
	"testing"

	"backend/internal/platform/config"
	"backend/internal/platform/database"

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

func TestCreateSession(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	repo := NewRepository(sqlDB)
	service := NewService(repo)
	ctx := context.Background()

	session, err := service.GetOrCreateSession(ctx, 1, 2, 3)
	require.NoError(t, err)
	assert.Greater(t, session.ID, int64(0))
	assert.Equal(t, int64(1), session.UserID)
}

func TestGetUserSessions(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	repo := NewRepository(sqlDB)
	service := NewService(repo)
	ctx := context.Background()

	sessions, err := service.GetUserSessions(ctx, 1)
	require.NoError(t, err)
	assert.NotNil(t, sessions)
}

func TestSendMessage(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	repo := NewRepository(sqlDB)
	service := NewService(repo)
	ctx := context.Background()

	session, err := service.GetOrCreateSession(ctx, 1, 2, 3)
	require.NoError(t, err)

	msg, err := service.SendMessage(ctx, session.ID, 1, "user", "hello", strPtr("text"), nil)
	require.NoError(t, err)
	assert.Greater(t, msg.ID, int64(0))
	assert.Equal(t, "hello", msg.Content)
}

func TestGetMessages(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	repo := NewRepository(sqlDB)
	service := NewService(repo)
	ctx := context.Background()

	session, err := service.GetOrCreateSession(ctx, 1, 2, 3)
	require.NoError(t, err)

	messages, total, err := service.GetSessionMessages(ctx, session.ID, 1, 10)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, 0)
	assert.NotNil(t, messages)
}

func TestGetUnreadCount(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	repo := NewRepository(sqlDB)
	service := NewService(repo)
	ctx := context.Background()

	count, err := service.GetUnreadCount(ctx, 1, false)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, 0)
}

func strPtr(s string) *string {
	return &s
}
