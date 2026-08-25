# 設計: 外部データソースのMySQL / Oracle / SQL Server対応復活

> 作成: 2026-08-25 / ブランチ: `feat/restore-multi-rdb-datasource`

## 1. 目的とスコープ

`2026-04-25-postgres-only-design.md`でPostgreSQLのみに縮約した外部データソース機能の対応RDBを、縮約前と同じ4種（PostgreSQL / MySQL / Oracle / SQL Server）に戻す。縮約後に入った機能と修正（SQL injection対策`d27c4a5`、JSONB/timestamp修正、CI共通化）はすべて維持する。

### 含むもの

- backend: DSN構築、テーブル・カラム一覧取得、レコード取得・集計・件数の4方言分岐、識別子クォート・プレースホルダ・LIMIT/OFFSETの方言差
- backend: `models.DBType`の4種復活、validateタグ、`GetDefaultPort`
- backend: `data_sources.db_type`のCHECK制約を4種に拡張
- backend: testcontainersによるMySQL / SQL Server / Oracleの統合テストとtesthelpers
- frontend: `DBType`ユニオン、ラベル、デフォルトポート、DB種別セレクト、一覧バッジ
- CI: 4RDBの統合テストjobを追加
- README: 対応RDB表と廃止告知の更新

### 含まないもの

- アプリ自身の保管DB（PostgreSQL固定）の多方言化。migrations本体、`database.go`、`config.go`、`dynamic_query.go`、`compose.yaml`、envは変更しない
- 機能追加・UI改修

## 2. 復元元と統合方針

復元元は`c0bd653^1`（PR #13マージ直前）の各ファイルです。そのまま戻さず、以下のとおり現行コードと統合します。

| 領域 | 縮約前（`c0bd653^1`） | 現行（HEAD） | 統合後 |
|---|---|---|---|
| `quoteIdentifierForDB` | `(dbType, name) string`。方言ごとにクォート（PG/Oracleは`"`二重化、MySQLはバッククォート二重化、MSSQLは`]]`）。Oracleは大文字化 | `(name) (string, error)`。空・長さ上限・制御文字を拒否し、CodeQL用サニタイザバリアのコメント付き | `(dbType, name) (string, error)`。現行の検証を先に行い、通過した識別子に縮約前の方言別クォートを適用する。コメントは現行を維持 |
| `getPlaceholder` / `buildLimitOffset` | 方言分岐 | `_ models.DBType`でPG固定 | 縮約前の方言分岐を復元する。シグネチャは現行と同じなので呼び出し側は無変更 |
| `buildDSN` / `openConnection` | 4方言 | PGのみ | 縮約前を復元する。SQL Serverのドライバは`github.com/microsoft/go-mssqldb`（`denisenkom/go-mssqldb`はアーカイブ済み。driver名`sqlserver`は同じ） |
| `GetTables` / `GetColumns` / `GetRecords` / `GetRecordByID` / `GetAggregatedData` / `CountRecords` | 方言分岐あり | PGのみ。`quoteIdentifierForDB`のエラーを扱う | 縮約前の分岐を、現行のエラー処理の形で復元する |
| 単体テスト`external_query_test.go` | 4方言×各関数 | PG+セキュリティ拒否ケース（制御文字・DEL・空・長さ） | 4方言×各関数に加え、セキュリティ拒否ケースを4方言すべてで検証する |
| 統合テスト | 4RDB、`//go:build integration` | PGのみ、同タグ+`testing.Short()`ガード | 4RDB、同タグ+同ガード |
| testhelpers | `mysql_external_container.go` / `mssql_container.go` / `oracle_container.go` / `external_db_helper.go` | `app_db.go`（`openTestDB`を内包）/ `postgres_container.go` | 3ファイルを復元する。`external_db_helper.go`は`app_db.go`の`openTestDB`と重複するため復元しない |
| `models/datasource.go` | 4種 | 1種 | 4種に戻す（定数・`ValidDBTypes`・validate `oneof`・`GetDefaultPort`） |
| `migrations/init.sql` | — | `CHECK (db_type IN ('postgresql'))` | `CHECK (db_type IN ('postgresql','mysql','oracle','sqlserver'))`。アプリにマイグレーションランナーは無く、init.sqlは初期投入専用のため追加マイグレーションは作らない |
| frontend | 4種 | 1種 | 縮約前の型・定数・セレクトを復元する。コンポーネントの現行スタイルは維持 |

## 3. 依存関係

| 追加 | バージョン | 用途 |
|---|---|---|
| `github.com/go-sql-driver/mysql` | v1.10.0 | MySQL driver |
| `github.com/microsoft/go-mssqldb` | v1.11.0 | SQL Server driver（driver名`sqlserver`） |
| `github.com/sijms/go-ora/v2` | v2.9.0 | Oracle driver（pure Go、driver名`oracle`） |
| `github.com/testcontainers/testcontainers-go/modules/mysql` | v0.42.0 | 既存`testcontainers-go v0.42.0`に合わせる |
| `github.com/testcontainers/testcontainers-go/modules/mssql` | v0.42.0 | 同上 |

コンテナイメージのうち`mysql:8.0`と`mcr.microsoft.com/mssql/server:2022-latest`は縮約前と同じです。Oracleは`gvenzl/oracle-free:23-slim`を使います（service名FREEPDB1、arm64/amd64両対応）。

## 4. テスト戦略

- 単体: `go test -short -race ./...`（既存CI job、変更なし）
- 統合: `go test -race -count=1 -tags=integration ./internal/repositories/...`をCIの新job `backend-integration`で実行する（`okamyuji/reusable-workflows/.github/workflows/go-ci.yml@v1`の2回目の呼び出し）。ubuntu-latestにはDockerがありtestcontainersが動く
- ローカル（colima）: `DOCKER_HOST=unix://$HOME/.colima/default/docker.sock TESTCONTAINERS_RYUK_DISABLED=true`を付けて実行する。SQL Serverイメージはamd64のためarm64では遅い。ローカルで不安定な場合の最終判定はCIの統合job
- frontend: `pnpm test` / `typecheck` / `lint` / `format:check` / `build`（既存CI job、変更なし）

## 5. リスク

| リスク | 対処 |
|---|---|
| 縮約前コードの復元でCodeQLのSQL injection指摘が再発する | `quoteIdentifierForDB`の検証とコメントを現行のまま残し、方言別クォートはその後段に置く。CodeQLはmainへのpushで自動実行されるためPRマージ後に確認する |
| Oracleコンテナの起動時間でCIが長くなる | 統合jobは既存jobと並列。timeoutは統合テスト側で5分/RDB |
| `denisenkom/go-mssqldb`からの置き換えで挙動差が出る | microsoft版は同一コードベースの後継。統合テストで検証する |
