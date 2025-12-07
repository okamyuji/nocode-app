package testhelpers

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mssql"
	"github.com/testcontainers/testcontainers-go/wait"
)

// MSSQLTestContainer SQL Serverテストコンテナの設定
type MSSQLTestContainer struct {
	Container *mssql.MSSQLServerContainer
	Host      string
	Port      int
	Database  string
	Username  string
	Password  string
}

// SetupMSSQLContainer SQL Serverテストコンテナをセットアップする
func SetupMSSQLContainer(ctx context.Context) (*MSSQLTestContainer, error) {
	dbPassword := "TestP@ssw0rd!"

	container, err := mssql.Run(ctx,
		"mcr.microsoft.com/mssql/server:2022-latest",
		mssql.WithAcceptEULA(),
		mssql.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("SQL Server is now ready for client connections").
				WithStartupTimeout(120*time.Second)),
	)
	if err != nil {
		return nil, fmt.Errorf("SQL Serverコンテナの起動に失敗しました: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("ホストの取得に失敗しました: %w", err)
	}

	mappedPort, err := container.MappedPort(ctx, "1433")
	if err != nil {
		return nil, fmt.Errorf("ポートの取得に失敗しました: %w", err)
	}

	return &MSSQLTestContainer{
		Container: container,
		Host:      host,
		Port:      mappedPort.Int(),
		Database:  "master",
		Username:  "sa",
		Password:  dbPassword,
	}, nil
}

// Terminate コンテナを終了する
func (m *MSSQLTestContainer) Terminate(ctx context.Context) error {
	if m.Container != nil {
		return m.Container.Terminate(ctx)
	}
	return nil
}

// CreateTestTable テスト用のテーブルを作成する
func (m *MSSQLTestContainer) CreateTestTable(ctx context.Context) error {
	// パスワードにURLエンコードが必要
	connStr := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
		url.QueryEscape(m.Username), url.QueryEscape(m.Password), m.Host, m.Port, m.Database)

	// SQL Serverの接続が安定するまでリトライ
	var db *sql.DB
	var err error
	for i := 0; i < 10; i++ {
		db, err = openTestDB("sqlserver", connStr)
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	// テストデータベースを作成
	_, err = db.ExecContext(ctx, `
		IF NOT EXISTS (SELECT name FROM sys.databases WHERE name = 'testdb')
		BEGIN
			CREATE DATABASE testdb
		END
	`)
	if err != nil {
		return fmt.Errorf("テストデータベースの作成に失敗しました: %w", err)
	}

	// testdbに接続し直す
	m.Database = "testdb"
	connStr = fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
		url.QueryEscape(m.Username), url.QueryEscape(m.Password), m.Host, m.Port, m.Database)

	db2, err := openTestDB("sqlserver", connStr)
	if err != nil {
		return err
	}
	defer func() { _ = db2.Close() }()

	// テストテーブルを作成
	_, err = db2.ExecContext(ctx, `
		IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'test_table')
		BEGIN
			CREATE TABLE test_table (
				id INT IDENTITY(1,1) PRIMARY KEY,
				name NVARCHAR(100) NOT NULL,
				email NVARCHAR(255),
				age INT,
				salary DECIMAL(10, 2),
				is_active BIT DEFAULT 1,
				created_at DATETIME DEFAULT GETDATE()
			)
		END
	`)
	if err != nil {
		return fmt.Errorf("テストテーブルの作成に失敗しました: %w", err)
	}

	// テストデータを挿入
	_, err = db2.ExecContext(ctx, `
		IF NOT EXISTS (SELECT * FROM test_table)
		BEGIN
			INSERT INTO test_table (name, email, age, salary, is_active) VALUES
			('Alice', 'alice@example.com', 30, 50000.00, 1),
			('Bob', 'bob@example.com', 25, 45000.00, 1),
			('Charlie', 'charlie@example.com', 35, 60000.00, 0)
		END
	`)
	if err != nil {
		return fmt.Errorf("テストデータの挿入に失敗しました: %w", err)
	}

	return nil
}
