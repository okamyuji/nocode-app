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

// createContextWithClaims テスト用にJWTクレームをもつコンテキストを作成する関数
func createContextWithClaims(ctx context.Context, userID uint64) context.Context {
	claims := &utils.JWTClaims{UserID: userID}
	return context.WithValue(ctx, middleware.UserContextKey, claims)
}

func TestAppHandler_List(t *testing.T) {
	validator := utils.NewValidator()

	t.Run("successful list apps", func(t *testing.T) {
		mockService := new(mocks.MockAppService)
		handler := handlers.NewAppHandler(mockService, validator)

		resp := &models.AppListResponse{
			Apps: []models.AppResponse{
				{ID: 1, Name: "App 1", Description: "Description 1"},
				{ID: 2, Name: "App 2", Description: "Description 2"},
			},
			Pagination: &models.Pagination{
				Total: 2,
				Page:  1,
				Limit: 20,
			},
		}

		mockService.On("GetApps", mock.Anything, 1, 20).Return(resp, nil)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/apps", nil)
		rr := httptest.NewRecorder()

		handler.List(rr, httpReq)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result models.AppListResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Len(t, result.Apps, 2)
		assert.Equal(t, int64(2), result.Pagination.Total)

		mockService.AssertExpectations(t)
	})

	t.Run("list apps with pagination", func(t *testing.T) {
		mockService := new(mocks.MockAppService)
		handler := handlers.NewAppHandler(mockService, validator)

		resp := &models.AppListResponse{
			Apps: []models.AppResponse{
				{ID: 3, Name: "App 3"},
			},
			Pagination: &models.Pagination{
				Total: 5,
				Page:  2,
				Limit: 2,
			},
		}

		mockService.On("GetApps", mock.Anything, 2, 2).Return(resp, nil)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/apps?page=2&limit=2", nil)
		rr := httptest.NewRecorder()

		handler.List(rr, httpReq)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result models.AppListResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Len(t, result.Apps, 1)

		mockService.AssertExpectations(t)
	})

	t.Run("method not allowed", func(t *testing.T) {
		mockService := new(mocks.MockAppService)
		handler := handlers.NewAppHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/apps", nil)
		rr := httptest.NewRecorder()

		handler.List(rr, httpReq)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}

func TestAppHandler_Create(t *testing.T) {
	validator := utils.NewValidator()

	t.Run("successful create app", func(t *testing.T) {
		mockService := new(mocks.MockAppService)
		handler := handlers.NewAppHandler(mockService, validator)

		req := models.CreateAppRequest{
			Name:        "New App",
			Description: "New Description",
			Fields: []models.CreateFieldRequest{
				{FieldCode: "name", FieldName: "Name", FieldType: "text"},
			},
		}
		now := time.Now()
		resp := &models.AppResponse{
			ID:          1,
			Name:        "New App",
			Description: "New Description",
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		mockService.On("CreateApp", mock.Anything, uint64(1), mock.AnythingOfType("*models.CreateAppRequest")).Return(resp, nil)

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/apps", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq = httpReq.WithContext(createContextWithClaims(httpReq.Context(), 1))
		rr := httptest.NewRecorder()

		handler.Create(rr, httpReq)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var result models.AppResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Equal(t, "New App", result.Name)

		mockService.AssertExpectations(t)
	})

	t.Run("unauthorized - no claims", func(t *testing.T) {
		mockService := new(mocks.MockAppService)
		handler := handlers.NewAppHandler(mockService, validator)

		req := models.CreateAppRequest{Name: "New App"}
		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/apps", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.Create(rr, httpReq)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("invalid request body", func(t *testing.T) {
		mockService := new(mocks.MockAppService)
		handler := handlers.NewAppHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/apps", bytes.NewReader([]byte("invalid json")))
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq = httpReq.WithContext(createContextWithClaims(httpReq.Context(), 1))
		rr := httptest.NewRecorder()

		handler.Create(rr, httpReq)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("method not allowed", func(t *testing.T) {
		mockService := new(mocks.MockAppService)
		handler := handlers.NewAppHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/apps", nil)
		rr := httptest.NewRecorder()

		handler.Create(rr, httpReq)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}

func TestAppHandler_Get(t *testing.T) {
	validator := utils.NewValidator()

	t.Run("successful get app", func(t *testing.T) {
		mockService := new(mocks.MockAppService)
		handler := handlers.NewAppHandler(mockService, validator)

		now := time.Now()
		resp := &models.AppResponse{
			ID:          1,
			Name:        "App 1",
			Description: "Description 1",
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		mockService.On("GetApp", mock.Anything, uint64(1)).Return(resp, nil)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/apps/1", nil)
		rr := httptest.NewRecorder()

		handler.Get(rr, httpReq)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result models.AppResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Equal(t, uint64(1), result.ID)
		assert.Equal(t, "App 1", result.Name)

		mockService.AssertExpectations(t)
	})

	t.Run("app not found", func(t *testing.T) {
		mockService := new(mocks.MockAppService)
		handler := handlers.NewAppHandler(mockService, validator)

		mockService.On("GetApp", mock.Anything, uint64(999)).Return(nil, services.ErrAppNotFound)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/apps/999", nil)
		rr := httptest.NewRecorder()

		handler.Get(rr, httpReq)

		assert.Equal(t, http.StatusNotFound, rr.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("invalid app id", func(t *testing.T) {
		mockService := new(mocks.MockAppService)
		handler := handlers.NewAppHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/apps/invalid", nil)
		rr := httptest.NewRecorder()

		handler.Get(rr, httpReq)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("method not allowed", func(t *testing.T) {
		mockService := new(mocks.MockAppService)
		handler := handlers.NewAppHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/apps/1", nil)
		rr := httptest.NewRecorder()

		handler.Get(rr, httpReq)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}

func TestAppHandler_Update(t *testing.T) {
	validator := utils.NewValidator()

	t.Run("successful update app", func(t *testing.T) {
		mockService := new(mocks.MockAppService)
		handler := handlers.NewAppHandler(mockService, validator)

		req := models.UpdateAppRequest{
			Name:        "Updated App",
			Description: "Updated Description",
		}
		now := time.Now()
		resp := &models.AppResponse{
			ID:          1,
			Name:        "Updated App",
			Description: "Updated Description",
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		mockService.On("UpdateApp", mock.Anything, uint64(1), mock.AnythingOfType("*models.UpdateAppRequest")).Return(resp, nil)

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPut, "/api/v1/apps/1", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.Update(rr, httpReq)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result models.AppResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Equal(t, "Updated App", result.Name)

		mockService.AssertExpectations(t)
	})

	t.Run("app not found", func(t *testing.T) {
		mockService := new(mocks.MockAppService)
		handler := handlers.NewAppHandler(mockService, validator)

		req := models.UpdateAppRequest{Name: "Updated"}
		mockService.On("UpdateApp", mock.Anything, uint64(999), mock.AnythingOfType("*models.UpdateAppRequest")).Return(nil, services.ErrAppNotFound)

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPut, "/api/v1/apps/999", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.Update(rr, httpReq)

		assert.Equal(t, http.StatusNotFound, rr.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("method not allowed", func(t *testing.T) {
		mockService := new(mocks.MockAppService)
		handler := handlers.NewAppHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/apps/1", nil)
		rr := httptest.NewRecorder()

		handler.Update(rr, httpReq)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}

func TestAppHandler_Delete(t *testing.T) {
	validator := utils.NewValidator()

	t.Run("successful delete app", func(t *testing.T) {
		mockService := new(mocks.MockAppService)
		handler := handlers.NewAppHandler(mockService, validator)

		mockService.On("DeleteApp", mock.Anything, uint64(1)).Return(nil)

		httpReq := httptest.NewRequest(http.MethodDelete, "/api/v1/apps/1", nil)
		rr := httptest.NewRecorder()

		handler.Delete(rr, httpReq)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result models.SuccessResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Contains(t, result.Message, "アプリを削除しました")

		mockService.AssertExpectations(t)
	})

	t.Run("app not found", func(t *testing.T) {
		mockService := new(mocks.MockAppService)
		handler := handlers.NewAppHandler(mockService, validator)

		mockService.On("DeleteApp", mock.Anything, uint64(999)).Return(services.ErrAppNotFound)

		httpReq := httptest.NewRequest(http.MethodDelete, "/api/v1/apps/999", nil)
		rr := httptest.NewRecorder()

		handler.Delete(rr, httpReq)

		assert.Equal(t, http.StatusNotFound, rr.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("method not allowed", func(t *testing.T) {
		mockService := new(mocks.MockAppService)
		handler := handlers.NewAppHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/apps/1", nil)
		rr := httptest.NewRecorder()

		handler.Delete(rr, httpReq)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}
