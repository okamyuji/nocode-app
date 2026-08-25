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
	"nocode-app/backend/internal/utils"
)

// encryptTestPassword テスト用にパスワードを暗号化する
func encryptTestPassword(t *testing.T, password string) string {
	t.Helper()
	encrypted, err := utils.Encrypt(password)
	require.NoError(t, err)
	return encrypted
}

func TestChartService_GetChartData(t *testing.T) {
	ctx := context.Background()

	t.Run("successful get chart data for internal app", func(t *testing.T) {
		mockChartRepo := new(mocks.MockChartRepository)
		mockAppRepo := new(mocks.MockAppRepository)
		mockFieldRepo := new(mocks.MockFieldRepository)
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

		service := services.NewChartService(mockChartRepo, mockAppRepo, mockFieldRepo, mockDynamicQuery, mockDSRepo, mockExternalQuery)

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

	t.Run("successful get chart data for external app with Japanese column names", func(t *testing.T) {
		setupEncryption(t)

		mockChartRepo := new(mocks.MockChartRepository)
		mockAppRepo := new(mocks.MockAppRepository)
		mockFieldRepo := new(mocks.MockFieldRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)
		mockDSRepo := new(mocks.MockDataSourceRepository)
		mockExternalQuery := new(mocks.MockExternalQueryExecutor)

		dataSourceID := uint64(1)
		sourceTableName := "顧客マスタ"
		app := &models.App{
			ID:              2,
			IsExternal:      true,
			DataSourceID:    &dataSourceID,
			SourceTableName: &sourceTableName,
		}

		// 日本語カラム名をsource_column_nameに持つフィールド
		processNameCol := "プロセス名"
		amountCol := "金額"
		fields := []models.AppField{
			{
				ID:               1,
				AppID:            2,
				FieldCode:        "field_1",
				FieldName:        "プロセス名",
				FieldType:        "text",
				SourceColumnName: &processNameCol,
			},
			{
				ID:               2,
				AppID:            2,
				FieldCode:        "field_2",
				FieldName:        "金額",
				FieldType:        "number",
				SourceColumnName: &amountCol,
			},
		}

		// パスワードを暗号化
		encryptedPassword := encryptTestPassword(t, "testpassword")

		dataSource := &models.DataSource{
			ID:                1,
			Name:              "Oracle DB",
			DBType:            models.DBTypeOracle,
			Host:              "localhost",
			Port:              1521,
			DatabaseName:      "ORCL",
			Username:          "testuser",
			EncryptedPassword: encryptedPassword,
		}

		chartData := &models.ChartDataResponse{
			Labels: []string{"初期活動", "提案・見積", "クローズ"},
			Datasets: []models.ChartDataset{
				{Label: "件数", Data: []float64{5, 10, 3}},
			},
		}

		mockAppRepo.On("GetByID", ctx, uint64(2)).Return(app, nil)
		mockDSRepo.On("GetByID", ctx, uint64(1)).Return(dataSource, nil)
		mockFieldRepo.On("GetByAppID", ctx, uint64(2)).Return(fields, nil)
		mockExternalQuery.On("GetAggregatedData", ctx, dataSource, "testpassword", sourceTableName, fields, mock.AnythingOfType("*models.ChartDataRequest")).Return(chartData, nil)

		service := services.NewChartService(mockChartRepo, mockAppRepo, mockFieldRepo, mockDynamicQuery, mockDSRepo, mockExternalQuery)

		req := &models.ChartDataRequest{
			ChartType: "bar",
			XAxis:     models.ChartAxis{Field: "field_1", Label: "プロセス名"},
			YAxis:     models.ChartAxis{Aggregation: "count", Label: "件数"},
		}

		resp, err := service.GetChartData(ctx, 2, req)
		require.NoError(t, err)
		assert.Equal(t, []string{"初期活動", "提案・見積", "クローズ"}, resp.Labels)
		assert.Len(t, resp.Datasets, 1)
		assert.Equal(t, "件数", resp.Datasets[0].Label)

		mockAppRepo.AssertExpectations(t)
		mockFieldRepo.AssertExpectations(t)
		mockDSRepo.AssertExpectations(t)
		mockExternalQuery.AssertExpectations(t)
	})

	t.Run("successful get chart data for external app with sum aggregation", func(t *testing.T) {
		setupEncryption(t)

		mockChartRepo := new(mocks.MockChartRepository)
		mockAppRepo := new(mocks.MockAppRepository)
		mockFieldRepo := new(mocks.MockFieldRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)
		mockDSRepo := new(mocks.MockDataSourceRepository)
		mockExternalQuery := new(mocks.MockExternalQueryExecutor)

		dataSourceID := uint64(1)
		sourceTableName := "売上テーブル"
		app := &models.App{
			ID:              3,
			IsExternal:      true,
			DataSourceID:    &dataSourceID,
			SourceTableName: &sourceTableName,
		}

		categoryCol := "カテゴリ"
		amountCol := "売上金額"
		fields := []models.AppField{
			{
				ID:               1,
				AppID:            3,
				FieldCode:        "category",
				FieldName:        "カテゴリ",
				FieldType:        "text",
				SourceColumnName: &categoryCol,
			},
			{
				ID:               2,
				AppID:            3,
				FieldCode:        "amount",
				FieldName:        "売上金額",
				FieldType:        "number",
				SourceColumnName: &amountCol,
			},
		}

		// パスワードを暗号化
		encryptedPassword := encryptTestPassword(t, "mysqlpassword")

		dataSource := &models.DataSource{
			ID:                1,
			Name:              "MySQL DB",
			DBType:            models.DBTypeMySQL,
			Host:              "localhost",
			Port:              3306,
			DatabaseName:      "testdb",
			Username:          "testuser",
			EncryptedPassword: encryptedPassword,
		}

		chartData := &models.ChartDataResponse{
			Labels: []string{"食品", "家電", "書籍"},
			Datasets: []models.ChartDataset{
				{Label: "売上金額合計", Data: []float64{150000, 300000, 50000}},
			},
		}

		mockAppRepo.On("GetByID", ctx, uint64(3)).Return(app, nil)
		mockDSRepo.On("GetByID", ctx, uint64(1)).Return(dataSource, nil)
		mockFieldRepo.On("GetByAppID", ctx, uint64(3)).Return(fields, nil)
		mockExternalQuery.On("GetAggregatedData", ctx, dataSource, "mysqlpassword", sourceTableName, fields, mock.AnythingOfType("*models.ChartDataRequest")).Return(chartData, nil)

		service := services.NewChartService(mockChartRepo, mockAppRepo, mockFieldRepo, mockDynamicQuery, mockDSRepo, mockExternalQuery)

		req := &models.ChartDataRequest{
			ChartType: "bar",
			XAxis:     models.ChartAxis{Field: "category", Label: "カテゴリ"},
			YAxis:     models.ChartAxis{Field: "amount", Aggregation: "sum", Label: "売上金額合計"},
		}

		resp, err := service.GetChartData(ctx, 3, req)
		require.NoError(t, err)
		assert.Equal(t, []string{"食品", "家電", "書籍"}, resp.Labels)
		assert.Len(t, resp.Datasets, 1)
		assert.Equal(t, "売上金額合計", resp.Datasets[0].Label)
		assert.Equal(t, []float64{150000, 300000, 50000}, resp.Datasets[0].Data)

		mockAppRepo.AssertExpectations(t)
		mockFieldRepo.AssertExpectations(t)
		mockDSRepo.AssertExpectations(t)
		mockExternalQuery.AssertExpectations(t)
	})

	t.Run("app not found", func(t *testing.T) {
		mockChartRepo := new(mocks.MockChartRepository)
		mockAppRepo := new(mocks.MockAppRepository)
		mockFieldRepo := new(mocks.MockFieldRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)
		mockDSRepo := new(mocks.MockDataSourceRepository)
		mockExternalQuery := new(mocks.MockExternalQueryExecutor)

		mockAppRepo.On("GetByID", ctx, uint64(999)).Return(nil, nil)

		service := services.NewChartService(mockChartRepo, mockAppRepo, mockFieldRepo, mockDynamicQuery, mockDSRepo, mockExternalQuery)

		req := &models.ChartDataRequest{}

		_, err := service.GetChartData(ctx, 999, req)
		assert.ErrorIs(t, err, services.ErrAppNotFound)

		mockAppRepo.AssertExpectations(t)
	})

	t.Run("data source not found for external app", func(t *testing.T) {
		setupEncryption(t)

		mockChartRepo := new(mocks.MockChartRepository)
		mockAppRepo := new(mocks.MockAppRepository)
		mockFieldRepo := new(mocks.MockFieldRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)
		mockDSRepo := new(mocks.MockDataSourceRepository)
		mockExternalQuery := new(mocks.MockExternalQueryExecutor)

		dataSourceID := uint64(999)
		sourceTableName := "test_table"
		app := &models.App{
			ID:              4,
			IsExternal:      true,
			DataSourceID:    &dataSourceID,
			SourceTableName: &sourceTableName,
		}

		mockAppRepo.On("GetByID", ctx, uint64(4)).Return(app, nil)
		mockDSRepo.On("GetByID", ctx, uint64(999)).Return(nil, nil)

		service := services.NewChartService(mockChartRepo, mockAppRepo, mockFieldRepo, mockDynamicQuery, mockDSRepo, mockExternalQuery)

		req := &models.ChartDataRequest{
			ChartType: "bar",
			XAxis:     models.ChartAxis{Field: "field_1"},
			YAxis:     models.ChartAxis{Aggregation: "count"},
		}

		_, err := service.GetChartData(ctx, 4, req)
		assert.ErrorIs(t, err, services.ErrDataSourceNotFound)

		mockAppRepo.AssertExpectations(t)
		mockDSRepo.AssertExpectations(t)
	})
}

func TestChartService_GetChartConfigs(t *testing.T) {
	ctx := context.Background()

	t.Run("successful get configs", func(t *testing.T) {
		mockChartRepo := new(mocks.MockChartRepository)
		mockAppRepo := new(mocks.MockAppRepository)
		mockFieldRepo := new(mocks.MockFieldRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)
		mockDSRepo := new(mocks.MockDataSourceRepository)
		mockExternalQuery := new(mocks.MockExternalQueryExecutor)

		configs := []models.ChartConfig{
			{ID: 1, AppID: 1, Name: "Chart 1", ChartType: "bar", CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: 2, AppID: 1, Name: "Chart 2", ChartType: "pie", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		}

		mockChartRepo.On("GetByAppID", ctx, uint64(1)).Return(configs, nil)

		service := services.NewChartService(mockChartRepo, mockAppRepo, mockFieldRepo, mockDynamicQuery, mockDSRepo, mockExternalQuery)

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
		mockFieldRepo := new(mocks.MockFieldRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)
		mockDSRepo := new(mocks.MockDataSourceRepository)
		mockExternalQuery := new(mocks.MockExternalQueryExecutor)

		app := &models.App{ID: 1}

		mockAppRepo.On("GetByID", ctx, uint64(1)).Return(app, nil)
		mockChartRepo.On("Create", ctx, mock.AnythingOfType("*models.ChartConfig")).Return(nil).Run(func(args mock.Arguments) {
			config := args.Get(1).(*models.ChartConfig)
			config.ID = 1
		})

		service := services.NewChartService(mockChartRepo, mockAppRepo, mockFieldRepo, mockDynamicQuery, mockDSRepo, mockExternalQuery)

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
		mockFieldRepo := new(mocks.MockFieldRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)
		mockDSRepo := new(mocks.MockDataSourceRepository)
		mockExternalQuery := new(mocks.MockExternalQueryExecutor)

		mockAppRepo.On("GetByID", ctx, uint64(999)).Return(nil, nil)

		service := services.NewChartService(mockChartRepo, mockAppRepo, mockFieldRepo, mockDynamicQuery, mockDSRepo, mockExternalQuery)

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
		mockFieldRepo := new(mocks.MockFieldRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)
		mockDSRepo := new(mocks.MockDataSourceRepository)
		mockExternalQuery := new(mocks.MockExternalQueryExecutor)

		config := &models.ChartConfig{ID: 1, AppID: 1}

		mockChartRepo.On("GetByID", ctx, uint64(1)).Return(config, nil)
		mockChartRepo.On("Delete", ctx, uint64(1)).Return(nil)

		service := services.NewChartService(mockChartRepo, mockAppRepo, mockFieldRepo, mockDynamicQuery, mockDSRepo, mockExternalQuery)

		err := service.DeleteChartConfig(ctx, 1)
		require.NoError(t, err)

		mockChartRepo.AssertExpectations(t)
	})

	t.Run("config not found", func(t *testing.T) {
		mockChartRepo := new(mocks.MockChartRepository)
		mockAppRepo := new(mocks.MockAppRepository)
		mockFieldRepo := new(mocks.MockFieldRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)
		mockDSRepo := new(mocks.MockDataSourceRepository)
		mockExternalQuery := new(mocks.MockExternalQueryExecutor)

		mockChartRepo.On("GetByID", ctx, uint64(999)).Return(nil, nil)

		service := services.NewChartService(mockChartRepo, mockAppRepo, mockFieldRepo, mockDynamicQuery, mockDSRepo, mockExternalQuery)

		err := service.DeleteChartConfig(ctx, 999)
		assert.ErrorIs(t, err, services.ErrChartConfigNotFound)

		mockChartRepo.AssertExpectations(t)
	})
}
