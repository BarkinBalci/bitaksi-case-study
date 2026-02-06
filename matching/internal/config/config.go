package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all application configuration.
type Config struct {
	DriverLocationApiKey  string
	Environment           string
	SwaggerEnabled        bool
	DriverLocationBaseURL string
	SearchRadius          int
}

// LoadConfig loads configuration from environment variables.
func LoadConfig() (*Config, error) {
	var missing []string

	getEnv := func(key, defaultValue string) string {
		if value, exists := os.LookupEnv(key); exists {
			return value
		}
		if defaultValue == "" {
			missing = append(missing, key)
		}
		return defaultValue
	}

	searchRadius, err := parseInt(getEnv("SEARCH_RADIUS", "8000"), "SEARCH_RADIUS")
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		DriverLocationApiKey:  getEnv("DRIVER_LOCATION_X_API_KEY", ""),
		Environment:           getEnv("ENVIRONMENT", "development"),
		SwaggerEnabled:        parseBool(getEnv("SWAGGER_ENABLED", "true")),
		DriverLocationBaseURL: getEnv("DRIVER_LOCATION_BASE_URL", "http://localhost:8080"),
		SearchRadius:          searchRadius,
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return cfg, nil
}

func parseBool(s string) bool {
	return strings.EqualFold(s, "true")
}

func parseInt(s, fieldName string) (int, error) {
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid %s value '%s': %w", fieldName, s, err)
	}
	return v, nil
}
