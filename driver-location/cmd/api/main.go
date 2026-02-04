package main

import (
	"github.com/gin-gonic/gin"
	files "github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"

	"github.com/BarkinBalci/bitaksi-case-study/driver-location/docs"
	"github.com/BarkinBalci/bitaksi-case-study/driver-location/internal/handler"
	"github.com/BarkinBalci/bitaksi-case-study/driver-location/internal/service"
)

func main() {
	// TODO: Add a config to the service
	srv := service.NewService()

	// Create handlers
	locationHandler := handler.NewLocationHandler(srv)
	healthHandler := handler.NewHealthHandler()

	// Create a gin router
	r := gin.Default()

	// Set a base path for swagger
	docs.SwaggerInfo.BasePath = "/"

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(files.Handler))

	// Create router groups
	v1 := r.Group("/api/v1")
	root := r.Group("/")

	// Register routes
	locationHandler.RegisterRoutes(v1)
	healthHandler.RegisterRoutes(root)

	// Start server
	if err := r.Run(); err != nil {
		panic(err)
	}

	// TODO: Implement graceful shutdown
}
