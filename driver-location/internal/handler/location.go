package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/BarkinBalci/bitaksi-case-study/driver-location/internal/config"
	"github.com/BarkinBalci/bitaksi-case-study/driver-location/internal/dto"
	"github.com/BarkinBalci/bitaksi-case-study/driver-location/internal/models"
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
	r.POST("/locations/import", h.importDriverLocations)
}

// @Summary Create a new driver location
// @Description Creates a new driver location
// @Tags locations
// @Accept json
// @Produce json
// @Param request body dto.CreateLocationRequest true "Create location request"
// @Success 200 {object} dto.CreateLocationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security ApiKeyAuth
// @Router /api/v1/locations [post]
func (h *LocationHandler) createDriverLocation(c *gin.Context) {
	var req dto.CreateLocationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	locationModel := models.NewDriverLocation(req.Latitude, req.Longitude)
	if err := h.service.CreateDriverLocation(c.Request.Context(), locationModel); err != nil {
		h.logger.Error("Failed to create driver location", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Success: false,
			Error:   config.ErrInternalServer,
		})
		return
	}

	c.JSON(http.StatusOK, dto.CreateLocationResponse{
		Success: true,
		Data: dto.CreateLocationData{
			Message: "Location updated successfully",
		},
	})
}

// @Summary Create driver locations in bulk
// @Description Creates multiple driver locations at once
// @Tags locations
// @Accept json
// @Produce json
// @Param request body dto.CreateLocationBulkRequest true "Create bulk location request"
// @Success 200 {object} dto.CreateLocationBulkResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security ApiKeyAuth
// @Router /api/v1/locations/batch [post]
func (h *LocationHandler) createDriverLocationBulk(c *gin.Context) {
	var req dto.CreateLocationBulkRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	locationModels := make([]*models.DriverLocation, len(req.Locations))
	for i, dtoReq := range req.Locations {
		locationModels[i] = models.NewDriverLocation(
			dtoReq.Latitude,
			dtoReq.Longitude,
		)
	}

	result, err := h.service.CreateDriverLocationBulk(c.Request.Context(), locationModels)
	if err != nil {
		h.logger.Error("Failed to create bulk locations", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Success: false,
			Error:   config.ErrInternalServer,
		})
		return
	}

	c.JSON(http.StatusOK, dto.CreateLocationBulkResponse{
		Success: true,
		Data: dto.CreateLocationBulkData{
			Total:      result.Total,
			Successful: result.Successful,
			Failed:     result.Failed,
		},
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

	searchResult, err := h.service.SearchDriverLocation(c.Request.Context(), req.Latitude, req.Longitude, req.Radius)
	if err != nil {
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
	if len(searchResult) == 0 {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Success: false,
			Error:   config.ErrNotFound,
		})
		return
	}

	drivers := make([]dto.SearchResultLocation, len(searchResult))
	for i, e := range searchResult {
		drivers[i] = dto.SearchResultLocation{
			Latitude:  e.Latitude,
			Longitude: e.Longitude,
			Distance:  e.Distance,
		}
	}

	c.JSON(http.StatusOK, dto.SearchLocationResponse{
		Success: true,
		Data: dto.SearchLocationData{
			Locations: drivers,
			Total:     len(drivers),
		},
	})
}

// @Summary Import driver locations from CSV
// @Description Imports driver locations from a CSV file.
// @Tags locations
// @Accept text/csv
// @Produce json
// @Param request body string true "CSV data"
// @Success 200 {object} dto.ImportLocationCSVResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security ApiKeyAuth
// @Router /api/v1/locations/import [post]
func (h *LocationHandler) importDriverLocations(c *gin.Context) {
	result, err := h.service.ImportDriverLocationsFromCSV(c.Request.Context(), c.Request.Body)
	if err != nil {
		h.logger.Error("Failed to import locations from CSV", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.ImportLocationCSVResponse{
		Success: true,
		Data: dto.ImportLocationCSVData{
			Total:      result.Total,
			Successful: result.Successful,
			Failed:     result.Failed,
		},
	})
}
