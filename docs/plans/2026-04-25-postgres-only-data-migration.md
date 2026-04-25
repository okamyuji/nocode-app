# データ移行計画: MySQL → PostgreSQL

> 対象: 既存の MySQL 環境で稼働している `nocode-app` データを PostgreSQL に移し替える手順。
> ローカル開発のみであれば「§4 クリーン再構築」で十分。本番・ステージングは「§5 データ保全移行」を参照。

---

## 1. 前提

- 移行対象データの存在は環境ごとに異なる。**本番に未デプロイ**の場合は「§4 クリーン再構築」のみで完了する。
- 動的に作られた `app_<id>` テーブルは事前に列挙不能。マイグレーションスクリプトはメタデータ (`apps`, `app_fields`) から各テーブルの DDL を再生成し、データを CSV 経由で移送する。
- 全工程はメンテナンス時間内に行う（書き込みを止める）。

---

## 2. 用語と参照値

| 名前 | 値 |
|---|---|
| 旧 DB | MySQL 8.0、`compose.yaml` の旧 `mysql` サービス、ボリューム `mysql_data` |
| 新 DB | PostgreSQL 16、`compose.yaml` の新 `postgres` サービス、ボリューム `postgres_data` |
| 旧 DB 名 | `nocode-app` |
| 新 DB 名 | `nocode-app` |
| ユーザー / パスワード | `nocode / nocodepassword` (env.example の値) |

---

## 3. 動作検証手順（必ず移行作業の最後に実施）

すべての移行手順の最後で以下が緑になることを確認する。バックエンド／フロントエンドの両方。

### 3.1 Backend
```bash
cd backend
go vet ./...
gofmt -l . | (! grep .)
go test -short ./...     # ユニット: すべて PASS
go test ./...            # 統合 (testcontainers): すべて PASS
```

### 3.2 Frontend
```bash
cd frontend
pnpm install
pnpm run typecheck
pnpm run lint
pnpm run format:check
pnpm test
pnpm run build
```

### 3.3 e2e (compose 起動)
```bash
docker compose up -d --build
sleep 15
docker compose ps                    # postgres / backend / frontend がすべて healthy
curl -fsS http://localhost:8080/health
# フロントを http://localhost:3000 で開いて admin@example.com / admin123 でログイン
docker compose down
```

---

## 4. クリーン再構築（本番未投入の場合）

ローカルで本番データを保持していない場合は、これだけで完了する。

```bash
# 1. 旧スタックを完全停止し、MySQL ボリュームも削除
docker compose down -v

# 2. main ブランチを最新化し、PostgreSQL 一択ブランチをチェックアウト済みであることを確認
git status

# 3. 環境変数を更新（POSTGRES_* と DB_PORT=5432 を反映）
cp env.example .env

# 4. PostgreSQL で起動
docker compose up -d --build

# 5. 動作検証手順 §3 を実施
```

---

## 5. データ保全移行（本番／ステージング向け）

### 5.1 全体フロー

1. メンテナンスモードに切替（書き込み停止）
2. 旧 MySQL からダンプ取得
3. ダンプを PostgreSQL 用に変換
4. 動的テーブル DDL を再生成
5. PostgreSQL に投入
6. シーケンス値を `MAX(id)+1` に揃える
7. 動作検証 §3
8. メンテナンス解除

### 5.2 Step 1: メンテナンスモード

- バックエンドの書き込みエンドポイントを 503 で返すよう nginx / LB で切替
- 確認: `curl -X POST http://app/api/v1/apps` が 503 を返す

### 5.3 Step 2: 旧 MySQL のダンプ

```bash
# 静的 6 テーブル + データソーステーブル + 動的テーブル全部
docker exec -i nocode-mysql mysqldump \
    --user=nocode --password=nocodepassword \
    --single-transaction --quick --skip-extended-insert \
    --no-create-info --complete-insert \
    nocode-app > /tmp/mysql-data.sql

# DDL は使い回さず PostgreSQL の init.sql を使うので、データのみ取得 (--no-create-info)
```

