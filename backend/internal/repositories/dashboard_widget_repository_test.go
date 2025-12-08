package repositories_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nocode-app/backend/internal/models"
	"nocode-app/backend/internal/repositories"
	"nocode-app/backend/internal/testhelpers"
)

// widgetTestCounter is used to generate unique names for tests
var widgetTestCounter int

// createTestUserForWidget creates a test user for dashboard widget tests
func createTestUserForWidget(ctx context.Context, t *testing.T, testName string) *models.User {
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)

	widgetTestCounter++
	email := fmt.Sprintf("widget_user_%s_%d_%d@example.com", testName, time.Now().UnixNano(), widgetTestCounter)

	user := &models.User{
		Email:        email,
		PasswordHash: "hash",
		Name:         "Widget Test User",
		Role:         "admin",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, userRepo.Create(ctx, user))
	return user
}

// createTestAppForWidget creates a test app for dashboard widget tests
func createTestAppForWidget(ctx context.Context, t *testing.T, name string, userID uint64) *models.App {
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	appRepo := repositories.NewAppRepository(db)

	widgetTestCounter++
	uniqueName := fmt.Sprintf("app_data_widget_%s_%d_%d", name, time.Now().UnixNano(), widgetTestCounter)

	app := &models.App{
		Name:        name,
		Description: "Test app for widgets",
		TableName:   uniqueName,
		Icon:        "test",
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	require.NoError(t, appRepo.Create(ctx, app))
	return app
}

func TestDashboardWidgetRepository_Create(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewDashboardWidgetRepository(db)
	user := createTestUserForWidget(ctx, t, "create")
	app := createTestAppForWidget(ctx, t, "create_test", user.ID)

	widget := &models.DashboardWidget{
		UserID:       user.ID,
		AppID:        app.ID,
		DisplayOrder: 0,
		ViewType:     "table",
		IsVisible:    true,
		WidgetSize:   "medium",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = repo.Create(ctx, widget)
	require.NoError(t, err)
	assert.NotZero(t, widget.ID)
}

func TestDashboardWidgetRepository_Create_ViewTypes(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewDashboardWidgetRepository(db)
	user := createTestUserForWidget(ctx, t, "viewtypes")

	tests := []struct {
		name       string
		viewType   models.WidgetViewType
		widgetSize models.WidgetSize
	}{
		{"テーブルビュー", models.WidgetViewTypeTable, models.WidgetSizeMedium},
		{"リストビュー", models.WidgetViewTypeList, models.WidgetSizeSmall},
		{"チャートビュー", models.WidgetViewTypeChart, models.WidgetSizeLarge},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := createTestAppForWidget(ctx, t, string(tt.viewType), user.ID)
			widget := &models.DashboardWidget{
				UserID:       user.ID,
				AppID:        app.ID,
				DisplayOrder: 0,
				ViewType:     tt.viewType,
				IsVisible:    true,
				WidgetSize:   tt.widgetSize,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}
			err := repo.Create(ctx, widget)
			require.NoError(t, err)
			assert.NotZero(t, widget.ID)
		})
	}
}

