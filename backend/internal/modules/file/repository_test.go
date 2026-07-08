package file

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

func TestCreateCategory(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	repo := NewRepository(sqlDB)
	service := NewService(repo)
	ctx := context.Background()

	id, err := service.CreateCategory(ctx, &FileCategory{Name: "Test", Sort: 1})
	require.NoError(t, err)
	assert.Greater(t, id, int64(0))
}

func TestListCategories(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	repo := NewRepository(sqlDB)
	service := NewService(repo)
	ctx := context.Background()

	categories, err := service.ListCategories(ctx)
	require.NoError(t, err)
	assert.NotNil(t, categories)
}

func TestCreateFile(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	repo := NewRepository(sqlDB)
	service := NewService(repo)
	ctx := context.Background()

	file := &File{
		URL:          "https://example.com/test.jpg",
		OriginalName: strPtr("test.jpg"),
		Type:         strPtr("image"),
		Size:         int64Ptr(1024),
	}
	err := service.CreateFile(ctx, file)
	require.NoError(t, err)
	assert.Greater(t, file.ID, int64(0))
}

func TestGetFile(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	repo := NewRepository(sqlDB)
	service := NewService(repo)
	ctx := context.Background()

	_, err := service.GetFile(ctx, 99999)
	assert.Error(t, err)
}

func TestDeleteFile(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	repo := NewRepository(sqlDB)
	service := NewService(repo)
	ctx := context.Background()

	file := &File{
		URL:          "https://example.com/delete-test.jpg",
		OriginalName: strPtr("delete-test.jpg"),
	}
	require.NoError(t, service.CreateFile(ctx, file))

	err := service.DeleteFile(ctx, file.ID)
	require.NoError(t, err)
}

func TestCreateCategoryEmptyName(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	repo := NewRepository(sqlDB)
	service := NewService(repo)
	ctx := context.Background()

	_, err := service.CreateCategory(ctx, &FileCategory{})
	assert.Error(t, err)
}

func strPtr(s string) *string {
	return &s
}

func int64Ptr(i int64) *int64 {
	return &i
}
