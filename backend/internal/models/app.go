package models

import (
	"time"

	"github.com/uptrace/bun"
)

// App ユーザーが作成したアプリケーション（データベーステーブル）を表す構造体
type App struct {
	bun.BaseModel `bun:"table:apps,alias:a"`

	ID          uint64     `bun:"id,pk,autoincrement" json:"id"`
	Name        string     `bun:"name,notnull" json:"name"`
	Description string     `bun:"description" json:"description"`
	TableName   string     `bun:"table_name,notnull,unique" json:"table_name"`
	Icon        string     `bun:"icon,notnull,default:'default'" json:"icon"`
	CreatedBy   uint64     `bun:"created_by,notnull" json:"created_by"`
	CreatedAt   time.Time  `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt   time.Time  `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
	Fields      []AppField `bun:"rel:has-many,join:id=app_id" json:"fields,omitempty"`
	Views       []AppView  `bun:"rel:has-many,join:id=app_id" json:"views,omitempty"`
	Creator     *User      `bun:"rel:belongs-to,join:created_by=id" json:"creator,omitempty"`
}

// CreateAppRequest アプリ作成リクエストの構造体
type CreateAppRequest struct {
	Name        string               `json:"name" validate:"required,min=1,max=100"`
	Description string               `json:"description"`
	Icon        string               `json:"icon"`
	Fields      []CreateFieldRequest `json:"fields" validate:"required,min=1,dive"`
}

// UpdateAppRequest アプリ更新リクエストの構造体
type UpdateAppRequest struct {
	Name        string `json:"name" validate:"omitempty,min=1,max=100"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

// AppResponse アプリデータのレスポンス構造体
type AppResponse struct {
	ID          uint64          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	TableName   string          `json:"table_name"`
	Icon        string          `json:"icon"`
	CreatedBy   uint64          `json:"created_by"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	Fields      []FieldResponse `json:"fields,omitempty"`
	FieldCount  int             `json:"field_count"`
}

// ToResponse AppをAppResponseに変換する
func (a *App) ToResponse() *AppResponse {
	resp := &AppResponse{
		ID:          a.ID,
		Name:        a.Name,
		Description: a.Description,
		TableName:   a.TableName,
		Icon:        a.Icon,
		CreatedBy:   a.CreatedBy,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
		FieldCount:  len(a.Fields),
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
