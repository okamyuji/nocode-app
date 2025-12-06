package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/uptrace/bun"

	"nocode-app/backend/internal/models"
)

// identifierRegex SQL識別子（テーブル名、カラム名）を検証する正規表現
// 英数字とアンダースコアのみ許可、先頭は文字またはアンダースコアで始まる必要がある
var identifierRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// maxIdentifierLength SQL識別子の最大長
const maxIdentifierLength = 64

// ValidateIdentifier 文字列が安全なSQL識別子かどうかを検証する
func ValidateIdentifier(name string) error {
	if name == "" {
		return fmt.Errorf("識別子を空にすることはできません")
	}
	if len(name) > maxIdentifierLength {
		return fmt.Errorf("識別子が長すぎます: 最大%d文字", maxIdentifierLength)
	}
	if !identifierRegex.MatchString(name) {
		return fmt.Errorf("無効な識別子: 英数字とアンダースコアのみ許可、先頭は文字またはアンダースコア")
	}
	return nil
}

// quoteIdentifier 検証後にSQL識別子を安全にクォートする
func quoteIdentifier(name string) (string, error) {
	if err := ValidateIdentifier(name); err != nil {
		return "", err
	}
	// バッククォートをエスケープ（正規表現で防止されているが念のため）
	escaped := strings.ReplaceAll(name, "`", "``")
	return "`" + escaped + "`", nil
}

// DynamicQueryExecutor 動的テーブル操作を処理する構造体
type DynamicQueryExecutor struct {
	db *bun.DB
}

// NewDynamicQueryExecutor 新しいDynamicQueryExecutorを作成する
func NewDynamicQueryExecutor(db *bun.DB) *DynamicQueryExecutor {
	return &DynamicQueryExecutor{db: db}
}

// CreateTable アプリ用の動的テーブルを作成する
func (e *DynamicQueryExecutor) CreateTable(ctx context.Context, tableName string, fields []models.AppField) error {
	quotedTable, err := quoteIdentifier(tableName)
	if err != nil {
		return fmt.Errorf("無効なテーブル名: %w", err)
	}

	// カラムスライスを事前確保: 1 id + len(fields) + 3 メタデータ
	columns := make([]string, 0, len(fields)+4)

	// 基本カラム
	columns = append(columns, "id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY")

	// フィールドからの動的カラム
	for i := range fields {
		quotedCol, colErr := quoteIdentifier(fields[i].FieldCode)
		if colErr != nil {
			return fmt.Errorf("無効なカラム名 %q: %w", fields[i].FieldCode, colErr)
		}
		colDef := fmt.Sprintf("%s %s", quotedCol, fields[i].GetMySQLColumnType())
		columns = append(columns, colDef)
	}

	// メタデータカラム
	columns = append(columns,
		"created_by BIGINT UNSIGNED NOT NULL",
		"created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP",
		"updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP",
	)

	query := fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS %s (%s) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci",
		quotedTable,
		strings.Join(columns, ", "),
	)

	_, err = e.db.ExecContext(ctx, query)
	return err
}

// DropTable 動的テーブルを削除する
func (e *DynamicQueryExecutor) DropTable(ctx context.Context, tableName string) error {
	quotedTable, err := quoteIdentifier(tableName)
	if err != nil {
		return fmt.Errorf("無効なテーブル名: %w", err)
	}

	query := fmt.Sprintf("DROP TABLE IF EXISTS %s", quotedTable)
	_, err = e.db.ExecContext(ctx, query)
	return err
}

// AddColumn 動的テーブルにカラムを追加する
func (e *DynamicQueryExecutor) AddColumn(ctx context.Context, tableName string, field *models.AppField) error {
	quotedTable, err := quoteIdentifier(tableName)
	if err != nil {
		return fmt.Errorf("無効なテーブル名: %w", err)
	}

	quotedCol, err := quoteIdentifier(field.FieldCode)
	if err != nil {
		return fmt.Errorf("無効なカラム名: %w", err)
	}

	query := fmt.Sprintf(
		"ALTER TABLE %s ADD COLUMN %s %s",
		quotedTable,
		quotedCol,
		field.GetMySQLColumnType(),
	)
	_, err = e.db.ExecContext(ctx, query)
	return err
}

