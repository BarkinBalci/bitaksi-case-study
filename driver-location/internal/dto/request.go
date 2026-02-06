package dto

type GeoJSONPoint struct {
	Type        string    `json:"type" binding:"required,eq=Point" example:"Point"`
	Coordinates []float64 `json:"coordinates" binding:"required,len=2" example:"28.9784,41.0082" swaggertype:"array,number"`
}

type CreateLocationRequest struct {
	Latitude  float64 `json:"latitude" binding:"required,latitude"`
	Longitude float64 `json:"longitude" binding:"required,longitude"`
}

type CreateLocationBulkRequest struct {
	Locations []CreateLocationRequest `json:"locations" binding:"required,min=1,max=1000,dive"`
}

type SearchLocationRequest struct {
	Location GeoJSONPoint `json:"location" binding:"required"`
	Radius   float64      `json:"radius" binding:"required,min=10,max=10000"`
}
