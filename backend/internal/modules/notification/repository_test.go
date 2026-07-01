package notification

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

func TestCreateNotification(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	repo := NewRepository(sqlDB)
	service := NewService(repo, database.NoopTxManager{})
	ctx := context.Background()

	err := service.CreateNotification(ctx, &Notification{
		Title: "Test",
		Type:  "system",
	})
	require.NoError(t, err)
}

func TestGetUserNotifications(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	repo := NewRepository(sqlDB)
	service := NewService(repo, database.NoopTxManager{})
	ctx := context.Background()

	notifications, total, err := service.GetUserNotifications(ctx, 1, 1, 10)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, 0)
	assert.NotNil(t, notifications)
}

func TestMarkAsRead(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	repo := NewRepository(sqlDB)
	service := NewService(repo, database.NoopTxManager{})
	ctx := context.Background()

	err := service.MarkAsRead(ctx, 1, 99999)
	require.NoError(t, err)
}

func TestMarkAllAsRead(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	repo := NewRepository(sqlDB)
	service := NewService(repo, database.NoopTxManager{})
	ctx := context.Background()

	err := service.MarkAllAsRead(ctx, 1)
	require.NoError(t, err)
}

func TestGetUnreadCount(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	repo := NewRepository(sqlDB)
	service := NewService(repo, database.NoopTxManager{})
	ctx := context.Background()

	count, err := service.GetUnreadCount(ctx, 1)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, 0)
}

func TestCreateTodo(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	repo := NewRepository(sqlDB)
	service := NewService(repo, database.NoopTxManager{})
	ctx := context.Background()

	id, err := service.CreateTodo(ctx, &SystemTodo{Title: "Test Todo"})
	require.NoError(t, err)
	assert.Greater(t, id, int64(0))
}

func TestGetTodos(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	repo := NewRepository(sqlDB)
	service := NewService(repo, database.NoopTxManager{})
	ctx := context.Background()

	todos, err := service.GetTodos(ctx, nil)
	require.NoError(t, err)
	assert.NotNil(t, todos)
}

func TestToggleTodo(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	repo := NewRepository(sqlDB)
	service := NewService(repo, database.NoopTxManager{})
	ctx := context.Background()

	id, err := service.CreateTodo(ctx, &SystemTodo{Title: "Toggle Test"})
	require.NoError(t, err)

	err = service.ToggleTodo(ctx, id, true, "admin")
	require.NoError(t, err)

	err = service.ToggleTodo(ctx, id, false, "admin")
	require.NoError(t, err)
}

func TestDeleteTodos(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	repo := NewRepository(sqlDB)
	service := NewService(repo, database.NoopTxManager{})
	ctx := context.Background()

	err := service.DeleteTodos(ctx, []int64{})
	require.NoError(t, err)
}

func TestCreateNotificationValidation(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	repo := NewRepository(sqlDB)
	service := NewService(repo, database.NoopTxManager{})
	ctx := context.Background()

	err := service.CreateNotification(ctx, &Notification{})
	assert.Error(t, err)

	err = service.CreateNotification(ctx, &Notification{Title: "Test"})
	assert.Error(t, err)
}
