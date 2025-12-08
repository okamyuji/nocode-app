package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"nocode-app/backend/internal/models"
	"nocode-app/backend/internal/testhelpers/mocks"
)

// TestDashboardWidgetService_GetWidgets GetWidgetsメソッドのテスト
func TestDashboardWidgetService_GetWidgets(t *testing.T) {
	tests := []struct {
		name          string
		userID        uint64
		mockSetup     func(*mocks.MockDashboardWidgetRepository)
		expectedLen   int
		expectedError bool
	}{
		{
			name:   "正常系_ウィジェット一覧取得",
			userID: 1,
			mockSetup: func(m *mocks.MockDashboardWidgetRepository) {
				m.On("GetByUserIDWithApps", mock.Anything, uint64(1)).Return([]models.DashboardWidget{
					{
						ID:           1,
						UserID:       1,
						AppID:        1,
						DisplayOrder: 0,
						ViewType:     models.WidgetViewTypeTable,
						IsVisible:    true,
						WidgetSize:   models.WidgetSizeMedium,
						App:          &models.App{ID: 1, Name: "Test App"},
					},
					{
						ID:           2,
						UserID:       1,
						AppID:        2,
						DisplayOrder: 1,
						ViewType:     models.WidgetViewTypeList,
						IsVisible:    false,
						WidgetSize:   models.WidgetSizeLarge,
						App:          &models.App{ID: 2, Name: "Test App 2"},
					},
				}, nil)
			},
			expectedLen:   2,
			expectedError: false,
		},
		{
			name:   "正常系_ウィジェットなし",
			userID: 1,
			mockSetup: func(m *mocks.MockDashboardWidgetRepository) {
				m.On("GetByUserIDWithApps", mock.Anything, uint64(1)).Return([]models.DashboardWidget{}, nil)
			},
			expectedLen:   0,
			expectedError: false,
		},
		{
			name:   "異常系_リポジトリエラー",
			userID: 1,
			mockSetup: func(m *mocks.MockDashboardWidgetRepository) {
				m.On("GetByUserIDWithApps", mock.Anything, uint64(1)).Return(nil, errors.New("database error"))
			},
			expectedLen:   0,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWidgetRepo := new(mocks.MockDashboardWidgetRepository)
			mockAppRepo := new(mocks.MockAppRepository)
			tt.mockSetup(mockWidgetRepo)

			service := NewDashboardWidgetService(mockWidgetRepo, mockAppRepo)
			result, err := service.GetWidgets(context.Background(), tt.userID)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result.Widgets, tt.expectedLen)
			}

			mockWidgetRepo.AssertExpectations(t)
		})
	}
}

// TestDashboardWidgetService_GetVisibleWidgets GetVisibleWidgetsメソッドのテスト
func TestDashboardWidgetService_GetVisibleWidgets(t *testing.T) {
	tests := []struct {
		name          string
		userID        uint64
		mockSetup     func(*mocks.MockDashboardWidgetRepository)
		expectedLen   int
		expectedError bool
	}{
		{
			name:   "正常系_表示中ウィジェット取得",
			userID: 1,
			mockSetup: func(m *mocks.MockDashboardWidgetRepository) {
				m.On("GetVisibleByUserID", mock.Anything, uint64(1)).Return([]models.DashboardWidget{
					{
						ID:           1,
						UserID:       1,
						AppID:        1,
						DisplayOrder: 0,
						ViewType:     models.WidgetViewTypeTable,
						IsVisible:    true,
						WidgetSize:   models.WidgetSizeMedium,
						App:          &models.App{ID: 1, Name: "Test App"},
					},
				}, nil)
			},
			expectedLen:   1,
			expectedError: false,
		},
		{
			name:   "異常系_リポジトリエラー",
			userID: 1,
			mockSetup: func(m *mocks.MockDashboardWidgetRepository) {
				m.On("GetVisibleByUserID", mock.Anything, uint64(1)).Return(nil, errors.New("database error"))
			},
			expectedLen:   0,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWidgetRepo := new(mocks.MockDashboardWidgetRepository)
			mockAppRepo := new(mocks.MockAppRepository)
			tt.mockSetup(mockWidgetRepo)

			service := NewDashboardWidgetService(mockWidgetRepo, mockAppRepo)
			result, err := service.GetVisibleWidgets(context.Background(), tt.userID)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result.Widgets, tt.expectedLen)
			}

			mockWidgetRepo.AssertExpectations(t)
		})
	}
}

