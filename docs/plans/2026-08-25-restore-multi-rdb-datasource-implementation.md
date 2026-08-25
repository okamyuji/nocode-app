# 実装計画: 外部データソースのMySQL / Oracle / SQL Server対応復活

> 設計: `2026-08-25-restore-multi-rdb-datasource-design.md` / ブランチ: `feat/restore-multi-rdb-datasource`

## 全体制約

- 復元元は`git show c0bd653^1:<path>`で取得する（PR #13マージ直前）。復元は「現行コードに縮約前の分岐を戻す」方向で行い、現行にある検証・エラー処理・コメントを消さない。
- アプリ保管DB（PostgreSQL）に関するファイルは変更しない: `backend/migrations/*.sql`の`db_type` CHECK以外、`backend/internal/database/`、`backend/internal/config/`、`backend/internal/repositories/dynamic_query*.go`、`compose.yaml`、`env.example`、`.env.example`。
- SQL Serverドライバは`github.com/microsoft/go-mssqldb v1.11.0`（import pathのみ縮約前と異なる。driver名`sqlserver`は同じ）。
- testcontainersモジュールは既存の`v0.42.0`に揃える。
- 統合テストは`//go:build integration`タグと`testing.Short()`ガードの両方を持つ。
- コミットはconventional commits。`Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>`を付け、`Claude-Session`行は付けない。
- ローカルでtestcontainersを使うときは`DOCKER_HOST=unix://$HOME/.colima/default/docker.sock TESTCONTAINERS_RYUK_DISABLED=true`を付ける。
- `cd`は使わず`git -C` / `go -C` / サブシェル`(cd dir && ...)`を使う。
- 各タスクの最後に`gofmt -l .`が空、`go vet ./...`と`go test -short -race -count=1 ./...`（backend）が通ること。frontendは`pnpm run typecheck && pnpm run lint && pnpm run format:check && pnpm test && pnpm run build`。

## タスク一覧

| # | タスク | 主なファイル |
|---|---|---|
| 1 | 依存追加・モデル復元・CHECK制約拡張 | `backend/go.mod`, `backend/go.sum`, `backend/internal/models/datasource.go`, `backend/internal/models/datasource_test.go`, `backend/internal/handlers/datasource_test.go`, `backend/internal/services/datasource_service_test.go`, `backend/internal/services/chart_service_test.go`, `backend/migrations/init.sql` |
| 2 | `external_query.go`の4方言復元と単体テスト | `backend/internal/repositories/external_query.go`, `backend/internal/repositories/external_query_test.go` |
| 3 | testhelpers復元と統合テスト（MySQL / SQL Server / Oracle） | `backend/internal/testhelpers/mysql_external_container.go`, `backend/internal/testhelpers/mssql_container.go`, `backend/internal/testhelpers/oracle_container.go`, `backend/internal/repositories/external_query_integration_test.go` |
| 4 | frontendの4種復元 | `frontend/src/types/datasource.ts`, `frontend/src/types/datasource.test.ts`, `frontend/src/api/datasources.test.ts`, `frontend/src/components/datasources/DataSourceForm.tsx`, `frontend/src/components/datasources/DataSourceList.tsx` |
| 5 | CI統合job追加とREADME更新 | `.github/workflows/ci.yml`, `README.md` |

---

### Task 1: 依存追加・モデル復元・CHECK制約拡張

Files:
- Modify: `backend/go.mod`, `backend/go.sum`
- Modify: `backend/internal/models/datasource.go`
- Modify: `backend/internal/models/datasource_test.go`
- Modify: `backend/internal/handlers/datasource_test.go`
- Modify: `backend/internal/services/datasource_service_test.go`
- Modify: `backend/internal/services/chart_service_test.go`
- Modify: `backend/migrations/init.sql`（`db_type`のCHECKのみ）

Produces: `models.DBTypeMySQL` / `models.DBTypeOracle` / `models.DBTypeSQLServer`、`models.ValidDBTypes`（4要素）、`models.GetDefaultPort(DBType) int`（5432 / 3306 / 1521 / 1433）、`validate:"required,oneof=postgresql mysql oracle sqlserver"`。

- [ ] Step 1: 復元元との差分を確認する

```bash
git -C /path/to/nocode-app diff c0bd653^1 HEAD -- backend/internal/models/datasource.go backend/internal/models/datasource_test.go backend/internal/handlers/datasource_test.go backend/internal/services/datasource_service_test.go backend/internal/services/chart_service_test.go
```

- [ ] Step 2: 失敗するテストを先に戻す。`datasource_test.go`に縮約前のMySQL / Oracle / SQL Serverのケース（`IsValidDBType`、`GetDefaultPort`、validateの`oneof`）を復元し、`handlers` / `services` / `chart_service`のテストにも縮約前のDBTypeケースを復元する。`go test -short ./internal/models/ ./internal/handlers/ ./internal/services/`が失敗することを確認する。

