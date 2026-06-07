# Config Package

This package handles environment variable loading and validation for the API server.

## Usage

### In main.go

```go
package main

import (
    "log"
    "speedcubinghungary/internal/config"
)

func main() {
    // Load and validate environment variables
    cfg, err := config.Load()
    if err != nil {
        log.Fatal("Config error:", err)
    }

    // Use the config
    println("Running in:", cfg.Deployment)
    println("JWT Secret loaded: ", cfg.JWTSecret != "")
}
```

## Features

- ✅ Loads all required environment variables
- ✅ Validates that required vars are set
- ✅ Validates JWT_SECRET strength (minimum 32 chars)
- ✅ Provides environment checking helpers (`IsDevelopment()`, `IsProduction()`)
- ✅ Defaults for non-critical variables

## Required Environment Variables

- `DATABASE_URL` - PostgreSQL connection string
- `JWT_SECRET` - JWT signing key (min 32 chars)
- `WCA_CLIENT_ID` - WCA OAuth client ID
- `WCA_CLIENT_SECRET` - WCA OAuth client secret
- `WCA_REDIRECT_URI` - OAuth redirect URL
- `FRONTEND_URL` - Frontend URL for CORS

## Optional Environment Variables

- `DEPLOYMENT` - deployment mode (default: "development")
- `LOG_LEVEL` - log level (default: "info")

## Loading Environment Files

### Option 1: Docker (Recommended)

Docker automatically loads `.env` files specified in `docker-compose.yml`.

### Option 2: godotenv package

```bash
go get github.com/joho/godotenv
```

Then in main.go:

```go
import "github.com/joho/godotenv"

func main() {
    godotenv.Load(".env.development.local")
    cfg, err := config.Load()
    // ...
}
```

### Option 3: Shell environment

```bash
export $(cat .env.development.local | xargs)
go run ./cmd/server
```

## Error Handling

The config package will fail fast if required variables are missing:

```
missing required environment variables: [JWT_SECRET DATABASE_URL]
Please check your .env file or environment setup
```

And if JWT_SECRET is too weak:

```
JWT_SECRET must be at least 32 characters long (use 'openssl rand -hex 32' to generate)
```

## Generating Secrets

```bash
# Generate a secure JWT_SECRET
openssl rand -hex 32
# Output: a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6a7b8c9d0e1f

# Copy this value to your .env file
```

## Testing

When running tests, set the config to test mode:

```bash
DEPLOYMENT=test go test ./...
```
