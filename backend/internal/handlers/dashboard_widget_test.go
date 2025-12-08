package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"nocode-app/backend/internal/middleware"
	"nocode-app/backend/internal/models"
	"nocode-app/backend/internal/testhelpers/mocks"
	"nocode-app/backend/internal/utils"
)

// setupDashboardWidgetHandler テスト用のハンドラーをセットアップ
func setupDashboardWidgetHandler() (*DashboardWidgetHandler, *mocks.MockDashboardWidgetService) {
	mockService := new(mocks.MockDashboardWidgetService)
	validator := utils.NewValidator()
	handler := NewDashboardWidgetHandler(mockService, validator)
	return handler, mockService
}

// createAuthenticatedRequest 認証済みリクエストを作成
func createAuthenticatedRequest(method, path string, body interface{}) *http.Request {
	var req *http.Request
	if body != nil {
		bodyBytes, _ := json.Marshal(body)
		req = httptest.NewRequest(method, path, bytes.NewReader(bodyBytes))
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	req.Header.Set("Content-Type", "application/json")

	// 認証コンテキストを追加
	claims := &utils.JWTClaims{
		UserID: 1,
		Email:  "test@example.com",
		Role:   "admin",
	}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, claims)
	return req.WithContext(ctx)
}

// TestDashboardWidgetHandler_List Listメソッドのテスト
func TestDashboardWidgetHandler_List(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		mockSetup      func(*mocks.MockDashboardWidgetService)
		expectedStatus int
	}{
		{
			name:  "正常系_全ウィジェット取得",
			query: "",
			mockSetup: func(m *mocks.MockDashboardWidgetService) {
				m.On("GetWidgets", mock.Anything, uint64(1)).Return(&models.DashboardWidgetListResponse{
					Widgets: []models.DashboardWidgetResponse{
						{
							ID:           1,
							UserID:       1,
							AppID:        1,
							DisplayOrder: 0,
							ViewType:     "table",
							IsVisible:    true,
							WidgetSize:   "medium",
						},
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "正常系_表示中のみ取得",
			query: "?visible=true",
			mockSetup: func(m *mocks.MockDashboardWidgetService) {
				m.On("GetVisibleWidgets", mock.Anything, uint64(1)).Return(&models.DashboardWidgetListResponse{
					Widgets: []models.DashboardWidgetResponse{
						{
							ID:           1,
							UserID:       1,
							AppID:        1,
							DisplayOrder: 0,
							ViewType:     "table",
							IsVisible:    true,
							WidgetSize:   "medium",
						},
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "異常系_サービスエラー",
			query: "",
			mockSetup: func(m *mocks.MockDashboardWidgetService) {
				m.On("GetWidgets", mock.Anything, uint64(1)).Return(nil, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := setupDashboardWidgetHandler()
			tt.mockSetup(mockService)

			req := createAuthenticatedRequest(http.MethodGet, "/api/v1/dashboard/widgets"+tt.query, nil)
			rr := httptest.NewRecorder()

			handler.List(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockService.AssertExpectations(t)
		})
	}
}

// TestDashboardWidgetHandler_List_MethodNotAllowed メソッド不許可テスト
func TestDashboardWidgetHandler_List_MethodNotAllowed(t *testing.T) {
	handler, _ := setupDashboardWidgetHandler()

	req := createAuthenticatedRequest(http.MethodPost, "/api/v1/dashboard/widgets", nil)
	rr := httptest.NewRecorder()

	handler.List(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

// TestDashboardWidgetHandler_Create Createメソッドのテスト
func TestDashboardWidgetHandler_Create(t *testing.T) {
	tests := []struct {
		name           string
		body           interface{}
		mockSetup      func(*mocks.MockDashboardWidgetService)
		expectedStatus int
	}{
		{
			name: "正常系_ウィジェット作成",
			body: models.CreateDashboardWidgetRequest{
				AppID:    1,
				ViewType: "table",
			},
			mockSetup: func(m *mocks.MockDashboardWidgetService) {
				m.On("CreateWidget", mock.Anything, uint64(1), mock.Anything).Return(&models.DashboardWidgetResponse{
					ID:           1,
					UserID:       1,
					AppID:        1,
					DisplayOrder: 0,
					ViewType:     "table",
					IsVisible:    true,
					WidgetSize:   "medium",
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "異常系_アプリなし",
			body: models.CreateDashboardWidgetRequest{
				AppID: 999,
			},
			mockSetup: func(m *mocks.MockDashboardWidgetService) {
				m.On("CreateWidget", mock.Anything, uint64(1), mock.Anything).Return(nil, errors.New("指定されたアプリが見つかりません"))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "異常系_既に存在",
			body: models.CreateDashboardWidgetRequest{
				AppID: 1,
			},
			mockSetup: func(m *mocks.MockDashboardWidgetService) {
				m.On("CreateWidget", mock.Anything, uint64(1), mock.Anything).Return(nil, errors.New("このアプリのウィジェットは既に存在します"))
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "異常系_無効なJSON",
			body:           "invalid json",
			mockSetup:      func(m *mocks.MockDashboardWidgetService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := setupDashboardWidgetHandler()
			tt.mockSetup(mockService)

			req := createAuthenticatedRequest(http.MethodPost, "/api/v1/dashboard/widgets", tt.body)
			rr := httptest.NewRecorder()

			handler.Create(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockService.AssertExpectations(t)
		})
	}
}

// TestDashboardWidgetHandler_Update Updateメソッドのテスト
func TestDashboardWidgetHandler_Update(t *testing.T) {
	isVisible := false

	tests := []struct {
		name           string
		path           string
		body           interface{}
		mockSetup      func(*mocks.MockDashboardWidgetService)
		expectedStatus int
	}{
		{
			name: "正常系_ウィジェット更新",
			path: "/api/v1/dashboard/widgets/1",
			body: models.UpdateDashboardWidgetRequest{
				ViewType:  "chart",
				IsVisible: &isVisible,
			},
			mockSetup: func(m *mocks.MockDashboardWidgetService) {
				m.On("UpdateWidget", mock.Anything, uint64(1), uint64(1), mock.Anything).Return(&models.DashboardWidgetResponse{
					ID:           1,
					UserID:       1,
					AppID:        1,
					DisplayOrder: 0,
					ViewType:     "chart",
					IsVisible:    false,
					WidgetSize:   "medium",
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "異常系_ウィジェットなし",
			path: "/api/v1/dashboard/widgets/999",
			body: models.UpdateDashboardWidgetRequest{},
			mockSetup: func(m *mocks.MockDashboardWidgetService) {
				m.On("UpdateWidget", mock.Anything, uint64(1), uint64(999), mock.Anything).Return(nil, errors.New("ウィジェットが見つかりません"))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "異常系_権限なし",
			path: "/api/v1/dashboard/widgets/1",
			body: models.UpdateDashboardWidgetRequest{},
			mockSetup: func(m *mocks.MockDashboardWidgetService) {
				m.On("UpdateWidget", mock.Anything, uint64(1), uint64(1), mock.Anything).Return(nil, errors.New("このウィジェットを更新する権限がありません"))
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "異常系_無効なID",
			path:           "/api/v1/dashboard/widgets/invalid",
			body:           models.UpdateDashboardWidgetRequest{},
			mockSetup:      func(m *mocks.MockDashboardWidgetService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := setupDashboardWidgetHandler()
			tt.mockSetup(mockService)

			req := createAuthenticatedRequest(http.MethodPut, tt.path, tt.body)
			rr := httptest.NewRecorder()

			handler.Update(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockService.AssertExpectations(t)
		})
	}
}

// TestDashboardWidgetHandler_Delete Deleteメソッドのテスト
func TestDashboardWidgetHandler_Delete(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		mockSetup      func(*mocks.MockDashboardWidgetService)
		expectedStatus int
	}{
		{
			name: "正常系_ウィジェット削除",
			path: "/api/v1/dashboard/widgets/1",
			mockSetup: func(m *mocks.MockDashboardWidgetService) {
				m.On("DeleteWidget", mock.Anything, uint64(1), uint64(1)).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name: "異常系_ウィジェットなし",
			path: "/api/v1/dashboard/widgets/999",
			mockSetup: func(m *mocks.MockDashboardWidgetService) {
				m.On("DeleteWidget", mock.Anything, uint64(1), uint64(999)).Return(errors.New("ウィジェットが見つかりません"))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "異常系_権限なし",
			path: "/api/v1/dashboard/widgets/1",
			mockSetup: func(m *mocks.MockDashboardWidgetService) {
				m.On("DeleteWidget", mock.Anything, uint64(1), uint64(1)).Return(errors.New("このウィジェットを削除する権限がありません"))
			},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := setupDashboardWidgetHandler()
			tt.mockSetup(mockService)

			req := createAuthenticatedRequest(http.MethodDelete, tt.path, nil)
			rr := httptest.NewRecorder()

			handler.Delete(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockService.AssertExpectations(t)
		})
	}
}

// TestDashboardWidgetHandler_Reorder Reorderメソッドのテスト
func TestDashboardWidgetHandler_Reorder(t *testing.T) {
	tests := []struct {
		name           string
		body           interface{}
		mockSetup      func(*mocks.MockDashboardWidgetService)
		expectedStatus int
	}{
		{
			name: "正常系_並び替え成功",
			body: models.ReorderWidgetsRequest{
				WidgetIDs: []uint64{2, 1, 3},
			},
			mockSetup: func(m *mocks.MockDashboardWidgetService) {
				m.On("ReorderWidgets", mock.Anything, uint64(1), mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "異常系_無効なウィジェットID",
			body: models.ReorderWidgetsRequest{
				WidgetIDs: []uint64{1, 999},
			},
			mockSetup: func(m *mocks.MockDashboardWidgetService) {
				m.On("ReorderWidgets", mock.Anything, uint64(1), mock.Anything).Return(errors.New("無効なウィジェットIDが含まれています"))
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "異常系_空の配列",
			body: models.ReorderWidgetsRequest{
				WidgetIDs: []uint64{},
			},
			mockSetup:      func(m *mocks.MockDashboardWidgetService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := setupDashboardWidgetHandler()
			tt.mockSetup(mockService)

			req := createAuthenticatedRequest(http.MethodPut, "/api/v1/dashboard/widgets/reorder", tt.body)
			rr := httptest.NewRecorder()

			handler.Reorder(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockService.AssertExpectations(t)
		})
	}
}

// TestDashboardWidgetHandler_ToggleVisibility ToggleVisibilityメソッドのテスト
func TestDashboardWidgetHandler_ToggleVisibility(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		mockSetup      func(*mocks.MockDashboardWidgetService)
		expectedStatus int
	}{
		{
			name: "正常系_表示切り替え",
			path: "/api/v1/dashboard/widgets/1/toggle",
			mockSetup: func(m *mocks.MockDashboardWidgetService) {
				m.On("ToggleVisibility", mock.Anything, uint64(1), uint64(1)).Return(&models.DashboardWidgetResponse{
					ID:        1,
					UserID:    1,
					AppID:     1,
					IsVisible: false,
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "異常系_ウィジェットなし",
			path: "/api/v1/dashboard/widgets/999/toggle",
			mockSetup: func(m *mocks.MockDashboardWidgetService) {
				m.On("ToggleVisibility", mock.Anything, uint64(1), uint64(999)).Return(nil, errors.New("ウィジェットが見つかりません"))
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := setupDashboardWidgetHandler()
			tt.mockSetup(mockService)

			req := createAuthenticatedRequest(http.MethodPost, tt.path, nil)
			rr := httptest.NewRecorder()

			handler.ToggleVisibility(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockService.AssertExpectations(t)
		})
	}
}

// TestDashboardWidgetHandler_Unauthorized 未認証テスト
func TestDashboardWidgetHandler_Unauthorized(t *testing.T) {
	handler, _ := setupDashboardWidgetHandler()

	// 認証なしリクエスト
	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/widgets", nil)
	rr := httptest.NewRecorder()

	handler.List(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
