package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"nocode-app/backend/internal/handlers"
	"nocode-app/backend/internal/middleware"
	"nocode-app/backend/internal/models"
	"nocode-app/backend/internal/services"
	"nocode-app/backend/internal/testhelpers/mocks"
	"nocode-app/backend/internal/utils"
)

// recordContextWithClaims creates a context with JWT claims for testing
func recordContextWithClaims(ctx context.Context, userID uint64) context.Context {
	claims := &utils.JWTClaims{UserID: userID}
	return context.WithValue(ctx, middleware.UserContextKey, claims)
}

func TestRecordHandler_List(t *testing.T) {
	validator := utils.NewValidator()

	t.Run("successful list records", func(t *testing.T) {
		mockService := new(mocks.MockRecordService)
		handler := handlers.NewRecordHandler(mockService, validator)

		now := time.Now().Format(time.RFC3339)
		resp := &models.RecordListResponse{
			Records: []models.RecordResponse{
				{ID: 1, Data: models.RecordData{"name": "Record 1"}, CreatedAt: now, UpdatedAt: now},
				{ID: 2, Data: models.RecordData{"name": "Record 2"}, CreatedAt: now, UpdatedAt: now},
			},
			Pagination: &models.Pagination{Total: 2, Page: 1, Limit: 20},
		}

		mockService.On("GetRecords", mock.Anything, uint64(1), mock.AnythingOfType("repositories.RecordQueryOptions")).Return(resp, nil)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/apps/1/records", nil)
		rr := httptest.NewRecorder()

		handler.List(rr, httpReq)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result models.RecordListResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Len(t, result.Records, 2)

		mockService.AssertExpectations(t)
	})

	t.Run("app not found", func(t *testing.T) {
		mockService := new(mocks.MockRecordService)
		handler := handlers.NewRecordHandler(mockService, validator)

		mockService.On("GetRecords", mock.Anything, uint64(999), mock.AnythingOfType("repositories.RecordQueryOptions")).Return(nil, services.ErrAppNotFound)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/apps/999/records", nil)
		rr := httptest.NewRecorder()

		handler.List(rr, httpReq)

		assert.Equal(t, http.StatusNotFound, rr.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("invalid app id", func(t *testing.T) {
		mockService := new(mocks.MockRecordService)
		handler := handlers.NewRecordHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/apps/invalid/records", nil)
		rr := httptest.NewRecorder()

		handler.List(rr, httpReq)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("method not allowed", func(t *testing.T) {
		mockService := new(mocks.MockRecordService)
		handler := handlers.NewRecordHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/apps/1/records", nil)
		rr := httptest.NewRecorder()

		handler.List(rr, httpReq)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}

func TestRecordHandler_Create(t *testing.T) {
	validator := utils.NewValidator()

	t.Run("successful create record", func(t *testing.T) {
		mockService := new(mocks.MockRecordService)
		handler := handlers.NewRecordHandler(mockService, validator)

		req := models.CreateRecordRequest{
			Data: models.RecordData{"name": "New Record"},
		}
		now := time.Now().Format(time.RFC3339)
		resp := &models.RecordResponse{
			ID:        1,
			Data:      models.RecordData{"name": "New Record"},
			CreatedBy: 1,
			CreatedAt: now,
			UpdatedAt: now,
		}

		mockService.On("CreateRecord", mock.Anything, uint64(1), uint64(1), mock.AnythingOfType("*models.CreateRecordRequest")).Return(resp, nil)

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/apps/1/records", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq = httpReq.WithContext(recordContextWithClaims(httpReq.Context(), 1))
		rr := httptest.NewRecorder()

		handler.Create(rr, httpReq)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var result models.RecordResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Equal(t, "New Record", result.Data["name"])

		mockService.AssertExpectations(t)
	})

	t.Run("unauthorized", func(t *testing.T) {
		mockService := new(mocks.MockRecordService)
		handler := handlers.NewRecordHandler(mockService, validator)

		req := models.CreateRecordRequest{Data: models.RecordData{"name": "Test"}}
		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/apps/1/records", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.Create(rr, httpReq)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("method not allowed", func(t *testing.T) {
		mockService := new(mocks.MockRecordService)
		handler := handlers.NewRecordHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/apps/1/records", nil)
		rr := httptest.NewRecorder()

		handler.Create(rr, httpReq)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}

func TestRecordHandler_Get(t *testing.T) {
	validator := utils.NewValidator()

	t.Run("successful get record", func(t *testing.T) {
		mockService := new(mocks.MockRecordService)
		handler := handlers.NewRecordHandler(mockService, validator)

		now := time.Now().Format(time.RFC3339)
		resp := &models.RecordResponse{
			ID:        1,
			Data:      models.RecordData{"name": "Record 1"},
			CreatedBy: 1,
			CreatedAt: now,
			UpdatedAt: now,
		}

		mockService.On("GetRecord", mock.Anything, uint64(1), uint64(1)).Return(resp, nil)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/apps/1/records/1", nil)
		rr := httptest.NewRecorder()

		handler.Get(rr, httpReq)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result models.RecordResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Equal(t, uint64(1), result.ID)

		mockService.AssertExpectations(t)
	})

	t.Run("record not found", func(t *testing.T) {
		mockService := new(mocks.MockRecordService)
		handler := handlers.NewRecordHandler(mockService, validator)

		mockService.On("GetRecord", mock.Anything, uint64(1), uint64(999)).Return(nil, services.ErrRecordNotFound)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/apps/1/records/999", nil)
		rr := httptest.NewRecorder()

		handler.Get(rr, httpReq)

		assert.Equal(t, http.StatusNotFound, rr.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("invalid record id", func(t *testing.T) {
		mockService := new(mocks.MockRecordService)
		handler := handlers.NewRecordHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/apps/1/records/invalid", nil)
		rr := httptest.NewRecorder()

		handler.Get(rr, httpReq)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("method not allowed", func(t *testing.T) {
		mockService := new(mocks.MockRecordService)
		handler := handlers.NewRecordHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/apps/1/records/1", nil)
		rr := httptest.NewRecorder()

		handler.Get(rr, httpReq)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}

func TestRecordHandler_Update(t *testing.T) {
	validator := utils.NewValidator()

	t.Run("successful update record", func(t *testing.T) {
		mockService := new(mocks.MockRecordService)
		handler := handlers.NewRecordHandler(mockService, validator)

		req := models.UpdateRecordRequest{
			Data: models.RecordData{"name": "Updated Record"},
		}
		now := time.Now().Format(time.RFC3339)
		resp := &models.RecordResponse{
			ID:        1,
			Data:      models.RecordData{"name": "Updated Record"},
			CreatedBy: 1,
			CreatedAt: now,
			UpdatedAt: now,
		}

		mockService.On("UpdateRecord", mock.Anything, uint64(1), uint64(1), mock.AnythingOfType("*models.UpdateRecordRequest")).Return(resp, nil)

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPut, "/api/v1/apps/1/records/1", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.Update(rr, httpReq)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result models.RecordResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Equal(t, "Updated Record", result.Data["name"])

		mockService.AssertExpectations(t)
	})

	t.Run("method not allowed", func(t *testing.T) {
		mockService := new(mocks.MockRecordService)
		handler := handlers.NewRecordHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/apps/1/records/1", nil)
		rr := httptest.NewRecorder()

		handler.Update(rr, httpReq)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})

	t.Run("invalid app id", func(t *testing.T) {
		mockService := new(mocks.MockRecordService)
		handler := handlers.NewRecordHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodPut, "/api/v1/apps/invalid/records/1", nil)
		rr := httptest.NewRecorder()

		handler.Update(rr, httpReq)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("invalid json body", func(t *testing.T) {
		mockService := new(mocks.MockRecordService)
		handler := handlers.NewRecordHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodPut, "/api/v1/apps/1/records/1", bytes.NewReader([]byte("invalid json")))
		httpReq.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.Update(rr, httpReq)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("app not found", func(t *testing.T) {
		mockService := new(mocks.MockRecordService)
		handler := handlers.NewRecordHandler(mockService, validator)

		req := models.UpdateRecordRequest{
			Data: models.RecordData{"name": "Test"},
		}

		mockService.On("UpdateRecord", mock.Anything, uint64(999), uint64(1), mock.AnythingOfType("*models.UpdateRecordRequest")).Return(nil, services.ErrAppNotFound)

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPut, "/api/v1/apps/999/records/1", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.Update(rr, httpReq)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		mockService.AssertExpectations(t)
	})
}

func TestRecordHandler_Delete(t *testing.T) {
	validator := utils.NewValidator()

	t.Run("successful delete record", func(t *testing.T) {
		mockService := new(mocks.MockRecordService)
		handler := handlers.NewRecordHandler(mockService, validator)

		mockService.On("DeleteRecord", mock.Anything, uint64(1), uint64(1)).Return(nil)

		httpReq := httptest.NewRequest(http.MethodDelete, "/api/v1/apps/1/records/1", nil)
		rr := httptest.NewRecorder()

		handler.Delete(rr, httpReq)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result models.SuccessResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Contains(t, result.Message, "削除しました")

		mockService.AssertExpectations(t)
	})

	t.Run("method not allowed", func(t *testing.T) {
		mockService := new(mocks.MockRecordService)
		handler := handlers.NewRecordHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/apps/1/records/1", nil)
		rr := httptest.NewRecorder()

		handler.Delete(rr, httpReq)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})

	t.Run("invalid app id", func(t *testing.T) {
		mockService := new(mocks.MockRecordService)
		handler := handlers.NewRecordHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodDelete, "/api/v1/apps/invalid/records/1", nil)
		rr := httptest.NewRecorder()

		handler.Delete(rr, httpReq)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("app not found", func(t *testing.T) {
		mockService := new(mocks.MockRecordService)
		handler := handlers.NewRecordHandler(mockService, validator)

		mockService.On("DeleteRecord", mock.Anything, uint64(999), uint64(1)).Return(services.ErrAppNotFound)

		httpReq := httptest.NewRequest(http.MethodDelete, "/api/v1/apps/999/records/1", nil)
		rr := httptest.NewRecorder()

		handler.Delete(rr, httpReq)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("service error", func(t *testing.T) {
		mockService := new(mocks.MockRecordService)
		handler := handlers.NewRecordHandler(mockService, validator)

		mockService.On("DeleteRecord", mock.Anything, uint64(1), uint64(999)).Return(services.ErrRecordNotFound)

		httpReq := httptest.NewRequest(http.MethodDelete, "/api/v1/apps/1/records/999", nil)
		rr := httptest.NewRecorder()

		handler.Delete(rr, httpReq)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		mockService.AssertExpectations(t)
	})
}

func TestRecordHandler_BulkCreate(t *testing.T) {
	validator := utils.NewValidator()

	t.Run("successful bulk create", func(t *testing.T) {
		mockService := new(mocks.MockRecordService)
		handler := handlers.NewRecordHandler(mockService, validator)

		req := models.BulkCreateRecordRequest{
			Records: []models.RecordData{
				{"name": "R1"},
				{"name": "R2"},
			},
		}
		resp := []models.RecordResponse{
			{ID: 1, Data: models.RecordData{"name": "R1"}},
			{ID: 2, Data: models.RecordData{"name": "R2"}},
		}

		mockService.On("BulkCreateRecords", mock.Anything, uint64(1), uint64(1), mock.AnythingOfType("*models.BulkCreateRecordRequest")).Return(resp, nil)

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/apps/1/records/bulk", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq = httpReq.WithContext(recordContextWithClaims(httpReq.Context(), 1))
		rr := httptest.NewRecorder()

		handler.BulkCreate(rr, httpReq)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var result map[string][]models.RecordResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Len(t, result["records"], 2)

		mockService.AssertExpectations(t)
	})

	t.Run("unauthorized", func(t *testing.T) {
		mockService := new(mocks.MockRecordService)
		handler := handlers.NewRecordHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/apps/1/records/bulk", nil)
		rr := httptest.NewRecorder()

		handler.BulkCreate(rr, httpReq)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("method not allowed", func(t *testing.T) {
		mockService := new(mocks.MockRecordService)
		handler := handlers.NewRecordHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/apps/1/records/bulk", nil)
		rr := httptest.NewRecorder()

		handler.BulkCreate(rr, httpReq)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}

func TestRecordHandler_BulkDelete(t *testing.T) {
	validator := utils.NewValidator()

	t.Run("successful bulk delete", func(t *testing.T) {
		mockService := new(mocks.MockRecordService)
		handler := handlers.NewRecordHandler(mockService, validator)

		req := models.BulkDeleteRecordRequest{
			IDs: []uint64{1, 2, 3},
		}

		mockService.On("BulkDeleteRecords", mock.Anything, uint64(1), mock.AnythingOfType("*models.BulkDeleteRecordRequest")).Return(nil)

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodDelete, "/api/v1/apps/1/records/bulk", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.BulkDelete(rr, httpReq)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result models.SuccessResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Contains(t, result.Message, "削除しました")

		mockService.AssertExpectations(t)
	})

	t.Run("method not allowed", func(t *testing.T) {
		mockService := new(mocks.MockRecordService)
		handler := handlers.NewRecordHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/apps/1/records/bulk", nil)
		rr := httptest.NewRecorder()

		handler.BulkDelete(rr, httpReq)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})

	t.Run("invalid app id", func(t *testing.T) {
		mockService := new(mocks.MockRecordService)
		handler := handlers.NewRecordHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodDelete, "/api/v1/apps/invalid/records/bulk", nil)
		rr := httptest.NewRecorder()

		handler.BulkDelete(rr, httpReq)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("invalid json body", func(t *testing.T) {
		mockService := new(mocks.MockRecordService)
		handler := handlers.NewRecordHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodDelete, "/api/v1/apps/1/records/bulk", bytes.NewReader([]byte("invalid json")))
		httpReq.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.BulkDelete(rr, httpReq)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("service error", func(t *testing.T) {
		mockService := new(mocks.MockRecordService)
		handler := handlers.NewRecordHandler(mockService, validator)

		req := models.BulkDeleteRecordRequest{
			IDs: []uint64{1, 2, 3},
		}

		mockService.On("BulkDeleteRecords", mock.Anything, uint64(1), mock.AnythingOfType("*models.BulkDeleteRecordRequest")).Return(services.ErrAppNotFound)

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodDelete, "/api/v1/apps/1/records/bulk", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.BulkDelete(rr, httpReq)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		mockService.AssertExpectations(t)
	})
}

func TestRecordHandler_Create_AdditionalCases(t *testing.T) {
	validator := utils.NewValidator()

	t.Run("invalid app id", func(t *testing.T) {
		mockService := new(mocks.MockRecordService)
		handler := handlers.NewRecordHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/apps/invalid/records", nil)
		httpReq = httpReq.WithContext(recordContextWithClaims(httpReq.Context(), 1))
		rr := httptest.NewRecorder()

		handler.Create(rr, httpReq)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("invalid json body", func(t *testing.T) {
		mockService := new(mocks.MockRecordService)
		handler := handlers.NewRecordHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/apps/1/records", bytes.NewReader([]byte("invalid json")))
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq = httpReq.WithContext(recordContextWithClaims(httpReq.Context(), 1))
		rr := httptest.NewRecorder()

		handler.Create(rr, httpReq)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("app not found", func(t *testing.T) {
		mockService := new(mocks.MockRecordService)
		handler := handlers.NewRecordHandler(mockService, validator)

		req := models.CreateRecordRequest{
			Data: models.RecordData{"name": "Test"},
		}

		mockService.On("CreateRecord", mock.Anything, uint64(999), uint64(1), mock.AnythingOfType("*models.CreateRecordRequest")).Return(nil, services.ErrAppNotFound)

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/apps/999/records", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq = httpReq.WithContext(recordContextWithClaims(httpReq.Context(), 1))
		rr := httptest.NewRecorder()

		handler.Create(rr, httpReq)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		mockService.AssertExpectations(t)
	})
}

func TestRecordHandler_BulkCreate_AdditionalCases(t *testing.T) {
	validator := utils.NewValidator()

	t.Run("invalid app id", func(t *testing.T) {
		mockService := new(mocks.MockRecordService)
		handler := handlers.NewRecordHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/apps/invalid/records/bulk", nil)
		httpReq = httpReq.WithContext(recordContextWithClaims(httpReq.Context(), 1))
		rr := httptest.NewRecorder()

		handler.BulkCreate(rr, httpReq)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("invalid json body", func(t *testing.T) {
		mockService := new(mocks.MockRecordService)
		handler := handlers.NewRecordHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/apps/1/records/bulk", bytes.NewReader([]byte("invalid json")))
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq = httpReq.WithContext(recordContextWithClaims(httpReq.Context(), 1))
		rr := httptest.NewRecorder()

		handler.BulkCreate(rr, httpReq)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("service error", func(t *testing.T) {
		mockService := new(mocks.MockRecordService)
		handler := handlers.NewRecordHandler(mockService, validator)

		req := models.BulkCreateRecordRequest{
			Records: []models.RecordData{
				{"name": "R1"},
			},
		}

		mockService.On("BulkCreateRecords", mock.Anything, uint64(999), uint64(1), mock.AnythingOfType("*models.BulkCreateRecordRequest")).Return(nil, services.ErrAppNotFound)

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/apps/999/records/bulk", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq = httpReq.WithContext(recordContextWithClaims(httpReq.Context(), 1))
		rr := httptest.NewRecorder()

		handler.BulkCreate(rr, httpReq)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		mockService.AssertExpectations(t)
	})
}

func TestRecordHandler_List_WithFilters(t *testing.T) {
	validator := utils.NewValidator()

	t.Run("with filter query params", func(t *testing.T) {
		mockService := new(mocks.MockRecordService)
		handler := handlers.NewRecordHandler(mockService, validator)

		resp := &models.RecordListResponse{
			Records:    []models.RecordResponse{},
			Pagination: &models.Pagination{Total: 0, Page: 1, Limit: 20},
		}

		mockService.On("GetRecords", mock.Anything, uint64(1), mock.AnythingOfType("repositories.RecordQueryOptions")).Return(resp, nil)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/apps/1/records?filter[name][eq]=test&page=2&limit=10", nil)
		rr := httptest.NewRecorder()

		handler.List(rr, httpReq)

		assert.Equal(t, http.StatusOK, rr.Code)
		mockService.AssertExpectations(t)
	})
}
