package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/uptrace/bun"

	"nocode-app/backend/internal/models"
)

// AppRepository アプリデータベース操作を処理する構造体
type AppRepository struct {
	db *bun.DB
}

// NewAppRepository 新しいAppRepositoryを作成する
func NewAppRepository(db *bun.DB) *AppRepository {
	return &AppRepository{db: db}
}

// Create 新しいアプリを作成する
func (r *AppRepository) Create(ctx context.Context, app *models.App) error {
	_, err := r.db.NewInsert().
		Model(app).
		Exec(ctx)
	return err
}

// GetByID IDでアプリを取得する
func (r *AppRepository) GetByID(ctx context.Context, id uint64) (*models.App, error) {
	app := new(models.App)
	err := r.db.NewSelect().
		Model(app).
		Where("a.id = ?", id).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return app, nil
}

// GetByIDWithFields IDでアプリをフィールド付きで取得する
func (r *AppRepository) GetByIDWithFields(ctx context.Context, id uint64) (*models.App, error) {
	app := new(models.App)
	err := r.db.NewSelect().
		Model(app).
		Relation("Fields", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("display_order ASC")
		}).
		Where("a.id = ?", id).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return app, nil
}

// GetAll ページネーション付きで全アプリを取得する
func (r *AppRepository) GetAll(ctx context.Context, page, limit int) ([]models.App, int64, error) {
	var apps []models.App
	count, err := r.db.NewSelect().
		Model(&apps).
		Order("created_at DESC").
		Limit(limit).
		Offset((page - 1) * limit).
		ScanAndCount(ctx)
	if err != nil {
		return nil, 0, err
	}
	return apps, int64(count), nil
}

// GetByUserID ユーザーが作成した全アプリを取得する
func (r *AppRepository) GetByUserID(ctx context.Context, userID uint64, page, limit int) ([]models.App, int64, error) {
	var apps []models.App
	count, err := r.db.NewSelect().
		Model(&apps).
		Where("created_by = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset((page - 1) * limit).
		ScanAndCount(ctx)
	if err != nil {
		return nil, 0, err
	}
	return apps, int64(count), nil
}

// Update アプリを更新する
func (r *AppRepository) Update(ctx context.Context, app *models.App) error {
	_, err := r.db.NewUpdate().
		Model(app).
		WherePK().
		Exec(ctx)
	return err
}

// Delete アプリを削除する
func (r *AppRepository) Delete(ctx context.Context, id uint64) error {
	_, err := r.db.NewDelete().
		Model((*models.App)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

// GetTableName アプリのテーブル名を取得する
func (r *AppRepository) GetTableName(ctx context.Context, appID uint64) (string, error) {
	app := new(models.App)
	err := r.db.NewSelect().
		Model(app).
		Column("table_name").
		Where("id = ?", appID).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return app.TableName, nil
}

// GetAllTableNames 全アプリのテーブル名を返す
func (r *AppRepository) GetAllTableNames(ctx context.Context) ([]string, error) {
	var tableNames []string
	err := r.db.NewSelect().
		Model((*models.App)(nil)).
		Column("table_name").
		Scan(ctx, &tableNames)
	if err != nil {
		return nil, err
	}
	return tableNames, nil
}
