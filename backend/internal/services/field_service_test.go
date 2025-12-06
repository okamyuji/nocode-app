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

func TestFieldService_GetFields(t *testing.T) {
	ctx := context.Background()

	t.Run("successful get fields", func(t *testing.T) {
		mockFieldRepo := new(mocks.MockFieldRepository)
		mockAppRepo := new(mocks.MockAppRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)

		fields := []models.AppField{
			{ID: 1, AppID: 1, FieldCode: "field1", FieldName: "Field 1", FieldType: "TEXT", DisplayOrder: 1},
			{ID: 2, AppID: 1, FieldCode: "field2", FieldName: "Field 2", FieldType: "NUMBER", DisplayOrder: 2},
		}

		mockFieldRepo.On("GetByAppID", ctx, uint64(1)).Return(fields, nil)

		service := services.NewFieldService(mockFieldRepo, mockAppRepo, mockDynamicQuery)

		resp, err := service.GetFields(ctx, 1)
		require.NoError(t, err)
		assert.Len(t, resp, 2)
		assert.Equal(t, "field1", resp[0].FieldCode)

		mockFieldRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockFieldRepo := new(mocks.MockFieldRepository)
		mockAppRepo := new(mocks.MockAppRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)

		mockFieldRepo.On("GetByAppID", ctx, uint64(1)).Return(nil, errors.New("db error"))

		service := services.NewFieldService(mockFieldRepo, mockAppRepo, mockDynamicQuery)

		_, err := service.GetFields(ctx, 1)
		assert.Error(t, err)

		mockFieldRepo.AssertExpectations(t)
	})
}

func TestFieldService_CreateField(t *testing.T) {
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		mockFieldRepo := new(mocks.MockFieldRepository)
		mockAppRepo := new(mocks.MockAppRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)

		mockAppRepo.On("GetTableName", ctx, uint64(1)).Return("app_data_1", nil)
		mockFieldRepo.On("FieldCodeExists", ctx, uint64(1), "new_field").Return(false, nil)
		mockFieldRepo.On("GetMaxDisplayOrder", ctx, uint64(1)).Return(5, nil)
		mockFieldRepo.On("Create", ctx, mock.AnythingOfType("*models.AppField")).Return(nil).Run(func(args mock.Arguments) {
			field := args.Get(1).(*models.AppField)
			field.ID = 1
		})
		mockDynamicQuery.On("AddColumn", ctx, "app_data_1", mock.AnythingOfType("*models.AppField")).Return(nil)

		service := services.NewFieldService(mockFieldRepo, mockAppRepo, mockDynamicQuery)

		req := &models.CreateFieldRequest{
			FieldCode: "new_field",
			FieldName: "New Field",
			FieldType: "TEXT",
			Required:  false,
		}

		resp, err := service.CreateField(ctx, 1, req)
		require.NoError(t, err)
		assert.Equal(t, "new_field", resp.FieldCode)
		assert.Equal(t, 6, resp.DisplayOrder)

		mockFieldRepo.AssertExpectations(t)
		mockAppRepo.AssertExpectations(t)
		mockDynamicQuery.AssertExpectations(t)
	})

	t.Run("field code already exists", func(t *testing.T) {
		mockFieldRepo := new(mocks.MockFieldRepository)
		mockAppRepo := new(mocks.MockAppRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)

		mockAppRepo.On("GetTableName", ctx, uint64(1)).Return("app_data_1", nil)
		mockFieldRepo.On("FieldCodeExists", ctx, uint64(1), "existing_field").Return(true, nil)

		service := services.NewFieldService(mockFieldRepo, mockAppRepo, mockDynamicQuery)

		req := &models.CreateFieldRequest{
			FieldCode: "existing_field",
			FieldName: "Existing Field",
			FieldType: "TEXT",
		}

		_, err := service.CreateField(ctx, 1, req)
		assert.ErrorIs(t, err, services.ErrFieldCodeExists)

		mockFieldRepo.AssertExpectations(t)
		mockAppRepo.AssertExpectations(t)
	})

	t.Run("app not found", func(t *testing.T) {
		mockFieldRepo := new(mocks.MockFieldRepository)
		mockAppRepo := new(mocks.MockAppRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)

		mockAppRepo.On("GetTableName", ctx, uint64(999)).Return("", nil)

		service := services.NewFieldService(mockFieldRepo, mockAppRepo, mockDynamicQuery)

		req := &models.CreateFieldRequest{
			FieldCode: "new_field",
			FieldName: "New Field",
			FieldType: "TEXT",
		}

		_, err := service.CreateField(ctx, 999, req)
		assert.ErrorIs(t, err, services.ErrAppNotFound)

		mockAppRepo.AssertExpectations(t)
	})
}

