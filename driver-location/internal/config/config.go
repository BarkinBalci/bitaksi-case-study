package config

import (
	"fmt"
	"os"
	"strings"
)

// Config holds all application configuration.
type Config struct {
	ApiKey              string
	Environment         string
	SwaggerEnabled      bool
	MongoURI            string
	MongoDBName         string
	MongoCollectionName string
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

	cfg := &Config{
		ApiKey:              getEnv("X_API_KEY", ""),
		Environment:         getEnv("ENVIRONMENT", "development"),
		SwaggerEnabled:      parseBool(getEnv("SWAGGER_ENABLED", "true")),
		MongoURI:            getEnv("MONGO_URI", ""),
		MongoDBName:         getEnv("MONGO_DB_NAME", ""),
		MongoCollectionName: getEnv("MONGO_COLLECTION_NAME", ""),
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return cfg, nil
}

func parseBool(s string) bool {
	return strings.EqualFold(s, "true")
}
