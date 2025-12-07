package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"nocode-app/backend/internal/models"
	"nocode-app/backend/internal/repositories"
)

// アプリ関連エラー
var (
	ErrAppNotFound = errors.New("アプリが見つかりません")
)

// AppService アプリ操作を処理する構造体
type AppService struct {
	appRepo        repositories.AppRepositoryInterface
	fieldRepo      repositories.FieldRepositoryInterface
	dynamicQuery   repositories.DynamicQueryExecutorInterface
	dataSourceRepo repositories.DataSourceRepositoryInterface
}

// NewAppService 新しいAppServiceを作成する
func NewAppService(
	appRepo repositories.AppRepositoryInterface,
	fieldRepo repositories.FieldRepositoryInterface,
	dynamicQuery repositories.DynamicQueryExecutorInterface,
	dataSourceRepo repositories.DataSourceRepositoryInterface,
) *AppService {
	return &AppService{
		appRepo:        appRepo,
		fieldRepo:      fieldRepo,
		dynamicQuery:   dynamicQuery,
		dataSourceRepo: dataSourceRepo,
	}
}

// CreateApp 新しいアプリをフィールドと動的テーブル付きで作成する
func (s *AppService) CreateApp(ctx context.Context, userID uint64, req *models.CreateAppRequest) (*models.AppResponse, error) {
	now := time.Now()

	// 一時的なユニークテーブル名を生成（NOT NULL UNIQUE制約を満たすため）
	tempTableName := fmt.Sprintf("temp_%s", uuid.New().String())

	// アプリを作成
	app := &models.App{
		Name:        req.Name,
		Description: req.Description,
		Icon:        req.Icon,
		TableName:   tempTableName,
		CreatedBy:   userID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if app.Icon == "" {
		app.Icon = "default"
	}

	// IDを取得するために先にデータベースにアプリを作成
	if err := s.appRepo.Create(ctx, app); err != nil {
		return nil, err
	}

	// アプリIDに基づいて正式なテーブル名を設定
	app.TableName = fmt.Sprintf("app_data_%d", app.ID)
	if err := s.appRepo.Update(ctx, app); err != nil {
		return nil, err
	}

	// フィールドを作成
	fields := make([]models.AppField, len(req.Fields))
	for i, fieldReq := range req.Fields {
		fields[i] = models.AppField{
			AppID:        app.ID,
			FieldCode:    fieldReq.FieldCode,
			FieldName:    fieldReq.FieldName,
			FieldType:    fieldReq.FieldType,
			Options:      fieldReq.Options,
			Required:     fieldReq.Required,
			DisplayOrder: fieldReq.DisplayOrder,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
	}

	if err := s.fieldRepo.CreateBatch(ctx, fields); err != nil {
		return nil, err
	}

	// 動的テーブルを作成
	if err := s.dynamicQuery.CreateTable(ctx, app.TableName, fields); err != nil {
		return nil, err
	}

	// 作成したアプリをフィールド付きで取得
	createdApp, err := s.appRepo.GetByIDWithFields(ctx, app.ID)
	if err != nil {
		return nil, err
	}
	if createdApp == nil {
		return nil, ErrAppNotFound
	}

	return createdApp.ToResponse(), nil
}

// GetApp IDでアプリを取得する
func (s *AppService) GetApp(ctx context.Context, appID uint64) (*models.AppResponse, error) {
	app, err := s.appRepo.GetByIDWithFields(ctx, appID)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, ErrAppNotFound
	}

	return app.ToResponse(), nil
}

// GetApps ページネーション付きで全アプリを取得する
func (s *AppService) GetApps(ctx context.Context, page, limit int) (*models.AppListResponse, error) {
	apps, total, err := s.appRepo.GetAll(ctx, page, limit)
	if err != nil {
		return nil, err
	}

	appResponses := make([]models.AppResponse, len(apps))
	for i := range apps {
		// 各アプリのフィールド数を取得
		fields, fieldErr := s.fieldRepo.GetByAppID(ctx, apps[i].ID)
		if fieldErr == nil {
			apps[i].Fields = fields
		}
		appResponses[i] = *apps[i].ToResponse()
		// ペイロードサイズを減らすためにレスポンスからフィールドをクリア
		appResponses[i].Fields = nil
	}

	return &models.AppListResponse{
		Apps:       appResponses,
		Pagination: models.NewPagination(page, limit, total),
	}, nil
}

// UpdateApp アプリを更新する
func (s *AppService) UpdateApp(ctx context.Context, appID uint64, req *models.UpdateAppRequest) (*models.AppResponse, error) {
	app, err := s.appRepo.GetByID(ctx, appID)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, ErrAppNotFound
	}

	// フィールドを更新
	if req.Name != "" {
		app.Name = req.Name
	}
	if req.Description != "" {
		app.Description = req.Description
	}
	if req.Icon != "" {
		app.Icon = req.Icon
	}
	app.UpdatedAt = time.Now()

	if err := s.appRepo.Update(ctx, app); err != nil {
		return nil, err
	}

	return app.ToResponse(), nil
}

