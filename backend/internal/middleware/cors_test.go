package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"nocode-app/backend/internal/middleware"
)

func TestDefaultCORSConfig(t *testing.T) {
	config := middleware.DefaultCORSConfig()

	assert.NotNil(t, config)
	assert.Contains(t, config.AllowedOrigins, "*")
	assert.Contains(t, config.AllowedMethods, http.MethodGet)
	assert.Contains(t, config.AllowedMethods, http.MethodPost)
	assert.Contains(t, config.AllowedMethods, http.MethodPut)
	assert.Contains(t, config.AllowedMethods, http.MethodDelete)
	assert.Contains(t, config.AllowedMethods, http.MethodOptions)
	assert.Contains(t, config.AllowedHeaders, "Authorization")
	assert.Contains(t, config.AllowedHeaders, "Content-Type")
	assert.True(t, config.AllowCredentials)
	assert.Equal(t, 86400, config.MaxAge)
}

func TestCORSMiddleware(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	})

	tests := []struct {
		name            string
		config          *middleware.CORSConfig
		origin          string
		method          string
		wantOrigin      string
		wantStatusCode  int
		wantCredentials bool
	}{
		{
			name:            "default config - allow all origins",
			config:          nil,
			origin:          "http://localhost:3000",
			method:          http.MethodGet,
			wantOrigin:      "*",
			wantStatusCode:  http.StatusOK,
			wantCredentials: true,
		},
		{
			name: "specific origin - allowed",
			config: &middleware.CORSConfig{
				AllowedOrigins:   []string{"http://localhost:3000"},
				AllowedMethods:   []string{"GET", "POST"},
				AllowedHeaders:   []string{"Content-Type"},
				AllowCredentials: true,
			},
			origin:          "http://localhost:3000",
			method:          http.MethodGet,
			wantOrigin:      "http://localhost:3000",
			wantStatusCode:  http.StatusOK,
			wantCredentials: true,
		},
		{
			name: "specific origin - not allowed",
			config: &middleware.CORSConfig{
				AllowedOrigins:   []string{"http://localhost:3000"},
				AllowedMethods:   []string{"GET"},
				AllowedHeaders:   []string{"Content-Type"},
				AllowCredentials: false,
			},
			origin:          "http://evil.com",
			method:          http.MethodGet,
			wantOrigin:      "",
			wantStatusCode:  http.StatusOK,
			wantCredentials: false,
		},
		{
			name:            "preflight request",
			config:          nil,
			origin:          "http://localhost:3000",
			method:          http.MethodOptions,
			wantOrigin:      "*",
			wantStatusCode:  http.StatusNoContent,
			wantCredentials: true,
		},
		{
			name: "with exposed headers",
			config: &middleware.CORSConfig{
				AllowedOrigins:   []string{"*"},
				AllowedMethods:   []string{"GET"},
				AllowedHeaders:   []string{"Content-Type"},
				ExposedHeaders:   []string{"X-Custom-Header"},
				AllowCredentials: false,
			},
			origin:          "http://localhost:3000",
			method:          http.MethodGet,
			wantOrigin:      "*",
			wantStatusCode:  http.StatusOK,
			wantCredentials: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/test", nil)
			req.Header.Set("Origin", tt.origin)
			rr := httptest.NewRecorder()

			handler := middleware.CORSMiddleware(tt.config)(nextHandler)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatusCode, rr.Code)

			if tt.wantOrigin != "" {
				assert.Equal(t, tt.wantOrigin, rr.Header().Get("Access-Control-Allow-Origin"))
			}

			if tt.wantCredentials {
				assert.Equal(t, "true", rr.Header().Get("Access-Control-Allow-Credentials"))
			}

			// Check that methods are set
			assert.NotEmpty(t, rr.Header().Get("Access-Control-Allow-Methods"))
		})
	}
}

func TestCORSMiddleware_MultipleOrigins(t *testing.T) {
	config := &middleware.CORSConfig{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:8080"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.CORSMiddleware(config)(nextHandler)

	tests := []struct {
		name       string
		origin     string
		wantOrigin string
	}{
		{
			name:       "first allowed origin",
			origin:     "http://localhost:3000",
			wantOrigin: "http://localhost:3000",
		},
		{
			name:       "second allowed origin",
			origin:     "http://localhost:8080",
			wantOrigin: "http://localhost:8080",
		},
		{
			name:       "not allowed origin",
			origin:     "http://evil.com",
			wantOrigin: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Origin", tt.origin)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantOrigin, rr.Header().Get("Access-Control-Allow-Origin"))
		})
	}
}
