package utils_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nocode-app/backend/internal/utils"
)

func TestNewValidator(t *testing.T) {
	v := utils.NewValidator()
	assert.NotNil(t, v)
}

type ValidateTestStruct struct {
	Email    string `validate:"required,email"`
	Name     string `validate:"required,min=1,max=100"`
	Age      int    `validate:"gte=0,lte=150"`
	Password string `validate:"required,min=8"`
}

func TestValidator_Validate(t *testing.T) {
	v := utils.NewValidator()

	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name: "valid struct",
			input: ValidateTestStruct{
				Email:    "test@example.com",
				Name:     "Test User",
				Age:      25,
				Password: "password123",
			},
			wantErr: false,
		},
		{
			name: "invalid email",
			input: ValidateTestStruct{
				Email:    "not-an-email",
				Name:     "Test User",
				Age:      25,
				Password: "password123",
			},
			wantErr: true,
		},
		{
			name: "missing required field",
			input: ValidateTestStruct{
				Email:    "test@example.com",
				Name:     "",
				Age:      25,
				Password: "password123",
			},
			wantErr: true,
		},
		{
			name: "password too short",
			input: ValidateTestStruct{
				Email:    "test@example.com",
				Name:     "Test User",
				Age:      25,
				Password: "short",
			},
			wantErr: true,
		},
		{
			name: "age out of range",
			input: ValidateTestStruct{
				Email:    "test@example.com",
				Name:     "Test User",
				Age:      -1,
				Password: "password123",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidator_ValidateVar(t *testing.T) {
	v := utils.NewValidator()

	tests := []struct {
		name    string
		field   interface{}
		tag     string
		wantErr bool
	}{
		{
			name:    "valid email",
			field:   "test@example.com",
			tag:     "email",
			wantErr: false,
		},
		{
			name:    "invalid email",
			field:   "not-an-email",
			tag:     "email",
			wantErr: true,
		},
		{
			name:    "valid min length",
			field:   "hello",
			tag:     "min=3",
			wantErr: false,
		},
		{
			name:    "invalid min length",
			field:   "hi",
			tag:     "min=3",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateVar(tt.field, tt.tag)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

type ParseAndValidateStruct struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

func TestValidator_ParseAndValidate(t *testing.T) {
	v := utils.NewValidator()

	tests := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{
			name:    "valid request",
			body:    `{"name":"Test","email":"test@example.com"}`,
			wantErr: false,
		},
		{
			name:    "missing required field",
			body:    `{"email":"test@example.com"}`,
			wantErr: true,
		},
		{
			name:    "invalid email",
			body:    `{"name":"Test","email":"not-an-email"}`,
			wantErr: true,
		},
		{
			name:    "invalid JSON",
			body:    `{invalid json}`,
			wantErr: true,
		},
		{
			name:    "empty body",
			body:    ``,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(tt.body))
			var dest ParseAndValidateStruct
			err := v.ParseAndValidate(r, &dest)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "Test", dest.Name)
				assert.Equal(t, "test@example.com", dest.Email)
			}
		})
	}
}

// fieldcodeカスタムバリデーターのテスト用構造体
type FieldCodeTestStruct struct {
	FieldCode string `validate:"fieldcode"`
}

