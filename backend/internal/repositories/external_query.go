package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	_ "github.com/denisenkom/go-mssqldb" // SQL Server driver
	_ "github.com/go-sql-driver/mysql"   // MySQL driver
	_ "github.com/lib/pq"                // PostgreSQL driver
	_ "github.com/sijms/go-ora/v2"       // Oracle driver (Pure Go)

	"nocode-app/backend/internal/models"
)

// ExternalQueryExecutor 外部データベースへのクエリ実行を処理する構造体
type ExternalQueryExecutor struct{}

// NewExternalQueryExecutor 新しいExternalQueryExecutorを作成する
func NewExternalQueryExecutor() *ExternalQueryExecutor {
	return &ExternalQueryExecutor{}
}

// buildDSN データソース情報からDSN文字列を構築する
func buildDSN(ds *models.DataSource, password string) (string, string, error) {
	switch ds.DBType {
	case models.DBTypePostgreSQL:
		// PostgreSQLはキーワード形式を使用（特殊文字のエスケープが不要）
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			ds.Host, ds.Port, ds.Username, escapePostgresPassword(password), ds.DatabaseName)
		return "postgres", dsn, nil

	case models.DBTypeMySQL:
		// MySQLはDSN形式を使用
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
			ds.Username, password, ds.Host, ds.Port, ds.DatabaseName)
		return "mysql", dsn, nil

	case models.DBTypeOracle:
		// go-ora v2 format: oracle://user:pass@host:port/service_name
		// パスワードとユーザー名はURLエンコードが必要
		dsn := fmt.Sprintf("oracle://%s:%s@%s:%d/%s",
			url.QueryEscape(ds.Username),
			url.QueryEscape(password),
			ds.Host,
			ds.Port,
			ds.DatabaseName)
		return "oracle", dsn, nil

	case models.DBTypeSQLServer:
		// SQL ServerはURLエンコードが必要
		dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
			url.QueryEscape(ds.Username),
			url.QueryEscape(password),
			ds.Host,
			ds.Port,
			ds.DatabaseName)
		return "sqlserver", dsn, nil

	default:
		return "", "", fmt.Errorf("サポートされていないデータベースタイプ: %s", ds.DBType)
	}
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
	db, err := openConnection(ctx, ds, password)
	if err != nil {
		return nil, err
	}
	defer func() { _ = db.Close() }()

	var query string
	switch ds.DBType {
	case models.DBTypePostgreSQL:
		query = `SELECT table_name, table_schema,
			CASE WHEN table_type = 'BASE TABLE' THEN 'TABLE' ELSE 'VIEW' END as table_type
			FROM information_schema.tables 
			WHERE table_schema NOT IN ('pg_catalog', 'information_schema') 
			AND table_type IN ('BASE TABLE', 'VIEW')
			ORDER BY table_schema, table_name`

	case models.DBTypeMySQL:
		query = `SELECT table_name, table_schema,
			CASE WHEN table_type = 'BASE TABLE' THEN 'TABLE' ELSE 'VIEW' END as table_type
			FROM information_schema.tables 
			WHERE table_schema = DATABASE() 
			AND table_type IN ('BASE TABLE', 'VIEW')
			ORDER BY table_name`

	case models.DBTypeOracle:
		// OracleはUNION ALLでテーブルとViewを結合
		query = `SELECT table_name, owner as table_schema, 'TABLE' as table_type
			FROM all_tables 
			WHERE owner = USER
			UNION ALL
			SELECT view_name as table_name, owner as table_schema, 'VIEW' as table_type
			FROM all_views
			WHERE owner = USER
			ORDER BY 1`

	case models.DBTypeSQLServer:
		query = `SELECT table_name, table_schema,
			CASE WHEN table_type = 'BASE TABLE' THEN 'TABLE' ELSE 'VIEW' END as table_type
			FROM information_schema.tables 
			WHERE table_type IN ('BASE TABLE', 'VIEW')
			ORDER BY table_schema, table_name`

	default:
		return nil, fmt.Errorf("サポートされていないデータベースタイプ: %s", ds.DBType)
	}

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
	db, err := openConnection(ctx, ds, password)
	if err != nil {
		return nil, err
	}
	defer func() { _ = db.Close() }()

	var query string
	var args []interface{}

	switch ds.DBType {
	case models.DBTypePostgreSQL:
		query = `SELECT 
			c.column_name,
			c.data_type,
			CASE WHEN c.is_nullable = 'YES' THEN true ELSE false END as is_nullable,
			CASE WHEN tc.constraint_type = 'PRIMARY KEY' THEN true ELSE false END as is_primary_key,
			COALESCE(c.column_default, '') as default_value
		FROM information_schema.columns c
		LEFT JOIN information_schema.key_column_usage kcu 
			ON c.table_name = kcu.table_name 
			AND c.column_name = kcu.column_name
		LEFT JOIN information_schema.table_constraints tc 
			ON kcu.constraint_name = tc.constraint_name 
			AND tc.constraint_type = 'PRIMARY KEY'
		WHERE c.table_name = $1
		ORDER BY c.ordinal_position`
		args = []interface{}{tableName}

	case models.DBTypeMySQL:
		query = `SELECT 
			column_name,
			data_type,
			CASE WHEN is_nullable = 'YES' THEN true ELSE false END as is_nullable,
			CASE WHEN column_key = 'PRI' THEN true ELSE false END as is_primary_key,
			COALESCE(column_default, '') as default_value
		FROM information_schema.columns
		WHERE table_schema = DATABASE() AND table_name = ?
		ORDER BY ordinal_position`
		args = []interface{}{tableName}

	case models.DBTypeOracle:
		// DATA_DEFAULTはLONG型のため、TO_CHARは使用できない
		// 代わりに空文字を返す（デフォルト値は必須ではない）
		query = `SELECT 
			c.column_name,
			c.data_type,
			CASE WHEN c.nullable = 'Y' THEN 1 ELSE 0 END as is_nullable,
			CASE WHEN cc.constraint_type = 'P' THEN 1 ELSE 0 END as is_primary_key,
			'' as default_value
		FROM all_tab_columns c
		LEFT JOIN (
			SELECT acc.column_name, ac.constraint_type
			FROM all_cons_columns acc
			JOIN all_constraints ac ON acc.constraint_name = ac.constraint_name
			WHERE ac.constraint_type = 'P' AND acc.table_name = :1 AND acc.owner = USER
		) cc ON c.column_name = cc.column_name
		WHERE c.table_name = :2 AND c.owner = USER
		ORDER BY c.column_id`
		args = []interface{}{tableName, tableName}

	case models.DBTypeSQLServer:
		query = `SELECT 
			c.COLUMN_NAME,
			c.DATA_TYPE,
			CASE WHEN c.IS_NULLABLE = 'YES' THEN 1 ELSE 0 END as is_nullable,
			CASE WHEN pk.COLUMN_NAME IS NOT NULL THEN 1 ELSE 0 END as is_primary_key,
			ISNULL(c.COLUMN_DEFAULT, '') as default_value
		FROM INFORMATION_SCHEMA.COLUMNS c
		LEFT JOIN (
			SELECT ku.COLUMN_NAME, ku.TABLE_NAME
			FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS tc
			JOIN INFORMATION_SCHEMA.KEY_COLUMN_USAGE ku 
				ON tc.CONSTRAINT_NAME = ku.CONSTRAINT_NAME
			WHERE tc.CONSTRAINT_TYPE = 'PRIMARY KEY'
		) pk ON c.TABLE_NAME = pk.TABLE_NAME AND c.COLUMN_NAME = pk.COLUMN_NAME
		WHERE c.TABLE_NAME = @p1
		ORDER BY c.ORDINAL_POSITION`
		args = []interface{}{tableName}

	default:
		return nil, fmt.Errorf("サポートされていないデータベースタイプ: %s", ds.DBType)
	}

	rows, err := db.QueryContext(ctx, query, args...)
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

	// カラムリストを構築（source_column_nameを使用）
	columns := make([]string, 0, len(fields))
	for _, f := range fields {
		colName := f.FieldCode
		if f.SourceColumnName != nil && *f.SourceColumnName != "" {
			colName = *f.SourceColumnName
		}
		columns = append(columns, quoteIdentifierForDB(ds.DBType, colName))
	}

	// COUNT クエリ
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", quoteIdentifierForDB(ds.DBType, tableName))
	var total int64
	if err := db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("レコード数の取得に失敗しました: %w", err)
	}

	// メインクエリ
	query := fmt.Sprintf("SELECT %s FROM %s",
		strings.Join(columns, ", "),
		quoteIdentifierForDB(ds.DBType, tableName))

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
		order := "ASC"
		if opts.Order == "desc" {
			order = "DESC"
		}
		query += fmt.Sprintf(" ORDER BY %s %s", quoteIdentifierForDB(ds.DBType, sortCol), order)
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

	// カラムリストを構築
	columns := make([]string, 0, len(fields))
	for _, f := range fields {
		colName := f.FieldCode
		if f.SourceColumnName != nil && *f.SourceColumnName != "" {
			colName = *f.SourceColumnName
		}
		columns = append(columns, quoteIdentifierForDB(ds.DBType, colName))
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

	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = %s",
		strings.Join(columns, ", "),
		quoteIdentifierForDB(ds.DBType, tableName),
		quoteIdentifierForDB(ds.DBType, pkColumn),
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
			// source_column_nameがない場合はfield_codeをそのまま使用
			fieldCodeToColumn[f.FieldCode] = f.FieldCode
		}
	}

	// X軸のカラム名を取得
	xColumnName, ok := fieldCodeToColumn[req.XAxis.Field]
	if !ok {
		return nil, fmt.Errorf("x-axis field '%s' not found", req.XAxis.Field)
	}
	xField := quoteIdentifierForDB(ds.DBType, xColumnName)

	var selectClause string
	switch req.YAxis.Aggregation {
	case "count":
		selectClause = fmt.Sprintf("%s, COUNT(*) as value", xField)
	case "sum", "avg", "min", "max":
		// Y軸のカラム名を取得
		yColumnName, ok := fieldCodeToColumn[req.YAxis.Field]
		if !ok {
			return nil, fmt.Errorf("y-axis field '%s' not found", req.YAxis.Field)
		}
		yField := quoteIdentifierForDB(ds.DBType, yColumnName)
		selectClause = fmt.Sprintf("%s, %s(%s) as value", xField, strings.ToUpper(req.YAxis.Aggregation), yField)
	default:
		selectClause = fmt.Sprintf("%s, COUNT(*) as value", xField)
	}

	query := fmt.Sprintf("SELECT %s FROM %s GROUP BY %s ORDER BY %s",
		selectClause,
		quoteIdentifierForDB(ds.DBType, tableName),
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

	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", quoteIdentifierForDB(ds.DBType, tableName))
	var count int64
	if err := db.QueryRowContext(ctx, query).Scan(&count); err != nil {
		return 0, fmt.Errorf("レコード数の取得に失敗しました: %w", err)
	}
	return count, nil
}

