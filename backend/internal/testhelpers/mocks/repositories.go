// Package mocks テスト用のモック実装を提供
package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"nocode-app/backend/internal/models"
	"nocode-app/backend/internal/repositories"
)

// MockUserRepository UserRepositoryInterfaceのモック実装
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uint64) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uint64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) GetAll(ctx context.Context, page, limit int) ([]models.User, int64, error) {
	args := m.Called(ctx, page, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserRepository) EmailExistsExcludingUser(ctx context.Context, email string, excludeID uint64) (bool, error) {
	args := m.Called(ctx, email, excludeID)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

// MockAppRepository AppRepositoryInterfaceのモック実装
type MockAppRepository struct {
	mock.Mock
}

func (m *MockAppRepository) Create(ctx context.Context, app *models.App) error {
	args := m.Called(ctx, app)
	return args.Error(0)
}

func (m *MockAppRepository) GetByID(ctx context.Context, id uint64) (*models.App, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.App), args.Error(1)
}

func (m *MockAppRepository) GetByIDWithFields(ctx context.Context, id uint64) (*models.App, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.App), args.Error(1)
}

func (m *MockAppRepository) GetAll(ctx context.Context, page, limit int) ([]models.App, int64, error) {
	args := m.Called(ctx, page, limit)
	return args.Get(0).([]models.App), args.Get(1).(int64), args.Error(2)
}

func (m *MockAppRepository) GetByUserID(ctx context.Context, userID uint64, page, limit int) ([]models.App, int64, error) {
	args := m.Called(ctx, userID, page, limit)
	return args.Get(0).([]models.App), args.Get(1).(int64), args.Error(2)
}

func (m *MockAppRepository) Update(ctx context.Context, app *models.App) error {
	args := m.Called(ctx, app)
	return args.Error(0)
}

func (m *MockAppRepository) Delete(ctx context.Context, id uint64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAppRepository) GetTableName(ctx context.Context, appID uint64) (string, error) {
	args := m.Called(ctx, appID)
	return args.String(0), args.Error(1)
}

func (m *MockAppRepository) GetAllTableNames(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

// MockFieldRepository FieldRepositoryInterfaceのモック実装
type MockFieldRepository struct {
	mock.Mock
}

func (m *MockFieldRepository) Create(ctx context.Context, field *models.AppField) error {
	args := m.Called(ctx, field)
	return args.Error(0)
}

func (m *MockFieldRepository) CreateBatch(ctx context.Context, fields []models.AppField) error {
	args := m.Called(ctx, fields)
	return args.Error(0)
}

func (m *MockFieldRepository) GetByID(ctx context.Context, id uint64) (*models.AppField, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AppField), args.Error(1)
}

func (m *MockFieldRepository) GetByAppID(ctx context.Context, appID uint64) ([]models.AppField, error) {
	args := m.Called(ctx, appID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.AppField), args.Error(1)
}

func (m *MockFieldRepository) GetByAppIDAndCode(ctx context.Context, appID uint64, fieldCode string) (*models.AppField, error) {
	args := m.Called(ctx, appID, fieldCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AppField), args.Error(1)
}

func (m *MockFieldRepository) Update(ctx context.Context, field *models.AppField) error {
	args := m.Called(ctx, field)
	return args.Error(0)
}

func (m *MockFieldRepository) Delete(ctx context.Context, id uint64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockFieldRepository) UpdateOrder(ctx context.Context, items []models.FieldOrderItem) error {
	args := m.Called(ctx, items)
	return args.Error(0)
}

func (m *MockFieldRepository) FieldCodeExists(ctx context.Context, appID uint64, fieldCode string) (bool, error) {
	args := m.Called(ctx, appID, fieldCode)
	return args.Bool(0), args.Error(1)
}

func (m *MockFieldRepository) GetMaxDisplayOrder(ctx context.Context, appID uint64) (int, error) {
	args := m.Called(ctx, appID)
	return args.Int(0), args.Error(1)
}

// MockViewRepository ViewRepositoryInterfaceのモック実装
type MockViewRepository struct {
	mock.Mock
}

func (m *MockViewRepository) Create(ctx context.Context, view *models.AppView) error {
	args := m.Called(ctx, view)
	return args.Error(0)
}

func (m *MockViewRepository) GetByID(ctx context.Context, id uint64) (*models.AppView, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AppView), args.Error(1)
}

