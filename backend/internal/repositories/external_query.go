package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"    // MySQL driver（DSN組み立てにmysql.Configを使うため非ブランクimport）
	_ "github.com/lib/pq"               // PostgreSQL driver
	_ "github.com/microsoft/go-mssqldb" // SQL Server driver
	_ "github.com/sijms/go-ora/v2"      // Oracle driver (Pure Go)

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

// dsnForbiddenHostChars ホスト名に含まれるとDSNの構造を壊しうる文字。
// URL形式のDSN（Oracle/SQL Server）では認証情報・パス・クエリの区切りとして解釈され、
// キーワード形式（PostgreSQL）やmysql.Configでも接続先のすり替えにつながる。
const dsnForbiddenHostChars = "/?@"

// validateDSNHost DSNに埋め込むホスト名を検証する
func validateDSNHost(host string) error {
	if strings.ContainsAny(host, dsnForbiddenHostChars) {
		return fmt.Errorf("ホスト名に使用できない文字が含まれています: %q", host)
	}
	return nil
}

// buildDSN データソース情報からDSN文字列を構築する。
// DatabaseName・ユーザー名・パスワードは利用者入力であるため、各方言のエンコーダ
// （mysql.Config.FormatDSN / url.URL / PostgreSQLの引用符付き値）を通し、
// 接続オプションとして再解釈されないようにする。
func buildDSN(ds *models.DataSource, password string) (string, string, error) {
	if err := validateDSNHost(ds.Host); err != nil {
		return "", "", err
	}
	addr := net.JoinHostPort(ds.Host, strconv.Itoa(ds.Port))

	switch ds.DBType {
	case models.DBTypePostgreSQL:
		// PostgreSQLはキーワード形式を使用する。空白を含む値は次のキーワードとして
		// 解釈されるため、利用者入力である4つの値すべてを引用符で囲む。
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			quotePostgresValue(ds.Host),
			ds.Port,
			quotePostgresValue(ds.Username),
			quotePostgresValue(password),
			quotePostgresValue(ds.DatabaseName))
		return "postgres", dsn, nil

	case models.DBTypeMySQL:
		// mysql.NewConfig()の既定値（AllowNativePasswords=true など）を土台にすることで、
		// 文字列連結していた頃のDSNと挙動を揃える。
		cfg := mysql.NewConfig()
		cfg.User = ds.Username
		cfg.Passwd = password
		cfg.Net = "tcp"
		cfg.Addr = addr
		cfg.DBName = ds.DatabaseName
		cfg.ParseTime = true
		return "mysql", cfg.FormatDSN(), nil

	case models.DBTypeOracle:
		// go-ora v2 format: oracle://user:pass@host:port/service_name
		u := url.URL{
			Scheme: "oracle",
			User:   url.UserPassword(ds.Username, password),
			Host:   addr,
		}
		// RawPathにPathEscape済みの値を入れることで、サービス名中の "/" や "?" も
		// パス区切り・クエリ開始として解釈されない完全エスケープ形になる。
		u.Path = "/" + ds.DatabaseName
		u.RawPath = "/" + url.PathEscape(ds.DatabaseName)
		return "oracle", u.String(), nil

	case models.DBTypeSQLServer:
		// url.Valuesではなくurl.URLに組ませることで、パスワード中の空白が
		// "+"（クエリ表現）ではなく "%20" になり、パーサが空白として復元できる。
		u := url.URL{
			Scheme:   "sqlserver",
			User:     url.UserPassword(ds.Username, password),
			Host:     addr,
			RawQuery: url.Values{"database": {ds.DatabaseName}}.Encode(),
		}
		return "sqlserver", u.String(), nil

	default:
		return "", "", fmt.Errorf("サポートされていないデータベースタイプ: %s", ds.DBType)
	}
}

