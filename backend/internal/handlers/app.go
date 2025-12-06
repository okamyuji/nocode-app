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

// AppHandler アプリエンドポイントを処理する構造体
type AppHandler struct {
	appService services.AppServiceInterface
	validator  *utils.Validator
}

// NewAppHandler 新しいAppHandlerを作成する
func NewAppHandler(appService services.AppServiceInterface, validator *utils.Validator) *AppHandler {
	return &AppHandler{
		appService: appService,
		validator:  validator,
	}
}

// List 全アプリを一覧表示する
func (h *AppHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "メソッドが許可されていません")
		return
	}

	page := utils.GetQueryParamInt(r, "page", 1)
	limit := utils.GetQueryParamInt(r, "limit", 20)

	resp, err := h.appService.GetApps(r.Context(), page, limit)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "アプリの取得に失敗しました")
		return
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

// Create 新しいアプリを作成する
func (h *AppHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "メソッドが許可されていません")
		return
	}

	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "認証されていません")
		return
	}

	var req models.CreateAppRequest
	if err := h.validator.ParseAndValidate(r, &req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.appService.CreateApp(r.Context(), claims.UserID, &req)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "アプリの作成に失敗しました")
		return
	}

	utils.WriteJSON(w, http.StatusCreated, resp)
}

// Get IDでアプリを取得する
func (h *AppHandler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "メソッドが許可されていません")
		return
	}

	appID, err := extractAppID(r.URL.Path)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "無効なアプリIDです")
		return
	}

	resp, err := h.appService.GetApp(r.Context(), appID)
	if err != nil {
		if errors.Is(err, services.ErrAppNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "アプリの取得に失敗しました")
		return
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

// Update アプリを更新する
func (h *AppHandler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "メソッドが許可されていません")
		return
	}

	appID, err := extractAppID(r.URL.Path)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "無効なアプリIDです")
		return
	}

	var req models.UpdateAppRequest
	if err := h.validator.ParseAndValidate(r, &req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.appService.UpdateApp(r.Context(), appID, &req)
	if err != nil {
		if errors.Is(err, services.ErrAppNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "アプリの更新に失敗しました")
		return
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

// Delete アプリを削除する
func (h *AppHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "メソッドが許可されていません")
		return
	}

	appID, err := extractAppID(r.URL.Path)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "無効なアプリIDです")
		return
	}

	if err := h.appService.DeleteApp(r.Context(), appID); err != nil {
		if errors.Is(err, services.ErrAppNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "アプリの削除に失敗しました")
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.SuccessResponse{Message: "アプリを削除しました"})
}

// extractAppID URLパスからアプリIDを抽出する
// 期待されるパス形式: /api/v1/apps/{id}
func extractAppID(path string) (uint64, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 4 {
		return 0, errors.New("無効なパスです")
	}
	return strconv.ParseUint(parts[3], 10, 64)
}
