// Package repositories テスト容易性のためのインターフェースを備えたデータベースアクセス層を提供
package repositories

import (
	"context"

	"nocode-app/backend/internal/models"
)

// UserRepositoryInterface ユーザーデータベース操作のインターフェースを定義
type UserRepositoryInterface interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uint64) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetAll(ctx context.Context, page, limit int) ([]models.User, int64, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uint64) error
	EmailExists(ctx context.Context, email string) (bool, error)
	EmailExistsExcludingUser(ctx context.Context, email string, excludeID uint64) (bool, error)
	Count(ctx context.Context) (int64, error)
}

// AppRepositoryInterface アプリデータベース操作のインターフェースを定義
type AppRepositoryInterface interface {
	Create(ctx context.Context, app *models.App) error
	GetByID(ctx context.Context, id uint64) (*models.App, error)
	GetByIDWithFields(ctx context.Context, id uint64) (*models.App, error)
	GetAll(ctx context.Context, page, limit int) ([]models.App, int64, error)
	GetByUserID(ctx context.Context, userID uint64, page, limit int) ([]models.App, int64, error)
	Update(ctx context.Context, app *models.App) error
	Delete(ctx context.Context, id uint64) error
	GetTableName(ctx context.Context, appID uint64) (string, error)
	GetAllTableNames(ctx context.Context) ([]string, error)
}

// FieldRepositoryInterface フィールドデータベース操作のインターフェースを定義
type FieldRepositoryInterface interface {
	Create(ctx context.Context, field *models.AppField) error
	CreateBatch(ctx context.Context, fields []models.AppField) error
	GetByID(ctx context.Context, id uint64) (*models.AppField, error)
	GetByAppID(ctx context.Context, appID uint64) ([]models.AppField, error)
	GetByAppIDAndCode(ctx context.Context, appID uint64, fieldCode string) (*models.AppField, error)
	Update(ctx context.Context, field *models.AppField) error
	Delete(ctx context.Context, id uint64) error
	UpdateOrder(ctx context.Context, items []models.FieldOrderItem) error
	FieldCodeExists(ctx context.Context, appID uint64, fieldCode string) (bool, error)
	GetMaxDisplayOrder(ctx context.Context, appID uint64) (int, error)
}

// ViewRepositoryInterface ビューデータベース操作のインターフェースを定義
type ViewRepositoryInterface interface {
	Create(ctx context.Context, view *models.AppView) error
	GetByID(ctx context.Context, id uint64) (*models.AppView, error)
	GetByAppID(ctx context.Context, appID uint64) ([]models.AppView, error)
	GetDefaultByAppID(ctx context.Context, appID uint64) (*models.AppView, error)
	Update(ctx context.Context, view *models.AppView) error
	Delete(ctx context.Context, id uint64) error
	ClearDefaultByAppID(ctx context.Context, appID uint64) error
}

// ChartRepositoryInterface チャート設定データベース操作のインターフェースを定義
type ChartRepositoryInterface interface {
	Create(ctx context.Context, config *models.ChartConfig) error
	GetByID(ctx context.Context, id uint64) (*models.ChartConfig, error)
	GetByAppID(ctx context.Context, appID uint64) ([]models.ChartConfig, error)
	Update(ctx context.Context, config *models.ChartConfig) error
	Delete(ctx context.Context, id uint64) error
}

// DynamicQueryExecutorInterface 動的テーブル操作のインターフェースを定義
type DynamicQueryExecutorInterface interface {
	CreateTable(ctx context.Context, tableName string, fields []models.AppField) error
	DropTable(ctx context.Context, tableName string) error
	AddColumn(ctx context.Context, tableName string, field *models.AppField) error
	DropColumn(ctx context.Context, tableName, columnName string) error
	InsertRecord(ctx context.Context, tableName string, data models.RecordData, userID uint64) (uint64, error)
	UpdateRecord(ctx context.Context, tableName string, recordID uint64, data models.RecordData) error
	DeleteRecord(ctx context.Context, tableName string, recordID uint64) error
	DeleteRecords(ctx context.Context, tableName string, recordIDs []uint64) error
	GetRecords(ctx context.Context, tableName string, fields []models.AppField, opts RecordQueryOptions) ([]models.RecordResponse, int64, error)
	GetRecordByID(ctx context.Context, tableName string, fields []models.AppField, recordID uint64) (*models.RecordResponse, error)
	GetAggregatedData(ctx context.Context, tableName string, req *models.ChartDataRequest) (*models.ChartDataResponse, error)
	CountRecords(ctx context.Context, tableName string) (int64, error)
	CountTodaysUpdates(ctx context.Context, tableName string) (int64, error)
}

// 実装がインターフェースを満たすことを確認
var (
	_ UserRepositoryInterface       = (*UserRepository)(nil)
	_ AppRepositoryInterface        = (*AppRepository)(nil)
	_ FieldRepositoryInterface      = (*FieldRepository)(nil)
	_ ViewRepositoryInterface       = (*ViewRepository)(nil)
	_ ChartRepositoryInterface      = (*ChartRepository)(nil)
	_ DynamicQueryExecutorInterface = (*DynamicQueryExecutor)(nil)
)
