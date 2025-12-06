package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"nocode-app/backend/internal/handlers"
	"nocode-app/backend/internal/models"
	"nocode-app/backend/internal/services"
	"nocode-app/backend/internal/testhelpers/mocks"
	"nocode-app/backend/internal/utils"
)

func TestFieldHandler_List(t *testing.T) {
	validator := utils.NewValidator()

	t.Run("successful list fields", func(t *testing.T) {
		mockService := new(mocks.MockFieldService)
		handler := handlers.NewFieldHandler(mockService, validator)

		fields := []models.FieldResponse{
			{ID: 1, FieldCode: "name", FieldName: "Name", FieldType: "TEXT"},
			{ID: 2, FieldCode: "age", FieldName: "Age", FieldType: "NUMBER"},
		}

		mockService.On("GetFields", mock.Anything, uint64(1)).Return(fields, nil)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/apps/1/fields", nil)
		rr := httptest.NewRecorder()

		handler.List(rr, httpReq)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result map[string][]models.FieldResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Len(t, result["fields"], 2)

		mockService.AssertExpectations(t)
	})

	t.Run("invalid app id", func(t *testing.T) {
		mockService := new(mocks.MockFieldService)
		handler := handlers.NewFieldHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/apps/invalid/fields", nil)
		rr := httptest.NewRecorder()

		handler.List(rr, httpReq)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("method not allowed", func(t *testing.T) {
		mockService := new(mocks.MockFieldService)
		handler := handlers.NewFieldHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/apps/1/fields", nil)
		rr := httptest.NewRecorder()

		handler.List(rr, httpReq)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}

func TestFieldHandler_Create(t *testing.T) {
	validator := utils.NewValidator()

	t.Run("successful create field", func(t *testing.T) {
		mockService := new(mocks.MockFieldService)
		handler := handlers.NewFieldHandler(mockService, validator)

		req := models.CreateFieldRequest{
			FieldCode: "email",
			FieldName: "Email",
			FieldType: "text",
		}
		resp := &models.FieldResponse{
			ID:        1,
			FieldCode: "email",
			FieldName: "Email",
			FieldType: "text",
		}

		mockService.On("CreateField", mock.Anything, uint64(1), mock.AnythingOfType("*models.CreateFieldRequest")).Return(resp, nil)

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/apps/1/fields", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.Create(rr, httpReq)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var result models.FieldResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Equal(t, "email", result.FieldCode)

		mockService.AssertExpectations(t)
	})

	t.Run("field code exists", func(t *testing.T) {
		mockService := new(mocks.MockFieldService)
		handler := handlers.NewFieldHandler(mockService, validator)

		req := models.CreateFieldRequest{
			FieldCode: "existing",
			FieldName: "Existing",
			FieldType: "text",
		}

		mockService.On("CreateField", mock.Anything, uint64(1), mock.AnythingOfType("*models.CreateFieldRequest")).Return(nil, services.ErrFieldCodeExists)

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/apps/1/fields", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.Create(rr, httpReq)

		assert.Equal(t, http.StatusConflict, rr.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("app not found", func(t *testing.T) {
		mockService := new(mocks.MockFieldService)
		handler := handlers.NewFieldHandler(mockService, validator)

		req := models.CreateFieldRequest{
			FieldCode: "name",
			FieldName: "Name",
			FieldType: "text",
		}

		mockService.On("CreateField", mock.Anything, uint64(999), mock.AnythingOfType("*models.CreateFieldRequest")).Return(nil, services.ErrAppNotFound)

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/apps/999/fields", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.Create(rr, httpReq)

		assert.Equal(t, http.StatusNotFound, rr.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("method not allowed", func(t *testing.T) {
		mockService := new(mocks.MockFieldService)
		handler := handlers.NewFieldHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/apps/1/fields", nil)
		rr := httptest.NewRecorder()

		handler.Create(rr, httpReq)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}

func TestFieldHandler_Update(t *testing.T) {
	validator := utils.NewValidator()

	t.Run("successful update field", func(t *testing.T) {
		mockService := new(mocks.MockFieldService)
		handler := handlers.NewFieldHandler(mockService, validator)

		req := models.UpdateFieldRequest{
			FieldName: "Updated Name",
		}
		resp := &models.FieldResponse{
			ID:        1,
			FieldCode: "name",
			FieldName: "Updated Name",
			FieldType: "TEXT",
		}

		mockService.On("UpdateField", mock.Anything, uint64(1), mock.AnythingOfType("*models.UpdateFieldRequest")).Return(resp, nil)

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPut, "/api/v1/apps/1/fields/1", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.Update(rr, httpReq)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result models.FieldResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", result.FieldName)

		mockService.AssertExpectations(t)
	})

	t.Run("field not found", func(t *testing.T) {
		mockService := new(mocks.MockFieldService)
		handler := handlers.NewFieldHandler(mockService, validator)

		req := models.UpdateFieldRequest{FieldName: "Updated"}
		mockService.On("UpdateField", mock.Anything, uint64(999), mock.AnythingOfType("*models.UpdateFieldRequest")).Return(nil, services.ErrFieldNotFound)

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPut, "/api/v1/apps/1/fields/999", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.Update(rr, httpReq)

		assert.Equal(t, http.StatusNotFound, rr.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("method not allowed", func(t *testing.T) {
		mockService := new(mocks.MockFieldService)
		handler := handlers.NewFieldHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/apps/1/fields/1", nil)
		rr := httptest.NewRecorder()

		handler.Update(rr, httpReq)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}

func TestFieldHandler_Delete(t *testing.T) {
	validator := utils.NewValidator()

	t.Run("successful delete field", func(t *testing.T) {
		mockService := new(mocks.MockFieldService)
		handler := handlers.NewFieldHandler(mockService, validator)

		mockService.On("DeleteField", mock.Anything, uint64(1), uint64(1)).Return(nil)

		httpReq := httptest.NewRequest(http.MethodDelete, "/api/v1/apps/1/fields/1", nil)
		rr := httptest.NewRecorder()

		handler.Delete(rr, httpReq)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result models.SuccessResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Contains(t, result.Message, "フィールドを削除しました")

		mockService.AssertExpectations(t)
	})

	t.Run("field not found", func(t *testing.T) {
		mockService := new(mocks.MockFieldService)
		handler := handlers.NewFieldHandler(mockService, validator)

		mockService.On("DeleteField", mock.Anything, uint64(1), uint64(999)).Return(services.ErrFieldNotFound)

		httpReq := httptest.NewRequest(http.MethodDelete, "/api/v1/apps/1/fields/999", nil)
		rr := httptest.NewRecorder()

		handler.Delete(rr, httpReq)

		assert.Equal(t, http.StatusNotFound, rr.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("method not allowed", func(t *testing.T) {
		mockService := new(mocks.MockFieldService)
		handler := handlers.NewFieldHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/apps/1/fields/1", nil)
		rr := httptest.NewRecorder()

		handler.Delete(rr, httpReq)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})

	t.Run("invalid app id", func(t *testing.T) {
		mockService := new(mocks.MockFieldService)
		handler := handlers.NewFieldHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodDelete, "/api/v1/apps/invalid/fields/1", nil)
		rr := httptest.NewRecorder()

		handler.Delete(rr, httpReq)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("invalid field id", func(t *testing.T) {
		mockService := new(mocks.MockFieldService)
		handler := handlers.NewFieldHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodDelete, "/api/v1/apps/1/fields/invalid", nil)
		rr := httptest.NewRecorder()

		handler.Delete(rr, httpReq)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("app not found", func(t *testing.T) {
		mockService := new(mocks.MockFieldService)
		handler := handlers.NewFieldHandler(mockService, validator)

		mockService.On("DeleteField", mock.Anything, uint64(999), uint64(1)).Return(services.ErrAppNotFound)

		httpReq := httptest.NewRequest(http.MethodDelete, "/api/v1/apps/999/fields/1", nil)
		rr := httptest.NewRecorder()

		handler.Delete(rr, httpReq)

		assert.Equal(t, http.StatusNotFound, rr.Code)

		mockService.AssertExpectations(t)
	})
}

func TestFieldHandler_UpdateOrder(t *testing.T) {
	validator := utils.NewValidator()

	t.Run("successful update order", func(t *testing.T) {
		mockService := new(mocks.MockFieldService)
		handler := handlers.NewFieldHandler(mockService, validator)

		req := models.UpdateFieldOrderRequest{
			Fields: []models.FieldOrderItem{
				{ID: 1, DisplayOrder: 1},
				{ID: 2, DisplayOrder: 2},
			},
		}

		mockService.On("UpdateFieldOrder", mock.Anything, uint64(1), mock.AnythingOfType("*models.UpdateFieldOrderRequest")).Return(nil)

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPut, "/api/v1/apps/1/fields/order", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.UpdateOrder(rr, httpReq)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result models.SuccessResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Contains(t, result.Message, "フィールド順序を更新しました")

		mockService.AssertExpectations(t)
	})

	t.Run("method not allowed", func(t *testing.T) {
		mockService := new(mocks.MockFieldService)
		handler := handlers.NewFieldHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/apps/1/fields/order", nil)
		rr := httptest.NewRecorder()

		handler.UpdateOrder(rr, httpReq)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})

	t.Run("invalid app id", func(t *testing.T) {
		mockService := new(mocks.MockFieldService)
		handler := handlers.NewFieldHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodPut, "/api/v1/apps/invalid/fields/order", nil)
		rr := httptest.NewRecorder()

		handler.UpdateOrder(rr, httpReq)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("invalid json body", func(t *testing.T) {
		mockService := new(mocks.MockFieldService)
		handler := handlers.NewFieldHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodPut, "/api/v1/apps/1/fields/order", bytes.NewReader([]byte("invalid json")))
		httpReq.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.UpdateOrder(rr, httpReq)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("service error", func(t *testing.T) {
		mockService := new(mocks.MockFieldService)
		handler := handlers.NewFieldHandler(mockService, validator)

		req := models.UpdateFieldOrderRequest{
			Fields: []models.FieldOrderItem{
				{ID: 1, DisplayOrder: 1},
			},
		}

		mockService.On("UpdateFieldOrder", mock.Anything, uint64(1), mock.AnythingOfType("*models.UpdateFieldOrderRequest")).Return(services.ErrAppNotFound)

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPut, "/api/v1/apps/1/fields/order", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.UpdateOrder(rr, httpReq)

		// Handler doesn't specifically check for ErrAppNotFound in UpdateOrder
		assert.Equal(t, http.StatusInternalServerError, rr.Code)

		mockService.AssertExpectations(t)
	})
}

func TestFieldHandler_Create_AdditionalCases(t *testing.T) {
	validator := utils.NewValidator()

	t.Run("invalid app id", func(t *testing.T) {
		mockService := new(mocks.MockFieldService)
		handler := handlers.NewFieldHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/apps/invalid/fields", nil)
		httpReq.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.Create(rr, httpReq)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("invalid json body", func(t *testing.T) {
		mockService := new(mocks.MockFieldService)
		handler := handlers.NewFieldHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/apps/1/fields", bytes.NewReader([]byte("invalid json")))
		httpReq.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.Create(rr, httpReq)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestFieldHandler_Update_AdditionalCases(t *testing.T) {
	validator := utils.NewValidator()

	t.Run("invalid app id", func(t *testing.T) {
		mockService := new(mocks.MockFieldService)
		handler := handlers.NewFieldHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodPut, "/api/v1/apps/invalid/fields/1", nil)
		httpReq.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.Update(rr, httpReq)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("invalid json body", func(t *testing.T) {
		mockService := new(mocks.MockFieldService)
		handler := handlers.NewFieldHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodPut, "/api/v1/apps/1/fields/1", bytes.NewReader([]byte("invalid json")))
		httpReq.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.Update(rr, httpReq)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestFieldHandler_List_ServiceError(t *testing.T) {
	validator := utils.NewValidator()

	t.Run("service error", func(t *testing.T) {
		mockService := new(mocks.MockFieldService)
		handler := handlers.NewFieldHandler(mockService, validator)

		mockService.On("GetFields", mock.Anything, uint64(999)).Return(nil, services.ErrAppNotFound)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/apps/999/fields", nil)
		rr := httptest.NewRecorder()

		handler.List(rr, httpReq)

		// Handler doesn't specifically check for ErrAppNotFound in List
		assert.Equal(t, http.StatusInternalServerError, rr.Code)

		mockService.AssertExpectations(t)
	})
}
