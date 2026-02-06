package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/BarkinBalci/bitaksi-case-study/matching/internal/config"
)

// NewLogger initializes a logger.
func NewLogger(cfg config.Config) (*zap.Logger, error) {
	zapConfig := zap.NewProductionConfig()
	if cfg.Environment != "production" {
		zapConfig = zap.NewDevelopmentConfig()
	}

	logger, err := zapConfig.Build(zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	if err != nil {
		return nil, err
	}
	return logger, nil
}

// LoggerMiddleware creates Gin middleware for logging.
func LoggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		if raw != "" {
			path = path + "?" + raw
		}

		logger.Info("Request completed",
			zap.Int("status_code", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Duration("latency", latency),
			zap.String("error", c.Errors.ByType(gin.ErrorTypePrivate).String()),
		)
	}
}
