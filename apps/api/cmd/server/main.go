package main

import (
	"log"
	"net/http"
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
