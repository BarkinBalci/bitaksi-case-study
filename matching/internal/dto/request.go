package dto

type GeoJSONPoint struct {
	Type        string    `json:"type" binding:"required,eq=Point" example:"Point"`
	Coordinates []float64 `json:"coordinates" binding:"required,len=2" example:"28.9784,41.0082" swaggertype:"array,number"`
}

type MatchRequest struct {
	Location GeoJSONPoint `json:"location" binding:"required"`
}
