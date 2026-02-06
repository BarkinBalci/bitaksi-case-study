package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/BarkinBalci/bitaksi-case-study/matching/internal/dto"
	"github.com/BarkinBalci/bitaksi-case-study/matching/internal/service"
)

type HealthHandler struct {
	service service.Service
}

func NewHealthHandler(service service.Service) *HealthHandler {
	return &HealthHandler{service: service}
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
	if err := h.service.HealthCheck(c.Request.Context()); err != nil {
		c.JSON(http.StatusServiceUnavailable, dto.HealthCheckResponse{
			Status: "unavailable",
		})
		return
	}
	c.JSON(http.StatusOK, dto.HealthCheckResponse{
		Status: "ok",
	})
}
