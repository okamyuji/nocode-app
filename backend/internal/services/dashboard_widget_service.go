package services

import (
	"context"
	"errors"

	"nocode-app/backend/internal/models"
	"nocode-app/backend/internal/repositories"
)

// DashboardWidgetService ダッシュボードウィジェットサービスの実装
type DashboardWidgetService struct {
	widgetRepo repositories.DashboardWidgetRepositoryInterface
	appRepo    repositories.AppRepositoryInterface
}

// NewDashboardWidgetService 新しいDashboardWidgetServiceを作成
func NewDashboardWidgetService(
	widgetRepo repositories.DashboardWidgetRepositoryInterface,
	appRepo repositories.AppRepositoryInterface,
) *DashboardWidgetService {
	return &DashboardWidgetService{
		widgetRepo: widgetRepo,
		appRepo:    appRepo,
	}
}

// GetWidgets ユーザーのダッシュボードウィジェット一覧を取得
func (s *DashboardWidgetService) GetWidgets(ctx context.Context, userID uint64) (*models.DashboardWidgetListResponse, error) {
	widgets, err := s.widgetRepo.GetByUserIDWithApps(ctx, userID)
	if err != nil {
		return nil, err
	}

	response := &models.DashboardWidgetListResponse{
		Widgets: make([]models.DashboardWidgetResponse, len(widgets)),
	}
	for i := range widgets {
		response.Widgets[i] = *widgets[i].ToResponse()
	}

	return response, nil
}

// GetVisibleWidgets ユーザーの表示中ダッシュボードウィジェット一覧を取得
func (s *DashboardWidgetService) GetVisibleWidgets(ctx context.Context, userID uint64) (*models.DashboardWidgetListResponse, error) {
	widgets, err := s.widgetRepo.GetVisibleByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	response := &models.DashboardWidgetListResponse{
		Widgets: make([]models.DashboardWidgetResponse, len(widgets)),
	}
	for i := range widgets {
		response.Widgets[i] = *widgets[i].ToResponse()
	}

	return response, nil
}

// CreateWidget 新しいダッシュボードウィジェットを作成
func (s *DashboardWidgetService) CreateWidget(ctx context.Context, userID uint64, req *models.CreateDashboardWidgetRequest) (*models.DashboardWidgetResponse, error) {
	// アプリの存在確認
	app, err := s.appRepo.GetByID(ctx, req.AppID)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, errors.New("指定されたアプリが見つかりません")
	}

	// 既に同じアプリのウィジェットが存在するか確認
	exists, err := s.widgetRepo.Exists(ctx, userID, req.AppID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("このアプリのウィジェットは既に存在します")
	}

	// デフォルトの表示順序を取得
	maxOrder, err := s.widgetRepo.GetMaxDisplayOrder(ctx, userID)
	if err != nil {
		return nil, err
	}

	// ウィジェットを作成
	widget := &models.DashboardWidget{
		UserID:       userID,
		AppID:        req.AppID,
		DisplayOrder: maxOrder + 1,
		ViewType:     models.WidgetViewTypeTable,
		IsVisible:    true,
		WidgetSize:   models.WidgetSizeMedium,
		Config:       req.Config,
	}

	// リクエストから値を設定
	if req.DisplayOrder != nil {
		widget.DisplayOrder = *req.DisplayOrder
	}
	if req.ViewType != "" {
		widget.ViewType = models.WidgetViewType(req.ViewType)
	}
	if req.IsVisible != nil {
		widget.IsVisible = *req.IsVisible
	}
	if req.WidgetSize != "" {
		widget.WidgetSize = models.WidgetSize(req.WidgetSize)
	}

	if err := s.widgetRepo.Create(ctx, widget); err != nil {
		return nil, err
	}

	// 作成したウィジェットをアプリ情報付きで再取得
	createdWidget, err := s.widgetRepo.GetByID(ctx, widget.ID)
	if err != nil {
		return nil, err
	}
	createdWidget.App = app

	return createdWidget.ToResponse(), nil
}