// TestDashboardWidgetService_CreateWidget CreateWidgetメソッドのテスト
func TestDashboardWidgetService_CreateWidget(t *testing.T) {
	tests := []struct {
		name          string
		userID        uint64
		req           *models.CreateDashboardWidgetRequest
		mockSetup     func(*mocks.MockDashboardWidgetRepository, *mocks.MockAppRepository)
		expectedError bool
		errorContains string
	}{
		{
			name:   "正常系_ウィジェット作成",
			userID: 1,
			req: &models.CreateDashboardWidgetRequest{
				AppID:    1,
				ViewType: "table",
			},
			mockSetup: func(mr *mocks.MockDashboardWidgetRepository, ma *mocks.MockAppRepository) {
				ma.On("GetByID", mock.Anything, uint64(1)).Return(&models.App{
					ID:   1,
					Name: "Test App",
				}, nil)
				mr.On("Exists", mock.Anything, uint64(1), uint64(1)).Return(false, nil)
				mr.On("GetMaxDisplayOrder", mock.Anything, uint64(1)).Return(0, nil)
				mr.On("Create", mock.Anything, mock.Anything).Return(nil)
				mr.On("GetByID", mock.Anything, mock.Anything).Return(&models.DashboardWidget{
					ID:           1,
					UserID:       1,
					AppID:        1,
					DisplayOrder: 1,
					ViewType:     models.WidgetViewTypeTable,
					IsVisible:    true,
					WidgetSize:   models.WidgetSizeMedium,
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}, nil)
			},
			expectedError: false,
		},
		{
			name:   "異常系_アプリが存在しない",
			userID: 1,
			req: &models.CreateDashboardWidgetRequest{
				AppID: 999,
			},
			mockSetup: func(mr *mocks.MockDashboardWidgetRepository, ma *mocks.MockAppRepository) {
				ma.On("GetByID", mock.Anything, uint64(999)).Return(nil, nil)
			},
			expectedError: true,
			errorContains: "見つかりません",
		},
		{
			name:   "異常系_ウィジェット既存",
			userID: 1,
			req: &models.CreateDashboardWidgetRequest{
				AppID: 1,
			},
			mockSetup: func(mr *mocks.MockDashboardWidgetRepository, ma *mocks.MockAppRepository) {
				ma.On("GetByID", mock.Anything, uint64(1)).Return(&models.App{
					ID:   1,
					Name: "Test App",
				}, nil)
				mr.On("Exists", mock.Anything, uint64(1), uint64(1)).Return(true, nil)
			},
			expectedError: true,
			errorContains: "既に存在",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWidgetRepo := new(mocks.MockDashboardWidgetRepository)
			mockAppRepo := new(mocks.MockAppRepository)
			tt.mockSetup(mockWidgetRepo, mockAppRepo)

			service := NewDashboardWidgetService(mockWidgetRepo, mockAppRepo)
			result, err := service.CreateWidget(context.Background(), tt.userID, tt.req)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockWidgetRepo.AssertExpectations(t)
			mockAppRepo.AssertExpectations(t)
		})
	}
}

// TestDashboardWidgetService_UpdateWidget UpdateWidgetメソッドのテスト
func TestDashboardWidgetService_UpdateWidget(t *testing.T) {
	viewTypeChart := "chart"
	isVisibleFalse := false

	tests := []struct {
		name          string
		userID        uint64
		widgetID      uint64
		req           *models.UpdateDashboardWidgetRequest
		mockSetup     func(*mocks.MockDashboardWidgetRepository, *mocks.MockAppRepository)
		expectedError bool
		errorContains string
	}{
		{
			name:     "正常系_ウィジェット更新",
			userID:   1,
			widgetID: 1,
			req: &models.UpdateDashboardWidgetRequest{
				ViewType:  viewTypeChart,
				IsVisible: &isVisibleFalse,
			},
			mockSetup: func(mr *mocks.MockDashboardWidgetRepository, ma *mocks.MockAppRepository) {
				mr.On("GetByID", mock.Anything, uint64(1)).Return(&models.DashboardWidget{
					ID:           1,
					UserID:       1,
					AppID:        1,
					DisplayOrder: 0,
					ViewType:     models.WidgetViewTypeTable,
					IsVisible:    true,
					WidgetSize:   models.WidgetSizeMedium,
				}, nil)
				mr.On("Update", mock.Anything, mock.Anything).Return(nil)
				ma.On("GetByIDWithFields", mock.Anything, uint64(1)).Return(&models.App{
					ID:   1,
					Name: "Test App",
				}, nil)
			},
			expectedError: false,
		},
		{
			name:     "異常系_ウィジェットなし",
			userID:   1,
			widgetID: 999,
			req:      &models.UpdateDashboardWidgetRequest{},
			mockSetup: func(mr *mocks.MockDashboardWidgetRepository, ma *mocks.MockAppRepository) {
				mr.On("GetByID", mock.Anything, uint64(999)).Return(nil, nil)
			},
			expectedError: true,
			errorContains: "見つかりません",
		},
		{
			name:     "異常系_権限なし",
			userID:   2,
			widgetID: 1,
			req:      &models.UpdateDashboardWidgetRequest{},
			mockSetup: func(mr *mocks.MockDashboardWidgetRepository, ma *mocks.MockAppRepository) {
				mr.On("GetByID", mock.Anything, uint64(1)).Return(&models.DashboardWidget{
					ID:     1,
					UserID: 1, // 別のユーザー
					AppID:  1,
				}, nil)
			},
			expectedError: true,
			errorContains: "権限がありません",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWidgetRepo := new(mocks.MockDashboardWidgetRepository)
			mockAppRepo := new(mocks.MockAppRepository)
			tt.mockSetup(mockWidgetRepo, mockAppRepo)

			service := NewDashboardWidgetService(mockWidgetRepo, mockAppRepo)
			result, err := service.UpdateWidget(context.Background(), tt.userID, tt.widgetID, tt.req)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockWidgetRepo.AssertExpectations(t)
			mockAppRepo.AssertExpectations(t)
		})
	}
}

