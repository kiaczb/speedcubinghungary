package config

import (
	"fmt"
	"os"
)

// Config holds all environment configuration
type Config struct {
	Deployment      string
	FrontendURL     string
	DatabaseURL     string
	JWTSecret       string
	WCAClientID     string
	WCAClientSecret string
	WCARedirectURI  string
	LogLevel        string
	// UploadDir a feltöltött képek könyvtára a konténerben. Ide egy named volume
	// van csatolva (lásd infra/docker-compose.yml), hogy a fájlok túléljék a
	// konténer újraépítését. Nem kötelező env — az alapértelmezés a szokásos út.
	UploadDir string
	// WCALiveURL a WCA Live példány alap-URL-je (GraphQL + Phoenix socket).
	// Nem kötelező env; teszteléshez lehet saját példányra állítani.
	WCALiveURL string
}

// Load loads and validates environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Deployment:      getEnv("DEPLOYMENT", "development"),
		FrontendURL:     getEnv("FRONTEND_URL", "http://localhost:3000"),
		DatabaseURL:     getEnv("DATABASE_URL", ""),
		JWTSecret:       getEnv("JWT_SECRET", ""),
		WCAClientID:     getEnv("WCA_CLIENT_ID", ""),
		WCAClientSecret: getEnv("WCA_CLIENT_SECRET", ""),
		WCARedirectURI:  getEnv("WCA_REDIRECT_URI", ""),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		UploadDir:       getEnv("UPLOAD_DIR", "/data/uploads"),
		WCALiveURL:      getEnv("WCA_LIVE_URL", "https://live.worldcubeassociation.org"),
	}

	// Validate required variables
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks that all required variables are set
func (c *Config) Validate() error {
	required := map[string]string{
		"DATABASE_URL":      c.DatabaseURL,
		"JWT_SECRET":        c.JWTSecret,
		"WCA_CLIENT_ID":     c.WCAClientID,
		"WCA_CLIENT_SECRET": c.WCAClientSecret,
		"WCA_REDIRECT_URI":  c.WCARedirectURI,
		"FRONTEND_URL":      c.FrontendURL,
	}

	var missing []string
	for name, value := range required {
		if value == "" {
			missing = append(missing, name)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %v\n\nPlease check your .env file or environment setup", missing)
	}

	// Validate JWT_SECRET length (should be at least 32 bytes / 64 hex chars)
	if len(c.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters long (use 'openssl rand -hex 32' to generate)")
	}

	return nil
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Deployment == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Deployment == "production"
}

// IsTest returns true if running in test mode
func (c *Config) IsTest() bool {
	return c.Deployment == "test"
}
