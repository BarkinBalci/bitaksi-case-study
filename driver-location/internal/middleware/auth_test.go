package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/BarkinBalci/bitaksi-case-study/driver-location/internal/config"
	"github.com/BarkinBalci/bitaksi-case-study/driver-location/internal/dto"
)

const testAPIKey = "test-api-key-123"

func TestAuthMiddleware(t *testing.T) {
	cfg := config.Config{
		ApiKey: testAPIKey,
	}

	tests := []struct {
		name               string
		setupAuth          func() string
		expectedStatusCode int
		expectedError      string
		shouldCallNext     bool
	}{
		{
			name: "success - valid API key",
			setupAuth: func() string {
				return testAPIKey
			},
			expectedStatusCode: http.StatusOK,
			shouldCallNext:     true,
		},
		{
			name: "failure - missing API key header",
			setupAuth: func() string {
				return ""
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedError:      config.ErrUnauthorized,
			shouldCallNext:     false,
		},
		{
			name: "failure - wrong API key",
			setupAuth: func() string {
				return "wrong-api-key"
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedError:      config.ErrUnauthorized,
			shouldCallNext:     false,
		},
		{
			name: "failure - empty API key",
			setupAuth: func() string {
				return ""
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedError:      config.ErrUnauthorized,
			shouldCallNext:     false,
		},
		{
			name: "failure - API key with extra spaces",
			setupAuth: func() string {
				return testAPIKey + " "
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedError:      config.ErrUnauthorized,
			shouldCallNext:     false,
		},
		{
			name: "failure - API key with leading spaces",
			setupAuth: func() string {
				return " " + testAPIKey
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedError:      config.ErrUnauthorized,
			shouldCallNext:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			apiKey := tt.setupAuth()
			router := gin.New()
			router.Use(AuthMiddleware(cfg))

			// Execute
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if apiKey != "" {
				req.Header.Set("X-API-Key", apiKey)
			}
			router.ServeHTTP(recorder, req)

			// Assert
			var resp dto.ErrorResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &resp)
			if err == nil {
				assert.False(t, resp.Success)
				assert.Equal(t, tt.expectedError, resp.Error)
			}
		})
	}
}

func TestAuthMiddleware_IntegrationWithGinRouter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{
		ApiKey: testAPIKey,
	}

	tests := []struct {
		name               string
		setupAuth          func() string
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name: "success - protected route accessible with valid API key",
			setupAuth: func() string {
				return testAPIKey
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       `{"message":"success"}`,
		},
		{
			name: "failure - protected route blocked without API key",
			setupAuth: func() string {
				return ""
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedBody:       fmt.Sprintf(`{"success":false,"error":"%s"}`, config.ErrUnauthorized),
		},
		{
			name: "failure - protected route blocked with invalid API key",
			setupAuth: func() string {
				return "invalid-key"
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedBody:       fmt.Sprintf(`{"success":false,"error":"%s"}`, config.ErrUnauthorized),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			router := gin.New()
			router.Use(AuthMiddleware(cfg))
			router.GET("/protected", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"message": "success",
				})
			})

			// Execute
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if apiKey := tt.setupAuth(); apiKey != "" {
				req.Header.Set("X-API-Key", apiKey)
			}

			router.ServeHTTP(recorder, req)

			// Assert
			assert.Equal(t, tt.expectedStatusCode, recorder.Code)
			assert.JSONEq(t, tt.expectedBody, recorder.Body.String())
		})
	}
}
