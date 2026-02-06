package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"

	"github.com/BarkinBalci/bitaksi-case-study/matching/internal/config"
	"github.com/BarkinBalci/bitaksi-case-study/matching/internal/dto"
)

const testJWTSecret = "test-secret-key"

// Helper function to create a valid JWT token
func createTestToken(secret string, claims jwt.MapClaims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

func TestJWTAuthMiddleware(t *testing.T) {
	cfg := config.Config{
		JWTSecret: testJWTSecret,
	}

	tests := []struct {
		name               string
		setupAuth          func() string
		expectedStatusCode int
		expectedError      string
		shouldCallNext     bool
		assertContext      func(*testing.T, *gin.Context)
	}{
		{
			name: "success - valid token",
			setupAuth: func() string {
				claims := jwt.MapClaims{
					"authenticated": true,
					"user_id":       "user123",
					"exp":           time.Now().Add(time.Hour).Unix(),
				}
				return "Bearer " + createTestToken(testJWTSecret, claims)
			},
			expectedStatusCode: http.StatusOK,
			shouldCallNext:     true,
			assertContext: func(t *testing.T, ctx *gin.Context) {
				authenticated, exists := ctx.Get("authenticated")
				assert.True(t, exists)
				assert.Equal(t, true, authenticated)

				userID, exists := ctx.Get("user_id")
				assert.True(t, exists)
				assert.Equal(t, "user123", userID)
			},
		},
		{
			name: "failure - missing authorization header",
			setupAuth: func() string {
				return ""
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedError:      config.ErrUnauthorized,
			shouldCallNext:     false,
		},
		{
			name: "failure - expired token",
			setupAuth: func() string {
				claims := jwt.MapClaims{
					"authenticated": true,
					"user_id":       "user123",
					"exp":           time.Now().Add(-time.Hour).Unix(),
				}
				return "Bearer " + createTestToken(testJWTSecret, claims)
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedError:      config.ErrTokenExpired,
			shouldCallNext:     false,
		},
		{
			name: "failure - wrong secret",
			setupAuth: func() string {
				claims := jwt.MapClaims{
					"authenticated": true,
					"exp":           time.Now().Add(time.Hour).Unix(),
				}
				return "Bearer " + createTestToken("wrong-secret", claims)
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedError:      config.ErrUnauthorized,
			shouldCallNext:     false,
		},
		{
			name: "failure - missing authenticated claim",
			setupAuth: func() string {
				claims := jwt.MapClaims{
					"user_id": "user123",
					"exp":     time.Now().Add(time.Hour).Unix(),
				}
				return "Bearer " + createTestToken(testJWTSecret, claims)
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedError:      config.ErrUnauthorized,
			shouldCallNext:     false,
		},
		{
			name: "failure - authenticated claim is false",
			setupAuth: func() string {
				claims := jwt.MapClaims{
					"authenticated": false,
					"user_id":       "user123",
					"exp":           time.Now().Add(time.Hour).Unix(),
				}
				return "Bearer " + createTestToken(testJWTSecret, claims)
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedError:      config.ErrUnauthorized,
			shouldCallNext:     false,
		},
		{
			name: "failure - missing expiration",
			setupAuth: func() string {
				claims := jwt.MapClaims{
					"authenticated": true,
					"user_id":       "user123",
				}
				return "Bearer " + createTestToken(testJWTSecret, claims)
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedError:      config.ErrUnauthorized,
			shouldCallNext:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			authHeader := tt.setupAuth()
			router := gin.New()
			nextCalled := false
			router.Use(JWTAuthMiddleware(cfg))
			router.GET("/test", func(c *gin.Context) {
				nextCalled = true
				if tt.assertContext != nil {
					tt.assertContext(t, c)
				}
				c.Status(http.StatusOK)
			})

			// Execute
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if authHeader != "" {
				req.Header.Set("Authorization", authHeader)
			}
			router.ServeHTTP(recorder, req)

			// Assert
			assert.Equal(t, tt.expectedStatusCode, recorder.Code)
			assert.Equal(t, tt.shouldCallNext, nextCalled)

			if tt.expectedError != "" {
				var resp dto.ErrorResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.False(t, resp.Success)
				assert.Equal(t, tt.expectedError, resp.Error)
			}
		})
	}
}

func TestJWTAuthMiddleware_IntegrationWithGinRouter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{
		JWTSecret: testJWTSecret,
	}

	tests := []struct {
		name               string
		setupAuth          func() string
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name: "success - protected route accessible with valid token",
			setupAuth: func() string {
				claims := jwt.MapClaims{
					"authenticated": true,
					"user_id":       "user123",
					"exp":           time.Now().Add(time.Hour).Unix(),
				}
				return "Bearer " + createTestToken(testJWTSecret, claims)
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       `{"message":"success","user_id":"user123"}`,
		},
		{
			name: "failure - protected route blocked without token",
			setupAuth: func() string {
				return ""
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedBody:       fmt.Sprintf(`{"success":false,"error":"%s"}`, config.ErrUnauthorized),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup router
			router := gin.New()
			router.Use(JWTAuthMiddleware(cfg))
			router.GET("/protected", func(c *gin.Context) {
				userID, _ := c.Get("user_id")
				c.JSON(http.StatusOK, gin.H{
					"message": "success",
					"user_id": userID,
				})
			})

			// Execute
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if authHeader := tt.setupAuth(); authHeader != "" {
				req.Header.Set("Authorization", authHeader)
			}

			router.ServeHTTP(recorder, req)

			// Assert
			assert.Equal(t, tt.expectedStatusCode, recorder.Code)
			assert.JSONEq(t, tt.expectedBody, recorder.Body.String())
		})
	}
}
