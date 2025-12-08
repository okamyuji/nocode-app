package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"nocode-app/backend/internal/middleware"
	"nocode-app/backend/internal/models"
	"nocode-app/backend/internal/services"
	"nocode-app/backend/internal/utils"
)

// DashboardWidgetHandler ダッシュボードウィジェット関連のHTTPリクエストを処理
type DashboardWidgetHandler struct {
	widgetService services.DashboardWidgetServiceInterface
	validator     *utils.Validator
}

// NewDashboardWidgetHandler 新しいDashboardWidgetHandlerを作成
func NewDashboardWidgetHandler(widgetService services.DashboardWidgetServiceInterface, validator *utils.Validator) *DashboardWidgetHandler {
	return &DashboardWidgetHandler{
		widgetService: widgetService,
		validator:     validator,
	}
}

// List GET /api/v1/dashboard/widgets を処理してウィジェット一覧を返す
func (h *DashboardWidgetHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "認証が必要です")
		return
	}

	// クエリパラメータで表示中のみを取得するオプション
	visibleOnly := r.URL.Query().Get("visible") == "true"

	var response *models.DashboardWidgetListResponse
	var err error

	if visibleOnly {
		response, err = h.widgetService.GetVisibleWidgets(r.Context(), claims.UserID)
	} else {
		response, err = h.widgetService.GetWidgets(r.Context(), claims.UserID)
	}

	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "ウィジェット一覧の取得に失敗しました")
		return
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

// Create POST /api/v1/dashboard/widgets を処理して新しいウィジェットを作成
func (h *DashboardWidgetHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "認証が必要です")
		return
	}

	var req models.CreateDashboardWidgetRequest
	if err := h.validator.ParseAndValidate(r, &req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	response, err := h.widgetService.CreateWidget(r.Context(), claims.UserID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "見つかりません") {
			utils.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		if strings.Contains(err.Error(), "既に存在") {
			utils.WriteErrorResponse(w, http.StatusConflict, err.Error())
			return
		}
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "ウィジェットの作成に失敗しました")
		return
	}

	utils.WriteJSON(w, http.StatusCreated, response)
}

// Update PUT /api/v1/dashboard/widgets/{id} を処理してウィジェットを更新
func (h *DashboardWidgetHandler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "認証が必要です")
		return
	}

	// URLからウィジェットIDを取得
	widgetID, err := h.extractWidgetID(r.URL.Path)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "無効なウィジェットIDです")
		return
	}

	var req models.UpdateDashboardWidgetRequest
	if err := h.validator.ParseAndValidate(r, &req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	response, err := h.widgetService.UpdateWidget(r.Context(), claims.UserID, widgetID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "見つかりません") {
			utils.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		if strings.Contains(err.Error(), "権限がありません") {
			utils.WriteErrorResponse(w, http.StatusForbidden, err.Error())
			return
		}
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "ウィジェットの更新に失敗しました")
		return
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

// Delete DELETE /api/v1/dashboard/widgets/{id} を処理してウィジェットを削除
func (h *DashboardWidgetHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "認証が必要です")
		return
	}

	// URLからウィジェットIDを取得
	widgetID, err := h.extractWidgetID(r.URL.Path)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "無効なウィジェットIDです")
		return
	}

	if err := h.widgetService.DeleteWidget(r.Context(), claims.UserID, widgetID); err != nil {
		if strings.Contains(err.Error(), "見つかりません") {
			utils.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		if strings.Contains(err.Error(), "権限がありません") {
			utils.WriteErrorResponse(w, http.StatusForbidden, err.Error())
			return
		}
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "ウィジェットの削除に失敗しました")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Reorder PUT /api/v1/dashboard/widgets/reorder を処理してウィジェットの並び順を更新
func (h *DashboardWidgetHandler) Reorder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "認証が必要です")
		return
	}

	var req models.ReorderWidgetsRequest
	if err := h.validator.ParseAndValidate(r, &req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.widgetService.ReorderWidgets(r.Context(), claims.UserID, &req); err != nil {
		if strings.Contains(err.Error(), "無効なウィジェットID") {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "並び順の更新に失敗しました")
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "並び順を更新しました"})
}

// ToggleVisibility POST /api/v1/dashboard/widgets/{id}/toggle を処理して表示/非表示を切り替え
func (h *DashboardWidgetHandler) ToggleVisibility(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "認証が必要です")
		return
	}

	// URLからウィジェットIDを取得
	widgetID, err := h.extractWidgetIDForToggle(r.URL.Path)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "無効なウィジェットIDです")
		return
	}

	response, err := h.widgetService.ToggleVisibility(r.Context(), claims.UserID, widgetID)
	if err != nil {
		if strings.Contains(err.Error(), "見つかりません") {
			utils.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		if strings.Contains(err.Error(), "権限がありません") {
			utils.WriteErrorResponse(w, http.StatusForbidden, err.Error())
			return
		}
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "表示状態の切り替えに失敗しました")
		return
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

// extractWidgetID URLからウィジェットIDを取得
// /api/v1/dashboard/widgets/{id} の形式を想定
func (h *DashboardWidgetHandler) extractWidgetID(path string) (uint64, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	// parts: [api, v1, dashboard, widgets, {id}]
	if len(parts) < 5 {
		return 0, strconv.ErrSyntax
	}
	return strconv.ParseUint(parts[4], 10, 64)
}

// extractWidgetIDForToggle URLからウィジェットIDを取得（toggleエンドポイント用）
// /api/v1/dashboard/widgets/{id}/toggle の形式を想定
func (h *DashboardWidgetHandler) extractWidgetIDForToggle(path string) (uint64, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	// parts: [api, v1, dashboard, widgets, {id}, toggle]
	if len(parts) < 6 {
		return 0, strconv.ErrSyntax
	}
	return strconv.ParseUint(parts[4], 10, 64)
}
