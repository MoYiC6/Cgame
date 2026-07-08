package teacher

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
	db, err := database.NewPgxPool(ctx, config.DBConfig{DSN: "postgres://djclub:djclub_local_pass@localhost:15432/cgame?sslmode=disable"})
	require.NoError(t, err)
	sqlDB, err := db.SQLDB()
	require.NoError(t, err)
	return sqlDB
}

func TestCreateTeacher(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	repo := NewRepository(sqlDB)
	service := NewService(repo)
	ctx := context.Background()

	teacher := &Teacher{
		UserID: 1,
		Name:   strPtr("Test Teacher"),
		Status: intPtr(1),
	}
	id, err := service.CreateTeacher(ctx, teacher)
	require.NoError(t, err)
	assert.Greater(t, id, int64(0))
}

func TestGetTeacher(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	repo := NewRepository(sqlDB)
	service := NewService(repo)
	ctx := context.Background()

	teacher, err := service.GetTeacher(ctx, 1)
	require.NoError(t, err)
	assert.NotNil(t, teacher)
}

func TestCreateTeacherLevel(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	repo := NewRepository(sqlDB)
	service := NewService(repo)
	ctx := context.Background()

	level := &TeacherLevel{Name: "Level 1", MinOrders: 0, CommissionRate: 1.0}
	id, err := service.CreateTeacherLevel(ctx, level)
	require.NoError(t, err)
	assert.Greater(t, id, int64(0))
}

func TestGetTeacherLevels(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	repo := NewRepository(sqlDB)
	service := NewService(repo)
	ctx := context.Background()

	_, err := service.CreateTeacherLevel(ctx, &TeacherLevel{Name: "Level 1", MinOrders: 0, CommissionRate: 1.0})
	require.NoError(t, err)

	levels, err := service.GetTeacherLevels(ctx)
	require.NoError(t, err)
	assert.NotNil(t, levels)
	assert.GreaterOrEqual(t, len(levels), 1)
}

func strPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}
