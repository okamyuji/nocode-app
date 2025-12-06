package models

// Pagination ページネーション情報を表す構造体
type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// NewPagination 新しいPaginationインスタンスを作成する
func NewPagination(page, limit int, total int64) *Pagination {
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}
	return &Pagination{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}
}

// RecordData 動的なレコードデータを表す型
type RecordData map[string]interface{}

// CreateRecordRequest レコード作成リクエストの構造体
type CreateRecordRequest struct {
	Data RecordData `json:"data" validate:"required"`
}

// UpdateRecordRequest レコード更新リクエストの構造体
type UpdateRecordRequest struct {
	Data RecordData `json:"data" validate:"required"`
}

// BulkCreateRecordRequest レコード一括作成リクエストの構造体
type BulkCreateRecordRequest struct {
	Records []RecordData `json:"records" validate:"required,min=1"`
}

// BulkDeleteRecordRequest レコード一括削除リクエストの構造体
type BulkDeleteRecordRequest struct {
	IDs []uint64 `json:"ids" validate:"required,min=1"`
}

// RecordResponse レコードデータのレスポンス構造体
type RecordResponse struct {
	ID        uint64     `json:"id"`
	Data      RecordData `json:"data"`
	CreatedBy uint64     `json:"created_by"`
	CreatedAt string     `json:"created_at"`
	UpdatedAt string     `json:"updated_at"`
}

// RecordListResponse レコード一覧のレスポンス構造体
type RecordListResponse struct {
	Records    []RecordResponse `json:"records"`
	Pagination *Pagination      `json:"pagination"`
}

// ChartDataRequest チャートデータリクエストの構造体
type ChartDataRequest struct {
	ChartType string       `json:"chart_type" validate:"required,oneof=bar horizontal_bar line pie doughnut scatter area"`
	XAxis     ChartAxis    `json:"x_axis" validate:"required"`
	YAxis     ChartAxis    `json:"y_axis" validate:"required"`
	Filters   []FilterItem `json:"filters"`
	GroupBy   string       `json:"group_by"`
}

// ChartAxis チャートの軸設定を表す構造体
type ChartAxis struct {
	Field       string `json:"field"` // X軸および非count集計のY軸で必須
	Label       string `json:"label"`
	Aggregation string `json:"aggregation"` // count, sum, avg, min, max
}

// FilterItem フィルター条件を表す構造体
type FilterItem struct {
	Field    string `json:"field" validate:"required"`
	Operator string `json:"operator" validate:"required,oneof=eq ne gt gte lt lte like in"`
	Value    string `json:"value"`
}

// ChartDataResponse チャートデータのレスポンス構造体
type ChartDataResponse struct {
	Labels   []string       `json:"labels"`
	Datasets []ChartDataset `json:"datasets"`
}

// ChartDataset チャート内の単一データセットを表す構造体
type ChartDataset struct {
	Label string    `json:"label"`
	Data  []float64 `json:"data"`
}

// SaveChartConfigRequest チャート設定保存リクエストの構造体
type SaveChartConfigRequest struct {
	Name      string           `json:"name" validate:"required,min=1,max=100"`
	ChartType string           `json:"chart_type" validate:"required"`
	Config    ChartDataRequest `json:"config" validate:"required"`
}

// ErrorResponse エラーレスポンスを表す構造体
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code,omitempty"`
}

// SuccessResponse 成功レスポンスを表す構造体
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
