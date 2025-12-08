package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/uptrace/bun"

	"nocode-app/backend/internal/models"
)

// DashboardWidgetRepository ダッシュボードウィジェットのデータベース操作を実装
type DashboardWidgetRepository struct {
	db *bun.DB
}

// NewDashboardWidgetRepository 新しいDashboardWidgetRepositoryを作成
func NewDashboardWidgetRepository(db *bun.DB) *DashboardWidgetRepository {
	return &DashboardWidgetRepository{db: db}
}

// Create 新しいダッシュボードウィジェットを作成
func (r *DashboardWidgetRepository) Create(ctx context.Context, widget *models.DashboardWidget) error {
	_, err := r.db.NewInsert().Model(widget).Exec(ctx)
	if err != nil {
		return fmt.Errorf("ダッシュボードウィジェットの作成に失敗しました: %w", err)
	}
	return nil
}

// GetByID IDでダッシュボードウィジェットを取得
func (r *DashboardWidgetRepository) GetByID(ctx context.Context, id uint64) (*models.DashboardWidget, error) {
	widget := new(models.DashboardWidget)
	err := r.db.NewSelect().
		Model(widget).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("ダッシュボードウィジェットの取得に失敗しました: %w", err)
	}
	return widget, nil
}

// GetByUserID ユーザーIDでダッシュボードウィジェット一覧を取得
func (r *DashboardWidgetRepository) GetByUserID(ctx context.Context, userID uint64) ([]models.DashboardWidget, error) {
	var widgets []models.DashboardWidget
	err := r.db.NewSelect().
		Model(&widgets).
		Where("user_id = ?", userID).
		Order("display_order ASC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("ダッシュボードウィジェット一覧の取得に失敗しました: %w", err)
	}
	return widgets, nil
}

// GetByUserIDWithApps ユーザーIDでダッシュボードウィジェット一覧をアプリ情報付きで取得
func (r *DashboardWidgetRepository) GetByUserIDWithApps(ctx context.Context, userID uint64) ([]models.DashboardWidget, error) {
	var widgets []models.DashboardWidget
	err := r.db.NewSelect().
		Model(&widgets).
		Relation("App").
		Relation("App.Fields").
		Where("dw.user_id = ?", userID).
		Order("dw.display_order ASC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("ダッシュボードウィジェット一覧の取得に失敗しました: %w", err)
	}
	return widgets, nil
}

// GetByUserIDAndAppID ユーザーIDとアプリIDでダッシュボードウィジェットを取得
func (r *DashboardWidgetRepository) GetByUserIDAndAppID(ctx context.Context, userID, appID uint64) (*models.DashboardWidget, error) {
	widget := new(models.DashboardWidget)
	err := r.db.NewSelect().
		Model(widget).
		Where("user_id = ? AND app_id = ?", userID, appID).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("ダッシュボードウィジェットの取得に失敗しました: %w", err)
	}
	return widget, nil
}

// GetVisibleByUserID ユーザーIDで表示中のダッシュボードウィジェット一覧を取得
func (r *DashboardWidgetRepository) GetVisibleByUserID(ctx context.Context, userID uint64) ([]models.DashboardWidget, error) {
	var widgets []models.DashboardWidget
	err := r.db.NewSelect().
		Model(&widgets).
		Relation("App").
		Relation("App.Fields").
		Where("dw.user_id = ? AND dw.is_visible = ?", userID, true).
		Order("dw.display_order ASC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("表示中ダッシュボードウィジェット一覧の取得に失敗しました: %w", err)
	}
	return widgets, nil
}

// Update ダッシュボードウィジェットを更新
func (r *DashboardWidgetRepository) Update(ctx context.Context, widget *models.DashboardWidget) error {
	_, err := r.db.NewUpdate().
		Model(widget).
		WherePK().
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("ダッシュボードウィジェットの更新に失敗しました: %w", err)
	}
	return nil
}

// Delete ダッシュボードウィジェットを削除
func (r *DashboardWidgetRepository) Delete(ctx context.Context, id uint64) error {
	_, err := r.db.NewDelete().
		Model((*models.DashboardWidget)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("ダッシュボードウィジェットの削除に失敗しました: %w", err)
	}
	return nil
}

// DeleteByUserIDAndAppID ユーザーIDとアプリIDでダッシュボードウィジェットを削除
func (r *DashboardWidgetRepository) DeleteByUserIDAndAppID(ctx context.Context, userID, appID uint64) error {
	_, err := r.db.NewDelete().
		Model((*models.DashboardWidget)(nil)).
		Where("user_id = ? AND app_id = ?", userID, appID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("ダッシュボードウィジェットの削除に失敗しました: %w", err)
	}
	return nil
}

// UpdateDisplayOrders ウィジェットの表示順序を一括更新
func (r *DashboardWidgetRepository) UpdateDisplayOrders(ctx context.Context, userID uint64, widgetIDs []uint64) error {
	if len(widgetIDs) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("トランザクションの開始に失敗しました: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	for i, widgetID := range widgetIDs {
		_, err := tx.NewUpdate().
			Model((*models.DashboardWidget)(nil)).
			Set("display_order = ?", i).
			Where("id = ? AND user_id = ?", widgetID, userID).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("表示順序の更新に失敗しました (widget_id=%d): %w", widgetID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("トランザクションのコミットに失敗しました: %w", err)
	}

	return nil
}

// GetMaxDisplayOrder ユーザーの最大表示順序を取得
func (r *DashboardWidgetRepository) GetMaxDisplayOrder(ctx context.Context, userID uint64) (int, error) {
	var maxOrder int
	err := r.db.NewSelect().
		Model((*models.DashboardWidget)(nil)).
		ColumnExpr("COALESCE(MAX(display_order), -1)").
		Where("user_id = ?", userID).
		Scan(ctx, &maxOrder)
	if err != nil {
		return 0, fmt.Errorf("最大表示順序の取得に失敗しました: %w", err)
	}
	return maxOrder, nil
}

// Exists ユーザーIDとアプリIDの組み合わせが存在するかチェック
func (r *DashboardWidgetRepository) Exists(ctx context.Context, userID, appID uint64) (bool, error) {
	exists, err := r.db.NewSelect().
		Model((*models.DashboardWidget)(nil)).
		Where("user_id = ? AND app_id = ?", userID, appID).
		Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("ダッシュボードウィジェットの存在確認に失敗しました: %w", err)
	}
	return exists, nil
}
