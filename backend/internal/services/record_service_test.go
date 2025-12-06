package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"nocode-app/backend/internal/models"
	"nocode-app/backend/internal/repositories"
	"nocode-app/backend/internal/services"
	"nocode-app/backend/internal/testhelpers/mocks"
)

func TestRecordService_GetRecords(t *testing.T) {
	ctx := context.Background()

	t.Run("successful get records", func(t *testing.T) {
		mockAppRepo := new(mocks.MockAppRepository)
		mockFieldRepo := new(mocks.MockFieldRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)

		fields := []models.AppField{
			{ID: 1, FieldCode: "name", FieldName: "Name", FieldType: "TEXT"},
		}
		records := []models.RecordResponse{
			{ID: 1, Data: models.RecordData{"name": "Record 1"}, CreatedBy: 1, CreatedAt: time.Now().Format(time.RFC3339), UpdatedAt: time.Now().Format(time.RFC3339)},
			{ID: 2, Data: models.RecordData{"name": "Record 2"}, CreatedBy: 1, CreatedAt: time.Now().Format(time.RFC3339), UpdatedAt: time.Now().Format(time.RFC3339)},
		}

		mockAppRepo.On("GetTableName", ctx, uint64(1)).Return("app_data_1", nil)
		mockFieldRepo.On("GetByAppID", ctx, uint64(1)).Return(fields, nil)
		mockDynamicQuery.On("GetRecords", ctx, "app_data_1", fields, mock.AnythingOfType("repositories.RecordQueryOptions")).Return(records, int64(2), nil)

		service := services.NewRecordService(mockAppRepo, mockFieldRepo, mockDynamicQuery)

		opts := repositories.RecordQueryOptions{Page: 1, Limit: 10}
		resp, err := service.GetRecords(ctx, 1, opts)
		require.NoError(t, err)
		assert.Len(t, resp.Records, 2)
		assert.Equal(t, int64(2), resp.Pagination.Total)

		mockAppRepo.AssertExpectations(t)
		mockFieldRepo.AssertExpectations(t)
		mockDynamicQuery.AssertExpectations(t)
	})

	t.Run("app not found", func(t *testing.T) {
		mockAppRepo := new(mocks.MockAppRepository)
		mockFieldRepo := new(mocks.MockFieldRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)

		mockAppRepo.On("GetTableName", ctx, uint64(999)).Return("", nil)

		service := services.NewRecordService(mockAppRepo, mockFieldRepo, mockDynamicQuery)

		opts := repositories.RecordQueryOptions{Page: 1, Limit: 10}
		_, err := service.GetRecords(ctx, 999, opts)
		assert.ErrorIs(t, err, services.ErrAppNotFound)

		mockAppRepo.AssertExpectations(t)
	})
}

func TestRecordService_GetRecord(t *testing.T) {
	ctx := context.Background()

	t.Run("successful get record", func(t *testing.T) {
		mockAppRepo := new(mocks.MockAppRepository)
		mockFieldRepo := new(mocks.MockFieldRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)

		fields := []models.AppField{
			{ID: 1, FieldCode: "name", FieldName: "Name", FieldType: "TEXT"},
		}
		record := &models.RecordResponse{
			ID:        1,
			Data:      models.RecordData{"name": "Record 1"},
			CreatedBy: 1,
			CreatedAt: time.Now().Format(time.RFC3339),
			UpdatedAt: time.Now().Format(time.RFC3339),
		}

		mockAppRepo.On("GetTableName", ctx, uint64(1)).Return("app_data_1", nil)
		mockFieldRepo.On("GetByAppID", ctx, uint64(1)).Return(fields, nil)
		mockDynamicQuery.On("GetRecordByID", ctx, "app_data_1", fields, uint64(1)).Return(record, nil)

		service := services.NewRecordService(mockAppRepo, mockFieldRepo, mockDynamicQuery)

		resp, err := service.GetRecord(ctx, 1, 1)
		require.NoError(t, err)
		assert.Equal(t, uint64(1), resp.ID)
		assert.Equal(t, "Record 1", resp.Data["name"])

		mockAppRepo.AssertExpectations(t)
		mockFieldRepo.AssertExpectations(t)
		mockDynamicQuery.AssertExpectations(t)
	})

	t.Run("record not found", func(t *testing.T) {
		mockAppRepo := new(mocks.MockAppRepository)
		mockFieldRepo := new(mocks.MockFieldRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)

		fields := []models.AppField{}

		mockAppRepo.On("GetTableName", ctx, uint64(1)).Return("app_data_1", nil)
		mockFieldRepo.On("GetByAppID", ctx, uint64(1)).Return(fields, nil)
		mockDynamicQuery.On("GetRecordByID", ctx, "app_data_1", fields, uint64(999)).Return(nil, nil)

		service := services.NewRecordService(mockAppRepo, mockFieldRepo, mockDynamicQuery)

		_, err := service.GetRecord(ctx, 1, 999)
		assert.ErrorIs(t, err, services.ErrRecordNotFound)

		mockAppRepo.AssertExpectations(t)
		mockFieldRepo.AssertExpectations(t)
		mockDynamicQuery.AssertExpectations(t)
	})
}