func TestDashboardWidgetRepository_GetByID(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewDashboardWidgetRepository(db)
	user := createTestUserForWidget(ctx, t, "getbyid")
	app := createTestAppForWidget(ctx, t, "getbyid_test", user.ID)

	// Create test widget
	widget := &models.DashboardWidget{
		UserID:       user.ID,
		AppID:        app.ID,
		DisplayOrder: 0,
		ViewType:     "table",
		IsVisible:    true,
		WidgetSize:   "medium",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, repo.Create(ctx, widget))

	t.Run("正常系_存在するウィジェット", func(t *testing.T) {
		result, err := repo.GetByID(ctx, widget.ID)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, widget.ID, result.ID)
		assert.Equal(t, user.ID, result.UserID)
		assert.Equal(t, models.WidgetViewTypeTable, result.ViewType)
	})

	t.Run("正常系_存在しないウィジェット", func(t *testing.T) {
		result, err := repo.GetByID(ctx, 99999)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("境界値_ID=0", func(t *testing.T) {
		result, err := repo.GetByID(ctx, 0)
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestDashboardWidgetRepository_GetByUserID(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewDashboardWidgetRepository(db)

	// Create two users
	user1 := createTestUserForWidget(ctx, t, "getbyuserid1")
	user2 := createTestUserForWidget(ctx, t, "getbyuserid2")

	// Create widgets for user1 (3 widgets)
	for i := 0; i < 3; i++ {
		app := createTestAppForWidget(ctx, t, fmt.Sprintf("user1_app_%d", i), user1.ID)
		widget := &models.DashboardWidget{
			UserID:       user1.ID,
			AppID:        app.ID,
			DisplayOrder: i,
			ViewType:     "table",
			IsVisible:    true,
			WidgetSize:   "medium",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		require.NoError(t, repo.Create(ctx, widget))
	}

	// Create widgets for user2 (1 widget)
	app := createTestAppForWidget(ctx, t, "user2_app", user2.ID)
	widget := &models.DashboardWidget{
		UserID:       user2.ID,
		AppID:        app.ID,
		DisplayOrder: 0,
		ViewType:     "chart",
		IsVisible:    true,
		WidgetSize:   "large",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, repo.Create(ctx, widget))

	t.Run("正常系_user1のウィジェット", func(t *testing.T) {
		widgets, err := repo.GetByUserID(ctx, user1.ID)
		require.NoError(t, err)
		assert.Len(t, widgets, 3)
	})

	t.Run("正常系_user2のウィジェット", func(t *testing.T) {
		widgets, err := repo.GetByUserID(ctx, user2.ID)
		require.NoError(t, err)
		assert.Len(t, widgets, 1)
	})

	t.Run("正常系_存在しないユーザー", func(t *testing.T) {
		widgets, err := repo.GetByUserID(ctx, 99999)
		require.NoError(t, err)
		assert.Len(t, widgets, 0)
	})
}

func TestDashboardWidgetRepository_GetByUserIDAndAppID(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewDashboardWidgetRepository(db)
	user := createTestUserForWidget(ctx, t, "getuserandapp")
	app := createTestAppForWidget(ctx, t, "userandapp_test", user.ID)

	// Create test widget
	widget := &models.DashboardWidget{
		UserID:       user.ID,
		AppID:        app.ID,
		DisplayOrder: 0,
		ViewType:     "list",
		IsVisible:    true,
		WidgetSize:   "small",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, repo.Create(ctx, widget))

	t.Run("正常系_存在する組み合わせ", func(t *testing.T) {
		result, err := repo.GetByUserIDAndAppID(ctx, user.ID, app.ID)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, models.WidgetViewTypeList, result.ViewType)
	})

	t.Run("正常系_存在しないユーザー", func(t *testing.T) {
		result, err := repo.GetByUserIDAndAppID(ctx, 99999, app.ID)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("正常系_存在しないアプリ", func(t *testing.T) {
		result, err := repo.GetByUserIDAndAppID(ctx, user.ID, 99999)
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestDashboardWidgetRepository_GetVisibleByUserID(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewDashboardWidgetRepository(db)
	user := createTestUserForWidget(ctx, t, "getvisible")

	// Create 3 apps
	app1 := createTestAppForWidget(ctx, t, "visible_test_1", user.ID)
	app2 := createTestAppForWidget(ctx, t, "visible_test_2", user.ID)
	app3 := createTestAppForWidget(ctx, t, "visible_test_3", user.ID)

	// Create widgets (2 visible, 1 hidden)
	visibleWidget1 := &models.DashboardWidget{
		UserID:       user.ID,
		AppID:        app1.ID,
		DisplayOrder: 0,
		ViewType:     models.WidgetViewTypeTable,
		IsVisible:    true,
		WidgetSize:   models.WidgetSizeMedium,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, repo.Create(ctx, visibleWidget1))

	// Verify hiddenWidget is created with is_visible = false
	hiddenWidget := &models.DashboardWidget{
		UserID:       user.ID,
		AppID:        app2.ID,
		DisplayOrder: 1,
		ViewType:     models.WidgetViewTypeChart,
		IsVisible:    false, // This should be false
		WidgetSize:   models.WidgetSizeLarge,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, repo.Create(ctx, hiddenWidget))

	// Verify the hidden widget was created with IsVisible = false
	createdHidden, err := repo.GetByID(ctx, hiddenWidget.ID)
	require.NoError(t, err)
	require.NotNil(t, createdHidden)
	assert.False(t, createdHidden.IsVisible, "Hidden widget should have IsVisible = false")

	visibleWidget2 := &models.DashboardWidget{
		UserID:       user.ID,
		AppID:        app3.ID,
		DisplayOrder: 2,
		ViewType:     models.WidgetViewTypeList,
		IsVisible:    true,
		WidgetSize:   models.WidgetSizeSmall,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, repo.Create(ctx, visibleWidget2))

	// Get visible widgets
	visibleWidgets, err := repo.GetVisibleByUserID(ctx, user.ID)
	require.NoError(t, err)
	assert.Len(t, visibleWidgets, 2, "Should return only visible widgets")

	// Verify all returned widgets are visible
	for _, w := range visibleWidgets {
		assert.True(t, w.IsVisible, "All returned widgets should be visible")
	}

	// Verify widgets are ordered by display_order (ascending)
	if len(visibleWidgets) >= 2 {
		assert.True(t, visibleWidgets[0].DisplayOrder < visibleWidgets[1].DisplayOrder,
			"Widgets should be ordered by display_order")
	}
}

func TestDashboardWidgetRepository_Update(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewDashboardWidgetRepository(db)
	user := createTestUserForWidget(ctx, t, "update")
	app := createTestAppForWidget(ctx, t, "update_test", user.ID)

	// Create test widget
	widget := &models.DashboardWidget{
		UserID:       user.ID,
		AppID:        app.ID,
		DisplayOrder: 0,
		ViewType:     "table",
		IsVisible:    true,
		WidgetSize:   "medium",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, repo.Create(ctx, widget))

	t.Run("正常系_ビュータイプ変更", func(t *testing.T) {
		widget.ViewType = models.WidgetViewTypeChart
		err := repo.Update(ctx, widget)
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, widget.ID)
		require.NoError(t, err)
		assert.Equal(t, models.WidgetViewTypeChart, updated.ViewType)
	})

	t.Run("正常系_サイズ変更", func(t *testing.T) {
		widget.WidgetSize = models.WidgetSizeLarge
		err := repo.Update(ctx, widget)
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, widget.ID)
		require.NoError(t, err)
		assert.Equal(t, models.WidgetSizeLarge, updated.WidgetSize)
	})

	t.Run("正常系_非表示に変更", func(t *testing.T) {
		widget.IsVisible = false
		err := repo.Update(ctx, widget)
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, widget.ID)
		require.NoError(t, err)
		assert.False(t, updated.IsVisible)
	})

	t.Run("正常系_表示順序変更", func(t *testing.T) {
		widget.DisplayOrder = 5
		err := repo.Update(ctx, widget)
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, widget.ID)
		require.NoError(t, err)
		assert.Equal(t, 5, updated.DisplayOrder)
	})
}

func TestDashboardWidgetRepository_Delete(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewDashboardWidgetRepository(db)
	user := createTestUserForWidget(ctx, t, "delete")
	app := createTestAppForWidget(ctx, t, "delete_test", user.ID)

	// Create test widget
	widget := &models.DashboardWidget{
		UserID:       user.ID,
		AppID:        app.ID,
		DisplayOrder: 0,
		ViewType:     "table",
		IsVisible:    true,
		WidgetSize:   "medium",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, repo.Create(ctx, widget))

	// Delete widget
	err = repo.Delete(ctx, widget.ID)
	require.NoError(t, err)

	// Verify deletion
	deleted, err := repo.GetByID(ctx, widget.ID)
	require.NoError(t, err)
	assert.Nil(t, deleted)

	// Delete non-existing widget (should not error)
	err = repo.Delete(ctx, 99999)
	require.NoError(t, err)
}

func TestDashboardWidgetRepository_DeleteByUserIDAndAppID(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewDashboardWidgetRepository(db)
	user := createTestUserForWidget(ctx, t, "deletebyuserapp")
	app := createTestAppForWidget(ctx, t, "delete_user_app_test", user.ID)

	// Create test widget
	widget := &models.DashboardWidget{
		UserID:       user.ID,
		AppID:        app.ID,
		DisplayOrder: 0,
		ViewType:     "table",
		IsVisible:    true,
		WidgetSize:   "medium",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, repo.Create(ctx, widget))

	// Delete by user and app ID
	err = repo.DeleteByUserIDAndAppID(ctx, user.ID, app.ID)
	require.NoError(t, err)

	// Verify deletion
	deleted, err := repo.GetByUserIDAndAppID(ctx, user.ID, app.ID)
	require.NoError(t, err)
	assert.Nil(t, deleted)
}

func TestDashboardWidgetRepository_UpdateDisplayOrders(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewDashboardWidgetRepository(db)
	user := createTestUserForWidget(ctx, t, "updateorders")

	// Create 3 apps and widgets with initial order
	widgets := make([]*models.DashboardWidget, 3)
	for i := 0; i < 3; i++ {
		app := createTestAppForWidget(ctx, t, fmt.Sprintf("order_test_%d", i), user.ID)
		widgets[i] = &models.DashboardWidget{
			UserID:       user.ID,
			AppID:        app.ID,
			DisplayOrder: i,
			ViewType:     "table",
			IsVisible:    true,
			WidgetSize:   "medium",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		require.NoError(t, repo.Create(ctx, widgets[i]))
	}

	t.Run("正常系_順序逆転", func(t *testing.T) {
		// Reverse order: [2, 1, 0]
		newOrder := []uint64{widgets[2].ID, widgets[1].ID, widgets[0].ID}
		err := repo.UpdateDisplayOrders(ctx, user.ID, newOrder)
		require.NoError(t, err)

		result, err := repo.GetByUserID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, widgets[2].ID, result[0].ID)
		assert.Equal(t, 0, result[0].DisplayOrder)
		assert.Equal(t, widgets[1].ID, result[1].ID)
		assert.Equal(t, 1, result[1].DisplayOrder)
		assert.Equal(t, widgets[0].ID, result[2].ID)
		assert.Equal(t, 2, result[2].DisplayOrder)
	})

	t.Run("境界値_空の配列", func(t *testing.T) {
		err := repo.UpdateDisplayOrders(ctx, user.ID, []uint64{})
		require.NoError(t, err)
		// Should not error and not change anything
	})
}

func TestDashboardWidgetRepository_GetMaxDisplayOrder(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewDashboardWidgetRepository(db)

	// Create user with no widgets
	userNoWidgets := createTestUserForWidget(ctx, t, "maxorder_empty")

	t.Run("正常系_ウィジェットなし", func(t *testing.T) {
		maxOrder, err := repo.GetMaxDisplayOrder(ctx, userNoWidgets.ID)
		require.NoError(t, err)
		assert.Equal(t, -1, maxOrder) // COALESCE returns -1 when no widgets
	})

	// Create user with widgets
	userWithWidgets := createTestUserForWidget(ctx, t, "maxorder_has")
	for i := 0; i < 3; i++ {
		app := createTestAppForWidget(ctx, t, fmt.Sprintf("maxorder_app_%d", i), userWithWidgets.ID)
		widget := &models.DashboardWidget{
			UserID:       userWithWidgets.ID,
			AppID:        app.ID,
			DisplayOrder: i * 2, // 0, 2, 4
			ViewType:     "table",
			IsVisible:    true,
			WidgetSize:   "medium",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		require.NoError(t, repo.Create(ctx, widget))
	}

	t.Run("正常系_最大順序を取得", func(t *testing.T) {
		maxOrder, err := repo.GetMaxDisplayOrder(ctx, userWithWidgets.ID)
		require.NoError(t, err)
		assert.Equal(t, 4, maxOrder)
	})
}

func TestDashboardWidgetRepository_Exists(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewDashboardWidgetRepository(db)
	user := createTestUserForWidget(ctx, t, "exists")
	app := createTestAppForWidget(ctx, t, "exists_test", user.ID)

	t.Run("正常系_存在しない", func(t *testing.T) {
		exists, err := repo.Exists(ctx, user.ID, app.ID)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	// Create widget
	widget := &models.DashboardWidget{
		UserID:       user.ID,
		AppID:        app.ID,
		DisplayOrder: 0,
		ViewType:     "table",
		IsVisible:    true,
		WidgetSize:   "medium",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, repo.Create(ctx, widget))

	t.Run("正常系_存在する", func(t *testing.T) {
		exists, err := repo.Exists(ctx, user.ID, app.ID)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("異常系_異なるユーザー", func(t *testing.T) {
		exists, err := repo.Exists(ctx, 99999, app.ID)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("異常系_異なるアプリ", func(t *testing.T) {
		exists, err := repo.Exists(ctx, user.ID, 99999)
		require.NoError(t, err)
		assert.False(t, exists)
	})
}