// DropColumn 動的テーブルからカラムを削除する
func (e *DynamicQueryExecutor) DropColumn(ctx context.Context, tableName, columnName string) error {
	quotedTable, err := quoteIdentifier(tableName)
	if err != nil {
		return fmt.Errorf("無効なテーブル名: %w", err)
	}

	quotedCol, err := quoteIdentifier(columnName)
	if err != nil {
		return fmt.Errorf("無効なカラム名: %w", err)
	}

	query := fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", quotedTable, quotedCol)
	_, err = e.db.ExecContext(ctx, query)
	return err
}

// InsertRecord 動的テーブルにレコードを挿入する
func (e *DynamicQueryExecutor) InsertRecord(ctx context.Context, tableName string, data models.RecordData, userID uint64) (uint64, error) {
	quotedTable, err := quoteIdentifier(tableName)
	if err != nil {
		return 0, fmt.Errorf("無効なテーブル名: %w", err)
	}

	columns := []string{"created_by"}
	placeholders := []string{"?"}
	values := []interface{}{userID}

	for key, value := range data {
		quotedCol, colErr := quoteIdentifier(key)
		if colErr != nil {
			return 0, fmt.Errorf("無効なカラム名 %q: %w", key, colErr)
		}
		columns = append(columns, quotedCol)
		placeholders = append(placeholders, "?")
		values = append(values, value)
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		quotedTable,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	result, err := e.db.ExecContext(ctx, query, values...)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return uint64(id), nil
}

// UpdateRecord 動的テーブルのレコードを更新する
func (e *DynamicQueryExecutor) UpdateRecord(ctx context.Context, tableName string, recordID uint64, data models.RecordData) error {
	quotedTable, err := quoteIdentifier(tableName)
	if err != nil {
		return fmt.Errorf("無効なテーブル名: %w", err)
	}

	// スライスを事前確保
	setClauses := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data)+1)

	for key, value := range data {
		quotedCol, colErr := quoteIdentifier(key)
		if colErr != nil {
			return fmt.Errorf("無効なカラム名 %q: %w", key, colErr)
		}
		setClauses = append(setClauses, quotedCol+" = ?")
		values = append(values, value)
	}

	values = append(values, recordID)

	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE id = ?",
		quotedTable,
		strings.Join(setClauses, ", "),
	)

	_, err = e.db.ExecContext(ctx, query, values...)
	return err
}

