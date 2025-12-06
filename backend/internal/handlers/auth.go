package handlers

import (
	"errors"
	"net/http"

	"nocode-app/backend/internal/middleware"
	"nocode-app/backend/internal/models"
	"nocode-app/backend/internal/services"
	"nocode-app/backend/internal/utils"
)

// AuthHandler 認証エンドポイントを処理する構造体
type AuthHandler struct {
	authService services.AuthServiceInterface
	validator   *utils.Validator
}

// NewAuthHandler 新しいAuthHandlerを作成する
func NewAuthHandler(authService services.AuthServiceInterface, validator *utils.Validator) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validator:   validator,
	}
}

// Register ユーザー登録を処理する
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "メソッドが許可されていません")
		return
	}

	var req models.RegisterRequest
	if err := h.validator.ParseAndValidate(r, &req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.authService.Register(r.Context(), &req)
	if err != nil {
		if errors.Is(err, services.ErrEmailAlreadyExists) {
			utils.WriteErrorResponse(w, http.StatusConflict, err.Error())
			return
		}
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "ユーザー登録に失敗しました")
		return
	}

	utils.WriteJSON(w, http.StatusCreated, resp)
}

// Login ユーザーログインを処理する
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "メソッドが許可されていません")
		return
	}

	var req models.LoginRequest
	if err := h.validator.ParseAndValidate(r, &req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.authService.Login(r.Context(), &req)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, err.Error())
			return
		}
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "ログインに失敗しました")
		return
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

// Me 現在のユーザーを返す
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "メソッドが許可されていません")
		return
	}

	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "認証されていません")
		return
	}

	user, err := h.authService.GetCurrentUser(r.Context(), claims.UserID)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "ユーザーの取得に失敗しました")
		return
	}

	utils.WriteJSON(w, http.StatusOK, user)
}

// Refresh JWTトークンを更新する
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "メソッドが許可されていません")
		return
	}

	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "認証されていません")
		return
	}

	token, err := h.authService.RefreshToken(claims)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "トークンの更新に失敗しました")
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"token": token})
}
