package services

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"nocode-app/backend/internal/models"
	"nocode-app/backend/internal/repositories"
)

// チャート関連エラー
var (
	ErrChartConfigNotFound = errors.New("チャート設定が見つかりません")
)

// ChartService チャート操作を処理する構造体
type ChartService struct {
	chartRepo    repositories.ChartRepositoryInterface
	appRepo      repositories.AppRepositoryInterface
	dynamicQuery repositories.DynamicQueryExecutorInterface
}

// NewChartService 新しいChartServiceを作成する
func NewChartService(
	chartRepo repositories.ChartRepositoryInterface,
	appRepo repositories.AppRepositoryInterface,
	dynamicQuery repositories.DynamicQueryExecutorInterface,
) *ChartService {
	return &ChartService{
		chartRepo:    chartRepo,
		appRepo:      appRepo,
		dynamicQuery: dynamicQuery,
	}
}

// GetChartData チャート用の集計データを取得する
func (s *ChartService) GetChartData(ctx context.Context, appID uint64, req *models.ChartDataRequest) (*models.ChartDataResponse, error) {
	// アプリ情報を取得
	tableName, err := s.appRepo.GetTableName(ctx, appID)
	if err != nil {
		return nil, err
	}
	if tableName == "" {
		return nil, ErrAppNotFound
	}

	return s.dynamicQuery.GetAggregatedData(ctx, tableName, req)
}

// GetChartConfigs アプリの全チャート設定を取得する
func (s *ChartService) GetChartConfigs(ctx context.Context, appID uint64) ([]models.ChartConfig, error) {
	return s.chartRepo.GetByAppID(ctx, appID)
}

// SaveChartConfig チャート設定を保存する
func (s *ChartService) SaveChartConfig(ctx context.Context, appID, userID uint64, req *models.SaveChartConfigRequest) (*models.ChartConfig, error) {
	// アプリの存在確認
	app, err := s.appRepo.GetByID(ctx, appID)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, ErrAppNotFound
	}

	// 設定をViewConfig（JSON保存形式）に変換
	configJSON, err := json.Marshal(req.Config)
	if err != nil {
		return nil, err
	}

	var viewConfig models.ViewConfig
	if err := json.Unmarshal(configJSON, &viewConfig); err != nil {
		return nil, err
	}

	now := time.Now()
	config := &models.ChartConfig{
		AppID:     appID,
		Name:      req.Name,
		ChartType: req.ChartType,
		Config:    viewConfig,
		CreatedBy: userID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.chartRepo.Create(ctx, config); err != nil {
		return nil, err
	}

	return config, nil
}

// DeleteChartConfig チャート設定を削除する
func (s *ChartService) DeleteChartConfig(ctx context.Context, configID uint64) error {
	config, err := s.chartRepo.GetByID(ctx, configID)
	if err != nil {
		return err
	}
	if config == nil {
		return ErrChartConfigNotFound
	}

	return s.chartRepo.Delete(ctx, configID)
}
