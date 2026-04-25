//go:build integration
// +build integration

package repositories

import (
	"context"
	"testing"
	"time"

	"nocode-app/backend/internal/models"
	"nocode-app/backend/internal/testhelpers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExternalQueryExecutor_PostgreSQL_Integration PostgreSQLの統合テスト
func TestExternalQueryExecutor_PostgreSQL_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("統合テストはショートモードでスキップされます")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// PostgreSQLコンテナをセットアップ
	container, err := testhelpers.SetupPostgresContainer(ctx)
	require.NoError(t, err, "PostgreSQLコンテナのセットアップに失敗しました")
	defer func() {
		if termErr := container.Terminate(ctx); termErr != nil {
			t.Logf("コンテナの終了に失敗しました: %v", termErr)
		}
	}()

	// テストテーブルを作成
	err = container.CreateTestTable(ctx)
	require.NoError(t, err, "テストテーブルの作成に失敗しました")

	// ExternalQueryExecutorを作成
	executor := NewExternalQueryExecutor()

	// DataSourceを作成
	ds := &models.DataSource{
		DBType:       models.DBTypePostgreSQL,
		Host:         container.Host,
		Port:         container.Port,
		DatabaseName: container.Database,
		Username:     container.Username,
	}

	t.Run("TestConnection", func(t *testing.T) {
		err := executor.TestConnection(ctx, ds, container.Password)
		assert.NoError(t, err, "PostgreSQL接続テストに失敗しました")
	})

	t.Run("GetTables", func(t *testing.T) {
		tables, err := executor.GetTables(ctx, ds, container.Password)
		require.NoError(t, err, "テーブル一覧の取得に失敗しました")

		// test_tableが存在することを確認
		found := false
		for _, table := range tables {
			if table.Name == "test_table" {
				found = true
				assert.Equal(t, models.TableTypeTable, table.Type, "test_tableのタイプがTABLEではありません")
				break
			}
		}
		assert.True(t, found, "test_tableが見つかりませんでした")
	})

	t.Run("GetTablesIncludingViews", func(t *testing.T) {
		// テストビューを作成
		err := container.CreateTestView(ctx)
		require.NoError(t, err, "テストビューの作成に失敗しました")

		tables, err := executor.GetTables(ctx, ds, container.Password)
		require.NoError(t, err, "テーブル一覧の取得に失敗しました")

		// test_viewが存在することを確認
		foundView := false
		foundTable := false
		for _, table := range tables {
			if table.Name == "test_view" {
				foundView = true
				assert.Equal(t, models.TableTypeView, table.Type, "test_viewのタイプがVIEWではありません")
			}
			if table.Name == "test_table" {
				foundTable = true
				assert.Equal(t, models.TableTypeTable, table.Type, "test_tableのタイプがTABLEではありません")
			}
		}
		assert.True(t, foundView, "test_viewが見つかりませんでした")
		assert.True(t, foundTable, "test_tableが見つかりませんでした")
	})

	t.Run("GetViewColumns", func(t *testing.T) {
		// ビューのカラム一覧を取得
		columns, err := executor.GetColumns(ctx, ds, container.Password, "test_view")
		require.NoError(t, err, "ビューのカラム一覧の取得に失敗しました")

		// カラムが存在することを確認（ビューはid, name, email, age, salaryの5カラム）
		assert.GreaterOrEqual(t, len(columns), 5, "ビューのカラム数が不足しています")

		// 各カラムの存在を確認
		columnNames := make(map[string]bool)
		for _, col := range columns {
			columnNames[col.Name] = true
		}

		expectedColumns := []string{"id", "name", "email", "age", "salary"}
		for _, expected := range expectedColumns {
			assert.True(t, columnNames[expected], "カラム %s が見つかりませんでした", expected)
		}
	})

	t.Run("GetColumns", func(t *testing.T) {
		columns, err := executor.GetColumns(ctx, ds, container.Password, "test_table")
		require.NoError(t, err, "カラム一覧の取得に失敗しました")

		// カラムが存在することを確認
		assert.GreaterOrEqual(t, len(columns), 7, "カラム数が不足しています")

		// 各カラムの存在を確認
		columnNames := make(map[string]bool)
		for _, col := range columns {
			columnNames[col.Name] = true
		}

		expectedColumns := []string{"id", "name", "email", "age", "salary", "is_active", "created_at"}
		for _, expected := range expectedColumns {
			assert.True(t, columnNames[expected], "カラム %s が見つかりませんでした", expected)
		}
	})

	t.Run("GetRecords", func(t *testing.T) {
		fields := createTestFields()
		opts := RecordQueryOptions{
			Page:  1,
			Limit: 10,
		}

		records, total, err := executor.GetRecords(ctx, ds, container.Password, "test_table", fields, opts)
		require.NoError(t, err, "レコードの取得に失敗しました")

		assert.Equal(t, int64(3), total, "レコード数が一致しません")
		assert.Len(t, records, 3, "取得されたレコード数が一致しません")
	})

	t.Run("CountRecords", func(t *testing.T) {
		count, err := executor.CountRecords(ctx, ds, container.Password, "test_table")
		require.NoError(t, err, "レコード数の取得に失敗しました")

		assert.Equal(t, int64(3), count, "レコード数が一致しません")
	})
}

// createTestFields PostgreSQL 用のテストフィールドセットを作成する
func createTestFields() []models.AppField {
	idCol := "id"
	nameCol := "name"
	emailCol := "email"
	ageCol := "age"
	salaryCol := "salary"
	isActiveCol := "is_active"
	createdAtCol := "created_at"

	return []models.AppField{
		{FieldCode: "id", FieldName: "ID", FieldType: "number", SourceColumnName: &idCol},
		{FieldCode: "name", FieldName: "名前", FieldType: "text", SourceColumnName: &nameCol},
		{FieldCode: "email", FieldName: "メール", FieldType: "text", SourceColumnName: &emailCol},
		{FieldCode: "age", FieldName: "年齢", FieldType: "number", SourceColumnName: &ageCol},
		{FieldCode: "salary", FieldName: "給与", FieldType: "number", SourceColumnName: &salaryCol},
		{FieldCode: "is_active", FieldName: "有効", FieldType: "checkbox", SourceColumnName: &isActiveCol},
		{FieldCode: "created_at", FieldName: "作成日時", FieldType: "datetime", SourceColumnName: &createdAtCol},
	}
}
