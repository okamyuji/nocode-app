package services

import (
	"context"
	"errors"
	"time"

	"nocode-app/backend/internal/models"
	"nocode-app/backend/internal/repositories"
)

// フィールド関連エラー
var (
	ErrFieldNotFound    = errors.New("フィールドが見つかりません")
	ErrFieldCodeExists  = errors.New("フィールドコードは既に存在します")
	ErrInvalidFieldCode = errors.New("無効なフィールドコードです")
)

// FieldService フィールド操作を処理する構造体
type FieldService struct {
	fieldRepo    repositories.FieldRepositoryInterface
	appRepo      repositories.AppRepositoryInterface
	dynamicQuery repositories.DynamicQueryExecutorInterface
}

// NewFieldService 新しいFieldServiceを作成する
func NewFieldService(
	fieldRepo repositories.FieldRepositoryInterface,
	appRepo repositories.AppRepositoryInterface,
	dynamicQuery repositories.DynamicQueryExecutorInterface,
) *FieldService {
	return &FieldService{
		fieldRepo:    fieldRepo,
		appRepo:      appRepo,
		dynamicQuery: dynamicQuery,
	}
}

// GetFields アプリの全フィールドを取得する
func (s *FieldService) GetFields(ctx context.Context, appID uint64) ([]models.FieldResponse, error) {
	fields, err := s.fieldRepo.GetByAppID(ctx, appID)
	if err != nil {
		return nil, err
	}

	responses := make([]models.FieldResponse, len(fields))
	for i := range fields {
		responses[i] = *fields[i].ToResponse()
	}

	return responses, nil
}

// CreateField 新しいフィールドを作成し動的テーブルにカラムを追加する
func (s *FieldService) CreateField(ctx context.Context, appID uint64, req *models.CreateFieldRequest) (*models.FieldResponse, error) {
	// テーブル名を取得するためにアプリを取得
	tableName, err := s.appRepo.GetTableName(ctx, appID)
	if err != nil {
		return nil, err
	}
	if tableName == "" {
		return nil, ErrAppNotFound
	}

	// フィールドコードの存在確認
	exists, err := s.fieldRepo.FieldCodeExists(ctx, appID, req.FieldCode)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrFieldCodeExists
	}

	// 指定がない場合は最大表示順序を取得
	displayOrder := req.DisplayOrder
	if displayOrder == 0 {
		maxOrder, err := s.fieldRepo.GetMaxDisplayOrder(ctx, appID)
		if err != nil {
			return nil, err
		}
		displayOrder = maxOrder + 1
	}

	now := time.Now()
	field := &models.AppField{
		AppID:        appID,
		FieldCode:    req.FieldCode,
		FieldName:    req.FieldName,
		FieldType:    req.FieldType,
		Options:      req.Options,
		Required:     req.Required,
		DisplayOrder: displayOrder,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// データベースにフィールドを作成
	if err := s.fieldRepo.Create(ctx, field); err != nil {
		return nil, err
	}

	// 動的テーブルにカラムを追加
	if err := s.dynamicQuery.AddColumn(ctx, tableName, field); err != nil {
		// フィールド作成をロールバック
		_ = s.fieldRepo.Delete(ctx, field.ID)
		return nil, err
	}

	return field.ToResponse(), nil
}

// UpdateField フィールドを更新する
func (s *FieldService) UpdateField(ctx context.Context, fieldID uint64, req *models.UpdateFieldRequest) (*models.FieldResponse, error) {
	field, err := s.fieldRepo.GetByID(ctx, fieldID)
	if err != nil {
		return nil, err
	}
	if field == nil {
		return nil, ErrFieldNotFound
	}

	// フィールドを更新
	if req.FieldName != "" {
		field.FieldName = req.FieldName
	}
	if req.Options != nil {
		field.Options = req.Options
	}
	if req.Required != nil {
		field.Required = *req.Required
	}
	if req.DisplayOrder != nil {
		field.DisplayOrder = *req.DisplayOrder
	}
	field.UpdatedAt = time.Now()

	if err := s.fieldRepo.Update(ctx, field); err != nil {
		return nil, err
	}

	return field.ToResponse(), nil
}

// DeleteField フィールドを削除し動的テーブルからカラムを削除する
func (s *FieldService) DeleteField(ctx context.Context, appID, fieldID uint64) error {
	field, err := s.fieldRepo.GetByID(ctx, fieldID)
	if err != nil {
		return err
	}
	if field == nil {
		return ErrFieldNotFound
	}

	// テーブル名を取得
	tableName, err := s.appRepo.GetTableName(ctx, appID)
	if err != nil {
		return err
	}
	if tableName == "" {
		return ErrAppNotFound
	}

	// 動的テーブルからカラムを削除
	if err := s.dynamicQuery.DropColumn(ctx, tableName, field.FieldCode); err != nil {
		return err
	}

	// フィールドを削除
	return s.fieldRepo.Delete(ctx, fieldID)
}

// UpdateFieldOrder フィールドの表示順序を更新する
func (s *FieldService) UpdateFieldOrder(ctx context.Context, _ uint64, req *models.UpdateFieldOrderRequest) error {
	return s.fieldRepo.UpdateOrder(ctx, req.Fields)
}
