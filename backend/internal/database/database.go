package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	// MySQLドライバーのインポート - sql.Open("mysql", ...) に必要
	_ "github.com/go-sql-driver/mysql"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"

	"nocode-app/backend/internal/config"
)

// NewDB 新しいデータベース接続を作成する
func NewDB(cfg *config.DBConfig) (*bun.DB, error) {
	sqldb, err := sql.Open("mysql", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("データベースのオープンに失敗しました: %w", err)
	}

	// コネクションプールの設定
	sqldb.SetMaxOpenConns(cfg.MaxOpenConns)
	sqldb.SetMaxIdleConns(cfg.MaxIdleConns)
	sqldb.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// コンテキスト付きで接続を確認
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqldb.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("データベースへのPingに失敗しました: %w", err)
	}

	// BUNデータベースインスタンスを作成
	db := bun.NewDB(sqldb, mysqldialect.New())

	return db, nil
}

// Close データベース接続を閉じる
func Close(db *bun.DB) error {
	return db.Close()
}
