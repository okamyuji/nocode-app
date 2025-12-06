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

func TestDynamicQueryExecutor_CreateTable(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	executor := repositories.NewDynamicQueryExecutor(db)

	fields := []models.AppField{
		{FieldCode: "title", FieldName: "Title", FieldType: "TEXT"},
		{FieldCode: "amount", FieldName: "Amount", FieldType: "NUMBER"},
		{FieldCode: "is_active", FieldName: "Is Active", FieldType: "CHECKBOX"},
		{FieldCode: "due_date", FieldName: "Due Date", FieldType: "DATE"},
	}

	err = executor.CreateTable(ctx, "app_data_dynamic_1", fields)
	require.NoError(t, err)

	// レコードを挿入してテーブルの存在を確認
	adminID := getAdminUserID(ctx, t)
	recordID, err := executor.InsertRecord(ctx, "app_data_dynamic_1", models.RecordData{
		"title":     "Test",
		"amount":    100,
		"is_active": true,
	}, adminID)
	require.NoError(t, err)
	assert.NotZero(t, recordID)
}

func TestDynamicQueryExecutor_CreateTable_InvalidName(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	executor := repositories.NewDynamicQueryExecutor(db)

	tests := []struct {
		name      string
		tableName string
	}{
		{"空の名前", ""},
		{"SQLインジェクション", "users; DROP TABLE users;--"},
		{"特殊文字", "table-name"},
		{"数字で始まる", "1table"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := executor.CreateTable(ctx, tt.tableName, nil)
			assert.Error(t, err)
		})
	}
}

func TestDynamicQueryExecutor_DropTable(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	executor := repositories.NewDynamicQueryExecutor(db)

	// まずテーブルを作成
	err = executor.CreateTable(ctx, "app_data_drop_test", []models.AppField{
		{FieldCode: "field1", FieldName: "Field 1", FieldType: "TEXT"},
	})
	require.NoError(t, err)

	// テーブルを削除
	err = executor.DropTable(ctx, "app_data_drop_test")
	require.NoError(t, err)

	// 挿入を試みてテーブルが削除されたことを確認（失敗するはず）
	_, err = executor.InsertRecord(ctx, "app_data_drop_test", models.RecordData{"field1": "test"}, 1)
	assert.Error(t, err)
}

func TestDynamicQueryExecutor_AddColumn(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	executor := repositories.NewDynamicQueryExecutor(db)

	// テーブルを作成
	err = executor.CreateTable(ctx, "app_data_addcol", []models.AppField{
		{FieldCode: "existing", FieldName: "Existing", FieldType: "TEXT"},
	})
	require.NoError(t, err)

	// 新しいカラムを追加
	newField := &models.AppField{
		FieldCode: "new_field",
		FieldName: "New Field",
		FieldType: "NUMBER",
	}
	err = executor.AddColumn(ctx, "app_data_addcol", newField)
	require.NoError(t, err)

	// 新しいカラムが機能することを確認
	adminID := getAdminUserID(ctx, t)
	recordID, err := executor.InsertRecord(ctx, "app_data_addcol", models.RecordData{
		"existing":  "test",
		"new_field": 42,
	}, adminID)
	require.NoError(t, err)
	assert.NotZero(t, recordID)
}

func TestDynamicQueryExecutor_DropColumn(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	executor := repositories.NewDynamicQueryExecutor(db)

	// 2つのカラムを持つテーブルを作成
	err = executor.CreateTable(ctx, "app_data_dropcol", []models.AppField{
		{FieldCode: "keep_field", FieldName: "Keep", FieldType: "TEXT"},
		{FieldCode: "drop_field", FieldName: "Drop", FieldType: "TEXT"},
	})
	require.NoError(t, err)

	// カラムを削除
	err = executor.DropColumn(ctx, "app_data_dropcol", "drop_field")
	require.NoError(t, err)

	// 削除されたカラムの使用を試みてカラムが削除されたことを確認（失敗するはず）
	adminID := getAdminUserID(ctx, t)
	_, err = executor.InsertRecord(ctx, "app_data_dropcol", models.RecordData{
		"drop_field": "test",
	}, adminID)
	assert.Error(t, err)

	// keep_fieldはまだ機能するはず
	recordID, err := executor.InsertRecord(ctx, "app_data_dropcol", models.RecordData{
		"keep_field": "test",
	}, adminID)
	require.NoError(t, err)
	assert.NotZero(t, recordID)
}

