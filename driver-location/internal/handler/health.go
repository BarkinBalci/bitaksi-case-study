package handler

import (
	"net/http"

	"github.com/BarkinBalci/bitaksi-case-study/driver-location/internal/dto"
	"github.com/gin-gonic/gin"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/health", h.healthCheck)
}

// @Summary Check if the service is healthy
// @Description Check if the service is healthy
// @Tags health
// @Produce json
// @Success 200 {object} dto.HealthCheckResponse
// @Router /health [get]
func (h *HealthHandler) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, dto.HealthCheckResponse{
		Status: "ok",
	})
}
