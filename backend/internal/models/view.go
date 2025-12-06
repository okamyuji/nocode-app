package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/uptrace/bun"
)

// ViewType ビューの種類を表す型
type ViewType string

// ビュー種類の定数
const (
	ViewTypeTable    ViewType = "table"
	ViewTypeList     ViewType = "list"
	ViewTypeCalendar ViewType = "calendar"
	ViewTypeChart    ViewType = "chart"
)

// ViewConfig ビュー固有の設定をJSONとして保持する型
type ViewConfig map[string]interface{}

// Value ViewConfigのdriver.Valuer実装
func (vc ViewConfig) Value() (driver.Value, error) {
	if vc == nil {
		return "{}", nil
	}
	bytes, err := json.Marshal(vc)
	if err != nil {
		return nil, err
	}
	return string(bytes), nil
}

// Scan ViewConfigのsql.Scanner実装
func (vc *ViewConfig) Scan(value interface{}) error {
	if value == nil {
		*vc = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("ViewConfigのスキャンに失敗しました")
	}
	return json.Unmarshal(bytes, vc)
}

// AppView アプリのビュー設定を表す構造体
type AppView struct {
	bun.BaseModel `bun:"table:app_views,alias:av"`

	ID        uint64     `bun:"id,pk,autoincrement" json:"id"`
	AppID     uint64     `bun:"app_id,notnull" json:"app_id"`
	Name      string     `bun:"name,notnull" json:"name"`
	ViewType  string     `bun:"view_type,notnull,default:'table'" json:"view_type"`
	Config    ViewConfig `bun:"config,type:json" json:"config,omitempty"`
	IsDefault bool       `bun:"is_default,notnull,default:false" json:"is_default"`
	CreatedAt time.Time  `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time  `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

// CreateViewRequest ビュー作成リクエストの構造体
type CreateViewRequest struct {
	Name      string     `json:"name" validate:"required,min=1,max=100"`
	ViewType  string     `json:"view_type" validate:"required,oneof=table list calendar chart"`
	Config    ViewConfig `json:"config"`
	IsDefault bool       `json:"is_default"`
}

// UpdateViewRequest ビュー更新リクエストの構造体
type UpdateViewRequest struct {
	Name      string     `json:"name" validate:"omitempty,min=1,max=100"`
	Config    ViewConfig `json:"config"`
	IsDefault *bool      `json:"is_default"`
}

// ViewResponse ビューデータのレスポンス構造体
type ViewResponse struct {
	ID        uint64     `json:"id"`
	AppID     uint64     `json:"app_id"`
	Name      string     `json:"name"`
	ViewType  string     `json:"view_type"`
	Config    ViewConfig `json:"config,omitempty"`
	IsDefault bool       `json:"is_default"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// ToResponse AppViewをViewResponseに変換する
func (v *AppView) ToResponse() *ViewResponse {
	return &ViewResponse{
		ID:        v.ID,
		AppID:     v.AppID,
		Name:      v.Name,
		ViewType:  v.ViewType,
		Config:    v.Config,
		IsDefault: v.IsDefault,
		CreatedAt: v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
	}
}

// ChartConfig 保存されたチャート設定を表す構造体
type ChartConfig struct {
	bun.BaseModel `bun:"table:chart_configs,alias:cc"`

	ID        uint64     `bun:"id,pk,autoincrement" json:"id"`
	AppID     uint64     `bun:"app_id,notnull" json:"app_id"`
	Name      string     `bun:"name,notnull" json:"name"`
	ChartType string     `bun:"chart_type,notnull" json:"chart_type"`
	Config    ViewConfig `bun:"config,type:json,notnull" json:"config"`
	CreatedBy uint64     `bun:"created_by,notnull" json:"created_by"`
	CreatedAt time.Time  `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time  `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}
