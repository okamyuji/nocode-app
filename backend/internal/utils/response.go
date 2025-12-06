package utils

import (
	"encoding/json"
	"net/http"

	"nocode-app/backend/internal/models"
)

// WriteJSON JSONレスポンスを書き込む
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			// エラーをログに記録するが、この時点でできることは限られている
			http.Error(w, `{"error":"レスポンスのエンコードに失敗しました"}`, http.StatusInternalServerError)
		}
	}
}

// WriteErrorResponse エラーレスポンスを書き込む
func WriteErrorResponse(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, models.ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
		Code:    status,
	})
}

// WriteSuccessResponse 成功レスポンスを書き込む
func WriteSuccessResponse(w http.ResponseWriter, message string, data interface{}) {
	WriteJSON(w, http.StatusOK, models.SuccessResponse{
		Message: message,
		Data:    data,
	})
}

// ParseJSON JSONリクエストボディをパースする
func ParseJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}
