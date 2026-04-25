# 実装計画: PostgreSQL 一択化

> 設計計画: `2026-04-25-postgres-only-design.md` 参照

**Goal:** 対応 RDB を PostgreSQL のみに統一し、MySQL/Oracle/MSSQL に関するコードと依存とテストを完全に除去する。

**Architecture:** バックエンド（migrations / database.go / config.go / dynamic_query.go / external_query.go / models / testhelpers）→ フロントエンド（types / components / tests）→ infra（compose / env）→ 品質検証 の順で進める。各タスクは独立に commit する。

**Tech Stack:** Go 1.25 / bun ORM / `lib/pq` / PostgreSQL 16 / React 18 / TypeScript / pnpm / Vite / Vitest

---

## タスク全体マップ

| # | タスク | 概要 | コミット粒度 |
|---|---|---|---|
| 1 | 設計書コミット | `docs/plans/*.md` を最初の commit に固める | 1 |
| 2 | migrations 書き換え | init/002/003 の SQL を PG 化 | 1 |
| 3 | `database.go` 切替 | `pgdialect` + `lib/pq` | 1 |
| 4 | `config.go` DSN 変更 | PostgreSQL DSN | 1 |
| 5 | `field.go` カラム型変更 | `GetPostgresColumnType` | 1 |
| 6 | `dynamic_query.go` 構文変換 | バックティック→`"`、`?`→`$N`、`RETURNING id`、`CURRENT_DATE` | 1 |
| 7 | `models/datasource.go` 縮約 | `DBType` を PG のみに | 1 |
| 8 | `external_query.go` 縮約 | switch/case 削除 | 1 |
| 9 | testhelpers の整理 | mssql/mysql/oracle/external_db_helper を削除、postgres_container にアプリ DB 用 GetTestDB/ResetDatabase を実装 | 1 |
| 10 | テスト書き換え | dbtype 関連テストの縮小 | 1 |
| 11 | go.mod クリーンアップ | 不要 driver / dialect を削除 | 1 |
| 12 | フロント縮約 | `DBType` ユニオン縮約と関連コンポーネント・テスト | 1 |
| 13 | infra 切替 | compose.yaml / env / Dockerfile / README | 1 |
| 14 | 品質ゲート | `go test`, `pnpm test`, `compose up` 動作確認 | （commit せず PR 直前にまとめて） |

---

## Task 1: 設計・実装・移行計画書を初回 commit する

**Files:**
- Create: `docs/plans/2026-04-25-postgres-only-design.md` ✅ (既に作成済)
- Create: `docs/plans/2026-04-25-postgres-only-implementation.md` ✅ (本書)
- Create: `docs/plans/2026-04-25-postgres-only-data-migration.md` ✅ (これから作成)

- [ ] **Step 1.1:** 3 つの計画書がすべて作成済みであることを確認する

```bash
ls docs/plans/2026-04-25-postgres-only-*.md
# 3 ファイルが出力されること
```

- [ ] **Step 1.2:** commit する

```bash
git add docs/plans/2026-04-25-postgres-only-*.md
git commit -m "docs: add PostgreSQL-only refactor design/implementation/migration plans"
```

---

## Task 2: migrations を PostgreSQL 構文に書き換える

**Files:**
- Modify: `backend/migrations/init.sql`
- Modify: `backend/migrations/002_datasources.sql`
- Modify: `backend/migrations/003_dashboard_widgets.sql`

- [ ] **Step 2.1:** `backend/migrations/init.sql` を以下の内容で完全置換する

```sql
-- ノーコードアプリ 初期スキーマ (PostgreSQL)

-- updated_at 自動更新用の共通トリガ関数
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ユーザーテーブル
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(100) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'user' CHECK (role IN ('admin', 'user')),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

CREATE TRIGGER trg_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- データソーステーブル
CREATE TABLE IF NOT EXISTS data_sources (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    db_type VARCHAR(20) NOT NULL CHECK (db_type IN ('postgresql')),
    host VARCHAR(255) NOT NULL,
    port INT NOT NULL,
    database_name VARCHAR(100) NOT NULL,
    username VARCHAR(100) NOT NULL,
    encrypted_password TEXT NOT NULL,
    created_by BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_data_sources_created_by ON data_sources(created_by);

CREATE TRIGGER trg_data_sources_updated_at
    BEFORE UPDATE ON data_sources
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- アプリテーブル
CREATE TABLE IF NOT EXISTS apps (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    table_name VARCHAR(64) NOT NULL UNIQUE,
    icon VARCHAR(50) DEFAULT 'default',
    is_external BOOLEAN NOT NULL DEFAULT FALSE,
    data_source_id BIGINT NULL REFERENCES data_sources(id) ON DELETE SET NULL,
    source_table_name VARCHAR(255) NULL,
    created_by BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_apps_created_by ON apps(created_by);
CREATE INDEX IF NOT EXISTS idx_apps_data_source_id ON apps(data_source_id);

CREATE TRIGGER trg_apps_updated_at
    BEFORE UPDATE ON apps
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- アプリフィールドテーブル
CREATE TABLE IF NOT EXISTS app_fields (
    id BIGSERIAL PRIMARY KEY,
    app_id BIGINT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    field_code VARCHAR(64) NOT NULL,
    field_name VARCHAR(100) NOT NULL,
    field_type VARCHAR(20) NOT NULL,
    source_column_name VARCHAR(255) NULL,
    options JSONB,
    required BOOLEAN NOT NULL DEFAULT FALSE,
    display_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uk_app_field_code UNIQUE (app_id, field_code)
);

CREATE INDEX IF NOT EXISTS idx_app_fields_app_id ON app_fields(app_id);

CREATE TRIGGER trg_app_fields_updated_at
    BEFORE UPDATE ON app_fields
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- アプリビューテーブル
CREATE TABLE IF NOT EXISTS app_views (
    id BIGSERIAL PRIMARY KEY,
    app_id BIGINT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    view_type VARCHAR(20) NOT NULL DEFAULT 'table'
        CHECK (view_type IN ('table', 'list', 'calendar', 'chart')),
    config JSONB,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_app_views_app_id ON app_views(app_id);

CREATE TRIGGER trg_app_views_updated_at
    BEFORE UPDATE ON app_views
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- チャート設定テーブル
CREATE TABLE IF NOT EXISTS chart_configs (
    id BIGSERIAL PRIMARY KEY,
    app_id BIGINT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    chart_type VARCHAR(20) NOT NULL,
    config JSONB NOT NULL,
    created_by BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_chart_configs_app_id ON chart_configs(app_id);

CREATE TRIGGER trg_chart_configs_updated_at
    BEFORE UPDATE ON chart_configs
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ダッシュボードウィジェットテーブル
CREATE TABLE IF NOT EXISTS dashboard_widgets (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    app_id BIGINT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    display_order INT NOT NULL DEFAULT 0,
    view_type VARCHAR(20) NOT NULL DEFAULT 'table'
        CHECK (view_type IN ('table', 'list', 'chart')),
    is_visible BOOLEAN NOT NULL DEFAULT TRUE,
    widget_size VARCHAR(20) NOT NULL DEFAULT 'medium'
        CHECK (widget_size IN ('small', 'medium', 'large')),
    config JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uk_dashboard_user_app UNIQUE (user_id, app_id)
);

CREATE INDEX IF NOT EXISTS idx_dashboard_widgets_user_id ON dashboard_widgets(user_id);
CREATE INDEX IF NOT EXISTS idx_dashboard_widgets_app_id ON dashboard_widgets(app_id);
CREATE INDEX IF NOT EXISTS idx_dashboard_widgets_user_order ON dashboard_widgets(user_id, display_order);

CREATE TRIGGER trg_dashboard_widgets_updated_at
    BEFORE UPDATE ON dashboard_widgets
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- デフォルト管理者ユーザーを挿入（パスワード: admin123）
INSERT INTO users (email, password_hash, name, role) VALUES
('admin@example.com', '$2a$10$e8i3egbnenpqzZlow/3Q0.5L6uN8vNyktEYkgRdWwP13xSkCtR1re', 'Admin', 'admin')
ON CONFLICT (email) DO NOTHING;
```

