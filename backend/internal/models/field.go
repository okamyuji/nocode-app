package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/uptrace/bun"
)

// FieldType フィールドの種類を表す型
type FieldType string

// フィールド種類の定数
const (
	FieldTypeText        FieldType = "text"
	FieldTypeTextArea    FieldType = "textarea"
	FieldTypeNumber      FieldType = "number"
	FieldTypeDate        FieldType = "date"
	FieldTypeDateTime    FieldType = "datetime"
	FieldTypeSelect      FieldType = "select"
	FieldTypeMultiSelect FieldType = "multiselect"
	FieldTypeCheckbox    FieldType = "checkbox"
	FieldTypeRadio       FieldType = "radio"
	FieldTypeLink        FieldType = "link"
	FieldTypeAttachment  FieldType = "attachment"
)

// MySQLカラム型の定数
const (
	mysqlVarchar255 = "VARCHAR(255)"
)

// FieldOptions フィールド固有のオプションをJSONとして保持する型
type FieldOptions map[string]interface{}

// Value FieldOptionsのdriver.Valuer実装
func (fo FieldOptions) Value() (driver.Value, error) {
	if fo == nil {
		return nil, nil
	}
	return json.Marshal(fo)
}

// Scan FieldOptionsのsql.Scanner実装
func (fo *FieldOptions) Scan(value interface{}) error {
	if value == nil {
		*fo = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("FieldOptionsのスキャンに失敗しました")
	}
	return json.Unmarshal(bytes, fo)
}

// AppField アプリ内のフィールド定義を表す構造体
type AppField struct {
	bun.BaseModel `bun:"table:app_fields,alias:af"`

	ID           uint64       `bun:"id,pk,autoincrement" json:"id"`
	AppID        uint64       `bun:"app_id,notnull" json:"app_id"`
	FieldCode    string       `bun:"field_code,notnull" json:"field_code"`
	FieldName    string       `bun:"field_name,notnull" json:"field_name"`
	FieldType    string       `bun:"field_type,notnull" json:"field_type"`
	Options      FieldOptions `bun:"options,type:json" json:"options,omitempty"`
	Required     bool         `bun:"required,notnull,default:false" json:"required"`
	DisplayOrder int          `bun:"display_order,notnull,default:0" json:"display_order"`
	CreatedAt    time.Time    `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt    time.Time    `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

// CreateFieldRequest フィールド作成リクエストの構造体
type CreateFieldRequest struct {
	FieldCode    string       `json:"field_code" validate:"required,min=1,max=64,alphanum"`
	FieldName    string       `json:"field_name" validate:"required,min=1,max=100"`
	FieldType    string       `json:"field_type" validate:"required,oneof=text textarea number date datetime select multiselect checkbox radio link attachment"`
	Options      FieldOptions `json:"options"`
	Required     bool         `json:"required"`
	DisplayOrder int          `json:"display_order"`
}

// UpdateFieldRequest フィールド更新リクエストの構造体
type UpdateFieldRequest struct {
	FieldName    string       `json:"field_name" validate:"omitempty,min=1,max=100"`
	Options      FieldOptions `json:"options"`
	Required     *bool        `json:"required"`
	DisplayOrder *int         `json:"display_order"`
}

// UpdateFieldOrderRequest フィールド順序更新リクエストの構造体
type UpdateFieldOrderRequest struct {
	Fields []FieldOrderItem `json:"fields" validate:"required,min=1,dive"`
}

// FieldOrderItem 単一のフィールド順序アイテムを表す構造体
type FieldOrderItem struct {
	ID           uint64 `json:"id" validate:"required"`
	DisplayOrder int    `json:"display_order"`
}

// FieldResponse フィールドデータのレスポンス構造体
type FieldResponse struct {
	ID           uint64       `json:"id"`
	AppID        uint64       `json:"app_id"`
	FieldCode    string       `json:"field_code"`
	FieldName    string       `json:"field_name"`
	FieldType    string       `json:"field_type"`
	Options      FieldOptions `json:"options,omitempty"`
	Required     bool         `json:"required"`
	DisplayOrder int          `json:"display_order"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

// ToResponse AppFieldをFieldResponseに変換する
func (f *AppField) ToResponse() *FieldResponse {
	return &FieldResponse{
		ID:           f.ID,
		AppID:        f.AppID,
		FieldCode:    f.FieldCode,
		FieldName:    f.FieldName,
		FieldType:    f.FieldType,
		Options:      f.Options,
		Required:     f.Required,
		DisplayOrder: f.DisplayOrder,
		CreatedAt:    f.CreatedAt,
		UpdatedAt:    f.UpdatedAt,
	}
}

// GetMySQLColumnType このフィールドのMySQLカラム型を返す
func (f *AppField) GetMySQLColumnType() string {
	switch FieldType(f.FieldType) {
	case FieldTypeText:
		return mysqlVarchar255
	case FieldTypeTextArea:
		return "TEXT"
	case FieldTypeNumber:
		return "DECIMAL(18,4)"
	case FieldTypeDate:
		return "DATE"
	case FieldTypeDateTime:
		return "DATETIME"
	case FieldTypeSelect:
		return mysqlVarchar255
	case FieldTypeMultiSelect:
		return "JSON"
	case FieldTypeCheckbox:
		return "BOOLEAN"
	case FieldTypeRadio:
		return mysqlVarchar255
	case FieldTypeLink:
		return "VARCHAR(500)"
	case FieldTypeAttachment:
		return "JSON"
	default:
		return mysqlVarchar255
	}
}