func TestValidator_FieldCodeValidation(t *testing.T) {
	v := utils.NewValidator()

	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{
			name:    "valid simple code",
			code:    "customer_name",
			wantErr: false,
		},
		{
			name:    "valid with numbers",
			code:    "field1",
			wantErr: false,
		},
		{
			name:    "valid uppercase",
			code:    "FieldName",
			wantErr: false,
		},
		{
			name:    "valid camelCase",
			code:    "fieldCode",
			wantErr: false,
		},
		{
			name:    "empty code",
			code:    "",
			wantErr: true,
		},
		{
			name:    "starts with number",
			code:    "1field",
			wantErr: true,
		},
		{
			name:    "starts with underscore",
			code:    "_field",
			wantErr: true,
		},
		{
			name:    "contains hyphen",
			code:    "field-name",
			wantErr: true,
		},
		{
			name:    "contains dot",
			code:    "field.name",
			wantErr: true,
		},
		{
			name:    "contains space",
			code:    "field name",
			wantErr: true,
		},
		{
			name:    "only underscores",
			code:    "___",
			wantErr: true,
		},
		{
			name:    "single letter",
			code:    "a",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := FieldCodeTestStruct{FieldCode: tt.code}
			err := v.Validate(input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsValidFieldCode(t *testing.T) {
	tests := []struct {
		name  string
		code  string
		valid bool
	}{
		{
			name:  "valid simple code",
			code:  "field1",
			valid: true,
		},
		{
			name:  "valid with underscore",
			code:  "field_name",
			valid: true,
		},
		{
			name:  "valid uppercase",
			code:  "FieldName",
			valid: true,
		},
		{
			name:  "valid mixed case with numbers",
			code:  "field123Name",
			valid: true,
		},
		{
			name:  "empty code",
			code:  "",
			valid: false,
		},
		{
			name:  "starts with number",
			code:  "1field",
			valid: false,
		},
		{
			name:  "contains hyphen",
			code:  "field-name",
			valid: false,
		},
		{
			name:  "contains space",
			code:  "field name",
			valid: false,
		},
		{
			name:  "too long",
			code:  strings.Repeat("a", 65),
			valid: false,
		},
		{
			name:  "exactly 64 chars",
			code:  "a" + strings.Repeat("b", 63),
			valid: true,
		},
		{
			name:  "starts with underscore",
			code:  "_field",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.IsValidFieldCode(tt.code)
			assert.Equal(t, tt.valid, got)
		})
	}
}

func TestSanitizeTableName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple name",
			input:    "users",
			expected: "users",
		},
		{
			name:     "name with underscore",
			input:    "app_data",
			expected: "app_data",
		},
		{
			name:     "name with special chars",
			input:    "app-data!@#",
			expected: "appdata",
		},
		{
			name:     "name with spaces",
			input:    "app data table",
			expected: "appdatatable",
		},
		{
			name:     "name with numbers",
			input:    "app_data_123",
			expected: "app_data_123",
		},
		{
			name:     "SQL injection attempt",
			input:    "users; DROP TABLE users;",
			expected: "usersDROPTABLEusers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.SanitizeTableName(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestGetQueryParam(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		key          string
		defaultValue string
		expected     string
	}{
		{
			name:         "existing param",
			url:          "/test?name=value",
			key:          "name",
			defaultValue: "default",
			expected:     "value",
		},
		{
			name:         "missing param",
			url:          "/test",
			key:          "name",
			defaultValue: "default",
			expected:     "default",
		},
		{
			name:         "empty param value",
			url:          "/test?name=",
			key:          "name",
			defaultValue: "default",
			expected:     "default",
		},
		{
			name:         "multiple params",
			url:          "/test?name=value&other=123",
			key:          "name",
			defaultValue: "default",
			expected:     "value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, tt.url, nil)
			got := utils.GetQueryParam(r, tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestGetQueryParamInt(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		key          string
		defaultValue int
		expected     int
	}{
		{
			name:         "existing integer param",
			url:          "/test?page=5",
			key:          "page",
			defaultValue: 1,
			expected:     5,
		},
		{
			name:         "missing param",
			url:          "/test",
			key:          "page",
			defaultValue: 1,
			expected:     1,
		},
		{
			name:         "empty param value",
			url:          "/test?page=",
			key:          "page",
			defaultValue: 1,
			expected:     1,
		},
		{
			name:         "non-integer param",
			url:          "/test?page=abc",
			key:          "page",
			defaultValue: 1,
			expected:     1,
		},
		{
			name:         "zero value",
			url:          "/test?page=0",
			key:          "page",
			defaultValue: 1,
			expected:     0,
		},
		{
			name:         "large number",
			url:          "/test?page=100",
			key:          "page",
			defaultValue: 1,
			expected:     100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, tt.url, nil)
			got := utils.GetQueryParamInt(r, tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestExtractPathParam(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		pattern  string
		expected map[string]string
	}{
		{
			name:    "single param",
			path:    "/users/123",
			pattern: "/users/:id",
			expected: map[string]string{
				"id": "123",
			},
		},
		{
			name:    "multiple params",
			path:    "/apps/1/fields/2",
			pattern: "/apps/:appId/fields/:fieldId",
			expected: map[string]string{
				"appId":   "1",
				"fieldId": "2",
			},
		},
		{
			name:     "no params",
			path:     "/users",
			pattern:  "/users",
			expected: map[string]string{},
		},
		{
			name:     "mismatched length",
			path:     "/users/123/extra",
			pattern:  "/users/:id",
			expected: map[string]string{},
		},
		{
			name:    "trailing slash",
			path:    "/users/123/",
			pattern: "/users/:id/",
			expected: map[string]string{
				"id": "123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.ExtractPathParam(tt.path, tt.pattern)
			require.Equal(t, len(tt.expected), len(got))
			for k, v := range tt.expected {
				assert.Equal(t, v, got[k])
			}
		})
	}
}
