package main

import (
	"log"
	"net/http"
	"speedcubinghungary/internal/config"
	"speedcubinghungary/internal/db"
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

	// 3. Budapest Special: WCA Live alapú eredménytábla. A versenyt a NÉZŐ
	//    választja ki kérésenként — nincs szerveroldali "figyelt verseny". A
	//    hivatalos eredményt NEM tároljuk: a tracker a memóriában tartja, és
	//    amíg legalább egy nézet nyitva van, a WCA Live subscription tartja
	//    frissen. A tracker és a HTTP réteg ugyanazon a Broadcasteren osztozik,
	//    hogy a beérkező frissítésekről a megnyitott felületek azonnal (SSE-n)
	//    értesüljenek.
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

	// 4. Create router with config
	r := router.New(dbConn, cfg, specialHandler)

	log.Printf("Server started on :8080 (mode: %s)", cfg.Deployment)
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", r))
}