- [ ] Step 3: `models/datasource.go`を復元する。定数4種、`ValidDBTypes`、validateタグ、`GetDefaultPort`。現行の他フィールド・コメントは維持する。

- [ ] Step 4: `backend/migrations/init.sql`の`CHECK (db_type IN ('postgresql'))`を`CHECK (db_type IN ('postgresql', 'mysql', 'oracle', 'sqlserver'))`に変更する。他行は変更しない。

- [ ] Step 5: 依存を追加する

```bash
(cd backend && go get github.com/go-sql-driver/mysql@v1.10.0 github.com/microsoft/go-mssqldb@v1.11.0 github.com/sijms/go-ora/v2@v2.9.0 github.com/testcontainers/testcontainers-go/modules/mysql@v0.42.0 github.com/testcontainers/testcontainers-go/modules/mssql@v0.42.0 && go mod tidy)
```

`go mod tidy`は未使用importを落とすため、Task 2 / 3でimportされるまでは`require`から消える。消えた場合はTask 2 / 3で再度`go get`する。この時点では`go.mod`に残らなくてよい。

- [ ] Step 6: 検証

```bash
(cd backend && gofmt -l . && go vet ./... && DOCKER_HOST=unix://$HOME/.colima/default/docker.sock TESTCONTAINERS_RYUK_DISABLED=true go test -short -race -count=1 ./...)
```

- [ ] Step 7: コミット `feat(models): restore mysql/oracle/sqlserver DBType and widen db_type CHECK`

---

### Task 2: `external_query.go`の4方言復元と単体テスト

Files:
- Modify: `backend/internal/repositories/external_query.go`
- Modify: `backend/internal/repositories/external_query_test.go`
- Modify: `backend/go.mod`, `backend/go.sum`（driverのblank import追加で`require`に戻る）

Consumes: Task 1の`models.DBType*`。
Produces: `quoteIdentifierForDB(dbType models.DBType, name string) (string, error)`、`getPlaceholder(dbType models.DBType, index int) string`、`buildLimitOffset(dbType models.DBType, limit, offset int) string`、`buildDSN(ds *models.DataSource, password string) (driverName, dsn string, err error)`。

- [ ] Step 1: 復元元を取り出す

```bash
git -C /path/to/nocode-app show c0bd653^1:backend/internal/repositories/external_query.go > /tmp/old_external_query.go
git -C /path/to/nocode-app show c0bd653^1:backend/internal/repositories/external_query_test.go > /tmp/old_external_query_test.go
git -C /path/to/nocode-app diff c0bd653 HEAD -- backend/internal/repositories/external_query.go   # d27c4a5 のセキュリティ変更
```

- [ ] Step 2: 失敗するテストを先に書く。`external_query_test.go`を次の構成にする。
  - `TestQuoteIdentifierForDB`: 縮約前の4方言ケース（PG `"x"`、MySQL `` `x` ``、Oracle `"X"`（大文字化）、MSSQL `[x]`、各エスケープ）。呼び出しは`quoteIdentifierForDB(dbType, name)`で`(string, error)`を受ける。
  - `TestQuoteIdentifierForDB_RejectsInvalid`: 現行のセキュリティ拒否ケース（空文字、`maxExternalIdentifierLength`超、NUL、DEL 0x7f、改行）を4方言すべてでループして`err != nil`を確認する。
  - `TestQuoteIdentifierForDBWithJapaneseEdgeCases`: 縮約前の日本語・絵文字ケースを4方言で。
  - `TestGetPlaceholder`: PG `$1`、MySQL `?`、Oracle `:1`、MSSQL `@p1`。
  - `TestBuildLimitOffset`: PG/MySQL ` LIMIT 10 OFFSET 20`、Oracle/MSSQL ` OFFSET 20 ROWS FETCH NEXT 10 ROWS ONLY`。
  - `TestBuildDSN`: 縮約前の4方言ケース（driver名`postgres` / `mysql` / `oracle` / `sqlserver`、URLエンコード、未対応typeでerror）。
  - `TestEscapePostgresPassword`、`TestConvertScannedValue`: 現行のまま。
  `go test -short ./internal/repositories/ -run 'TestQuote|TestGetPlaceholder|TestBuildLimitOffset|TestBuildDSN'`がコンパイルエラーまたは失敗になることを確認する。

