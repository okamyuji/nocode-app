package repositories_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nocode-app/backend/internal/models"
	"nocode-app/backend/internal/repositories"
	"nocode-app/backend/internal/testhelpers"
)

// createTestApp テスト用アプリを作成して返す
func createTestApp(ctx context.Context, t *testing.T, tableName string) *models.App {
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	appRepo := repositories.NewAppRepository(db)
	adminID := getAdminUserID(ctx, t)

	app := &models.App{
		Name:        "Test App for Fields",
		Description: "Test",
		TableName:   tableName,
		Icon:        "test",
		CreatedBy:   adminID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	require.NoError(t, appRepo.Create(ctx, app))
	return app
}

func TestFieldRepository_Create(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewFieldRepository(db)
	app := createTestApp(ctx, t, "app_data_field_create")

	field := &models.AppField{
		AppID:        app.ID,
		FieldCode:    "test_field",
		FieldName:    "Test Field",
		FieldType:    "TEXT",
		Required:     false,
		DisplayOrder: 1,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = repo.Create(ctx, field)
	require.NoError(t, err)
	assert.NotZero(t, field.ID)
}

func TestFieldRepository_CreateBatch(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewFieldRepository(db)
	app := createTestApp(ctx, t, "app_data_field_batch")

	fields := []models.AppField{
		{
			AppID:        app.ID,
			FieldCode:    "batch_field_1",
			FieldName:    "Batch Field 1",
			FieldType:    "TEXT",
			Required:     false,
			DisplayOrder: 1,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			AppID:        app.ID,
			FieldCode:    "batch_field_2",
			FieldName:    "Batch Field 2",
			FieldType:    "NUMBER",
			Required:     true,
			DisplayOrder: 2,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			AppID:        app.ID,
			FieldCode:    "batch_field_3",
			FieldName:    "Batch Field 3",
			FieldType:    "DATE",
			Required:     false,
			DisplayOrder: 3,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}

	err = repo.CreateBatch(ctx, fields)
	require.NoError(t, err)

	// 全フィールドが作成されたことを確認
	result, err := repo.GetByAppID(ctx, app.ID)
	require.NoError(t, err)
	assert.Len(t, result, 3)
}

func TestFieldRepository_CreateBatch_Empty(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	repo := repositories.NewFieldRepository(db)

	// 空のスライスでエラーにならないことを確認
	err = repo.CreateBatch(ctx, []models.AppField{})
	require.NoError(t, err)
}

func TestFieldRepository_GetByID(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewFieldRepository(db)
	app := createTestApp(ctx, t, "app_data_field_getbyid")

	field := &models.AppField{
		AppID:        app.ID,
		FieldCode:    "getbyid_field",
		FieldName:    "GetByID Field",
		FieldType:    "TEXT",
		Required:     true,
		DisplayOrder: 1,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, repo.Create(ctx, field))

	tests := []struct {
		name      string
		id        uint64
		wantField bool
		wantCode  string
	}{
		{
			name:      "存在するフィールド",
			id:        field.ID,
			wantField: true,
			wantCode:  "getbyid_field",
		},
		{
			name:      "存在しないフィールド",
			id:        99999,
			wantField: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByID(ctx, tt.id)
			require.NoError(t, err)
			if tt.wantField {
				require.NotNil(t, result)
				assert.Equal(t, tt.wantCode, result.FieldCode)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

func TestFieldRepository_GetByAppID(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewFieldRepository(db)
	app := createTestApp(ctx, t, "app_data_field_getbyappid")

	// 異なる表示順序でフィールドを作成
	fields := []models.AppField{
		{AppID: app.ID, FieldCode: "c_field", FieldName: "C Field", FieldType: "TEXT", DisplayOrder: 3, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{AppID: app.ID, FieldCode: "a_field", FieldName: "A Field", FieldType: "TEXT", DisplayOrder: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{AppID: app.ID, FieldCode: "b_field", FieldName: "B Field", FieldType: "TEXT", DisplayOrder: 2, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	require.NoError(t, repo.CreateBatch(ctx, fields))

	// フィールドを取得 - display_orderで順序付けされるべき
	result, err := repo.GetByAppID(ctx, app.ID)
	require.NoError(t, err)
	require.Len(t, result, 3)
	assert.Equal(t, "a_field", result[0].FieldCode)
	assert.Equal(t, "b_field", result[1].FieldCode)
	assert.Equal(t, "c_field", result[2].FieldCode)
}

func TestFieldRepository_GetByAppIDAndCode(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewFieldRepository(db)
	app := createTestApp(ctx, t, "app_data_field_getbycode")

	field := &models.AppField{
		AppID:        app.ID,
		FieldCode:    "unique_code",
		FieldName:    "Unique Field",
		FieldType:    "TEXT",
		DisplayOrder: 1,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, repo.Create(ctx, field))

	tests := []struct {
		name      string
		appID     uint64
		fieldCode string
		wantField bool
	}{
		{
			name:      "存在するフィールド",
			appID:     app.ID,
			fieldCode: "unique_code",
			wantField: true,
		},
		{
			name:      "不正なアプリID",
			appID:     99999,
			fieldCode: "unique_code",
			wantField: false,
		},
		{
			name:      "不正なフィールドコード",
			appID:     app.ID,
			fieldCode: "wrong_code",
			wantField: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByAppIDAndCode(ctx, tt.appID, tt.fieldCode)
			require.NoError(t, err)
			if tt.wantField {
				require.NotNil(t, result)
				assert.Equal(t, "Unique Field", result.FieldName)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

func TestFieldRepository_Update(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewFieldRepository(db)
	app := createTestApp(ctx, t, "app_data_field_update")

	field := &models.AppField{
		AppID:        app.ID,
		FieldCode:    "update_field",
		FieldName:    "Original Name",
		FieldType:    "TEXT",
		Required:     false,
		DisplayOrder: 1,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, repo.Create(ctx, field))

	// フィールドを更新
	field.FieldName = "Updated Name"
	field.Required = true
	err = repo.Update(ctx, field)
	require.NoError(t, err)

	// 更新を確認
	updated, err := repo.GetByID(ctx, field.ID)
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, "Updated Name", updated.FieldName)
	assert.True(t, updated.Required)
}

func TestFieldRepository_Delete(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewFieldRepository(db)
	app := createTestApp(ctx, t, "app_data_field_delete")

	field := &models.AppField{
		AppID:        app.ID,
		FieldCode:    "delete_field",
		FieldName:    "Delete Field",
		FieldType:    "TEXT",
		DisplayOrder: 1,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, repo.Create(ctx, field))

	// フィールドを削除
	err = repo.Delete(ctx, field.ID)
	require.NoError(t, err)

	// 削除を確認
	deleted, err := repo.GetByID(ctx, field.ID)
	require.NoError(t, err)
	assert.Nil(t, deleted)
}

func TestFieldRepository_UpdateOrder(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewFieldRepository(db)
	app := createTestApp(ctx, t, "app_data_field_order")

	// フィールドを作成
	fields := []models.AppField{
		{AppID: app.ID, FieldCode: "field_1", FieldName: "Field 1", FieldType: "TEXT", DisplayOrder: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{AppID: app.ID, FieldCode: "field_2", FieldName: "Field 2", FieldType: "TEXT", DisplayOrder: 2, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{AppID: app.ID, FieldCode: "field_3", FieldName: "Field 3", FieldType: "TEXT", DisplayOrder: 3, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	require.NoError(t, repo.CreateBatch(ctx, fields))

	// 作成されたフィールドを取得してIDを取得
	createdFields, err := repo.GetByAppID(ctx, app.ID)
	require.NoError(t, err)
	require.Len(t, createdFields, 3)

	// 順序を変更: 3, 1, 2
	orderItems := []models.FieldOrderItem{
		{ID: createdFields[0].ID, DisplayOrder: 2}, // field_1 -> 2
		{ID: createdFields[1].ID, DisplayOrder: 3}, // field_2 -> 3
		{ID: createdFields[2].ID, DisplayOrder: 1}, // field_3 -> 1
	}
	err = repo.UpdateOrder(ctx, orderItems)
	require.NoError(t, err)

	// 新しい順序を確認
	reordered, err := repo.GetByAppID(ctx, app.ID)
	require.NoError(t, err)
	assert.Equal(t, "field_3", reordered[0].FieldCode)
	assert.Equal(t, "field_1", reordered[1].FieldCode)
	assert.Equal(t, "field_2", reordered[2].FieldCode)
}

func TestFieldRepository_FieldCodeExists(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewFieldRepository(db)
	app := createTestApp(ctx, t, "app_data_field_exists")

	field := &models.AppField{
		AppID:        app.ID,
		FieldCode:    "existing_code",
		FieldName:    "Existing Field",
		FieldType:    "TEXT",
		DisplayOrder: 1,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, repo.Create(ctx, field))

	tests := []struct {
		name      string
		appID     uint64
		fieldCode string
		exists    bool
	}{
		{
			name:      "存在するコード",
			appID:     app.ID,
			fieldCode: "existing_code",
			exists:    true,
		},
		{
			name:      "存在しないコード",
			appID:     app.ID,
			fieldCode: "nonexistent",
			exists:    false,
		},
		{
			name:      "不正なアプリ",
			appID:     99999,
			fieldCode: "existing_code",
			exists:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := repo.FieldCodeExists(ctx, tt.appID, tt.fieldCode)
			require.NoError(t, err)
			assert.Equal(t, tt.exists, exists)
		})
	}
}

func TestFieldRepository_GetMaxDisplayOrder(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewFieldRepository(db)
	app := createTestApp(ctx, t, "app_data_field_maxorder")

	// 初期状態ではフィールドなし
	maxOrder, err := repo.GetMaxDisplayOrder(ctx, app.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, maxOrder)

	// フィールドを追加
	fields := []models.AppField{
		{AppID: app.ID, FieldCode: "f1", FieldName: "F1", FieldType: "TEXT", DisplayOrder: 5, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{AppID: app.ID, FieldCode: "f2", FieldName: "F2", FieldType: "TEXT", DisplayOrder: 10, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{AppID: app.ID, FieldCode: "f3", FieldName: "F3", FieldType: "TEXT", DisplayOrder: 3, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	require.NoError(t, repo.CreateBatch(ctx, fields))

	// 最大値は10であるべき
	maxOrder, err = repo.GetMaxDisplayOrder(ctx, app.ID)
	require.NoError(t, err)
	assert.Equal(t, 10, maxOrder)
}
