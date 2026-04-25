package repositories

import (
	"testing"

	"nocode-app/backend/internal/models"

	"github.com/stretchr/testify/assert"
)

// TestQuoteIdentifierForDB 識別子クォート（PostgreSQL）をテストする
func TestQuoteIdentifierForDB(t *testing.T) {
	tests := []struct {
		name     string
		dbType   models.DBType
		input    string
		expected string
	}{
		{
			name:     "PostgreSQL: 英語の単純な名前",
			dbType:   models.DBTypePostgreSQL,
			input:    "users",
			expected: `"users"`,
		},
		{
			name:     "PostgreSQL: 日本語テーブル名",
			dbType:   models.DBTypePostgreSQL,
			input:    "顧客マスタ",
			expected: `"顧客マスタ"`,
		},
		{
			name:     "PostgreSQL: 日本語カラム名",
			dbType:   models.DBTypePostgreSQL,
			input:    "プロセス名",
			expected: `"プロセス名"`,
		},
		{
			name:     "PostgreSQL: ダブルクォートを含む名前",
			dbType:   models.DBTypePostgreSQL,
			input:    `user"name`,
			expected: `"user""name"`,
		},
		{
			name:     "PostgreSQL: スペースを含む名前",
			dbType:   models.DBTypePostgreSQL,
			input:    "user name",
			expected: `"user name"`,
		},
		{
			name:     "PostgreSQL: 予約語",
			dbType:   models.DBTypePostgreSQL,
			input:    "select",
			expected: `"select"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := quoteIdentifierForDB(tt.dbType, tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGetPlaceholder PostgreSQL の $N プレースホルダをテストする
func TestGetPlaceholder(t *testing.T) {
	tests := []struct {
		name     string
		index    int
		expected string
	}{
		{name: "index 1", index: 1, expected: "$1"},
		{name: "index 5", index: 5, expected: "$5"},
		{name: "index 100", index: 100, expected: "$100"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPlaceholder(models.DBTypePostgreSQL, tt.index)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestBuildLimitOffset PostgreSQL の LIMIT/OFFSET 句をテストする
func TestBuildLimitOffset(t *testing.T) {
	tests := []struct {
		name     string
		limit    int
		offset   int
		expected string
	}{
		{name: "limit 10 offset 0", limit: 10, offset: 0, expected: " LIMIT 10 OFFSET 0"},
		{name: "limit 20 offset 40", limit: 20, offset: 40, expected: " LIMIT 20 OFFSET 40"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildLimitOffset(models.DBTypePostgreSQL, tt.limit, tt.offset)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestBuildDSN PostgreSQL の DSN 構築と非対応 DB のエラーをテストする
func TestBuildDSN(t *testing.T) {
	tests := []struct {
		name           string
		dataSource     *models.DataSource
		password       string
		expectedDriver string
		expectedDSN    string
		expectedError  bool
	}{
		{
			name: "PostgreSQL: 基本的なDSN",
			dataSource: &models.DataSource{
				DBType:       models.DBTypePostgreSQL,
				Host:         "localhost",
				Port:         5432,
				Username:     "testuser",
				DatabaseName: "testdb",
			},
			password:       "testpass",
			expectedDriver: "postgres",
			expectedDSN:    "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable",
			expectedError:  false,
		},
		{
			name: "PostgreSQL: 特殊文字を含むパスワード",
			dataSource: &models.DataSource{
				DBType:       models.DBTypePostgreSQL,
				Host:         "localhost",
				Port:         5432,
				Username:     "testuser",
				DatabaseName: "testdb",
			},
			password:       "test'pass\\word",
			expectedDriver: "postgres",
			expectedDSN:    "host=localhost port=5432 user=testuser password=test\\'pass\\\\word dbname=testdb sslmode=disable",
			expectedError:  false,
		},
		{
			name: "PostgreSQL: 日本語データベース名",
			dataSource: &models.DataSource{
				DBType:       models.DBTypePostgreSQL,
				Host:         "localhost",
				Port:         5432,
				Username:     "testuser",
				DatabaseName: "テストDB",
			},
			password:       "testpass",
			expectedDriver: "postgres",
			expectedDSN:    "host=localhost port=5432 user=testuser password=testpass dbname=テストDB sslmode=disable",
			expectedError:  false,
		},
		{
			name: "Unknown: エラーを返す",
			dataSource: &models.DataSource{
				DBType:       "unknown",
				Host:         "localhost",
				Port:         5432,
				Username:     "testuser",
				DatabaseName: "testdb",
			},
			password:      "testpass",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			driver, dsn, err := buildDSN(tt.dataSource, tt.password)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedDriver, driver)
				assert.Equal(t, tt.expectedDSN, dsn)
			}
		})
	}
}

// TestEscapePostgresPassword PostgreSQLパスワードエスケープをテストする
func TestEscapePostgresPassword(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "通常のパスワード", input: "password123", expected: "password123"},
		{name: "シングルクォートを含む", input: "pass'word", expected: "pass\\'word"},
		{name: "バックスラッシュを含む", input: "pass\\word", expected: "pass\\\\word"},
		{name: "両方を含む", input: "pass'\\word", expected: "pass\\'\\\\word"},
		{name: "日本語パスワード", input: "パスワード123", expected: "パスワード123"},
		{name: "空文字列", input: "", expected: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapePostgresPassword(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestConvertScannedValue データベースからスキャンした値の変換をテストする
func TestConvertScannedValue(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{name: "nil値", input: nil, expected: nil},
		{name: "string値", input: "テスト文字列", expected: "テスト文字列"},
		{name: "[]byte値（日本語）", input: []byte("日本語テキスト"), expected: "日本語テキスト"},
		{name: "int64値", input: int64(12345), expected: int64(12345)},
		{name: "float64値", input: float64(123.45), expected: float64(123.45)},
		{name: "bool値 true", input: true, expected: true},
		{name: "bool値 false", input: false, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertScannedValue(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestQuoteIdentifierForDBWithJapaneseEdgeCases 日本語の境界ケースをテストする
func TestQuoteIdentifierForDBWithJapaneseEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "PostgreSQL: ひらがな", input: "てすと"},
		{name: "PostgreSQL: カタカナ", input: "テスト"},
		{name: "PostgreSQL: 漢字", input: "顧客管理"},
		{name: "PostgreSQL: 全角数字", input: "テーブル１２３"},
		{name: "PostgreSQL: 全角記号", input: "テスト＿テーブル"},
		{name: "PostgreSQL: 混合名", input: "user_テーブル_123"},
		{name: "PostgreSQL: 長い日本語名", input: "非常に長い日本語のテーブル名前をテストする"},
		{name: "PostgreSQL: 絵文字を含む", input: "テスト😀テーブル"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := quoteIdentifierForDB(models.DBTypePostgreSQL, tt.input)
			assert.NotEmpty(t, result)
			assert.Contains(t, result, tt.input)
		})
	}
}
