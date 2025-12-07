package handlers

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"nocode-app/backend/internal/middleware"
	"nocode-app/backend/internal/models"
	"nocode-app/backend/internal/services"
	"nocode-app/backend/internal/utils"
)

// DataSourceHandler データソースエンドポイントを処理する構造体
type DataSourceHandler struct {
	dsService services.DataSourceServiceInterface
	validator *utils.Validator
}

// NewDataSourceHandler 新しいDataSourceHandlerを作成する
func NewDataSourceHandler(dsService services.DataSourceServiceInterface, validator *utils.Validator) *DataSourceHandler {
	return &DataSourceHandler{
		dsService: dsService,
		validator: validator,
	}
}

// List 全データソースを一覧表示する
func (h *DataSourceHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "メソッドが許可されていません")
		return
	}

	page := utils.GetQueryParamInt(r, "page", 1)
	limit := utils.GetQueryParamInt(r, "limit", 20)

	resp, err := h.dsService.GetDataSources(r.Context(), page, limit)
	if err != nil {
		log.Printf("データソース一覧取得エラー: %v", err)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "データソースの取得に失敗しました")
		return
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

// Create 新しいデータソースを作成する
func (h *DataSourceHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "メソッドが許可されていません")
		return
	}

	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "認証されていません")
		return
	}

	var req models.CreateDataSourceRequest
	if err := h.validator.ParseAndValidate(r, &req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.dsService.CreateDataSource(r.Context(), claims.UserID, &req)
	if err != nil {
		if errors.Is(err, services.ErrDataSourceNameExists) {
			utils.WriteErrorResponse(w, http.StatusConflict, err.Error())
			return
		}
		if errors.Is(err, services.ErrInvalidDBType) {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, services.ErrEncryptionNotInitialized) {
			utils.WriteErrorResponse(w, http.StatusServiceUnavailable, err.Error())
			return
		}
		log.Printf("データソース作成エラー: %v", err)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "データソースの作成に失敗しました")
		return
	}

	utils.WriteJSON(w, http.StatusCreated, resp)
}

// Get IDでデータソースを取得する
func (h *DataSourceHandler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "メソッドが許可されていません")
		return
	}

	dsID, err := extractDataSourceID(r.URL.Path)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "無効なデータソースIDです")
		return
	}

	resp, err := h.dsService.GetDataSource(r.Context(), dsID)
	if err != nil {
		if errors.Is(err, services.ErrDataSourceNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		log.Printf("データソース取得エラー: %v", err)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "データソースの取得に失敗しました")
		return
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

// Update データソースを更新する
func (h *DataSourceHandler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "メソッドが許可されていません")
		return
	}

	dsID, err := extractDataSourceID(r.URL.Path)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "無効なデータソースIDです")
		return
	}

	var req models.UpdateDataSourceRequest
	if err := h.validator.ParseAndValidate(r, &req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.dsService.UpdateDataSource(r.Context(), dsID, &req)
	if err != nil {
		if errors.Is(err, services.ErrDataSourceNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, services.ErrDataSourceNameExists) {
			utils.WriteErrorResponse(w, http.StatusConflict, err.Error())
			return
		}
		log.Printf("データソース更新エラー: %v", err)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "データソースの更新に失敗しました")
		return
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

// Delete データソースを削除する
func (h *DataSourceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "メソッドが許可されていません")
		return
	}

	dsID, err := extractDataSourceID(r.URL.Path)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "無効なデータソースIDです")
		return
	}

	if err := h.dsService.DeleteDataSource(r.Context(), dsID); err != nil {
		if errors.Is(err, services.ErrDataSourceNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		log.Printf("データソース削除エラー: %v", err)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "データソースの削除に失敗しました")
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.SuccessResponse{Message: "データソースを削除しました"})
}

// TestConnection テスト接続を実行する
func (h *DataSourceHandler) TestConnection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "メソッドが許可されていません")
		return
	}

	var req models.TestConnectionRequest
	if err := h.validator.ParseAndValidate(r, &req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.dsService.TestConnection(r.Context(), &req)
	if err != nil {
		log.Printf("テスト接続エラー: %v", err)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "テスト接続の実行に失敗しました")
		return
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

// GetTables データソースのテーブル一覧を取得する
func (h *DataSourceHandler) GetTables(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "メソッドが許可されていません")
		return
	}

	dsID, err := extractDataSourceID(r.URL.Path)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "無効なデータソースIDです")
		return
	}

	resp, err := h.dsService.GetTables(r.Context(), dsID)
	if err != nil {
		if errors.Is(err, services.ErrDataSourceNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, services.ErrEncryptionNotInitialized) {
			utils.WriteErrorResponse(w, http.StatusServiceUnavailable, err.Error())
			return
		}
		log.Printf("テーブル一覧取得エラー: %v", err)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "テーブル一覧の取得に失敗しました")
		return
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

// GetColumns テーブルのカラム一覧を取得する
func (h *DataSourceHandler) GetColumns(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "メソッドが許可されていません")
		return
	}

	dsID, tableName, err := extractDataSourceIDAndTableName(r.URL.Path)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.dsService.GetColumns(r.Context(), dsID, tableName)
	if err != nil {
		if errors.Is(err, services.ErrDataSourceNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, services.ErrEncryptionNotInitialized) {
			utils.WriteErrorResponse(w, http.StatusServiceUnavailable, err.Error())
			return
		}
		log.Printf("カラム一覧取得エラー: %v", err)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "カラム一覧の取得に失敗しました")
		return
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

// extractDataSourceID URLパスからデータソースIDを抽出する
// 期待されるパス形式: /api/v1/datasources/{id}
func extractDataSourceID(path string) (uint64, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 4 {
		return 0, errors.New("無効なパスです")
	}
	return strconv.ParseUint(parts[3], 10, 64)
}

// extractDataSourceIDAndTableName URLパスからデータソースIDとテーブル名を抽出する
// 期待されるパス形式: /api/v1/datasources/{id}/tables/{tableName}/columns
func extractDataSourceIDAndTableName(path string) (uint64, string, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	// api/v1/datasources/{id}/tables/{tableName}/columns
	// [0]  [1]    [2]       [3]   [4]     [5]       [6]
	if len(parts) < 7 {
		return 0, "", errors.New("無効なパスです")
	}

	id, err := strconv.ParseUint(parts[3], 10, 64)
	if err != nil {
		return 0, "", errors.New("無効なデータソースIDです")
	}

	tableName := parts[5]
	if tableName == "" {
		return 0, "", errors.New("テーブル名が指定されていません")
	}

	// URLエンコードされた日本語テーブル名などをデコード
	decodedTableName, err := url.PathUnescape(tableName)
	if err != nil {
		return 0, "", errors.New("テーブル名のデコードに失敗しました")
	}

	return id, decodedTableName, nil
}