### 5.4 Step 3: データ変換

`pgloader` を使うのがもっとも確実。Docker で実行可能:

```bash
docker run --rm -it \
  --network host \
  -v $(pwd):/work \
  dimitri/pgloader:latest \
  pgloader \
    "mysql://nocode:nocodepassword@localhost:3306/nocode-app" \
    "postgresql://nocode:nocodepassword@localhost:5432/nocode-app"
```

`pgloader` は型変換（`tinyint(1)`→`boolean`、`json`→`jsonb`、`enum`→`varchar+CHECK`）を自動で行う。詳細は <https://pgloader.readthedocs.io/>。

⚠ 動的に作られた `app_<id>` テーブルも一緒に移送される。pgloader は MySQL 側の DDL から再構築するので、`BIGINT UNSIGNED AUTO_INCREMENT` は `BIGINT` + `serial` 相当に変換される。

### 5.5 Step 4: 静的スキーマの整合性確認

`init.sql` で定義された **CHECK 制約・トリガ・インデックス** は pgloader 経由では入らないため、インポート後に補正が必要。

⚠️ **重要**: `CREATE TABLE IF NOT EXISTS` は **テーブルが既に存在する場合は何もしない**（既存テーブルに新たな CHECK 制約を追加してくれない）。pgloader が先にテーブルを作っているため、init.sql を後から流してもインライン CHECK 制約は反映されない。トリガと `CREATE INDEX IF NOT EXISTS` の方は別ステートメントなので問題なく追加される。

選択肢のいずれかで進める:

**(a) pgloader 後に既存テーブルへ ALTER で CHECK を追加 (推奨)**

```bash
docker exec -i nocode-postgres psql -U nocode -d nocode-app <<'SQL'
-- pgloader が型を text にマップしている可能性があるため、必要に応じ TYPE も合わせる
ALTER TABLE users
    ADD CONSTRAINT users_role_check CHECK (role IN ('admin', 'user'));
ALTER TABLE data_sources
    ADD CONSTRAINT data_sources_db_type_check CHECK (db_type IN ('postgresql'));
ALTER TABLE app_views
    ADD CONSTRAINT app_views_view_type_check CHECK (view_type IN ('table', 'list', 'calendar', 'chart'));
ALTER TABLE dashboard_widgets
    ADD CONSTRAINT dashboard_widgets_view_type_check CHECK (view_type IN ('table', 'list', 'chart'));
ALTER TABLE dashboard_widgets
    ADD CONSTRAINT dashboard_widgets_widget_size_check CHECK (widget_size IN ('small', 'medium', 'large'));
SQL

# その後 init.sql を流してトリガ・インデックスを補完
docker exec -i nocode-postgres psql -U nocode -d nocode-app < backend/migrations/init.sql
```

**(b) pgloader が作ったテーブルを drop して init.sql で作り直す（より確実）**

```bash
# 静的テーブルだけ drop (動的テーブル app_data_* は残す)
docker exec -i nocode-postgres psql -U nocode -d nocode-app -c '
DROP TABLE IF EXISTS dashboard_widgets, chart_configs, app_views, app_fields, apps, data_sources, users CASCADE;'

# init.sql でクリーンに作成
docker exec -i nocode-postgres psql -U nocode -d nocode-app < backend/migrations/init.sql

# その上で pgloader し直す（データのみ、--no-create フラグ等で対応）
```

### 5.6 Step 5: シーケンス値リセット

pgloader 後、`BIGSERIAL` のシーケンスが既存データの最大 ID と同期していないことがある。次のスクリプトで全テーブル一括同期:

`backend/migrations/postmigration_reset_sequences.sql` を新規作成:

