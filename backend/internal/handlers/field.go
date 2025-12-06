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

// FieldHandler フィールドエンドポイントを処理する構造体
type FieldHandler struct {
	fieldService services.FieldServiceInterface
	validator    *utils.Validator
}

// NewFieldHandler 新しいFieldHandlerを作成する
func NewFieldHandler(fieldService services.FieldServiceInterface, validator *utils.Validator) *FieldHandler {
	return &FieldHandler{
		fieldService: fieldService,
		validator:    validator,
	}
}

// List アプリの全フィールドを一覧表示する
func (h *FieldHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "メソッドが許可されていません")
		return
	}

	appID, err := extractAppIDFromFieldPath(r.URL.Path)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "無効なアプリIDです")
		return
	}

	fields, err := h.fieldService.GetFields(r.Context(), appID)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "フィールドの取得に失敗しました")
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{"fields": fields})
}

// Create 新しいフィールドを作成する
func (h *FieldHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "メソッドが許可されていません")
		return
	}

	appID, err := extractAppIDFromFieldPath(r.URL.Path)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "無効なアプリIDです")
		return
	}

	var req models.CreateFieldRequest
	if err := h.validator.ParseAndValidate(r, &req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	field, err := h.fieldService.CreateField(r.Context(), appID, &req)
	if err != nil {
		if errors.Is(err, services.ErrFieldCodeExists) {
			utils.WriteErrorResponse(w, http.StatusConflict, err.Error())
			return
		}
		if errors.Is(err, services.ErrAppNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "フィールドの作成に失敗しました")
		return
	}

	utils.WriteJSON(w, http.StatusCreated, field)
}

// Update フィールドを更新する
func (h *FieldHandler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "メソッドが許可されていません")
		return
	}

	_, fieldID, err := extractAppAndFieldID(r.URL.Path)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "無効なフィールドIDです")
		return
	}

	var req models.UpdateFieldRequest
	if err := h.validator.ParseAndValidate(r, &req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	field, err := h.fieldService.UpdateField(r.Context(), fieldID, &req)
	if err != nil {
		if errors.Is(err, services.ErrFieldNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "フィールドの更新に失敗しました")
		return
	}

	utils.WriteJSON(w, http.StatusOK, field)
}

// Delete フィールドを削除する
func (h *FieldHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "メソッドが許可されていません")
		return
	}

	appID, fieldID, err := extractAppAndFieldID(r.URL.Path)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "無効なフィールドIDです")
		return
	}

	if err := h.fieldService.DeleteField(r.Context(), appID, fieldID); err != nil {
		if errors.Is(err, services.ErrFieldNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, services.ErrAppNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "フィールドの削除に失敗しました")
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.SuccessResponse{Message: "フィールドを削除しました"})
}

// UpdateOrder フィールドの表示順序を更新する
func (h *FieldHandler) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "メソッドが許可されていません")
		return
	}

	appID, err := extractAppIDFromFieldPath(r.URL.Path)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "無効なアプリIDです")
		return
	}

	var req models.UpdateFieldOrderRequest
	if err := h.validator.ParseAndValidate(r, &req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.fieldService.UpdateFieldOrder(r.Context(), appID, &req); err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "フィールド順序の更新に失敗しました")
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.SuccessResponse{Message: "フィールド順序を更新しました"})
}

// extractAppIDFromFieldPath URLパスからアプリIDを抽出する
// 期待されるパス形式: /api/v1/apps/{appId}/fields
func extractAppIDFromFieldPath(path string) (uint64, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 4 {
		return 0, errors.New("無効なパスです")
	}
	return strconv.ParseUint(parts[3], 10, 64)
}

// extractAppAndFieldID URLパスからアプリIDとフィールドIDの両方を抽出する
// 期待されるパス形式: /api/v1/apps/{appId}/fields/{fieldId}
func extractAppAndFieldID(path string) (appID, fieldID uint64, err error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 6 {
		return 0, 0, errors.New("無効なパスです")
	}
	appID, err = strconv.ParseUint(parts[3], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	fieldID, err = strconv.ParseUint(parts[5], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	return appID, fieldID, nil
}
