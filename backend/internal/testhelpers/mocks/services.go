// Package mocks テスト用のモック実装を提供
package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"nocode-app/backend/internal/models"
	"nocode-app/backend/internal/repositories"
	"nocode-app/backend/internal/utils"
)

// MockAuthService AuthServiceInterfaceのモック実装
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(ctx context.Context, req *models.RegisterRequest) (*models.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AuthResponse), args.Error(1)
}

func (m *MockAuthService) Login(ctx context.Context, req *models.LoginRequest) (*models.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AuthResponse), args.Error(1)
}

func (m *MockAuthService) GetCurrentUser(ctx context.Context, userID uint64) (*models.UserResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserResponse), args.Error(1)
}

func (m *MockAuthService) RefreshToken(claims *utils.JWTClaims) (string, error) {
	args := m.Called(claims)
	return args.String(0), args.Error(1)
}

// MockAppService AppServiceInterfaceのモック実装
type MockAppService struct {
	mock.Mock
}

func (m *MockAppService) CreateApp(ctx context.Context, userID uint64, req *models.CreateAppRequest) (*models.AppResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AppResponse), args.Error(1)
}

func (m *MockAppService) GetApp(ctx context.Context, appID uint64) (*models.AppResponse, error) {
	args := m.Called(ctx, appID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AppResponse), args.Error(1)
}

func (m *MockAppService) GetApps(ctx context.Context, page, limit int) (*models.AppListResponse, error) {
	args := m.Called(ctx, page, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AppListResponse), args.Error(1)
}

func (m *MockAppService) UpdateApp(ctx context.Context, appID uint64, req *models.UpdateAppRequest) (*models.AppResponse, error) {
	args := m.Called(ctx, appID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AppResponse), args.Error(1)
}

func (m *MockAppService) DeleteApp(ctx context.Context, appID uint64) error {
	args := m.Called(ctx, appID)
	return args.Error(0)
}

func (m *MockAppService) CreateExternalApp(ctx context.Context, userID uint64, req *models.CreateExternalAppRequest) (*models.AppResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AppResponse), args.Error(1)
}

// MockFieldService FieldServiceInterfaceのモック実装
type MockFieldService struct {
	mock.Mock
}

func (m *MockFieldService) GetFields(ctx context.Context, appID uint64) ([]models.FieldResponse, error) {
	args := m.Called(ctx, appID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.FieldResponse), args.Error(1)
}

func (m *MockFieldService) CreateField(ctx context.Context, appID uint64, req *models.CreateFieldRequest) (*models.FieldResponse, error) {
	args := m.Called(ctx, appID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.FieldResponse), args.Error(1)
}

func (m *MockFieldService) UpdateField(ctx context.Context, fieldID uint64, req *models.UpdateFieldRequest) (*models.FieldResponse, error) {
	args := m.Called(ctx, fieldID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.FieldResponse), args.Error(1)
}

func (m *MockFieldService) DeleteField(ctx context.Context, appID, fieldID uint64) error {
	args := m.Called(ctx, appID, fieldID)
	return args.Error(0)
}

func (m *MockFieldService) UpdateFieldOrder(ctx context.Context, appID uint64, req *models.UpdateFieldOrderRequest) error {
	args := m.Called(ctx, appID, req)
	return args.Error(0)
}

// MockRecordService RecordServiceInterfaceのモック実装
type MockRecordService struct {
	mock.Mock
}

func (m *MockRecordService) GetRecords(ctx context.Context, appID uint64, opts repositories.RecordQueryOptions) (*models.RecordListResponse, error) {
	args := m.Called(ctx, appID, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RecordListResponse), args.Error(1)
}

func (m *MockRecordService) GetRecord(ctx context.Context, appID, recordID uint64) (*models.RecordResponse, error) {
	args := m.Called(ctx, appID, recordID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RecordResponse), args.Error(1)
}

func (m *MockRecordService) CreateRecord(ctx context.Context, appID, userID uint64, req *models.CreateRecordRequest) (*models.RecordResponse, error) {
	args := m.Called(ctx, appID, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RecordResponse), args.Error(1)
}

