# Core Service for Shopper App

This service provides the core business logic and API endpoints for the Shopper app.

## Overview

The core service implements:
- Flyer search
- Shop list management
- User profile

## Development

### Prerequisites

- Go 1.16+
- Redis 7.0+
- MySQL 8.0+

### Running Locally
You can obtain the docker compose files from [ai_shopper_docker_compose](https://github.com/kdjuwidja/ai_shopper_docker_compose)

1. Start Redis and MySQL using Docker Compose:
   ```
   docker-compose -f docker-compose-infra.yml up -d
   ```

2. Build the service with 
    ```
    docker-compose build core
    ```

3. Run the service with 
    ```
    docker-compose up
    ```

### Testing

Run tests with:
```
go test ./... -v
```

Requires the testing infra to be up and running. Run testing infra with
```
docker-compose -f docker-compose-test.yml up -d
```