func TestRecordService_CreateRecord(t *testing.T) {
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		mockAppRepo := new(mocks.MockAppRepository)
		mockFieldRepo := new(mocks.MockFieldRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)

		fields := []models.AppField{
			{ID: 1, FieldCode: "name", FieldName: "Name", FieldType: "TEXT"},
		}
		createdRecord := &models.RecordResponse{
			ID:        1,
			Data:      models.RecordData{"name": "New Record"},
			CreatedBy: 1,
			CreatedAt: time.Now().Format(time.RFC3339),
			UpdatedAt: time.Now().Format(time.RFC3339),
		}

		mockAppRepo.On("GetTableName", ctx, uint64(1)).Return("app_data_1", nil)
		mockFieldRepo.On("GetByAppID", ctx, uint64(1)).Return(fields, nil)
		mockDynamicQuery.On("InsertRecord", ctx, "app_data_1", mock.AnythingOfType("models.RecordData"), uint64(1)).Return(uint64(1), nil)
		mockDynamicQuery.On("GetRecordByID", ctx, "app_data_1", fields, uint64(1)).Return(createdRecord, nil)

		service := services.NewRecordService(mockAppRepo, mockFieldRepo, mockDynamicQuery)

		req := &models.CreateRecordRequest{
			Data: models.RecordData{"name": "New Record"},
		}

		resp, err := service.CreateRecord(ctx, 1, 1, req)
		require.NoError(t, err)
		assert.Equal(t, uint64(1), resp.ID)

		mockAppRepo.AssertExpectations(t)
		mockFieldRepo.AssertExpectations(t)
		mockDynamicQuery.AssertExpectations(t)
	})
}

func TestRecordService_UpdateRecord(t *testing.T) {
	ctx := context.Background()

	t.Run("successful update", func(t *testing.T) {
		mockAppRepo := new(mocks.MockAppRepository)
		mockFieldRepo := new(mocks.MockFieldRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)

		fields := []models.AppField{
			{ID: 1, FieldCode: "name", FieldName: "Name", FieldType: "TEXT"},
		}
		updatedRecord := &models.RecordResponse{
			ID:        1,
			Data:      models.RecordData{"name": "Updated"},
			CreatedBy: 1,
			CreatedAt: time.Now().Format(time.RFC3339),
			UpdatedAt: time.Now().Format(time.RFC3339),
		}

		mockAppRepo.On("GetTableName", ctx, uint64(1)).Return("app_data_1", nil)
		mockFieldRepo.On("GetByAppID", ctx, uint64(1)).Return(fields, nil)
		mockDynamicQuery.On("UpdateRecord", ctx, "app_data_1", uint64(1), mock.AnythingOfType("models.RecordData")).Return(nil)
		mockDynamicQuery.On("GetRecordByID", ctx, "app_data_1", fields, uint64(1)).Return(updatedRecord, nil)

		service := services.NewRecordService(mockAppRepo, mockFieldRepo, mockDynamicQuery)

		req := &models.UpdateRecordRequest{
			Data: models.RecordData{"name": "Updated"},
		}

		resp, err := service.UpdateRecord(ctx, 1, 1, req)
		require.NoError(t, err)
		assert.Equal(t, "Updated", resp.Data["name"])

		mockAppRepo.AssertExpectations(t)
		mockFieldRepo.AssertExpectations(t)
		mockDynamicQuery.AssertExpectations(t)
	})
}

