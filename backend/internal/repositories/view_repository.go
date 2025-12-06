package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/uptrace/bun"

	"nocode-app/backend/internal/models"
)

// ViewRepository ビューデータベース操作を処理する構造体
type ViewRepository struct {
	db *bun.DB
}

// NewViewRepository 新しいViewRepositoryを作成する
func NewViewRepository(db *bun.DB) *ViewRepository {
	return &ViewRepository{db: db}
}

// Create 新しいビューを作成する
func (r *ViewRepository) Create(ctx context.Context, view *models.AppView) error {
	_, err := r.db.NewInsert().
		Model(view).
		Exec(ctx)
	return err
}

// GetByID IDでビューを取得する
func (r *ViewRepository) GetByID(ctx context.Context, id uint64) (*models.AppView, error) {
	view := new(models.AppView)
	err := r.db.NewSelect().
		Model(view).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return view, nil
}

// GetByAppID アプリの全ビューを取得する
func (r *ViewRepository) GetByAppID(ctx context.Context, appID uint64) ([]models.AppView, error) {
	var views []models.AppView
	err := r.db.NewSelect().
		Model(&views).
		Where("app_id = ?", appID).
		Order("created_at ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return views, nil
}

// GetDefaultByAppID アプリのデフォルトビューを取得する
func (r *ViewRepository) GetDefaultByAppID(ctx context.Context, appID uint64) (*models.AppView, error) {
	view := new(models.AppView)
	err := r.db.NewSelect().
		Model(view).
		Where("app_id = ? AND is_default = ?", appID, true).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return view, nil
}

// Update ビューを更新する
func (r *ViewRepository) Update(ctx context.Context, view *models.AppView) error {
	_, err := r.db.NewUpdate().
		Model(view).
		WherePK().
		Exec(ctx)
	return err
}

// Delete ビューを削除する
func (r *ViewRepository) Delete(ctx context.Context, id uint64) error {
	_, err := r.db.NewDelete().
		Model((*models.AppView)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

// ClearDefaultByAppID アプリの全ビューのデフォルトフラグをクリアする
func (r *ViewRepository) ClearDefaultByAppID(ctx context.Context, appID uint64) error {
	_, err := r.db.NewUpdate().
		Model((*models.AppView)(nil)).
		Set("is_default = ?", false).
		Where("app_id = ?", appID).
		Exec(ctx)
	return err
}
