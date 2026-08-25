package repositories

import (
	"strings"
	"testing"

	"nocode-app/backend/internal/models"

	"github.com/stretchr/testify/assert"
)

// TestQuoteIdentifierForDB 各データベースタイプでの識別子クォートをテストする
func TestQuoteIdentifierForDB(t *testing.T) {
	tests := []struct {
		name     string
		dbType   models.DBType
		input    string
		expected string
	}{
		// PostgreSQL テスト
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

		// MySQL テスト
		{
			name:     "MySQL: 英語の単純な名前",
			dbType:   models.DBTypeMySQL,
			input:    "users",
			expected: "`users`",
		},
		{
			name:     "MySQL: 日本語テーブル名",
			dbType:   models.DBTypeMySQL,
			input:    "顧客マスタ",
			expected: "`顧客マスタ`",
		},
		{
			name:     "MySQL: 日本語カラム名",
			dbType:   models.DBTypeMySQL,
			input:    "プロセス名",
			expected: "`プロセス名`",
		},
		{
			name:     "MySQL: バッククォートを含む名前",
			dbType:   models.DBTypeMySQL,
			input:    "user`name",
			expected: "`user``name`",
		},
		{
			name:     "MySQL: スペースを含む名前",
			dbType:   models.DBTypeMySQL,
			input:    "user name",
			expected: "`user name`",
		},
		{
			name:     "MySQL: 予約語",
			dbType:   models.DBTypeMySQL,
			input:    "select",
			expected: "`select`",
		},

		// Oracle テスト（大文字変換あり）
		{
			name:     "Oracle: 英語の単純な名前",
			dbType:   models.DBTypeOracle,
			input:    "users",
			expected: `"USERS"`,
		},
		{
			name:     "Oracle: 日本語テーブル名（大文字変換なし）",
			dbType:   models.DBTypeOracle,
			input:    "顧客マスタ",
			expected: `"顧客マスタ"`,
		},
		{
			name:     "Oracle: 日本語カラム名（大文字変換なし）",
			dbType:   models.DBTypeOracle,
			input:    "プロセス名",
			expected: `"プロセス名"`,
		},
		{
			name:     "Oracle: 混合（大文字英語+日本語）",
			dbType:   models.DBTypeOracle,
			input:    "SPR2_プロセスマスタ",
			expected: `"SPR2_プロセスマスタ"`,
		},
		{
			name:     "Oracle: 混合（小文字英語+日本語）大文字変換を確認",
			dbType:   models.DBTypeOracle,
			input:    "spr2_プロセスマスタ",
			expected: `"SPR2_プロセスマスタ"`,
		},
		{
			name:     "Oracle: ダブルクォートを含む名前",
			dbType:   models.DBTypeOracle,
			input:    `user"name`,
			expected: `"USER""NAME"`,
		},
		{
			name:     "Oracle: スペースを含む名前",
			dbType:   models.DBTypeOracle,
			input:    "user name",
			expected: `"USER NAME"`,
		},

		// SQL Server テスト
		{
			name:     "SQLServer: 英語の単純な名前",
			dbType:   models.DBTypeSQLServer,
			input:    "users",
			expected: "[users]",
		},
		{
			name:     "SQLServer: 日本語テーブル名",
			dbType:   models.DBTypeSQLServer,
			input:    "顧客マスタ",
			expected: "[顧客マスタ]",
		},
		{
			name:     "SQLServer: 日本語カラム名",
			dbType:   models.DBTypeSQLServer,
			input:    "プロセス名",
			expected: "[プロセス名]",
		},
		{
			name:     "SQLServer: 閉じ括弧を含む名前",
			dbType:   models.DBTypeSQLServer,
			input:    "user]name",
			expected: "[user]]name]",
		},
		{
			name:     "SQLServer: スペースを含む名前",
			dbType:   models.DBTypeSQLServer,
			input:    "user name",
			expected: "[user name]",
		},
		{
			name:     "SQLServer: 予約語",
			dbType:   models.DBTypeSQLServer,
			input:    "select",
			expected: "[select]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := quoteIdentifierForDB(tt.dbType, tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestQuoteIdentifierForDB_UnsupportedDBType 未対応のデータベースタイプでエラーを返すことをテストする
func TestQuoteIdentifierForDB_UnsupportedDBType(t *testing.T) {
	_, err := quoteIdentifierForDB("unknown", "users")
	assert.Error(t, err)
}

// TestQuoteIdentifierForDB_RejectsInvalid 制御文字・空文字など不正な識別子を全方言で拒否することをテストする
func TestQuoteIdentifierForDB_RejectsInvalid(t *testing.T) {
	dbTypes := []models.DBType{
		models.DBTypePostgreSQL,
		models.DBTypeMySQL,
		models.DBTypeOracle,
		models.DBTypeSQLServer,
	}

	tests := []struct {
		name  string
		input string
	}{
		{name: "空文字", input: ""},
		{name: "ヌルバイトを含む", input: "users\x00"},
		{name: "改行（制御文字）を含む", input: "user\nname"},
		{name: "タブ（制御文字）を含む", input: "user\tname"},
		{name: "DEL（制御文字）を含む", input: "user\x7fname"},
		{name: "長すぎる識別子", input: strings.Repeat("a", maxExternalIdentifierLength+1)},
	}

	for _, dbType := range dbTypes {
		for _, tt := range tests {
			t.Run(string(dbType)+": "+tt.name, func(t *testing.T) {
				_, err := quoteIdentifierForDB(dbType, tt.input)
				assert.Error(t, err)
			})
		}
	}
}

// TestQuoteIdentifierForDBWithJapaneseEdgeCases 日本語の境界ケースをテストする
func TestQuoteIdentifierForDBWithJapaneseEdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		dbType models.DBType
		input  string
	}{
		{name: "PostgreSQL: ひらがな", dbType: models.DBTypePostgreSQL, input: "てすと"},
		{name: "MySQL: カタカナ", dbType: models.DBTypeMySQL, input: "テスト"},
		{name: "Oracle: 漢字", dbType: models.DBTypeOracle, input: "顧客管理"},
		{name: "SQLServer: 全角数字", dbType: models.DBTypeSQLServer, input: "テーブル１２３"},
		{name: "PostgreSQL: 全角記号", dbType: models.DBTypePostgreSQL, input: "テスト＿テーブル"},
		{name: "MySQL: 混合名", dbType: models.DBTypeMySQL, input: "user_テーブル_123"},
		{name: "Oracle: 長い日本語名", dbType: models.DBTypeOracle, input: "非常に長い日本語のテーブル名前をテストする"},
		{name: "SQLServer: 絵文字を含む", dbType: models.DBTypeSQLServer, input: "テスト😀テーブル"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := quoteIdentifierForDB(tt.dbType, tt.input)
			assert.NoError(t, err)
			assert.NotEmpty(t, result)
			// 入力が結果に含まれていることを確認（クォート文字を除く）
			assert.Contains(t, result, tt.input)
		})
	}
}