- [ ] **Step 2.2:** `backend/migrations/002_datasources.sql` は init.sql で吸収済みなので **空の no-op スクリプト** に書き換える（履歴用に残す）

```sql
-- 002: 当初は MySQL 用に分離されていたが、PostgreSQL 一択化に伴い
-- init.sql に統合済み。このスクリプトは互換目的で残置されているのみ。
SELECT 1;
```

- [ ] **Step 2.3:** `backend/migrations/003_dashboard_widgets.sql` も同様

```sql
-- 003: dashboard_widgets は init.sql に統合済み。
SELECT 1;
```

- [ ] **Step 2.4:** commit する

```bash
git add backend/migrations/
git commit -m "refactor(db): rewrite migrations for PostgreSQL"
```

---

## Task 3: `database.go` を pgdialect + lib/pq に切替

**Files:**
- Modify: `backend/internal/database/database.go`

- [ ] **Step 3.1:** ファイルを以下で完全置換する

```go
package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	// PostgreSQL ドライバー
	_ "github.com/lib/pq"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"nocode-app/backend/internal/config"
)

// NewDB 新しいデータベース接続を作成する
func NewDB(cfg *config.DBConfig) (*bun.DB, error) {
	sqldb, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("データベースのオープンに失敗しました: %w", err)
	}

	sqldb.SetMaxOpenConns(cfg.MaxOpenConns)
	sqldb.SetMaxIdleConns(cfg.MaxIdleConns)
	sqldb.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqldb.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("データベースへのPingに失敗しました: %w", err)
	}

	db := bun.NewDB(sqldb, pgdialect.New())

	return db, nil
}

// Close データベース接続を閉じる
func Close(db *bun.DB) error {
	return db.Close()
}
```

- [ ] **Step 3.2:** commit する

```bash
git add backend/internal/database/database.go
git commit -m "refactor(db): switch bun dialect to pgdialect + lib/pq"
```

---

## Task 4: `config.go` の DSN を PostgreSQL 形式に変更

**Files:**
- Modify: `backend/internal/config/config.go`

- [ ] **Step 4.1:** `DSN()` を書き換える

`backend/internal/config/config.go` 内の以下を置換:

旧:
```go
// DSN MySQLのデータソース名を返す
func (c *DBConfig) DSN() string {
	return c.User + ":" + c.Password + "@tcp(" + c.Host + ":" + c.Port + ")/" + c.Name + "?charset=utf8mb4&parseTime=True&loc=Local"
}
```

新:
```go
// DSN PostgreSQLのデータソース名を返す
func (c *DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Password, c.Name,
	)
}
```

`fmt` の import を追加（既存になければ）。

- [ ] **Step 4.2:** デフォルトの `DB_PORT` 値を `3306` から `5432` に変更

旧: `Port: getEnv("DB_PORT", "3306"),`
新: `Port: getEnv("DB_PORT", "5432"),`

- [ ] **Step 4.3:** commit する

```bash
git add backend/internal/config/config.go
git commit -m "refactor(db): use PostgreSQL DSN format and default port 5432"
```

---

## Task 5: `field.go` のカラム型を PG 化

**Files:**
- Modify: `backend/internal/models/field.go`
- Modify: `backend/internal/models/field_test.go`

- [ ] **Step 5.1:** `field.go` の 30〜33 行付近の定数を置換

旧:
```go
// MySQLカラム型の定数
const (
	mysqlVarchar255 = "VARCHAR(255)"
)
```

新:
```go
// PostgreSQLカラム型の定数
const (
	pgVarchar255 = "VARCHAR(255)"
)
```

- [ ] **Step 5.2:** 関数名変更 `GetMySQLColumnType` → `GetPostgresColumnType`、戻り値も PG 型へ

