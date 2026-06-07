# Security Best Practices

## Authentication Security Review

### ✅ Current Implementation Strengths

Your `apps/api/internal/middleware/auth.go` has good security practices:

1. **HMAC Algorithm Validation** ✅
   - Checks that token uses HMAC signing method
   - Prevents algorithm substitution attacks
   ```go
   if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
       return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
   }
   ```

2. **JWT Token Validation** ✅
   - Validates token signature and expiration
   - Checks claims type properly

3. **RBAC Implementation** ✅
   - Role-based access control enforced
   - Proper unauthorized/forbidden responses

### ⚠️ Security Recommendations

#### 1. **JWT Secret Management** (HIGH PRIORITY)

**Current Risk**: JWT_SECRET is hardcoded and visible in `.env` files

**Fix**: Implemented in this update
- Create separate secrets per environment (dev/test/prod)
- Generate strong secrets: `openssl rand -hex 32`
- Never commit real secrets to git
- Use `.gitignore` to exclude `.env` files

**Production**: Use secure secret management
```go
// Example: Load from AWS Secrets Manager
import "github.com/aws/aws-secretsmanager-caching-go/cache"

secret, _ := client.GetSecretString(ctx, "api/jwt-secret")
```

---

#### 2. **JWT Token Expiration** (HIGH PRIORITY)

**Current Issue**: Claims structure includes `jwt.RegisteredClaims` but expiration isn't validated in middleware

**Recommended Fix**:
```go
// In middleware/auth.go - Add to UserClaims validation

if claims, ok := token.Claims.(*UserClaims); ok {
    // Check expiration
    if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
        http.Error(w, "token expired", http.StatusUnauthorized)
        return
    }
}
```

---

#### 3. **Cookie Security** (MEDIUM PRIORITY)

**Current Issue**: Cookie is set without security flags

**Recommended Fix**:
```go
// When setting auth cookie in your login handler:

http.SetCookie(w, &http.Cookie{
    Name:     "auth",
    Value:    tokenString,
    Path:     "/",
    HttpOnly: true,      // Prevent JS access (XSS protection)
    Secure:   isProduction, // HTTPS only in production
    SameSite: http.SameSiteLaxMode, // CSRF protection
    MaxAge:   86400 * 7, // 7 days
})
```

---

#### 4. **Algorithm Configuration** (MEDIUM PRIORITY)

**Current**: Uses HS256 (HMAC-SHA256) - suitable for symmetric secrets

**Production Recommendation**: Consider RS256 (RSA) for microservices
```go
// More secure for multi-service architectures
// Public key for verification, private key for signing
// Key management is more complex but more secure
```

---

#### 5. **Rate Limiting** (HIGH PRIORITY)

**Current**: No rate limiting on login/token endpoints

**Add Rate Limiter**:
```bash
go get github.com/didip/tollbooth
```

```go
import "github.com/didip/tollbooth"

lmt := tollbooth.NewLimiter(5, time.Minute) // 5 requests per minute

func LoginHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        err := tollbooth.LimitByRequest(lmt, w, r)
        if err != nil {
            http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

---

#### 6. **CORS Configuration** (MEDIUM PRIORITY)

**Recommended**:
```go
import "github.com/rs/cors"

// Restrict CORS to your frontend domain
handler := cors.New(cors.Options{
    AllowedOrigins: []string{cfg.FrontendURL}, // Only your domain
    AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
    AllowedHeaders: []string{"Content-Type", "Authorization"},
    AllowCredentials: true,
    MaxAge: 3600,
}).Handler(router)
```

---

#### 7. **Logging & Monitoring** (HIGH PRIORITY)

**Do NOT log**:
- ❌ Full JWT tokens
- ❌ Passwords or secrets
- ❌ Personal user data

**DO log**:
- ✅ Failed auth attempts with user ID
- ✅ Unusual access patterns
- ✅ Rate limit violations

```go
// Good logging example
log.Printf("auth_failure user_id=%d ip=%s", userID, r.RemoteAddr)

// Bad logging example - DO NOT DO THIS
log.Printf("token=%s secret=%s", token, jwtSecret) // NEVER!
```

---

## Secrets Management by Environment

### Development
```bash
# Simple secret for local development
JWT_SECRET=dev-secret-12345-for-local-testing-only
# Note: No special requirements, simple to test
```

### Test/CI
```bash
# Fixed secret for reproducible tests
JWT_SECRET=test-secret-fixed-for-ci-pipeline-12345
# Regenerate: openssl rand -hex 32
```

### Production
```bash
# CRITICAL: Strong random secret
JWT_SECRET=$(openssl rand -hex 32)

# Store in secure vault:
# - AWS Secrets Manager: production/api/jwt-secret
# - Kubernetes Secrets: api-jwt-secret
# - HashiCorp Vault: secret/prod/jwt-secret
```

---

## Checklist for Production Deployment

- [ ] JWT_SECRET is strong (64+ hex characters)
- [ ] Different JWT_SECRET for each environment
- [ ] Rate limiting enabled on auth endpoints
- [ ] CORS configured to specific domain
- [ ] HTTPS enforced (Secure cookie flag)
- [ ] HttpOnly flag set on auth cookie
- [ ] SameSite cookie policy configured
- [ ] Token expiration implemented
- [ ] Failed auth attempts logged
- [ ] Secrets never logged
- [ ] Regular security audits scheduled
- [ ] Dependencies kept updated (`go mod tidy`, `npm audit fix`)

---

## Related Files

- **Auth Middleware**: `apps/api/internal/middleware/auth.go`
- **Config Validation**: `apps/api/internal/config/config.go`
- **Environment Setup**: `ENV_SETUP.md`

## Additional Resources

- [OWASP JWT Best Practices](https://cheatsheetseries.owasp.org/cheatsheets/JSON_Web_Token_for_Java_Cheat_Sheet.html)
- [JWT.io Documentation](https://jwt.io/)
- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
