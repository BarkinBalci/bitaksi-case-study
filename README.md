# bitaksi-case-study
A driver-rider matching system consisting of two microservices.

[![CI](https://github.com/BarkinBalci/bitaksi-case-study/actions/workflows/ci-driver-location.yml/badge.svg)](https://github.com/BarkinBalci/bitaksi-case-study/actions/workflows/ci-driver-location.yml)
[![CI](https://github.com/BarkinBalci/bitaksi-case-study/actions/workflows/ci-matching.yml/badge.svg)](https://github.com/BarkinBalci/bitaksi-case-study/actions/workflows/ci-matching.yml)

## Overview
**Driver Location Service** stores driver locations in MongoDB using GeoJSON Points with a [2dsphere geospatial index](https://www.mongodb.com/docs/manual/core/indexes/index-types/geospatial/2dsphere/) for proximity searches via [$geoNear aggregation step](https://www.mongodb.com/docs/manual/reference/operator/aggregation/geoNear/). It provides endpoints for location management and radius-based queries. **Matching Service** acts as the client facing API, authenticating riders with JWT and calling the Driver Location Service to find the nearest available driver.

## Quick Start

### Prerequisites
- [Go 1.25+](https://go.dev/doc/install)
- [Docker](https://docs.docker.com/get-started/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install)

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

### Accessing the Swagger UIs
- Driver Location Service: http://localhost:8080/swagger/index.html
- Matching Service: http://localhost:8081/swagger/index.html

## Examples

### Driver Location Service
All `/api/v1/*` endpoints require an `X-API-Key` header with a valid API key.

#### Create a driver location
```bash
curl -X POST http://localhost:8080/api/v1/locations \
  -H "Content-Type: application/json" \
  -H "X-API-Key: an-api-key" \
  -d '{
    "latitude": 41.015137,
    "longitude": 28.979530
  }'
```

#### Batch create driver locations
```bash
curl -X POST http://localhost:8080/api/v1/locations/batch \
  -H "Content-Type: application/json" \
  -H "X-API-Key: an-api-key" \
  -d '{
    "locations": [
      {
        "latitude": 41.015137,
        "longitude": 28.979530
      },
      {
        "latitude": 41.016137,
        "longitude": 28.980530
      }
    ]
  }'
```

#### Search for nearby drivers
```bash
curl -X POST http://localhost:8080/api/v1/locations/search \
  -H "Content-Type: application/json" \
  -H "X-API-Key: an-api-key" \
  -d '{
    "location": {
      "type": "Point",
      "coordinates": [28.979530, 41.015137]
    },
    "radius": 5000
  }'
```

#### Health check
```bash
curl http://localhost:8080/health
```

### Matching Service
All `/api/v1/*` endpoints require an Authorization header with a valid JWT Bearer token. 
A development token is provided in `.env.example` as `DEV_JWT_TOKEN`.

#### Find nearest driver
```bash
# Load service .env file
source matching/.env

curl -X POST http://localhost:8081/api/v1/match \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $DEV_JWT_TOKEN" \
  -d '{
    "location": {
      "type": "Point",
      "coordinates": [28.979530, 41.015137]
    }
  }'
```

**Health check:**
```bash
curl http://localhost:8081/health
```

## Limitations & Future Enhancements

### Current Limitations
- Matching Service can be impacted if Driver Location Service is slow or down
- No retry logic for transient failures
- No rate limiting or caching layer

### Enhancement Opportunities
- **Resilience**: Circuit breaker implementation, service mesh, retry with backoff, graceful degradation
- **Performance**: Redis caching, connection pool tuning, MongoDB replica set, async processing with message queues
- **Operations**: Integration tests, load testing, observability