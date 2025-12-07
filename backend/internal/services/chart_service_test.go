package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"nocode-app/backend/internal/models"
	"nocode-app/backend/internal/services"
	"nocode-app/backend/internal/testhelpers/mocks"
)

func TestChartService_GetChartData(t *testing.T) {
	ctx := context.Background()

	t.Run("successful get chart data", func(t *testing.T) {
		mockChartRepo := new(mocks.MockChartRepository)
		mockAppRepo := new(mocks.MockAppRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)
		mockDSRepo := new(mocks.MockDataSourceRepository)
		mockExternalQuery := new(mocks.MockExternalQueryExecutor)

		app := &models.App{
			ID:         1,
			TableName:  "app_data_1",
			IsExternal: false,
		}
		chartData := &models.ChartDataResponse{
			Labels: []string{"A", "B", "C"},
			Datasets: []models.ChartDataset{
				{Label: "Count", Data: []float64{10, 20, 30}},
			},
		}

		mockAppRepo.On("GetByID", ctx, uint64(1)).Return(app, nil)
		mockDynamicQuery.On("GetAggregatedData", ctx, "app_data_1", mock.AnythingOfType("*models.ChartDataRequest")).Return(chartData, nil)

		service := services.NewChartService(mockChartRepo, mockAppRepo, mockDynamicQuery, mockDSRepo, mockExternalQuery)

		req := &models.ChartDataRequest{
			ChartType: "bar",
			XAxis:     models.ChartAxis{Field: "category"},
			YAxis:     models.ChartAxis{Aggregation: "count", Label: "Count"},
		}

		resp, err := service.GetChartData(ctx, 1, req)
		require.NoError(t, err)
		assert.Equal(t, []string{"A", "B", "C"}, resp.Labels)
		assert.Len(t, resp.Datasets, 1)

		mockAppRepo.AssertExpectations(t)
		mockDynamicQuery.AssertExpectations(t)
	})

	t.Run("app not found", func(t *testing.T) {
		mockChartRepo := new(mocks.MockChartRepository)
		mockAppRepo := new(mocks.MockAppRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)
		mockDSRepo := new(mocks.MockDataSourceRepository)
		mockExternalQuery := new(mocks.MockExternalQueryExecutor)

		mockAppRepo.On("GetByID", ctx, uint64(999)).Return(nil, nil)

		service := services.NewChartService(mockChartRepo, mockAppRepo, mockDynamicQuery, mockDSRepo, mockExternalQuery)

		req := &models.ChartDataRequest{}

		_, err := service.GetChartData(ctx, 999, req)
		assert.ErrorIs(t, err, services.ErrAppNotFound)

		mockAppRepo.AssertExpectations(t)
	})
}

