package models_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nocode-app/backend/internal/models"
)

func TestViewConfig_Value(t *testing.T) {
	t.Run("nil config", func(t *testing.T) {
		var vc models.ViewConfig
		val, err := vc.Value()
		assert.NoError(t, err)
		assert.Equal(t, "{}", val)
	})

	t.Run("with values", func(t *testing.T) {
		vc := models.ViewConfig{
			"columns":    []string{"field1", "field2"},
			"sortOrder":  "asc",
			"sortColumn": "field1",
		}
		val, err := vc.Value()
		assert.NoError(t, err)
		assert.NotNil(t, val)

		str, ok := val.(string)
		require.True(t, ok)
		assert.Contains(t, str, "columns")
	})
}

func TestViewConfig_Scan(t *testing.T) {
	t.Run("nil value", func(t *testing.T) {
		var vc models.ViewConfig
		err := vc.Scan(nil)
		assert.NoError(t, err)
		assert.Nil(t, vc)
	})

	t.Run("valid json bytes", func(t *testing.T) {
		var vc models.ViewConfig
		err := vc.Scan([]byte(`{"columns":["field1","field2"],"sortOrder":"asc"}`))
		assert.NoError(t, err)
		assert.NotNil(t, vc)
		assert.Equal(t, "asc", vc["sortOrder"])
	})

	t.Run("invalid type", func(t *testing.T) {
		var vc models.ViewConfig
		err := vc.Scan("not bytes")
		assert.Error(t, err)
	})

	t.Run("invalid json", func(t *testing.T) {
		var vc models.ViewConfig
		err := vc.Scan([]byte(`{invalid json}`))
		assert.Error(t, err)
	})
}

func TestAppView_ToResponse(t *testing.T) {
	now := time.Now()
	view := &models.AppView{
		ID:        1,
		AppID:     1,
		Name:      "Test View",
		ViewType:  "table",
		Config:    models.ViewConfig{"columns": []string{"field1"}},
		IsDefault: true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	resp := view.ToResponse()

	assert.NotNil(t, resp)
	assert.Equal(t, view.ID, resp.ID)
	assert.Equal(t, view.AppID, resp.AppID)
	assert.Equal(t, view.Name, resp.Name)
	assert.Equal(t, view.ViewType, resp.ViewType)
	assert.Equal(t, view.Config, resp.Config)
	assert.Equal(t, view.IsDefault, resp.IsDefault)
	assert.Equal(t, view.CreatedAt, resp.CreatedAt)
	assert.Equal(t, view.UpdatedAt, resp.UpdatedAt)
}

func TestViewType_Constants(t *testing.T) {
	assert.Equal(t, models.ViewType("table"), models.ViewTypeTable)
	assert.Equal(t, models.ViewType("list"), models.ViewTypeList)
	assert.Equal(t, models.ViewType("calendar"), models.ViewTypeCalendar)
	assert.Equal(t, models.ViewType("chart"), models.ViewTypeChart)
}

func TestFieldType_Constants(t *testing.T) {
	assert.Equal(t, models.FieldType("text"), models.FieldTypeText)
	assert.Equal(t, models.FieldType("textarea"), models.FieldTypeTextArea)
	assert.Equal(t, models.FieldType("number"), models.FieldTypeNumber)
	assert.Equal(t, models.FieldType("date"), models.FieldTypeDate)
	assert.Equal(t, models.FieldType("datetime"), models.FieldTypeDateTime)
	assert.Equal(t, models.FieldType("select"), models.FieldTypeSelect)
	assert.Equal(t, models.FieldType("multiselect"), models.FieldTypeMultiSelect)
	assert.Equal(t, models.FieldType("checkbox"), models.FieldTypeCheckbox)
	assert.Equal(t, models.FieldType("radio"), models.FieldTypeRadio)
	assert.Equal(t, models.FieldType("link"), models.FieldTypeLink)
	assert.Equal(t, models.FieldType("attachment"), models.FieldTypeAttachment)
}
