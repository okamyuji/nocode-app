// Package services テスト容易性のためのインターフェースを備えたビジネスロジック層を提供
package services

import (
	"context"

	"nocode-app/backend/internal/models"
	"nocode-app/backend/internal/repositories"
	"nocode-app/backend/internal/utils"
)

// AuthServiceInterface 認証操作のインターフェースを定義
type AuthServiceInterface interface {
	Register(ctx context.Context, req *models.RegisterRequest) (*models.AuthResponse, error)
	Login(ctx context.Context, req *models.LoginRequest) (*models.AuthResponse, error)
	GetCurrentUser(ctx context.Context, userID uint64) (*models.UserResponse, error)
	RefreshToken(claims *utils.JWTClaims) (string, error)
}

// AppServiceInterface アプリ操作のインターフェースを定義
type AppServiceInterface interface {
	CreateApp(ctx context.Context, userID uint64, req *models.CreateAppRequest) (*models.AppResponse, error)
	GetApp(ctx context.Context, appID uint64) (*models.AppResponse, error)
	GetApps(ctx context.Context, page, limit int) (*models.AppListResponse, error)
	UpdateApp(ctx context.Context, appID uint64, req *models.UpdateAppRequest) (*models.AppResponse, error)
	DeleteApp(ctx context.Context, appID uint64) error
}

// FieldServiceInterface フィールド操作のインターフェースを定義
type FieldServiceInterface interface {
	GetFields(ctx context.Context, appID uint64) ([]models.FieldResponse, error)
	CreateField(ctx context.Context, appID uint64, req *models.CreateFieldRequest) (*models.FieldResponse, error)
	UpdateField(ctx context.Context, fieldID uint64, req *models.UpdateFieldRequest) (*models.FieldResponse, error)
	DeleteField(ctx context.Context, appID, fieldID uint64) error
	UpdateFieldOrder(ctx context.Context, appID uint64, req *models.UpdateFieldOrderRequest) error
}

// RecordServiceInterface レコード操作のインターフェースを定義
type RecordServiceInterface interface {
	GetRecords(ctx context.Context, appID uint64, opts repositories.RecordQueryOptions) (*models.RecordListResponse, error)
	GetRecord(ctx context.Context, appID, recordID uint64) (*models.RecordResponse, error)
	CreateRecord(ctx context.Context, appID, userID uint64, req *models.CreateRecordRequest) (*models.RecordResponse, error)
	UpdateRecord(ctx context.Context, appID, recordID uint64, req *models.UpdateRecordRequest) (*models.RecordResponse, error)
	DeleteRecord(ctx context.Context, appID, recordID uint64) error
	BulkCreateRecords(ctx context.Context, appID, userID uint64, req *models.BulkCreateRecordRequest) ([]models.RecordResponse, error)
	BulkDeleteRecords(ctx context.Context, appID uint64, req *models.BulkDeleteRecordRequest) error
}

// ViewServiceInterface ビュー操作のインターフェースを定義
type ViewServiceInterface interface {
	GetViews(ctx context.Context, appID uint64) ([]models.ViewResponse, error)
	CreateView(ctx context.Context, appID uint64, req *models.CreateViewRequest) (*models.ViewResponse, error)
	UpdateView(ctx context.Context, viewID uint64, req *models.UpdateViewRequest) (*models.ViewResponse, error)
	DeleteView(ctx context.Context, viewID uint64) error
}

// ChartServiceInterface チャート操作のインターフェースを定義
type ChartServiceInterface interface {
	GetChartData(ctx context.Context, appID uint64, req *models.ChartDataRequest) (*models.ChartDataResponse, error)
	GetChartConfigs(ctx context.Context, appID uint64) ([]models.ChartConfig, error)
	SaveChartConfig(ctx context.Context, appID, userID uint64, req *models.SaveChartConfigRequest) (*models.ChartConfig, error)
	DeleteChartConfig(ctx context.Context, configID uint64) error
}

// UserServiceInterface ユーザー管理操作のインターフェースを定義
type UserServiceInterface interface {
	GetUsers(ctx context.Context, callerRole string, page, limit int) (*models.UserListResponse, error)
	GetUser(ctx context.Context, callerRole string, userID uint64) (*models.UserResponse, error)
	CreateUser(ctx context.Context, callerRole string, req *models.CreateUserRequest) (*models.UserResponse, error)
	UpdateUser(ctx context.Context, callerID uint64, callerRole string, userID uint64, req *models.UpdateUserRequest) (*models.UserResponse, error)
	DeleteUser(ctx context.Context, callerID uint64, callerRole string, userID uint64) error
	UpdateProfile(ctx context.Context, userID uint64, req *models.UpdateProfileRequest) (*models.UserResponse, error)
	ChangePassword(ctx context.Context, userID uint64, req *models.ChangePasswordRequest) error
}

// 実装がインターフェースを満たすことを確認
var (
	_ AuthServiceInterface   = (*AuthService)(nil)
	_ AppServiceInterface    = (*AppService)(nil)
	_ FieldServiceInterface  = (*FieldService)(nil)
	_ RecordServiceInterface = (*RecordService)(nil)
	_ ViewServiceInterface   = (*ViewService)(nil)
	_ ChartServiceInterface  = (*ChartService)(nil)
	_ UserServiceInterface   = (*UserService)(nil)
)