```sql
-- 各 BIGSERIAL のシーケンスを当該テーブルの MAX(id)+1 にリセット
DO $$
DECLARE
    t TEXT;
BEGIN
    FOR t IN
        SELECT tablename FROM pg_tables
        WHERE schemaname = 'public'
          AND tablename NOT IN ('pg_stat_statements')
    LOOP
        EXECUTE format(
            $f$
              SELECT setval(
                pg_get_serial_sequence(%L, 'id'),
                COALESCE((SELECT MAX(id) FROM %I), 0) + 1,
                false
              )
            $f$,
            t, t
        );
    END LOOP;
END$$;
```

実行:

```bash
docker exec -i nocode-postgres psql -U nocode -d nocode-app \
  < backend/migrations/postmigration_reset_sequences.sql
```

> このファイルは恒久的に残す必要はないが、再移行時に役立つので `backend/migrations/` に残す。

### 5.7 Step 6: 動作検証

- §3 の手順をすべて実施
- ログイン、アプリ一覧表示、新規アプリ作成、レコード追加・更新・削除、ダッシュボード表示、外部 PostgreSQL データソース接続テストが正常動作することを画面操作で確認

### 5.8 Step 7: メンテナンス解除

- nginx / LB を通常モードに戻す
- `docker compose logs --tail 200` で異常無いことを最終確認

---

## 6. ロールバック手順

問題発生時は以下の順で原状復帰:

1. メンテナンスモード ON（書き込み停止）
2. PostgreSQL コンテナを停止: `docker compose stop postgres`
3. ブランチを切り戻し: `git checkout main && git pull` （本変更前の `d11a532` 直前へ）
4. 旧 `compose.yaml` (MySQL 版) で起動: `docker compose up -d --build`
5. MySQL ボリューム `mysql_data` を残しておく前提（少なくとも 7 日間は破棄しない）
6. メンテナンス解除

---

## 7. 失敗時の代替手段

`pgloader` が型不一致でこける場合の代替:

### 7.1 個別テーブル CSV 転送（最終手段）

```bash
# MySQL 側から CSV エクスポート
docker exec nocode-mysql mysql -u nocode -pnocodepassword \
  -e "SELECT * FROM nocode-app.users INTO OUTFILE '/tmp/users.csv' \
      FIELDS TERMINATED BY ',' OPTIONALLY ENCLOSED BY '\"' \
      LINES TERMINATED BY '\n'"
docker cp nocode-mysql:/tmp/users.csv /tmp/users.csv

# PostgreSQL 側に COPY で投入
docker cp /tmp/users.csv nocode-postgres:/tmp/users.csv
docker exec -i nocode-postgres psql -U nocode -d nocode-app \
  -c "COPY users (id, email, password_hash, name, role, created_at, updated_at) \
      FROM '/tmp/users.csv' WITH (FORMAT csv);"
```

各テーブルについて同様に実施し、最後に §5.6 のシーケンスリセットを必ず行う。

---

## 8. 想定実行時間

| データ量 | クリーン再構築 (§4) | 保全移行 (§5) |
|---|---|---|
| 開発環境 (1 ユーザー、admin のみ) | 5 分 | n/a |
| ステージング (〜10 アプリ、〜1k レコード) | n/a | 15 分 |
| 本番 (〜100 アプリ、〜100k レコード) | n/a | 30〜60 分（pgloader 依存） |

---

## 9. チェックリスト（移行担当者用）

- [ ] 本書 §3 動作検証手順を**冒頭に**読んでおく
- [ ] 環境変数 `.env` を更新済み（`DB_PORT=5432`, `POSTGRES_*`）
- [ ] `docker compose down -v` で旧ボリュームを破棄するか、保全移行のためバックアップ取得済みか、**どちらかを明示的に選んだ**
- [ ] §5 を選んだ場合: `pgloader` 実行後にシーケンスリセット (§5.6) を必ず実行
- [ ] §3 の Backend / Frontend / e2e すべて PASS を確認
- [ ] PR の説明欄に本書へのリンクを貼った
