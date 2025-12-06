package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"nocode-app/backend/internal/middleware"
)

func TestLoggerMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		wantStatusCode int
		writeBody      bool
	}{
		{
			name:           "GET request",
			method:         http.MethodGet,
			path:           "/api/test",
			wantStatusCode: http.StatusOK,
			writeBody:      true,
		},
		{
			name:           "POST request",
			method:         http.MethodPost,
			path:           "/api/users",
			wantStatusCode: http.StatusCreated,
			writeBody:      true,
		},
		{
			name:           "no body",
			method:         http.MethodDelete,
			path:           "/api/users/1",
			wantStatusCode: http.StatusNoContent,
			writeBody:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.wantStatusCode)
				if tt.writeBody {
					_, _ = w.Write([]byte("response body"))
				}
			})

			req := httptest.NewRequest(tt.method, tt.path, nil)
			rr := httptest.NewRecorder()

			handler := middleware.LoggerMiddleware(nextHandler)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatusCode, rr.Code)
		})
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	t.Run("no panic", func(t *testing.T) {
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("success"))
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()

		handler := middleware.RecoveryMiddleware(nextHandler)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "success", rr.Body.String())
	})

	t.Run("with panic", func(t *testing.T) {
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("test panic")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()

		handler := middleware.RecoveryMiddleware(nextHandler)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "internal server error")
	})

	t.Run("with panic error value", func(t *testing.T) {
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("some error occurred")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()

		handler := middleware.RecoveryMiddleware(nextHandler)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestResponseWriter(t *testing.T) {
	t.Run("captures status code", func(t *testing.T) {
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("not found"))
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()

		handler := middleware.LoggerMiddleware(nextHandler)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("captures body size", func(t *testing.T) {
		body := "this is a test response body"
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(body))
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()

		handler := middleware.LoggerMiddleware(nextHandler)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, body, rr.Body.String())
		assert.Equal(t, len(body), rr.Body.Len())
	})
}
