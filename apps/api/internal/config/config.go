package config

import (
	"fmt"
	"os"
)

type Config struct {
	Deployment      string
	FrontendURL     string
	DatabaseURL     string
	JWTSecret       string
	WCAClientID     string
	WCAClientSecret string
	WCARedirectURI  string
	// UploadDir a feltöltött képek könyvtára; ide egy named volume van csatolva
	// (lásd infra/docker-compose.yml), hogy a fájlok túléljék az újraépítést.
	UploadDir  string
	WCALiveURL string
}

func Load() (*Config, error) {
	cfg := &Config{
		// ⚠️ DEPLOYMENT és FRONTEND_URL szándékosan DEFAULT NÉLKÜL töltődik: a
		// DEPLOYMENT szabja meg az auth cookie `Secure` flagjét, a FRONTEND_URL a
		// CORS origint és a login utáni átirányítás célját. Default mellett egy
		// prodban hiányzó DEPLOYMENT nem leállást, hanem némán Secure nélküli
		// sütit eredményezne.
		Deployment:      os.Getenv("DEPLOYMENT"),
		FrontendURL:     os.Getenv("FRONTEND_URL"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		JWTSecret:       os.Getenv("JWT_SECRET"),
		WCAClientID:     os.Getenv("WCA_CLIENT_ID"),
		WCAClientSecret: os.Getenv("WCA_CLIENT_SECRET"),
		WCARedirectURI:  os.Getenv("WCA_REDIRECT_URI"),
		UploadDir:       getEnv("UPLOAD_DIR", "/data/uploads"),
		WCALiveURL:      getEnv("WCA_LIVE_URL", "https://live.worldcubeassociation.org"),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	required := map[string]string{
		"DATABASE_URL":      c.DatabaseURL,
		"JWT_SECRET":        c.JWTSecret,
		"WCA_CLIENT_ID":     c.WCAClientID,
		"WCA_CLIENT_SECRET": c.WCAClientSecret,
		"WCA_REDIRECT_URI":  c.WCARedirectURI,
		"FRONTEND_URL":      c.FrontendURL,
		"DEPLOYMENT":        c.Deployment,
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

	// Egy elgépelt "produciton" nem üres, tehát a fenti ellenőrzésen átcsúszna,
	// és némán fejlesztői módként viselkedne.
	switch c.Deployment {
	case "development", "test", "production":
	default:
		return fmt.Errorf("DEPLOYMENT must be one of development|test|production, got %q", c.Deployment)
	}

	if len(c.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters long (use 'openssl rand -hex 32' to generate)")
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (c *Config) IsDevelopment() bool {
	return c.Deployment == "development"
}
