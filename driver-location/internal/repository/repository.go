package repository

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"

	"github.com/BarkinBalci/bitaksi-case-study/driver-location/internal/config"
	"github.com/BarkinBalci/bitaksi-case-study/driver-location/internal/models"
)

type DriverLocationRepository interface {
	Create(ctx context.Context, location *models.DriverLocation) error
	CreateMany(ctx context.Context, locations []*models.DriverLocation) (int, error)
	Search(ctx context.Context, longitude, latitude, radius float64) ([]*models.SearchResult, error)
	Ping(ctx context.Context) error
}

type driverLocationRepository struct {
	collection *mongo.Collection
	logger     *zap.Logger
}

func NewDriverLocationRepository(ctx context.Context, collection *mongo.Collection, logger *zap.Logger) (DriverLocationRepository, error) {
	repo := &driverLocationRepository{
		collection: collection,
		logger:     logger,
	}

	twodsphere := mongo.IndexModel{
		Keys: bson.D{{Key: "location", Value: "2dsphere"}},
	}

	_, err := collection.Indexes().CreateOne(ctx, twodsphere)
	if err != nil {
		return nil, fmt.Errorf("failed to create geospatial index: %w", err)
	}

	return repo, nil
}

func (d driverLocationRepository) Create(ctx context.Context, location *models.DriverLocation) error {
	_, err := d.collection.InsertOne(ctx, location)
	if err != nil {
		return fmt.Errorf("failed to insert driver location: %w", err)
	}
	return nil
}

func (d driverLocationRepository) CreateMany(ctx context.Context, locations []*models.DriverLocation) (int, error) {
	docs := make([]interface{}, len(locations))
	for i, loc := range locations {
		docs[i] = loc
	}

	opts := options.InsertMany().SetOrdered(false)
	result, err := d.collection.InsertMany(ctx, docs, opts)

	insertedCount := 0
	if result != nil {
		insertedCount = len(result.InsertedIDs)
	}

	if err != nil {
		return insertedCount, err
	}

	return insertedCount, nil
}

func (d driverLocationRepository) Search(ctx context.Context, longitude, latitude, radius float64) ([]*models.SearchResult, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$geoNear", Value: bson.D{
			{Key: "near", Value: bson.D{
				{Key: "type", Value: "Point"},
				{Key: "coordinates", Value: bson.A{longitude, latitude}},
			}},
			{Key: "distanceField", Value: "distance"},
			{Key: "maxDistance", Value: radius},
			{Key: "spherical", Value: true},
		}}},
		{{Key: "$limit", Value: config.MaxSearchResults}},
	}

	cursor, err := d.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate driver locations: %w", err)
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			d.logger.Error("failed to close cursor", zap.Error(err))
		}
	}()

	var results []struct {
		ID       bson.ObjectID  `bson:"_id"`
		Location models.GeoJSON `bson:"location"`
		Distance float64        `bson:"distance"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode aggregation results: %w", err)
	}

	searchResults := make([]*models.SearchResult, len(results))
	for i, r := range results {
		searchResults[i] = &models.SearchResult{
			DriverID:  r.ID.Hex(),
			Latitude:  r.Location.Coordinates[1],
			Longitude: r.Location.Coordinates[0],
			Distance:  r.Distance,
		}
	}

	return searchResults, nil
}

func (d driverLocationRepository) Ping(ctx context.Context) error {
	return d.collection.Database().Client().Ping(ctx, nil)
}
