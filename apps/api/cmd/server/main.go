package main

import (
	"log"
	"net/http"
	"speedcubinghungary/internal/allrounder"
	"speedcubinghungary/internal/auth"
	"speedcubinghungary/internal/certificate"
	"speedcubinghungary/internal/competition"
	"speedcubinghungary/internal/config"
	"speedcubinghungary/internal/db"
	"speedcubinghungary/internal/middleware"
	"speedcubinghungary/internal/post"
	"speedcubinghungary/internal/prediction"
	"speedcubinghungary/internal/router"
	"speedcubinghungary/internal/special"
	"speedcubinghungary/internal/stats"
	"speedcubinghungary/internal/wcalive"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Config error:", err)
	}

	dbConn, err := db.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("DB error:", err)
	}
	defer dbConn.Close()

	m, err := migrate.New("file:///migrations", cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Migration init error:", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal("Migration run error:", err)
	}

	specialLive := special.NewLiveClient(cfg.WCALiveURL)
	specialBroadcaster := wcalive.NewBroadcaster()
	specialTracker := special.NewTracker(specialLive, wcalive.NewSocketClient(cfg.WCALiveURL), specialBroadcaster)
	specialHandler := special.NewHandler(
		special.NewService(special.NewPostgresRepository(dbConn), specialLive, specialTracker, specialBroadcaster),
	)

	allrounderLive := allrounder.NewLiveClient(cfg.WCALiveURL)
	allrounderBroadcaster := wcalive.NewBroadcaster()
	allrounderTracker := allrounder.NewLiveTracker(allrounderLive, wcalive.NewSocketClient(cfg.WCALiveURL), allrounderBroadcaster)
	allrounderHandler := allrounder.NewHandler(allrounder.NewService(
		allrounder.NewPostgresRepository(dbConn),
		allrounder.NewWCAClient(),
		allrounderLive,
		allrounderTracker,
		allrounderBroadcaster,
	))

	postHandler := post.NewHandler(post.NewService(post.NewPostgresRepository(dbConn)), cfg.UploadDir)

	competitionService := competition.NewService(competition.NewWCARepository())
	competitionHandler := competition.NewHandler(competitionService)

	predictionHandler := prediction.NewHandler(prediction.NewService(
		prediction.NewPostgresRepository(dbConn),
		competitionService,
	))

	authService := auth.NewService(auth.NewPostgresRepository(dbConn), cfg)

	statsService := stats.NewService(competitionService, stats.NewLiveRepository(cfg.WCALiveURL), authService)
	statsHandler := stats.NewHandler(statsService)

	certificateHandler := certificate.NewHandler(certificate.NewService(
		certificate.NewPostgresRepository(dbConn),
		competitionService,
		statsService,
	))

	authHandler := auth.NewHandler(authService, cfg)
	authMiddleware := middleware.AuthMiddlewareWithConfig(cfg, authService)

	r := router.New(cfg, authMiddleware, postHandler, competitionHandler, predictionHandler, authHandler, specialHandler, allrounderHandler, statsHandler, certificateHandler)

	log.Printf("Server started on :8080 (mode: %s)", cfg.Deployment)
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", r))
}
