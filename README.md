# bitaksi-case-study
A driver-rider matching system consisting of two microservices.

[![CI](https://github.com/BarkinBalci/bitaksi-case-study/actions/workflows/ci-driver-location.yml/badge.svg)](https://github.com/BarkinBalci/bitaksi-case-study/actions/workflows/ci-driver-location.yml)
[![CI](https://github.com/BarkinBalci/bitaksi-case-study/actions/workflows/ci-matching.yml/badge.svg)](https://github.com/BarkinBalci/bitaksi-case-study/actions/workflows/ci-matching.yml)

## Prerequisites
- Go 1.25+
- Docker
- Docker Compose

## Quick Start

### Setting up the environment variables

```bash
cp driver-location/.env.example driver-location/.env
cp matching/.env.example matching/.env
```

### Building and running the services

```bash
docker-compose up -d --build
```

### Bootstrapping data to the database
```bash
curl -X POST http://localhost:8080/api/v1/locations/import \
  -H "Content-Type: text/csv" \
  -H "X-API-Key: an-api-key" \
  --data-binary @bootstrap.csv
```

## API Documentation

### Driver Location Service

Swagger UI: `http://localhost:8080/swagger/index.html`

**Endpoints:**
- `POST /api/v1/locations` - Create a driver location
- `POST /api/v1/locations/bulk` - Bulk create driver locations
- `POST /api/v1/locations/import` - Import locations from CSV
- `POST /api/v1/locations/search` - Search locations by coordinates and radius
- `GET /health` - Health check

**Authentication:** Requires `X-API-Key` header for service-to-service communication

### Matching Service

Swagger UI: `http://localhost:8081/swagger/index.html`

**Endpoints:**
- `POST /api/v1/match` - Find nearest driver for given coordinates
- `GET /health` - Health check

**Authentication:** Requires JWT token with `authenticated: true` in payload