// DeleteApp アプリとその動的テーブルを削除する
func (s *AppService) DeleteApp(ctx context.Context, appID uint64) error {
	app, err := s.appRepo.GetByID(ctx, appID)
	if err != nil {
		return err
	}
	if app == nil {
		return ErrAppNotFound
	}

	// 外部データソースのアプリは動的テーブルを削除しない
	if !app.IsExternal {
		if err := s.dynamicQuery.DropTable(ctx, app.TableName); err != nil {
			return err
		}
	}

	// アプリを削除（カスケードでフィールドとビューも削除される）
	return s.appRepo.Delete(ctx, appID)
}

// CreateExternalApp 外部データソースからアプリを作成する
func (s *AppService) CreateExternalApp(ctx context.Context, userID uint64, req *models.CreateExternalAppRequest) (*models.AppResponse, error) {
	// データソースの存在確認
	ds, err := s.dataSourceRepo.GetByID(ctx, req.DataSourceID)
	if err != nil {
		return nil, err
	}
	if ds == nil {
		return nil, ErrDataSourceNotFound
	}

	now := time.Now()

	// 一時的なユニークテーブル名を生成（NOT NULL UNIQUE制約を満たすため）
	tempTableName := fmt.Sprintf("temp_%s", uuid.New().String())

	// アプリを作成
	dataSourceID := req.DataSourceID
	sourceTableName := req.SourceTableName
	app := &models.App{
		Name:            req.Name,
		Description:     req.Description,
		Icon:            req.Icon,
		TableName:       tempTableName,
		IsExternal:      true,
		DataSourceID:    &dataSourceID,
		SourceTableName: &sourceTableName,
		CreatedBy:       userID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if app.Icon == "" {
		app.Icon = "database"
	}

	// IDを取得するために先にデータベースにアプリを作成
	if err := s.appRepo.Create(ctx, app); err != nil {
		return nil, err
	}

	// 外部アプリのテーブル名を設定（動的テーブルは作成しない）
	app.TableName = fmt.Sprintf("external_%d", app.ID)
	if err := s.appRepo.Update(ctx, app); err != nil {
		return nil, err
	}

	// フィールドを作成（source_column_nameを設定）
	fields := make([]models.AppField, len(req.Fields))
	for i, fieldReq := range req.Fields {
		sourceColumnName := fieldReq.SourceColumnName
		fields[i] = models.AppField{
			AppID:            app.ID,
			FieldCode:        fieldReq.FieldCode,
			FieldName:        fieldReq.FieldName,
			FieldType:        fieldReq.FieldType,
			SourceColumnName: &sourceColumnName,
			Options:          fieldReq.Options,
			Required:         fieldReq.Required,
			DisplayOrder:     fieldReq.DisplayOrder,
			CreatedAt:        now,
			UpdatedAt:        now,
		}
	}

	if err := s.fieldRepo.CreateBatch(ctx, fields); err != nil {
		return nil, err
	}

	// 作成したアプリをフィールド付きで取得
	createdApp, err := s.appRepo.GetByIDWithFields(ctx, app.ID)
	if err != nil {
		return nil, err
	}
	if createdApp == nil {
		return nil, ErrAppNotFound
	}

	return createdApp.ToResponse(), nil
}
