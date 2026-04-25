# 設計計画: PostgreSQL 一択化

> 作成: 2026-04-25 / ブランチ: `refactor/postgres-only-db` / 対象 commit: `d11a532` 起点

## 1. 目的とスコープ

### 目的
`nocode-app` の対応 RDB を **PostgreSQL のみ** に統一する。

### 含むもの
1. アプリ自身の保管 DB（`users`, `apps`, `app_fields`, `app_views`, `chart_configs`, `dashboard_widgets`, `data_sources` および動的に作られるアプリテーブル）を MySQL 8.0 → PostgreSQL 16 に切替
2. 外部データソース機能の対応 DB を `postgresql / mysql / oracle / sqlserver` の 4 種から `postgresql` のみに縮約
3. 上記に伴う MySQL / Oracle / SQL Server 関連のコード・依存・テスト・ドキュメントの完全削除

### 含まないもの
- 既存の **動作している MySQL 本番データの自動移行ツール**（→ 別途データ移行計画書 `2026-04-25-postgres-only-data-migration.md` で手順化）
- 機能追加・UI 改修

---

## 2. 影響範囲（ファイル一覧）

### 2.1 削除するファイル

| パス | 行数 | 理由 |
|---|---:|---|
| `backend/internal/testhelpers/mssql_container.go` | 183 | MSSQL 外部接続テスト用 |
| `backend/internal/testhelpers/oracle_container.go` | 157 | Oracle 外部接続テスト用 |
| `backend/internal/testhelpers/mysql_container.go` | 304 | アプリ保管 MySQL 統合テスト用（新 `postgres_container.go` に置換） |
| `backend/internal/testhelpers/mysql_external_container.go` | 145 | MySQL 外部接続テスト用 |
| `backend/internal/testhelpers/external_db_helper.go` | 26 | 4 種ドライバの blank import を集約しているだけ |

### 2.2 書き換えるファイル

