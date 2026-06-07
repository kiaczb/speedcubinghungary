# Environment Management Guide

## Overview

The project uses a **single `.env` file per location** (root, apps/api, apps/web, db). Environment mode (development/test/production) is controlled by the **`NODE_ENV` and `DEPLOYMENT`** variables within these files, not by separate files.

## Project Structure

```
.
├── .env                    # Root environment (frontend public vars)
├── .env.example           # Template reference
├── apps/api/
│   ├── .env               # API server config
│   └── .env.example       # Template reference
├── apps/web/
│   ├── .env               # Frontend config
│   └── .env.example       # Template reference
└── db/
    ├── .env               # Database config
    └── .env.example       # Template reference
```

## Environment Variables Explained

### Development (Default)

**Backend (apps/api/.env):**
```
DEPLOYMENT=development
DATABASE_URL=postgres://dbuser:Speedcubing-DEV@db:5432/speedcubinghungary_dev?sslmode=disable
```

**Frontend (apps/web/.env & root/.env):**
```
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_BACKEND_LOGIN=http://localhost:8080/login
SERVER_API_URL=http://api:8080
```

**Database (db/.env):**
```
POSTGRES_USER=dbuser
POSTGRES_PASSWORD=Speedcubing-DEV
POSTGRES_DB=speedcubinghungary_dev
```

### Production

Update the `.env` files with production values:

**Backend (apps/api/.env):**
```
DEPLOYMENT=production
DATABASE_URL=postgres://prod_user:prod_password@prod-db-host:5432/speedcubinghungary_prod?sslmode=require
```

**Frontend (apps/web/.env & root/.env):**
```
NEXT_PUBLIC_API_URL=https://api.yourdomain.com
NEXT_PUBLIC_BACKEND_LOGIN=https://api.yourdomain.com/login
SERVER_API_URL=https://api.yourdomain.com
```

**Database (db/.env):**
```
POSTGRES_USER=prod_user
POSTGRES_PASSWORD=secure_production_password
POSTGRES_DB=speedcubinghungary_prod
```

### Test

Update the `.env` files with test values:

**Backend (apps/api/.env):**
```
DEPLOYMENT=test
DATABASE_URL=postgres://test_user:test_password@localhost:5432/speedcubinghungary_test?sslmode=disable
```

## Switching Environments

### Docker Compose

The `docker-compose.yml` loads environment variables from `.env` files:

```bash
# Development (default)
docker-compose up

# For production, manually update .env files first, then:
docker-compose up
```

### Backend (Go)

The backend reads environment variables via `config.Load()`:

```bash
# Run with current .env
DEPLOYMENT=development go run ./cmd/server

# Or load from .env file first
export $(cat apps/api/.env | xargs)
go run ./cmd/server
```

### Frontend (Next.js)

Next.js automatically reads variables prefixed with `NEXT_PUBLIC_`:

```bash
# Development
npm run dev

# Production build
npm run build
npm start
```

## Best Practices

1. **Never commit actual `.env` files** - they contain secrets
2. **Only commit `.env.example`** files as templates
3. **Update `.env` files for each environment** you deploy to
4. **Use `.env.example` as a reference** for required variables
5. **Rotate secrets regularly** in production
6. **Use different database credentials** for each environment (dev/test/prod)

## How Environment Control Works

1. **Backend** reads `DEPLOYMENT` variable via `config.Load()`
   - Uses `IsDevelopment()`, `IsProduction()`, `IsTest()` helper methods
   - Adjusts logging, CORS, and database settings accordingly

2. **Frontend** reads `NEXT_PUBLIC_*` variables automatically
   - Different API URLs for different environments prevent cross-environment traffic

3. **Database** is configured via `POSTGRES_*` variables
   - Each environment has separate database name and credentials

## Environment-Specific Database Setup

For **production deployment** with a managed PostgreSQL service (AWS RDS, etc.):

1. Create separate databases:
   - `speedcubinghungary_dev` (development)
   - `speedcubinghungary_test` (testing)
   - `speedcubinghungary_prod` (production)

2. Update `DATABASE_URL` in respective `.env` files
3. Ensure SSL is enabled for production: `?sslmode=require`

## Troubleshooting

### "Missing required environment variables"

**Problem:** Config validation fails

**Solution:**
- Check `.env` file exists in the correct location
- Verify all required variables are set (see `.env.example`)
- For Docker: verify `env_file` path in `docker-compose.yml`

### Wrong environment active

**Problem:** Code is running in wrong environment mode

**Solution:**
- Verify `DEPLOYMENT` variable is set correctly in `.env`
- Check what value the code is reading: `echo $DEPLOYMENT`
- For Docker: inspect container env: `docker exec <container> env`

### Database connection refused

**Problem:** Can't connect to database

**Solution:**
- Verify `DATABASE_URL` in `.env`
- Ensure database is running and credentials are correct
- For Docker: check db service health: `docker-compose ps`

