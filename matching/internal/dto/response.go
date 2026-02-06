package dto

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

type HealthCheckResponse struct {
	Status string `json:"status" example:"ok"`
}

type DriverMatch struct {
	ID       string       `json:"id"`
	Location GeoJSONPoint `json:"location"`
	Distance float64      `json:"distance"`
}

type MatchResponse struct {
	Success bool         `json:"success"`
	Data    *DriverMatch `json:"data"`
}