// TestDashboardWidgetService_DeleteWidget DeleteWidgetメソッドのテスト
func TestDashboardWidgetService_DeleteWidget(t *testing.T) {
	tests := []struct {
		name          string
		userID        uint64
		widgetID      uint64
		mockSetup     func(*mocks.MockDashboardWidgetRepository)
		expectedError bool
		errorContains string
	}{
		{
			name:     "正常系_ウィジェット削除",
			userID:   1,
			widgetID: 1,
			mockSetup: func(mr *mocks.MockDashboardWidgetRepository) {
				mr.On("GetByID", mock.Anything, uint64(1)).Return(&models.DashboardWidget{
					ID:     1,
					UserID: 1,
					AppID:  1,
				}, nil)
				mr.On("Delete", mock.Anything, uint64(1)).Return(nil)
			},
			expectedError: false,
		},
		{
			name:     "異常系_ウィジェットなし",
			userID:   1,
			widgetID: 999,
			mockSetup: func(mr *mocks.MockDashboardWidgetRepository) {
				mr.On("GetByID", mock.Anything, uint64(999)).Return(nil, nil)
			},
			expectedError: true,
			errorContains: "見つかりません",
		},
		{
			name:     "異常系_権限なし",
			userID:   2,
			widgetID: 1,
			mockSetup: func(mr *mocks.MockDashboardWidgetRepository) {
				mr.On("GetByID", mock.Anything, uint64(1)).Return(&models.DashboardWidget{
					ID:     1,
					UserID: 1, // 別のユーザー
					AppID:  1,
				}, nil)
			},
			expectedError: true,
			errorContains: "権限がありません",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWidgetRepo := new(mocks.MockDashboardWidgetRepository)
			mockAppRepo := new(mocks.MockAppRepository)
			tt.mockSetup(mockWidgetRepo)

			service := NewDashboardWidgetService(mockWidgetRepo, mockAppRepo)
			err := service.DeleteWidget(context.Background(), tt.userID, tt.widgetID)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}

			mockWidgetRepo.AssertExpectations(t)
		})
	}
}

// TestDashboardWidgetService_ReorderWidgets ReorderWidgetsメソッドのテスト
func TestDashboardWidgetService_ReorderWidgets(t *testing.T) {
	tests := []struct {
		name          string
		userID        uint64
		req           *models.ReorderWidgetsRequest
		mockSetup     func(*mocks.MockDashboardWidgetRepository)
		expectedError bool
		errorContains string
	}{
		{
			name:   "正常系_並び替え成功",
			userID: 1,
			req: &models.ReorderWidgetsRequest{
				WidgetIDs: []uint64{2, 1, 3},
			},
			mockSetup: func(mr *mocks.MockDashboardWidgetRepository) {
				mr.On("GetByUserID", mock.Anything, uint64(1)).Return([]models.DashboardWidget{
					{ID: 1, UserID: 1},
					{ID: 2, UserID: 1},
					{ID: 3, UserID: 1},
				}, nil)
				mr.On("UpdateDisplayOrders", mock.Anything, uint64(1), []uint64{2, 1, 3}).Return(nil)
			},
			expectedError: false,
		},
		{
			name:   "異常系_無効なウィジェットID",
			userID: 1,
			req: &models.ReorderWidgetsRequest{
				WidgetIDs: []uint64{1, 999},
			},
			mockSetup: func(mr *mocks.MockDashboardWidgetRepository) {
				mr.On("GetByUserID", mock.Anything, uint64(1)).Return([]models.DashboardWidget{
					{ID: 1, UserID: 1},
				}, nil)
			},
			expectedError: true,
			errorContains: "無効なウィジェットID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWidgetRepo := new(mocks.MockDashboardWidgetRepository)
			mockAppRepo := new(mocks.MockAppRepository)
			tt.mockSetup(mockWidgetRepo)

			service := NewDashboardWidgetService(mockWidgetRepo, mockAppRepo)
			err := service.ReorderWidgets(context.Background(), tt.userID, tt.req)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}

			mockWidgetRepo.AssertExpectations(t)
		})
	}
}

