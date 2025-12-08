package models

import (
	"time"

	"github.com/uptrace/bun"
)

// DBType データベースタイプの定数
type DBType string

const (
	DBTypePostgreSQL DBType = "postgresql"
	DBTypeMySQL      DBType = "mysql"
	DBTypeOracle     DBType = "oracle"
	DBTypeSQLServer  DBType = "sqlserver"
)

// ValidDBTypes 有効なデータベースタイプ一覧
var ValidDBTypes = []DBType{
	DBTypePostgreSQL,
	DBTypeMySQL,
	DBTypeOracle,
	DBTypeSQLServer,
}

// IsValidDBType データベースタイプが有効かどうかを検証する
func IsValidDBType(dbType string) bool {
	for _, t := range ValidDBTypes {
		if string(t) == dbType {
			return true
		}
	}
	return false
}

// DataSource 外部データベース接続情報を表す構造体
type DataSource struct {
	bun.BaseModel `bun:"table:data_sources,alias:ds"`

	ID                uint64    `bun:"id,pk,autoincrement" json:"id"`
	Name              string    `bun:"name,notnull,unique" json:"name"`
	DBType            DBType    `bun:"db_type,notnull" json:"db_type"`
	Host              string    `bun:"host,notnull" json:"host"`
	Port              int       `bun:"port,notnull" json:"port"`
	DatabaseName      string    `bun:"database_name,notnull" json:"database_name"`
	Username          string    `bun:"username,notnull" json:"username"`
	EncryptedPassword string    `bun:"encrypted_password,notnull" json:"-"` // JSONには出力しない
	CreatedBy         uint64    `bun:"created_by,notnull" json:"created_by"`
	CreatedAt         time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt         time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
	Creator           *User     `bun:"rel:belongs-to,join:created_by=id" json:"creator,omitempty"`
}

// CreateDataSourceRequest データソース作成リクエストの構造体
type CreateDataSourceRequest struct {
	Name         string `json:"name" validate:"required,min=1,max=100"`
	DBType       string `json:"db_type" validate:"required,oneof=postgresql mysql oracle sqlserver"`
	Host         string `json:"host" validate:"required,min=1,max=255"`
	Port         int    `json:"port" validate:"required,min=1,max=65535"`
	DatabaseName string `json:"database_name" validate:"required,min=1,max=100"`
	Username     string `json:"username" validate:"required,min=1,max=100"`
	Password     string `json:"password" validate:"required,min=1"`
}

// UpdateDataSourceRequest データソース更新リクエストの構造体
type UpdateDataSourceRequest struct {
	Name         string `json:"name" validate:"omitempty,min=1,max=100"`
	Host         string `json:"host" validate:"omitempty,min=1,max=255"`
	Port         int    `json:"port" validate:"omitempty,min=1,max=65535"`
	DatabaseName string `json:"database_name" validate:"omitempty,min=1,max=100"`
	Username     string `json:"username" validate:"omitempty,min=1,max=100"`
	Password     string `json:"password" validate:"omitempty,min=1"` // 空の場合は更新しない
}

// TestConnectionRequest テスト接続リクエストの構造体
type TestConnectionRequest struct {
	DBType       string `json:"db_type" validate:"required,oneof=postgresql mysql oracle sqlserver"`
	Host         string `json:"host" validate:"required,min=1,max=255"`
	Port         int    `json:"port" validate:"required,min=1,max=65535"`
	DatabaseName string `json:"database_name" validate:"required,min=1,max=100"`
	Username     string `json:"username" validate:"required,min=1,max=100"`
	Password     string `json:"password" validate:"required,min=1"`
}

// DataSourceResponse データソースのレスポンス構造体（パスワードは含まない）
type DataSourceResponse struct {
	ID           uint64    `json:"id"`
	Name         string    `json:"name"`
	DBType       DBType    `json:"db_type"`
	Host         string    `json:"host"`
	Port         int       `json:"port"`
	DatabaseName string    `json:"database_name"`
	Username     string    `json:"username"`
	CreatedBy    uint64    `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ToResponse DataSourceをDataSourceResponseに変換する
func (ds *DataSource) ToResponse() *DataSourceResponse {
	return &DataSourceResponse{
		ID:           ds.ID,
		Name:         ds.Name,
		DBType:       ds.DBType,
		Host:         ds.Host,
		Port:         ds.Port,
		DatabaseName: ds.DatabaseName,
		Username:     ds.Username,
		CreatedBy:    ds.CreatedBy,
		CreatedAt:    ds.CreatedAt,
		UpdatedAt:    ds.UpdatedAt,
	}
}

// DataSourceListResponse データソース一覧のレスポンス構造体
type DataSourceListResponse struct {
	DataSources []DataSourceResponse `json:"data_sources"`
	Pagination  *Pagination          `json:"pagination"`
}

// TableType テーブルタイプの定数
type TableType string

const (
	TableTypeTable TableType = "TABLE"
	TableTypeView  TableType = "VIEW"
)

// TableInfo テーブル情報の構造体
type TableInfo struct {
	Name   string    `json:"name"`
	Schema string    `json:"schema,omitempty"`
	Type   TableType `json:"type"`
}

// TableListResponse テーブル一覧のレスポンス構造体
type TableListResponse struct {
	Tables []TableInfo `json:"tables"`
}

// ColumnInfo カラム情報の構造体
type ColumnInfo struct {
	Name         string `json:"name"`
	DataType     string `json:"data_type"`
	IsNullable   bool   `json:"is_nullable"`
	IsPrimaryKey bool   `json:"is_primary_key"`
	DefaultValue string `json:"default_value,omitempty"`
}

// ColumnListResponse カラム一覧のレスポンス構造体
type ColumnListResponse struct {
	Columns []ColumnInfo `json:"columns"`
}

// TestConnectionResponse テスト接続のレスポンス構造体
type TestConnectionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// GetDefaultPort データベースタイプに応じたデフォルトポートを返す
func GetDefaultPort(dbType DBType) int {
	switch dbType {
	case DBTypePostgreSQL:
		return 5432
	case DBTypeMySQL:
		return 3306
	case DBTypeOracle:
		return 1521
	case DBTypeSQLServer:
		return 1433
	default:
		return 0
	}
}
