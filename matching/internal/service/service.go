package service

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/BarkinBalci/bitaksi-case-study/matching/internal/client"
	"github.com/BarkinBalci/bitaksi-case-study/matching/internal/config"
	"github.com/BarkinBalci/bitaksi-case-study/matching/internal/dto"
)

var ErrNoDriverFound = errors.New("no driver found")

type Service interface {
	FindNearestDriver(ctx context.Context, req *dto.MatchRequest) (*dto.DriverMatch, error)
	HealthCheck(ctx context.Context) error
}

type service struct {
	driverLocationClient *client.DriverLocationClient
	config               *config.Config
	logger               *zap.Logger
}

func NewService(driverLocationClient *client.DriverLocationClient, cfg *config.Config, logger *zap.Logger) Service {
	return &service{
		driverLocationClient: driverLocationClient,
		config:               cfg,
		logger:               logger,
	}
}

func (s service) FindNearestDriver(ctx context.Context, req *dto.MatchRequest) (*dto.DriverMatch, error) {
	lon := req.Location.Coordinates[0]
	lat := req.Location.Coordinates[1]

	searchResp, err := s.driverLocationClient.SearchDrivers(ctx, lat, lon, float64(s.config.SearchRadius))
	if err != nil {
		return nil, fmt.Errorf("failed to search drivers: %w", err)
	}

	if !searchResp.Success || len(searchResp.Data.Locations) == 0 {
		return nil, ErrNoDriverFound
	}

	nearest := searchResp.Data.Locations[0]

	return &dto.DriverMatch{
		ID: nearest.ID,
		Location: dto.GeoJSONPoint{
			Type:        nearest.Location.Type,
			Coordinates: nearest.Location.Coordinates,
		},
		Distance: nearest.Distance,
	}, nil
}

func (s service) HealthCheck(ctx context.Context) error {
	return nil
}
