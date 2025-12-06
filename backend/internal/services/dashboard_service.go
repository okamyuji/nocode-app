package services

import (
	"context"

	"nocode-app/backend/internal/models"
	"nocode-app/backend/internal/repositories"
)

// DashboardServiceInterface ダッシュボード操作のインターフェースを定義
type DashboardServiceInterface interface {
	GetStats(ctx context.Context) (*models.DashboardStats, error)
}

// DashboardService DashboardServiceInterfaceを実装する構造体
type DashboardService struct {
	userRepo     repositories.UserRepositoryInterface
	appRepo      repositories.AppRepositoryInterface
	dynamicQuery repositories.DynamicQueryExecutorInterface
}

// NewDashboardService 新しいDashboardServiceを作成する
func NewDashboardService(
	userRepo repositories.UserRepositoryInterface,
	appRepo repositories.AppRepositoryInterface,
	dynamicQuery repositories.DynamicQueryExecutorInterface,
) *DashboardService {
	return &DashboardService{
		userRepo:     userRepo,
		appRepo:      appRepo,
		dynamicQuery: dynamicQuery,
	}
}

// GetStats ダッシュボード統計を返す
func (s *DashboardService) GetStats(ctx context.Context) (*models.DashboardStats, error) {
	stats := &models.DashboardStats{}

	// ユーザー数を取得
	userCount, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.UserCount = userCount

	// 全アプリのテーブル名を取得
	tableNames, err := s.appRepo.GetAllTableNames(ctx)
	if err != nil {
		return nil, err
	}
	stats.AppCount = int64(len(tableNames))

	// 全テーブルからレコード数と本日の更新数を集計
	var totalRecords int64
	var todaysUpdates int64

	for _, tableName := range tableNames {
		recordCount, countErr := s.dynamicQuery.CountRecords(ctx, tableName)
		if countErr != nil {
			// まだ存在しない可能性のあるテーブルはスキップ
			continue
		}
		totalRecords += recordCount

		updateCount, updateErr := s.dynamicQuery.CountTodaysUpdates(ctx, tableName)
		if updateErr != nil {
			// 問題のあるテーブルはスキップ
			continue
		}
		todaysUpdates += updateCount
	}

	stats.TotalRecords = totalRecords
	stats.TodaysUpdates = todaysUpdates

	return stats, nil
}

// 実装がインターフェースを満たすことを確認
var _ DashboardServiceInterface = (*DashboardService)(nil)