func (m *MockViewRepository) GetByAppID(ctx context.Context, appID uint64) ([]models.AppView, error) {
	args := m.Called(ctx, appID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.AppView), args.Error(1)
}

func (m *MockViewRepository) GetDefaultByAppID(ctx context.Context, appID uint64) (*models.AppView, error) {
	args := m.Called(ctx, appID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AppView), args.Error(1)
}

func (m *MockViewRepository) Update(ctx context.Context, view *models.AppView) error {
	args := m.Called(ctx, view)
	return args.Error(0)
}

func (m *MockViewRepository) Delete(ctx context.Context, id uint64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockViewRepository) ClearDefaultByAppID(ctx context.Context, appID uint64) error {
	args := m.Called(ctx, appID)
	return args.Error(0)
}

// MockChartRepository ChartRepositoryInterfaceのモック実装
type MockChartRepository struct {
	mock.Mock
}

func (m *MockChartRepository) Create(ctx context.Context, config *models.ChartConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockChartRepository) GetByID(ctx context.Context, id uint64) (*models.ChartConfig, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ChartConfig), args.Error(1)
}

func (m *MockChartRepository) GetByAppID(ctx context.Context, appID uint64) ([]models.ChartConfig, error) {
	args := m.Called(ctx, appID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ChartConfig), args.Error(1)
}

func (m *MockChartRepository) Update(ctx context.Context, config *models.ChartConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockChartRepository) Delete(ctx context.Context, id uint64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockDynamicQueryExecutor DynamicQueryExecutorInterfaceのモック実装
type MockDynamicQueryExecutor struct {
	mock.Mock
}

func (m *MockDynamicQueryExecutor) CreateTable(ctx context.Context, tableName string, fields []models.AppField) error {
	args := m.Called(ctx, tableName, fields)
	return args.Error(0)
}

func (m *MockDynamicQueryExecutor) DropTable(ctx context.Context, tableName string) error {
	args := m.Called(ctx, tableName)
	return args.Error(0)
}

func (m *MockDynamicQueryExecutor) AddColumn(ctx context.Context, tableName string, field *models.AppField) error {
	args := m.Called(ctx, tableName, field)
	return args.Error(0)
}

func (m *MockDynamicQueryExecutor) DropColumn(ctx context.Context, tableName, columnName string) error {
	args := m.Called(ctx, tableName, columnName)
	return args.Error(0)
}

func (m *MockDynamicQueryExecutor) InsertRecord(ctx context.Context, tableName string, data models.RecordData, userID uint64) (uint64, error) {
	args := m.Called(ctx, tableName, data, userID)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockDynamicQueryExecutor) UpdateRecord(ctx context.Context, tableName string, recordID uint64, data models.RecordData) error {
	args := m.Called(ctx, tableName, recordID, data)
	return args.Error(0)
}

func (m *MockDynamicQueryExecutor) DeleteRecord(ctx context.Context, tableName string, recordID uint64) error {
	args := m.Called(ctx, tableName, recordID)
	return args.Error(0)
}

func (m *MockDynamicQueryExecutor) DeleteRecords(ctx context.Context, tableName string, recordIDs []uint64) error {
	args := m.Called(ctx, tableName, recordIDs)
	return args.Error(0)
}

func (m *MockDynamicQueryExecutor) GetRecords(ctx context.Context, tableName string, fields []models.AppField, opts repositories.RecordQueryOptions) ([]models.RecordResponse, int64, error) {
	args := m.Called(ctx, tableName, fields, opts)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.RecordResponse), args.Get(1).(int64), args.Error(2)
}

func (m *MockDynamicQueryExecutor) GetRecordByID(ctx context.Context, tableName string, fields []models.AppField, recordID uint64) (*models.RecordResponse, error) {
	args := m.Called(ctx, tableName, fields, recordID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RecordResponse), args.Error(1)
}

func (m *MockDynamicQueryExecutor) GetAggregatedData(ctx context.Context, tableName string, req *models.ChartDataRequest) (*models.ChartDataResponse, error) {
	args := m.Called(ctx, tableName, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ChartDataResponse), args.Error(1)
}

func (m *MockDynamicQueryExecutor) CountRecords(ctx context.Context, tableName string) (int64, error) {
	args := m.Called(ctx, tableName)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockDynamicQueryExecutor) CountTodaysUpdates(ctx context.Context, tableName string) (int64, error) {
	args := m.Called(ctx, tableName)
	return args.Get(0).(int64), args.Error(1)
}
