package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/BarkinBalci/bitaksi-case-study/driver-location/internal/models"
)

// MockRepository is a mock implementation of repository.DriverLocationRepository
type MockRepository struct {
	mock.Mock
}

func NewMockRepository() *MockRepository {
	return &MockRepository{}
}

func (m *MockRepository) Create(ctx context.Context, location *models.DriverLocation) error {
	args := m.Called(ctx, location)
	return args.Error(0)
}

func (m *MockRepository) CreateMany(ctx context.Context, locations []*models.DriverLocation) (int, error) {
	args := m.Called(ctx, locations)
	return args.Int(0), args.Error(1)
}

func (m *MockRepository) Search(ctx context.Context, longitude, latitude, radius float64) ([]*models.SearchResult, error) {
	args := m.Called(ctx, longitude, latitude, radius)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.SearchResult), args.Error(1)
}

func (m *MockRepository) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Test helpers
func setupTest() (*MockRepository, Service, context.Context) {
	mockRepo := NewMockRepository()
	logger := zap.NewNop()
	svc := NewService(mockRepo, logger)
	ctx := context.Background()
	return mockRepo, svc, ctx
}

func TestHealthCheck(t *testing.T) {
	tests := []struct {
		name          string
		mockSetup     func(*MockRepository, context.Context)
		expectedError bool
		errorContains string
	}{
		{
			name: "success",
			mockSetup: func(m *MockRepository, ctx context.Context) {
				m.On("Ping", ctx).Return(nil).Once()
			},
			expectedError: false,
		},
		{
			name: "failure - db down",
			mockSetup: func(m *MockRepository, ctx context.Context) {
				m.On("Ping", ctx).Return(errors.New("db down")).Once()
			},
			expectedError: true,
			errorContains: "health check failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo, svc, ctx := setupTest()
			tt.mockSetup(mockRepo, ctx)

			// Execute
			err := svc.HealthCheck(ctx)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCreateDriverLocation(t *testing.T) {
	tests := []struct {
		name          string
		location      *models.DriverLocation
		mockSetup     func(*MockRepository, context.Context, *models.DriverLocation)
		expectedError bool
		errorContains string
	}{
		{
			name:     "success",
			location: models.NewDriverLocation(40.0, 29.0),
			mockSetup: func(m *MockRepository, ctx context.Context, loc *models.DriverLocation) {
				m.On("Create", ctx, loc).Return(nil).Once()
			},
			expectedError: false,
		},
		{
			name:     "failure - db error",
			location: models.NewDriverLocation(40.0, 29.0),
			mockSetup: func(m *MockRepository, ctx context.Context, loc *models.DriverLocation) {
				m.On("Create", ctx, loc).Return(errors.New("db error")).Once()
			},
			expectedError: true,
			errorContains: "failed to create driver location",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo, svc, ctx := setupTest()
			tt.mockSetup(mockRepo, ctx, tt.location)

			// Execute
			err := svc.CreateDriverLocation(ctx, tt.location)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCreateDriverLocationBulk(t *testing.T) {
	testLocations := []*models.DriverLocation{
		models.NewDriverLocation(40.0, 29.0),
		models.NewDriverLocation(41.0, 30.0),
	}

	tests := []struct {
		name               string
		locations          []*models.DriverLocation
		mockSetup          func(*MockRepository, context.Context, []*models.DriverLocation)
		expectedError      bool
		expectedTotal      int
		expectedSuccessful int
		expectedFailed     int
	}{
		{
			name:      "success - all inserted",
			locations: testLocations,
			mockSetup: func(m *MockRepository, ctx context.Context, locs []*models.DriverLocation) {
				m.On("CreateMany", ctx, locs).Return(len(locs), nil).Once()
			},
			expectedError:      false,
			expectedTotal:      2,
			expectedSuccessful: 2,
			expectedFailed:     0,
		},
		{
			name:      "partial success - one failed",
			locations: testLocations,
			mockSetup: func(m *MockRepository, ctx context.Context, locs []*models.DriverLocation) {
				m.On("CreateMany", ctx, locs).Return(1, nil).Once()
			},
			expectedError:      false,
			expectedTotal:      2,
			expectedSuccessful: 1,
			expectedFailed:     1,
		},
		{
			name:      "failure - db error",
			locations: testLocations,
			mockSetup: func(m *MockRepository, ctx context.Context, locs []*models.DriverLocation) {
				m.On("CreateMany", ctx, locs).Return(0, errors.New("db error")).Once()
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo, svc, ctx := setupTest()
			tt.mockSetup(mockRepo, ctx, tt.locations)

			// Execute
			result, err := svc.CreateDriverLocationBulk(ctx, tt.locations)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedTotal, result.Total)
				assert.Equal(t, tt.expectedSuccessful, result.Successful)
				assert.Equal(t, tt.expectedFailed, result.Failed)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSearchDriverLocation(t *testing.T) {
	const (
		testLatitude  = 40.0
		testLongitude = 29.0
		testRadius    = 1000.0
	)

	tests := []struct {
		name            string
		latitude        float64
		longitude       float64
		radius          float64
		mockSetup       func(*MockRepository, context.Context)
		expectedError   bool
		expectedResults []*models.SearchResult
	}{
		{
			name:      "success - results found",
			latitude:  testLatitude,
			longitude: testLongitude,
			radius:    testRadius,
			mockSetup: func(m *MockRepository, ctx context.Context) {
				expectedResults := []*models.SearchResult{
					{Latitude: 40.1, Longitude: 29.1, Distance: 500},
				}
				m.On("Search", ctx, testLongitude, testLatitude, testRadius).Return(expectedResults, nil).Once()
			},
			expectedError: false,
			expectedResults: []*models.SearchResult{
				{Latitude: 40.1, Longitude: 29.1, Distance: 500},
			},
		},
		{
			name:      "failure - db error",
			latitude:  testLatitude,
			longitude: testLongitude,
			radius:    testRadius,
			mockSetup: func(m *MockRepository, ctx context.Context) {
				m.On("Search", ctx, testLongitude, testLatitude, testRadius).Return(nil, errors.New("db error")).Once()
			},
			expectedError:   true,
			expectedResults: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo, svc, ctx := setupTest()
			tt.mockSetup(mockRepo, ctx)

			// Execute
			results, err := svc.SearchDriverLocation(ctx, tt.latitude, tt.longitude, tt.radius)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, results)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResults, results)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestImportDriverLocationsFromCSV(t *testing.T) {
	tests := []struct {
		name               string
		csvContent         string
		mockSetup          func(*MockRepository, context.Context)
		expectedError      bool
		expectedTotal      int
		expectedSuccessful int
	}{
		{
			name: "success - valid csv",
			csvContent: `lat,lon
40.0,29.0
41.0,30.0`,
			mockSetup: func(m *MockRepository, ctx context.Context) {
				m.On("CreateMany", ctx, mock.MatchedBy(func(locs []*models.DriverLocation) bool {
					return len(locs) == 2 && locs[0].Location.Coordinates[1] == 40.0
				})).Return(2, nil).Once()
			},
			expectedError:      false,
			expectedTotal:      2,
			expectedSuccessful: 2,
		},
		{
			name: "failure - invalid latitude",
			csvContent: `lat,lon
invalid,29.0`,
			mockSetup:     func(m *MockRepository, ctx context.Context) {},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo, svc, ctx := setupTest()
			tt.mockSetup(mockRepo, ctx)
			reader := strings.NewReader(tt.csvContent)

			// Execute
			result, err := svc.ImportDriverLocationsFromCSV(ctx, reader)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedTotal, result.Total)
				assert.Equal(t, tt.expectedSuccessful, result.Successful)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}
