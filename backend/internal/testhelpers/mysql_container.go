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

	_ "github.com/go-sql-driver/mysql"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
)

var (
	mysqlContainer *mysql.MySQLContainer
	testDB         *bun.DB
	containerOnce  sync.Once
	containerErr   error
	dbMutex        sync.Mutex
)

const (
	testDBName   = "nocode-app-test"
	testUser     = "test"
	testPassword = "testpass"
)

// GetTestDB 共有テストデータベース接続を返す
// MySQLコンテナは一度だけ起動され、全テストで再利用される
func GetTestDB(ctx context.Context) (*bun.DB, error) {
	containerOnce.Do(func() {
		containerErr = startMySQLContainer(ctx)
	})

	if containerErr != nil {
		return nil, containerErr
	}

	return testDB, nil
}

// startMySQLContainer テスト用MySQLコンテナを起動する
func startMySQLContainer(ctx context.Context) error {
	var err error
	mysqlContainer, err = mysql.Run(ctx,
		"mysql:8.0",
		mysql.WithDatabase(testDBName),
		mysql.WithUsername(testUser),
		mysql.WithPassword(testPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("ready for connections").
				WithOccurrence(2).
				WithStartupTimeout(120*time.Second),
		),
	)
	if err != nil {
		return fmt.Errorf("MySQLコンテナの起動に失敗しました: %w", err)
	}

	// 接続文字列を取得
	connStr, err := mysqlContainer.ConnectionString(ctx, "parseTime=true")
	if err != nil {
		return fmt.Errorf("接続文字列の取得に失敗しました: %w", err)
	}

	// リトライ付きでデータベースに接続
	var sqldb *sql.DB
	for i := 0; i < 10; i++ {
		sqldb, err = sql.Open("mysql", connStr)
		if err != nil {
			return fmt.Errorf("データベースのオープンに失敗しました: %w", err)
		}

		// コネクションプールの設定
		sqldb.SetMaxOpenConns(10)
		sqldb.SetMaxIdleConns(5)
		sqldb.SetConnMaxLifetime(time.Hour)

		// 接続テスト
		if pingErr := sqldb.PingContext(ctx); pingErr == nil {
			break
		}
		_ = sqldb.Close()
		time.Sleep(time.Second)
	}

	// 接続が機能することを確認
	if pingErr := sqldb.PingContext(ctx); pingErr != nil {
		return fmt.Errorf("リトライ後のデータベース接続に失敗しました: %w", pingErr)
	}

	// Bun DBを作成
	testDB = bun.NewDB(sqldb, mysqldialect.New())

	// マイグレーションを実行
	if err := runMigrations(ctx, sqldb); err != nil {
		return fmt.Errorf("マイグレーションの実行に失敗しました: %w", err)
	}

	return nil
}

// runMigrations init.sqlマイグレーションファイルを実行する
func runMigrations(ctx context.Context, db *sql.DB) error {
	// マイグレーションディレクトリを検索
	migrationPath := findMigrationsPath()
	if migrationPath == "" {
		return fmt.Errorf("マイグレーションディレクトリが見つかりません")
	}

	// init.sqlを読み込み
	initSQL, err := os.ReadFile(filepath.Join(migrationPath, "init.sql"))
	if err != nil {
		return fmt.Errorf("init.sqlの読み込みに失敗しました: %w", err)
	}

	// 各ステートメントを分割して個別に実行
	// MySQLドライバーはデフォルトでマルチステートメントをサポートしない
	statements := splitSQLStatements(string(initSQL))
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		_, err = db.ExecContext(ctx, stmt)
		if err != nil {
			return fmt.Errorf("ステートメントの実行に失敗しました: %w\nステートメント: %s", err, stmt)
		}
	}

	return nil
}

// splitSQLStatements SQLファイルの内容を個別のステートメントに分割する
func splitSQLStatements(sql string) []string {
	var statements []string
	var current strings.Builder
	inString := false
	stringChar := byte(0)

	for i := 0; i < len(sql); i++ {
		c := sql[i]

		// 文字列リテラルの処理
		if (c == '\'' || c == '"') && (i == 0 || sql[i-1] != '\\') {
			if !inString {
				inString = true
				stringChar = c
			} else if c == stringChar {
				inString = false
			}
		}

		// コメントの処理
		if !inString && c == '-' && i+1 < len(sql) && sql[i+1] == '-' {
			// 行末までスキップ
			for i < len(sql) && sql[i] != '\n' {
				i++
			}
			continue
		}

		// ステートメント終端の処理
		if !inString && c == ';' {
			statements = append(statements, current.String())
			current.Reset()
			continue
		}

		current.WriteByte(c)
	}

	// 残りの内容を追加
	if remaining := strings.TrimSpace(current.String()); remaining != "" {
		statements = append(statements, remaining)
	}

	return statements
}

// findMigrationsPath マイグレーションディレクトリを検索する
func findMigrationsPath() string {
	// このファイルのディレクトリを取得
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}

	// internal/testhelpersからbackend/migrationsに移動
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

// CleanupContainer MySQLコンテナを終了する
// 全テスト完了後のTestMainで呼び出す必要がある
func CleanupContainer(ctx context.Context) error {
	if mysqlContainer != nil {
		return mysqlContainer.Terminate(ctx)
	}
	return nil
}

// CleanupTables デフォルト管理者ユーザーを保持しながら全データテーブルをトランケートする
// テスト間の分離を確保するために呼び出す必要がある
func CleanupTables(ctx context.Context) error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if testDB == nil {
		return fmt.Errorf("テストデータベースが初期化されていません")
	}

	// 外部キーチェックを一時的に無効化
	if _, err := testDB.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 0"); err != nil {
		return err
	}

	// トランケートするテーブル（外部キーを考慮した順序）
	tables := []string{
		"chart_configs",
		"app_views",
		"app_fields",
		"apps",
	}

	for _, table := range tables {
		if _, err := testDB.ExecContext(ctx, fmt.Sprintf("TRUNCATE TABLE %s", table)); err != nil {
			return fmt.Errorf("%sのトランケートに失敗しました: %w", table, err)
		}
	}

	// 管理者以外のユーザーを削除
	if _, err := testDB.ExecContext(ctx, "DELETE FROM users WHERE email != 'admin@example.com'"); err != nil {
		return fmt.Errorf("ユーザーのクリーンアップに失敗しました: %w", err)
	}

	// 外部キーチェックを再度有効化
	if _, err := testDB.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 1"); err != nil {
		return err
	}

	return nil
}

// DropDynamicTables 全ての動的アプリテーブル（app_data_*）を削除する
func DropDynamicTables(ctx context.Context) error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if testDB == nil {
		return fmt.Errorf("テストデータベースが初期化されていません")
	}

	// 全ての動的テーブルを検索
	rows, err := testDB.QueryContext(ctx, "SHOW TABLES LIKE 'app_data_%'")
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return err
		}
		tables = append(tables, tableName)
	}

	// 各動的テーブルを削除
	for _, table := range tables {
		if _, err := testDB.ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS `%s`", table)); err != nil {
			return fmt.Errorf("%sの削除に失敗しました: %w", table, err)
		}
	}

	return nil
}

// ResetDatabase 完全なリセットを実行: テーブルのクリーンアップと動的テーブルの削除
func ResetDatabase(ctx context.Context) error {
	if err := DropDynamicTables(ctx); err != nil {
		return err
	}
	return CleanupTables(ctx)
}
