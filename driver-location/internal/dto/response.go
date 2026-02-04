package dto

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

type HealthCheckResponse struct {
	Status string `json:"status" example:"ok"`
}

type LocationData struct {
	DriverID string `json:"driver_id"`
	Message  string `json:"message"`
}

type LocationResponse struct {
	Success bool         `json:"success"`
	Data    LocationData `json:"data"`
}

type BulkResult struct {
	Total      int         `json:"total"`
	Successful int         `json:"successful"`
	Failed     int         `json:"failed"`
	Errors     []BulkError `json:"errors,omitempty"`
}

type BulkError struct {
	DriverID string `json:"driver_id"`
	Error    string `json:"error"`
}

type BulkLocationResponse struct {
	Success bool       `json:"success"`
	Data    BulkResult `json:"data"`
}

type DriverLocation struct {
	DriverID  string  `json:"driver_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Distance  float64 `json:"distance"`
}

type SearchLocationData struct {
	Drivers []*DriverLocation `json:"drivers"`
	Total   int               `json:"total"`
}

type SearchLocationResponse struct {
	Success bool               `json:"success"`
	Data    SearchLocationData `json:"data"`
}
