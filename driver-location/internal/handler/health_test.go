package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func (m *MockService) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestHealthHandler_HealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name               string
		mockSetup          func(*MockService)
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name: "success",
			mockSetup: func(m *MockService) {
				m.On("HealthCheck", mock.Anything).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       `{"status": "ok"}`,
		},
		{
			name: "failure",
			mockSetup: func(m *MockService) {
				m.On("HealthCheck", mock.Anything).Return(errors.New("db error"))
			},
			expectedStatusCode: http.StatusServiceUnavailable,
			expectedBody:       `{"status": "unavailable"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := NewMockService()
			tt.mockSetup(mockService)
			handler := NewHealthHandler(mockService)

			// Execute
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(http.MethodGet, "/health", nil)
			handler.healthCheck(ctx)

			// Assert
			assert.Equal(t, tt.expectedStatusCode, recorder.Code)
			assert.JSONEq(t, tt.expectedBody, recorder.Body.String())
			mockService.AssertExpectations(t)
		})
	}
}
