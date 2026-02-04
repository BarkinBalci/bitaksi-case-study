package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/BarkinBalci/bitaksi-case-study/driver-location/internal/config"
	"github.com/BarkinBalci/bitaksi-case-study/driver-location/internal/dto"
)

// AuthMiddleware creates Gin middleware for authentication.
func AuthMiddleware(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")

		if subtle.ConstantTimeCompare([]byte(apiKey), []byte(cfg.ApiKey)) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{
				Success: false,
				Error:   config.ErrUnauthorized,
			})
			return
		}

		c.Next()
	}
}