func TestFieldService_UpdateField(t *testing.T) {
	ctx := context.Background()

	t.Run("successful update", func(t *testing.T) {
		mockFieldRepo := new(mocks.MockFieldRepository)
		mockAppRepo := new(mocks.MockAppRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)

		field := &models.AppField{
			ID:           1,
			AppID:        1,
			FieldCode:    "field1",
			FieldName:    "Original Name",
			FieldType:    "TEXT",
			Required:     false,
			DisplayOrder: 1,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		mockFieldRepo.On("GetByID", ctx, uint64(1)).Return(field, nil)
		mockFieldRepo.On("Update", ctx, mock.AnythingOfType("*models.AppField")).Return(nil)

		service := services.NewFieldService(mockFieldRepo, mockAppRepo, mockDynamicQuery)

		required := true
		req := &models.UpdateFieldRequest{
			FieldName: "Updated Name",
			Required:  &required,
		}

		resp, err := service.UpdateField(ctx, 1, req)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", resp.FieldName)
		assert.True(t, resp.Required)

		mockFieldRepo.AssertExpectations(t)
	})

	t.Run("field not found", func(t *testing.T) {
		mockFieldRepo := new(mocks.MockFieldRepository)
		mockAppRepo := new(mocks.MockAppRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)

		mockFieldRepo.On("GetByID", ctx, uint64(999)).Return(nil, nil)

		service := services.NewFieldService(mockFieldRepo, mockAppRepo, mockDynamicQuery)

		req := &models.UpdateFieldRequest{}

		_, err := service.UpdateField(ctx, 999, req)
		assert.ErrorIs(t, err, services.ErrFieldNotFound)

		mockFieldRepo.AssertExpectations(t)
	})
}

func TestFieldService_DeleteField(t *testing.T) {
	ctx := context.Background()

	t.Run("successful delete", func(t *testing.T) {
		mockFieldRepo := new(mocks.MockFieldRepository)
		mockAppRepo := new(mocks.MockAppRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)

		field := &models.AppField{ID: 1, AppID: 1, FieldCode: "field1"}

		mockFieldRepo.On("GetByID", ctx, uint64(1)).Return(field, nil)
		mockAppRepo.On("GetTableName", ctx, uint64(1)).Return("app_data_1", nil)
		mockDynamicQuery.On("DropColumn", ctx, "app_data_1", "field1").Return(nil)
		mockFieldRepo.On("Delete", ctx, uint64(1)).Return(nil)

		service := services.NewFieldService(mockFieldRepo, mockAppRepo, mockDynamicQuery)

		err := service.DeleteField(ctx, 1, 1)
		require.NoError(t, err)

		mockFieldRepo.AssertExpectations(t)
		mockAppRepo.AssertExpectations(t)
		mockDynamicQuery.AssertExpectations(t)
	})

	t.Run("field not found", func(t *testing.T) {
		mockFieldRepo := new(mocks.MockFieldRepository)
		mockAppRepo := new(mocks.MockAppRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)

		mockFieldRepo.On("GetByID", ctx, uint64(999)).Return(nil, nil)

		service := services.NewFieldService(mockFieldRepo, mockAppRepo, mockDynamicQuery)

		err := service.DeleteField(ctx, 1, 999)
		assert.ErrorIs(t, err, services.ErrFieldNotFound)

		mockFieldRepo.AssertExpectations(t)
	})
}

func TestFieldService_UpdateFieldOrder(t *testing.T) {
	ctx := context.Background()

	t.Run("successful order update", func(t *testing.T) {
		mockFieldRepo := new(mocks.MockFieldRepository)
		mockAppRepo := new(mocks.MockAppRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)

		mockFieldRepo.On("UpdateOrder", ctx, mock.AnythingOfType("[]models.FieldOrderItem")).Return(nil)

		service := services.NewFieldService(mockFieldRepo, mockAppRepo, mockDynamicQuery)

		req := &models.UpdateFieldOrderRequest{
			Fields: []models.FieldOrderItem{
				{ID: 1, DisplayOrder: 2},
				{ID: 2, DisplayOrder: 1},
			},
		}

		err := service.UpdateFieldOrder(ctx, 1, req)
		require.NoError(t, err)

		mockFieldRepo.AssertExpectations(t)
	})
}

func TestFieldService_DeleteField_AppNotFound(t *testing.T) {
	ctx := context.Background()

	mockFieldRepo := new(mocks.MockFieldRepository)
	mockAppRepo := new(mocks.MockAppRepository)
	mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)

	field := &models.AppField{ID: 1, AppID: 999, FieldCode: "field1"}

	mockFieldRepo.On("GetByID", ctx, uint64(1)).Return(field, nil)
	mockAppRepo.On("GetTableName", ctx, uint64(999)).Return("", nil)

	service := services.NewFieldService(mockFieldRepo, mockAppRepo, mockDynamicQuery)

	err := service.DeleteField(ctx, 999, 1)
	assert.ErrorIs(t, err, services.ErrAppNotFound)

	mockFieldRepo.AssertExpectations(t)
	mockAppRepo.AssertExpectations(t)
}

func TestFieldService_UpdateField_DisplayOrder(t *testing.T) {
	ctx := context.Background()

	mockFieldRepo := new(mocks.MockFieldRepository)
	mockAppRepo := new(mocks.MockAppRepository)
	mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)

	field := &models.AppField{
		ID:           1,
		AppID:        1,
		FieldCode:    "field1",
		FieldName:    "Original Name",
		FieldType:    "TEXT",
		Required:     false,
		DisplayOrder: 1,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	mockFieldRepo.On("GetByID", ctx, uint64(1)).Return(field, nil)
	mockFieldRepo.On("Update", ctx, mock.AnythingOfType("*models.AppField")).Return(nil)

	service := services.NewFieldService(mockFieldRepo, mockAppRepo, mockDynamicQuery)

	newOrder := 5
	req := &models.UpdateFieldRequest{
		DisplayOrder: &newOrder,
	}

	resp, err := service.UpdateField(ctx, 1, req)
	require.NoError(t, err)
	assert.Equal(t, 5, resp.DisplayOrder)

	mockFieldRepo.AssertExpectations(t)
}
