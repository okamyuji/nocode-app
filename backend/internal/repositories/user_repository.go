package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/uptrace/bun"

	"nocode-app/backend/internal/models"
)

// UserRepository ユーザーデータベース操作を処理する構造体
type UserRepository struct {
	db *bun.DB
}

// NewUserRepository 新しいUserRepositoryを作成する
func NewUserRepository(db *bun.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create 新しいユーザーを作成する
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	_, err := r.db.NewInsert().
		Model(user).
		Exec(ctx)
	return err
}

// GetByID IDでユーザーを取得する
func (r *UserRepository) GetByID(ctx context.Context, id uint64) (*models.User, error) {
	user := new(models.User)
	err := r.db.NewSelect().
		Model(user).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

// GetByEmail メールアドレスでユーザーを取得する
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	user := new(models.User)
	err := r.db.NewSelect().
		Model(user).
		Where("email = ?", email).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

// Update ユーザーを更新する
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	_, err := r.db.NewUpdate().
		Model(user).
		WherePK().
		Exec(ctx)
	return err
}

// Delete ユーザーを削除する
func (r *UserRepository) Delete(ctx context.Context, id uint64) error {
	_, err := r.db.NewDelete().
		Model((*models.User)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

// EmailExists メールアドレスが既に存在するか確認する
func (r *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	count, err := r.db.NewSelect().
		Model((*models.User)(nil)).
		Where("email = ?", email).
		Count(ctx)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetAll ページネーション付きで全ユーザーを取得する
func (r *UserRepository) GetAll(ctx context.Context, page, limit int) ([]models.User, int64, error) {
	var users []models.User
	offset := (page - 1) * limit

	count, err := r.db.NewSelect().
		Model((*models.User)(nil)).
		Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	err = r.db.NewSelect().
		Model(&users).
		Order("id ASC").
		Limit(limit).
		Offset(offset).
		Scan(ctx)
	if err != nil {
		return nil, 0, err
	}

	return users, int64(count), nil
}

// EmailExistsExcludingUser 指定ユーザー以外でメールアドレスが存在するか確認する
func (r *UserRepository) EmailExistsExcludingUser(ctx context.Context, email string, excludeID uint64) (bool, error) {
	count, err := r.db.NewSelect().
		Model((*models.User)(nil)).
		Where("email = ?", email).
		Where("id != ?", excludeID).
		Count(ctx)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Count ユーザーの総数を返す
func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	count, err := r.db.NewSelect().
		Model((*models.User)(nil)).
		Count(ctx)
	if err != nil {
		return 0, err
	}
	return int64(count), nil
}
