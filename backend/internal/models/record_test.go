package models_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"nocode-app/backend/internal/models"
)

func TestNewPagination(t *testing.T) {
	tests := []struct {
		name           string
		page           int
		limit          int
		total          int64
		wantPage       int
		wantLimit      int
		wantTotal      int64
		wantTotalPages int
	}{
		{
			name:           "first page of many",
			page:           1,
			limit:          10,
			total:          100,
			wantPage:       1,
			wantLimit:      10,
			wantTotal:      100,
			wantTotalPages: 10,
		},
		{
			name:           "partial last page",
			page:           1,
			limit:          10,
			total:          25,
			wantPage:       1,
			wantLimit:      10,
			wantTotal:      25,
			wantTotalPages: 3,
		},
		{
			name:           "exactly one page",
			page:           1,
			limit:          10,
			total:          10,
			wantPage:       1,
			wantLimit:      10,
			wantTotal:      10,
			wantTotalPages: 1,
		},
		{
			name:           "no results",
			page:           1,
			limit:          10,
			total:          0,
			wantPage:       1,
			wantLimit:      10,
			wantTotal:      0,
			wantTotalPages: 0,
		},
		{
			name:           "single item",
			page:           1,
			limit:          10,
			total:          1,
			wantPage:       1,
			wantLimit:      10,
			wantTotal:      1,
			wantTotalPages: 1,
		},
		{
			name:           "limit 1",
			page:           1,
			limit:          1,
			total:          5,
			wantPage:       1,
			wantLimit:      1,
			wantTotal:      5,
			wantTotalPages: 5,
		},
		{
			name:           "middle page",
			page:           5,
			limit:          20,
			total:          200,
			wantPage:       5,
			wantLimit:      20,
			wantTotal:      200,
			wantTotalPages: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := models.NewPagination(tt.page, tt.limit, tt.total)

			assert.NotNil(t, p)
			assert.Equal(t, tt.wantPage, p.Page)
			assert.Equal(t, tt.wantLimit, p.Limit)
			assert.Equal(t, tt.wantTotal, p.Total)
			assert.Equal(t, tt.wantTotalPages, p.TotalPages)
		})
	}
}

func TestRecordData(t *testing.T) {
	data := models.RecordData{
		"field1": "value1",
		"field2": 123,
		"field3": true,
	}

	assert.Equal(t, "value1", data["field1"])
	assert.Equal(t, 123, data["field2"])
	assert.Equal(t, true, data["field3"])
}
