package handlers

import (
	"net/http"

	"nocode-app/backend/internal/services"
	"nocode-app/backend/internal/utils"
)

// DashboardHandler ダッシュボード関連のHTTPリクエストを処理
type DashboardHandler struct {
	dashboardService services.DashboardServiceInterface
}

// NewDashboardHandler 新しいDashboardHandlerを作成
func NewDashboardHandler(dashboardService services.DashboardServiceInterface) *DashboardHandler {
	return &DashboardHandler{
		dashboardService: dashboardService,
	}
}

// GetStats GET /api/v1/dashboard/stats を処理してダッシュボード統計情報を返す
func (h *DashboardHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	stats, err := h.dashboardService.GetStats(r.Context())
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "統計情報の取得に失敗しました")
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"stats": stats,
	})
}
