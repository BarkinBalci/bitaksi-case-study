package handler

import (
	"net/http"

	"github.com/BarkinBalci/bitaksi-case-study/driver-location/internal/config"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/BarkinBalci/bitaksi-case-study/driver-location/internal/dto"
	"github.com/BarkinBalci/bitaksi-case-study/driver-location/internal/service"
)

type LocationHandler struct {
	service service.Service
	logger  *zap.Logger
}

func NewLocationHandler(service service.Service, logger *zap.Logger) *LocationHandler {
	return &LocationHandler{service: service, logger: logger}
}

func (h *LocationHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/locations", h.createDriverLocation)
	r.POST("/locations/batch", h.createDriverLocationBulk)
	r.POST("/locations/search", h.searchDriverLocation)
}

// @Summary Create a new driver location
// @Description Creates a new driver location
// @Tags locations
// @Accept json
// @Produce json
// @Param request body dto.CreateLocationRequest true "Create location request"
// @Success 200 {object} dto.LocationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security ApiKeyAuth
// @Router /api/v1/locations [post]
func (h *LocationHandler) createDriverLocation(c *gin.Context) {
	var req dto.CreateLocationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON for createDriverLocation",
			zap.Error(err),
			zap.String("path", c.Request.URL.Path),
		)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	if err := h.service.CreateDriverLocation(c.Request.Context(), &req); err != nil {
		h.logger.Error("Failed to create driver location",
			zap.Error(err),
			zap.String("driver_id", req.DriverID),
			zap.String("path", c.Request.URL.Path),
		)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Success: false,
			Error:   config.ErrInternalServer,
		})
		return
	}

	c.JSON(http.StatusOK, dto.LocationResponse{
		Success: true,
		Data: dto.LocationData{
			DriverID: req.DriverID,
			Message:  "Location updated successfully",
		},
	})
}

// @Summary Create driver locations in bulk
// @Description Creates multiple driver locations at once
// @Tags locations
// @Accept json
// @Produce json
// @Param request body dto.CreateLocationBulkRequest true "Create bulk location request"
// @Success 200 {object} dto.BulkLocationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security ApiKeyAuth
// @Router /api/v1/locations/batch [post]
func (h *LocationHandler) createDriverLocationBulk(c *gin.Context) {
	var req dto.CreateLocationBulkRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON for createDriverLocationBulk",
			zap.Error(err),
			zap.String("path", c.Request.URL.Path),
		)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	result, err := h.service.CreateDriverLocationBulk(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create driver locations in bulk",
			zap.Error(err),
			zap.String("path", c.Request.URL.Path),
		)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Success: false,
			Error:   config.ErrInternalServer,
		})
		return
	}

	c.JSON(http.StatusOK, dto.BulkLocationResponse{
		Success: true,
		Data:    *result,
	})
}

// @Summary Search for driver locations
// @Description Searches for driver locations based on a GeoJSON point and radius
// @Tags locations
// @Accept json
// @Produce json
// @Param request body dto.SearchLocationRequest true "Search location request"
// @Success 200 {object} dto.SearchLocationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security ApiKeyAuth
// @Router /api/v1/locations/search [post]
func (h *LocationHandler) searchDriverLocation(c *gin.Context) {
	var req dto.SearchLocationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON for searchDriverLocation",
			zap.Error(err),
			zap.String("path", c.Request.URL.Path),
		)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	drivers, err := h.service.SearchDriverLocation(c.Request.Context(), &req)
	if err != nil {
		// TODO: Handle 404 separately
		h.logger.Error("Failed to search driver locations",
			zap.Error(err),
			zap.String("path", c.Request.URL.Path),
		)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Success: false,
			Error:   config.ErrInternalServer,
		})
		return
	}

	c.JSON(http.StatusOK, dto.SearchLocationResponse{
		Success: true,
		Data: dto.SearchLocationData{
			Drivers: drivers,
			Total:   len(drivers),
		},
	})
}
