# Quick Start: Environment & Security Setup

## What I've Created

### 📋 Configuration Files
- ✅ `.env.example` - Root template
- ✅ `apps/api/.env.example` - API template with all required variables
- ✅ `apps/web/.env.example` - Frontend template
- ✅ `db/.env.example` - Database template

### 🔐 .gitignore Updates
- ✅ Updated root `.gitignore` to exclude `.env.local` files
- ✅ Created `apps/api/.gitignore`
- ✅ Created `db/.gitignore`

### 💻 Code
- ✅ `apps/api/internal/config/config.go` - Environment validation package
- ✅ `apps/api/internal/config/README.md` - Config package documentation

### 📚 Documentation
- ✅ `ENV_SETUP.md` - Complete environment setup guide
- ✅ `SECURITY.md` - Security best practices and recommendations

## Next Steps (Copy & Paste Commands)

### 1️⃣ Create Local Environment Files

```bash
# From root directory
cp .env.example .env.development.local
cp apps/api/.env.example apps/api/.env.development.local
cp apps/web/.env.example apps/web/.env.development.local
cp db/.env.example db/.env.development.local
```

### 2️⃣ Edit Environment Files

```bash
# Replace placeholder values with your actual credentials
nano apps/api/.env.development.local
nano db/.env.development.local

# You need to change:
# 1. WCA_CLIENT_ID - Get from https://www.worldcubeassociation.org/oauth/applications
# 2. WCA_CLIENT_SECRET - Get from WCA OAuth settings
# 3. Database passwords - Use secure values
# 4. JWT_SECRET - Can use simple value for dev like: "dev-secret-12345"
```

### 3️⃣ Verify Everything Works

```bash
# Start Docker services
docker-compose up

# In another terminal, test the API
curl -X GET http://localhost:8080/health

# Check that environment variables are loaded
docker-compose exec api env | grep JWT_SECRET
```

### 4️⃣ (Optional) Integrate Config Validation in Go

Update `apps/api/cmd/server/main.go` to use the new config package:

```go
package main

import (
    "log"
    "net/http"
    "os"
    "speedcubinghungary/internal/config"
    "speedcubinghungary/internal/db"
    "speedcubinghungary/internal/router"

    "github.com/golang-migrate/migrate/v4"
    _ "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
    // Load and validate environment variables
    cfg, err := config.Load()
    if err != nil {
        log.Fatal("Config error:", err)
    }

    // 1. DB connection
    dbConn, err := db.NewPostgres()
    if err != nil {
        log.Fatal("DB error:", err)
    }
    defer dbConn.Close()

    // 2. Database migration
    m, err := migrate.New("file:///migrations", cfg.DatabaseURL)
    if err != nil {
        log.Fatal("Migration init error:", err)
    }
    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        log.Fatal("Migration run error:", err)
    }

    // 3. Create router
    r := router.New(dbConn)

    log.Printf("Server started on :8080 (mode: %s)", cfg.Deployment)
    log.Fatal(http.ListenAndServe("0.0.0.0:8080", r))
}
```

### 5️⃣ Commit to Git

```bash
# Verify .env files are NOT staged
git status

# You should see .env*.local files are NOT in the staged changes
# Only .env.example files should be tracked

# Stage the new files
git add ENV_SETUP.md SECURITY.md
git add .gitignore apps/api/.gitignore db/.gitignore
git add apps/api/.env.example apps/web/.env.example db/.env.example
git add .env.example apps/api/internal/config/

# Commit
git commit -m "feat: add secure environment configuration and validation

- Create environment templates (.env.example) for dev/test/prod
- Add Go config validation package for environment variable checking
- Update .gitignore to protect secrets (.env.local, .env.*.local)
- Add comprehensive environment setup and security documentation
- Implement validation for required variables and JWT secret strength

Fixes security issues with hardcoded credentials."
```

## File Structure Reference

```
speedcubinghungary/
├── .env.example                    # Template (committed)
├── .env.development.local          # Your dev secrets (NOT committed, git ignored)
├── ENV_SETUP.md                    # Setup guide
├── SECURITY.md                     # Security best practices
├── .gitignore                      # Updated with .env*.local
│
├── apps/api/
│   ├── .env.example                # Template (committed)
│   ├── .env.development.local      # Your dev secrets (NOT committed)
│   ├── .gitignore                  # New file
│   └── internal/config/
│       ├── config.go               # Environment validation
│       └── README.md               # Config package docs
│
├── apps/web/
│   └── .env.example                # Template (committed)
│   └── .env.development.local      # Your dev secrets (NOT committed)
│
└── db/
    ├── .env.example                # Template (committed)
    ├── .env.development.local      # Your dev secrets (NOT committed)
    └── .gitignore                  # New file
```

## Security Checklist ✅

- [ ] Created `.env.development.local` files with real values
- [ ] Added `.env*.local` to `.gitignore` (already done)
- [ ] Generated strong JWT_SECRET for production: `openssl rand -hex 32`
- [ ] Read `SECURITY.md` for recommendations
- [ ] Set up separate environments (dev/test/prod)
- [ ] Never commit `.env` files with secrets
- [ ] Verified Docker loads environment variables correctly

## Files to Review

1. **ENV_SETUP.md** - Complete setup guide with troubleshooting
2. **SECURITY.md** - Detailed security recommendations
3. **apps/api/internal/config/README.md** - Config package usage
4. **apps/api/internal/config/config.go** - Implementation details

## Common Issues & Solutions

### Q: Where do I get WCA_CLIENT_ID and WCA_CLIENT_SECRET?
**A:** Register your application at: https://www.worldcubeassociation.org/oauth/applications

### Q: What should JWT_SECRET be?
**A:** 
- Dev: Simple value like "dev-secret-12345-for-local-testing"
- Test: Fixed value like "test-secret-fixed-for-ci-12345"
- Prod: Run `openssl rand -hex 32` to generate strong secret

### Q: Docker can't find .env file?
**A:** Make sure `.env.development.local` exists and `docker-compose.yml` references it

### Q: Can I commit the `.env` files?
**A:** NO! Only `.env.example` files should be committed. `.env*.local` files are git-ignored for security.

## What Changed (Security Improvements)

### ✅ Before
- ❌ Hardcoded credentials in .env (visible in git)
- ❌ No environment separation
- ❌ No validation of required variables
- ❌ Same secrets across all environments

### ✅ After
- ✅ Environment files are git-ignored (secrets safe)
- ✅ Separate dev/test/prod configurations
- ✅ Automatic validation on app startup
- ✅ Different secrets per environment
- ✅ Strong JWT secret validation
- ✅ Configuration package for type-safe access
- ✅ Comprehensive security documentation

## Need Help?

1. Check `ENV_SETUP.md` for detailed setup instructions
2. Check `SECURITY.md` for security recommendations
3. Check `apps/api/internal/config/README.md` for config package usage