func (m *MockRecordService) UpdateRecord(ctx context.Context, appID, recordID uint64, req *models.UpdateRecordRequest) (*models.RecordResponse, error) {
	args := m.Called(ctx, appID, recordID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RecordResponse), args.Error(1)
}

func (m *MockRecordService) DeleteRecord(ctx context.Context, appID, recordID uint64) error {
	args := m.Called(ctx, appID, recordID)
	return args.Error(0)
}

func (m *MockRecordService) BulkCreateRecords(ctx context.Context, appID, userID uint64, req *models.BulkCreateRecordRequest) ([]models.RecordResponse, error) {
	args := m.Called(ctx, appID, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.RecordResponse), args.Error(1)
}

func (m *MockRecordService) BulkDeleteRecords(ctx context.Context, appID uint64, req *models.BulkDeleteRecordRequest) error {
	args := m.Called(ctx, appID, req)
	return args.Error(0)
}

// MockViewService ViewServiceInterfaceのモック実装
type MockViewService struct {
	mock.Mock
}

func (m *MockViewService) GetViews(ctx context.Context, appID uint64) ([]models.ViewResponse, error) {
	args := m.Called(ctx, appID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ViewResponse), args.Error(1)
}

func (m *MockViewService) CreateView(ctx context.Context, appID uint64, req *models.CreateViewRequest) (*models.ViewResponse, error) {
	args := m.Called(ctx, appID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ViewResponse), args.Error(1)
}

func (m *MockViewService) UpdateView(ctx context.Context, viewID uint64, req *models.UpdateViewRequest) (*models.ViewResponse, error) {
	args := m.Called(ctx, viewID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ViewResponse), args.Error(1)
}

func (m *MockViewService) DeleteView(ctx context.Context, viewID uint64) error {
	args := m.Called(ctx, viewID)
	return args.Error(0)
}

// MockChartService ChartServiceInterfaceのモック実装
type MockChartService struct {
	mock.Mock
}

func (m *MockChartService) GetChartData(ctx context.Context, appID uint64, req *models.ChartDataRequest) (*models.ChartDataResponse, error) {
	args := m.Called(ctx, appID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ChartDataResponse), args.Error(1)
}

func (m *MockChartService) GetChartConfigs(ctx context.Context, appID uint64) ([]models.ChartConfig, error) {
	args := m.Called(ctx, appID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ChartConfig), args.Error(1)
}

func (m *MockChartService) SaveChartConfig(ctx context.Context, appID, userID uint64, req *models.SaveChartConfigRequest) (*models.ChartConfig, error) {
	args := m.Called(ctx, appID, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ChartConfig), args.Error(1)
}

func (m *MockChartService) DeleteChartConfig(ctx context.Context, configID uint64) error {
	args := m.Called(ctx, configID)
	return args.Error(0)
}

// MockUserService UserServiceInterfaceのモック実装
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) GetUsers(ctx context.Context, callerRole string, page, limit int) (*models.UserListResponse, error) {
	args := m.Called(ctx, callerRole, page, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserListResponse), args.Error(1)
}

func (m *MockUserService) GetUser(ctx context.Context, callerRole string, userID uint64) (*models.UserResponse, error) {
	args := m.Called(ctx, callerRole, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserResponse), args.Error(1)
}

func (m *MockUserService) CreateUser(ctx context.Context, callerRole string, req *models.CreateUserRequest) (*models.UserResponse, error) {
	args := m.Called(ctx, callerRole, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserResponse), args.Error(1)
}

func (m *MockUserService) UpdateUser(ctx context.Context, callerID uint64, callerRole string, userID uint64, req *models.UpdateUserRequest) (*models.UserResponse, error) {
	args := m.Called(ctx, callerID, callerRole, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserResponse), args.Error(1)
}

func (m *MockUserService) DeleteUser(ctx context.Context, callerID uint64, callerRole string, userID uint64) error {
	args := m.Called(ctx, callerID, callerRole, userID)
	return args.Error(0)
}

