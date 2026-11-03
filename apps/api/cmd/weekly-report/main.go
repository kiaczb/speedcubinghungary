package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"speedcubinghungary/internal/config"
	"speedcubinghungary/internal/db"
	"speedcubinghungary/internal/weeklyreport"
)

func main() {
	force := flag.Bool("force", false, "A hét meglévő riport-posztját felülírja (alapból a hét első riportja marad meg).")
	skipEmail := flag.Bool("skip-email", false, "Kihagyja az e-mail küldést (teszteléshez, SMTP nélkül).")
	dryRun := flag.Bool("dry-run", false, "Az adatbázisba semmit nem ír (számlálók, poszt, rekordok csak a logban); az e-mail [TESZT] tárggyal elmegy.")
	countsOnly := flag.Bool("counts-only", false, "Csak a verseny-, delegált- és kirakásszámokat és a rekordokat frissíti a WCA-ból (nincs poszt, e-mail).")
	flag.Parse()

	cfg, err := config.LoadWeeklyReport(!*skipEmail && !*countsOnly)
	if err != nil {
		log.Fatal("Config error: ", err)
	}

	dbConn, err := db.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("DB error: ", err)
	}
	defer dbConn.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var repo weeklyreport.Repository = weeklyreport.NewPostgresRepository(dbConn)
	var mailer weeklyreport.Mailer = weeklyreport.NewSMTPMailer(cfg.EmailFrom, cfg.EmailAppPassword, cfg.EmailTo)
	if *dryRun {
		repo = weeklyreport.NewDryRunRepository(repo)
		mailer = weeklyreport.WithSubjectPrefix(mailer, "[TESZT] ")
	}

	service := weeklyreport.NewService(repo, weeklyreport.NewWCAClient(), mailer)

	if err := service.Run(ctx, weeklyreport.Options{Force: *force, SkipEmail: *skipEmail, CountsOnly: *countsOnly, DryRun: *dryRun}); err != nil {
		log.Printf("heti riport: sikertelen: %v", err)
		dbConn.Close()
		os.Exit(1)
	}
	log.Printf("heti riport: kész")
}
