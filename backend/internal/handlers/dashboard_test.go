package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"nocode-app/backend/internal/handlers"
	"nocode-app/backend/internal/middleware"
	"nocode-app/backend/internal/models"
	"nocode-app/backend/internal/testhelpers/mocks"
	"nocode-app/backend/internal/utils"
)

func TestDashboardHandler_GetStats(t *testing.T) {
	t.Run("successful get stats", func(t *testing.T) {
		mockService := new(mocks.MockDashboardService)

		stats := &models.DashboardStats{
			AppCount:      5,
			TotalRecords:  100,
			UserCount:     10,
			TodaysUpdates: 15,
		}
		mockService.On("GetStats", mock.Anything).Return(stats, nil)

		handler := handlers.NewDashboardHandler(mockService)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/stats", nil)
		claims := &utils.JWTClaims{
			UserID: 1,
			Role:   "admin",
		}
		ctx := middleware.SetUserInContext(req.Context(), claims)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.GetStats(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]models.DashboardStats
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, int64(5), response["stats"].AppCount)
		assert.Equal(t, int64(100), response["stats"].TotalRecords)
		assert.Equal(t, int64(10), response["stats"].UserCount)
		assert.Equal(t, int64(15), response["stats"].TodaysUpdates)

		mockService.AssertExpectations(t)
	})

	t.Run("service error", func(t *testing.T) {
		mockService := new(mocks.MockDashboardService)

		mockService.On("GetStats", mock.Anything).Return(nil, errors.New("db error"))

		handler := handlers.NewDashboardHandler(mockService)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/stats", nil)
		claims := &utils.JWTClaims{
			UserID: 1,
			Role:   "admin",
		}
		ctx := middleware.SetUserInContext(req.Context(), claims)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.GetStats(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("method not allowed", func(t *testing.T) {
		mockService := new(mocks.MockDashboardService)
		handler := handlers.NewDashboardHandler(mockService)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/dashboard/stats", nil)
		claims := &utils.JWTClaims{
			UserID: 1,
			Role:   "admin",
		}
		ctx := middleware.SetUserInContext(req.Context(), claims)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.GetStats(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}
