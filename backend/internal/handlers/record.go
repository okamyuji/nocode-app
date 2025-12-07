package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"nocode-app/backend/internal/middleware"
	"nocode-app/backend/internal/models"
	"nocode-app/backend/internal/repositories"
	"nocode-app/backend/internal/services"
	"nocode-app/backend/internal/utils"
)

// RecordHandler レコードエンドポイントを処理する構造体
type RecordHandler struct {
	recordService services.RecordServiceInterface
	validator     *utils.Validator
}

// NewRecordHandler 新しいRecordHandlerを作成する
func NewRecordHandler(recordService services.RecordServiceInterface, validator *utils.Validator) *RecordHandler {
	return &RecordHandler{
		recordService: recordService,
		validator:     validator,
	}
}

// List アプリの全レコードを一覧表示する
func (h *RecordHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "メソッドが許可されていません")
		return
	}

	appID, err := extractAppIDFromRecordPath(r.URL.Path)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "無効なアプリIDです")
		return
	}

	// クエリオプションをパース
	opts := repositories.RecordQueryOptions{
		Page:    utils.GetQueryParamInt(r, "page", 1),
		Limit:   utils.GetQueryParamInt(r, "limit", 20),
		Sort:    utils.GetQueryParam(r, "sort", ""),
		Order:   utils.GetQueryParam(r, "order", "desc"),
		Filters: parseFilters(r),
	}

	resp, err := h.recordService.GetRecords(r.Context(), appID, opts)
	if err != nil {
		if errors.Is(err, services.ErrAppNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, services.ErrEncryptionNotInitialized) {
			utils.WriteErrorResponse(w, http.StatusServiceUnavailable, err.Error())
			return
		}
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "レコードの取得に失敗しました")
		return
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

// Create 新しいレコードを作成する
func (h *RecordHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "メソッドが許可されていません")
		return
	}

	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "認証されていません")
		return
	}

	appID, err := extractAppIDFromRecordPath(r.URL.Path)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "無効なアプリIDです")
		return
	}

	var req models.CreateRecordRequest
	if err := h.validator.ParseAndValidate(r, &req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.recordService.CreateRecord(r.Context(), appID, claims.UserID, &req)
	if err != nil {
		if errors.Is(err, services.ErrAppNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, services.ErrExternalAppReadOnly) {
			utils.WriteErrorResponse(w, http.StatusForbidden, err.Error())
			return
		}
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "レコードの作成に失敗しました")
		return
	}

	utils.WriteJSON(w, http.StatusCreated, resp)
}

// Get IDでレコードを取得する
func (h *RecordHandler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "メソッドが許可されていません")
		return
	}

	appID, recordID, err := extractAppAndRecordID(r.URL.Path)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "無効なレコードIDです")
		return
	}

	resp, err := h.recordService.GetRecord(r.Context(), appID, recordID)
	if err != nil {
		if errors.Is(err, services.ErrRecordNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, services.ErrAppNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, services.ErrEncryptionNotInitialized) {
			utils.WriteErrorResponse(w, http.StatusServiceUnavailable, err.Error())
			return
		}
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "レコードの取得に失敗しました")
		return
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

// Update レコードを更新する
func (h *RecordHandler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "メソッドが許可されていません")
		return
	}

	appID, recordID, err := extractAppAndRecordID(r.URL.Path)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "無効なレコードIDです")
		return
	}

	var req models.UpdateRecordRequest
	if err := h.validator.ParseAndValidate(r, &req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.recordService.UpdateRecord(r.Context(), appID, recordID, &req)
	if err != nil {
		if errors.Is(err, services.ErrAppNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, services.ErrExternalAppReadOnly) {
			utils.WriteErrorResponse(w, http.StatusForbidden, err.Error())
			return
		}
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "レコードの更新に失敗しました")
		return
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

// Delete レコードを削除する
func (h *RecordHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "メソッドが許可されていません")
		return
	}

	appID, recordID, err := extractAppAndRecordID(r.URL.Path)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "無効なレコードIDです")
		return
	}

	if err := h.recordService.DeleteRecord(r.Context(), appID, recordID); err != nil {
		if errors.Is(err, services.ErrAppNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, services.ErrExternalAppReadOnly) {
			utils.WriteErrorResponse(w, http.StatusForbidden, err.Error())
			return
		}
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "レコードの削除に失敗しました")
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.SuccessResponse{Message: "レコードを削除しました"})
}

// BulkCreate 複数のレコードを作成する
func (h *RecordHandler) BulkCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "メソッドが許可されていません")
		return
	}

	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "認証されていません")
		return
	}

	appID, err := extractAppIDFromRecordPath(r.URL.Path)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "無効なアプリIDです")
		return
	}

	var req models.BulkCreateRecordRequest
	if err := h.validator.ParseAndValidate(r, &req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	records, err := h.recordService.BulkCreateRecords(r.Context(), appID, claims.UserID, &req)
	if err != nil {
		if errors.Is(err, services.ErrAppNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, services.ErrExternalAppReadOnly) {
			utils.WriteErrorResponse(w, http.StatusForbidden, err.Error())
			return
		}
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "レコードの作成に失敗しました")
		return
	}

	utils.WriteJSON(w, http.StatusCreated, map[string]interface{}{"records": records})
}

// BulkDelete 複数のレコードを削除する
func (h *RecordHandler) BulkDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "メソッドが許可されていません")
		return
	}

	appID, err := extractAppIDFromRecordPath(r.URL.Path)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "無効なアプリIDです")
		return
	}

	var req models.BulkDeleteRecordRequest
	if err := h.validator.ParseAndValidate(r, &req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.recordService.BulkDeleteRecords(r.Context(), appID, &req); err != nil {
		if errors.Is(err, services.ErrAppNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, services.ErrExternalAppReadOnly) {
			utils.WriteErrorResponse(w, http.StatusForbidden, err.Error())
			return
		}
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "レコードの削除に失敗しました")
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.SuccessResponse{Message: "レコードを削除しました"})
}

// extractAppIDFromRecordPath URLパスからアプリIDを抽出する
// 期待されるパス形式: /api/v1/apps/{appId}/records
func extractAppIDFromRecordPath(path string) (uint64, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 4 {
		return 0, errors.New("無効なパスです")
	}
	return strconv.ParseUint(parts[3], 10, 64)
}

// extractAppAndRecordID URLパスからアプリIDとレコードIDの両方を抽出する
// 期待されるパス形式: /api/v1/apps/{appId}/records/{recordId}
func extractAppAndRecordID(path string) (appID, recordID uint64, err error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 6 {
		return 0, 0, errors.New("無効なパスです")
	}
	appID, err = strconv.ParseUint(parts[3], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	recordID, err = strconv.ParseUint(parts[5], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	return appID, recordID, nil
}

// parseFilters フィルタークエリパラメータをパースする
// 形式: filter=field:op:value
func parseFilters(r *http.Request) []models.FilterItem {
	var filters []models.FilterItem
	filterParams := r.URL.Query()["filter"]

	for _, param := range filterParams {
		parts := strings.SplitN(param, ":", 3)
		if len(parts) == 3 {
			filters = append(filters, models.FilterItem{
				Field:    parts[0],
				Operator: parts[1],
				Value:    parts[2],
			})
		}
	}

	return filters
}
