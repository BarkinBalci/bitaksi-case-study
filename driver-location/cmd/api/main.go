package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	files "github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"go.uber.org/zap"

	"github.com/BarkinBalci/bitaksi-case-study/driver-location/docs"
	"github.com/BarkinBalci/bitaksi-case-study/driver-location/internal/config"
	"github.com/BarkinBalci/bitaksi-case-study/driver-location/internal/handler"
	"github.com/BarkinBalci/bitaksi-case-study/driver-location/internal/middleware"
	"github.com/BarkinBalci/bitaksi-case-study/driver-location/internal/service"
)

// @securityDefinitions.apiKey ApiKeyAuth
// @in header
// @name X-API-Key
func main() {
	cfg := config.LoadConfig()
	srv := service.NewService()

	// Initialize logger
	logger, err := middleware.NewLogger(*cfg)
	if err != nil {
		log.Fatal("failed to initialize logger: ", err)
	}
	defer func() {
		_ = logger.Sync()
	}()

	// Create handlers
	locationHandler := handler.NewLocationHandler(srv, logger)
	healthHandler := handler.NewHealthHandler()

	// Create a gin router and attach middlewares
	router := gin.New()
	router.Use(middleware.LoggerMiddleware(logger))
	router.Use(gin.Recovery())

	if cfg.SwaggerEnabled {
		// Serve Swagger documentation
		docs.SwaggerInfo.BasePath = "/"
		router.GET("/swagger/*any", ginSwagger.WrapHandler(files.Handler))
	}

	// Register public routes
	root := router.Group("/")
	healthHandler.RegisterRoutes(root)

	// Register private routes
	v1 := router.Group("/api/v1")
	v1.Use(middleware.AuthMiddleware(*cfg))
	locationHandler.RegisterRoutes(v1)

	// Create http server
	httpServer := &http.Server{
		Addr:              ":" + cfg.ServerPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Start the server in a goroutine to not block graceful shutdown handling
	go func() {
		err := httpServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("failed to start server", zap.Error(err))
		}
	}()

	// Wait for an interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down server...")

	// Inform the server it has 5 seconds to finish the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Fatal("server forced to shutdown", zap.Error(err))
	}
}