// DeleteRecord 動的テーブルからレコードを削除する
func (e *DynamicQueryExecutor) DeleteRecord(ctx context.Context, tableName string, recordID uint64) error {
	quotedTable, err := quoteIdentifier(tableName)
	if err != nil {
		return fmt.Errorf("無効なテーブル名: %w", err)
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", quotedTable)
	_, err = e.db.ExecContext(ctx, query, recordID)
	return err
}

// DeleteRecords 動的テーブルから複数のレコードを削除する
func (e *DynamicQueryExecutor) DeleteRecords(ctx context.Context, tableName string, recordIDs []uint64) error {
	if len(recordIDs) == 0 {
		return nil
	}

	quotedTable, err := quoteIdentifier(tableName)
	if err != nil {
		return fmt.Errorf("無効なテーブル名: %w", err)
	}

	placeholders := make([]string, len(recordIDs))
	values := make([]interface{}, len(recordIDs))
	for i, id := range recordIDs {
		placeholders[i] = "?"
		values[i] = id
	}

	query := fmt.Sprintf(
		"DELETE FROM %s WHERE id IN (%s)",
		quotedTable,
		strings.Join(placeholders, ", "),
	)

	_, err = e.db.ExecContext(ctx, query, values...)
	return err
}

// RecordQueryOptions レコードクエリのオプションを保持する構造体
type RecordQueryOptions struct {
	Page    int
	Limit   int
	Sort    string
	Order   string
	Filters []models.FilterItem
}

// GetRecords ページネーションとフィルタリング付きで動的テーブルからレコードを取得する
func (e *DynamicQueryExecutor) GetRecords(ctx context.Context, tableName string, fields []models.AppField, opts RecordQueryOptions) ([]models.RecordResponse, int64, error) {
	quotedTable, err := quoteIdentifier(tableName)
	if err != nil {
		return nil, 0, fmt.Errorf("無効なテーブル名: %w", err)
	}

	// カラムリストとWHERE句を構築
	columns, err := e.buildColumnList(fields)
	if err != nil {
		return nil, 0, err
	}

	whereSQL, whereValues, err := e.buildWhereClause(opts.Filters)
	if err != nil {
		return nil, 0, err
	}

	// 総件数を取得
	total, err := e.getRecordCount(ctx, quotedTable, whereSQL, whereValues)
	if err != nil {
		return nil, 0, err
	}

	// ORDER BY句を構築
	orderBy, err := e.buildOrderBy(opts.Sort, opts.Order)
	if err != nil {
		return nil, 0, err
	}

	// メインクエリを構築して実行
	return e.executeRecordsQuery(ctx, quotedTable, columns, whereSQL, whereValues, orderBy, opts, fields, total)
}

// buildColumnList SELECTカラムリストを構築する
func (e *DynamicQueryExecutor) buildColumnList(fields []models.AppField) ([]string, error) {
	columns := make([]string, 0, len(fields)+4)
	columns = append(columns, "id", "created_by", "created_at", "updated_at")

	for i := range fields {
		quotedCol, colErr := quoteIdentifier(fields[i].FieldCode)
		if colErr != nil {
			return nil, fmt.Errorf("無効なカラム名 %q: %w", fields[i].FieldCode, colErr)
		}
		columns = append(columns, quotedCol)
	}
	return columns, nil
}

// buildWhereClause フィルターからWHERE句を構築する
func (e *DynamicQueryExecutor) buildWhereClause(filters []models.FilterItem) (whereSQL string, whereValues []interface{}, err error) {
	if len(filters) == 0 {
		return "", nil, nil
	}

	whereClauses := make([]string, 0, len(filters))
	whereValues = make([]interface{}, 0, len(filters))

	for _, filter := range filters {
		clause, value, filterErr := buildFilterClause(filter)
		if filterErr != nil {
			return "", nil, filterErr
		}
		if clause != "" {
			whereClauses = append(whereClauses, clause)
			whereValues = append(whereValues, value)
		}
	}

	if len(whereClauses) == 0 {
		return "", nil, nil
	}
	return "WHERE " + strings.Join(whereClauses, " AND "), whereValues, nil
}

// getRecordCount レコードの総件数を取得する
func (e *DynamicQueryExecutor) getRecordCount(ctx context.Context, quotedTable, whereSQL string, whereValues []interface{}) (int64, error) {
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", quotedTable, whereSQL)
	var total int64
	err := e.db.QueryRowContext(ctx, countQuery, whereValues...).Scan(&total)
	return total, err
}

// buildOrderBy ORDER BY句を構築する
func (e *DynamicQueryExecutor) buildOrderBy(sort, order string) (string, error) {
	if sort == "" {
		return "id DESC", nil
	}

	quotedSort, sortErr := quoteIdentifier(sort)
	if sortErr != nil {
		return "", fmt.Errorf("無効なソートカラム: %w", sortErr)
	}

	orderDir := "ASC"
	if order == "desc" {
		orderDir = "DESC"
	}
	return fmt.Sprintf("%s %s", quotedSort, orderDir), nil
}

// executeRecordsQuery メインレコードクエリを実行する
func (e *DynamicQueryExecutor) executeRecordsQuery(
	ctx context.Context,
	quotedTable string,
	columns []string,
	whereSQL string,
	whereValues []interface{},
	orderBy string,
	opts RecordQueryOptions,
	fields []models.AppField,
	total int64,
) ([]models.RecordResponse, int64, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM %s %s ORDER BY %s LIMIT ? OFFSET ?",
		strings.Join(columns, ", "),
		quotedTable,
		whereSQL,
		orderBy,
	)

	offset := (opts.Page - 1) * opts.Limit
	args := make([]interface{}, 0, len(whereValues)+2)
	args = append(args, whereValues...)
	args = append(args, opts.Limit, offset)

	rows, err := e.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	var records []models.RecordResponse
	for rows.Next() {
		record, scanErr := scanRecordRow(rows, fields)
		if scanErr != nil {
			return nil, 0, scanErr
		}
		records = append(records, *record)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

// GetRecordByID IDで単一のレコードを取得する
func (e *DynamicQueryExecutor) GetRecordByID(ctx context.Context, tableName string, fields []models.AppField, recordID uint64) (*models.RecordResponse, error) {
	quotedTable, err := quoteIdentifier(tableName)
	if err != nil {
		return nil, fmt.Errorf("無効なテーブル名: %w", err)
	}

	// カラムリストを構築
	columns := make([]string, 0, len(fields)+4)
	columns = append(columns, "id", "created_by", "created_at", "updated_at")

	for i := range fields {
		quotedCol, colErr := quoteIdentifier(fields[i].FieldCode)
		if colErr != nil {
			return nil, fmt.Errorf("無効なカラム名 %q: %w", fields[i].FieldCode, colErr)
		}
		columns = append(columns, quotedCol)
	}

	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE id = ?",
		strings.Join(columns, ", "),
		quotedTable,
	)

	row := e.db.QueryRowContext(ctx, query, recordID)
	return scanSingleRecordRow(row, fields)
}

// buildFilterClause 単一のフィルター句を構築する
func buildFilterClause(filter models.FilterItem) (clause string, value interface{}, err error) {
	quotedCol, err := quoteIdentifier(filter.Field)
	if err != nil {
		return "", nil, fmt.Errorf("無効なフィルターフィールド %q: %w", filter.Field, err)
	}

	switch filter.Operator {
	case "eq":
		return quotedCol + " = ?", filter.Value, nil
	case "ne":
		return quotedCol + " != ?", filter.Value, nil
	case "gt":
		return quotedCol + " > ?", filter.Value, nil
	case "gte":
		return quotedCol + " >= ?", filter.Value, nil
	case "lt":
		return quotedCol + " < ?", filter.Value, nil
	case "lte":
		return quotedCol + " <= ?", filter.Value, nil
	case "like":
		return quotedCol + " LIKE ?", "%" + filter.Value + "%", nil
	default:
		return "", nil, nil
	}
}

// scanRecordRow 行からレコードをスキャンする
func scanRecordRow(rows *sql.Rows, fields []models.AppField) (*models.RecordResponse, error) {
	var id, createdBy uint64
	var createdAt, updatedAt time.Time

	// フィールド値をスキャンするスライスを作成
	fieldValues := make([]interface{}, len(fields))
	fieldPtrs := make([]interface{}, len(fields))
	for i := range fieldValues {
		fieldPtrs[i] = &fieldValues[i]
	}

	// スキャン先を構築
	scanDest := []interface{}{&id, &createdBy, &createdAt, &updatedAt}
	scanDest = append(scanDest, fieldPtrs...)

	if err := rows.Scan(scanDest...); err != nil {
		return nil, err
	}

	// レコードデータを構築
	data := make(models.RecordData)
	for i := range fields {
		data[fields[i].FieldCode] = convertScannedValue(fieldValues[i])
	}

	return &models.RecordResponse{
		ID:        id,
		Data:      data,
		CreatedBy: createdBy,
		CreatedAt: createdAt.Format(time.RFC3339),
		UpdatedAt: updatedAt.Format(time.RFC3339),
	}, nil
}

// scanSingleRecordRow 単一行からレコードをスキャンする
func scanSingleRecordRow(row *sql.Row, fields []models.AppField) (*models.RecordResponse, error) {
	var id, createdBy uint64
	var createdAt, updatedAt time.Time

	// フィールド値をスキャンするスライスを作成
	fieldValues := make([]interface{}, len(fields))
	fieldPtrs := make([]interface{}, len(fields))
	for i := range fieldValues {
		fieldPtrs[i] = &fieldValues[i]
	}

	// スキャン先を構築
	scanDest := []interface{}{&id, &createdBy, &createdAt, &updatedAt}
	scanDest = append(scanDest, fieldPtrs...)

	if err := row.Scan(scanDest...); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// レコードデータを構築
	data := make(models.RecordData)
	for i := range fields {
		data[fields[i].FieldCode] = convertScannedValue(fieldValues[i])
	}

	return &models.RecordResponse{
		ID:        id,
		Data:      data,
		CreatedBy: createdBy,
		CreatedAt: createdAt.Format(time.RFC3339),
		UpdatedAt: updatedAt.Format(time.RFC3339),
	}, nil
}

// convertScannedValue スキャンした値を変換する
func convertScannedValue(v interface{}) interface{} {
	if v == nil {
		return nil
	}

	switch val := v.(type) {
	case []byte:
		return string(val)
	case time.Time:
		return val.Format(time.RFC3339)
	default:
		return val
	}
}

// GetAggregatedData チャート用の集計データを取得する
func (e *DynamicQueryExecutor) GetAggregatedData(ctx context.Context, tableName string, req *models.ChartDataRequest) (*models.ChartDataResponse, error) {
	quotedTable, err := quoteIdentifier(tableName)
	if err != nil {
		return nil, fmt.Errorf("無効なテーブル名: %w", err)
	}

	quotedXField, err := quoteIdentifier(req.XAxis.Field)
	if err != nil {
		return nil, fmt.Errorf("無効なX軸フィールド: %w", err)
	}

	// SELECT句を構築
	selectClause, err := e.buildAggregationSelect(quotedXField, req.YAxis)
	if err != nil {
		return nil, err
	}

	// フィルターからWHERE句を構築
	whereSQL, whereValues, err := e.buildWhereClause(req.Filters)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(
		"SELECT %s FROM %s %s GROUP BY %s ORDER BY %s",
		selectClause,
		quotedTable,
		whereSQL,
		quotedXField,
		quotedXField,
	)

	rows, err := e.db.QueryContext(ctx, query, whereValues...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return e.scanAggregatedData(rows, req.YAxis.Label)
}

// buildAggregationSelect 集計用のSELECT句を構築する
func (e *DynamicQueryExecutor) buildAggregationSelect(quotedXField string, yAxis models.ChartAxis) (string, error) {
	switch yAxis.Aggregation {
	case "count":
		return fmt.Sprintf("%s, COUNT(*) as value", quotedXField), nil
	case "sum", "avg", "min", "max":
		quotedYField, yErr := quoteIdentifier(yAxis.Field)
		if yErr != nil {
			return "", fmt.Errorf("無効なY軸フィールド: %w", yErr)
		}
		aggFunc := strings.ToUpper(yAxis.Aggregation)
		return fmt.Sprintf("%s, %s(%s) as value", quotedXField, aggFunc, quotedYField), nil
	default:
		return fmt.Sprintf("%s, COUNT(*) as value", quotedXField), nil
	}
}

// scanAggregatedData 行から集計データをスキャンする
func (e *DynamicQueryExecutor) scanAggregatedData(rows *sql.Rows, yLabel string) (*models.ChartDataResponse, error) {
	var labels []string
	var data []float64

	for rows.Next() {
		var label interface{}
		var value float64
		if scanErr := rows.Scan(&label, &value); scanErr != nil {
			return nil, scanErr
		}

		// ラベルを文字列に変換
		var labelStr string
		switch v := label.(type) {
		case nil:
			labelStr = "(空)"
		case []byte:
			labelStr = string(v)
		case string:
			labelStr = v
		default:
			labelStr = fmt.Sprintf("%v", v)
		}

		labels = append(labels, labelStr)
		data = append(data, value)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &models.ChartDataResponse{
		Labels: labels,
		Datasets: []models.ChartDataset{
			{
				Label: yLabel,
				Data:  data,
			},
		},
	}, nil
}

// CountRecords テーブル内のレコード総数を返す
func (e *DynamicQueryExecutor) CountRecords(ctx context.Context, tableName string) (int64, error) {
	quotedTable, err := quoteIdentifier(tableName)
	if err != nil {
		return 0, fmt.Errorf("無効なテーブル名: %w", err)
	}

	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", quotedTable)
	var count int64
	err = e.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// CountTodaysUpdates テーブル内の本日更新されたレコード数を返す
func (e *DynamicQueryExecutor) CountTodaysUpdates(ctx context.Context, tableName string) (int64, error) {
	quotedTable, err := quoteIdentifier(tableName)
	if err != nil {
		return 0, fmt.Errorf("無効なテーブル名: %w", err)
	}

	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE DATE(updated_at) = CURDATE()", quotedTable)
	var count int64
	err = e.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
