package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"nocode-app/backend/internal/models"
	"nocode-app/backend/internal/services"
	"nocode-app/backend/internal/utils"
)

// ViewHandler ビュー関連のHTTPリクエストを処理
type ViewHandler struct {
	viewService services.ViewServiceInterface
	validator   *utils.Validator
}

// NewViewHandler 新しいViewHandlerを作成
func NewViewHandler(viewService services.ViewServiceInterface, validator *utils.Validator) *ViewHandler {
	return &ViewHandler{
		viewService: viewService,
		validator:   validator,
	}
}

// List アプリの全ビューを一覧取得
func (h *ViewHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	appID, err := extractAppIDFromViewPath(r.URL.Path)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "invalid app id")
		return
	}

	views, err := h.viewService.GetViews(r.Context(), appID)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to get views")
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{"views": views})
}

// Create 新しいビューを作成
func (h *ViewHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	appID, err := extractAppIDFromViewPath(r.URL.Path)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "invalid app id")
		return
	}

	var req models.CreateViewRequest
	if err := h.validator.ParseAndValidate(r, &req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	view, err := h.viewService.CreateView(r.Context(), appID, &req)
	if err != nil {
		if errors.Is(err, services.ErrAppNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to create view")
		return
	}

	utils.WriteJSON(w, http.StatusCreated, view)
}

// Update ビューを更新
func (h *ViewHandler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	viewID, err := extractViewID(r.URL.Path)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "invalid view id")
		return
	}

	var req models.UpdateViewRequest
	if err := h.validator.ParseAndValidate(r, &req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	view, err := h.viewService.UpdateView(r.Context(), viewID, &req)
	if err != nil {
		if errors.Is(err, services.ErrViewNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to update view")
		return
	}

	utils.WriteJSON(w, http.StatusOK, view)
}

// Delete ビューを削除
func (h *ViewHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	viewID, err := extractViewID(r.URL.Path)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "invalid view id")
		return
	}

	if err := h.viewService.DeleteView(r.Context(), viewID); err != nil {
		if errors.Is(err, services.ErrViewNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to delete view")
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.SuccessResponse{Message: "ビューを削除しました"})
}

// extractAppIDFromViewPath URLパスからアプリIDを抽出
// 想定パス形式: /api/v1/apps/{appId}/views
func extractAppIDFromViewPath(path string) (uint64, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 4 {
		return 0, errors.New("invalid path")
	}
	return strconv.ParseUint(parts[3], 10, 64)
}

// extractViewID URLパスからビューIDを抽出
// 想定パス形式: /api/v1/apps/{appId}/views/{viewId}
func extractViewID(path string) (uint64, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 6 {
		return 0, errors.New("invalid path")
	}
	// アプリIDの形式を検証
	if _, err := strconv.ParseUint(parts[3], 10, 64); err != nil {
		return 0, err
	}
	return strconv.ParseUint(parts[5], 10, 64)
}
