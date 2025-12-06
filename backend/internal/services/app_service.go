package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"nocode-app/backend/internal/models"
	"nocode-app/backend/internal/repositories"
)

// アプリ関連エラー
var (
	ErrAppNotFound = errors.New("アプリが見つかりません")
)

// AppService アプリ操作を処理する構造体
type AppService struct {
	appRepo      repositories.AppRepositoryInterface
	fieldRepo    repositories.FieldRepositoryInterface
	dynamicQuery repositories.DynamicQueryExecutorInterface
}

// NewAppService 新しいAppServiceを作成する
func NewAppService(
	appRepo repositories.AppRepositoryInterface,
	fieldRepo repositories.FieldRepositoryInterface,
	dynamicQuery repositories.DynamicQueryExecutorInterface,
) *AppService {
	return &AppService{
		appRepo:      appRepo,
		fieldRepo:    fieldRepo,
		dynamicQuery: dynamicQuery,
	}
}

// CreateApp 新しいアプリをフィールドと動的テーブル付きで作成する
func (s *AppService) CreateApp(ctx context.Context, userID uint64, req *models.CreateAppRequest) (*models.AppResponse, error) {
	now := time.Now()

	// アプリを作成
	app := &models.App{
		Name:        req.Name,
		Description: req.Description,
		Icon:        req.Icon,
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

	// アプリIDに基づいてテーブル名を設定
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

	// 動的テーブルを削除
	if err := s.dynamicQuery.DropTable(ctx, app.TableName); err != nil {
		return err
	}

	// アプリを削除（カスケードでフィールドとビューも削除される）
	return s.appRepo.Delete(ctx, appID)
}
