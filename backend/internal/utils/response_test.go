package utils_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nocode-app/backend/internal/models"
	"nocode-app/backend/internal/utils"
)

func TestWriteJSON(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		data       interface{}
		wantStatus int
		checkBody  bool
	}{
		{
			name:       "success with data",
			status:     http.StatusOK,
			data:       map[string]string{"key": "value"},
			wantStatus: http.StatusOK,
			checkBody:  true,
		},
		{
			name:       "created status",
			status:     http.StatusCreated,
			data:       map[string]int{"id": 1},
			wantStatus: http.StatusCreated,
			checkBody:  true,
		},
		{
			name:       "nil data",
			status:     http.StatusNoContent,
			data:       nil,
			wantStatus: http.StatusNoContent,
			checkBody:  false,
		},
		{
			name:       "struct data",
			status:     http.StatusOK,
			data:       models.SuccessResponse{Message: "ok"},
			wantStatus: http.StatusOK,
			checkBody:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			utils.WriteJSON(w, tt.status, tt.data)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			if tt.checkBody {
				var result map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &result)
				assert.NoError(t, err)
			}
		})
	}
}

func TestWriteErrorResponse(t *testing.T) {
	tests := []struct {
		name        string
		status      int
		message     string
		wantError   string
		wantMessage string
	}{
		{
			name:        "bad request",
			status:      http.StatusBadRequest,
			message:     "invalid input",
			wantError:   "Bad Request",
			wantMessage: "invalid input",
		},
		{
			name:        "unauthorized",
			status:      http.StatusUnauthorized,
			message:     "authentication required",
			wantError:   "Unauthorized",
			wantMessage: "authentication required",
		},
		{
			name:        "not found",
			status:      http.StatusNotFound,
			message:     "resource not found",
			wantError:   "Not Found",
			wantMessage: "resource not found",
		},
		{
			name:        "internal server error",
			status:      http.StatusInternalServerError,
			message:     "something went wrong",
			wantError:   "Internal Server Error",
			wantMessage: "something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			utils.WriteErrorResponse(w, tt.status, tt.message)

			assert.Equal(t, tt.status, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var result models.ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &result)
			require.NoError(t, err)
			assert.Equal(t, tt.wantError, result.Error)
			assert.Equal(t, tt.wantMessage, result.Message)
			assert.Equal(t, tt.status, result.Code)
		})
	}
}

func TestWriteSuccessResponse(t *testing.T) {
	tests := []struct {
		name    string
		message string
		data    interface{}
	}{
		{
			name:    "simple success",
			message: "operation successful",
			data:    nil,
		},
		{
			name:    "success with data",
			message: "created",
			data:    map[string]int{"id": 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			utils.WriteSuccessResponse(w, tt.message, tt.data)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var result models.SuccessResponse
			err := json.Unmarshal(w.Body.Bytes(), &result)
			require.NoError(t, err)
			assert.Equal(t, tt.message, result.Message)
		})
	}
}

func TestParseJSON(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		dest    interface{}
		wantErr bool
	}{
		{
			name:    "valid JSON",
			body:    `{"name":"test","value":123}`,
			dest:    &map[string]interface{}{},
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			body:    `{invalid json}`,
			dest:    &map[string]interface{}{},
			wantErr: true,
		},
		{
			name:    "empty body",
			body:    ``,
			dest:    &map[string]interface{}{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(tt.body))
			err := utils.ParseJSON(r, tt.dest)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

type TestStruct struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func TestParseJSON_Struct(t *testing.T) {
	body := `{"name":"test","value":42}`
	r := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))

	var dest TestStruct
	err := utils.ParseJSON(r, &dest)

	require.NoError(t, err)
	assert.Equal(t, "test", dest.Name)
	assert.Equal(t, 42, dest.Value)
}
