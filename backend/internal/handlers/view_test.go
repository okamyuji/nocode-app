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

func TestViewHandler_List(t *testing.T) {
	validator := utils.NewValidator()

	t.Run("successful list views", func(t *testing.T) {
		mockService := new(mocks.MockViewService)
		handler := handlers.NewViewHandler(mockService, validator)

		views := []models.ViewResponse{
			{ID: 1, Name: "Table View", ViewType: "table"},
			{ID: 2, Name: "Chart View", ViewType: "chart"},
		}

		mockService.On("GetViews", mock.Anything, uint64(1)).Return(views, nil)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/apps/1/views", nil)
		rr := httptest.NewRecorder()

		handler.List(rr, httpReq)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result map[string][]models.ViewResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Len(t, result["views"], 2)

		mockService.AssertExpectations(t)
	})

	t.Run("invalid app id", func(t *testing.T) {
		mockService := new(mocks.MockViewService)
		handler := handlers.NewViewHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/apps/invalid/views", nil)
		rr := httptest.NewRecorder()

		handler.List(rr, httpReq)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("method not allowed", func(t *testing.T) {
		mockService := new(mocks.MockViewService)
		handler := handlers.NewViewHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/apps/1/views", nil)
		rr := httptest.NewRecorder()

		handler.List(rr, httpReq)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}

func TestViewHandler_Create(t *testing.T) {
	validator := utils.NewValidator()

	t.Run("successful create view", func(t *testing.T) {
		mockService := new(mocks.MockViewService)
		handler := handlers.NewViewHandler(mockService, validator)

		req := models.CreateViewRequest{
			Name:     "New View",
			ViewType: "table",
		}
		resp := &models.ViewResponse{
			ID:       1,
			Name:     "New View",
			ViewType: "table",
		}

		mockService.On("CreateView", mock.Anything, uint64(1), mock.AnythingOfType("*models.CreateViewRequest")).Return(resp, nil)

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/apps/1/views", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.Create(rr, httpReq)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var result models.ViewResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Equal(t, "New View", result.Name)

		mockService.AssertExpectations(t)
	})

	t.Run("app not found", func(t *testing.T) {
		mockService := new(mocks.MockViewService)
		handler := handlers.NewViewHandler(mockService, validator)

		req := models.CreateViewRequest{Name: "New", ViewType: "table"}
		mockService.On("CreateView", mock.Anything, uint64(999), mock.AnythingOfType("*models.CreateViewRequest")).Return(nil, services.ErrAppNotFound)

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/apps/999/views", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.Create(rr, httpReq)

		assert.Equal(t, http.StatusNotFound, rr.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("method not allowed", func(t *testing.T) {
		mockService := new(mocks.MockViewService)
		handler := handlers.NewViewHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/apps/1/views", nil)
		rr := httptest.NewRecorder()

		handler.Create(rr, httpReq)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}

func TestViewHandler_Update(t *testing.T) {
	validator := utils.NewValidator()

	t.Run("successful update view", func(t *testing.T) {
		mockService := new(mocks.MockViewService)
		handler := handlers.NewViewHandler(mockService, validator)

		req := models.UpdateViewRequest{
			Name: "Updated View",
		}
		resp := &models.ViewResponse{
			ID:       1,
			Name:     "Updated View",
			ViewType: "table",
		}

		mockService.On("UpdateView", mock.Anything, uint64(1), mock.AnythingOfType("*models.UpdateViewRequest")).Return(resp, nil)

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPut, "/api/v1/apps/1/views/1", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.Update(rr, httpReq)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result models.ViewResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Equal(t, "Updated View", result.Name)

		mockService.AssertExpectations(t)
	})

	t.Run("view not found", func(t *testing.T) {
		mockService := new(mocks.MockViewService)
		handler := handlers.NewViewHandler(mockService, validator)

		req := models.UpdateViewRequest{Name: "Updated"}
		mockService.On("UpdateView", mock.Anything, uint64(999), mock.AnythingOfType("*models.UpdateViewRequest")).Return(nil, services.ErrViewNotFound)

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPut, "/api/v1/apps/1/views/999", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.Update(rr, httpReq)

		assert.Equal(t, http.StatusNotFound, rr.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("method not allowed", func(t *testing.T) {
		mockService := new(mocks.MockViewService)
		handler := handlers.NewViewHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/apps/1/views/1", nil)
		rr := httptest.NewRecorder()

		handler.Update(rr, httpReq)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}

func TestViewHandler_Delete(t *testing.T) {
	validator := utils.NewValidator()

	t.Run("successful delete view", func(t *testing.T) {
		mockService := new(mocks.MockViewService)
		handler := handlers.NewViewHandler(mockService, validator)

		mockService.On("DeleteView", mock.Anything, uint64(1)).Return(nil)

		httpReq := httptest.NewRequest(http.MethodDelete, "/api/v1/apps/1/views/1", nil)
		rr := httptest.NewRecorder()

		handler.Delete(rr, httpReq)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result models.SuccessResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Contains(t, result.Message, "ビューを削除しました")

		mockService.AssertExpectations(t)
	})

	t.Run("view not found", func(t *testing.T) {
		mockService := new(mocks.MockViewService)
		handler := handlers.NewViewHandler(mockService, validator)

		mockService.On("DeleteView", mock.Anything, uint64(999)).Return(services.ErrViewNotFound)

		httpReq := httptest.NewRequest(http.MethodDelete, "/api/v1/apps/1/views/999", nil)
		rr := httptest.NewRecorder()

		handler.Delete(rr, httpReq)

		assert.Equal(t, http.StatusNotFound, rr.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("method not allowed", func(t *testing.T) {
		mockService := new(mocks.MockViewService)
		handler := handlers.NewViewHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/apps/1/views/1", nil)
		rr := httptest.NewRecorder()

		handler.Delete(rr, httpReq)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}
