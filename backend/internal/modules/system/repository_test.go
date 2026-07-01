package system

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

func TestSystemSettings(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	repo := NewRepository(sqlDB)
	service := NewService(repo)
	ctx := context.Background()

	_, err := service.GetSetting(ctx, "nonexistent.key")
	require.NoError(t, err)

	err = service.SetSetting(ctx, "test.key", "test-value")
	require.NoError(t, err)

	value, err := service.GetSetting(ctx, "test.key")
	require.NoError(t, err)
	assert.Equal(t, "test-value", value)
}

func TestPartnerConfig(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	repo := NewRepository(sqlDB)
	service := NewService(repo)
	ctx := context.Background()

	_, err := service.GetPartnerConfig(ctx, "nonexistent")
	require.NoError(t, err)

	err = service.SetPartnerConfig(ctx, "test.partner", "true")
	require.NoError(t, err)

	value, err := service.GetPartnerConfig(ctx, "test.partner")
	require.NoError(t, err)
	assert.Equal(t, "true", value)
}

func TestFaceIdConfig(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	repo := NewRepository(sqlDB)
	service := NewService(repo)
	ctx := context.Background()

	config, err := service.GetActiveFaceIdConfig(ctx)
	require.NoError(t, err)
	assert.Nil(t, config)

	id, err := service.CreateFaceIdConfig(ctx, &FaceIdConfig{
		SecretID:  "test-secret-id",
		SecretKey: "test-secret-key",
		RuleID:    "test-rule-id",
	})
	require.NoError(t, err)
	assert.Greater(t, id, int64(0))

	err = service.DeleteFaceIdConfig(ctx, id)
	require.NoError(t, err)
}

func TestRealNameVerifyLog(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	repo := NewRepository(sqlDB)
	service := NewService(repo)
	ctx := context.Background()

	log := &RealNameVerifyLog{
		UserID:      1,
		EventType:   "MANUAL_SUBMITTED",
		OperatorID:  int64Ptr(1),
		OperatorType: strPtr("admin"),
	}
	err := service.CreateRealNameVerifyLog(ctx, log)
	require.NoError(t, err)
	assert.Greater(t, log.ID, int64(0))

	logs, total, err := service.ListRealNameVerifyLogs(ctx, nil, "", 1, 10)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, 1)
	assert.NotEmpty(t, logs)
}

func int64Ptr(i int64) *int64 {
	return &i
}