- [ ] Step 3: `external_query.go`を実装する。
  - import: `_ "github.com/go-sql-driver/mysql"`、`_ "github.com/microsoft/go-mssqldb"`、`_ "github.com/sijms/go-ora/v2"`、`net/url`を追加。既存の`_ "github.com/lib/pq"`は維持。
  - `buildDSN`: 縮約前のswitchを復元。SQL Serverのcaseは縮約前と同じDSN形式（`sqlserver://user:pass@host:port?database=name`）。
  - `quoteIdentifierForDB(dbType, name)`: 現行の3つの検証（空・長さ・`externalIdentifierRegex`）とコメントをそのまま先頭に置き、その後に縮約前のswitchでクォートする。`default`は`fmt.Errorf("サポートされていないデータベースタイプ: %s", dbType)`を返す。
  - `getPlaceholder` / `buildLimitOffset`: 縮約前のswitchを復元。
  - `GetTables` / `GetColumns`: 縮約前のswitchでクエリを分岐する。PGのcaseは現行のクエリをそのまま使う。Oracleの`GetColumns`はLONG型回避（`b2e1bcd`）を含む縮約前の実装を使う。
  - `GetRecords` / `GetRecordByID` / `GetAggregatedData` / `CountRecords`: 現行の実装をベースに、`quoteIdentifierForDB(ds.DBType, ...)`、`getPlaceholder(ds.DBType, ...)`、`buildLimitOffset(ds.DBType, ...)`を渡す。縮約前にあった方言固有の処理（SQL ServerのOFFSETに必須の`ORDER BY`、Oracleの大文字識別子）を復元する。
  - 各関数のdocコメントにある`PostgreSQL のみ`という文言を外す。

- [ ] Step 4: 検証

```bash
(cd backend && go mod tidy && gofmt -l . && go vet ./... && DOCKER_HOST=unix://$HOME/.colima/default/docker.sock TESTCONTAINERS_RYUK_DISABLED=true go test -short -race -count=1 ./...)
(cd backend && DOCKER_HOST=unix://$HOME/.colima/default/docker.sock TESTCONTAINERS_RYUK_DISABLED=true go test -count=1 -tags=integration -run TestExternalQueryExecutor_PostgreSQL_Integration ./internal/repositories/)
```

PGの統合テストが復元後も通ることを確認する（回帰確認）。

- [ ] Step 5: コミット `feat(repo): restore multi-dialect external query executor (mysql/oracle/sqlserver)`

---

### Task 3: testhelpers復元と統合テスト

Files:
- Create: `backend/internal/testhelpers/mysql_external_container.go`
- Create: `backend/internal/testhelpers/mssql_container.go`
- Create: `backend/internal/testhelpers/oracle_container.go`
- Modify: `backend/internal/repositories/external_query_integration_test.go`
- Modify: `backend/go.mod`, `backend/go.sum`

Consumes: Task 2の`ExternalQueryExecutor`。
Produces: `testhelpers.SetupMySQLExternalContainer(ctx) (*MySQLExternalTestContainer, error)`、`testhelpers.SetupMSSQLContainer(ctx) (*MSSQLTestContainer, error)`、`testhelpers.SetupOracleContainer(ctx) (*OracleTestContainer, error)`。各構造体は`Host` / `Port` / `Database` / `Username` / `Password`と`Terminate(ctx)` / `CreateTestTable(ctx)` / `CreateTestView(ctx)`を持つ（縮約前と同じ）。

- [ ] Step 1: 復元元を取り出す

```bash
for f in mysql_external_container mssql_container oracle_container; do git -C /path/to/nocode-app show c0bd653^1:backend/internal/testhelpers/$f.go > backend/internal/testhelpers/$f.go; done
git -C /path/to/nocode-app show c0bd653^1:backend/internal/repositories/external_query_integration_test.go > /tmp/old_integration_test.go
```

- [ ] Step 2: 復元した3ファイルを現行に合わせる。
  - `openTestDB`は`app_db.go`にあるものを使う（`external_db_helper.go`は復元しない）。
  - `mssql_container.go`のimportを`github.com/microsoft/go-mssqldb`に変える。
  - 現行`postgres_container.go`の構造（フィールド名、`Terminate`のシグネチャ）と揃える。

- [ ] Step 3: `external_query_integration_test.go`に縮約前の`TestExternalQueryExecutor_MySQL_Integration` / `_SQLServer_Integration` / `_Oracle_Integration`を復元する。各テストの先頭に現行PGテストと同じ`testing.Short()`ガードと`5*time.Minute`のcontextを置く。サブテスト構成（TestConnection / GetTables / GetColumns / GetRecords / GetRecordByID / CountRecords / GetAggregatedData / View）は現行PGテストに揃える。

- [ ] Step 4: 検証（順に実行。SQL Serverはarm64ではamd64エミュレーションのため数分かかる）

