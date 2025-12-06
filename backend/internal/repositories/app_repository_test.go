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

// getAdminUserID returns the ID of the default admin user
func getAdminUserID(ctx context.Context, t *testing.T) uint64 {
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	admin, err := userRepo.GetByEmail(ctx, "admin@example.com")
	require.NoError(t, err)
	require.NotNil(t, admin)
	return admin.ID
}

func TestAppRepository_Create(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewAppRepository(db)
	adminID := getAdminUserID(ctx, t)

	app := &models.App{
		Name:        "Test App",
		Description: "A test application",
		TableName:   "app_data_test_1",
		Icon:        "test",
		CreatedBy:   adminID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = repo.Create(ctx, app)
	require.NoError(t, err)
	assert.NotZero(t, app.ID)
}

func TestAppRepository_GetByID(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewAppRepository(db)
	adminID := getAdminUserID(ctx, t)

	// Create test app
	app := &models.App{
		Name:        "GetByID App",
		Description: "Test",
		TableName:   "app_data_getbyid",
		Icon:        "test",
		CreatedBy:   adminID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	require.NoError(t, repo.Create(ctx, app))

	tests := []struct {
		name     string
		id       uint64
		wantApp  bool
		wantName string
	}{
		{
			name:     "existing app",
			id:       app.ID,
			wantApp:  true,
			wantName: "GetByID App",
		},
		{
			name:    "non-existing app",
			id:      99999,
			wantApp: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByID(ctx, tt.id)
			require.NoError(t, err)
			if tt.wantApp {
				require.NotNil(t, result)
				assert.Equal(t, tt.wantName, result.Name)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

func TestAppRepository_GetByIDWithFields(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	appRepo := repositories.NewAppRepository(db)
	fieldRepo := repositories.NewFieldRepository(db)
	adminID := getAdminUserID(ctx, t)

	// Create test app
	app := &models.App{
		Name:        "App with Fields",
		Description: "Test",
		TableName:   "app_data_with_fields",
		Icon:        "test",
		CreatedBy:   adminID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	require.NoError(t, appRepo.Create(ctx, app))

	// Create test fields
	fields := []models.AppField{
		{
			AppID:        app.ID,
			FieldCode:    "field1",
			FieldName:    "Field 1",
			FieldType:    "TEXT",
			Required:     false,
			DisplayOrder: 1,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			AppID:        app.ID,
			FieldCode:    "field2",
			FieldName:    "Field 2",
			FieldType:    "NUMBER",
			Required:     true,
			DisplayOrder: 2,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}
	require.NoError(t, fieldRepo.CreateBatch(ctx, fields))

	// Get app with fields
	result, err := appRepo.GetByIDWithFields(ctx, app.ID)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "App with Fields", result.Name)
	assert.Len(t, result.Fields, 2)
	assert.Equal(t, "field1", result.Fields[0].FieldCode)
	assert.Equal(t, "field2", result.Fields[1].FieldCode)
}

func TestAppRepository_GetAll(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewAppRepository(db)
	adminID := getAdminUserID(ctx, t)

	// Create multiple apps
	for i := 1; i <= 5; i++ {
		app := &models.App{
			Name:        "App " + string(rune('A'+i-1)),
			Description: "Test",
			TableName:   "app_data_all_" + string(rune('0'+i)),
			Icon:        "test",
			CreatedBy:   adminID,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		require.NoError(t, repo.Create(ctx, app))
	}

	tests := []struct {
		name       string
		page       int
		limit      int
		wantCount  int
		totalCount int64
	}{
		{
			name:       "first page",
			page:       1,
			limit:      3,
			wantCount:  3,
			totalCount: 5,
		},
		{
			name:       "second page",
			page:       2,
			limit:      3,
			wantCount:  2,
			totalCount: 5,
		},
		{
			name:       "all items",
			page:       1,
			limit:      10,
			wantCount:  5,
			totalCount: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apps, total, err := repo.GetAll(ctx, tt.page, tt.limit)
			require.NoError(t, err)
			assert.Len(t, apps, tt.wantCount)
			assert.Equal(t, tt.totalCount, total)
		})
	}
}

func TestAppRepository_GetByUserID(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	appRepo := repositories.NewAppRepository(db)
	userRepo := repositories.NewUserRepository(db)
	adminID := getAdminUserID(ctx, t)

	// Create another user
	user := &models.User{
		Email:        "appowner@example.com",
		PasswordHash: "hash",
		Name:         "App Owner",
		Role:         "user",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, userRepo.Create(ctx, user))

	// Create apps for different users
	for i := 1; i <= 3; i++ {
		app := &models.App{
			Name:        "Admin App " + string(rune('0'+i)),
			Description: "Admin's app",
			TableName:   "app_data_admin_" + string(rune('0'+i)),
			Icon:        "test",
			CreatedBy:   adminID,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		require.NoError(t, appRepo.Create(ctx, app))
	}

	for i := 1; i <= 2; i++ {
		app := &models.App{
			Name:        "User App " + string(rune('0'+i)),
			Description: "User's app",
			TableName:   "app_data_user_" + string(rune('0'+i)),
			Icon:        "test",
			CreatedBy:   user.ID,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		require.NoError(t, appRepo.Create(ctx, app))
	}

	// Get apps by admin
	adminApps, adminTotal, err := appRepo.GetByUserID(ctx, adminID, 1, 10)
	require.NoError(t, err)
	assert.Len(t, adminApps, 3)
	assert.Equal(t, int64(3), adminTotal)

	// Get apps by user
	userApps, userTotal, err := appRepo.GetByUserID(ctx, user.ID, 1, 10)
	require.NoError(t, err)
	assert.Len(t, userApps, 2)
	assert.Equal(t, int64(2), userTotal)
}

func TestAppRepository_Update(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewAppRepository(db)
	adminID := getAdminUserID(ctx, t)

	// Create test app
	app := &models.App{
		Name:        "Original Name",
		Description: "Original Description",
		TableName:   "app_data_update",
		Icon:        "original",
		CreatedBy:   adminID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	require.NoError(t, repo.Create(ctx, app))

	// Update the app
	app.Name = "Updated Name"
	app.Description = "Updated Description"
	app.Icon = "updated"
	err = repo.Update(ctx, app)
	require.NoError(t, err)

	// Verify update
	updated, err := repo.GetByID(ctx, app.ID)
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, "Updated Name", updated.Name)
	assert.Equal(t, "Updated Description", updated.Description)
	assert.Equal(t, "updated", updated.Icon)
}

func TestAppRepository_Delete(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewAppRepository(db)
	adminID := getAdminUserID(ctx, t)

	// Create test app
	app := &models.App{
		Name:        "Delete Me",
		Description: "To be deleted",
		TableName:   "app_data_delete",
		Icon:        "test",
		CreatedBy:   adminID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	require.NoError(t, repo.Create(ctx, app))

	// Delete the app
	err = repo.Delete(ctx, app.ID)
	require.NoError(t, err)

	// Verify deletion
	deleted, err := repo.GetByID(ctx, app.ID)
	require.NoError(t, err)
	assert.Nil(t, deleted)
}

func TestAppRepository_GetTableName(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewAppRepository(db)
	adminID := getAdminUserID(ctx, t)

	// Create test app
	app := &models.App{
		Name:        "Table Name App",
		Description: "Test",
		TableName:   "app_data_tablename",
		Icon:        "test",
		CreatedBy:   adminID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	require.NoError(t, repo.Create(ctx, app))

	tests := []struct {
		name          string
		appID         uint64
		wantTableName string
	}{
		{
			name:          "existing app",
			appID:         app.ID,
			wantTableName: "app_data_tablename",
		},
		{
			name:          "non-existing app",
			appID:         99999,
			wantTableName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tableName, err := repo.GetTableName(ctx, tt.appID)
			require.NoError(t, err)
			assert.Equal(t, tt.wantTableName, tableName)
		})
	}
}