| パス | 内容 |
|---|---|
| `backend/migrations/init.sql` | PostgreSQL 構文に全面書き換え |
| `backend/migrations/002_datasources.sql` | 同上 |
| `backend/migrations/003_dashboard_widgets.sql` | 同上 |
| `backend/internal/database/database.go` | `mysqldialect` → `pgdialect`、`go-sql-driver/mysql` → `lib/pq` |
| `backend/internal/config/config.go` | `DSN()` を PostgreSQL 形式へ。`DB_PORT` のデフォルトを `3306` → `5432` |
| `backend/internal/repositories/external_query.go` | PostgreSQL 専用に縮小（switch/case 削除、ファイル本体を半分以下に） |
| `backend/internal/repositories/external_query_test.go` | PostgreSQL 関連テストのみ残す |
| `backend/internal/repositories/external_query_integration_test.go` | `TestExternalQueryExecutor_PostgreSQL_Integration` 以外を削除 |
| `backend/internal/repositories/dynamic_query.go` | 識別子クォートをバックティック (`` ` ``) → ダブルクォート (`"`)、`BIGINT UNSIGNED AUTO_INCREMENT` → `BIGSERIAL`、`CURDATE()` → `CURRENT_DATE`、`ON UPDATE CURRENT_TIMESTAMP` 廃止（トリガ `set_updated_at()` に移管）、`LastInsertId()` → `INSERT ... RETURNING id`。プレースホルダは bun の `?` のまま (ADR-5 参照) |
| `backend/internal/models/field.go` | `GetMySQLColumnType` → `GetPostgresColumnType`、戻り値も PG 型へ |
| `backend/internal/models/field_test.go` | 上記の改名と期待値変更 |
| `backend/internal/models/datasource.go` | `DBType` を `DBTypePostgreSQL` のみに縮約、`ValidDBTypes`、`IsValidDBType`、validate タグ、`GetDefaultPort` を縮小 |
| `backend/internal/models/datasource_test.go` | 同上 |
| `backend/internal/handlers/datasource_test.go` | MySQL/Oracle/MSSQL のテストケース削除 |
| `backend/internal/services/datasource_service_test.go` | 同上 |
| `backend/internal/services/chart_service_test.go` | DBType 関連テスト削除 |
| `backend/internal/repositories/external_query.go` 内の testhelper 連携箇所 | （該当なし、確認のみ） |
| `backend/internal/testhelpers/postgres_container.go` | アプリ保管 DB 用 `GetTestDB() / ResetDatabase()` を提供する形に拡張、外部 DB テスト用フィクスチャもここに統合 |
| `backend/go.mod`, `backend/go.sum` | `denisenkom/go-mssqldb`, `sijms/go-ora/v2`, `go-sql-driver/mysql`, `bun/dialect/mysqldialect` を削除（`bun/dialect/pgdialect`, `lib/pq` のみ残す） |
| `compose.yaml` | `mysql:8.0` サービスを `postgres:16-alpine` へ。volume 名・healthcheck・env を変更 |
| `env.example`, `.env.example` | `MYSQL_*` → `POSTGRES_*`、`DB_HOST=postgres`, `DB_PORT=5432` |
| `frontend/src/types/datasource.ts` | `DBType` ユニオンを `"postgresql"` のみに、`DB_TYPE_LABELS` / `DEFAULT_PORTS` を縮約 |
| `frontend/src/types/datasource.test.ts` | 縮約に合わせる |
| `frontend/src/api/datasources.test.ts` | 同上 |
| `frontend/src/components/datasources/DataSourceForm.tsx` | DB 種別セレクトボックスを PostgreSQL 固定の表示（または 1 件のみのセレクト）に |
| `frontend/src/components/datasources/DataSourceList.tsx` | DB 種別ラベルバッジを PostgreSQL 固定表示 |
| `README.md` | DB 章を PG 中心に |

### 2.3 新規作成するファイル

| パス | 内容 |
|---|---|
| `docs/plans/2026-04-25-postgres-only-design.md` | 本書 |
| `docs/plans/2026-04-25-postgres-only-implementation.md` | 実装計画（タスク分解） |
| `docs/plans/2026-04-25-postgres-only-data-migration.md` | データ移行計画 |

---

## 3. アーキテクチャ決定 (ADR)

### ADR-1: ドライバ選定

**選定: `github.com/lib/pq`（既存依存）**

| 候補 | 採用 | 理由 |
|---|---|---|
| `lib/pq` | ✅ | 既に外部接続用に依存しており、セキュリティアラートの直近履歴も少ない。bun の `pgdialect` と組み合わせ可能 |
| `jackc/pgx/v5` (database/sql wrap) | ❌ | より高機能だが新規依存。今回スコープでは過剰 |
| `bun/driver/pgdriver` | ❌ | bun 専用、`database/sql.DB` 経由の既存テストヘルパとの相性が悪い |

### ADR-2: 主キー型

**選定: `BIGSERIAL`**

| 候補 | 採用 | 理由 |
|---|---|---|
| `BIGSERIAL` | ✅ | PostgreSQL 慣用。bun の `id,pk,autoincrement` タグと素直に一致 |
| `BIGINT GENERATED ALWAYS AS IDENTITY` | ❌ | より新しいが、bun の autoincrement 経路が `currval` 系を期待する場合に取り回しが悪い |

### ADR-3: ENUM の扱い

**選定: `CHECK (col IN (...))` + `VARCHAR`**

| 候補 | 採用 | 理由 |
|---|---|---|
| `CREATE TYPE ... AS ENUM` | ❌ | 値追加に `ALTER TYPE` が要り、bun の string 型バインドと相性悪い |
| `CHECK (col IN (...))` + `VARCHAR(N)` | ✅ | 既存 Go コードはすべて `string` 型でやり取りしており追加改修ゼロ |

### ADR-4: `ON UPDATE CURRENT_TIMESTAMP` の代替

**選定: PL/pgSQL トリガ `set_updated_at()` を 1 関数で共有**

```sql
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
```

各テーブルに `BEFORE UPDATE` トリガを 1 個ずつ張る。

### ADR-5: 識別子クォートの正規化とプレースホルダ方針

`dynamic_query.go` のすべてのバックティックを **ダブルクォート (`"`)** に置換する。

プレースホルダは **bun の `?` をそのまま使う**。`bun.DB.QueryRowContext` / `ExecContext` / `QueryContext` は内部で `db.format(query, args...)` を呼び、引数を SQL リテラルとして **インライン化したクエリを生成**するため（pgdialect に応じた適切なエスケープが行われる）、`$N` 形式に手動で書き換える必要はない。`pgPlaceholders(n int)` のようなヘルパも追加しない。

> 当初案では `?` を `$N` に手動置換する予定だったが、bun の format 動作を確認した結果、`?` のままで pgdialect が正しく扱うため不要と判明した。

### ADR-6: `LastInsertId()` の代替

PostgreSQL は `LastInsertId()` を返さないため、`INSERT ... RETURNING id` で取得し `db.QueryRowContext(...).Scan(&id)` で受ける。

### ADR-7: フィールド型マッピング

| アプリ FieldType | MySQL 型（旧） | PostgreSQL 型（新） |
|---|---|---|
| text / select / radio | `VARCHAR(255)` | `VARCHAR(255)` |
| textarea | `TEXT` | `TEXT` |
| number | `DECIMAL(18,4)` | `NUMERIC(18,4)` |
| date | `DATE` | `DATE` |
| datetime | `DATETIME` | `TIMESTAMP` |
| multiselect / attachment | `JSON` | `JSONB` |
| checkbox | `BOOLEAN` | `BOOLEAN` |
| link | `VARCHAR(500)` | `VARCHAR(500)` |

---

## 4. テスト戦略

### 4.1 単体テスト
- `models/field_test.go` の `TestAppField_GetPostgresColumnType` で全 FieldType の戻り値を表検証
- `models/datasource_test.go` で `DBTypePostgreSQL` のみ受理、それ以外を拒否
- `repositories/dynamic_query` の SQL 文字列ビルダ系（プレースホルダ番号、`RETURNING id`、`CURRENT_DATE`）はモックなしの文字列比較で検証

### 4.2 統合テスト（testcontainers）
- `chart_repository_test.go`, `app_repository_test.go` 等の既存統合テストは `testhelpers.GetTestDB / ResetDatabase` 経由で PG コンテナを使うように切替（呼び出しは無変更、内部実装だけ差し替え）
- `external_query_integration_test.go` は `TestExternalQueryExecutor_PostgreSQL_Integration` 一本に縮約

### 4.3 品質ゲート（最終確認）
1. **Backend**: `go vet ./...`、`go test -short ./...`（ユニットのみ、CI 必須）、`go test ./...`（統合含む。ローカル & CI で testcontainers が動く想定）、`gofmt -l .`
2. **Frontend**: `pnpm run typecheck`、`pnpm run lint`、`pnpm run format:check`、`pnpm test`、`pnpm run build`
3. **Docker**: `docker compose up -d` で起動、ヘルスチェック緑、フロントから API 叩いて 200 系

---

## 5. リスクとロールバック

### リスク
- bun の autoincrement タグが PG では `id,pk,autoincrement` から `id,pk,default:nextval('...')` に変更が必要なケースがあるが、bun は dialect 切替で自動的に `BIGSERIAL` を期待する挙動なので原則そのまま動く。差分が出たらタグ書き換えで対処。
- `dynamic_query.go` のプレースホルダ変換漏れ → 統合テストでクラッシュするので発見可能
- `JSON` 型カラムの既存値が JSONB に入る際、bun のエンコードが `string` か `[]byte` かで挙動差。`FieldOptions` 型の `Value()/Scan()` に変更は不要（`string` を返しているため PG `JSONB` 列にもそのまま入る）

### ロールバック
- 本変更は 1 PR で完結する。リバート時は単純に PR を revert すればよい
- DB 物理移行はデータ移行計画書のスナップショット運用により、`docker compose down -v` ＋ MySQL ボリューム再構築で原状回復可能

---

## 6. 自己レビュー結果

- [x] 仕様書（ユーザーリクエスト）の各要件にタスクが対応している
  - PG 一択化 → §2 の全項目で覆われる
  - MySQL/Oracle/MSSQL の関連コード・テスト全削除 → §2.1, §2.2 で網羅
- [x] プレースホルダ語（"TBD", "適切に処理する" 等）が無い
- [x] 型・識別子の整合: `GetPostgresColumnType`、`DBTypePostgreSQL`、`pgdialect` を全文書で一貫使用