```bash
export DOCKER_HOST=unix://$HOME/.colima/default/docker.sock TESTCONTAINERS_RYUK_DISABLED=true
(cd backend && go mod tidy && gofmt -l . && go vet -tags=integration ./... )
(cd backend && go test -count=1 -tags=integration -run TestExternalQueryExecutor_MySQL_Integration -v ./internal/repositories/ 2>&1 | tail -20)
(cd backend && go test -count=1 -tags=integration -run TestExternalQueryExecutor_SQLServer_Integration -v ./internal/repositories/ 2>&1 | tail -20)
(cd backend && go test -count=1 -tags=integration -run TestExternalQueryExecutor_Oracle_Integration -v ./internal/repositories/ 2>&1 | tail -20)
(cd backend && go test -short -race -count=1 ./...)
```

イメージのpullに失敗した場合や、arm64でコンテナが起動しない場合は、その事実とログを報告し、CIの統合jobを最終判定にする。テストを`t.Skip`で回避しない。

- [ ] Step 5: コミット `test(repo): restore mysql/sqlserver/oracle testcontainers integration tests`

---

### Task 4: frontendの4種復元

Files:
- Modify: `frontend/src/types/datasource.ts`
- Modify: `frontend/src/types/datasource.test.ts`
- Modify: `frontend/src/api/datasources.test.ts`
- Modify: `frontend/src/components/datasources/DataSourceForm.tsx`
- Modify: `frontend/src/components/datasources/DataSourceList.tsx`

Produces: `DBType = "postgresql" | "mysql" | "oracle" | "sqlserver"`、`DB_TYPE_LABELS`（PostgreSQL / MySQL / Oracle / SQL Server）、`DEFAULT_PORTS`（5432 / 3306 / 1521 / 1433）。

- [ ] Step 1: 復元元との差分を確認する

```bash
git -C /path/to/nocode-app diff c0bd653^1 HEAD -- frontend/src/types/datasource.ts frontend/src/types/datasource.test.ts frontend/src/api/datasources.test.ts frontend/src/components/datasources/DataSourceForm.tsx frontend/src/components/datasources/DataSourceList.tsx
```

- [ ] Step 2: テストを先に戻す。`datasource.test.ts`と`datasources.test.ts`に縮約前の4種ケースを復元し、`pnpm test`が失敗することを確認する。

- [ ] Step 3: `datasource.ts`のユニオン・ラベル・ポートを4種に戻す。`DataSourceForm.tsx`の`DB_TYPES`を`["postgresql", "mysql", "oracle", "sqlserver"]`に戻す。`DataSourceList.tsx`のバッジを`DB_TYPE_LABELS[ds.db_type]`で表示する形に戻す（縮約前の差分を参照）。コメントにある`PostgreSQL のみサポート`という文言を外す。

- [ ] Step 4: 検証

```bash
(cd frontend && pnpm install --frozen-lockfile && pnpm run typecheck && pnpm run lint && pnpm run format:check && pnpm test -- --run && pnpm run build)
```

- [ ] Step 5: コミット `feat(frontend): restore mysql/oracle/sqlserver data source types`

---

### Task 5: CI統合job追加とREADME更新

Files:
- Modify: `.github/workflows/ci.yml`
- Modify: `README.md`

- [ ] Step 1: `ci.yml`の`backend` jobの直後に次のjobを追加する

```yaml
  backend-integration:
    name: Backend integration (testcontainers, 4 RDB)
    uses: okamyuji/reusable-workflows/.github/workflows/go-ci.yml@v1
    with:
      working-directory: backend
      test-command: "go test -race -count=1 -tags=integration -timeout 40m ./internal/repositories/..."
```

- [ ] Step 2: `README.md`を更新する。
  - 12行目付近の`DB 一択化（2026-04-25）`で始まる注記を、保管DBはPostgreSQL固定、外部データソースは4種対応である旨に書き換える。
  - 外部データソース節の対応RDB表を4行（PostgreSQL lib/pq、MySQL go-sql-driver/mysql、Oracle go-ora、SQL Server microsoft/go-mssqldb）にする。
  - `MySQL / Oracle / SQL Server のサポートは廃止されました`と書かれた段落を削除する。
  - `db_type`列の説明を`CHECK (db_type IN ('postgresql','mysql','oracle','sqlserver'))`にする。
  - 経緯や履歴は書かず、現在の仕様だけを書く。

- [ ] Step 3: 検証: `git diff --stat`が`ci.yml`と`README.md`のみ。YAMLは`python3 -c 'import yaml,sys; yaml.safe_load(open(".github/workflows/ci.yml"))'`で構文確認。

- [ ] Step 4: コミット `ci: run 4-RDB testcontainers integration tests; docs: update supported data source RDBs`
