package repositories

import (
	"context"

	"github.com/uptrace/bun"

	"nocode-app/backend/internal/models"
)

// DataSourceRepository データソースのデータベース操作を実装する構造体
type DataSourceRepository struct {
	db *bun.DB
}

// NewDataSourceRepository 新しいDataSourceRepositoryを作成する
func NewDataSourceRepository(db *bun.DB) *DataSourceRepository {
	return &DataSourceRepository{db: db}
}

// Create 新しいデータソースを作成する
func (r *DataSourceRepository) Create(ctx context.Context, ds *models.DataSource) error {
	_, err := r.db.NewInsert().Model(ds).Exec(ctx)
	return err
}

// GetByID IDでデータソースを取得する
func (r *DataSourceRepository) GetByID(ctx context.Context, id uint64) (*models.DataSource, error) {
	ds := new(models.DataSource)
	err := r.db.NewSelect().Model(ds).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return ds, nil
}

// GetByName 名前でデータソースを取得する
func (r *DataSourceRepository) GetByName(ctx context.Context, name string) (*models.DataSource, error) {
	ds := new(models.DataSource)
	err := r.db.NewSelect().Model(ds).Where("name = ?", name).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return ds, nil
}

// GetAll ページネーション付きで全データソースを取得する
func (r *DataSourceRepository) GetAll(ctx context.Context, page, limit int) ([]models.DataSource, int64, error) {
	var dataSources []models.DataSource
	offset := (page - 1) * limit

	count, err := r.db.NewSelect().Model(&dataSources).Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	err = r.db.NewSelect().
		Model(&dataSources).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(ctx)
	if err != nil {
		return nil, 0, err
	}

	return dataSources, int64(count), nil
}

// Update データソースを更新する
func (r *DataSourceRepository) Update(ctx context.Context, ds *models.DataSource) error {
	_, err := r.db.NewUpdate().Model(ds).WherePK().Exec(ctx)
	return err
}

// Delete データソースを削除する
func (r *DataSourceRepository) Delete(ctx context.Context, id uint64) error {
	_, err := r.db.NewDelete().Model((*models.DataSource)(nil)).Where("id = ?", id).Exec(ctx)
	return err
}

// NameExists 指定した名前のデータソースが存在するかチェックする
func (r *DataSourceRepository) NameExists(ctx context.Context, name string) (bool, error) {
	count, err := r.db.NewSelect().
		Model((*models.DataSource)(nil)).
		Where("name = ?", name).
		Count(ctx)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// NameExistsExcludingDataSource 指定した名前のデータソースが存在するかチェックする（特定IDを除く）
func (r *DataSourceRepository) NameExistsExcludingDataSource(ctx context.Context, name string, excludeID uint64) (bool, error) {
	count, err := r.db.NewSelect().
		Model((*models.DataSource)(nil)).
		Where("name = ?", name).
		Where("id != ?", excludeID).
		Count(ctx)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