// quotePostgresValue PostgreSQLのキーワード形式DSN用に値を引用符で囲む。
// libpqは単一引用符で囲まれた値の中の空白を値の一部として扱うため、
// 引用符で囲まない値に空白があると以降が別のキーワードとして解釈される。
func quotePostgresValue(value string) string {
	escaped := strings.ReplaceAll(value, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `'`, `\'`)
	return "'" + escaped + "'"
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
		// スキーマ未指定なら接続中の検索パスのスキーマ (current_schema()) で絞る。
		// マルチスキーマ DB で同名テーブルが存在しても他スキーマのカラムや PK 制約を
		// 拾わないようにする。
		query = `SELECT
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
			JOIN all_constraints ac
				ON ac.owner = acc.owner
				AND ac.constraint_name = acc.constraint_name
				AND ac.table_name = acc.table_name
			WHERE ac.constraint_type = 'P' AND acc.table_name = :1
				AND acc.owner = USER AND ac.owner = USER
		) cc ON c.column_name = cc.column_name
		WHERE c.table_name = :2 AND c.owner = USER
		ORDER BY c.column_id`
		// データディクショナリは非クォート識別子を大文字で保持するため、バインド値も大文字化する。
		// これを怠るとquoteIdentifierForDBが大文字化するGetRecordsは成功するのに
		// カラム一覧だけ空になる、という食い違いが起きる。
		oracleTable := oracleObjectName(tableName)
		args = []interface{}{oracleTable, oracleTable}

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

	// テーブル名を検証してクォート
	quotedTable, err := quoteIdentifierForDB(ds.DBType, tableName)
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
		quotedCol, colErr := quoteIdentifierForDB(ds.DBType, colName)
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

	// ORDER BY の並び替え式（クォート済みカラム + 方向）を組み立てる
	sortSQL := ""
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
		quotedSort, sortErr := quoteIdentifierForDB(ds.DBType, sortCol)
		if sortErr != nil {
			return nil, 0, fmt.Errorf("無効なソートカラム %q: %w", sortCol, sortErr)
		}
		order := "ASC"
		if opts.Order == "desc" {
			order = "DESC"
		}
		sortSQL = fmt.Sprintf("%s %s", quotedSort, order)
	}

	// ORDER BY と LIMIT/OFFSET
	offset := (opts.Page - 1) * opts.Limit
	query += buildOrderAndLimit(ds.DBType, sortSQL, opts.Limit, offset)

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
	quotedTable, err := quoteIdentifierForDB(ds.DBType, tableName)
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
		quotedCol, colErr := quoteIdentifierForDB(ds.DBType, colName)
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

	quotedPK, err := quoteIdentifierForDB(ds.DBType, pkColumn)
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
	xField, err := quoteIdentifierForDB(ds.DBType, xColumnName)
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
		yField, yErr := quoteIdentifierForDB(ds.DBType, yColumnName)
		if yErr != nil {
			return nil, fmt.Errorf("無効なY軸フィールド %q: %w", yColumnName, yErr)
		}
		selectClause = fmt.Sprintf("%s, %s(%s) as value", xField, strings.ToUpper(req.YAxis.Aggregation), yField)
	default:
		selectClause = fmt.Sprintf("%s, COUNT(*) as value", xField)
	}

	quotedTable, err := quoteIdentifierForDB(ds.DBType, tableName)
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

	quotedTable, err := quoteIdentifierForDB(ds.DBType, tableName)
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

// oracleObjectName Oracleのデータディクショナリ（all_tab_columns等）へバインドする
// オブジェクト名を大文字化する。Oracleは非クォート識別子を大文字で格納するため、
// quoteIdentifierForDBのOracle分岐と同じ規則を適用しないと、
// メタデータ検索だけが空振りして実データ取得と結果が食い違う。
func oracleObjectName(name string) string {
	return strings.ToUpper(name)
}