func TestRecordService_DeleteRecord(t *testing.T) {
	ctx := context.Background()

	t.Run("successful delete", func(t *testing.T) {
		mockAppRepo := new(mocks.MockAppRepository)
		mockFieldRepo := new(mocks.MockFieldRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)

		mockAppRepo.On("GetTableName", ctx, uint64(1)).Return("app_data_1", nil)
		mockDynamicQuery.On("DeleteRecord", ctx, "app_data_1", uint64(1)).Return(nil)

		service := services.NewRecordService(mockAppRepo, mockFieldRepo, mockDynamicQuery)

		err := service.DeleteRecord(ctx, 1, 1)
		require.NoError(t, err)

		mockAppRepo.AssertExpectations(t)
		mockDynamicQuery.AssertExpectations(t)
	})
}

func TestRecordService_BulkCreateRecords(t *testing.T) {
	ctx := context.Background()

	t.Run("successful bulk creation", func(t *testing.T) {
		mockAppRepo := new(mocks.MockAppRepository)
		mockFieldRepo := new(mocks.MockFieldRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)

		fields := []models.AppField{
			{ID: 1, FieldCode: "name", FieldName: "Name", FieldType: "TEXT"},
		}

		mockAppRepo.On("GetTableName", ctx, uint64(1)).Return("app_data_1", nil)
		mockFieldRepo.On("GetByAppID", ctx, uint64(1)).Return(fields, nil)
		mockDynamicQuery.On("InsertRecord", ctx, "app_data_1", mock.AnythingOfType("models.RecordData"), uint64(1)).Return(uint64(1), nil).Once()
		mockDynamicQuery.On("InsertRecord", ctx, "app_data_1", mock.AnythingOfType("models.RecordData"), uint64(1)).Return(uint64(2), nil).Once()
		mockDynamicQuery.On("GetRecordByID", ctx, "app_data_1", fields, uint64(1)).Return(&models.RecordResponse{ID: 1, Data: models.RecordData{"name": "R1"}}, nil)
		mockDynamicQuery.On("GetRecordByID", ctx, "app_data_1", fields, uint64(2)).Return(&models.RecordResponse{ID: 2, Data: models.RecordData{"name": "R2"}}, nil)

		service := services.NewRecordService(mockAppRepo, mockFieldRepo, mockDynamicQuery)

		req := &models.BulkCreateRecordRequest{
			Records: []models.RecordData{
				{"name": "R1"},
				{"name": "R2"},
			},
		}

		resp, err := service.BulkCreateRecords(ctx, 1, 1, req)
		require.NoError(t, err)
		assert.Len(t, resp, 2)

		mockAppRepo.AssertExpectations(t)
		mockFieldRepo.AssertExpectations(t)
		mockDynamicQuery.AssertExpectations(t)
	})
}

func TestRecordService_BulkDeleteRecords(t *testing.T) {
	ctx := context.Background()

	t.Run("successful bulk delete", func(t *testing.T) {
		mockAppRepo := new(mocks.MockAppRepository)
		mockFieldRepo := new(mocks.MockFieldRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)

		mockAppRepo.On("GetTableName", ctx, uint64(1)).Return("app_data_1", nil)
		mockDynamicQuery.On("DeleteRecords", ctx, "app_data_1", []uint64{1, 2, 3}).Return(nil)

		service := services.NewRecordService(mockAppRepo, mockFieldRepo, mockDynamicQuery)

		req := &models.BulkDeleteRecordRequest{
			IDs: []uint64{1, 2, 3},
		}

		err := service.BulkDeleteRecords(ctx, 1, req)
		require.NoError(t, err)

		mockAppRepo.AssertExpectations(t)
		mockDynamicQuery.AssertExpectations(t)
	})

	t.Run("app not found", func(t *testing.T) {
		mockAppRepo := new(mocks.MockAppRepository)
		mockFieldRepo := new(mocks.MockFieldRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)

		mockAppRepo.On("GetTableName", ctx, uint64(999)).Return("", nil)

		service := services.NewRecordService(mockAppRepo, mockFieldRepo, mockDynamicQuery)

		req := &models.BulkDeleteRecordRequest{
			IDs: []uint64{1, 2, 3},
		}

		err := service.BulkDeleteRecords(ctx, 999, req)
		assert.ErrorIs(t, err, services.ErrAppNotFound)

		mockAppRepo.AssertExpectations(t)
	})
}