旧 (147〜175):
```go
// GetMySQLColumnType このフィールドのMySQLカラム型を返す
func (f *AppField) GetMySQLColumnType() string {
	switch FieldType(f.FieldType) {
	case FieldTypeText:
		return mysqlVarchar255
	case FieldTypeTextArea:
		return "TEXT"
	case FieldTypeNumber:
		return "DECIMAL(18,4)"
	case FieldTypeDate:
		return "DATE"
	case FieldTypeDateTime:
		return "DATETIME"
	case FieldTypeSelect:
		return mysqlVarchar255
	case FieldTypeMultiSelect:
		return "JSON"
	case FieldTypeCheckbox:
		return "BOOLEAN"
	case FieldTypeRadio:
		return mysqlVarchar255
	case FieldTypeLink:
		return "VARCHAR(500)"
	case FieldTypeAttachment:
		return "JSON"
	default:
		return mysqlVarchar255
	}
}
```

新:
```go
// GetPostgresColumnType このフィールドのPostgreSQLカラム型を返す
func (f *AppField) GetPostgresColumnType() string {
	switch FieldType(f.FieldType) {
	case FieldTypeText:
		return pgVarchar255
	case FieldTypeTextArea:
		return "TEXT"
	case FieldTypeNumber:
		return "NUMERIC(18,4)"
	case FieldTypeDate:
		return "DATE"
	case FieldTypeDateTime:
		return "TIMESTAMP"
	case FieldTypeSelect:
		return pgVarchar255
	case FieldTypeMultiSelect:
		return "JSONB"
	case FieldTypeCheckbox:
		return "BOOLEAN"
	case FieldTypeRadio:
		return pgVarchar255
	case FieldTypeLink:
		return "VARCHAR(500)"
	case FieldTypeAttachment:
		return "JSONB"
	default:
		return pgVarchar255
	}
}
```

- [ ] **Step 5.3:** `field_test.go` の `TestAppField_GetMySQLColumnType` を `TestAppField_GetPostgresColumnType` にリネームし、期待値を上の表に合わせる（`DATETIME` → `TIMESTAMP`、`JSON` → `JSONB`、`DECIMAL(18,4)` → `NUMERIC(18,4)`）。テストの呼び出し側 `field.GetMySQLColumnType()` を `field.GetPostgresColumnType()` に書き換える

- [ ] **Step 5.4:** commit する

```bash
git add backend/internal/models/field.go backend/internal/models/field_test.go
git commit -m "refactor(models): rename GetMySQLColumnType to GetPostgresColumnType and switch types"
```

---

## Task 6: `dynamic_query.go` を PostgreSQL 構文に変換

**Files:**
- Modify: `backend/internal/repositories/dynamic_query.go`

このファイルが最大の書き換え対象なので、5 つのサブステップに分ける。

- [ ] **Step 6.1:** 識別子クォートをバックティック→ダブルクォート

`quoteIdentifier` 関数 (37〜45 行) を置換:

```go
// quoteIdentifier 検証後にSQL識別子を安全にクォートする (PostgreSQL: ダブルクォート)
func quoteIdentifier(name string) (string, error) {
	if err := ValidateIdentifier(name); err != nil {
		return "", err
	}
	escaped := strings.ReplaceAll(name, `"`, `""`)
	return `"` + escaped + `"`, nil
}
```

- [ ] **Step 6.2:** `CreateTable` を PostgreSQL 構文に

`CreateTable` (57〜95 行) を置換:

```go
func (e *DynamicQueryExecutor) CreateTable(ctx context.Context, tableName string, fields []models.AppField) error {
	quotedTable, err := quoteIdentifier(tableName)
	if err != nil {
		return fmt.Errorf("無効なテーブル名: %w", err)
	}

	columns := make([]string, 0, len(fields)+4)
	columns = append(columns, `id BIGSERIAL PRIMARY KEY`)

	for i := range fields {
		quotedCol, colErr := quoteIdentifier(fields[i].FieldCode)
		if colErr != nil {
			return fmt.Errorf("無効なカラム名 %q: %w", fields[i].FieldCode, colErr)
		}
		colDef := fmt.Sprintf("%s %s", quotedCol, fields[i].GetPostgresColumnType())
		columns = append(columns, colDef)
	}

	columns = append(columns,
		`created_by BIGINT NOT NULL`,
		`created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP`,
		`updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP`,
	)

	query := fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS %s (%s)",
		quotedTable,
		strings.Join(columns, ", "),
	)

	if _, err := e.db.ExecContext(ctx, query); err != nil {
		return err
	}

	// updated_at 自動更新トリガを張る
	triggerName := fmt.Sprintf("trg_dyn_updated_at_%s", tableName)
	triggerSQL := fmt.Sprintf(
		`CREATE TRIGGER %s BEFORE UPDATE ON %s FOR EACH ROW EXECUTE FUNCTION set_updated_at()`,
		`"`+strings.ReplaceAll(triggerName, `"`, `""`)+`"`,
		quotedTable,
	)
	_, err = e.db.ExecContext(ctx, triggerSQL)
	return err
}
```

- [ ] **Step 6.3:** `AddColumn` の `GetMySQLColumnType` 呼び出しを修正

旧 (125 行付近): `field.GetMySQLColumnType()` → 新: `field.GetPostgresColumnType()`

- [ ] **Step 6.4:** `InsertRecord` を `RETURNING id` ＆ `$N` プレースホルダ化

`InsertRecord` (148〜187 行) を置換:

```go
func (e *DynamicQueryExecutor) InsertRecord(ctx context.Context, tableName string, data models.RecordData, userID uint64) (uint64, error) {
	quotedTable, err := quoteIdentifier(tableName)
	if err != nil {
		return 0, fmt.Errorf("無効なテーブル名: %w", err)
	}

	columns := []string{"created_by"}
	placeholders := []string{"$1"}
	values := []interface{}{userID}

	idx := 2
	for key, value := range data {
		quotedCol, colErr := quoteIdentifier(key)
		if colErr != nil {
			return 0, fmt.Errorf("無効なカラム名 %q: %w", key, colErr)
		}
		columns = append(columns, quotedCol)
		placeholders = append(placeholders, fmt.Sprintf("$%d", idx))
		values = append(values, value)
		idx++
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s) RETURNING id",
		quotedTable,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	var id uint64
	if err := e.db.QueryRowContext(ctx, query, values...).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}
