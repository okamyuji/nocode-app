package repositories

import (
	"net/url"
	"strings"
	"testing"

	"nocode-app/backend/internal/models"

	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
// TestOracleObjectName データディクショナリ検索用の名前がquoteIdentifierForDBのOracle分岐と
// 同じ大文字化規則になることをテストする
func TestOracleObjectName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "小文字は大文字化される", input: "employees", expected: "EMPLOYEES"},
		{name: "大文字はそのまま", input: "EMPLOYEES", expected: "EMPLOYEES"},
		{name: "混在は大文字化される", input: "Test_Table", expected: "TEST_TABLE"},
		{name: "数字と記号を含む名前", input: "tbl_2024$x", expected: "TBL_2024$X"},
		{name: "空文字はそのまま", input: "", expected: ""},
		{name: "日本語は変化しない", input: "社員", expected: "社員"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, oracleObjectName(tt.input))

			if tt.input != "" {
				quoted, err := quoteIdentifierForDB(models.DBTypeOracle, tt.input)
				assert.NoError(t, err)
				assert.Equal(t, `"`+tt.expected+`"`, quoted,
					"メタデータ検索の名前とクォート済み識別子の中身は一致する必要がある")
			}
		})
	}
}

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
			expectedDSN:    "host='localhost' port=5432 user='testuser' password='testpass' dbname='testdb' sslmode=disable",
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
			expectedDSN:    "host='localhost' port=5432 user='testuser' password='test\\'pass\\\\word' dbname='testdb' sslmode=disable",
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
			expectedDSN:    "host='localhost' port=5432 user='testuser' password='testpass' dbname='テストDB' sslmode=disable",
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
			expectedDSN:    "testuser:testpass@tcp(localhost:3306)/%E3%83%86%E3%82%B9%E3%83%88DB?parseTime=true",
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

		{
			name: "PostgreSQL: 空白とクォートを含むパスワードは引用符で囲まれる",
			dataSource: &models.DataSource{
				DBType:       models.DBTypePostgreSQL,
				Host:         "localhost",
				Port:         5432,
				Username:     "testuser",
				DatabaseName: "testdb",
			},
			password:       "pa ss'wd",
			expectedDriver: "postgres",
			expectedDSN:    `host='localhost' port=5432 user='testuser' password='pa ss\'wd' dbname='testdb' sslmode=disable`,
			expectedError:  false,
		},
		{
			name: "MySQL: DatabaseNameのクエリ文字列はエスケープされる",
			dataSource: &models.DataSource{
				DBType:       models.DBTypeMySQL,
				Host:         "localhost",
				Port:         3306,
				Username:     "testuser",
				DatabaseName: "db?allowAllFiles=true",
			},
			password:       "testpass",
			expectedDriver: "mysql",
			expectedDSN:    "testuser:testpass@tcp(localhost:3306)/db%3FallowAllFiles=true?parseTime=true",
			expectedError:  false,
		},
		{
			name: "Oracle: サービス名のクエリ文字列はエスケープされる",
			dataSource: &models.DataSource{
				DBType:       models.DBTypeOracle,
				Host:         "localhost",
				Port:         1521,
				Username:     "testuser",
				DatabaseName: "FREEPDB1?TRACE FILE=/tmp/x",
			},
			password:       "testpass",
			expectedDriver: "oracle",
			expectedDSN:    "oracle://testuser:testpass@localhost:1521/FREEPDB1%3FTRACE%20FILE=%2Ftmp%2Fx",
			expectedError:  false,
		},
		{
			name: "SQLServer: DatabaseNameの追加パラメータはエスケープされる",
			dataSource: &models.DataSource{
				DBType:       models.DBTypeSQLServer,
				Host:         "localhost",
				Port:         1433,
				Username:     "testuser",
				DatabaseName: "db&log=63",
			},
			password:       "testpass",
			expectedDriver: "sqlserver",
			expectedDSN:    "sqlserver://testuser:testpass@localhost:1433?database=db%26log%3D63",
			expectedError:  false,
		},
		{
			name: "SQLServer: 空白を含むパスワードは+ではなく%20になる",
			dataSource: &models.DataSource{
				DBType:       models.DBTypeSQLServer,
				Host:         "localhost",
				Port:         1433,
				Username:     "testuser",
				DatabaseName: "testdb",
			},
			password:       "pa ss@wd",
			expectedDriver: "sqlserver",
			expectedDSN:    "sqlserver://testuser:pa%20ss%40wd@localhost:1433?database=testdb",
			expectedError:  false,
		},
		{
			name: "SQLServer: IPv6ホストはブラケットで囲まれる",
			dataSource: &models.DataSource{
				DBType:       models.DBTypeSQLServer,
				Host:         "::1",
				Port:         1433,
				Username:     "testuser",
				DatabaseName: "testdb",
			},
			password:       "testpass",
			expectedDriver: "sqlserver",
			expectedDSN:    "sqlserver://testuser:testpass@[::1]:1433?database=testdb",
			expectedError:  false,
		},

		{
			name: "PostgreSQL: DatabaseNameの空白でキーワードを注入できない",
			dataSource: &models.DataSource{
				DBType:       models.DBTypePostgreSQL,
				Host:         "localhost",
				Port:         5432,
				Username:     "testuser",
				DatabaseName: "testdb sslmode=require",
			},
			password:       "testpass",
			expectedDriver: "postgres",
			expectedDSN:    `host='localhost' port=5432 user='testuser' password='testpass' dbname='testdb sslmode=require' sslmode=disable`,
			expectedError:  false,
		},
		{
			name: "PostgreSQL: ホストの空白でキーワードを注入できない",
			dataSource: &models.DataSource{
				DBType:       models.DBTypePostgreSQL,
				Host:         "localhost sslmode=require",
				Port:         5432,
				Username:     "testuser",
				DatabaseName: "testdb",
			},
			password:       "testpass",
			expectedDriver: "postgres",
			expectedDSN:    `host='localhost sslmode=require' port=5432 user='testuser' password='testpass' dbname='testdb' sslmode=disable`,
			expectedError:  false,
		},
		{
			name: "PostgreSQL: ユーザー名の空白でキーワードを注入できない",
			dataSource: &models.DataSource{
				DBType:       models.DBTypePostgreSQL,
				Host:         "localhost",
				Port:         5432,
				Username:     "testuser sslmode=require",
				DatabaseName: "testdb",
			},
			password:       "testpass",
			expectedDriver: "postgres",
			expectedDSN:    `host='localhost' port=5432 user='testuser sslmode=require' password='testpass' dbname='testdb' sslmode=disable`,
			expectedError:  false,
		},

		// ホストに含めてはならない文字
		{
			name: "PostgreSQL: ホストに@が含まれるとエラー",
			dataSource: &models.DataSource{
				DBType: models.DBTypePostgreSQL, Host: "localhost@evil", Port: 5432,
				Username: "testuser", DatabaseName: "testdb",
			},
			password:      "testpass",
			expectedError: true,
		},
		{
			name: "MySQL: ホストに/が含まれるとエラー",
			dataSource: &models.DataSource{
				DBType: models.DBTypeMySQL, Host: "localhost/evil", Port: 3306,
				Username: "testuser", DatabaseName: "testdb",
			},
			password:      "testpass",
			expectedError: true,
		},
		{
			name: "Oracle: ホストに?が含まれるとエラー",
			dataSource: &models.DataSource{
				DBType: models.DBTypeOracle, Host: "localhost?x=1", Port: 1521,
				Username: "testuser", DatabaseName: "ORCL",
			},
			password:      "testpass",
			expectedError: true,
		},
		{
			name: "SQLServer: ホストに@が含まれるとエラー",
			dataSource: &models.DataSource{
				DBType: models.DBTypeSQLServer, Host: "localhost@evil", Port: 1433,
				Username: "testuser", DatabaseName: "testdb",
			},
			password:      "testpass",
			expectedError: true,
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

// TestQuotePostgresValue PostgreSQLキーワード形式DSNの値の引用をテストする
func TestQuotePostgresValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "通常のパスワード", input: "password123", expected: `'password123'`},
		{name: "シングルクォートを含む", input: "pass'word", expected: `'pass\'word'`},
		{name: "バックスラッシュを含む", input: `pass\word`, expected: `'pass\\word'`},
		{name: "両方を含む", input: `pass'\word`, expected: `'pass\'\\word'`},
		{name: "空白を含む", input: "pa ss wd", expected: `'pa ss wd'`},
		{name: "日本語パスワード", input: "パスワード123", expected: `'パスワード123'`},
		{name: "空文字列", input: "", expected: `''`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := quotePostgresValue(tt.input)
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

// TestBuildDSN_HardeningRoundTrip DatabaseNameやパスワードに混入した特殊文字が
// 接続オプションとして解釈されず、値としてそのまま往復することを確認する。
func TestBuildDSN_HardeningRoundTrip(t *testing.T) {
	t.Run("PostgreSQL: 値に含まれるキーワードは値のまま残る", func(t *testing.T) {
		_, dsn, err := buildDSN(&models.DataSource{
			DBType:       models.DBTypePostgreSQL,
			Host:         "localhost",
			Port:         5432,
			Username:     "testuser",
			DatabaseName: "testdb sslmode=require",
		}, "testpass")
		require.NoError(t, err)

		assert.Contains(t, dsn, "dbname='testdb sslmode=require'")
		assert.Equal(t, 1, strings.Count(dsn, "sslmode=disable"), "sslmode=disableは1回だけ現れるべきです")
		assert.True(t, strings.HasSuffix(dsn, "sslmode=disable"), "注入されたキーワードが値の外に出ています")
	})

	t.Run("MySQL: DatabaseNameでallowAllFilesを有効化できない", func(t *testing.T) {
		_, dsn, err := buildDSN(&models.DataSource{
			DBType:       models.DBTypeMySQL,
			Host:         "localhost",
			Port:         3306,
			Username:     "testuser",
			DatabaseName: "db?allowAllFiles=true",
		}, "testpass")
		require.NoError(t, err)

		cfg, parseErr := mysql.ParseDSN(dsn)
		require.NoError(t, parseErr)
		assert.Equal(t, "db?allowAllFiles=true", cfg.DBName)
		assert.False(t, cfg.AllowAllFiles)
		assert.True(t, cfg.ParseTime)
		assert.True(t, cfg.AllowNativePasswords)
	})

	t.Run("Oracle: サービス名にクエリを混入できない", func(t *testing.T) {
		_, dsn, err := buildDSN(&models.DataSource{
			DBType:       models.DBTypeOracle,
			Host:         "localhost",
			Port:         1521,
			Username:     "testuser",
			DatabaseName: "FREEPDB1?TRACE FILE=/tmp/x",
		}, "testpass")
		require.NoError(t, err)

		u, parseErr := url.Parse(dsn)
		require.NoError(t, parseErr)
		assert.Equal(t, "/FREEPDB1%3FTRACE%20FILE=%2Ftmp%2Fx", u.EscapedPath())
		assert.Equal(t, "/FREEPDB1?TRACE FILE=/tmp/x", u.Path)
		assert.Empty(t, u.RawQuery)
		assert.Equal(t, "localhost:1521", u.Host)
	})

	t.Run("Oracle: 空白と@を含むパスワードが往復する", func(t *testing.T) {
		_, dsn, err := buildDSN(&models.DataSource{
			DBType:       models.DBTypeOracle,
			Host:         "::1",
			Port:         1521,
			Username:     "testuser",
			DatabaseName: "FREEPDB1",
		}, "pa ss@wd")
		require.NoError(t, err)

		u, parseErr := url.Parse(dsn)
		require.NoError(t, parseErr)
		assert.Equal(t, "[::1]:1521", u.Host)
		pw, ok := u.User.Password()
		require.True(t, ok)
		assert.Equal(t, "pa ss@wd", pw)
	})

	t.Run("SQLServer: DatabaseNameで別パラメータを追加できない", func(t *testing.T) {
		_, dsn, err := buildDSN(&models.DataSource{
			DBType:       models.DBTypeSQLServer,
			Host:         "localhost",
			Port:         1433,
			Username:     "testuser",
			DatabaseName: "db&log=63",
		}, "testpass")
		require.NoError(t, err)

		u, parseErr := url.Parse(dsn)
		require.NoError(t, parseErr)
		q := u.Query()
		assert.Equal(t, "db&log=63", q.Get("database"))
		assert.NotContains(t, q, "log")
		assert.Len(t, q, 1)
	})

	t.Run("SQLServer: 空白と@を含むパスワードが往復する", func(t *testing.T) {
		_, dsn, err := buildDSN(&models.DataSource{
			DBType:       models.DBTypeSQLServer,
			Host:         "::1",
			Port:         1433,
			Username:     "testuser",
			DatabaseName: "testdb",
		}, "pa ss@wd")
		require.NoError(t, err)

		u, parseErr := url.Parse(dsn)
		require.NoError(t, parseErr)
		assert.Equal(t, "[::1]:1433", u.Host)
		pw, ok := u.User.Password()
		require.True(t, ok)
		assert.Equal(t, "pa ss@wd", pw)
	})
}

// TestParseRecordID スキャン値からレコードIDへの変換をテストする。
func TestParseRecordID(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected uint64
	}{
		{name: "int64", input: int64(42), expected: 42},
		{name: "uint64", input: uint64(43), expected: 43},
		{name: "float64", input: float64(44), expected: 44},
		{name: "int", input: 45, expected: 45},
		{name: "int32", input: int32(46), expected: 46},
		{name: "string (go-oraのNUMBER)", input: "47", expected: 47},
		{name: "string 前後空白", input: " 48 ", expected: 48},
		{name: "string 小数点付き", input: "49.00", expected: 49},
		{name: "[]byte", input: []byte("50"), expected: 50},
		{name: "string 数値でない", input: "abc", expected: 0},
		{name: "string 空", input: "", expected: 0},
		{name: "[]byte 数値でない", input: []byte("x"), expected: 0},
		{name: "string 負数", input: "-1", expected: 0},
		{name: "nil", input: nil, expected: 0},
		{name: "bool", input: true, expected: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, parseRecordID(tt.input))
		})
	}
}
