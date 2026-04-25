// Package testhelpers TestContainers管理を含むテストユーティリティを提供
package testhelpers

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

const (
	appTestDBName   = "nocode-app-test"
	appTestUser     = "test"
	appTestPassword = "testpass"
)

var (
	appPgContainer *postgres.PostgresContainer
	appTestDB      *bun.DB
	appOnce        sync.Once
	appErr         error
	appDBMutex     sync.Mutex
)

// GetTestDB アプリ保管 DB のテスト用接続 (シングルトン) を返す。
// 初回呼び出し時に PostgreSQL コンテナを起動し、init.sql でスキーマを作成する。
func GetTestDB(ctx context.Context) (*bun.DB, error) {
	appOnce.Do(func() {
		appErr = startAppDBContainer(ctx)
	})
	if appErr != nil {
		return nil, appErr
	}
	return appTestDB, nil
}

// CleanupContainer アプリ DB コンテナを終了する。TestMain で呼び出すこと。
func CleanupContainer(ctx context.Context) error {
	if appPgContainer != nil {
		return appPgContainer.Terminate(ctx)
	}
	return nil
}

// ResetDatabase 全テーブルをトランケートし、動的アプリテーブルを drop する。
// 管理者ユーザー (admin@example.com) は再投入する。
func ResetDatabase(ctx context.Context) error {
	appDBMutex.Lock()
	defer appDBMutex.Unlock()

	if appTestDB == nil {
		return fmt.Errorf("テストデータベースが初期化されていません")
	}

	// 動的テーブル (app_data_*) を drop
	if err := dropDynamicTablesLocked(ctx); err != nil {
		return err
	}

	// 静的テーブルを TRUNCATE
	if _, err := appTestDB.ExecContext(ctx,
		`TRUNCATE "dashboard_widgets", "chart_configs", "app_views", "app_fields", "apps", "data_sources", "users" RESTART IDENTITY CASCADE`); err != nil {
		return fmt.Errorf("TRUNCATE 失敗: %w", err)
	}

	// 管理者を再投入
	if _, err := appTestDB.ExecContext(ctx, `
		INSERT INTO users (email, password_hash, name, role) VALUES
		('admin@example.com', '$2a$10$e8i3egbnenpqzZlow/3Q0.5L6uN8vNyktEYkgRdWwP13xSkCtR1re', 'Admin', 'admin')
		ON CONFLICT (email) DO NOTHING`); err != nil {
		return fmt.Errorf("管理者の再投入に失敗: %w", err)
	}
	return nil
}

func dropDynamicTablesLocked(ctx context.Context) error {
	// LIKE の `_` はワイルドカードなので、リテラル下線にエスケープする。
	// "app_data_" 以外で始まるテーブル (例: app_dataset_cache) を誤って
	// 巻き込んで DROP しないため。
	rows, err := appTestDB.QueryContext(ctx, `
		SELECT tablename FROM pg_tables
		WHERE schemaname = 'public'
			AND tablename LIKE 'app\_data\_%' ESCAPE '\'`)
	if err != nil {
		return fmt.Errorf("動的テーブル一覧取得に失敗: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var dyn []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return err
		}
		dyn = append(dyn, t)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, t := range dyn {
		quoted := `"` + strings.ReplaceAll(t, `"`, `""`) + `"`
		if _, err := appTestDB.ExecContext(ctx, "DROP TABLE IF EXISTS "+quoted+" CASCADE"); err != nil {
			return fmt.Errorf("DROP TABLE %s 失敗: %w", t, err)
		}
	}
	return nil
}

func startAppDBContainer(ctx context.Context) error {
	c, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase(appTestDBName),
		postgres.WithUsername(appTestUser),
		postgres.WithPassword(appTestPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(120*time.Second),
		),
	)
	if err != nil {
		return fmt.Errorf("PostgreSQLコンテナの起動に失敗しました: %w", err)
	}
	appPgContainer = c

	connStr, err := c.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return fmt.Errorf("接続文字列の取得に失敗しました: %w", err)
	}

	var sqldb *sql.DB
	for i := 0; i < 10; i++ {
		sqldb, err = sql.Open("postgres", connStr)
		if err != nil {
			return fmt.Errorf("データベースのオープンに失敗しました: %w", err)
		}
		sqldb.SetMaxOpenConns(10)
		sqldb.SetMaxIdleConns(5)
		sqldb.SetConnMaxLifetime(time.Hour)

		pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		err = sqldb.PingContext(pingCtx)
		cancel()
		if err == nil {
			break
		}
		_ = sqldb.Close()
		time.Sleep(time.Second)
	}
	if err != nil {
		return fmt.Errorf("リトライ後のデータベース接続に失敗しました: %w", err)
	}

	appTestDB = bun.NewDB(sqldb, pgdialect.New())

	if err := runAppMigrations(ctx, sqldb); err != nil {
		return fmt.Errorf("マイグレーションの実行に失敗しました: %w", err)
	}
	return nil
}

func runAppMigrations(ctx context.Context, db *sql.DB) error {
	migrationPath := findMigrationsPath()
	if migrationPath == "" {
		return fmt.Errorf("マイグレーションディレクトリが見つかりません")
	}

	initSQL, err := os.ReadFile(filepath.Join(migrationPath, "init.sql"))
	if err != nil {
		return fmt.Errorf("init.sqlの読み込みに失敗しました: %w", err)
	}

	// PostgreSQL は libpq 経由でマルチステートメントを許容するため、
	// init.sql を一括投入する。CREATE FUNCTION / DO ブロック内のセミコロンを
	// 雑に分割すると壊れるため、分割せずに渡す。
	if _, err := db.ExecContext(ctx, string(initSQL)); err != nil {
		return fmt.Errorf("init.sql の実行に失敗しました: %w", err)
	}
	return nil
}

func findMigrationsPath() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}
	dir := filepath.Dir(filename)
	for i := 0; i < 10; i++ {
		migrationPath := filepath.Join(dir, "migrations")
		if _, err := os.Stat(migrationPath); err == nil {
			return migrationPath
		}
		dir = filepath.Dir(dir)
	}
	return ""
}

// openTestDB ドライバ名と DSN から *sql.DB を開く。外部 DB テストフィクスチャから利用される。
func openTestDB(driverName, dsn string) (*sql.DB, error) {
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("データベース接続の初期化に失敗しました: %w", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("データベースへの接続に失敗しました: %w", err)
	}
	return db, nil
}
