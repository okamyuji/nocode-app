package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"nocode-app/backend/internal/handlers"
	"nocode-app/backend/internal/models"
	"nocode-app/backend/internal/repositories"
	"nocode-app/backend/internal/testhelpers/mocks"
	"nocode-app/backend/internal/utils"
)

// 範囲外の page と limit は、user.go と同じく page=1、limit=20 に直してから service へ渡す。
var paginationBoundsCases = []struct {
	name      string
	query     string
	wantPage  int
	wantLimit int
}{
	{"limit が 0", "?limit=0", 1, 20},
	{"limit が上限 100 を超える", "?limit=101", 1, 20},
	{"limit が上限ちょうど", "?limit=100", 1, 100},
	{"limit が下限ちょうど", "?limit=1", 1, 1},
	{"limit が int を溢れる", "?limit=99999999999999999999", 1, 20},
	{"page が 0", "?page=0&limit=20", 1, 20},
	{"page が数字でない", "?page=-3&limit=5", 1, 5},
	{"page が正しい", "?page=3&limit=5", 3, 5},
	{"page が上限ちょうど", "?page=1000000&limit=100", 1000000, 100},
	{"page が上限を超える", "?page=1000001&limit=100", 1000000, 100},
	{"page が int の最大値", "?page=9223372036854775807&limit=100", 1000000, 100},
}

func TestAppHandler_List_PaginationBounds(t *testing.T) {
	for _, tc := range paginationBoundsCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := new(mocks.MockAppService)
			handler := handlers.NewAppHandler(mockService, utils.NewValidator())
			mockService.On("GetApps", mock.Anything, tc.wantPage, tc.wantLimit).
				Return(&models.AppListResponse{Pagination: models.NewPagination(tc.wantPage, tc.wantLimit, 0)}, nil)

			rr := httptest.NewRecorder()
			handler.List(rr, httptest.NewRequest(http.MethodGet, "/api/v1/apps"+tc.query, nil))

			assert.Equal(t, http.StatusOK, rr.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestDataSourceHandler_List_PaginationBounds(t *testing.T) {
	for _, tc := range paginationBoundsCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := new(mocks.MockDataSourceService)
			handler := createDataSourceHandler(mockService)
			mockService.On("GetDataSources", mock.Anything, tc.wantPage, tc.wantLimit).
				Return(&models.DataSourceListResponse{Pagination: models.NewPagination(tc.wantPage, tc.wantLimit, 0)}, nil)

			rr := httptest.NewRecorder()
			handler.List(rr, httptest.NewRequest(http.MethodGet, "/api/v1/datasources"+tc.query, nil))

			assert.Equal(t, http.StatusOK, rr.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestRecordHandler_List_PaginationBounds(t *testing.T) {
	for _, tc := range paginationBoundsCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := new(mocks.MockRecordService)
			handler := handlers.NewRecordHandler(mockService, utils.NewValidator())
			mockService.On("GetRecords", mock.Anything, uint64(1), mock.MatchedBy(func(o repositories.RecordQueryOptions) bool {
				return o.Page == tc.wantPage && o.Limit == tc.wantLimit
			})).Return(&models.RecordListResponse{Pagination: models.NewPagination(tc.wantPage, tc.wantLimit, 0)}, nil)

			rr := httptest.NewRecorder()
			handler.List(rr, httptest.NewRequest(http.MethodGet, "/api/v1/apps/1/records"+tc.query, nil))

			assert.Equal(t, http.StatusOK, rr.Code)
			mockService.AssertExpectations(t)
		})
	}
}