func TestRecordService_CreateRecord_AppNotFound(t *testing.T) {
	ctx := context.Background()

	mockAppRepo := new(mocks.MockAppRepository)
	mockFieldRepo := new(mocks.MockFieldRepository)
	mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)

	mockAppRepo.On("GetTableName", ctx, uint64(999)).Return("", nil)

	service := services.NewRecordService(mockAppRepo, mockFieldRepo, mockDynamicQuery)

	req := &models.CreateRecordRequest{
		Data: models.RecordData{"name": "Test"},
	}

	_, err := service.CreateRecord(ctx, 999, 1, req)
	assert.ErrorIs(t, err, services.ErrAppNotFound)

	mockAppRepo.AssertExpectations(t)
}

func TestRecordService_UpdateRecord_AppNotFound(t *testing.T) {
	ctx := context.Background()

	mockAppRepo := new(mocks.MockAppRepository)
	mockFieldRepo := new(mocks.MockFieldRepository)
	mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)

	mockAppRepo.On("GetTableName", ctx, uint64(999)).Return("", nil)

	service := services.NewRecordService(mockAppRepo, mockFieldRepo, mockDynamicQuery)

	req := &models.UpdateRecordRequest{
		Data: models.RecordData{"name": "Test"},
	}

	_, err := service.UpdateRecord(ctx, 999, 1, req)
	assert.ErrorIs(t, err, services.ErrAppNotFound)

	mockAppRepo.AssertExpectations(t)
}

func TestRecordService_DeleteRecord_AppNotFound(t *testing.T) {
	ctx := context.Background()

	mockAppRepo := new(mocks.MockAppRepository)
	mockFieldRepo := new(mocks.MockFieldRepository)
	mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)

	mockAppRepo.On("GetTableName", ctx, uint64(999)).Return("", nil)

	service := services.NewRecordService(mockAppRepo, mockFieldRepo, mockDynamicQuery)

	err := service.DeleteRecord(ctx, 999, 1)
	assert.ErrorIs(t, err, services.ErrAppNotFound)

	mockAppRepo.AssertExpectations(t)
}

func TestRecordService_GetRecord_AppNotFound(t *testing.T) {
	ctx := context.Background()

	mockAppRepo := new(mocks.MockAppRepository)
	mockFieldRepo := new(mocks.MockFieldRepository)
	mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)

	mockAppRepo.On("GetTableName", ctx, uint64(999)).Return("", nil)

	service := services.NewRecordService(mockAppRepo, mockFieldRepo, mockDynamicQuery)

	_, err := service.GetRecord(ctx, 999, 1)
	assert.ErrorIs(t, err, services.ErrAppNotFound)

	mockAppRepo.AssertExpectations(t)
}

func TestRecordService_BulkCreateRecords_AppNotFound(t *testing.T) {
	ctx := context.Background()

	mockAppRepo := new(mocks.MockAppRepository)
	mockFieldRepo := new(mocks.MockFieldRepository)
	mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)

	mockAppRepo.On("GetTableName", ctx, uint64(999)).Return("", nil)

	service := services.NewRecordService(mockAppRepo, mockFieldRepo, mockDynamicQuery)

	req := &models.BulkCreateRecordRequest{
		Records: []models.RecordData{{"name": "R1"}},
	}

	_, err := service.BulkCreateRecords(ctx, 999, 1, req)
	assert.ErrorIs(t, err, services.ErrAppNotFound)

	mockAppRepo.AssertExpectations(t)
}
