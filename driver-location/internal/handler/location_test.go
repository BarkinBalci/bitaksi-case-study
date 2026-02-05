package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/BarkinBalci/bitaksi-case-study/driver-location/internal/config"
	"github.com/BarkinBalci/bitaksi-case-study/driver-location/internal/dto"
	"github.com/BarkinBalci/bitaksi-case-study/driver-location/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockService implements the service interface for testing
type MockService struct {
	mock.Mock
}

func NewMockService() *MockService {
	return &MockService{}
}

func (m *MockService) CreateDriverLocation(ctx context.Context, location *models.DriverLocation) error {
	args := m.Called(ctx, location)
	return args.Error(0)
}

func (m *MockService) CreateDriverLocationBulk(ctx context.Context, locations []*models.DriverLocation) (*models.BulkResult, error) {
	args := m.Called(ctx, locations)
	if args.Get(0) != nil {
		return args.Get(0).(*models.BulkResult), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockService) SearchDriverLocation(ctx context.Context, latitude, longitude, radius float64) ([]*models.SearchResult, error) {
	args := m.Called(ctx, latitude, longitude, radius)
	if args.Get(0) != nil {
		return args.Get(0).([]*models.SearchResult), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockService) ImportDriverLocationsFromCSV(ctx context.Context, reader io.Reader) (*models.BulkResult, error) {
	args := m.Called(ctx, reader)
	if args.Get(0) != nil {
		return args.Get(0).(*models.BulkResult), args.Error(1)
	}
	return nil, args.Error(1)
}

// Test helpers
func setupTestContext(method, path string, body io.Reader) (*gin.Context, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(method, path, body)
	return ctx, recorder
}

func marshalJSON(t *testing.T, v interface{}) []byte {
	data, err := json.Marshal(v)
	assert.NoError(t, err)
	return data
}

func unmarshalJSON(t *testing.T, data []byte, v interface{}) {
	err := json.Unmarshal(data, v)
	assert.NoError(t, err)
}

func TestLocationHandler_CreateDriverLocation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()

	tests := []struct {
		name               string
		requestBody        interface{}
		mockSetup          func(*MockService)
		expectedStatusCode int
		assertResponse     func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "success",
			requestBody: dto.CreateLocationRequest{
				Latitude:  41.0,
				Longitude: 29.0,
			},
			mockSetup: func(m *MockService) {
				m.On("CreateDriverLocation", mock.Anything, mock.MatchedBy(func(loc *models.DriverLocation) bool {
					return loc.Location.Coordinates[0] == 29.0 && loc.Location.Coordinates[1] == 41.0
				})).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			assertResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var resp dto.CreateLocationResponse
				unmarshalJSON(t, recorder.Body.Bytes(), &resp)
				assert.True(t, resp.Success)
			},
		},
		{
			name:               "bad request - invalid json",
			requestBody:        "invalid json",
			mockSetup:          func(m *MockService) {},
			expectedStatusCode: http.StatusBadRequest,
			assertResponse:     func(t *testing.T, recorder *httptest.ResponseRecorder) {},
		},
		{
			name: "service error",
			requestBody: dto.CreateLocationRequest{
				Latitude:  41.0,
				Longitude: 29.0,
			},
			mockSetup: func(m *MockService) {
				m.On("CreateDriverLocation", mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			assertResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var resp dto.ErrorResponse
				unmarshalJSON(t, recorder.Body.Bytes(), &resp)
				assert.False(t, resp.Success)
				assert.Equal(t, config.ErrInternalServer, resp.Error)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := NewMockService()
			tt.mockSetup(mockService)
			handler := NewLocationHandler(mockService, logger)

			// Prepare request body
			var body io.Reader
			if str, ok := tt.requestBody.(string); ok {
				body = bytes.NewBufferString(str)
			} else {
				body = bytes.NewBuffer(marshalJSON(t, tt.requestBody))
			}

			// Execute
			ctx, recorder := setupTestContext(http.MethodPost, "/locations", body)
			handler.createDriverLocation(ctx)

			// Assert
			assert.Equal(t, tt.expectedStatusCode, recorder.Code)
			tt.assertResponse(t, recorder)
			mockService.AssertExpectations(t)
		})
	}
}

func TestLocationHandler_CreateDriverLocationBulk(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()

	tests := []struct {
		name               string
		requestBody        interface{}
		mockSetup          func(*MockService)
		expectedStatusCode int
		assertResponse     func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "success",
			requestBody: dto.CreateLocationBulkRequest{
				Locations: []dto.CreateLocationRequest{
					{Latitude: 41.0, Longitude: 29.0},
					{Latitude: 41.1, Longitude: 29.1},
				},
			},
			mockSetup: func(m *MockService) {
				expectedResult := &models.BulkResult{
					Total:      2,
					Successful: 2,
					Failed:     0,
				}
				m.On("CreateDriverLocationBulk", mock.Anything, mock.MatchedBy(func(locs []*models.DriverLocation) bool {
					return len(locs) == 2
				})).Return(expectedResult, nil)
			},
			expectedStatusCode: http.StatusOK,
			assertResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var resp dto.CreateLocationBulkResponse
				unmarshalJSON(t, recorder.Body.Bytes(), &resp)
				assert.True(t, resp.Success)
				assert.Equal(t, 2, resp.Data.Total)
			},
		},
		{
			name:               "bad request - invalid json",
			requestBody:        "invalid json",
			mockSetup:          func(m *MockService) {},
			expectedStatusCode: http.StatusBadRequest,
			assertResponse:     func(t *testing.T, recorder *httptest.ResponseRecorder) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := NewMockService()
			tt.mockSetup(mockService)
			handler := NewLocationHandler(mockService, logger)

			// Prepare request body
			var body io.Reader
			if str, ok := tt.requestBody.(string); ok {
				body = bytes.NewBufferString(str)
			} else {
				body = bytes.NewBuffer(marshalJSON(t, tt.requestBody))
			}

			// Execute
			ctx, recorder := setupTestContext(http.MethodPost, "/locations/batch", body)
			handler.createDriverLocationBulk(ctx)

			// Assert
			assert.Equal(t, tt.expectedStatusCode, recorder.Code)
			tt.assertResponse(t, recorder)
			mockService.AssertExpectations(t)
		})
	}
}