func TestDynamicQueryExecutor_InsertRecord(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	executor := repositories.NewDynamicQueryExecutor(db)
	adminID := getAdminUserID(ctx, t)

	// テーブルを作成
	fields := []models.AppField{
		{FieldCode: "name", FieldName: "Name", FieldType: "TEXT"},
		{FieldCode: "age", FieldName: "Age", FieldType: "NUMBER"},
	}
	require.NoError(t, executor.CreateTable(ctx, "app_data_insert", fields))

	// レコードを挿入
	data := models.RecordData{
		"name": "John Doe",
		"age":  30,
	}
	recordID, err := executor.InsertRecord(ctx, "app_data_insert", data, adminID)
	require.NoError(t, err)
	assert.NotZero(t, recordID)

	// 挿入を確認
	record, err := executor.GetRecordByID(ctx, "app_data_insert", fields, recordID)
	require.NoError(t, err)
	require.NotNil(t, record)
	assert.Equal(t, "John Doe", record.Data["name"])
	// 数値フィールドはデータベーススキャン後に文字列として返される
	assert.Equal(t, "30", record.Data["age"])
}

func TestDynamicQueryExecutor_UpdateRecord(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	executor := repositories.NewDynamicQueryExecutor(db)
	adminID := getAdminUserID(ctx, t)

	// テーブルを作成
	fields := []models.AppField{
		{FieldCode: "title", FieldName: "Title", FieldType: "TEXT"},
		{FieldCode: "status", FieldName: "Status", FieldType: "TEXT"},
	}
	require.NoError(t, executor.CreateTable(ctx, "app_data_update", fields))

	// レコードを挿入
	recordID, err := executor.InsertRecord(ctx, "app_data_update", models.RecordData{
		"title":  "Original",
		"status": "draft",
	}, adminID)
	require.NoError(t, err)

	// レコードを更新
	err = executor.UpdateRecord(ctx, "app_data_update", recordID, models.RecordData{
		"title":  "Updated",
		"status": "published",
	})
	require.NoError(t, err)

	// 更新を確認
	record, err := executor.GetRecordByID(ctx, "app_data_update", fields, recordID)
	require.NoError(t, err)
	require.NotNil(t, record)
	assert.Equal(t, "Updated", record.Data["title"])
	assert.Equal(t, "published", record.Data["status"])
}

func TestDynamicQueryExecutor_DeleteRecord(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	executor := repositories.NewDynamicQueryExecutor(db)
	adminID := getAdminUserID(ctx, t)

	// テーブルを作成
	fields := []models.AppField{
		{FieldCode: "name", FieldName: "Name", FieldType: "TEXT"},
	}
	require.NoError(t, executor.CreateTable(ctx, "app_data_delete", fields))

	// レコードを挿入
	recordID, err := executor.InsertRecord(ctx, "app_data_delete", models.RecordData{"name": "Delete Me"}, adminID)
	require.NoError(t, err)

	// レコードを削除
	err = executor.DeleteRecord(ctx, "app_data_delete", recordID)
	require.NoError(t, err)

	// 削除を確認
	record, err := executor.GetRecordByID(ctx, "app_data_delete", fields, recordID)
	require.NoError(t, err)
	assert.Nil(t, record)
}

