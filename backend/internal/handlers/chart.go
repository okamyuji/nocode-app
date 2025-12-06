package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"nocode-app/backend/internal/middleware"
	"nocode-app/backend/internal/models"
	"nocode-app/backend/internal/services"
	"nocode-app/backend/internal/utils"
)

// ChartHandler チャート関連のHTTPリクエストを処理
type ChartHandler struct {
	chartService services.ChartServiceInterface
	validator    *utils.Validator
}

// NewChartHandler 新しいChartHandlerを作成
func NewChartHandler(chartService services.ChartServiceInterface, validator *utils.Validator) *ChartHandler {
	return &ChartHandler{
		chartService: chartService,
		validator:    validator,
	}
}

// GetData チャート用の集計データを取得
func (h *ChartHandler) GetData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	appID, err := extractAppIDFromChartPath(r.URL.Path)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "invalid app id")
		return
	}

	var req models.ChartDataRequest
	if err := h.validator.ParseAndValidate(r, &req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// カスタムバリデーション: x_axis.fieldは必須
	if req.XAxis.Field == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "x_axis.field is required")
		return
	}

	// カスタムバリデーション: count以外の集計にはy_axis.fieldが必須
	if req.YAxis.Aggregation != "count" && req.YAxis.Aggregation != "" && req.YAxis.Field == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "y_axis.field is required for non-count aggregations")
		return
	}

	resp, err := h.chartService.GetChartData(r.Context(), appID, &req)
	if err != nil {
		if errors.Is(err, services.ErrAppNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to get chart data")
		return
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

// GetConfigs アプリの全チャート設定を取得
func (h *ChartHandler) GetConfigs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	appID, err := extractAppIDFromChartPath(r.URL.Path)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "invalid app id")
		return
	}

	configs, err := h.chartService.GetChartConfigs(r.Context(), appID)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to get chart configs")
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{"configs": configs})
}

// SaveConfig チャート設定を保存
func (h *ChartHandler) SaveConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	appID, err := extractAppIDFromChartPath(r.URL.Path)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "invalid app id")
		return
	}

	var req models.SaveChartConfigRequest
	if err := h.validator.ParseAndValidate(r, &req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	config, err := h.chartService.SaveChartConfig(r.Context(), appID, claims.UserID, &req)
	if err != nil {
		if errors.Is(err, services.ErrAppNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to save chart config")
		return
	}

	utils.WriteJSON(w, http.StatusCreated, config)
}

// DeleteConfig チャート設定を削除
func (h *ChartHandler) DeleteConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	configID, err := extractChartConfigID(r.URL.Path)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "invalid config id")
		return
	}

	if err := h.chartService.DeleteChartConfig(r.Context(), configID); err != nil {
		if errors.Is(err, services.ErrChartConfigNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to delete chart config")
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.SuccessResponse{Message: "チャート設定を削除しました"})
}

// extractAppIDFromChartPath URLパスからアプリIDを抽出
// 想定パス形式: /api/v1/apps/{appId}/charts/...
func extractAppIDFromChartPath(path string) (uint64, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 4 {
		return 0, errors.New("invalid path")
	}
	return strconv.ParseUint(parts[3], 10, 64)
}

// extractChartConfigID URLパスからチャート設定IDを抽出
// 想定パス形式: /api/v1/apps/{appId}/charts/config/{configId}
func extractChartConfigID(path string) (uint64, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 7 {
		return 0, errors.New("invalid path")
	}
	return strconv.ParseUint(parts[6], 10, 64)
}
