package models

import (
	"time"

	"github.com/uptrace/bun"
)

// App ユーザーが作成したアプリケーション（データベーステーブル）を表す構造体
type App struct {
	bun.BaseModel `bun:"table:apps,alias:a"`

	ID              uint64      `bun:"id,pk,autoincrement" json:"id"`
	Name            string      `bun:"name,notnull" json:"name"`
	Description     string      `bun:"description" json:"description"`
	TableName       string      `bun:"table_name,notnull,unique" json:"table_name"`
	Icon            string      `bun:"icon,notnull,default:'default'" json:"icon"`
	IsExternal      bool        `bun:"is_external,notnull,default:false" json:"is_external"`
	DataSourceID    *uint64     `bun:"data_source_id" json:"data_source_id,omitempty"`
	SourceTableName *string     `bun:"source_table_name" json:"source_table_name,omitempty"`
	CreatedBy       uint64      `bun:"created_by,notnull" json:"created_by"`
	CreatedAt       time.Time   `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt       time.Time   `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
	Fields          []AppField  `bun:"rel:has-many,join:id=app_id" json:"fields,omitempty"`
	Views           []AppView   `bun:"rel:has-many,join:id=app_id" json:"views,omitempty"`
	Creator         *User       `bun:"rel:belongs-to,join:created_by=id" json:"creator,omitempty"`
	DataSource      *DataSource `bun:"rel:belongs-to,join:data_source_id=id" json:"data_source,omitempty"`
}

// CreateAppRequest アプリ作成リクエストの構造体
type CreateAppRequest struct {
	Name        string               `json:"name" validate:"required,min=1,max=100"`
	Description string               `json:"description"`
	Icon        string               `json:"icon"`
	Fields      []CreateFieldRequest `json:"fields" validate:"required,min=1,dive"`
}

// CreateExternalAppRequest 外部データソースからのアプリ作成リクエストの構造体
type CreateExternalAppRequest struct {
	Name            string                       `json:"name" validate:"required,min=1,max=100"`
	Description     string                       `json:"description"`
	Icon            string                       `json:"icon"`
	DataSourceID    uint64                       `json:"data_source_id" validate:"required"`
	SourceTableName string                       `json:"source_table_name" validate:"required,min=1,max=100"`
	Fields          []CreateExternalFieldRequest `json:"fields" validate:"required,min=1,dive"`
}

// CreateExternalFieldRequest 外部データソース用のフィールド作成リクエスト
// 外部データソースは読み取り専用のため、Requiredフィールドは不要
type CreateExternalFieldRequest struct {
	SourceColumnName string                 `json:"source_column_name" validate:"required,min=1,max=100"`
	FieldCode        string                 `json:"field_code" validate:"required,min=1,max=64,fieldcode"`
	FieldName        string                 `json:"field_name" validate:"required,min=1,max=100"`
	FieldType        string                 `json:"field_type" validate:"required,oneof=text textarea number date datetime select multiselect checkbox radio link attachment"`
	Options          map[string]interface{} `json:"options"`
	DisplayOrder     int                    `json:"display_order"`
}

// UpdateAppRequest アプリ更新リクエストの構造体
type UpdateAppRequest struct {
	Name        string `json:"name" validate:"omitempty,min=1,max=100"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

// AppResponse アプリデータのレスポンス構造体
type AppResponse struct {
	ID              uint64          `json:"id"`
	Name            string          `json:"name"`
	Description     string          `json:"description"`
	TableName       string          `json:"table_name"`
	Icon            string          `json:"icon"`
	IsExternal      bool            `json:"is_external"`
	DataSourceID    *uint64         `json:"data_source_id,omitempty"`
	SourceTableName *string         `json:"source_table_name,omitempty"`
	CreatedBy       uint64          `json:"created_by"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	Fields          []FieldResponse `json:"fields,omitempty"`
	FieldCount      int             `json:"field_count"`
}

// ToResponse AppをAppResponseに変換する
func (a *App) ToResponse() *AppResponse {
	resp := &AppResponse{
		ID:              a.ID,
		Name:            a.Name,
		Description:     a.Description,
		TableName:       a.TableName,
		Icon:            a.Icon,
		IsExternal:      a.IsExternal,
		DataSourceID:    a.DataSourceID,
		SourceTableName: a.SourceTableName,
		CreatedBy:       a.CreatedBy,
		CreatedAt:       a.CreatedAt,
		UpdatedAt:       a.UpdatedAt,
		FieldCount:      len(a.Fields),
	}

	if len(a.Fields) > 0 {
		resp.Fields = make([]FieldResponse, len(a.Fields))
		for i := range a.Fields {
			resp.Fields[i] = *a.Fields[i].ToResponse()
		}
	}

	return resp
}

// AppListResponse アプリ一覧のレスポンス構造体
type AppListResponse struct {
	Apps       []AppResponse `json:"apps"`
	Pagination *Pagination   `json:"pagination"`
}
