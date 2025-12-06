package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/uptrace/bun"

	"nocode-app/backend/internal/models"
)

// FieldRepository フィールドデータベース操作を処理する構造体
type FieldRepository struct {
	db *bun.DB
}

// NewFieldRepository 新しいFieldRepositoryを作成する
func NewFieldRepository(db *bun.DB) *FieldRepository {
	return &FieldRepository{db: db}
}

// Create 新しいフィールドを作成する
func (r *FieldRepository) Create(ctx context.Context, field *models.AppField) error {
	_, err := r.db.NewInsert().
		Model(field).
		Exec(ctx)
	return err
}

// CreateBatch 複数のフィールドを作成する
func (r *FieldRepository) CreateBatch(ctx context.Context, fields []models.AppField) error {
	if len(fields) == 0 {
		return nil
	}
	_, err := r.db.NewInsert().
		Model(&fields).
		Exec(ctx)
	return err
}

// GetByID IDでフィールドを取得する
func (r *FieldRepository) GetByID(ctx context.Context, id uint64) (*models.AppField, error) {
	field := new(models.AppField)
	err := r.db.NewSelect().
		Model(field).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return field, nil
}

// GetByAppID アプリの全フィールドを取得する
func (r *FieldRepository) GetByAppID(ctx context.Context, appID uint64) ([]models.AppField, error) {
	var fields []models.AppField
	err := r.db.NewSelect().
		Model(&fields).
		Where("app_id = ?", appID).
		Order("display_order ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return fields, nil
}

// GetByAppIDAndCode アプリIDとフィールドコードでフィールドを取得する
func (r *FieldRepository) GetByAppIDAndCode(ctx context.Context, appID uint64, fieldCode string) (*models.AppField, error) {
	field := new(models.AppField)
	err := r.db.NewSelect().
		Model(field).
		Where("app_id = ? AND field_code = ?", appID, fieldCode).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return field, nil
}

// Update フィールドを更新する
func (r *FieldRepository) Update(ctx context.Context, field *models.AppField) error {
	_, err := r.db.NewUpdate().
		Model(field).
		WherePK().
		Exec(ctx)
	return err
}

// Delete フィールドを削除する
func (r *FieldRepository) Delete(ctx context.Context, id uint64) error {
	_, err := r.db.NewDelete().
		Model((*models.AppField)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

// UpdateOrder フィールドの表示順序を更新する
func (r *FieldRepository) UpdateOrder(ctx context.Context, items []models.FieldOrderItem) error {
	for _, item := range items {
		_, err := r.db.NewUpdate().
			Model((*models.AppField)(nil)).
			Set("display_order = ?", item.DisplayOrder).
			Where("id = ?", item.ID).
			Exec(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

// FieldCodeExists アプリ内にフィールドコードが既に存在するか確認する
func (r *FieldRepository) FieldCodeExists(ctx context.Context, appID uint64, fieldCode string) (bool, error) {
	count, err := r.db.NewSelect().
		Model((*models.AppField)(nil)).
		Where("app_id = ? AND field_code = ?", appID, fieldCode).
		Count(ctx)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetMaxDisplayOrder アプリのフィールドの最大表示順序を取得する
func (r *FieldRepository) GetMaxDisplayOrder(ctx context.Context, appID uint64) (int, error) {
	var maxOrder int
	err := r.db.NewSelect().
		Model((*models.AppField)(nil)).
		ColumnExpr("COALESCE(MAX(display_order), 0)").
		Where("app_id = ?", appID).
		Scan(ctx, &maxOrder)
	if err != nil {
		return 0, err
	}
	return maxOrder, nil
}
