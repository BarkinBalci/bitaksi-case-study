package dto

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

type HealthCheckResponse struct {
	Status string `json:"status" example:"ok"`
}

type CreateLocationData struct {
	Message string `json:"message"`
}

type CreateLocationResponse struct {
	Success bool               `json:"success"`
	Data    CreateLocationData `json:"data"`
}

type CreateLocationBulkResponse struct {
	Success bool                   `json:"success"`
	Data    CreateLocationBulkData `json:"data"`
}

type CreateLocationBulkData struct {
	Total      int `json:"total"`
	Successful int `json:"successful"`
	Failed     int `json:"failed"`
}

type SearchLocationResponse struct {
	Success bool               `json:"success"`
	Data    SearchLocationData `json:"data"`
}

type SearchLocationData struct {
	Locations []SearchResultLocation `json:"locations"`
	Total     int                    `json:"total"`
}

type SearchResultLocation struct {
	ID       string       `json:"id"`
	Location GeoJSONPoint `json:"location"`
	Distance float64      `json:"distance"`
}

type ImportLocationCSVResponse struct {
	Success bool                  `json:"success"`
	Data    ImportLocationCSVData `json:"data"`
}

type ImportLocationCSVData struct {
	Total      int `json:"total"`
	Successful int `json:"successful"`
	Failed     int `json:"failed"`
}
