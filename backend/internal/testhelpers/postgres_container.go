package testhelpers

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// PostgresTestContainer PostgreSQLテストコンテナの設定
type PostgresTestContainer struct {
	Container *postgres.PostgresContainer
	Host      string
	Port      int
	Database  string
	Username  string
	Password  string
}

// SetupPostgresContainer PostgreSQLテストコンテナをセットアップする
func SetupPostgresContainer(ctx context.Context) (*PostgresTestContainer, error) {
	dbName := "testdb"
	dbUser := "testuser"
	dbPassword := "testpass"

	container, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second)),
	)
	if err != nil {
		return nil, fmt.Errorf("PostgreSQLコンテナの起動に失敗しました: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("ホストの取得に失敗しました: %w", err)
	}

	mappedPort, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return nil, fmt.Errorf("ポートの取得に失敗しました: %w", err)
	}

	return &PostgresTestContainer{
		Container: container,
		Host:      host,
		Port:      mappedPort.Int(),
		Database:  dbName,
		Username:  dbUser,
		Password:  dbPassword,
	}, nil
}

// Terminate コンテナを終了する
func (p *PostgresTestContainer) Terminate(ctx context.Context) error {
	if p.Container != nil {
		return p.Container.Terminate(ctx)
	}
	return nil
}

// CreateTestTable テスト用のテーブルを作成する
func (p *PostgresTestContainer) CreateTestTable(ctx context.Context) error {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		p.Host, p.Port, p.Username, p.Password, p.Database)

	db, err := openTestDB("postgres", connStr)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	// テストテーブルを作成
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS test_table (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			email VARCHAR(255),
			age INTEGER,
			salary DECIMAL(10, 2),
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("テストテーブルの作成に失敗しました: %w", err)
	}

	// テストデータを挿入
	_, err = db.ExecContext(ctx, `
		INSERT INTO test_table (name, email, age, salary, is_active) VALUES
		('Alice', 'alice@example.com', 30, 50000.00, true),
		('Bob', 'bob@example.com', 25, 45000.00, true),
		('Charlie', 'charlie@example.com', 35, 60000.00, false)
	`)
	if err != nil {
		return fmt.Errorf("テストデータの挿入に失敗しました: %w", err)
	}

	return nil
}

// CreateTestView テスト用のビューを作成する
func (p *PostgresTestContainer) CreateTestView(ctx context.Context) error {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		p.Host, p.Port, p.Username, p.Password, p.Database)

	db, err := openTestDB("postgres", connStr)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	// テストビューを作成（アクティブなユーザーのみを表示）
	_, err = db.ExecContext(ctx, `
		CREATE OR REPLACE VIEW test_view AS
		SELECT id, name, email, age, salary
		FROM test_table
		WHERE is_active = true
	`)
	if err != nil {
		return fmt.Errorf("テストビューの作成に失敗しました: %w", err)
	}

	return nil
}
