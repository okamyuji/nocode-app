package testhelpers

import (
	"database/sql"
	"fmt"

	_ "github.com/denisenkom/go-mssqldb" // SQL Server driver
	_ "github.com/go-sql-driver/mysql"   // MySQL driver
	_ "github.com/lib/pq"                // PostgreSQL driver
	_ "github.com/sijms/go-ora/v2"       // Oracle driver
)

// openTestDB テスト用のデータベース接続を開く
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

