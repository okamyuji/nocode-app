package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nocode-app/backend/internal/services"
	"nocode-app/backend/internal/testhelpers/mocks"
)

func TestDashboardService_GetStats(t *testing.T) {
	ctx := context.Background()

	t.Run("successful get dashboard stats", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockAppRepo := new(mocks.MockAppRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)

		mockUserRepo.On("Count", ctx).Return(int64(10), nil)
		mockAppRepo.On("GetAllTableNames", ctx).Return([]string{"app_data_1", "app_data_2"}, nil)
		mockDynamicQuery.On("CountRecords", ctx, "app_data_1").Return(int64(100), nil)
		mockDynamicQuery.On("CountRecords", ctx, "app_data_2").Return(int64(50), nil)
		mockDynamicQuery.On("CountTodaysUpdates", ctx, "app_data_1").Return(int64(5), nil)
		mockDynamicQuery.On("CountTodaysUpdates", ctx, "app_data_2").Return(int64(3), nil)

		service := services.NewDashboardService(mockUserRepo, mockAppRepo, mockDynamicQuery)

		stats, err := service.GetStats(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(2), stats.AppCount)
		assert.Equal(t, int64(150), stats.TotalRecords)
		assert.Equal(t, int64(10), stats.UserCount)
		assert.Equal(t, int64(8), stats.TodaysUpdates)

		mockUserRepo.AssertExpectations(t)
		mockAppRepo.AssertExpectations(t)
		mockDynamicQuery.AssertExpectations(t)
	})

	t.Run("no apps", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockAppRepo := new(mocks.MockAppRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)

		mockUserRepo.On("Count", ctx).Return(int64(5), nil)
		mockAppRepo.On("GetAllTableNames", ctx).Return([]string{}, nil)

		service := services.NewDashboardService(mockUserRepo, mockAppRepo, mockDynamicQuery)

		stats, err := service.GetStats(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(0), stats.AppCount)
		assert.Equal(t, int64(0), stats.TotalRecords)
		assert.Equal(t, int64(5), stats.UserCount)
		assert.Equal(t, int64(0), stats.TodaysUpdates)

		mockUserRepo.AssertExpectations(t)
		mockAppRepo.AssertExpectations(t)
	})

	t.Run("user count error", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockAppRepo := new(mocks.MockAppRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)

		mockUserRepo.On("Count", ctx).Return(int64(0), errors.New("db error"))

		service := services.NewDashboardService(mockUserRepo, mockAppRepo, mockDynamicQuery)

		_, err := service.GetStats(ctx)
		assert.Error(t, err)

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("get table names error", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockAppRepo := new(mocks.MockAppRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)

		mockUserRepo.On("Count", ctx).Return(int64(5), nil)
		mockAppRepo.On("GetAllTableNames", ctx).Return(nil, errors.New("db error"))

		service := services.NewDashboardService(mockUserRepo, mockAppRepo, mockDynamicQuery)

		_, err := service.GetStats(ctx)
		assert.Error(t, err)

		mockUserRepo.AssertExpectations(t)
		mockAppRepo.AssertExpectations(t)
	})

	t.Run("count records error skipped", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockAppRepo := new(mocks.MockAppRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)

		mockUserRepo.On("Count", ctx).Return(int64(5), nil)
		mockAppRepo.On("GetAllTableNames", ctx).Return([]string{"app_data_1"}, nil)
		mockDynamicQuery.On("CountRecords", ctx, "app_data_1").Return(int64(0), errors.New("db error"))

		service := services.NewDashboardService(mockUserRepo, mockAppRepo, mockDynamicQuery)

		// Count errors are skipped in the service implementation
		stats, err := service.GetStats(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(1), stats.AppCount)
		assert.Equal(t, int64(0), stats.TotalRecords)

		mockUserRepo.AssertExpectations(t)
		mockAppRepo.AssertExpectations(t)
		mockDynamicQuery.AssertExpectations(t)
	})

	t.Run("count todays updates error skipped", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockAppRepo := new(mocks.MockAppRepository)
		mockDynamicQuery := new(mocks.MockDynamicQueryExecutor)

		mockUserRepo.On("Count", ctx).Return(int64(5), nil)
		mockAppRepo.On("GetAllTableNames", ctx).Return([]string{"app_data_1"}, nil)
		mockDynamicQuery.On("CountRecords", ctx, "app_data_1").Return(int64(100), nil)
		mockDynamicQuery.On("CountTodaysUpdates", ctx, "app_data_1").Return(int64(0), errors.New("db error"))

		service := services.NewDashboardService(mockUserRepo, mockAppRepo, mockDynamicQuery)

		// Update count errors are skipped in the service implementation
		stats, err := service.GetStats(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(100), stats.TotalRecords)
		assert.Equal(t, int64(0), stats.TodaysUpdates)

		mockUserRepo.AssertExpectations(t)
		mockAppRepo.AssertExpectations(t)
		mockDynamicQuery.AssertExpectations(t)
	})
}
