package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver

	"nocode-app/backend/internal/models"
)

// externalIdentifierRegex 外部DBの識別子（テーブル名・カラム名）として許可する文字を定義する。
// PostgreSQL のクォート済み識別子は日本語や絵文字などの任意の Unicode 文字を含められるため、
// 内部テーブル用の identifierRegex（ASCII 限定の許可リスト）のように文字種を狭めず、
// 「クォート済み識別子として安全に表現できない文字」＝制御文字（ヌルバイトを含む C0 制御 0x00-0x1f
// および DEL 0x7f）のみを拒否するデナイリストとする。ダブルクォート自体は quoteIdentifierForDB 内で
// `"` を `""` にエスケープして無害化する。
var externalIdentifierRegex = regexp.MustCompile(`^[^\x00-\x1f\x7f]+$`)

// maxExternalIdentifierLength 外部DB識別子の最大長（バイト）
const maxExternalIdentifierLength = 128

// ExternalQueryExecutor 外部データベースへのクエリ実行を処理する構造体
type ExternalQueryExecutor struct{}

// NewExternalQueryExecutor 新しいExternalQueryExecutorを作成する
func NewExternalQueryExecutor() *ExternalQueryExecutor {
	return &ExternalQueryExecutor{}
}

// buildDSN データソース情報からDSN文字列を構築する (PostgreSQL のみ対応)
func buildDSN(ds *models.DataSource, password string) (string, string, error) {
	if ds.DBType != models.DBTypePostgreSQL {
		return "", "", fmt.Errorf("サポートされていないデータベースタイプ: %s", ds.DBType)
	}
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		ds.Host, ds.Port, ds.Username, escapePostgresPassword(password), ds.DatabaseName)
	return "postgres", dsn, nil
}

// escapePostgresPassword PostgreSQLのパスワードをエスケープする
func escapePostgresPassword(password string) string {
	// PostgreSQLのキーワード形式では、シングルクォートとバックスラッシュをエスケープ
	escaped := strings.ReplaceAll(password, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `'`, `\'`)
	return escaped
}

// openConnection 外部データベースへの接続を開く
func openConnection(ctx context.Context, ds *models.DataSource, password string) (*sql.DB, error) {
	driverName, dsn, err := buildDSN(ds, password)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("データベース接続の初期化に失敗しました: %w", err)
	}

	// 接続プールの設定
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(5 * time.Minute)

	// 接続テスト
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("データベースへの接続に失敗しました: %w", err)
	}

	return db, nil
}

// TestConnection データベースへのテスト接続を実行する
func (e *ExternalQueryExecutor) TestConnection(ctx context.Context, ds *models.DataSource, password string) error {
	db, err := openConnection(ctx, ds, password)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	return nil
}

