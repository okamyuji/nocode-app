package models_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nocode-app/backend/internal/models"
)

func TestApp_ToResponse(t *testing.T) {
	now := time.Now()

	t.Run("without fields", func(t *testing.T) {
		app := &models.App{
			ID:          1,
			Name:        "Test App",
			Description: "A test application",
			TableName:   "app_data_1",
			Icon:        "default",
			CreatedBy:   1,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		resp := app.ToResponse()

		assert.NotNil(t, resp)
		assert.Equal(t, app.ID, resp.ID)
		assert.Equal(t, app.Name, resp.Name)
		assert.Equal(t, app.Description, resp.Description)
		assert.Equal(t, app.TableName, resp.TableName)
		assert.Equal(t, app.Icon, resp.Icon)
		assert.Equal(t, app.CreatedBy, resp.CreatedBy)
		assert.Equal(t, app.CreatedAt, resp.CreatedAt)
		assert.Equal(t, app.UpdatedAt, resp.UpdatedAt)
		assert.Empty(t, resp.Fields)
	})

	t.Run("with fields", func(t *testing.T) {
		app := &models.App{
			ID:          1,
			Name:        "Test App",
			Description: "A test application",
			TableName:   "app_data_1",
			Icon:        "default",
			CreatedBy:   1,
			CreatedAt:   now,
			UpdatedAt:   now,
			Fields: []models.AppField{
				{
					ID:           1,
					AppID:        1,
					FieldCode:    "field1",
					FieldName:    "Field 1",
					FieldType:    "text",
					Required:     true,
					DisplayOrder: 0,
					CreatedAt:    now,
					UpdatedAt:    now,
				},
				{
					ID:           2,
					AppID:        1,
					FieldCode:    "field2",
					FieldName:    "Field 2",
					FieldType:    "number",
					Required:     false,
					DisplayOrder: 1,
					CreatedAt:    now,
					UpdatedAt:    now,
				},
			},
		}

		resp := app.ToResponse()

		assert.NotNil(t, resp)
		require.Len(t, resp.Fields, 2)
		assert.Equal(t, "field1", resp.Fields[0].FieldCode)
		assert.Equal(t, "field2", resp.Fields[1].FieldCode)
	})
}
