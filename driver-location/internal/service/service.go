package service

import (
	"context"
	"errors"

	"github.com/BarkinBalci/bitaksi-case-study/driver-location/internal/dto"
)

var ErrNotImplemented = errors.New("not implemented")

type Service interface {
	CreateDriverLocation(ctx context.Context, req *dto.CreateLocationRequest) error
	CreateDriverLocationBulk(ctx context.Context, req *dto.CreateLocationBulkRequest) (*dto.BulkResult, error)
	SearchDriverLocation(ctx context.Context, req *dto.SearchLocationRequest) ([]*dto.DriverLocation, error)
}

type service struct{}

func NewService() Service {
	return &service{}
}

func (s service) CreateDriverLocation(ctx context.Context, req *dto.CreateLocationRequest) error {
	//TODO: Implement me
	return ErrNotImplemented
}

func (s service) CreateDriverLocationBulk(ctx context.Context, req *dto.CreateLocationBulkRequest) (*dto.BulkResult, error) {
	//TODO: Implement me
	return nil, ErrNotImplemented
}

func (s service) SearchDriverLocation(ctx context.Context, req *dto.SearchLocationRequest) ([]*dto.DriverLocation, error) {
	//TODO: Implement me
	return nil, ErrNotImplemented
}