func TestLocationHandler_SearchDriverLocation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()

	tests := []struct {
		name               string
		requestBody        dto.SearchLocationRequest
		mockSetup          func(*MockService)
		expectedStatusCode int
		assertResponse     func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "success - found results",
			requestBody: dto.SearchLocationRequest{
				Latitude:  41.0,
				Longitude: 29.0,
				Radius:    10.0,
			},
			mockSetup: func(m *MockService) {
				expectedResults := []*models.SearchResult{
					{Latitude: 41.0, Longitude: 29.0, Distance: 100},
				}
				m.On("SearchDriverLocation", mock.Anything, 41.0, 29.0, 10.0).Return(expectedResults, nil)
			},
			expectedStatusCode: http.StatusOK,
			assertResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var resp dto.SearchLocationResponse
				unmarshalJSON(t, recorder.Body.Bytes(), &resp)
				assert.True(t, resp.Success)
				assert.Len(t, resp.Data.Locations, 1)
			},
		},
		{
			name: "not found - no results",
			requestBody: dto.SearchLocationRequest{
				Latitude:  41.0,
				Longitude: 29.0,
				Radius:    10.0,
			},
			mockSetup: func(m *MockService) {
				m.On("SearchDriverLocation", mock.Anything, 41.0, 29.0, 10.0).Return([]*models.SearchResult{}, nil)
			},
			expectedStatusCode: http.StatusNotFound,
			assertResponse:     func(t *testing.T, recorder *httptest.ResponseRecorder) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := NewMockService()
			tt.mockSetup(mockService)
			handler := NewLocationHandler(mockService, logger)

			// Execute
			body := bytes.NewBuffer(marshalJSON(t, tt.requestBody))
			ctx, recorder := setupTestContext(http.MethodPost, "/locations/search", body)
			handler.searchDriverLocation(ctx)

			// Assert
			assert.Equal(t, tt.expectedStatusCode, recorder.Code)
			tt.assertResponse(t, recorder)
			mockService.AssertExpectations(t)
		})
	}
}

func TestLocationHandler_ImportDriverLocations(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()

	tests := []struct {
		name               string
		csvData            string
		mockSetup          func(*MockService)
		expectedStatusCode int
		assertResponse     func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:    "success",
			csvData: "csv data",
			mockSetup: func(m *MockService) {
				expectedResult := &models.BulkResult{
					Total:      10,
					Successful: 10,
					Failed:     0,
				}
				m.On("ImportDriverLocationsFromCSV", mock.Anything, mock.Anything).Return(expectedResult, nil)
			},
			expectedStatusCode: http.StatusOK,
			assertResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var resp dto.ImportLocationCSVResponse
				unmarshalJSON(t, recorder.Body.Bytes(), &resp)
				assert.True(t, resp.Success)
				assert.Equal(t, 10, resp.Data.Total)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := NewMockService()
			tt.mockSetup(mockService)
			handler := NewLocationHandler(mockService, logger)

			// Execute
			body := bytes.NewBufferString(tt.csvData)
			ctx, recorder := setupTestContext(http.MethodPost, "/locations/import", body)
			handler.importDriverLocations(ctx)

			// Assert
			assert.Equal(t, tt.expectedStatusCode, recorder.Code)
			tt.assertResponse(t, recorder)
			mockService.AssertExpectations(t)
		})
	}
}