// TestGetPlaceholder 各データベースタイプでのプレースホルダーをテストする
func TestGetPlaceholder(t *testing.T) {
	tests := []struct {
		name     string
		dbType   models.DBType
		index    int
		expected string
	}{
		{name: "PostgreSQL: index 1", dbType: models.DBTypePostgreSQL, index: 1, expected: "$1"},
		{name: "PostgreSQL: index 5", dbType: models.DBTypePostgreSQL, index: 5, expected: "$5"},
		{name: "PostgreSQL: index 100", dbType: models.DBTypePostgreSQL, index: 100, expected: "$100"},
		{name: "MySQL: index 1", dbType: models.DBTypeMySQL, index: 1, expected: "?"},
		{name: "MySQL: index 5", dbType: models.DBTypeMySQL, index: 5, expected: "?"},
		{name: "Oracle: index 1", dbType: models.DBTypeOracle, index: 1, expected: ":1"},
		{name: "Oracle: index 5", dbType: models.DBTypeOracle, index: 5, expected: ":5"},
		{name: "SQLServer: index 1", dbType: models.DBTypeSQLServer, index: 1, expected: "@p1"},
		{name: "SQLServer: index 5", dbType: models.DBTypeSQLServer, index: 5, expected: "@p5"},
		{name: "Unknown: MySQL形式にフォールバック", dbType: "unknown", index: 1, expected: "?"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPlaceholder(tt.dbType, tt.index)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestBuildLimitOffset 各データベースタイプでのLIMIT/OFFSET句をテストする
func TestBuildLimitOffset(t *testing.T) {
	tests := []struct {
		name     string
		dbType   models.DBType
		limit    int
		offset   int
		expected string
	}{
		{name: "PostgreSQL: limit 10 offset 0", dbType: models.DBTypePostgreSQL, limit: 10, offset: 0, expected: " LIMIT 10 OFFSET 0"},
		{name: "PostgreSQL: limit 20 offset 40", dbType: models.DBTypePostgreSQL, limit: 20, offset: 40, expected: " LIMIT 20 OFFSET 40"},
		{name: "PostgreSQL: limit 10 offset 20", dbType: models.DBTypePostgreSQL, limit: 10, offset: 20, expected: " LIMIT 10 OFFSET 20"},
		{name: "MySQL: limit 10 offset 0", dbType: models.DBTypeMySQL, limit: 10, offset: 0, expected: " LIMIT 10 OFFSET 0"},
		{name: "MySQL: limit 10 offset 20", dbType: models.DBTypeMySQL, limit: 10, offset: 20, expected: " LIMIT 10 OFFSET 20"},
		{name: "Oracle: limit 10 offset 0", dbType: models.DBTypeOracle, limit: 10, offset: 0, expected: " OFFSET 0 ROWS FETCH NEXT 10 ROWS ONLY"},
		{name: "Oracle: limit 10 offset 20", dbType: models.DBTypeOracle, limit: 10, offset: 20, expected: " OFFSET 20 ROWS FETCH NEXT 10 ROWS ONLY"},
		{name: "SQLServer: limit 10 offset 0", dbType: models.DBTypeSQLServer, limit: 10, offset: 0, expected: " OFFSET 0 ROWS FETCH NEXT 10 ROWS ONLY"},
		{name: "SQLServer: limit 10 offset 20", dbType: models.DBTypeSQLServer, limit: 10, offset: 20, expected: " OFFSET 20 ROWS FETCH NEXT 10 ROWS ONLY"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildLimitOffset(tt.dbType, tt.limit, tt.offset)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestBuildOrderAndLimit ORDER BY句とLIMIT/OFFSET句の組み立てを各データベースタイプでテストする
func TestBuildOrderAndLimit(t *testing.T) {
	tests := []struct {
		name     string
		dbType   models.DBType
		sortSQL  string
		limit    int
		offset   int
		expected string
	}{
		// 並び替え指定あり: 全方言でORDER BYをそのまま出す
		{
			name:     "PostgreSQL: ソートあり",
			dbType:   models.DBTypePostgreSQL,
			sortSQL:  `"id" ASC`,
			limit:    10,
			offset:   20,
			expected: ` ORDER BY "id" ASC LIMIT 10 OFFSET 20`,
		},
		{
			name:     "MySQL: ソートあり",
			dbType:   models.DBTypeMySQL,
			sortSQL:  "`id` DESC",
			limit:    10,
			offset:   20,
			expected: " ORDER BY `id` DESC LIMIT 10 OFFSET 20",
		},
		{
			name:     "Oracle: ソートあり",
			dbType:   models.DBTypeOracle,
			sortSQL:  `"ID" ASC`,
			limit:    10,
			offset:   20,
			expected: ` ORDER BY "ID" ASC OFFSET 20 ROWS FETCH NEXT 10 ROWS ONLY`,
		},
		{
			name:     "SQLServer: ソートあり（フォールバックしない）",
			dbType:   models.DBTypeSQLServer,
			sortSQL:  "[id] ASC",
			limit:    10,
			offset:   20,
			expected: " ORDER BY [id] ASC OFFSET 20 ROWS FETCH NEXT 10 ROWS ONLY",
		},

		// 並び替え指定なし: SQL ServerだけORDER BY (SELECT NULL)を補う
		{
			name:     "PostgreSQL: ソートなしはORDER BYを付けない",
			dbType:   models.DBTypePostgreSQL,
			sortSQL:  "",
			limit:    10,
			offset:   20,
			expected: " LIMIT 10 OFFSET 20",
		},
		{
			name:     "MySQL: ソートなしはORDER BYを付けない",
			dbType:   models.DBTypeMySQL,
			sortSQL:  "",
			limit:    10,
			offset:   20,
			expected: " LIMIT 10 OFFSET 20",
		},
		{
			name:     "Oracle: ソートなしはORDER BYを付けない",
			dbType:   models.DBTypeOracle,
			sortSQL:  "",
			limit:    10,
			offset:   20,
			expected: " OFFSET 20 ROWS FETCH NEXT 10 ROWS ONLY",
		},
		{
			name:     "SQLServer: ソートなしはORDER BY (SELECT NULL)を補う",
			dbType:   models.DBTypeSQLServer,
			sortSQL:  "",
			limit:    10,
			offset:   20,
			expected: " ORDER BY (SELECT NULL) OFFSET 20 ROWS FETCH NEXT 10 ROWS ONLY",
		},
		{
			name:     "SQLServer: ソートなし・offset 0",
			dbType:   models.DBTypeSQLServer,
			sortSQL:  "",
			limit:    25,
			offset:   0,
			expected: " ORDER BY (SELECT NULL) OFFSET 0 ROWS FETCH NEXT 25 ROWS ONLY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildOrderAndLimit(tt.dbType, tt.sortSQL, tt.limit, tt.offset)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestBuildDSN 各データベースタイプでのDSN構築をテストする
func TestBuildDSN(t *testing.T) {
	tests := []struct {
		name           string
		dataSource     *models.DataSource
		password       string
		expectedDriver string
		expectedDSN    string
		expectedError  bool
	}{
		// PostgreSQL
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

		// MySQL
		{
			name: "MySQL: 基本的なDSN",
			dataSource: &models.DataSource{
				DBType:       models.DBTypeMySQL,
				Host:         "localhost",
				Port:         3306,
				Username:     "testuser",
				DatabaseName: "testdb",
			},
			password:       "testpass",
			expectedDriver: "mysql",
			expectedDSN:    "testuser:testpass@tcp(localhost:3306)/testdb?parseTime=true",
			expectedError:  false,
		},
		{
			name: "MySQL: 日本語データベース名",
			dataSource: &models.DataSource{
				DBType:       models.DBTypeMySQL,
				Host:         "localhost",
				Port:         3306,
				Username:     "testuser",
				DatabaseName: "テストDB",
			},
			password:       "testpass",
			expectedDriver: "mysql",
			expectedDSN:    "testuser:testpass@tcp(localhost:3306)/テストDB?parseTime=true",
			expectedError:  false,
		},

		// Oracle
		{
			name: "Oracle: 基本的なDSN",
			dataSource: &models.DataSource{
				DBType:       models.DBTypeOracle,
				Host:         "localhost",
				Port:         1521,
				Username:     "testuser",
				DatabaseName: "ORCL",
			},
			password:       "testpass",
			expectedDriver: "oracle",
			expectedDSN:    "oracle://testuser:testpass@localhost:1521/ORCL",
			expectedError:  false,
		},
		{
			name: "Oracle: 特殊文字を含むパスワード",
			dataSource: &models.DataSource{
				DBType:       models.DBTypeOracle,
				Host:         "localhost",
				Port:         1521,
				Username:     "testuser",
				DatabaseName: "ORCL",
			},
			password:       "test@pass/word",
			expectedDriver: "oracle",
			expectedDSN:    "oracle://testuser:test%40pass%2Fword@localhost:1521/ORCL",
			expectedError:  false,
		},

		// SQL Server
		{
			name: "SQLServer: 基本的なDSN",
			dataSource: &models.DataSource{
				DBType:       models.DBTypeSQLServer,
				Host:         "localhost",
				Port:         1433,
				Username:     "testuser",
				DatabaseName: "testdb",
			},
			password:       "testpass",
			expectedDriver: "sqlserver",
			expectedDSN:    "sqlserver://testuser:testpass@localhost:1433?database=testdb",
			expectedError:  false,
		},
		{
			name: "SQLServer: 特殊文字を含むパスワード",
			dataSource: &models.DataSource{
				DBType:       models.DBTypeSQLServer,
				Host:         "localhost",
				Port:         1433,
				Username:     "testuser",
				DatabaseName: "testdb",
			},
			password:       "test@pass/word",
			expectedDriver: "sqlserver",
			expectedDSN:    "sqlserver://testuser:test%40pass%2Fword@localhost:1433?database=testdb",
			expectedError:  false,
		},

		// 不明なデータベースタイプ
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