```

- [ ] **Step 6.5:** `UpdateRecord`, `DeleteRecord`, `DeleteRecords` を `$N` プレースホルダに

`UpdateRecord` (189〜219 行) を置換:

```go
func (e *DynamicQueryExecutor) UpdateRecord(ctx context.Context, tableName string, recordID uint64, data models.RecordData) error {
	quotedTable, err := quoteIdentifier(tableName)
	if err != nil {
		return fmt.Errorf("無効なテーブル名: %w", err)
	}

	setClauses := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data)+1)

	idx := 1
	for key, value := range data {
		quotedCol, colErr := quoteIdentifier(key)
		if colErr != nil {
			return fmt.Errorf("無効なカラム名 %q: %w", key, colErr)
		}
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", quotedCol, idx))
		values = append(values, value)
		idx++
	}

	values = append(values, recordID)

	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE id = $%d",
		quotedTable,
		strings.Join(setClauses, ", "),
		idx,
	)

	_, err = e.db.ExecContext(ctx, query, values...)
	return err
}
```

`DeleteRecord` (222〜231 行) を置換:

```go
func (e *DynamicQueryExecutor) DeleteRecord(ctx context.Context, tableName string, recordID uint64) error {
	quotedTable, err := quoteIdentifier(tableName)
	if err != nil {
		return fmt.Errorf("無効なテーブル名: %w", err)
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", quotedTable)
	_, err = e.db.ExecContext(ctx, query, recordID)
	return err
}
```

`DeleteRecords` (234〜259 行) を置換:

```go
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
		placeholders[i] = fmt.Sprintf("$%d", i+1)
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
```

- [ ] **Step 6.6:** `GetRecords` 系のクエリを `$N` 化

`GetRecords` (271〜302 行) では `LIMIT ? OFFSET ?` と WHERE 句があるので、`buildWhereClause` を `$N` 番号付きに作り直す。

`buildWhereClause` (320〜343 行) を置換:

```go
func (e *DynamicQueryExecutor) buildWhereClause(filters []models.FilterItem) (whereSQL string, whereValues []interface{}, nextIdx int, err error) {
	if len(filters) == 0 {
		return "", nil, 1, nil
	}

	whereClauses := make([]string, 0, len(filters))
	whereValues = make([]interface{}, 0, len(filters))
	idx := 1

	for _, filter := range filters {
		clause, value, filterErr := buildFilterClause(filter, idx)
		if filterErr != nil {
			return "", nil, 0, filterErr
		}
		if clause != "" {
			whereClauses = append(whereClauses, clause)
			whereValues = append(whereValues, value)
			idx++
		}
	}

	if len(whereClauses) == 0 {
		return "", nil, idx, nil
	}
	return "WHERE " + strings.Join(whereClauses, " AND "), whereValues, idx, nil
}
```

`buildFilterClause` (448〜472 行) を `idx int` 引数付きに置換:

```go
func buildFilterClause(filter models.FilterItem, idx int) (clause string, value interface{}, err error) {
	quotedCol, err := quoteIdentifier(filter.Field)
	if err != nil {
		return "", nil, fmt.Errorf("無効なフィルターフィールド %q: %w", filter.Field, err)
	}

	ph := fmt.Sprintf("$%d", idx)
	switch filter.Operator {
	case "eq":
		return quotedCol + " = " + ph, filter.Value, nil
	case "ne":
		return quotedCol + " != " + ph, filter.Value, nil
	case "gt":
		return quotedCol + " > " + ph, filter.Value, nil
	case "gte":
		return quotedCol + " >= " + ph, filter.Value, nil
	case "lt":
		return quotedCol + " < " + ph, filter.Value, nil
	case "lte":
		return quotedCol + " <= " + ph, filter.Value, nil
	case "like":
		return quotedCol + " LIKE " + ph, "%" + filter.Value + "%", nil
	default:
		return "", nil, nil
	}
}
```

`GetRecords`、`getRecordCount`、`executeRecordsQuery`、`GetRecordByID`、`GetAggregatedData` は `buildWhereClause` の戻り値が `(whereSQL, whereValues, nextIdx, err)` の 4 値に変わったことに合わせ、呼び出し側で `nextIdx` を受け取り、`LIMIT $N OFFSET $N+1`、`WHERE id = $1` 等を組み立てるよう修正する。

`getRecordCount` (346〜351 行) は変更なし（プレースホルダはすでに `whereValues` の数だけある）。

`executeRecordsQuery` (372〜416 行) を置換:

```go
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
	startIdx int,
) ([]models.RecordResponse, int64, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM %s %s ORDER BY %s LIMIT $%d OFFSET $%d",
		strings.Join(columns, ", "),
		quotedTable,
		whereSQL,
		orderBy,
		startIdx,
		startIdx+1,
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
```

`GetRecords` を呼び出し側調整:

```go
func (e *DynamicQueryExecutor) GetRecords(ctx context.Context, tableName string, fields []models.AppField, opts RecordQueryOptions) ([]models.RecordResponse, int64, error) {
	quotedTable, err := quoteIdentifier(tableName)
	if err != nil {
		return nil, 0, fmt.Errorf("無効なテーブル名: %w", err)
	}

	columns, err := e.buildColumnList(fields)
	if err != nil {
		return nil, 0, err
	}

	whereSQL, whereValues, nextIdx, err := e.buildWhereClause(opts.Filters)
	if err != nil {
		return nil, 0, err
	}

	total, err := e.getRecordCount(ctx, quotedTable, whereSQL, whereValues)
	if err != nil {
		return nil, 0, err
	}

	orderBy, err := e.buildOrderBy(opts.Sort, opts.Order)
	if err != nil {
		return nil, 0, err
	}

	return e.executeRecordsQuery(ctx, quotedTable, columns, whereSQL, whereValues, orderBy, opts, fields, total, nextIdx)
}
```

`GetRecordByID` の `WHERE id = ?` を `WHERE id = $1` に変更:

```go
query := fmt.Sprintf(
	"SELECT %s FROM %s WHERE id = $1",
	strings.Join(columns, ", "),
	quotedTable,
)
```

`GetAggregatedData` (564〜603 行) の WHERE は idx を考慮:

```go
whereSQL, whereValues, _, err := e.buildWhereClause(req.Filters)
```

`CountTodaysUpdates` (683〜696 行) の `CURDATE()` を `CURRENT_DATE` に置換:

```go
query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE DATE(updated_at) = CURRENT_DATE", quotedTable)
```

- [ ] **Step 6.7:** `go build ./...` で構文エラーが無いか確認

```bash
cd backend && go build ./...
```

期待: エラーなし。エラーが出る場合は修正を続ける。

- [ ] **Step 6.8:** commit する

```bash
git add backend/internal/repositories/dynamic_query.go
git commit -m "refactor(repo): port dynamic_query.go to PostgreSQL syntax"
```

---

## Task 7: `models/datasource.go` を PostgreSQL 一択に縮約

**Files:**
- Modify: `backend/internal/models/datasource.go`
- Modify: `backend/internal/models/datasource_test.go`

- [ ] **Step 7.1:** `datasource.go` の定数を縮約

旧 (12〜25 行):
```go
const (
	DBTypePostgreSQL DBType = "postgresql"
	DBTypeMySQL      DBType = "mysql"
	DBTypeOracle     DBType = "oracle"
	DBTypeSQLServer  DBType = "sqlserver"
)

var ValidDBTypes = []DBType{
	DBTypePostgreSQL,
	DBTypeMySQL,
	DBTypeOracle,
	DBTypeSQLServer,
}
```

新:
```go
const (
	DBTypePostgreSQL DBType = "postgresql"
)

var ValidDBTypes = []DBType{
	DBTypePostgreSQL,
}
```

- [ ] **Step 7.2:** validate タグを縮約

`CreateDataSourceRequest`, `TestConnectionRequest` の `validate:"required,oneof=postgresql mysql oracle sqlserver"` を `validate:"required,oneof=postgresql"` に。

- [ ] **Step 7.3:** `GetDefaultPort` 縮約

旧 (162〜176 行):
```go
func GetDefaultPort(dbType DBType) int {
	switch dbType {
	case DBTypePostgreSQL:
		return 5432
	case DBTypeMySQL:
		return 3306
	case DBTypeOracle:
		return 1521
	case DBTypeSQLServer:
		return 1433
	default:
		return 0
	}
}
```

新:
```go
func GetDefaultPort(dbType DBType) int {
	if dbType == DBTypePostgreSQL {
		return 5432
	}
	return 0
}
```

- [ ] **Step 7.4:** `datasource_test.go` を縮約

`DBTypeMySQL`, `DBTypeOracle`, `DBTypeSQLServer` への参照を含むテストケースを削除。`IsValidDBType` のテストは「`postgresql` は true、`mysql`/`oracle`/`sqlserver`/`unknown` は false」に書き換える。`GetDefaultPort` のテストも同様に縮約。

- [ ] **Step 7.5:** commit する

```bash
git add backend/internal/models/datasource.go backend/internal/models/datasource_test.go
git commit -m "refactor(models): restrict DBType to postgresql only"
```

---

## Task 8: `external_query.go` を PostgreSQL 専用に縮約

**Files:**
- Modify: `backend/internal/repositories/external_query.go`
- Modify: `backend/internal/repositories/external_query_test.go`
- Modify: `backend/internal/repositories/external_query_integration_test.go`

- [ ] **Step 8.1:** `external_query.go` の `import` から MSSQL/MySQL/Oracle ドライバを削除し、`lib/pq` のみ残す

```go
import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver

	"nocode-app/backend/internal/models"
)
```

`net/url` も MySQL/MSSQL のエンコードでしか使っていないので削除可能か確認のうえ削除（PG 用 `escapePostgresPassword` は残す）。

- [ ] **Step 8.2:** `buildDSN` (28〜66 行) を縮約

```go
func buildDSN(ds *models.DataSource, password string) (string, string, error) {
	if ds.DBType != models.DBTypePostgreSQL {
		return "", "", fmt.Errorf("サポートされていないデータベースタイプ: %s", ds.DBType)
	}
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		ds.Host, ds.Port, ds.Username, escapePostgresPassword(password), ds.DatabaseName)
	return "postgres", dsn, nil
}
```

- [ ] **Step 8.3:** `GetTables` の switch を縮約 (113〜158 行)

```go
func (e *ExternalQueryExecutor) GetTables(ctx context.Context, ds *models.DataSource, password string) ([]models.TableInfo, error) {
	db, err := openConnection(ctx, ds, password)
	if err != nil {
		return nil, err
	}
	defer func() { _ = db.Close() }()

	if ds.DBType != models.DBTypePostgreSQL {
		return nil, fmt.Errorf("サポートされていないデータベースタイプ: %s", ds.DBType)
	}

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
```

- [ ] **Step 8.4:** `GetColumns` (185〜293 行) も同様に PostgreSQL 専用に縮約

```go
func (e *ExternalQueryExecutor) GetColumns(ctx context.Context, ds *models.DataSource, password string, tableName string) ([]models.ColumnInfo, error) {
	db, err := openConnection(ctx, ds, password)
	if err != nil {
		return nil, err
	}
	defer func() { _ = db.Close() }()

	if ds.DBType != models.DBTypePostgreSQL {
		return nil, fmt.Errorf("サポートされていないデータベースタイプ: %s", ds.DBType)
	}

	const query = `SELECT
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
```

- [ ] **Step 8.5:** `quoteIdentifierForDB`, `getPlaceholder`, `buildLimitOffset` を PostgreSQL 専用化

```go
func quoteIdentifierForDB(_ models.DBType, name string) string {
	return fmt.Sprintf(`"%s"`, strings.ReplaceAll(name, `"`, `""`))
}

func getPlaceholder(_ models.DBType, index int) string {
	return fmt.Sprintf("$%d", index)
}

func buildLimitOffset(_ models.DBType, limit, offset int) string {
	return fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)
}
```

（引数の `models.DBType` は呼び出し側互換性のため残し `_` で受ける）

- [ ] **Step 8.6:** `external_query_test.go` の `dbType` テーブル駆動テストから MySQL/Oracle/MSSQL ケースを削除し、PostgreSQL ケースのみ残す

- [ ] **Step 8.7:** `external_query_integration_test.go` から以下を削除:
  - `TestExternalQueryExecutor_MySQL_Integration` (160〜300 行)
  - `TestExternalQueryExecutor_SQLServer_Integration` (301〜443 行)
  - `TestExternalQueryExecutor_Oracle_Integration` (444〜624 行)
  - 上 3 つに紐づく import / helper

`TestExternalQueryExecutor_PostgreSQL_Integration` のみ残す。

- [ ] **Step 8.8:** `go build ./...` で OK を確認

- [ ] **Step 8.9:** commit する

```bash
git add backend/internal/repositories/external_query.go backend/internal/repositories/external_query_test.go backend/internal/repositories/external_query_integration_test.go
git commit -m "refactor(repo): restrict external_query to PostgreSQL only"
```

---

## Task 9: testhelpers の整理

**Files:**
- Delete: `backend/internal/testhelpers/mssql_container.go`
- Delete: `backend/internal/testhelpers/mysql_container.go`
- Delete: `backend/internal/testhelpers/mysql_external_container.go`
- Delete: `backend/internal/testhelpers/oracle_container.go`
- Delete: `backend/internal/testhelpers/external_db_helper.go`
- Modify: `backend/internal/testhelpers/postgres_container.go`

- [ ] **Step 9.1:** `mysql_container.go` の `GetTestDB() / ResetDatabase()` を新ファイル `backend/internal/testhelpers/app_db.go` に PostgreSQL 版で移植

新規ファイル `backend/internal/testhelpers/app_db.go` を作成:

```go
package testhelpers

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

