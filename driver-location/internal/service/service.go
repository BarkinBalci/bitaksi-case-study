package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"

	"go.uber.org/zap"

	"github.com/BarkinBalci/bitaksi-case-study/driver-location/internal/models"
	"github.com/BarkinBalci/bitaksi-case-study/driver-location/internal/repository"
)

type Service interface {
	CreateDriverLocation(ctx context.Context, location *models.DriverLocation) error
	CreateDriverLocationBulk(ctx context.Context, locations []*models.DriverLocation) (*models.BulkResult, error)
	SearchDriverLocation(ctx context.Context, latitude, longitude, radius float64) ([]*models.SearchResult, error)
	ImportDriverLocationsFromCSV(ctx context.Context, reader io.Reader) (*models.BulkResult, error)
	HealthCheck(ctx context.Context) error
}

type service struct {
	repo   repository.DriverLocationRepository
	logger *zap.Logger
}

func NewService(repo repository.DriverLocationRepository, logger *zap.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

func (s service) HealthCheck(ctx context.Context) error {
	if err := s.repo.Ping(ctx); err != nil {
		s.logger.Error("health check failed", zap.Error(err))
		return fmt.Errorf("health check failed: %w", err)
	}
	return nil
}

func (s service) CreateDriverLocation(ctx context.Context, location *models.DriverLocation) error {
	err := s.repo.Create(ctx, location)
	if err != nil {
		s.logger.Error("failed to create driver location",
			zap.Error(err),
		)
		return fmt.Errorf("failed to create driver location: %w", err)
	}

	return nil
}

func (s service) CreateDriverLocationBulk(ctx context.Context, locations []*models.DriverLocation) (*models.BulkResult, error) {
	successCount, err := s.repo.CreateMany(ctx, locations)
	if err != nil {
		s.logger.Error("failed to create driver locations due to general error",
			zap.Error(err))
		return nil, fmt.Errorf("failed to create driver locations: %w", err)
	}

	totalCount := len(locations)
	failCount := totalCount - successCount

	result := &models.BulkResult{
		Total:      totalCount,
		Successful: successCount,
		Failed:     failCount,
	}

	if failCount > 0 {
		s.logger.Warn("some driver locations failed to be created in bulk operation",
			zap.Int("total", totalCount),
			zap.Int("successful", successCount),
			zap.Int("failed", failCount),
		)
	}

	return result, nil
}

func (s service) SearchDriverLocation(ctx context.Context, latitude, longitude, radius float64) ([]*models.SearchResult, error) {
	results, err := s.repo.Search(ctx, longitude, latitude, radius)
	if err != nil {
		s.logger.Error("failed to search driver locations",
			zap.Error(err),
			zap.Float64("latitude", latitude),
			zap.Float64("longitude", longitude),
			zap.Float64("radius", radius),
		)
		return nil, fmt.Errorf("failed to search driver locations: %w", err)
	}

	return results, nil
}

func (s service) ImportDriverLocationsFromCSV(ctx context.Context, reader io.Reader) (*models.BulkResult, error) {
	csvReader := csv.NewReader(reader)
	records, err := csvReader.ReadAll()
	if err != nil {
		s.logger.Error("Failed to read CSV data", zap.Error(err))
		return nil, fmt.Errorf("failed to read CSV data: %w", err)
	}

	// Assuming the first row is a header, so we skip it.
	records = records[1:]

	var locations []*models.DriverLocation
	for _, record := range records {
		lat, err := strconv.ParseFloat(record[0], 64)
		if err != nil {
			s.logger.Error("Failed to parse latitude from CSV record",
				zap.Error(err),
				zap.Strings("record", record),
			)
			return nil, fmt.Errorf("failed to parse latitude: %w", err)
		}

		lon, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			s.logger.Error("Failed to parse longitude from CSV record",
				zap.Error(err),
				zap.Strings("record", record),
			)
			return nil, fmt.Errorf("failed to parse longitude: %w", err)
		}
		locations = append(locations, models.NewDriverLocation(lat, lon))
	}

	return s.CreateDriverLocationBulk(ctx, locations)
}
