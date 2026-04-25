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
