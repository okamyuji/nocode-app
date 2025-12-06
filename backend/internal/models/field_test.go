package models_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nocode-app/backend/internal/models"
)

func TestFieldOptions_Value(t *testing.T) {
	t.Run("nil options", func(t *testing.T) {
		var fo models.FieldOptions
		val, err := fo.Value()
		assert.NoError(t, err)
		assert.Nil(t, val)
	})

	t.Run("with values", func(t *testing.T) {
		fo := models.FieldOptions{
			"choices": []string{"a", "b", "c"},
			"default": "a",
		}
		val, err := fo.Value()
		assert.NoError(t, err)
		assert.NotNil(t, val)

		str, ok := val.(string)
		require.True(t, ok)
		assert.Contains(t, str, "choices")
	})
}

func TestFieldOptions_Scan(t *testing.T) {
	t.Run("nil value", func(t *testing.T) {
		var fo models.FieldOptions
		err := fo.Scan(nil)
		assert.NoError(t, err)
		assert.Nil(t, fo)
	})

	t.Run("valid json bytes", func(t *testing.T) {
		var fo models.FieldOptions
		err := fo.Scan([]byte(`{"choices":["a","b"],"default":"a"}`))
		assert.NoError(t, err)
		assert.NotNil(t, fo)
		assert.Equal(t, "a", fo["default"])
	})

	t.Run("valid json string", func(t *testing.T) {
		var fo models.FieldOptions
		err := fo.Scan(`{"choices":["x","y"],"default":"x"}`)
		assert.NoError(t, err)
		assert.NotNil(t, fo)
		assert.Equal(t, "x", fo["default"])
	})

	t.Run("invalid type", func(t *testing.T) {
		var fo models.FieldOptions
		err := fo.Scan(12345)
		assert.Error(t, err)
	})

	t.Run("invalid json", func(t *testing.T) {
		var fo models.FieldOptions
		err := fo.Scan([]byte(`{invalid json}`))
		assert.Error(t, err)
	})
}

func TestAppField_ToResponse(t *testing.T) {
	now := time.Now()
	field := &models.AppField{
		ID:           1,
		AppID:        1,
		FieldCode:    "test_field",
		FieldName:    "Test Field",
		FieldType:    "text",
		Options:      models.FieldOptions{"placeholder": "Enter text"},
		Required:     true,
		DisplayOrder: 0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	resp := field.ToResponse()

	assert.NotNil(t, resp)
	assert.Equal(t, field.ID, resp.ID)
	assert.Equal(t, field.AppID, resp.AppID)
	assert.Equal(t, field.FieldCode, resp.FieldCode)
	assert.Equal(t, field.FieldName, resp.FieldName)
	assert.Equal(t, field.FieldType, resp.FieldType)
	assert.Equal(t, field.Options, resp.Options)
	assert.Equal(t, field.Required, resp.Required)
	assert.Equal(t, field.DisplayOrder, resp.DisplayOrder)
	assert.Equal(t, field.CreatedAt, resp.CreatedAt)
	assert.Equal(t, field.UpdatedAt, resp.UpdatedAt)
}

func TestAppField_GetMySQLColumnType(t *testing.T) {
	tests := []struct {
		name      string
		fieldType string
		want      string
	}{
		{
			name:      "text field",
			fieldType: "text",
			want:      "VARCHAR(255)",
		},
		{
			name:      "textarea field",
			fieldType: "textarea",
			want:      "TEXT",
		},
		{
			name:      "number field",
			fieldType: "number",
			want:      "DECIMAL(18,4)",
		},
		{
			name:      "date field",
			fieldType: "date",
			want:      "DATE",
		},
		{
			name:      "datetime field",
			fieldType: "datetime",
			want:      "DATETIME",
		},
		{
			name:      "select field",
			fieldType: "select",
			want:      "VARCHAR(255)",
		},
		{
			name:      "multiselect field",
			fieldType: "multiselect",
			want:      "JSON",
		},
		{
			name:      "checkbox field",
			fieldType: "checkbox",
			want:      "BOOLEAN",
		},
		{
			name:      "radio field",
			fieldType: "radio",
			want:      "VARCHAR(255)",
		},
		{
			name:      "link field",
			fieldType: "link",
			want:      "VARCHAR(500)",
		},
		{
			name:      "attachment field",
			fieldType: "attachment",
			want:      "JSON",
		},
		{
			name:      "unknown field type",
			fieldType: "unknown",
			want:      "VARCHAR(255)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := &models.AppField{FieldType: tt.fieldType}
			got := field.GetMySQLColumnType()
			assert.Equal(t, tt.want, got)
		})
	}
}