func TestDynamicQueryExecutor_DeleteRecords(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	executor := repositories.NewDynamicQueryExecutor(db)
	adminID := getAdminUserID(ctx, t)

	// テーブルを作成
	fields := []models.AppField{
		{FieldCode: "name", FieldName: "Name", FieldType: "TEXT"},
	}
	require.NoError(t, executor.CreateTable(ctx, "app_data_bulk_delete", fields))

	// 複数のレコードを挿入
	var ids []uint64
	for i := 0; i < 5; i++ {
		recordID, insertErr := executor.InsertRecord(ctx, "app_data_bulk_delete", models.RecordData{"name": "Record"}, adminID)
		require.NoError(t, insertErr)
		ids = append(ids, recordID)
	}

	// 最初の3レコードを削除
	err = executor.DeleteRecords(ctx, "app_data_bulk_delete", ids[:3])
	require.NoError(t, err)

	// 削除を確認
	records, total, err := executor.GetRecords(ctx, "app_data_bulk_delete", fields, repositories.RecordQueryOptions{Page: 1, Limit: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, records, 2)
}

func TestDynamicQueryExecutor_DeleteRecords_Empty(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	executor := repositories.NewDynamicQueryExecutor(db)

	// 空のスライスでエラーにならないことを確認
	err = executor.DeleteRecords(ctx, "any_table", []uint64{})
	require.NoError(t, err)
}

func TestDynamicQueryExecutor_GetRecords(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	executor := repositories.NewDynamicQueryExecutor(db)
	adminID := getAdminUserID(ctx, t)

	// テーブルを作成
	fields := []models.AppField{
		{FieldCode: "name", FieldName: "Name", FieldType: "TEXT"},
		{FieldCode: "priority", FieldName: "Priority", FieldType: "NUMBER"},
	}
	require.NoError(t, executor.CreateTable(ctx, "app_data_getrecords", fields))

	// テストレコードを挿入
	for i := 1; i <= 10; i++ {
		_, insertErr := executor.InsertRecord(ctx, "app_data_getrecords", models.RecordData{
			"name":     "Record " + string(rune('A'+i-1)),
			"priority": i,
		}, adminID)
		require.NoError(t, insertErr)
	}

	tests := []struct {
		name       string
		opts       repositories.RecordQueryOptions
		wantCount  int
		wantTotal  int64
		firstField string
	}{
		{
			name:       "最初のページ",
			opts:       repositories.RecordQueryOptions{Page: 1, Limit: 5},
			wantCount:  5,
			wantTotal:  10,
			firstField: "Record J", // デフォルトの順序はid DESC
		},
		{
			name:       "2番目のページ",
			opts:       repositories.RecordQueryOptions{Page: 2, Limit: 5},
			wantCount:  5,
			wantTotal:  10,
			firstField: "Record E",
		},
		{
			name:       "優先度昇順でソート",
			opts:       repositories.RecordQueryOptions{Page: 1, Limit: 3, Sort: "priority", Order: "asc"},
			wantCount:  3,
			wantTotal:  10,
			firstField: "Record A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			records, total, err := executor.GetRecords(ctx, "app_data_getrecords", fields, tt.opts)
			require.NoError(t, err)
			assert.Len(t, records, tt.wantCount)
			assert.Equal(t, tt.wantTotal, total)
			assert.Equal(t, tt.firstField, records[0].Data["name"])
		})
	}
}

func TestDynamicQueryExecutor_GetRecords_WithFilters(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	executor := repositories.NewDynamicQueryExecutor(db)
	adminID := getAdminUserID(ctx, t)

	// テーブルを作成
	fields := []models.AppField{
		{FieldCode: "category", FieldName: "Category", FieldType: "TEXT"},
		{FieldCode: "amount", FieldName: "Amount", FieldType: "NUMBER"},
	}
	require.NoError(t, executor.CreateTable(ctx, "app_data_filter", fields))

	// テストレコードを挿入
	testData := []models.RecordData{
		{"category": "A", "amount": 100},
		{"category": "A", "amount": 200},
		{"category": "B", "amount": 150},
		{"category": "B", "amount": 250},
		{"category": "C", "amount": 300},
	}
	for _, data := range testData {
		_, insertErr := executor.InsertRecord(ctx, "app_data_filter", data, adminID)
		require.NoError(t, insertErr)
	}

	tests := []struct {
		name      string
		filters   []models.FilterItem
		wantCount int64
	}{
		{
			name:      "カテゴリがeqでフィルタ",
			filters:   []models.FilterItem{{Field: "category", Operator: "eq", Value: "A"}},
			wantCount: 2,
		},
		{
			name:      "金額がgtでフィルタ",
			filters:   []models.FilterItem{{Field: "amount", Operator: "gt", Value: "200"}},
			wantCount: 2,
		},
		{
			name:      "金額がgteでフィルタ",
			filters:   []models.FilterItem{{Field: "amount", Operator: "gte", Value: "200"}},
			wantCount: 3,
		},
		{
			name:      "複数フィルタ",
			filters:   []models.FilterItem{{Field: "category", Operator: "eq", Value: "B"}, {Field: "amount", Operator: "lt", Value: "200"}},
			wantCount: 1,
		},
		{
			name:      "likeでフィルタ",
			filters:   []models.FilterItem{{Field: "category", Operator: "like", Value: "A"}},
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := repositories.RecordQueryOptions{
				Page:    1,
				Limit:   10,
				Filters: tt.filters,
			}
			_, total, err := executor.GetRecords(ctx, "app_data_filter", fields, opts)
			require.NoError(t, err)
			assert.Equal(t, tt.wantCount, total)
		})
	}
}