const (
	testDBName     = "nocode_test"
	testDBUser     = "nocode"
	testDBPassword = "nocodepassword"
)

var (
	sharedDB        *bun.DB
	sharedContainer *postgres.PostgresContainer
	sharedDBOnce    sync.Once
	sharedDBErr     error
)

// GetTestDB アプリ保管 DB のテスト用接続 (シングルトン) を返す。
// 初回呼び出し時に Postgres コンテナを起動し、init.sql で初期スキーマを流す。
func GetTestDB(ctx context.Context) (*bun.DB, error) {
	sharedDBOnce.Do(func() {
		sharedDB, sharedContainer, sharedDBErr = bootstrapTestDB(ctx)
	})
	return sharedDB, sharedDBErr
}

func bootstrapTestDB(ctx context.Context) (*bun.DB, *postgres.PostgresContainer, error) {
	initSQL, err := locateMigrationFile("init.sql")
	if err != nil {
		return nil, nil, err
	}

	c, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase(testDBName),
		postgres.WithUsername(testDBUser),
		postgres.WithPassword(testDBPassword),
		postgres.WithInitScripts(initSQL),
		postgres.BasicWaitStrategies(),
		postgres.WithSQLDriver("postgres"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("PostgreSQL コンテナの起動に失敗: %w", err)
	}

	dsn, err := c.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, nil, fmt.Errorf("接続文字列の取得に失敗: %w", err)
	}

	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("sql.Open 失敗: %w", err)
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(2)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(pingCtx); err != nil {
		return nil, nil, fmt.Errorf("Ping 失敗: %w", err)
	}

	// テストの待ち合わせ用に wait 戦略を追加（postgres モジュールが既に readiness を見ているが念のため）
	_ = wait.ForListeningPort

	return bun.NewDB(sqlDB, pgdialect.New()), c, nil
}