// TestDashboardWidgetService_ToggleVisibility ToggleVisibilityメソッドのテスト
func TestDashboardWidgetService_ToggleVisibility(t *testing.T) {
	tests := []struct {
		name            string
		userID          uint64
		widgetID        uint64
		mockSetup       func(*mocks.MockDashboardWidgetRepository, *mocks.MockAppRepository)
		expectedVisible bool
		expectedError   bool
		errorContains   string
	}{
		{
			name:     "正常系_表示から非表示へ切り替え",
			userID:   1,
			widgetID: 1,
			mockSetup: func(mr *mocks.MockDashboardWidgetRepository, ma *mocks.MockAppRepository) {
				mr.On("GetByID", mock.Anything, uint64(1)).Return(&models.DashboardWidget{
					ID:        1,
					UserID:    1,
					AppID:     1,
					IsVisible: true,
				}, nil)
				mr.On("Update", mock.Anything, mock.Anything).Return(nil)
				ma.On("GetByIDWithFields", mock.Anything, uint64(1)).Return(&models.App{
					ID:   1,
					Name: "Test App",
				}, nil)
			},
			expectedVisible: false,
			expectedError:   false,
		},
		{
			name:     "正常系_非表示から表示へ切り替え",
			userID:   1,
			widgetID: 1,
			mockSetup: func(mr *mocks.MockDashboardWidgetRepository, ma *mocks.MockAppRepository) {
				mr.On("GetByID", mock.Anything, uint64(1)).Return(&models.DashboardWidget{
					ID:        1,
					UserID:    1,
					AppID:     1,
					IsVisible: false,
				}, nil)
				mr.On("Update", mock.Anything, mock.Anything).Return(nil)
				ma.On("GetByIDWithFields", mock.Anything, uint64(1)).Return(&models.App{
					ID:   1,
					Name: "Test App",
				}, nil)
			},
			expectedVisible: true,
			expectedError:   false,
		},
		{
			name:     "異常系_ウィジェットなし",
			userID:   1,
			widgetID: 999,
			mockSetup: func(mr *mocks.MockDashboardWidgetRepository, ma *mocks.MockAppRepository) {
				mr.On("GetByID", mock.Anything, uint64(999)).Return(nil, nil)
			},
			expectedError: true,
			errorContains: "見つかりません",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWidgetRepo := new(mocks.MockDashboardWidgetRepository)
			mockAppRepo := new(mocks.MockAppRepository)
			tt.mockSetup(mockWidgetRepo, mockAppRepo)

			service := NewDashboardWidgetService(mockWidgetRepo, mockAppRepo)
			result, err := service.ToggleVisibility(context.Background(), tt.userID, tt.widgetID)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedVisible, result.IsVisible)
			}

			mockWidgetRepo.AssertExpectations(t)
			mockAppRepo.AssertExpectations(t)
		})
	}
}

// TestWidgetViewType_IsValid WidgetViewTypeのバリデーションテスト
func TestWidgetViewType_IsValid(t *testing.T) {
	tests := []struct {
		viewType models.WidgetViewType
		expected bool
	}{
		{models.WidgetViewTypeTable, true},
		{models.WidgetViewTypeList, true},
		{models.WidgetViewTypeChart, true},
		{models.WidgetViewType("invalid"), false},
		{models.WidgetViewType(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.viewType), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.viewType.IsValid())
		})
	}
}

// TestWidgetSize_IsValid WidgetSizeのバリデーションテスト
func TestWidgetSize_IsValid(t *testing.T) {
	tests := []struct {
		size     models.WidgetSize
		expected bool
	}{
		{models.WidgetSizeSmall, true},
		{models.WidgetSizeMedium, true},
		{models.WidgetSizeLarge, true},
		{models.WidgetSize("invalid"), false},
		{models.WidgetSize(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.size), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.size.IsValid())
		})
	}
}
