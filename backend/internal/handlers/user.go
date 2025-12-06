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

// UserHandler ユーザー管理のHTTPリクエストを処理
type UserHandler struct {
	userService services.UserServiceInterface
	validator   *utils.Validator
}

// NewUserHandler 新しいUserHandlerを作成
func NewUserHandler(userService services.UserServiceInterface, validator *utils.Validator) *UserHandler {
	return &UserHandler{
		userService: userService,
		validator:   validator,
	}
}

// List GET /api/v1/users を処理
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// ページネーションパラメータをパース
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	resp, err := h.userService.GetUsers(r.Context(), claims.Role, page, limit)
	if err != nil {
		if errors.Is(err, services.ErrNotAdmin) {
			utils.WriteErrorResponse(w, http.StatusForbidden, err.Error())
			return
		}
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to get users")
		return
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

// Get GET /api/v1/users/:id を処理
func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := extractUserID(r.URL.Path)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	user, err := h.userService.GetUser(r.Context(), claims.Role, userID)
	if err != nil {
		if errors.Is(err, services.ErrNotAdmin) {
			utils.WriteErrorResponse(w, http.StatusForbidden, err.Error())
			return
		}
		if errors.Is(err, services.ErrUserNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	utils.WriteJSON(w, http.StatusOK, user)
}

// Create POST /api/v1/users を処理
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req models.CreateUserRequest
	if err := h.validator.ParseAndValidate(r, &req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.userService.CreateUser(r.Context(), claims.Role, &req)
	if err != nil {
		if errors.Is(err, services.ErrNotAdmin) {
			utils.WriteErrorResponse(w, http.StatusForbidden, err.Error())
			return
		}
		if errors.Is(err, services.ErrEmailAlreadyExists) {
			utils.WriteErrorResponse(w, http.StatusConflict, err.Error())
			return
		}
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	utils.WriteJSON(w, http.StatusCreated, user)
}

// Update PUT /api/v1/users/:id を処理
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := extractUserID(r.URL.Path)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	var req models.UpdateUserRequest
	if err := h.validator.ParseAndValidate(r, &req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.userService.UpdateUser(r.Context(), claims.UserID, claims.Role, userID, &req)
	if err != nil {
		if errors.Is(err, services.ErrNotAdmin) {
			utils.WriteErrorResponse(w, http.StatusForbidden, err.Error())
			return
		}
		if errors.Is(err, services.ErrUserNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, services.ErrCannotChangeSelfRole) {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to update user")
		return
	}

	utils.WriteJSON(w, http.StatusOK, user)
}

// Delete DELETE /api/v1/users/:id を処理
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := extractUserID(r.URL.Path)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	err = h.userService.DeleteUser(r.Context(), claims.UserID, claims.Role, userID)
	if err != nil {
		if errors.Is(err, services.ErrNotAdmin) {
			utils.WriteErrorResponse(w, http.StatusForbidden, err.Error())
			return
		}
		if errors.Is(err, services.ErrUserNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, services.ErrCannotDeleteSelf) {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to delete user")
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "user deleted"})
}

// UpdateProfile PUT /api/v1/auth/profile を処理
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req models.UpdateProfileRequest
	if err := h.validator.ParseAndValidate(r, &req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.userService.UpdateProfile(r.Context(), claims.UserID, &req)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to update profile")
		return
	}

	utils.WriteJSON(w, http.StatusOK, user)
}

// ChangePassword PUT /api/v1/auth/password を処理
func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req models.ChangePasswordRequest
	if err := h.validator.ParseAndValidate(r, &req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	err := h.userService.ChangePassword(r.Context(), claims.UserID, &req)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, services.ErrInvalidPassword) {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to change password")
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "パスワードを変更しました"})
}

// extractUserID URLパスからユーザーIDを抽出（例: /api/v1/users/123）
func extractUserID(path string) (uint64, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 4 {
		return 0, errors.New("invalid path")
	}
	return strconv.ParseUint(parts[3], 10, 64)
}
