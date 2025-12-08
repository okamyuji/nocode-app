package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/uptrace/bun"
)

// DashboardStats ダッシュボードに表示する統計情報を表す構造体
type DashboardStats struct {
	AppCount      int64 `json:"app_count"`
	TotalRecords  int64 `json:"total_records"`
	UserCount     int64 `json:"user_count"`
	TodaysUpdates int64 `json:"todays_updates"`
}

// DashboardStatsResponse ダッシュボード統計のAPIレスポンス構造体
type DashboardStatsResponse struct {
	Stats DashboardStats `json:"stats"`
}

// WidgetViewType ウィジェットの表示形式を表す型
type WidgetViewType string

// ウィジェット表示形式の定数
const (
	WidgetViewTypeTable WidgetViewType = "table"
	WidgetViewTypeList  WidgetViewType = "list"
	WidgetViewTypeChart WidgetViewType = "chart"
)

// ValidWidgetViewTypes 有効なウィジェット表示形式のリスト
var ValidWidgetViewTypes = []WidgetViewType{
	WidgetViewTypeTable,
	WidgetViewTypeList,
	WidgetViewTypeChart,
}

// IsValid ウィジェット表示形式が有効かどうかを確認
func (v WidgetViewType) IsValid() bool {
	for _, validType := range ValidWidgetViewTypes {
		if v == validType {
			return true
		}
	}
	return false
}

// WidgetSize ウィジェットのサイズを表す型
type WidgetSize string

// ウィジェットサイズの定数
const (
	WidgetSizeSmall  WidgetSize = "small"
	WidgetSizeMedium WidgetSize = "medium"
	WidgetSizeLarge  WidgetSize = "large"
)

// ValidWidgetSizes 有効なウィジェットサイズのリスト
var ValidWidgetSizes = []WidgetSize{
	WidgetSizeSmall,
	WidgetSizeMedium,
	WidgetSizeLarge,
}

// IsValid ウィジェットサイズが有効かどうかを確認
func (s WidgetSize) IsValid() bool {
	for _, validSize := range ValidWidgetSizes {
		if s == validSize {
			return true
		}
	}
	return false
}

// WidgetConfig ウィジェット固有の設定をJSONとして保持する型
type WidgetConfig map[string]interface{}

// Value WidgetConfigのdriver.Valuer実装
func (wc WidgetConfig) Value() (driver.Value, error) {
	if wc == nil {
		return "{}", nil
	}
	bytes, err := json.Marshal(wc)
	if err != nil {
		return nil, err
	}
	return string(bytes), nil
}

// Scan WidgetConfigのsql.Scanner実装
func (wc *WidgetConfig) Scan(value interface{}) error {
	if value == nil {
		*wc = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("WidgetConfigのスキャンに失敗しました")
	}
	return json.Unmarshal(bytes, wc)
}

// DashboardWidget ダッシュボードウィジェットを表す構造体
type DashboardWidget struct {
	bun.BaseModel `bun:"table:dashboard_widgets,alias:dw"`

	ID           uint64         `bun:"id,pk,autoincrement" json:"id"`
	UserID       uint64         `bun:"user_id,notnull" json:"user_id"`
	AppID        uint64         `bun:"app_id,notnull" json:"app_id"`
	DisplayOrder int            `bun:"display_order,notnull,default:0" json:"display_order"`
	ViewType     WidgetViewType `bun:"view_type,notnull,default:'table'" json:"view_type"`
	IsVisible    bool           `bun:"is_visible,notnull,default:true" json:"is_visible"`
	WidgetSize   WidgetSize     `bun:"widget_size,notnull,default:'medium'" json:"widget_size"`
	Config       WidgetConfig   `bun:"config,type:json" json:"config,omitempty"`
	CreatedAt    time.Time      `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt    time.Time      `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`

	// リレーション
	App *App `bun:"rel:belongs-to,join:app_id=id" json:"app,omitempty"`
}

// CreateDashboardWidgetRequest ダッシュボードウィジェット作成リクエストの構造体
type CreateDashboardWidgetRequest struct {
	AppID        uint64       `json:"app_id" validate:"required"`
	DisplayOrder *int         `json:"display_order"`
	ViewType     string       `json:"view_type" validate:"omitempty,oneof=table list chart"`
	IsVisible    *bool        `json:"is_visible"`
	WidgetSize   string       `json:"widget_size" validate:"omitempty,oneof=small medium large"`
	Config       WidgetConfig `json:"config"`
}

// UpdateDashboardWidgetRequest ダッシュボードウィジェット更新リクエストの構造体
type UpdateDashboardWidgetRequest struct {
	DisplayOrder *int         `json:"display_order"`
	ViewType     string       `json:"view_type" validate:"omitempty,oneof=table list chart"`
	IsVisible    *bool        `json:"is_visible"`
	WidgetSize   string       `json:"widget_size" validate:"omitempty,oneof=small medium large"`
	Config       WidgetConfig `json:"config"`
}

// ReorderWidgetsRequest ウィジェット並び替えリクエストの構造体
type ReorderWidgetsRequest struct {
	WidgetIDs []uint64 `json:"widget_ids" validate:"required,min=1"`
}

// DashboardWidgetResponse ダッシュボードウィジェットのレスポンス構造体
type DashboardWidgetResponse struct {
	ID           uint64       `json:"id"`
	UserID       uint64       `json:"user_id"`
	AppID        uint64       `json:"app_id"`
	DisplayOrder int          `json:"display_order"`
	ViewType     string       `json:"view_type"`
	IsVisible    bool         `json:"is_visible"`
	WidgetSize   string       `json:"widget_size"`
	Config       WidgetConfig `json:"config,omitempty"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
	App          *AppResponse `json:"app,omitempty"`
}

// ToResponse DashboardWidgetをDashboardWidgetResponseに変換する
func (w *DashboardWidget) ToResponse() *DashboardWidgetResponse {
	resp := &DashboardWidgetResponse{
		ID:           w.ID,
		UserID:       w.UserID,
		AppID:        w.AppID,
		DisplayOrder: w.DisplayOrder,
		ViewType:     string(w.ViewType),
		IsVisible:    w.IsVisible,
		WidgetSize:   string(w.WidgetSize),
		Config:       w.Config,
		CreatedAt:    w.CreatedAt,
		UpdatedAt:    w.UpdatedAt,
	}

	if w.App != nil {
		resp.App = w.App.ToResponse()
	}

	return resp
}

// DashboardWidgetListResponse ダッシュボードウィジェット一覧のレスポンス構造体
type DashboardWidgetListResponse struct {
	Widgets []DashboardWidgetResponse `json:"widgets"`
}
