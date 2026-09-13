package main

import (
	"log"
	"net/http"
	"speedcubinghungary/internal/allrounder"
	"speedcubinghungary/internal/auth"
	"speedcubinghungary/internal/competition"
	"speedcubinghungary/internal/config"
	"speedcubinghungary/internal/db"
	"speedcubinghungary/internal/post"
	"speedcubinghungary/internal/prediction"
	"speedcubinghungary/internal/router"
	"speedcubinghungary/internal/special"

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

	specialRepo := special.NewPostgresRepository(dbConn)
	liveClient := special.NewLiveClient(cfg.WCALiveURL)
	specialBroadcaster := special.NewBroadcaster()
	specialTracker := special.NewTracker(
		liveClient,
		special.NewSocketClient(cfg.WCALiveURL),
		specialBroadcaster,
	)
	specialHandler := special.NewHandler(
		special.NewService(specialRepo, liveClient, specialTracker, specialBroadcaster),
	)

	// 4. Hungarian All-Rounder
	allrounderLiveClient := allrounder.NewLiveClient(cfg.WCALiveURL)
	allrounderBroadcaster := allrounder.NewBroadcaster()
	allrounderTracker := allrounder.NewLiveTracker(
		allrounderLiveClient,
		allrounder.NewSocketClient(cfg.WCALiveURL),
		allrounderBroadcaster,
	)
	allrounderHandler := allrounder.NewHandler(allrounder.NewService(
		allrounder.NewPostgresRepository(dbConn),
		allrounder.NewWCAClient(),
		allrounderLiveClient,
		allrounderTracker,
		allrounderBroadcaster,
	))

	// 5. Post feature
	postHandler := post.NewHandler(post.NewService(post.NewPostgresRepository(dbConn)), cfg.UploadDir)

	// 6. Competition feature (WCA API proxy)
	competitionService := competition.NewService(competition.NewWCARepository())
	competitionHandler := competition.NewHandler(competitionService)

	// 7. Tippverseny (a versenyadatokat a competition service-en keresztül kéri le)
	predictionHandler := prediction.NewHandler(prediction.NewService(
		prediction.NewPostgresRepository(dbConn),
		competitionService,
	))

	// 8. Auth (WCA OAuth)
	authHandler := auth.NewHandler(auth.NewService(auth.NewPostgresRepository(dbConn)))

	// 9. Create router with config
	r := router.New(cfg, postHandler, competitionHandler, predictionHandler, authHandler, specialHandler, allrounderHandler)

	log.Printf("Server started on :8080 (mode: %s)", cfg.Deployment)
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", r))
}