// GetTables データベースのテーブル一覧を取得する（テーブルとViewの両方を含む）
func (e *ExternalQueryExecutor) GetTables(ctx context.Context, ds *models.DataSource, password string) ([]models.TableInfo, error) {
	if ds.DBType != models.DBTypePostgreSQL {
		return nil, fmt.Errorf("サポートされていないデータベースタイプ: %s", ds.DBType)
	}

	db, err := openConnection(ctx, ds, password)
	if err != nil {
		return nil, err
	}
	defer func() { _ = db.Close() }()

	const query = `SELECT table_name, table_schema,
			CASE WHEN table_type = 'BASE TABLE' THEN 'TABLE' ELSE 'VIEW' END as table_type
			FROM information_schema.tables
			WHERE table_schema NOT IN ('pg_catalog', 'information_schema')
			AND table_type IN ('BASE TABLE', 'VIEW')
			ORDER BY table_schema, table_name`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("テーブル一覧の取得に失敗しました: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var tables []models.TableInfo
	for rows.Next() {
		var table models.TableInfo
		var tableType string
		if err := rows.Scan(&table.Name, &table.Schema, &tableType); err != nil {
			return nil, fmt.Errorf("テーブル情報のスキャンに失敗しました: %w", err)
		}
		table.Type = models.TableType(tableType)
		tables = append(tables, table)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tables, nil
}

// GetColumns テーブルのカラム一覧を取得する
func (e *ExternalQueryExecutor) GetColumns(ctx context.Context, ds *models.DataSource, password string, tableName string) ([]models.ColumnInfo, error) {
	if ds.DBType != models.DBTypePostgreSQL {
		return nil, fmt.Errorf("サポートされていないデータベースタイプ: %s", ds.DBType)
	}

	db, err := openConnection(ctx, ds, password)
	if err != nil {
		return nil, err
	}
	defer func() { _ = db.Close() }()

	// スキーマ未指定なら接続中の検索パスのスキーマ (current_schema()) で絞る。
	// マルチスキーマ DB で同名テーブルが存在しても他スキーマのカラムや PK 制約を
	// 拾わないようにする。
	const query = `SELECT
			c.column_name,
			c.data_type,
			CASE WHEN c.is_nullable = 'YES' THEN true ELSE false END as is_nullable,
			CASE WHEN tc.constraint_type = 'PRIMARY KEY' THEN true ELSE false END as is_primary_key,
			COALESCE(c.column_default, '') as default_value
		FROM information_schema.columns c
		LEFT JOIN information_schema.key_column_usage kcu
			ON c.table_schema = kcu.table_schema
			AND c.table_name = kcu.table_name
			AND c.column_name = kcu.column_name
		LEFT JOIN information_schema.table_constraints tc
			ON kcu.table_schema = tc.table_schema
			AND kcu.constraint_name = tc.constraint_name
			AND tc.constraint_type = 'PRIMARY KEY'
		WHERE c.table_name = $1
			AND c.table_schema = current_schema()
		ORDER BY c.ordinal_position`

	rows, err := db.QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, fmt.Errorf("カラム一覧の取得に失敗しました: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var columns []models.ColumnInfo
	for rows.Next() {
		var col models.ColumnInfo
		var defaultValue sql.NullString
		if err := rows.Scan(&col.Name, &col.DataType, &col.IsNullable, &col.IsPrimaryKey, &defaultValue); err != nil {
			return nil, fmt.Errorf("カラム情報のスキャンに失敗しました: %w", err)
		}
		if defaultValue.Valid {
			col.DefaultValue = defaultValue.String
		}
		columns = append(columns, col)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return columns, nil
}

// GetRecords 外部テーブルからレコードを取得する
func (e *ExternalQueryExecutor) GetRecords(ctx context.Context, ds *models.DataSource, password string, tableName string, fields []models.AppField, opts RecordQueryOptions) ([]models.RecordResponse, int64, error) {
	db, err := openConnection(ctx, ds, password)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = db.Close() }()

	// テーブル名を検証してクォート
	quotedTable, err := quoteIdentifierForDB(tableName)
	if err != nil {
		return nil, 0, fmt.Errorf("無効なテーブル名: %w", err)
	}

	// カラムリストを構築（source_column_nameを使用）
	columns := make([]string, 0, len(fields))
	for _, f := range fields {
		colName := f.FieldCode
		if f.SourceColumnName != nil && *f.SourceColumnName != "" {
			colName = *f.SourceColumnName
		}
		quotedCol, colErr := quoteIdentifierForDB(colName)
		if colErr != nil {
			return nil, 0, fmt.Errorf("無効なカラム名 %q: %w", colName, colErr)
		}
		columns = append(columns, quotedCol)
	}

	// COUNT クエリ
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", quotedTable)
	var total int64
	if err := db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("レコード数の取得に失敗しました: %w", err)
	}

	// メインクエリ
	query := fmt.Sprintf("SELECT %s FROM %s",
		strings.Join(columns, ", "),
		quotedTable)

	// ORDER BY
	if opts.Sort != "" {
		sortCol := opts.Sort
		// フィールドからsource_column_nameを取得
		for _, f := range fields {
			if f.FieldCode == opts.Sort {
				if f.SourceColumnName != nil && *f.SourceColumnName != "" {
					sortCol = *f.SourceColumnName
				}
				break
			}
		}
		quotedSort, sortErr := quoteIdentifierForDB(sortCol)
		if sortErr != nil {
			return nil, 0, fmt.Errorf("無効なソートカラム %q: %w", sortCol, sortErr)
		}
		order := "ASC"
		if opts.Order == "desc" {
			order = "DESC"
		}
		query += fmt.Sprintf(" ORDER BY %s %s", quotedSort, order)
	}

	// LIMIT/OFFSET
	offset := (opts.Page - 1) * opts.Limit
	query += buildLimitOffset(ds.DBType, opts.Limit, offset)

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, 0, fmt.Errorf("レコードの取得に失敗しました: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var records []models.RecordResponse
	for rows.Next() {
		record, scanErr := scanExternalRecordRow(rows, fields)
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

// GetRecordByID 外部テーブルから単一のレコードを取得する
func (e *ExternalQueryExecutor) GetRecordByID(ctx context.Context, ds *models.DataSource, password string, tableName string, fields []models.AppField, recordID uint64) (*models.RecordResponse, error) {
	db, err := openConnection(ctx, ds, password)
	if err != nil {
		return nil, err
	}
	defer func() { _ = db.Close() }()

	// テーブル名を検証してクォート
	quotedTable, err := quoteIdentifierForDB(tableName)
	if err != nil {
		return nil, fmt.Errorf("無効なテーブル名: %w", err)
	}

	// カラムリストを構築
	columns := make([]string, 0, len(fields))
	for _, f := range fields {
		colName := f.FieldCode
		if f.SourceColumnName != nil && *f.SourceColumnName != "" {
			colName = *f.SourceColumnName
		}
		quotedCol, colErr := quoteIdentifierForDB(colName)
		if colErr != nil {
			return nil, fmt.Errorf("無効なカラム名 %q: %w", colName, colErr)
		}
		columns = append(columns, quotedCol)
	}

	// PKカラムを特定（最初のフィールドまたはidカラムを使用）
	pkColumn := "id"
	for _, f := range fields {
		sourceColName := ""
		if f.SourceColumnName != nil {
			sourceColName = *f.SourceColumnName
		}
		if sourceColName == "id" || f.FieldCode == "id" {
			if sourceColName != "" {
				pkColumn = sourceColName
			}
			break
		}
	}

	quotedPK, err := quoteIdentifierForDB(pkColumn)
	if err != nil {
		return nil, fmt.Errorf("無効な主キーカラム %q: %w", pkColumn, err)
	}

	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = %s",
		strings.Join(columns, ", "),
		quotedTable,
		quotedPK,
		getPlaceholder(ds.DBType, 1))

	row := db.QueryRowContext(ctx, query, recordID)

	// フィールド値をスキャン
	fieldValues := make([]interface{}, len(fields))
	fieldPtrs := make([]interface{}, len(fields))
	for i := range fieldValues {
		fieldPtrs[i] = &fieldValues[i]
	}

	if err := row.Scan(fieldPtrs...); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("レコードのスキャンに失敗しました: %w", err)
	}

	// レコードデータを構築
	data := make(models.RecordData)
	for i, f := range fields {
		data[f.FieldCode] = convertScannedValue(fieldValues[i])
	}

	return &models.RecordResponse{
		ID:   recordID,
		Data: data,
	}, nil
}

// GetAggregatedData 外部テーブルから集計データを取得する
func (e *ExternalQueryExecutor) GetAggregatedData(ctx context.Context, ds *models.DataSource, password string, tableName string, fields []models.AppField, req *models.ChartDataRequest) (*models.ChartDataResponse, error) {
	db, err := openConnection(ctx, ds, password)
	if err != nil {
		return nil, err
	}
	defer func() { _ = db.Close() }()

	// field_codeからsource_column_nameへのマッピングを構築
	fieldCodeToColumn := make(map[string]string)
	for _, f := range fields {
		if f.SourceColumnName != nil {
			fieldCodeToColumn[f.FieldCode] = *f.SourceColumnName
		} else {
			fieldCodeToColumn[f.FieldCode] = f.FieldCode
		}
	}

	xColumnName, ok := fieldCodeToColumn[req.XAxis.Field]
	if !ok {
		return nil, fmt.Errorf("x-axis field '%s' not found", req.XAxis.Field)
	}
	xField, err := quoteIdentifierForDB(xColumnName)
	if err != nil {
		return nil, fmt.Errorf("無効なX軸フィールド %q: %w", xColumnName, err)
	}

	var selectClause string
	switch req.YAxis.Aggregation {
	case "count":
		selectClause = fmt.Sprintf("%s, COUNT(*) as value", xField)
	case "sum", "avg", "min", "max":
		yColumnName, ok := fieldCodeToColumn[req.YAxis.Field]
		if !ok {
			return nil, fmt.Errorf("y-axis field '%s' not found", req.YAxis.Field)
		}
		yField, yErr := quoteIdentifierForDB(yColumnName)
		if yErr != nil {
			return nil, fmt.Errorf("無効なY軸フィールド %q: %w", yColumnName, yErr)
		}
		selectClause = fmt.Sprintf("%s, %s(%s) as value", xField, strings.ToUpper(req.YAxis.Aggregation), yField)
	default:
		selectClause = fmt.Sprintf("%s, COUNT(*) as value", xField)
	}

	quotedTable, err := quoteIdentifierForDB(tableName)
	if err != nil {
		return nil, fmt.Errorf("無効なテーブル名: %w", err)
	}

	query := fmt.Sprintf("SELECT %s FROM %s GROUP BY %s ORDER BY %s",
		selectClause,
		quotedTable,
		xField,
		xField)

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("集計データの取得に失敗しました: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var labels []string
	var data []float64

	for rows.Next() {
		var label interface{}
		var value float64
		if err := rows.Scan(&label, &value); err != nil {
			return nil, err
		}

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
				Label: req.YAxis.Label,
				Data:  data,
			},
		},
	}, nil
}

// CountRecords 外部テーブルのレコード数を取得する
func (e *ExternalQueryExecutor) CountRecords(ctx context.Context, ds *models.DataSource, password string, tableName string) (int64, error) {
	db, err := openConnection(ctx, ds, password)
	if err != nil {
		return 0, err
	}
	defer func() { _ = db.Close() }()

	quotedTable, err := quoteIdentifierForDB(tableName)
	if err != nil {
		return 0, fmt.Errorf("無効なテーブル名: %w", err)
	}

	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", quotedTable)
	var count int64
	if err := db.QueryRowContext(ctx, query).Scan(&count); err != nil {
		return 0, fmt.Errorf("レコード数の取得に失敗しました: %w", err)
	}
	return count, nil
}

// quoteIdentifierForDB 識別子を検証してダブルクォートでクォートする (PostgreSQL)。
//
// externalIdentifierRegex（制御文字を拒否するデナイリスト）による MatchString ガードを
// 本関数内に直接置くことで、戻り値（クォート済み識別子）のデータフロー上にサニタイザバリアを乗せ、
// 静的解析（CodeQL go/sql-injection 等）が「検証済みの識別子のみがクエリへ流れる」ことを
// 認識できるようにする。検証に失敗した識別子はクエリに使わずエラーを返す。
// 識別子に含まれうるダブルクォートは `"` を `""` にエスケープして無害化する。
func quoteIdentifierForDB(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("識別子を空にすることはできません")
	}
	if len(name) > maxExternalIdentifierLength {
		return "", fmt.Errorf("識別子が長すぎます: 最大%dバイト", maxExternalIdentifierLength)
	}
	if !externalIdentifierRegex.MatchString(name) {
		return "", fmt.Errorf("無効な識別子: 制御文字（ヌルバイト等）を含めることはできません")
	}
	// ダブルクォートを二重化してエスケープし、クォート済み識別子を破壊できないようにする
	escaped := strings.ReplaceAll(name, `"`, `""`)
	return `"` + escaped + `"`, nil
}

// getPlaceholder PostgreSQL の $N プレースホルダを返す。
func getPlaceholder(_ models.DBType, index int) string {
	return fmt.Sprintf("$%d", index)
}

// buildLimitOffset PostgreSQL の LIMIT/OFFSET 句を構築する。
func buildLimitOffset(_ models.DBType, limit, offset int) string {
	return fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)
}

// scanExternalRecordRow 外部DBの行からレコードをスキャンする
func scanExternalRecordRow(rows *sql.Rows, fields []models.AppField) (*models.RecordResponse, error) {
	fieldValues := make([]interface{}, len(fields))
	fieldPtrs := make([]interface{}, len(fields))
	for i := range fieldValues {
		fieldPtrs[i] = &fieldValues[i]
	}

	if err := rows.Scan(fieldPtrs...); err != nil {
		return nil, fmt.Errorf("レコードのスキャンに失敗しました: %w", err)
	}

	data := make(models.RecordData)
	for i, f := range fields {
		data[f.FieldCode] = convertScannedValue(fieldValues[i])
	}

	// IDを取得（idフィールドがあれば）
	var id uint64
	if idVal, ok := data["id"]; ok {
		switch v := idVal.(type) {
		case int64:
			id = uint64(v)
		case uint64:
			id = v
		case float64:
			id = uint64(v)
		}
	}

	return &models.RecordResponse{
		ID:   id,
		Data: data,
	}, nil
}