func TestDynamicQueryExecutor_GetRecordByID(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	executor := repositories.NewDynamicQueryExecutor(db)
	adminID := getAdminUserID(ctx, t)

	// テーブルを作成
	fields := []models.AppField{
		{FieldCode: "title", FieldName: "Title", FieldType: "TEXT"},
	}
	require.NoError(t, executor.CreateTable(ctx, "app_data_getbyid", fields))

	// レコードを挿入
	recordID, err := executor.InsertRecord(ctx, "app_data_getbyid", models.RecordData{"title": "Test Record"}, adminID)
	require.NoError(t, err)

	tests := []struct {
		name       string
		recordID   uint64
		wantRecord bool
		wantTitle  string
	}{
		{
			name:       "存在するレコード",
			recordID:   recordID,
			wantRecord: true,
			wantTitle:  "Test Record",
		},
		{
			name:       "存在しないレコード",
			recordID:   99999,
			wantRecord: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := executor.GetRecordByID(ctx, "app_data_getbyid", fields, tt.recordID)
			require.NoError(t, err)
			if tt.wantRecord {
				require.NotNil(t, result)
				assert.Equal(t, tt.wantTitle, result.Data["title"])
				assert.Equal(t, adminID, result.CreatedBy)
				assert.NotEmpty(t, result.CreatedAt)
				assert.NotEmpty(t, result.UpdatedAt)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

func TestDynamicQueryExecutor_GetAggregatedData(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	executor := repositories.NewDynamicQueryExecutor(db)
	adminID := getAdminUserID(ctx, t)

	// テーブルを作成
	fields := []models.AppField{
		{FieldCode: "category", FieldName: "Category", FieldType: "TEXT"},
		{FieldCode: "amount", FieldName: "Amount", FieldType: "NUMBER"},
	}
	require.NoError(t, executor.CreateTable(ctx, "app_data_aggregate", fields))

	// テストレコードを挿入
	testData := []models.RecordData{
		{"category": "A", "amount": 100},
		{"category": "A", "amount": 200},
		{"category": "B", "amount": 150},
		{"category": "B", "amount": 250},
		{"category": "C", "amount": 300},
	}
	for _, data := range testData {
		_, insertErr := executor.InsertRecord(ctx, "app_data_aggregate", data, adminID)
		require.NoError(t, insertErr)
	}

	tests := []struct {
		name           string
		request        *models.ChartDataRequest
		expectedLabels []string
		expectedData   []float64
	}{
		{
			name: "count集計",
			request: &models.ChartDataRequest{
				XAxis: models.ChartAxis{Field: "category"},
				YAxis: models.ChartAxis{Aggregation: "count", Label: "Count"},
			},
			expectedLabels: []string{"A", "B", "C"},
			expectedData:   []float64{2, 2, 1},
		},
		{
			name: "sum集計",
			request: &models.ChartDataRequest{
				XAxis: models.ChartAxis{Field: "category"},
				YAxis: models.ChartAxis{Field: "amount", Aggregation: "sum", Label: "Total"},
			},
			expectedLabels: []string{"A", "B", "C"},
			expectedData:   []float64{300, 400, 300},
		},
		{
			name: "avg集計",
			request: &models.ChartDataRequest{
				XAxis: models.ChartAxis{Field: "category"},
				YAxis: models.ChartAxis{Field: "amount", Aggregation: "avg", Label: "Average"},
			},
			expectedLabels: []string{"A", "B", "C"},
			expectedData:   []float64{150, 200, 300},
		},
		{
			name: "max集計",
			request: &models.ChartDataRequest{
				XAxis: models.ChartAxis{Field: "category"},
				YAxis: models.ChartAxis{Field: "amount", Aggregation: "max", Label: "Max"},
			},
			expectedLabels: []string{"A", "B", "C"},
			expectedData:   []float64{200, 250, 300},
		},
		{
			name: "min集計",
			request: &models.ChartDataRequest{
				XAxis: models.ChartAxis{Field: "category"},
				YAxis: models.ChartAxis{Field: "amount", Aggregation: "min", Label: "Min"},
			},
			expectedLabels: []string{"A", "B", "C"},
			expectedData:   []float64{100, 150, 300},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := executor.GetAggregatedData(ctx, "app_data_aggregate", tt.request)
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tt.expectedLabels, result.Labels)
			require.Len(t, result.Datasets, 1)
			assert.Equal(t, tt.expectedData, result.Datasets[0].Data)
		})
	}
}

func TestDynamicQueryExecutor_GetAggregatedData_WithFilter(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	executor := repositories.NewDynamicQueryExecutor(db)
	adminID := getAdminUserID(ctx, t)

	// テーブルを作成
	fields := []models.AppField{
		{FieldCode: "category", FieldName: "Category", FieldType: "TEXT"},
		{FieldCode: "amount", FieldName: "Amount", FieldType: "NUMBER"},
	}
	require.NoError(t, executor.CreateTable(ctx, "app_data_agg_filter", fields))

	// テストレコードを挿入
	testData := []models.RecordData{
		{"category": "A", "amount": 100},
		{"category": "A", "amount": 200},
		{"category": "B", "amount": 150},
	}
	for _, data := range testData {
		_, insertErr := executor.InsertRecord(ctx, "app_data_agg_filter", data, adminID)
		require.NoError(t, insertErr)
	}

	// フィルタ付きで集計データを取得
	request := &models.ChartDataRequest{
		XAxis:   models.ChartAxis{Field: "category"},
		YAxis:   models.ChartAxis{Aggregation: "count", Label: "Count"},
		Filters: []models.FilterItem{{Field: "category", Operator: "eq", Value: "A"}},
	}

	result, err := executor.GetAggregatedData(ctx, "app_data_agg_filter", request)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, []string{"A"}, result.Labels)
	assert.Equal(t, []float64{2}, result.Datasets[0].Data)
}

func TestValidateIdentifier(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"有効なシンプルな名前", "my_table", false},
		{"数字を含む有効な名前", "table_123", false},
		{"アンダースコア開始の有効な名前", "_private", false},
		{"空", "", true},
		{"数字で開始", "123table", true},
		{"ダッシュを含む", "my-table", true},
		{"スペースを含む", "my table", true},
		{"セミコロンを含む", "table;", true},
		{"SQLインジェクション試行", "users; DROP TABLE", true},
		{"長すぎる", string(make([]byte, 65)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repositories.ValidateIdentifier(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDynamicQueryExecutor_FieldTypes(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	executor := repositories.NewDynamicQueryExecutor(db)
	adminID := getAdminUserID(ctx, t)

	// 全フィールドタイプを含むテーブルを作成
	fields := []models.AppField{
		{FieldCode: "text_field", FieldName: "Text", FieldType: "TEXT"},
		{FieldCode: "number_field", FieldName: "Number", FieldType: "NUMBER"},
		{FieldCode: "checkbox_field", FieldName: "Checkbox", FieldType: "CHECKBOX"},
		{FieldCode: "date_field", FieldName: "Date", FieldType: "DATE"},
		{FieldCode: "datetime_field", FieldName: "DateTime", FieldType: "DATETIME"},
		{FieldCode: "dropdown_field", FieldName: "Dropdown", FieldType: "DROPDOWN"},
		{FieldCode: "radio_field", FieldName: "Radio", FieldType: "RADIO"},
		{FieldCode: "textarea_field", FieldName: "Textarea", FieldType: "TEXTAREA"},
	}
	require.NoError(t, executor.CreateTable(ctx, "app_data_fieldtypes", fields))

	// 様々なフィールドタイプのレコードを挿入
	now := time.Now().Format("2006-01-02")
	nowTime := time.Now().Format("2006-01-02 15:04:05")

	data := models.RecordData{
		"text_field":     "Sample text",
		"number_field":   42.5,
		"checkbox_field": true,
		"date_field":     now,
		"datetime_field": nowTime,
		"dropdown_field": "option1",
		"radio_field":    "choice1",
		"textarea_field": "Long text content\nwith multiple lines",
	}

	recordID, err := executor.InsertRecord(ctx, "app_data_fieldtypes", data, adminID)
	require.NoError(t, err)

	// 取得して検証
	record, err := executor.GetRecordByID(ctx, "app_data_fieldtypes", fields, recordID)
	require.NoError(t, err)
	require.NotNil(t, record)

	assert.Equal(t, "Sample text", record.Data["text_field"])
	// 数値フィールドはデータベーススキャン後に文字列として返される
	assert.Equal(t, "42.5", record.Data["number_field"])
	assert.Equal(t, "1", record.Data["checkbox_field"]) // MySQLはスキャン後にTINYINTを文字列として返す
	assert.Equal(t, now, record.Data["date_field"])
}
