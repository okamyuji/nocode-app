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

func TestViewService_GetViews(t *testing.T) {
	ctx := context.Background()

	t.Run("successful get views", func(t *testing.T) {
		mockViewRepo := new(mocks.MockViewRepository)
		mockAppRepo := new(mocks.MockAppRepository)

		views := []models.AppView{
			{ID: 1, AppID: 1, Name: "Table View", ViewType: "table", IsDefault: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: 2, AppID: 1, Name: "List View", ViewType: "list", IsDefault: false, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		}

		mockViewRepo.On("GetByAppID", ctx, uint64(1)).Return(views, nil)

		service := services.NewViewService(mockViewRepo, mockAppRepo)

		resp, err := service.GetViews(ctx, 1)
		require.NoError(t, err)
		assert.Len(t, resp, 2)
		assert.Equal(t, "Table View", resp[0].Name)

		mockViewRepo.AssertExpectations(t)
	})
}

func TestViewService_CreateView(t *testing.T) {
	ctx := context.Background()

	t.Run("successful creation with default", func(t *testing.T) {
		mockViewRepo := new(mocks.MockViewRepository)
		mockAppRepo := new(mocks.MockAppRepository)

		app := &models.App{ID: 1}

		mockAppRepo.On("GetByID", ctx, uint64(1)).Return(app, nil)
		mockViewRepo.On("ClearDefaultByAppID", ctx, uint64(1)).Return(nil)
		mockViewRepo.On("Create", ctx, mock.AnythingOfType("*models.AppView")).Return(nil).Run(func(args mock.Arguments) {
			view := args.Get(1).(*models.AppView)
			view.ID = 1
		})

		service := services.NewViewService(mockViewRepo, mockAppRepo)

		req := &models.CreateViewRequest{
			Name:      "New View",
			ViewType:  "table",
			IsDefault: true,
		}

		resp, err := service.CreateView(ctx, 1, req)
		require.NoError(t, err)
		assert.Equal(t, "New View", resp.Name)
		assert.True(t, resp.IsDefault)

		mockViewRepo.AssertExpectations(t)
		mockAppRepo.AssertExpectations(t)
	})

	t.Run("app not found", func(t *testing.T) {
		mockViewRepo := new(mocks.MockViewRepository)
		mockAppRepo := new(mocks.MockAppRepository)

		mockAppRepo.On("GetByID", ctx, uint64(999)).Return(nil, nil)

		service := services.NewViewService(mockViewRepo, mockAppRepo)

		req := &models.CreateViewRequest{
			Name:     "New View",
			ViewType: "table",
		}

		_, err := service.CreateView(ctx, 999, req)
		assert.ErrorIs(t, err, services.ErrAppNotFound)

		mockAppRepo.AssertExpectations(t)
	})
}

func TestViewService_UpdateView(t *testing.T) {
	ctx := context.Background()

	t.Run("successful update", func(t *testing.T) {
		mockViewRepo := new(mocks.MockViewRepository)
		mockAppRepo := new(mocks.MockAppRepository)

		view := &models.AppView{
			ID:        1,
			AppID:     1,
			Name:      "Original Name",
			ViewType:  "table",
			IsDefault: false,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockViewRepo.On("GetByID", ctx, uint64(1)).Return(view, nil)
		mockViewRepo.On("Update", ctx, mock.AnythingOfType("*models.AppView")).Return(nil)

		service := services.NewViewService(mockViewRepo, mockAppRepo)

		name := "Updated Name"
		req := &models.UpdateViewRequest{
			Name: name,
		}

		resp, err := service.UpdateView(ctx, 1, req)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", resp.Name)

		mockViewRepo.AssertExpectations(t)
	})

	t.Run("update to default", func(t *testing.T) {
		mockViewRepo := new(mocks.MockViewRepository)
		mockAppRepo := new(mocks.MockAppRepository)

		view := &models.AppView{
			ID:        1,
			AppID:     1,
			Name:      "View",
			ViewType:  "table",
			IsDefault: false,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockViewRepo.On("GetByID", ctx, uint64(1)).Return(view, nil)
		mockViewRepo.On("ClearDefaultByAppID", ctx, uint64(1)).Return(nil)
		mockViewRepo.On("Update", ctx, mock.AnythingOfType("*models.AppView")).Return(nil)

		service := services.NewViewService(mockViewRepo, mockAppRepo)

		isDefault := true
		req := &models.UpdateViewRequest{
			IsDefault: &isDefault,
		}

		resp, err := service.UpdateView(ctx, 1, req)
		require.NoError(t, err)
		assert.True(t, resp.IsDefault)

		mockViewRepo.AssertExpectations(t)
	})

	t.Run("view not found", func(t *testing.T) {
		mockViewRepo := new(mocks.MockViewRepository)
		mockAppRepo := new(mocks.MockAppRepository)

		mockViewRepo.On("GetByID", ctx, uint64(999)).Return(nil, nil)

		service := services.NewViewService(mockViewRepo, mockAppRepo)

		req := &models.UpdateViewRequest{}

		_, err := service.UpdateView(ctx, 999, req)
		assert.ErrorIs(t, err, services.ErrViewNotFound)

		mockViewRepo.AssertExpectations(t)
	})
}

func TestViewService_DeleteView(t *testing.T) {
	ctx := context.Background()

	t.Run("successful delete", func(t *testing.T) {
		mockViewRepo := new(mocks.MockViewRepository)
		mockAppRepo := new(mocks.MockAppRepository)

		view := &models.AppView{ID: 1, AppID: 1}

		mockViewRepo.On("GetByID", ctx, uint64(1)).Return(view, nil)
		mockViewRepo.On("Delete", ctx, uint64(1)).Return(nil)

		service := services.NewViewService(mockViewRepo, mockAppRepo)

		err := service.DeleteView(ctx, 1)
		require.NoError(t, err)

		mockViewRepo.AssertExpectations(t)
	})

	t.Run("view not found", func(t *testing.T) {
		mockViewRepo := new(mocks.MockViewRepository)
		mockAppRepo := new(mocks.MockAppRepository)

		mockViewRepo.On("GetByID", ctx, uint64(999)).Return(nil, nil)

		service := services.NewViewService(mockViewRepo, mockAppRepo)

		err := service.DeleteView(ctx, 999)
		assert.ErrorIs(t, err, services.ErrViewNotFound)

		mockViewRepo.AssertExpectations(t)
	})
}
