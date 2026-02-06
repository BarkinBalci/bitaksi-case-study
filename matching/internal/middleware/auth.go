package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/BarkinBalci/bitaksi-case-study/matching/internal/config"
	"github.com/BarkinBalci/bitaksi-case-study/matching/internal/dto"
)

// JWTAuthMiddleware creates Gin middleware for JWT authentication.
func JWTAuthMiddleware(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{
				Success: false,
				Error:   config.ErrUnauthorized,
			})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{
				Success: false,
				Error:   config.ErrUnauthorized,
			})
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(cfg.JWTSecret), nil
		}, jwt.WithExpirationRequired())

		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{
					Success: false,
					Error:   config.ErrTokenExpired,
				})
				return
			}

			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{
				Success: false,
				Error:   config.ErrUnauthorized,
			})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{
				Success: false,
				Error:   config.ErrUnauthorized,
			})
			return
		}

		authenticatedVal, exists := claims["authenticated"]
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{
				Success: false,
				Error:   config.ErrUnauthorized,
			})
			return
		}

		authenticated, ok := authenticatedVal.(bool)
		if !ok || !authenticated {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{
				Success: false,
				Error:   config.ErrUnauthorized,
			})
			return
		}

		if userID, exists := claims["user_id"]; exists {
			c.Set("user_id", userID)
		}
		c.Set("authenticated", authenticated)

		c.Next()
	}
}
