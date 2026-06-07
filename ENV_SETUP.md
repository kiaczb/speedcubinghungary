# Environment Configuration Guide

## Overview

This project uses environment-based configuration to manage different deployment environments (development, test, production) securely. **Credentials and secrets are NEVER committed to git.**

## Setup Structure

```
speedcubinghungary/
├── .env.example              # Root frontend template
├── apps/
│   ├── api/
│   │   ├── .env.example      # API template
│   │   └── internal/config/  # Config validation package
│   ├── web/
│   │   └── .env.example      # Next.js template
├── db/
│   └── .env.example          # PostgreSQL template
└── docker-compose.yml        # References .env files
```

## Quick Start

### 1. Create Local Environment Files

For **development** with Docker:

```bash
# Copy templates and create local files
cp .env.example .env.development.local
cp apps/api/.env.example apps/api/.env.development.local
cp apps/web/.env.example apps/web/.env.development.local
cp db/.env.example db/.env.development.local

# Edit with your local values
# You only need to change:
# - WCA_CLIENT_ID and WCA_CLIENT_SECRET (from WCA)
# - DB passwords
# - JWT_SECRET (keep it simple for dev, e.g., "dev-secret-12345-for-local-testing")
```

### 2. Load Environment Files

#### Option A: Docker Compose (Recommended for development)

```bash
# Docker automatically loads .env.development.local files
docker-compose up
```

#### Option B: Manual Environment Loading

```bash
# Load from .env file before running
export $(cat apps/api/.env.development.local | grep -v '#' | xargs)
go run ./cmd/server
```

#### Option C: Use .env file directly (for local development only)

```bash
# Go - install godotenv support
go get github.com/joho/godotenv
```

Then in your main.go:
```go
import "github.com/joho/godotenv"

func main() {
    godotenv.Load(".env.development.local")
    // ... rest of your code
}
```

### 3. Next.js Frontend

Next.js automatically loads `.env.*.local` files based on `NODE_ENV`:

```bash
# Development
npm run dev
# Automatically loads: .env.development.local

# Production build
npm run build
# Automatically loads: .env.production.local

# Start production server
npm start
```

## Environment Variables Reference

### Root Level (.env.example)

| Variable | Purpose | Example |
|----------|---------|---------|
| `NEXT_PUBLIC_WCA_CLIENT_ID` | WCA OAuth Client ID (visible in browser) | From WCA |
| `NEXT_PUBLIC_API_URL` | API endpoint from browser | `http://localhost:8080` |
| `NEXT_PUBLIC_BACKEND_LOGIN` | Login redirect | `http://localhost:8080/login` |
| `SERVER_API_URL` | API endpoint for SSR | `http://api:8080` |

### API Server (apps/api/.env.example)

| Variable | Purpose | Dev Example | Prod Example |
|----------|---------|-------|------|
| `DEPLOYMENT` | Environment mode | `development` | `production` |
| `JWT_SECRET` | JWT signing key | `dev-secret-12345` | Random 64-char hex |
| `DATABASE_URL` | PostgreSQL connection | `postgres://user:pass@localhost:5432/db` | Managed database URL |
| `WCA_CLIENT_ID` | OAuth ID | From WCA | From WCA |
| `WCA_CLIENT_SECRET` | OAuth Secret | From WCA | From WCA |

### Database (db/.env.example)

| Variable | Purpose |
|----------|---------|
| `POSTGRES_USER` | Admin username |
| `POSTGRES_PASSWORD` | Admin password |
| `POSTGRES_DB` | Database name |

### Frontend (apps/web/.env.example)

Same as root level, but specifically for Next.js.

## Security Best Practices

### ✅ DO

- ✅ Generate strong JWT secrets: `openssl rand -hex 32`
- ✅ Use `.env.development.local`, `.env.test.local`, `.env.production.local`
- ✅ Add `*.local` and `.env.*` to `.gitignore`
- ✅ Use different secrets for each environment
- ✅ Store production secrets in secure vaults (AWS Secrets Manager, HashiCorp Vault, etc.)
- ✅ Validate all required variables on startup (using the config package)
- ✅ Rotate JWT secrets periodically in production

### ❌ DON'T

- ❌ Commit `.env` files with real secrets
- ❌ Use the same JWT secret across environments
- ❌ Use weak secrets like "123456" in production
- ❌ Hardcode secrets in code
- ❌ Share `.env` files in unencrypted emails or chats
- ❌ Log environment variables in production
- ❌ Commit credentials to git (ever!)

## Environment-Specific Setups

### Development (Local with Docker)

```bash
# Setup
cp apps/api/.env.example apps/api/.env.development.local
# Edit with local values

# Run
docker-compose up

# The API will validate all required env vars and fail if any are missing
```

### Test Environment

```bash
# Setup
cp apps/api/.env.example apps/api/.env.test.local
# Use a test database and test OAuth credentials

# Run tests
export $(cat apps/api/.env.test.local | xargs)
go test ./...
```

### Production Deployment

```bash
# DO NOT create .env.production.local in git!
# Instead, set environment variables in your deployment platform:

# Docker/Kubernetes:
# - Use secrets management (e.g., Docker Secrets, K8s Secrets)
# - Mount secrets as environment variables

# AWS:
# - Use AWS Secrets Manager or Parameter Store

# Heroku/Cloud Run:
# - Set environment variables in dashboard or CLI

# Example:
# export JWT_SECRET=$(aws secretsmanager get-secret-value --secret-id prod/jwt-secret)
# export DATABASE_URL=$(aws secretsmanager get-secret-value --secret-id prod/db-url)
```

## Troubleshooting

### Error: "missing required environment variables"

The API validates all required variables on startup. Check:

```bash
# Check which variables are set
env | grep -E 'JWT_SECRET|DATABASE_URL|WCA_'

# Check your .env file
cat apps/api/.env.development.local

# Make sure the file is being loaded
echo "Test: $JWT_SECRET"
```

### Docker not picking up .env file

- Ensure `.env.development.local` is in the correct directory
- Docker Compose loads from current working directory
- Run `docker-compose config` to verify variable expansion

### JWT_SECRET validation fails

The JWT_SECRET must be at least 32 characters:

```bash
# Generate a secure random secret
openssl rand -hex 32
# Output: a1b2c3d4e5f6... (64 hex characters = 32 bytes)
```

### Next.js can't access environment variables

Next.js only exposes variables prefixed with `NEXT_PUBLIC_` to the browser. Server-side variables:
- Are available in `.env.local` or environment
- Can be accessed in API routes and `getServerSideProps`
- Are NOT exposed to browser

## Git Workflow

```bash
# Initial setup
git clone <repo>
cd speedcubinghungary

# Create your local environment files (not tracked)
cp .env.example .env.development.local
cp apps/api/.env.example apps/api/.env.development.local
cp apps/web/.env.example apps/web/.env.development.local
cp db/.env.example db/.env.development.local

# Edit the files with your local values
nano apps/api/.env.development.local

# Start developing
docker-compose up

# The .env*.local files are ignored by git (in .gitignore)
git status  # Should not show .env files
```

## Related Files

- **Validation**: `apps/api/internal/config/config.go`
- **Auth Middleware**: `apps/api/internal/middleware/auth.go`
- **Docker Setup**: `docker-compose.yml`
