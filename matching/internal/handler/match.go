package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/BarkinBalci/bitaksi-case-study/matching/internal/config"
	"github.com/BarkinBalci/bitaksi-case-study/matching/internal/dto"
	"github.com/BarkinBalci/bitaksi-case-study/matching/internal/service"
)

type MatchHandler struct {
	service service.Service
	logger  *zap.Logger
}

func NewMatchHandler(service service.Service, logger *zap.Logger) *MatchHandler {
	return &MatchHandler{service: service, logger: logger}
}

func (h *MatchHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/match", h.findNearestDriver)
}

// @Summary Find nearest driver
// @Description Finds the nearest available driver for the given rider location
// @Tags match
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.MatchRequest true "Match request"
// @Success 200 {object} dto.MatchResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/match [post]
func (h *MatchHandler) findNearestDriver(c *gin.Context) {
	var req dto.MatchRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	match, err := h.service.FindNearestDriver(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrNoDriverFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Success: false,
				Error:   config.ErrNotFound,
			})
			return
		}

		h.logger.Error("failed to find nearest driver", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Success: false,
			Error:   config.ErrInternalServer,
		})
		return
	}

	c.JSON(http.StatusOK, dto.MatchResponse{
		Success: true,
		Data:    match,
	})
}
