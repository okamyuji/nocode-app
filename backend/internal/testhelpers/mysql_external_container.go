package testhelpers

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go/wait"
)

// MySQLExternalTestContainer 外部データソーステスト用MySQLコンテナの設定
type MySQLExternalTestContainer struct {
	Container *mysql.MySQLContainer
	Host      string
	Port      int
	Database  string
	Username  string
	Password  string
}

// SetupMySQLExternalContainer 外部データソーステスト用MySQLコンテナをセットアップする
func SetupMySQLExternalContainer(ctx context.Context) (*MySQLExternalTestContainer, error) {
	dbName := "external_testdb"
	dbUser := "extuser"
	dbPassword := "extpass"

	container, err := mysql.Run(ctx,
		"mysql:8.0",
		mysql.WithDatabase(dbName),
		mysql.WithUsername(dbUser),
		mysql.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("ready for connections").
				WithOccurrence(2).
				WithStartupTimeout(120*time.Second)),
	)
	if err != nil {
		return nil, fmt.Errorf("MySQLコンテナの起動に失敗しました: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("ホストの取得に失敗しました: %w", err)
	}

	mappedPort, err := container.MappedPort(ctx, "3306")
	if err != nil {
		return nil, fmt.Errorf("ポートの取得に失敗しました: %w", err)
	}

	return &MySQLExternalTestContainer{
		Container: container,
		Host:      host,
		Port:      mappedPort.Int(),
		Database:  dbName,
		Username:  dbUser,
		Password:  dbPassword,
	}, nil
}

// Terminate コンテナを終了する
func (m *MySQLExternalTestContainer) Terminate(ctx context.Context) error {
	if m.Container != nil {
		return m.Container.Terminate(ctx)
	}
	return nil
}

// CreateTestTable テスト用のテーブルを作成する
func (m *MySQLExternalTestContainer) CreateTestTable(ctx context.Context) error {
	connStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		m.Username, m.Password, m.Host, m.Port, m.Database)

	// MySQLの接続が安定するまでリトライ
	var db *sql.DB
	var err error
	for i := 0; i < 10; i++ {
		db, err = openTestDB("mysql", connStr)
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	// テストテーブルを作成
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS test_table (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			email VARCHAR(255),
			age INT,
			salary DECIMAL(10, 2),
			is_active TINYINT(1) DEFAULT 1,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("テストテーブルの作成に失敗しました: %w", err)
	}

	// テストデータを挿入
	_, err = db.ExecContext(ctx, `
		INSERT INTO test_table (name, email, age, salary, is_active) VALUES
		('Alice', 'alice@example.com', 30, 50000.00, 1),
		('Bob', 'bob@example.com', 25, 45000.00, 1),
		('Charlie', 'charlie@example.com', 35, 60000.00, 0)
	`)
	if err != nil {
		return fmt.Errorf("テストデータの挿入に失敗しました: %w", err)
	}

	return nil
}
