package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"nocode-app/backend/internal/models"
	"nocode-app/backend/internal/services"
	"nocode-app/backend/internal/testhelpers/mocks"
)

func TestAppService_CreateApp(t *testing.T) {
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		mockAppRepo := new(mocks.MockAppRepository)
		mockFieldRepo := new(mocks.MockFieldRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)
		mockDataSourceRepo := new(mocks.MockDataSourceRepository)

		createdApp := &models.App{
			ID:          1,
			Name:        "Test App",
			Description: "A test application",
			Icon:        "test",
			TableName:   "app_data_1",
			CreatedBy:   1,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Fields: []models.AppField{
				{ID: 1, FieldCode: "field1", FieldName: "Field 1", FieldType: "TEXT"},
			},
		}

		mockAppRepo.On("Create", ctx, mock.AnythingOfType("*models.App")).Return(nil).Run(func(args mock.Arguments) {
			app := args.Get(1).(*models.App)
			app.ID = 1
		})
		mockAppRepo.On("Update", ctx, mock.AnythingOfType("*models.App")).Return(nil)
		mockFieldRepo.On("CreateBatch", ctx, mock.AnythingOfType("[]models.AppField")).Return(nil)
		mockDynamicQuery.On("CreateTable", ctx, "app_data_1", mock.AnythingOfType("[]models.AppField")).Return(nil)
		mockAppRepo.On("GetByIDWithFields", ctx, uint64(1)).Return(createdApp, nil)

		service := services.NewAppService(mockAppRepo, mockFieldRepo, mockDynamicQuery, mockDataSourceRepo)

		req := &models.CreateAppRequest{
			Name:        "Test App",
			Description: "A test application",
			Icon:        "test",
			Fields: []models.CreateFieldRequest{
				{FieldCode: "field1", FieldName: "Field 1", FieldType: "text"},
			},
		}

		resp, err := service.CreateApp(ctx, 1, req)
		require.NoError(t, err)
		assert.NotZero(t, resp.ID)
		assert.Equal(t, "Test App", resp.Name)

		mockAppRepo.AssertExpectations(t)
		mockFieldRepo.AssertExpectations(t)
		mockDynamicQuery.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockAppRepo := new(mocks.MockAppRepository)
		mockFieldRepo := new(mocks.MockFieldRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)
		mockDataSourceRepo := new(mocks.MockDataSourceRepository)

		mockAppRepo.On("Create", ctx, mock.AnythingOfType("*models.App")).Return(errors.New("db error"))

		service := services.NewAppService(mockAppRepo, mockFieldRepo, mockDynamicQuery, mockDataSourceRepo)

		req := &models.CreateAppRequest{
			Name:        "Test App",
			Description: "A test application",
		}

		_, err := service.CreateApp(ctx, 1, req)
		assert.Error(t, err)

		mockAppRepo.AssertExpectations(t)
	})
}

func TestAppService_GetApp(t *testing.T) {
	ctx := context.Background()

	t.Run("successful get", func(t *testing.T) {
		mockAppRepo := new(mocks.MockAppRepository)
		mockFieldRepo := new(mocks.MockFieldRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)
		mockDataSourceRepo := new(mocks.MockDataSourceRepository)

		app := &models.App{
			ID:          1,
			Name:        "Test App",
			Description: "A test application",
			TableName:   "app_data_1",
			Icon:        "test",
			CreatedBy:   1,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Fields: []models.AppField{
				{ID: 1, FieldCode: "field1", FieldName: "Field 1", FieldType: "TEXT"},
			},
		}

		mockAppRepo.On("GetByIDWithFields", ctx, uint64(1)).Return(app, nil)

		service := services.NewAppService(mockAppRepo, mockFieldRepo, mockDynamicQuery, mockDataSourceRepo)

		resp, err := service.GetApp(ctx, 1)
		require.NoError(t, err)
		assert.Equal(t, "Test App", resp.Name)
		assert.Len(t, resp.Fields, 1)

		mockAppRepo.AssertExpectations(t)
	})

	t.Run("app not found", func(t *testing.T) {
		mockAppRepo := new(mocks.MockAppRepository)
		mockFieldRepo := new(mocks.MockFieldRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)
		mockDataSourceRepo := new(mocks.MockDataSourceRepository)

		mockAppRepo.On("GetByIDWithFields", ctx, uint64(999)).Return(nil, nil)

		service := services.NewAppService(mockAppRepo, mockFieldRepo, mockDynamicQuery, mockDataSourceRepo)

		_, err := service.GetApp(ctx, 999)
		assert.ErrorIs(t, err, services.ErrAppNotFound)

		mockAppRepo.AssertExpectations(t)
	})
}