func TestChartService_GetChartConfigs(t *testing.T) {
	ctx := context.Background()

	t.Run("successful get configs", func(t *testing.T) {
		mockChartRepo := new(mocks.MockChartRepository)
		mockAppRepo := new(mocks.MockAppRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)
		mockDSRepo := new(mocks.MockDataSourceRepository)
		mockExternalQuery := new(mocks.MockExternalQueryExecutor)

		configs := []models.ChartConfig{
			{ID: 1, AppID: 1, Name: "Chart 1", ChartType: "bar", CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: 2, AppID: 1, Name: "Chart 2", ChartType: "pie", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		}

		mockChartRepo.On("GetByAppID", ctx, uint64(1)).Return(configs, nil)

		service := services.NewChartService(mockChartRepo, mockAppRepo, mockDynamicQuery, mockDSRepo, mockExternalQuery)

		resp, err := service.GetChartConfigs(ctx, 1)
		require.NoError(t, err)
		assert.Len(t, resp, 2)

		mockChartRepo.AssertExpectations(t)
	})
}

func TestChartService_SaveChartConfig(t *testing.T) {
	ctx := context.Background()

	t.Run("create new config", func(t *testing.T) {
		mockChartRepo := new(mocks.MockChartRepository)
		mockAppRepo := new(mocks.MockAppRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)
		mockDSRepo := new(mocks.MockDataSourceRepository)
		mockExternalQuery := new(mocks.MockExternalQueryExecutor)

		app := &models.App{ID: 1}

		mockAppRepo.On("GetByID", ctx, uint64(1)).Return(app, nil)
		mockChartRepo.On("Create", ctx, mock.AnythingOfType("*models.ChartConfig")).Return(nil).Run(func(args mock.Arguments) {
			config := args.Get(1).(*models.ChartConfig)
			config.ID = 1
		})

		service := services.NewChartService(mockChartRepo, mockAppRepo, mockDynamicQuery, mockDSRepo, mockExternalQuery)

		req := &models.SaveChartConfigRequest{
			Name:      "New Chart",
			ChartType: "bar",
			Config: models.ChartDataRequest{
				ChartType: "bar",
				XAxis:     models.ChartAxis{Field: "category"},
				YAxis:     models.ChartAxis{Aggregation: "count"},
			},
		}

		resp, err := service.SaveChartConfig(ctx, 1, 1, req)
		require.NoError(t, err)
		assert.Equal(t, "New Chart", resp.Name)

		mockAppRepo.AssertExpectations(t)
		mockChartRepo.AssertExpectations(t)
	})

	t.Run("app not found", func(t *testing.T) {
		mockChartRepo := new(mocks.MockChartRepository)
		mockAppRepo := new(mocks.MockAppRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)
		mockDSRepo := new(mocks.MockDataSourceRepository)
		mockExternalQuery := new(mocks.MockExternalQueryExecutor)

		mockAppRepo.On("GetByID", ctx, uint64(999)).Return(nil, nil)

		service := services.NewChartService(mockChartRepo, mockAppRepo, mockDynamicQuery, mockDSRepo, mockExternalQuery)

		req := &models.SaveChartConfigRequest{
			Name:      "Test",
			ChartType: "bar",
		}

		_, err := service.SaveChartConfig(ctx, 999, 1, req)
		assert.ErrorIs(t, err, services.ErrAppNotFound)

		mockAppRepo.AssertExpectations(t)
	})
}

func TestChartService_DeleteChartConfig(t *testing.T) {
	ctx := context.Background()

	t.Run("successful delete", func(t *testing.T) {
		mockChartRepo := new(mocks.MockChartRepository)
		mockAppRepo := new(mocks.MockAppRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)
		mockDSRepo := new(mocks.MockDataSourceRepository)
		mockExternalQuery := new(mocks.MockExternalQueryExecutor)

		config := &models.ChartConfig{ID: 1, AppID: 1}

		mockChartRepo.On("GetByID", ctx, uint64(1)).Return(config, nil)
		mockChartRepo.On("Delete", ctx, uint64(1)).Return(nil)

		service := services.NewChartService(mockChartRepo, mockAppRepo, mockDynamicQuery, mockDSRepo, mockExternalQuery)

		err := service.DeleteChartConfig(ctx, 1)
		require.NoError(t, err)

		mockChartRepo.AssertExpectations(t)
	})

	t.Run("config not found", func(t *testing.T) {
		mockChartRepo := new(mocks.MockChartRepository)
		mockAppRepo := new(mocks.MockAppRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)
		mockDSRepo := new(mocks.MockDataSourceRepository)
		mockExternalQuery := new(mocks.MockExternalQueryExecutor)

		mockChartRepo.On("GetByID", ctx, uint64(999)).Return(nil, nil)

		service := services.NewChartService(mockChartRepo, mockAppRepo, mockDynamicQuery, mockDSRepo, mockExternalQuery)

		err := service.DeleteChartConfig(ctx, 999)
		assert.ErrorIs(t, err, services.ErrChartConfigNotFound)

		mockChartRepo.AssertExpectations(t)
	})
}
