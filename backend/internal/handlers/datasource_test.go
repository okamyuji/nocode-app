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

func createDataSourceHandler(mockService *mocks.MockDataSourceService) *handlers.DataSourceHandler {
	validator := utils.NewValidator()
	return handlers.NewDataSourceHandler(mockService, validator)
}

func createAuthenticatedRequest(method, path string, body interface{}) *http.Request {
	var reqBody []byte
	if body != nil {
		reqBody, _ = json.Marshal(body)
	}

	req := httptest.NewRequest(method, path, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	claims := &utils.JWTClaims{
		UserID: 1,
		Email:  "test@example.com",
		Role:   "admin",
	}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, claims)
	return req.WithContext(ctx)
}

func TestDataSourceHandler_List(t *testing.T) {
	t.Run("successful list", func(t *testing.T) {
		mockService := new(mocks.MockDataSourceService)
		handler := createDataSourceHandler(mockService)

		resp := &models.DataSourceListResponse{
			DataSources: []models.DataSourceResponse{
				{ID: 1, Name: "ds1", DBType: models.DBTypePostgreSQL},
				{ID: 2, Name: "ds2", DBType: models.DBTypeMySQL},
			},
			Pagination: models.NewPagination(1, 20, 2),
		}

		mockService.On("GetDataSources", mock.Anything, 1, 20).Return(resp, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/datasources", nil)
		w := httptest.NewRecorder()

		handler.List(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var result models.DataSourceListResponse
		err := json.Unmarshal(w.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Len(t, result.DataSources, 2)

		mockService.AssertExpectations(t)
	})

	t.Run("method not allowed", func(t *testing.T) {
		mockService := new(mocks.MockDataSourceService)
		handler := createDataSourceHandler(mockService)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/datasources", nil)
		w := httptest.NewRecorder()

		handler.List(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestDataSourceHandler_Create(t *testing.T) {
	t.Run("successful create", func(t *testing.T) {
		mockService := new(mocks.MockDataSourceService)
		handler := createDataSourceHandler(mockService)

		createReq := &models.CreateDataSourceRequest{
			Name:         "test-ds",
			DBType:       "postgresql",
			Host:         "localhost",
			Port:         5432,
			DatabaseName: "testdb",
			Username:     "testuser",
			Password:     "testpass",
		}

		resp := &models.DataSourceResponse{
			ID:           1,
			Name:         "test-ds",
			DBType:       models.DBTypePostgreSQL,
			Host:         "localhost",
			Port:         5432,
			DatabaseName: "testdb",
			Username:     "testuser",
			CreatedBy:    1,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		mockService.On("CreateDataSource", mock.Anything, uint64(1), createReq).Return(resp, nil)

		req := createAuthenticatedRequest(http.MethodPost, "/api/v1/datasources", createReq)
		w := httptest.NewRecorder()

		handler.Create(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var result models.DataSourceResponse
		err := json.Unmarshal(w.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Equal(t, "test-ds", result.Name)

		mockService.AssertExpectations(t)
	})

	t.Run("name already exists", func(t *testing.T) {
		mockService := new(mocks.MockDataSourceService)
		handler := createDataSourceHandler(mockService)

		createReq := &models.CreateDataSourceRequest{
			Name:         "existing-ds",
			DBType:       "postgresql",
			Host:         "localhost",
			Port:         5432,
			DatabaseName: "testdb",
			Username:     "testuser",
			Password:     "testpass",
		}

		mockService.On("CreateDataSource", mock.Anything, uint64(1), createReq).
			Return(nil, services.ErrDataSourceNameExists)

		req := createAuthenticatedRequest(http.MethodPost, "/api/v1/datasources", createReq)
		w := httptest.NewRecorder()

		handler.Create(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("invalid db type", func(t *testing.T) {
		mockService := new(mocks.MockDataSourceService)
		handler := createDataSourceHandler(mockService)

		createReq := &models.CreateDataSourceRequest{
			Name:         "test-ds",
			DBType:       "mongodb",
			Host:         "localhost",
			Port:         27017,
			DatabaseName: "testdb",
			Username:     "testuser",
			Password:     "testpass",
		}

		// Validation should fail before reaching service
		req := createAuthenticatedRequest(http.MethodPost, "/api/v1/datasources", createReq)
		w := httptest.NewRecorder()

		handler.Create(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		mockService := new(mocks.MockDataSourceService)
		handler := createDataSourceHandler(mockService)

		createReq := &models.CreateDataSourceRequest{
			Name:         "test-ds",
			DBType:       "postgresql",
			Host:         "localhost",
			Port:         5432,
			DatabaseName: "testdb",
			Username:     "testuser",
			Password:     "testpass",
		}

		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/datasources", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Create(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestDataSourceHandler_Get(t *testing.T) {
	t.Run("successful get", func(t *testing.T) {
		mockService := new(mocks.MockDataSourceService)
		handler := createDataSourceHandler(mockService)

		resp := &models.DataSourceResponse{
			ID:     1,
			Name:   "test-ds",
			DBType: models.DBTypePostgreSQL,
		}

		mockService.On("GetDataSource", mock.Anything, uint64(1)).Return(resp, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/datasources/1", nil)
		w := httptest.NewRecorder()

		handler.Get(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var result models.DataSourceResponse
		err := json.Unmarshal(w.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Equal(t, "test-ds", result.Name)

		mockService.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockService := new(mocks.MockDataSourceService)
		handler := createDataSourceHandler(mockService)

		mockService.On("GetDataSource", mock.Anything, uint64(999)).
			Return(nil, services.ErrDataSourceNotFound)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/datasources/999", nil)
		w := httptest.NewRecorder()

		handler.Get(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		mockService.AssertExpectations(t)
	})
}

func TestDataSourceHandler_Update(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		mockService := new(mocks.MockDataSourceService)
		handler := createDataSourceHandler(mockService)

		updateReq := &models.UpdateDataSourceRequest{
			Name: "updated-ds",
			Host: "newhost",
		}

		resp := &models.DataSourceResponse{
			ID:     1,
			Name:   "updated-ds",
			Host:   "newhost",
			DBType: models.DBTypePostgreSQL,
		}

		mockService.On("UpdateDataSource", mock.Anything, uint64(1), updateReq).Return(resp, nil)

		req := createAuthenticatedRequest(http.MethodPut, "/api/v1/datasources/1", updateReq)
		w := httptest.NewRecorder()

		handler.Update(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var result models.DataSourceResponse
		err := json.Unmarshal(w.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Equal(t, "updated-ds", result.Name)

		mockService.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockService := new(mocks.MockDataSourceService)
		handler := createDataSourceHandler(mockService)

		updateReq := &models.UpdateDataSourceRequest{
			Name: "updated-ds",
		}

		mockService.On("UpdateDataSource", mock.Anything, uint64(999), updateReq).
			Return(nil, services.ErrDataSourceNotFound)

		req := createAuthenticatedRequest(http.MethodPut, "/api/v1/datasources/999", updateReq)
		w := httptest.NewRecorder()

		handler.Update(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		mockService.AssertExpectations(t)
	})
}

func TestDataSourceHandler_Delete(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		mockService := new(mocks.MockDataSourceService)
		handler := createDataSourceHandler(mockService)

		mockService.On("DeleteDataSource", mock.Anything, uint64(1)).Return(nil)

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/datasources/1", nil)
		w := httptest.NewRecorder()

		handler.Delete(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockService := new(mocks.MockDataSourceService)
		handler := createDataSourceHandler(mockService)

		mockService.On("DeleteDataSource", mock.Anything, uint64(999)).
			Return(services.ErrDataSourceNotFound)

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/datasources/999", nil)
		w := httptest.NewRecorder()

		handler.Delete(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		mockService.AssertExpectations(t)
	})
}

func TestDataSourceHandler_TestConnection(t *testing.T) {
	t.Run("successful connection", func(t *testing.T) {
		mockService := new(mocks.MockDataSourceService)
		handler := createDataSourceHandler(mockService)

		testReq := &models.TestConnectionRequest{
			DBType:       "postgresql",
			Host:         "localhost",
			Port:         5432,
			DatabaseName: "testdb",
			Username:     "testuser",
			Password:     "testpass",
		}

		resp := &models.TestConnectionResponse{
			Success: true,
			Message: "接続に成功しました",
		}

		mockService.On("TestConnection", mock.Anything, testReq).Return(resp, nil)

		body, _ := json.Marshal(testReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/datasources/test-connection", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.TestConnection(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var result models.TestConnectionResponse
		err := json.Unmarshal(w.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.True(t, result.Success)

		mockService.AssertExpectations(t)
	})

	t.Run("connection failed", func(t *testing.T) {
		mockService := new(mocks.MockDataSourceService)
		handler := createDataSourceHandler(mockService)

		testReq := &models.TestConnectionRequest{
			DBType:       "postgresql",
			Host:         "localhost",
			Port:         5432,
			DatabaseName: "testdb",
			Username:     "testuser",
			Password:     "wrongpass",
		}

		resp := &models.TestConnectionResponse{
			Success: false,
			Message: "authentication failed",
		}

		mockService.On("TestConnection", mock.Anything, testReq).Return(resp, nil)

		body, _ := json.Marshal(testReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/datasources/test-connection", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.TestConnection(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var result models.TestConnectionResponse
		err := json.Unmarshal(w.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.False(t, result.Success)

		mockService.AssertExpectations(t)
	})
}

func TestDataSourceHandler_GetTables(t *testing.T) {
	t.Run("successful get tables", func(t *testing.T) {
		mockService := new(mocks.MockDataSourceService)
		handler := createDataSourceHandler(mockService)

		resp := &models.TableListResponse{
			Tables: []models.TableInfo{
				{Name: "users", Schema: "public"},
				{Name: "orders", Schema: "public"},
			},
		}

		mockService.On("GetTables", mock.Anything, uint64(1)).Return(resp, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/datasources/1/tables", nil)
		w := httptest.NewRecorder()

		handler.GetTables(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var result models.TableListResponse
		err := json.Unmarshal(w.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Len(t, result.Tables, 2)

		mockService.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockService := new(mocks.MockDataSourceService)
		handler := createDataSourceHandler(mockService)

		mockService.On("GetTables", mock.Anything, uint64(999)).
			Return(nil, services.ErrDataSourceNotFound)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/datasources/999/tables", nil)
		w := httptest.NewRecorder()

		handler.GetTables(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		mockService.AssertExpectations(t)
	})
}

func TestDataSourceHandler_GetColumns(t *testing.T) {
	t.Run("successful get columns", func(t *testing.T) {
		mockService := new(mocks.MockDataSourceService)
		handler := createDataSourceHandler(mockService)

		resp := &models.ColumnListResponse{
			Columns: []models.ColumnInfo{
				{Name: "id", DataType: "integer", IsPrimaryKey: true},
				{Name: "name", DataType: "varchar", IsNullable: true},
			},
		}

		mockService.On("GetColumns", mock.Anything, uint64(1), "users").Return(resp, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/datasources/1/tables/users/columns", nil)
		w := httptest.NewRecorder()

		handler.GetColumns(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var result models.ColumnListResponse
		err := json.Unmarshal(w.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Len(t, result.Columns, 2)

		mockService.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockService := new(mocks.MockDataSourceService)
		handler := createDataSourceHandler(mockService)

		mockService.On("GetColumns", mock.Anything, uint64(999), "users").
			Return(nil, services.ErrDataSourceNotFound)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/datasources/999/tables/users/columns", nil)
		w := httptest.NewRecorder()

		handler.GetColumns(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		mockService.AssertExpectations(t)
	})
}
