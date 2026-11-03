package config

import (
	"fmt"
	"os"
)

type WeeklyReportConfig struct {
	DatabaseURL      string
	EmailFrom        string
	EmailTo          string
	EmailAppPassword string
}

func LoadWeeklyReport(requireEmail bool) (*WeeklyReportConfig, error) {
	cfg := &WeeklyReportConfig{
		DatabaseURL:      os.Getenv("DATABASE_URL"),
		EmailFrom:        os.Getenv("EMAIL_FROM"),
		EmailTo:          os.Getenv("EMAIL_TO"),
		EmailAppPassword: os.Getenv("EMAIL_APP_PASSWORD"),
	}

	required := []struct{ name, value string }{{"DATABASE_URL", cfg.DatabaseURL}}
	if requireEmail {
		required = append(required,
			struct{ name, value string }{"EMAIL_FROM", cfg.EmailFrom},
			struct{ name, value string }{"EMAIL_TO", cfg.EmailTo},
			struct{ name, value string }{"EMAIL_APP_PASSWORD", cfg.EmailAppPassword},
		)
	}

	var missing []string
	for _, r := range required {
		if r.value == "" {
			missing = append(missing, r.name)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %v", missing)
	}

	return cfg, nil
}