func TestAppService_GetApps(t *testing.T) {
	ctx := context.Background()

	t.Run("successful get apps", func(t *testing.T) {
		mockAppRepo := new(mocks.MockAppRepository)
		mockFieldRepo := new(mocks.MockFieldRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)
		mockDataSourceRepo := new(mocks.MockDataSourceRepository)

		apps := []models.App{
			{ID: 1, Name: "App 1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: 2, Name: "App 2", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		}

		mockAppRepo.On("GetAll", ctx, 1, 10).Return(apps, int64(2), nil)
		// Mock field count for each app
		mockFieldRepo.On("GetByAppID", ctx, uint64(1)).Return([]models.AppField{
			{ID: 1, AppID: 1, FieldCode: "f1"},
			{ID: 2, AppID: 1, FieldCode: "f2"},
		}, nil)
		mockFieldRepo.On("GetByAppID", ctx, uint64(2)).Return([]models.AppField{
			{ID: 3, AppID: 2, FieldCode: "f3"},
		}, nil)

		service := services.NewAppService(mockAppRepo, mockFieldRepo, mockDynamicQuery, mockDataSourceRepo)

		resp, err := service.GetApps(ctx, 1, 10)
		require.NoError(t, err)
		assert.Len(t, resp.Apps, 2)
		assert.Equal(t, int64(2), resp.Pagination.Total)
		assert.Equal(t, 2, resp.Apps[0].FieldCount)
		assert.Equal(t, 1, resp.Apps[1].FieldCount)

		mockAppRepo.AssertExpectations(t)
		mockFieldRepo.AssertExpectations(t)
	})

	t.Run("empty result", func(t *testing.T) {
		mockAppRepo := new(mocks.MockAppRepository)
		mockFieldRepo := new(mocks.MockFieldRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)
		mockDataSourceRepo := new(mocks.MockDataSourceRepository)

		apps := []models.App{}

		mockAppRepo.On("GetAll", ctx, 1, 10).Return(apps, int64(0), nil)

		service := services.NewAppService(mockAppRepo, mockFieldRepo, mockDynamicQuery, mockDataSourceRepo)

		resp, err := service.GetApps(ctx, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, 1, resp.Pagination.Page)
		assert.Equal(t, 10, resp.Pagination.Limit)
		assert.Empty(t, resp.Apps)

		mockAppRepo.AssertExpectations(t)
	})
}

func TestAppService_UpdateApp(t *testing.T) {
	ctx := context.Background()

	t.Run("successful update", func(t *testing.T) {
		mockAppRepo := new(mocks.MockAppRepository)
		mockFieldRepo := new(mocks.MockFieldRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)
		mockDataSourceRepo := new(mocks.MockDataSourceRepository)

		existingApp := &models.App{
			ID:          1,
			Name:        "Original Name",
			Description: "Original Description",
			TableName:   "app_data_1",
			Icon:        "original",
			CreatedBy:   1,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		mockAppRepo.On("GetByID", ctx, uint64(1)).Return(existingApp, nil)
		mockAppRepo.On("Update", ctx, mock.AnythingOfType("*models.App")).Return(nil)

		service := services.NewAppService(mockAppRepo, mockFieldRepo, mockDynamicQuery, mockDataSourceRepo)

		req := &models.UpdateAppRequest{
			Name:        "Updated Name",
			Description: "Updated Description",
			Icon:        "updated",
		}

		resp, err := service.UpdateApp(ctx, 1, req)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", resp.Name)
		assert.Equal(t, "Updated Description", resp.Description)
		assert.Equal(t, "updated", resp.Icon)

		mockAppRepo.AssertExpectations(t)
	})

	t.Run("app not found for update", func(t *testing.T) {
		mockAppRepo := new(mocks.MockAppRepository)
		mockFieldRepo := new(mocks.MockFieldRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)
		mockDataSourceRepo := new(mocks.MockDataSourceRepository)

		mockAppRepo.On("GetByID", ctx, uint64(999)).Return(nil, nil)

		service := services.NewAppService(mockAppRepo, mockFieldRepo, mockDynamicQuery, mockDataSourceRepo)

		req := &models.UpdateAppRequest{}

		_, err := service.UpdateApp(ctx, 999, req)
		assert.ErrorIs(t, err, services.ErrAppNotFound)

		mockAppRepo.AssertExpectations(t)
	})
}

func TestAppService_DeleteApp(t *testing.T) {
	ctx := context.Background()

	t.Run("successful delete", func(t *testing.T) {
		mockAppRepo := new(mocks.MockAppRepository)
		mockFieldRepo := new(mocks.MockFieldRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)
		mockDataSourceRepo := new(mocks.MockDataSourceRepository)

		app := &models.App{
			ID:        1,
			Name:      "Test App",
			TableName: "app_data_1",
		}

		mockAppRepo.On("GetByID", ctx, uint64(1)).Return(app, nil)
		mockDynamicQuery.On("DropTable", ctx, "app_data_1").Return(nil)
		mockAppRepo.On("Delete", ctx, uint64(1)).Return(nil)

		service := services.NewAppService(mockAppRepo, mockFieldRepo, mockDynamicQuery, mockDataSourceRepo)

		err := service.DeleteApp(ctx, 1)
		require.NoError(t, err)

		mockAppRepo.AssertExpectations(t)
		mockDynamicQuery.AssertExpectations(t)
	})

	t.Run("app not found", func(t *testing.T) {
		mockAppRepo := new(mocks.MockAppRepository)
		mockFieldRepo := new(mocks.MockFieldRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)
		mockDataSourceRepo := new(mocks.MockDataSourceRepository)

		mockAppRepo.On("GetByID", ctx, uint64(999)).Return(nil, nil)

		service := services.NewAppService(mockAppRepo, mockFieldRepo, mockDynamicQuery, mockDataSourceRepo)

		err := service.DeleteApp(ctx, 999)
		assert.ErrorIs(t, err, services.ErrAppNotFound)

		mockAppRepo.AssertExpectations(t)
	})
}

func TestAppService_CreateExternalApp(t *testing.T) {
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		mockAppRepo := new(mocks.MockAppRepository)
		mockFieldRepo := new(mocks.MockFieldRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)
		mockDataSourceRepo := new(mocks.MockDataSourceRepository)

		dataSource := &models.DataSource{
			ID:     1,
			Name:   "Test DB",
			DBType: "postgresql",
		}

		dataSourceID := uint64(1)
		sourceTableName := "users"
		sourceColumnName := "id"
		createdApp := &models.App{
			ID:              1,
			Name:            "External App",
			Description:     "An external app",
			Icon:            "database",
			TableName:       "external_1",
			IsExternal:      true,
			DataSourceID:    &dataSourceID,
			SourceTableName: &sourceTableName,
			CreatedBy:       1,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
			Fields: []models.AppField{
				{ID: 1, FieldCode: "user_id", FieldName: "User ID", FieldType: "NUMBER", SourceColumnName: &sourceColumnName},
			},
		}

		mockDataSourceRepo.On("GetByID", ctx, uint64(1)).Return(dataSource, nil)
		mockAppRepo.On("Create", ctx, mock.AnythingOfType("*models.App")).Return(nil).Run(func(args mock.Arguments) {
			app := args.Get(1).(*models.App)
			app.ID = 1
		})
		mockAppRepo.On("Update", ctx, mock.AnythingOfType("*models.App")).Return(nil)
		mockFieldRepo.On("CreateBatch", ctx, mock.AnythingOfType("[]models.AppField")).Return(nil)
		mockAppRepo.On("GetByIDWithFields", ctx, uint64(1)).Return(createdApp, nil)

		service := services.NewAppService(mockAppRepo, mockFieldRepo, mockDynamicQuery, mockDataSourceRepo)

		req := &models.CreateExternalAppRequest{
			Name:            "External App",
			Description:     "An external app",
			DataSourceID:    1,
			SourceTableName: "users",
			Fields: []models.CreateExternalFieldRequest{
				{SourceColumnName: "id", FieldCode: "user_id", FieldName: "User ID", FieldType: "NUMBER"},
			},
		}

		resp, err := service.CreateExternalApp(ctx, 1, req)
		require.NoError(t, err)
		assert.NotZero(t, resp.ID)
		assert.Equal(t, "External App", resp.Name)
		assert.True(t, resp.IsExternal)

		mockDataSourceRepo.AssertExpectations(t)
		mockAppRepo.AssertExpectations(t)
		mockFieldRepo.AssertExpectations(t)
	})

	t.Run("datasource not found", func(t *testing.T) {
		mockAppRepo := new(mocks.MockAppRepository)
		mockFieldRepo := new(mocks.MockFieldRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)
		mockDataSourceRepo := new(mocks.MockDataSourceRepository)

		mockDataSourceRepo.On("GetByID", ctx, uint64(999)).Return(nil, nil)

		service := services.NewAppService(mockAppRepo, mockFieldRepo, mockDynamicQuery, mockDataSourceRepo)

		req := &models.CreateExternalAppRequest{
			Name:            "External App",
			Description:     "An external app",
			DataSourceID:    999,
			SourceTableName: "users",
			Fields: []models.CreateExternalFieldRequest{
				{SourceColumnName: "id", FieldCode: "user_id", FieldName: "User ID", FieldType: "NUMBER"},
			},
		}

		_, err := service.CreateExternalApp(ctx, 1, req)
		assert.ErrorIs(t, err, services.ErrDataSourceNotFound)

		mockDataSourceRepo.AssertExpectations(t)
	})

	t.Run("datasource repository error", func(t *testing.T) {
		mockAppRepo := new(mocks.MockAppRepository)
		mockFieldRepo := new(mocks.MockFieldRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)
		mockDataSourceRepo := new(mocks.MockDataSourceRepository)

		mockDataSourceRepo.On("GetByID", ctx, uint64(1)).Return(nil, errors.New("db error"))

		service := services.NewAppService(mockAppRepo, mockFieldRepo, mockDynamicQuery, mockDataSourceRepo)

		req := &models.CreateExternalAppRequest{
			Name:            "External App",
			Description:     "An external app",
			DataSourceID:    1,
			SourceTableName: "users",
			Fields: []models.CreateExternalFieldRequest{
				{SourceColumnName: "id", FieldCode: "user_id", FieldName: "User ID", FieldType: "NUMBER"},
			},
		}

		_, err := service.CreateExternalApp(ctx, 1, req)
		assert.Error(t, err)

		mockDataSourceRepo.AssertExpectations(t)
	})
}
