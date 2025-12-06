package models

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
