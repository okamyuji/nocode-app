package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/uptrace/bun"

	"nocode-app/backend/internal/models"
)

// ChartRepository チャート設定データベース操作を処理する構造体
type ChartRepository struct {
	db *bun.DB
}

// NewChartRepository 新しいChartRepositoryを作成する
func NewChartRepository(db *bun.DB) *ChartRepository {
	return &ChartRepository{db: db}
}

// Create 新しいチャート設定を作成する
func (r *ChartRepository) Create(ctx context.Context, config *models.ChartConfig) error {
	_, err := r.db.NewInsert().
		Model(config).
		Exec(ctx)
	return err
}

// GetByID IDでチャート設定を取得する
func (r *ChartRepository) GetByID(ctx context.Context, id uint64) (*models.ChartConfig, error) {
	config := new(models.ChartConfig)
	err := r.db.NewSelect().
		Model(config).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return config, nil
}

// GetByAppID アプリの全チャート設定を取得する
func (r *ChartRepository) GetByAppID(ctx context.Context, appID uint64) ([]models.ChartConfig, error) {
	var configs []models.ChartConfig
	err := r.db.NewSelect().
		Model(&configs).
		Where("app_id = ?", appID).
		Order("created_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return configs, nil
}

// Update チャート設定を更新する
func (r *ChartRepository) Update(ctx context.Context, config *models.ChartConfig) error {
	_, err := r.db.NewUpdate().
		Model(config).
		WherePK().
		Exec(ctx)
	return err
}

// Delete チャート設定を削除する
func (r *ChartRepository) Delete(ctx context.Context, id uint64) error {
	_, err := r.db.NewDelete().
		Model((*models.ChartConfig)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	return err
}