// locateMigrationFile testbinary 実行ディレクトリから上位を辿り、migrations/<name> を探す。
func locateMigrationFile(name string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := wd
	for i := 0; i < 6; i++ {
		candidate := filepath.Join(dir, "migrations", name)
		if _, statErr := os.Stat(candidate); statErr == nil {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("%s が見つかりません (起点: %s)", name, wd)
}

// ResetDatabase 既存のテーブルをすべて空にし、動的テーブルを破棄する。
func ResetDatabase(ctx context.Context) error {
	db, err := GetTestDB(ctx)
	if err != nil {
		return err
	}

	staticTables := []string{
		"dashboard_widgets",
		"chart_configs",
		"app_views",
		"app_fields",
		"apps",
		"data_sources",
		"users",
	}

	if _, err := db.ExecContext(ctx,
		`TRUNCATE `+joinIdent(staticTables)+` RESTART IDENTITY CASCADE`); err != nil {
		return fmt.Errorf("TRUNCATE 失敗: %w", err)
	}

	// 動的テーブル (app_<id>) をすべて drop
	rows, err := db.QueryContext(ctx, `
		SELECT tablename FROM pg_tables
		WHERE schemaname = 'public' AND tablename LIKE 'app_%'`)
	if err != nil {
		return fmt.Errorf("動的テーブル一覧取得に失敗: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var dynamicTables []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return err
		}
		// app_fields, app_views は静的なので除外
		if t == "app_fields" || t == "app_views" {
			continue
		}
		dynamicTables = append(dynamicTables, t)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, t := range dynamicTables {
		if _, err := db.ExecContext(ctx, `DROP TABLE IF EXISTS "`+t+`" CASCADE`); err != nil {
			return fmt.Errorf("DROP TABLE %s 失敗: %w", t, err)
		}
	}

	// admin ユーザーを再投入
	_, err = db.ExecContext(ctx, `
		INSERT INTO users (email, password_hash, name, role) VALUES
		('admin@example.com', '$2a$10$e8i3egbnenpqzZlow/3Q0.5L6uN8vNyktEYkgRdWwP13xSkCtR1re', 'Admin', 'admin')
		ON CONFLICT (email) DO NOTHING`)
	return err
}

func joinIdent(names []string) string {
	out := make([]string, len(names))
	for i, n := range names {
		out[i] = `"` + n + `"`
	}
	return joinWithComma(out)
}

func joinWithComma(s []string) string {
	if len(s) == 0 {
		return ""
	}
	res := s[0]
	for _, x := range s[1:] {
		res += ", " + x
	}
	return res
}
```

- [ ] **Step 9.2:** `postgres_container.go` を **外部 PostgreSQL テストコンテナ専用** にトリミング（既存実装を維持しつつ、`GetTestDB`/`ResetDatabase` は `app_db.go` に移したので衝突しないことを確認）

- [ ] **Step 9.3:** 不要ファイルを削除

```bash
rm backend/internal/testhelpers/mssql_container.go
rm backend/internal/testhelpers/mysql_container.go
rm backend/internal/testhelpers/mysql_external_container.go
rm backend/internal/testhelpers/oracle_container.go
rm backend/internal/testhelpers/external_db_helper.go
```

- [ ] **Step 9.4:** `go build ./...` で OK を確認

- [ ] **Step 9.5:** commit する

```bash
git add backend/internal/testhelpers/
git commit -m "refactor(testhelpers): consolidate to PostgreSQL-only fixtures"
```

---

## Task 10: 残りのテスト書き換え

**Files:**
- Modify: `backend/internal/handlers/datasource_test.go`
- Modify: `backend/internal/services/datasource_service_test.go`
- Modify: `backend/internal/services/chart_service_test.go`

- [ ] **Step 10.1:** 各ファイル内の `DBTypeMySQL`, `DBTypeOracle`, `DBTypeSQLServer` 参照行・サブテストを削除し、`DBTypePostgreSQL` のみ残す

- [ ] **Step 10.2:** `go test -short ./...` で全パッケージのコンパイルとユニットテスト通過を確認

```bash
cd backend && go test -short ./...
```

期待: すべて PASS

- [ ] **Step 10.3:** commit する

```bash
git add backend/internal/handlers backend/internal/services
git commit -m "refactor(test): trim non-postgres cases from handler/service tests"
```

---

## Task 11: `go.mod` / `go.sum` クリーンアップ

**Files:**
- Modify: `backend/go.mod`
- Modify: `backend/go.sum`

- [ ] **Step 11.1:** `go mod tidy` を実行して未使用依存を除去

```bash
cd backend && go mod tidy
```

- [ ] **Step 11.2:** `denisenkom/go-mssqldb`, `sijms/go-ora/v2`, `go-sql-driver/mysql`, `bun/dialect/mysqldialect` が消えていることを確認

```bash
grep -E "denisenkom|sijms|go-sql-driver|mysqldialect" go.mod
# 出力が空なら OK
```

- [ ] **Step 11.3:** `go test -short ./...` をもう一度回して通過を確認

- [ ] **Step 11.4:** commit する

```bash
git add backend/go.mod backend/go.sum
git commit -m "chore(deps): drop mssql/mysql/oracle drivers and mysqldialect"
```

---

## Task 12: フロントエンド縮約

**Files:**
- Modify: `frontend/src/types/datasource.ts`
- Modify: `frontend/src/types/datasource.test.ts`
- Modify: `frontend/src/api/datasources.test.ts`
- Modify: `frontend/src/components/datasources/DataSourceForm.tsx`
- Modify: `frontend/src/components/datasources/DataSourceList.tsx`

- [ ] **Step 12.1:** `frontend/src/types/datasource.ts` を縮約

```ts
export type DBType = "postgresql";

export const DB_TYPE_LABELS: Record<DBType, string> = {
  postgresql: "PostgreSQL",
};

export const DEFAULT_PORTS: Record<DBType, number> = {
  postgresql: 5432,
};
```

`DataSource`, `CreateDataSourceRequest`, `UpdateDataSourceRequest`, `TestConnectionRequest` の `db_type: DBType` はそのまま (型は縮約されている)。それ以外の interface はそのまま。

- [ ] **Step 12.2:** `frontend/src/types/datasource.test.ts` から MySQL/Oracle/SQLServer 期待値のテストを削除

- [ ] **Step 12.3:** `frontend/src/api/datasources.test.ts` も同様

- [ ] **Step 12.4:** `frontend/src/components/datasources/DataSourceForm.tsx` の DB 種別セレクトを「PostgreSQL 固定」に簡素化（Select 要素を残しつつ option を 1 つだけ表示するか、disabled で固定表示にする）

具体的にはセレクトボックス部分を以下のような表示に変更（既存マークアップに合わせて調整）:

```tsx
<FormControl>
  <FormLabel>DB タイプ</FormLabel>
  <Input value={DB_TYPE_LABELS.postgresql} isReadOnly />
</FormControl>
```

または既存 Select で `<option value="postgresql">PostgreSQL</option>` 1 件のみに。

- [ ] **Step 12.5:** `frontend/src/components/datasources/DataSourceList.tsx` の DB 種別バッジは PostgreSQL 一種なので、表示は `DB_TYPE_LABELS[ds.db_type]` のままで問題なし。MySQL/Oracle/SQLServer 専用のスタイル分岐があれば削除

- [ ] **Step 12.6:** フロント品質ゲート

```bash
cd frontend
pnpm run typecheck
pnpm run lint
pnpm run format:check
pnpm test
pnpm run build
```

すべて PASS を確認。

- [ ] **Step 12.7:** commit する

```bash
git add frontend/src
git commit -m "refactor(frontend): restrict DBType union to postgresql"
```

---

## Task 13: infra (compose / env / Dockerfile / README)

**Files:**
- Modify: `compose.yaml`
- Modify: `env.example`
- Modify: `.env.example`
- Modify: `README.md`

- [ ] **Step 13.1:** `compose.yaml` の MySQL サービスを Postgres に切り替える

```yaml
services:
  postgres:
    image: postgres:16-alpine
    container_name: nocode-postgres
    restart: unless-stopped
    environment:
      POSTGRES_DB: ${POSTGRES_DB}
      POSTGRES_USER: ${POSTGRES_USER}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./backend/migrations/init.sql:/docker-entrypoint-initdb.d/init.sql:ro
    healthcheck:
      test:
        ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER} -d ${POSTGRES_DB}"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - nocode-network

  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile
    container_name: nocode-backend
    restart: unless-stopped
    environment:
      DB_HOST: ${DB_HOST}
      DB_PORT: ${DB_PORT}
      DB_USER: ${DB_USER}
      DB_PASSWORD: ${DB_PASSWORD}
      DB_NAME: ${DB_NAME}
      JWT_SECRET: ${JWT_SECRET}
      JWT_EXPIRY_HOURS: ${JWT_EXPIRY_HOURS}
      SERVER_PORT: 8080
      ALLOWED_ORIGINS: ${ALLOWED_ORIGINS:-http://localhost:3000}
      ENCRYPTION_KEY: ${ENCRYPTION_KEY}
    ports:
      - "8080:8080"
    depends_on:
      postgres:
        condition: service_healthy
    networks:
      - nocode-network

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    container_name: nocode-frontend
    restart: unless-stopped
    ports:
      - "3000:80"
    depends_on:
      - backend
    networks:
      - nocode-network

volumes:
  postgres_data:

networks:
  nocode-network:
    driver: bridge
```

- [ ] **Step 13.2:** `env.example` を書き換え

```env
# PostgreSQL
POSTGRES_DB=nocode-app
POSTGRES_USER=nocode
POSTGRES_PASSWORD=nocodepassword

# Backend
DB_HOST=postgres
DB_PORT=5432
DB_USER=nocode
DB_PASSWORD=nocodepassword
DB_NAME=nocode-app
JWT_SECRET=your-super-secret-jwt-key-change-in-production
JWT_EXPIRY_HOURS=24
ALLOWED_ORIGINS=http://localhost:3000
ENCRYPTION_KEY=please-change-me-32-bytes-aes-key

# Frontend
VITE_API_URL=http://localhost:8080/api/v1
```

`.env.example` も同様の方針で書き換え（既存の `JWT_SECRET` 等のヒント文言は維持）。

- [ ] **Step 13.3:** `README.md` の MySQL 言及をすべて PostgreSQL に書き換える

具体的には以下の各箇所を書き換える:

| 行番号 | 旧 | 新 |
|---|---|---|
| L6 (バッジ) | `![MySQL](https://img.shields.io/badge/MySQL-8.0-4479A1?logo=mysql&logoColor=white)` | `![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?logo=postgresql&logoColor=white)` |
| L48 (フィールド型表) | `\| タイプ \| 説明 \| MySQLカラム型 \|` | `\| タイプ \| 説明 \| PostgreSQLカラム型 \|`、各行の MySQL 型を §5 (field.go) で決めた PostgreSQL 型に書き換え |
| L75 (依存表) | `\| MySQL \| go-sql-driver/mysql \| Pure Go実装 \|` | `\| PostgreSQL \| lib/pq \| Pure Go実装 \|` |
| L340 (要件表) | `\| MySQL \| 8.0+ \| データベース \|` | `\| PostgreSQL \| 16+ \| データベース \|` |
| L376 (アーキ図) | `subgraph Database["MySQL 8.0"]` | `subgraph Database["PostgreSQL 16"]` |
| L678 (DDL 表) | `\| db_type \| ENUM('postgresql','mysql','oracle','sqlserver') \| ...` | `\| db_type \| VARCHAR(20) CHECK (db_type IN ('postgresql')) \| ...` |
| L1269〜L1273 (env サンプル) | `# MySQL` ブロック全体 (`MYSQL_ROOT_PASSWORD`, `MYSQL_DATABASE`, `MYSQL_USER`, `MYSQL_PASSWORD`) | `# PostgreSQL` + `POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD` |
| L1276〜L1277 (env) | `DB_HOST=mysql` / `DB_PORT=3306` | `DB_HOST=postgres` / `DB_PORT=5432` |
| L1295 (テーブル) | `\| MySQL \| localhost:3306 \|` | `\| PostgreSQL \| localhost:5432 \|` |

加えて、外部データソース機能の説明箇所で「PostgreSQL / MySQL / Oracle / SQL Server」と列挙している部分を「PostgreSQL のみ」に書き換える。`grep -nE "MySQL|Oracle|SQLServer|SQL Server" README.md` の結果をもとに残存箇所を一掃する。

データ移行に関する記述があれば `docs/plans/2026-04-25-postgres-only-data-migration.md` へのリンクを冒頭セクションに追記する:

```md
> **DB 一択化（2026-04-25）:** 本プロジェクトは PostgreSQL 16 のみをサポートします。MySQL から移行する場合は [`docs/plans/2026-04-25-postgres-only-data-migration.md`](docs/plans/2026-04-25-postgres-only-data-migration.md) を参照してください。
```

- [ ] **Step 13.4:** commit する

```bash
git add compose.yaml env.example .env.example README.md
git commit -m "infra: switch local stack from MySQL to PostgreSQL"
```

---

## Task 14: 最終品質ゲート

> 詳細は `2026-04-25-postgres-only-data-migration.md` の §3「動作検証手順」と重複するが、PR を出す前にここで一度ローカル全部緑を確認する。

- [ ] **Step 14.1:** Backend ユニット & 統合テスト

```bash
cd backend
go vet ./...
gofmt -l . | (! grep .)   # フォーマット差分なし
go test -short ./...      # ユニット: すべて PASS
go test ./...             # 統合（testcontainers）: すべて PASS
```

- [ ] **Step 14.2:** Frontend テスト & ビルド

```bash
cd frontend
pnpm run typecheck
pnpm run lint
pnpm run format:check
pnpm test
pnpm run build
```

すべて PASS。

- [ ] **Step 14.3:** Docker Compose で起動確認

```bash
cd ..
cp env.example .env  # 必要なら
docker compose up -d --build
sleep 15
docker compose ps    # postgres / backend / frontend が healthy
curl -fsS http://localhost:8080/health
docker compose down
```

- [ ] **Step 14.4:** PR 作成

```bash
git push -u origin refactor/postgres-only-db
gh pr create --base main \
  --title "refactor: PostgreSQL-only support (drop MySQL/Oracle/SQL Server)" \
  --body-file <(cat docs/plans/2026-04-25-postgres-only-design.md)
```
