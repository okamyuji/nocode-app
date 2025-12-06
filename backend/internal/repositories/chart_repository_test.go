package repositories_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nocode-app/backend/internal/models"
	"nocode-app/backend/internal/repositories"
	"nocode-app/backend/internal/testhelpers"
)

func TestChartRepository_Create(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewChartRepository(db)
	app := createTestApp(ctx, t, "app_data_chart_create")
	adminID := getAdminUserID(ctx, t)

	config := &models.ChartConfig{
		AppID:     app.ID,
		Name:      "Test Chart",
		ChartType: "bar",
		Config:    make(models.ViewConfig), // Empty JSON object
		CreatedBy: adminID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = repo.Create(ctx, config)
	require.NoError(t, err)
	assert.NotZero(t, config.ID)
}

func TestChartRepository_GetByID(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewChartRepository(db)
	app := createTestApp(ctx, t, "app_data_chart_getbyid")
	adminID := getAdminUserID(ctx, t)

	config := &models.ChartConfig{
		AppID:     app.ID,
		Name:      "GetByID Chart",
		ChartType: "pie",
		Config:    make(models.ViewConfig),
		CreatedBy: adminID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, repo.Create(ctx, config))

	tests := []struct {
		name       string
		id         uint64
		wantConfig bool
		wantName   string
	}{
		{
			name:       "existing config",
			id:         config.ID,
			wantConfig: true,
			wantName:   "GetByID Chart",
		},
		{
			name:       "non-existing config",
			id:         99999,
			wantConfig: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByID(ctx, tt.id)
			require.NoError(t, err)
			if tt.wantConfig {
				require.NotNil(t, result)
				assert.Equal(t, tt.wantName, result.Name)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

func TestChartRepository_GetByAppID(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewChartRepository(db)
	app := createTestApp(ctx, t, "app_data_chart_getbyappid")
	adminID := getAdminUserID(ctx, t)

	// Create multiple chart configs
	configs := []*models.ChartConfig{
		{AppID: app.ID, Name: "Chart 1", ChartType: "bar", Config: make(models.ViewConfig), CreatedBy: adminID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{AppID: app.ID, Name: "Chart 2", ChartType: "line", Config: make(models.ViewConfig), CreatedBy: adminID, CreatedAt: time.Now().Add(time.Second), UpdatedAt: time.Now()},
		{AppID: app.ID, Name: "Chart 3", ChartType: "pie", Config: make(models.ViewConfig), CreatedBy: adminID, CreatedAt: time.Now().Add(2 * time.Second), UpdatedAt: time.Now()},
	}
	for _, c := range configs {
		require.NoError(t, repo.Create(ctx, c))
	}

	result, err := repo.GetByAppID(ctx, app.ID)
	require.NoError(t, err)
	assert.Len(t, result, 3)
	// Should be ordered by created_at DESC
	assert.Equal(t, "Chart 3", result[0].Name)
	assert.Equal(t, "Chart 2", result[1].Name)
	assert.Equal(t, "Chart 1", result[2].Name)
}

func TestChartRepository_Update(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewChartRepository(db)
	app := createTestApp(ctx, t, "app_data_chart_update")
	adminID := getAdminUserID(ctx, t)

	config := &models.ChartConfig{
		AppID:     app.ID,
		Name:      "Original Chart",
		ChartType: "bar",
		Config:    make(models.ViewConfig),
		CreatedBy: adminID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, repo.Create(ctx, config))

	// Update config
	config.Name = "Updated Chart"
	config.ChartType = "line"
	err = repo.Update(ctx, config)
	require.NoError(t, err)

	// Verify update
	updated, err := repo.GetByID(ctx, config.ID)
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, "Updated Chart", updated.Name)
	assert.Equal(t, "line", updated.ChartType)
}

func TestChartRepository_Delete(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewChartRepository(db)
	app := createTestApp(ctx, t, "app_data_chart_delete")
	adminID := getAdminUserID(ctx, t)

	config := &models.ChartConfig{
		AppID:     app.ID,
		Name:      "Delete Me",
		ChartType: "bar",
		Config:    make(models.ViewConfig),
		CreatedBy: adminID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, repo.Create(ctx, config))

	// Delete config
	err = repo.Delete(ctx, config.ID)
	require.NoError(t, err)

	// Verify deletion
	deleted, err := repo.GetByID(ctx, config.ID)
	require.NoError(t, err)
	assert.Nil(t, deleted)
}
