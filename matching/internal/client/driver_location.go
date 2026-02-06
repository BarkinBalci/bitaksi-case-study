package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type DriverLocationClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewDriverLocationClient(baseURL, apiKey string) *DriverLocationClient {
	return &DriverLocationClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type GeoJSONPoint struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

type SearchRequest struct {
	Location GeoJSONPoint `json:"location"`
	Radius   float64      `json:"radius"`
}

type SearchResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Locations []struct {
			ID       string       `json:"id"`
			Location GeoJSONPoint `json:"location"`
			Distance float64      `json:"distance"`
		} `json:"locations"`
	} `json:"data"`
}

func (c *DriverLocationClient) SearchDrivers(ctx context.Context, lat, lon, radius float64) (*SearchResponse, error) {
	searchReq := SearchRequest{
		Location: GeoJSONPoint{
			Type:        "Point",
			Coordinates: []float64{lon, lat},
		},
		Radius: radius,
	}

	body, err := json.Marshal(searchReq)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/locations/search", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return &SearchResponse{Success: false}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}