// UpdateWidget ダッシュボードウィジェットを更新
func (s *DashboardWidgetService) UpdateWidget(ctx context.Context, userID, widgetID uint64, req *models.UpdateDashboardWidgetRequest) (*models.DashboardWidgetResponse, error) {
	// ウィジェットの取得と所有権確認
	widget, err := s.widgetRepo.GetByID(ctx, widgetID)
	if err != nil {
		return nil, err
	}
	if widget == nil {
		return nil, errors.New("ウィジェットが見つかりません")
	}
	if widget.UserID != userID {
		return nil, errors.New("このウィジェットを更新する権限がありません")
	}

	// リクエストから値を更新
	if req.DisplayOrder != nil {
		widget.DisplayOrder = *req.DisplayOrder
	}
	if req.ViewType != "" {
		widget.ViewType = models.WidgetViewType(req.ViewType)
	}
	if req.IsVisible != nil {
		widget.IsVisible = *req.IsVisible
	}
	if req.WidgetSize != "" {
		widget.WidgetSize = models.WidgetSize(req.WidgetSize)
	}
	if req.Config != nil {
		widget.Config = req.Config
	}

	if err := s.widgetRepo.Update(ctx, widget); err != nil {
		return nil, err
	}

	// アプリ情報を取得
	app, err := s.appRepo.GetByIDWithFields(ctx, widget.AppID)
	if err != nil {
		return nil, err
	}
	widget.App = app

	return widget.ToResponse(), nil
}

// DeleteWidget ダッシュボードウィジェットを削除
func (s *DashboardWidgetService) DeleteWidget(ctx context.Context, userID, widgetID uint64) error {
	// ウィジェットの取得と所有権確認
	widget, err := s.widgetRepo.GetByID(ctx, widgetID)
	if err != nil {
		return err
	}
	if widget == nil {
		return errors.New("ウィジェットが見つかりません")
	}
	if widget.UserID != userID {
		return errors.New("このウィジェットを削除する権限がありません")
	}

	return s.widgetRepo.Delete(ctx, widgetID)
}

// ReorderWidgets ウィジェットの並び順を更新
func (s *DashboardWidgetService) ReorderWidgets(ctx context.Context, userID uint64, req *models.ReorderWidgetsRequest) error {
	// 指定されたウィジェットIDが全てユーザーのものか確認
	widgets, err := s.widgetRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	widgetIDSet := make(map[uint64]bool)
	for _, w := range widgets {
		widgetIDSet[w.ID] = true
	}

	for _, reqID := range req.WidgetIDs {
		if !widgetIDSet[reqID] {
			return errors.New("無効なウィジェットIDが含まれています")
		}
	}

	return s.widgetRepo.UpdateDisplayOrders(ctx, userID, req.WidgetIDs)
}

// ToggleVisibility ウィジェットの表示/非表示を切り替え
func (s *DashboardWidgetService) ToggleVisibility(ctx context.Context, userID, widgetID uint64) (*models.DashboardWidgetResponse, error) {
	// ウィジェットの取得と所有権確認
	widget, err := s.widgetRepo.GetByID(ctx, widgetID)
	if err != nil {
		return nil, err
	}
	if widget == nil {
		return nil, errors.New("ウィジェットが見つかりません")
	}
	if widget.UserID != userID {
		return nil, errors.New("このウィジェットを更新する権限がありません")
	}

	// 表示状態を切り替え
	widget.IsVisible = !widget.IsVisible

	if err := s.widgetRepo.Update(ctx, widget); err != nil {
		return nil, err
	}

	// アプリ情報を取得
	app, err := s.appRepo.GetByIDWithFields(ctx, widget.AppID)
	if err != nil {
		return nil, err
	}
	widget.App = app

	return widget.ToResponse(), nil
}