// quoteIdentifierForDB 識別子を検証してデータベースタイプに応じてクォートする。
//
// externalIdentifierRegex（制御文字を拒否するデナイリスト）による MatchString ガードを
// 本関数内に直接置くことで、戻り値（クォート済み識別子）のデータフロー上にサニタイザバリアを乗せ、
// 静的解析（CodeQL go/sql-injection 等）が「検証済みの識別子のみがクエリへ流れる」ことを
// 認識できるようにする。検証に失敗した識別子はクエリに使わずエラーを返す。
// 識別子に含まれうるクォート文字は方言ごとに二重化して無害化する
// （PostgreSQL / Oracle はダブルクォート、MySQL はバッククォート、SQL Server は閉じ角括弧を二重化する）。
func quoteIdentifierForDB(dbType models.DBType, name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("識別子を空にすることはできません")
	}
	if len(name) > maxExternalIdentifierLength {
		return "", fmt.Errorf("識別子が長すぎます: 最大%dバイト", maxExternalIdentifierLength)
	}
	if !externalIdentifierRegex.MatchString(name) {
		return "", fmt.Errorf("無効な識別子: 制御文字（ヌルバイト等）を含めることはできません")
	}

	switch dbType {
	case models.DBTypePostgreSQL:
		return fmt.Sprintf(`"%s"`, strings.ReplaceAll(name, `"`, `""`)), nil
	case models.DBTypeMySQL:
		return fmt.Sprintf("`%s`", strings.ReplaceAll(name, "`", "``")), nil
	case models.DBTypeOracle:
		// Oracleの非クォート識別子は大文字で格納されるため、大文字化してからクォートする
		return fmt.Sprintf(`"%s"`, strings.ReplaceAll(strings.ToUpper(name), `"`, `""`)), nil
	case models.DBTypeSQLServer:
		return fmt.Sprintf("[%s]", strings.ReplaceAll(name, "]", "]]")), nil
	default:
		return "", fmt.Errorf("サポートされていないデータベースタイプ: %s", dbType)
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

// buildOrderAndLimit ORDER BY句とLIMIT/OFFSET句を組み立てる。
// sortSQLは検証・クォート済みの並び替え式（例: `"id" ASC`）で、空なら並び替えを指定しない。
//
// SQL ServerはOFFSET/FETCHをORDER BY無しで受け付けないため、並び替え指定が無い場合だけ
// ORDER BY (SELECT NULL) を補って構文を成立させる。他の方言の出力は変えない。
func buildOrderAndLimit(dbType models.DBType, sortSQL string, limit, offset int) string {
	orderBy := ""
	switch {
	case sortSQL != "":
		orderBy = " ORDER BY " + sortSQL
	case dbType == models.DBTypeSQLServer:
		orderBy = " ORDER BY (SELECT NULL)"
	}
	return orderBy + buildLimitOffset(dbType, limit, offset)
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
		id = parseRecordID(idVal)
	}

	return &models.RecordResponse{
		ID:   id,
		Data: data,
	}, nil
}

// parseRecordID スキャン済みの値をレコードIDへ変換する。変換できない値は0を返す。
//
// go-oraはNUMBER列（IDENTITYのidを含む）をinterface{}にスキャンすると文字列で返すため、
// 数値型だけを見ているとOracleのレコードIDが常に0になる。
func parseRecordID(v interface{}) uint64 {
	switch val := v.(type) {
	case int64:
		if val < 0 {
			return 0
		}
		return uint64(val)
	case int32:
		if val < 0 {
			return 0
		}
		return uint64(val)
	case int:
		if val < 0 {
			return 0
		}
		return uint64(val)
	case uint64:
		return val
	case float64:
		if val < 0 {
			return 0
		}
		return uint64(val)
	case string:
		return parseRecordIDString(val)
	case []byte:
		return parseRecordIDString(string(val))
	default:
		return 0
	}
}

// parseRecordIDString 文字列表現のレコードIDを解釈する。
// go-oraはNUMBERを "1" だけでなく "1.00" のような小数表記で返すことがあるため、
// 整数として読めなければ浮動小数点としても試す。
func parseRecordIDString(s string) uint64 {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return 0
	}
	if n, err := strconv.ParseUint(trimmed, 10, 64); err == nil {
		return n
	}
	if f, err := strconv.ParseFloat(trimmed, 64); err == nil && f >= 0 {
		return uint64(f)
	}
	return 0
}
