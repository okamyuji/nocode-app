package services

import (
	"context"
	"errors"
	"time"

	"nocode-app/backend/internal/models"
	"nocode-app/backend/internal/repositories"
)

// ビュー関連エラー
var (
	ErrViewNotFound = errors.New("ビューが見つかりません")
)

// ViewService ビュー操作を処理する構造体
type ViewService struct {
	viewRepo repositories.ViewRepositoryInterface
	appRepo  repositories.AppRepositoryInterface
}

// NewViewService 新しいViewServiceを作成する
func NewViewService(viewRepo repositories.ViewRepositoryInterface, appRepo repositories.AppRepositoryInterface) *ViewService {
	return &ViewService{
		viewRepo: viewRepo,
		appRepo:  appRepo,
	}
}

// GetViews アプリの全ビューを取得する
func (s *ViewService) GetViews(ctx context.Context, appID uint64) ([]models.ViewResponse, error) {
	views, err := s.viewRepo.GetByAppID(ctx, appID)
	if err != nil {
		return nil, err
	}

	responses := make([]models.ViewResponse, len(views))
	for i, view := range views {
		responses[i] = *view.ToResponse()
	}

	return responses, nil
}

// CreateView 新しいビューを作成する
func (s *ViewService) CreateView(ctx context.Context, appID uint64, req *models.CreateViewRequest) (*models.ViewResponse, error) {
	// アプリの存在確認
	app, err := s.appRepo.GetByID(ctx, appID)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, ErrAppNotFound
	}

	now := time.Now()
	view := &models.AppView{
		AppID:     appID,
		Name:      req.Name,
		ViewType:  req.ViewType,
		Config:    req.Config,
		IsDefault: req.IsDefault,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// デフォルトに設定する場合は他のデフォルトをクリア
	if view.IsDefault {
		if err := s.viewRepo.ClearDefaultByAppID(ctx, appID); err != nil {
			return nil, err
		}
	}

	if err := s.viewRepo.Create(ctx, view); err != nil {
		return nil, err
	}

	return view.ToResponse(), nil
}

// UpdateView ビューを更新する
func (s *ViewService) UpdateView(ctx context.Context, viewID uint64, req *models.UpdateViewRequest) (*models.ViewResponse, error) {
	view, err := s.viewRepo.GetByID(ctx, viewID)
	if err != nil {
		return nil, err
	}
	if view == nil {
		return nil, ErrViewNotFound
	}

	// フィールドを更新
	if req.Name != "" {
		view.Name = req.Name
	}
	if req.Config != nil {
		view.Config = req.Config
	}
	if req.IsDefault != nil {
		// デフォルトに設定する場合は先に他のデフォルトをクリア
		if *req.IsDefault {
			if err := s.viewRepo.ClearDefaultByAppID(ctx, view.AppID); err != nil {
				return nil, err
			}
		}
		view.IsDefault = *req.IsDefault
	}
	view.UpdatedAt = time.Now()

	if err := s.viewRepo.Update(ctx, view); err != nil {
		return nil, err
	}

	return view.ToResponse(), nil
}

// DeleteView ビューを削除する
func (s *ViewService) DeleteView(ctx context.Context, viewID uint64) error {
	view, err := s.viewRepo.GetByID(ctx, viewID)
	if err != nil {
		return err
	}
	if view == nil {
		return ErrViewNotFound
	}

	return s.viewRepo.Delete(ctx, viewID)
}