// quoteIdentifierForDB データベースタイプに応じた識別子のクォート
func quoteIdentifierForDB(dbType models.DBType, name string) string {
	switch dbType {
	case models.DBTypePostgreSQL:
		return fmt.Sprintf(`"%s"`, strings.ReplaceAll(name, `"`, `""`))
	case models.DBTypeMySQL:
		return fmt.Sprintf("`%s`", strings.ReplaceAll(name, "`", "``"))
	case models.DBTypeOracle:
		return fmt.Sprintf(`"%s"`, strings.ReplaceAll(strings.ToUpper(name), `"`, `""`))
	case models.DBTypeSQLServer:
		return fmt.Sprintf("[%s]", strings.ReplaceAll(name, "]", "]]"))
	default:
		return name
	}
}

// getPlaceholder データベースタイプに応じたプレースホルダーを返す
func getPlaceholder(dbType models.DBType, index int) string {
	switch dbType {
	case models.DBTypePostgreSQL:
		return fmt.Sprintf("$%d", index)
	case models.DBTypeOracle:
		return fmt.Sprintf(":%d", index)
	case models.DBTypeSQLServer:
		return fmt.Sprintf("@p%d", index)
	default: // MySQL
		return "?"
	}
}

// buildLimitOffset データベースタイプに応じたLIMIT/OFFSET句を構築する
func buildLimitOffset(dbType models.DBType, limit, offset int) string {
	switch dbType {
	case models.DBTypeOracle:
		return fmt.Sprintf(" OFFSET %d ROWS FETCH NEXT %d ROWS ONLY", offset, limit)
	case models.DBTypeSQLServer:
		// SQL Serverの場合、ORDER BYが必要
		return fmt.Sprintf(" OFFSET %d ROWS FETCH NEXT %d ROWS ONLY", offset, limit)
	default: // PostgreSQL, MySQL
		return fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)
	}
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
