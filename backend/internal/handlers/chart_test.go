package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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

// chartContextWithClaims テストのためにJWTクレームを持つコンテキストを作成する関数
func chartContextWithClaims(ctx context.Context, userID uint64) context.Context {
	claims := &utils.JWTClaims{UserID: userID}
	return context.WithValue(ctx, middleware.UserContextKey, claims)
}

func TestChartHandler_GetData(t *testing.T) {
	validator := utils.NewValidator()

	t.Run("successful get chart data", func(t *testing.T) {
		mockService := new(mocks.MockChartService)
		handler := handlers.NewChartHandler(mockService, validator)

		req := models.ChartDataRequest{
			ChartType: "bar",
			XAxis:     models.ChartAxis{Field: "category"},
			YAxis:     models.ChartAxis{Field: "amount", Aggregation: "sum"},
		}
		resp := &models.ChartDataResponse{
			Labels: []string{"A", "B", "C"},
			Datasets: []models.ChartDataset{
				{Label: "Amount", Data: []float64{100, 200, 300}},
			},
		}

		mockService.On("GetChartData", mock.Anything, uint64(1), mock.AnythingOfType("*models.ChartDataRequest")).Return(resp, nil)

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/apps/1/charts/data", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.GetData(rr, httpReq)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result models.ChartDataResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Len(t, result.Labels, 3)

		mockService.AssertExpectations(t)
	})

	t.Run("missing x_axis field", func(t *testing.T) {
		mockService := new(mocks.MockChartService)
		handler := handlers.NewChartHandler(mockService, validator)

		req := models.ChartDataRequest{
			ChartType: "bar",
			XAxis:     models.ChartAxis{Field: ""},
			YAxis:     models.ChartAxis{Field: "amount", Aggregation: "sum"},
		}

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/apps/1/charts/data", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.GetData(rr, httpReq)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("app not found", func(t *testing.T) {
		mockService := new(mocks.MockChartService)
		handler := handlers.NewChartHandler(mockService, validator)

		req := models.ChartDataRequest{
			ChartType: "bar",
			XAxis:     models.ChartAxis{Field: "category"},
			YAxis:     models.ChartAxis{Aggregation: "count"},
		}

		mockService.On("GetChartData", mock.Anything, uint64(999), mock.AnythingOfType("*models.ChartDataRequest")).Return(nil, services.ErrAppNotFound)

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/apps/999/charts/data", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.GetData(rr, httpReq)

		assert.Equal(t, http.StatusNotFound, rr.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("invalid app id", func(t *testing.T) {
		mockService := new(mocks.MockChartService)
		handler := handlers.NewChartHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/apps/invalid/charts/data", nil)
		rr := httptest.NewRecorder()

		handler.GetData(rr, httpReq)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("method not allowed", func(t *testing.T) {
		mockService := new(mocks.MockChartService)
		handler := handlers.NewChartHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/apps/1/charts/data", nil)
		rr := httptest.NewRecorder()

		handler.GetData(rr, httpReq)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}

func TestChartHandler_GetConfigs(t *testing.T) {
	validator := utils.NewValidator()

	t.Run("successful get chart configs", func(t *testing.T) {
		mockService := new(mocks.MockChartService)
		handler := handlers.NewChartHandler(mockService, validator)

		configs := []models.ChartConfig{
			{ID: 1, Name: "Config 1", AppID: 1},
			{ID: 2, Name: "Config 2", AppID: 1},
		}

		mockService.On("GetChartConfigs", mock.Anything, uint64(1)).Return(configs, nil)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/apps/1/charts/config", nil)
		rr := httptest.NewRecorder()

		handler.GetConfigs(rr, httpReq)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result map[string][]models.ChartConfig
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Len(t, result["configs"], 2)

		mockService.AssertExpectations(t)
	})

	t.Run("invalid app id", func(t *testing.T) {
		mockService := new(mocks.MockChartService)
		handler := handlers.NewChartHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/apps/invalid/charts/config", nil)
		rr := httptest.NewRecorder()

		handler.GetConfigs(rr, httpReq)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("method not allowed", func(t *testing.T) {
		mockService := new(mocks.MockChartService)
		handler := handlers.NewChartHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/apps/1/charts/config", nil)
		rr := httptest.NewRecorder()

		handler.GetConfigs(rr, httpReq)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}

func TestChartHandler_SaveConfig(t *testing.T) {
	validator := utils.NewValidator()

	t.Run("successful save chart config", func(t *testing.T) {
		mockService := new(mocks.MockChartService)
		handler := handlers.NewChartHandler(mockService, validator)

		req := models.SaveChartConfigRequest{
			Name:      "New Config",
			ChartType: "bar",
			Config: models.ChartDataRequest{
				ChartType: "bar",
				XAxis:     models.ChartAxis{Field: "category"},
				YAxis:     models.ChartAxis{Aggregation: "count"},
			},
		}
		resp := &models.ChartConfig{
			ID:    1,
			Name:  "New Config",
			AppID: 1,
		}

		mockService.On("SaveChartConfig", mock.Anything, uint64(1), uint64(1), mock.AnythingOfType("*models.SaveChartConfigRequest")).Return(resp, nil)

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/apps/1/charts/config", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq = httpReq.WithContext(chartContextWithClaims(httpReq.Context(), 1))
		rr := httptest.NewRecorder()

		handler.SaveConfig(rr, httpReq)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var result models.ChartConfig
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Equal(t, "New Config", result.Name)

		mockService.AssertExpectations(t)
	})

	t.Run("unauthorized", func(t *testing.T) {
		mockService := new(mocks.MockChartService)
		handler := handlers.NewChartHandler(mockService, validator)

		req := models.SaveChartConfigRequest{Name: "Config"}
		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/apps/1/charts/config", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.SaveConfig(rr, httpReq)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("method not allowed", func(t *testing.T) {
		mockService := new(mocks.MockChartService)
		handler := handlers.NewChartHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/apps/1/charts/config", nil)
		rr := httptest.NewRecorder()

		handler.SaveConfig(rr, httpReq)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})

	t.Run("invalid app id", func(t *testing.T) {
		mockService := new(mocks.MockChartService)
		handler := handlers.NewChartHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/apps/invalid/charts/config", nil)
		httpReq = httpReq.WithContext(chartContextWithClaims(httpReq.Context(), 1))
		rr := httptest.NewRecorder()

		handler.SaveConfig(rr, httpReq)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("invalid json body", func(t *testing.T) {
		mockService := new(mocks.MockChartService)
		handler := handlers.NewChartHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/apps/1/charts/config", bytes.NewReader([]byte("invalid json")))
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq = httpReq.WithContext(chartContextWithClaims(httpReq.Context(), 1))
		rr := httptest.NewRecorder()

		handler.SaveConfig(rr, httpReq)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("app not found", func(t *testing.T) {
		mockService := new(mocks.MockChartService)
		handler := handlers.NewChartHandler(mockService, validator)

		req := models.SaveChartConfigRequest{
			Name:      "Config",
			ChartType: "bar",
			Config: models.ChartDataRequest{
				ChartType: "bar",
				XAxis:     models.ChartAxis{Field: "category"},
				YAxis:     models.ChartAxis{Aggregation: "count"},
			},
		}

		mockService.On("SaveChartConfig", mock.Anything, uint64(999), uint64(1), mock.AnythingOfType("*models.SaveChartConfigRequest")).Return(nil, services.ErrAppNotFound)

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/apps/999/charts/config", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq = httpReq.WithContext(chartContextWithClaims(httpReq.Context(), 1))
		rr := httptest.NewRecorder()

		handler.SaveConfig(rr, httpReq)

		assert.Equal(t, http.StatusNotFound, rr.Code)

		mockService.AssertExpectations(t)
	})
}

func TestChartHandler_GetData_AdditionalCases(t *testing.T) {
	validator := utils.NewValidator()

	t.Run("invalid json body", func(t *testing.T) {
		mockService := new(mocks.MockChartService)
		handler := handlers.NewChartHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/apps/1/charts/data", bytes.NewReader([]byte("invalid json")))
		httpReq.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.GetData(rr, httpReq)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestChartHandler_GetConfigs_ServiceError(t *testing.T) {
	validator := utils.NewValidator()

	t.Run("service error", func(t *testing.T) {
		mockService := new(mocks.MockChartService)
		handler := handlers.NewChartHandler(mockService, validator)

		mockService.On("GetChartConfigs", mock.Anything, uint64(999)).Return(nil, services.ErrAppNotFound)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/apps/999/charts/config", nil)
		rr := httptest.NewRecorder()

		handler.GetConfigs(rr, httpReq)

		// Handler doesn't specifically check for ErrAppNotFound in GetConfigs
		assert.Equal(t, http.StatusInternalServerError, rr.Code)

		mockService.AssertExpectations(t)
	})
}

func TestChartHandler_DeleteConfig(t *testing.T) {
	validator := utils.NewValidator()

	t.Run("successful delete chart config", func(t *testing.T) {
		mockService := new(mocks.MockChartService)
		handler := handlers.NewChartHandler(mockService, validator)

		mockService.On("DeleteChartConfig", mock.Anything, uint64(1)).Return(nil)

		httpReq := httptest.NewRequest(http.MethodDelete, "/api/v1/apps/1/charts/config/1", nil)
		rr := httptest.NewRecorder()

		handler.DeleteConfig(rr, httpReq)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result models.SuccessResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Contains(t, result.Message, "チャート設定を削除しました")

		mockService.AssertExpectations(t)
	})

	t.Run("config not found", func(t *testing.T) {
		mockService := new(mocks.MockChartService)
		handler := handlers.NewChartHandler(mockService, validator)

		mockService.On("DeleteChartConfig", mock.Anything, uint64(999)).Return(services.ErrChartConfigNotFound)

		httpReq := httptest.NewRequest(http.MethodDelete, "/api/v1/apps/1/charts/config/999", nil)
		rr := httptest.NewRecorder()

		handler.DeleteConfig(rr, httpReq)

		assert.Equal(t, http.StatusNotFound, rr.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("invalid config id", func(t *testing.T) {
		mockService := new(mocks.MockChartService)
		handler := handlers.NewChartHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodDelete, "/api/v1/apps/1/charts/config/invalid", nil)
		rr := httptest.NewRecorder()

		handler.DeleteConfig(rr, httpReq)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("method not allowed", func(t *testing.T) {
		mockService := new(mocks.MockChartService)
		handler := handlers.NewChartHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/apps/1/charts/config/1", nil)
		rr := httptest.NewRecorder()

		handler.DeleteConfig(rr, httpReq)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}
