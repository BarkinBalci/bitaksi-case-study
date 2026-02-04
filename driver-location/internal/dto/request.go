package dto

type CreateLocationRequest struct {
	DriverID  string  `json:"driver_id" binding:"required"`
	Latitude  float64 `json:"latitude" binding:"required,latitude"`
	Longitude float64 `json:"longitude" binding:"required,longitude"`
}

type CreateLocationBulkRequest struct {
	Locations []CreateLocationRequest `json:"locations" binding:"required,min=1,max=1000,dive"`
}

type SearchLocationRequest struct {
	Latitude  float64 `json:"latitude" binding:"required,latitude"`
	Longitude float64 `json:"longitude" binding:"required,longitude"`
	Radius    float64 `json:"radius" binding:"required,min=0.1,max=50"`
}