func (m *MockUserService) UpdateProfile(ctx context.Context, userID uint64, req *models.UpdateProfileRequest) (*models.UserResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserResponse), args.Error(1)
}

func (m *MockUserService) ChangePassword(ctx context.Context, userID uint64, req *models.ChangePasswordRequest) error {
	args := m.Called(ctx, userID, req)
	return args.Error(0)
}

// MockDashboardService DashboardServiceInterfaceのモック実装
type MockDashboardService struct {
	mock.Mock
}

func (m *MockDashboardService) GetStats(ctx context.Context) (*models.DashboardStats, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DashboardStats), args.Error(1)
}

// MockDataSourceService DataSourceServiceInterfaceのモック実装
type MockDataSourceService struct {
	mock.Mock
}

func (m *MockDataSourceService) CreateDataSource(ctx context.Context, userID uint64, req *models.CreateDataSourceRequest) (*models.DataSourceResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DataSourceResponse), args.Error(1)
}

func (m *MockDataSourceService) GetDataSource(ctx context.Context, id uint64) (*models.DataSourceResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DataSourceResponse), args.Error(1)
}

func (m *MockDataSourceService) GetDataSources(ctx context.Context, page, limit int) (*models.DataSourceListResponse, error) {
	args := m.Called(ctx, page, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DataSourceListResponse), args.Error(1)
}

func (m *MockDataSourceService) UpdateDataSource(ctx context.Context, id uint64, req *models.UpdateDataSourceRequest) (*models.DataSourceResponse, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DataSourceResponse), args.Error(1)
}

func (m *MockDataSourceService) DeleteDataSource(ctx context.Context, id uint64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDataSourceService) TestConnection(ctx context.Context, req *models.TestConnectionRequest) (*models.TestConnectionResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TestConnectionResponse), args.Error(1)
}

func (m *MockDataSourceService) GetTables(ctx context.Context, id uint64) (*models.TableListResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TableListResponse), args.Error(1)
}

func (m *MockDataSourceService) GetColumns(ctx context.Context, id uint64, tableName string) (*models.ColumnListResponse, error) {
	args := m.Called(ctx, id, tableName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ColumnListResponse), args.Error(1)
}

func (m *MockDataSourceService) GetDecryptedPassword(ctx context.Context, id uint64) (string, error) {
	args := m.Called(ctx, id)
	return args.String(0), args.Error(1)
}

// MockDashboardWidgetService DashboardWidgetServiceInterfaceのモック実装
type MockDashboardWidgetService struct {
	mock.Mock
}

func (m *MockDashboardWidgetService) GetWidgets(ctx context.Context, userID uint64) (*models.DashboardWidgetListResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DashboardWidgetListResponse), args.Error(1)
}

func (m *MockDashboardWidgetService) GetVisibleWidgets(ctx context.Context, userID uint64) (*models.DashboardWidgetListResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DashboardWidgetListResponse), args.Error(1)
}

func (m *MockDashboardWidgetService) CreateWidget(ctx context.Context, userID uint64, req *models.CreateDashboardWidgetRequest) (*models.DashboardWidgetResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DashboardWidgetResponse), args.Error(1)
}

func (m *MockDashboardWidgetService) UpdateWidget(ctx context.Context, userID, widgetID uint64, req *models.UpdateDashboardWidgetRequest) (*models.DashboardWidgetResponse, error) {
	args := m.Called(ctx, userID, widgetID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DashboardWidgetResponse), args.Error(1)
}

func (m *MockDashboardWidgetService) DeleteWidget(ctx context.Context, userID, widgetID uint64) error {
	args := m.Called(ctx, userID, widgetID)
	return args.Error(0)
}

func (m *MockDashboardWidgetService) ReorderWidgets(ctx context.Context, userID uint64, req *models.ReorderWidgetsRequest) error {
	args := m.Called(ctx, userID, req)
	return args.Error(0)
}

func (m *MockDashboardWidgetService) ToggleVisibility(ctx context.Context, userID, widgetID uint64) (*models.DashboardWidgetResponse, error) {
	args := m.Called(ctx, userID, widgetID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DashboardWidgetResponse), args.Error(1)
}
