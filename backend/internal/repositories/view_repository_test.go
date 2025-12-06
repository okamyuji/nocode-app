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

func TestViewRepository_Create(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewViewRepository(db)
	app := createTestApp(ctx, t, "app_data_view_create")

	view := &models.AppView{
		AppID:     app.ID,
		Name:      "Test View",
		ViewType:  "table",
		IsDefault: true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = repo.Create(ctx, view)
	require.NoError(t, err)
	assert.NotZero(t, view.ID)
}

func TestViewRepository_GetByID(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewViewRepository(db)
	app := createTestApp(ctx, t, "app_data_view_getbyid")

	view := &models.AppView{
		AppID:     app.ID,
		Name:      "GetByID View",
		ViewType:  "list",
		IsDefault: false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, repo.Create(ctx, view))

	tests := []struct {
		name     string
		id       uint64
		wantView bool
		wantName string
	}{
		{
			name:     "existing view",
			id:       view.ID,
			wantView: true,
			wantName: "GetByID View",
		},
		{
			name:     "non-existing view",
			id:       99999,
			wantView: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByID(ctx, tt.id)
			require.NoError(t, err)
			if tt.wantView {
				require.NotNil(t, result)
				assert.Equal(t, tt.wantName, result.Name)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

func TestViewRepository_GetByAppID(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewViewRepository(db)
	app := createTestApp(ctx, t, "app_data_view_getbyappid")

	// Create multiple views
	views := []*models.AppView{
		{AppID: app.ID, Name: "View 1", ViewType: "table", IsDefault: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{AppID: app.ID, Name: "View 2", ViewType: "list", IsDefault: false, CreatedAt: time.Now().Add(time.Second), UpdatedAt: time.Now()},
		{AppID: app.ID, Name: "View 3", ViewType: "calendar", IsDefault: false, CreatedAt: time.Now().Add(2 * time.Second), UpdatedAt: time.Now()},
	}
	for _, v := range views {
		require.NoError(t, repo.Create(ctx, v))
	}

	result, err := repo.GetByAppID(ctx, app.ID)
	require.NoError(t, err)
	assert.Len(t, result, 3)
	// Should be ordered by created_at ASC
	assert.Equal(t, "View 1", result[0].Name)
	assert.Equal(t, "View 2", result[1].Name)
	assert.Equal(t, "View 3", result[2].Name)
}

func TestViewRepository_GetDefaultByAppID(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewViewRepository(db)
	app := createTestApp(ctx, t, "app_data_view_default")

	// Create views with one default
	nonDefault := &models.AppView{AppID: app.ID, Name: "Non-Default", ViewType: "table", IsDefault: false, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	defaultView := &models.AppView{AppID: app.ID, Name: "Default View", ViewType: "list", IsDefault: true, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	require.NoError(t, repo.Create(ctx, nonDefault))
	require.NoError(t, repo.Create(ctx, defaultView))

	result, err := repo.GetDefaultByAppID(ctx, app.ID)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "Default View", result.Name)
	assert.True(t, result.IsDefault)
}

func TestViewRepository_GetDefaultByAppID_NoDefault(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewViewRepository(db)
	app := createTestApp(ctx, t, "app_data_view_nodefault")

	// Create view without default
	view := &models.AppView{AppID: app.ID, Name: "Not Default", ViewType: "table", IsDefault: false, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	require.NoError(t, repo.Create(ctx, view))

	result, err := repo.GetDefaultByAppID(ctx, app.ID)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestViewRepository_Update(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewViewRepository(db)
	app := createTestApp(ctx, t, "app_data_view_update")

	view := &models.AppView{
		AppID:     app.ID,
		Name:      "Original Name",
		ViewType:  "table",
		IsDefault: false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, repo.Create(ctx, view))

	// Update view
	view.Name = "Updated Name"
	view.ViewType = "chart"
	view.IsDefault = true
	err = repo.Update(ctx, view)
	require.NoError(t, err)

	// Verify update
	updated, err := repo.GetByID(ctx, view.ID)
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, "Updated Name", updated.Name)
	assert.Equal(t, "chart", string(updated.ViewType))
	assert.True(t, updated.IsDefault)
}

func TestViewRepository_Delete(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewViewRepository(db)
	app := createTestApp(ctx, t, "app_data_view_delete")

	view := &models.AppView{
		AppID:     app.ID,
		Name:      "Delete Me",
		ViewType:  "table",
		IsDefault: false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, repo.Create(ctx, view))

	// Delete view
	err = repo.Delete(ctx, view.ID)
	require.NoError(t, err)

	// Verify deletion
	deleted, err := repo.GetByID(ctx, view.ID)
	require.NoError(t, err)
	assert.Nil(t, deleted)
}

func TestViewRepository_ClearDefaultByAppID(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewViewRepository(db)
	app := createTestApp(ctx, t, "app_data_view_cleardefault")

	// Create multiple default views
	views := []*models.AppView{
		{AppID: app.ID, Name: "Default 1", ViewType: "table", IsDefault: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{AppID: app.ID, Name: "Default 2", ViewType: "list", IsDefault: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{AppID: app.ID, Name: "Non-Default", ViewType: "chart", IsDefault: false, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	for _, v := range views {
		require.NoError(t, repo.Create(ctx, v))
	}

	// Clear all defaults
	err = repo.ClearDefaultByAppID(ctx, app.ID)
	require.NoError(t, err)

	// Verify no defaults
	result, err := repo.GetByAppID(ctx, app.ID)
	require.NoError(t, err)
	for _, v := range result {
		assert.False(t, v.IsDefault, "View %s should not be default", v.Name)
	}
}
