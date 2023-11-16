package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/timohahaa/device-service/config"
	v1 "github.com/timohahaa/device-service/internal/controllers/http/v1"
	"github.com/timohahaa/device-service/internal/filescanner"
	"github.com/timohahaa/postgres"
)

func main() {
	slog.Info("reading config...")
	config, err := config.NewConfig("config/config.yaml")
	if err != nil {
		panic(err)
	}

	slog.Info("validating scanner duration...")
	dur, err := time.ParseDuration(config.Scanner.RescanDurationInSeconds)
	if err != nil {
		panic(err)
	}

	slog.Info("connecting to postgres...")
	dbUrl := fmt.Sprintf("postgres://%v:%v@%v:%v/%v",
		config.Database.Username,
		config.Database.Password,
		config.Database.Host,
		config.Database.Port,
		config.Database.DatabaseName,
	)
	pg, err := postgres.New(dbUrl, postgres.MaxConnPoolSize(5))
	if err != nil {
		panic(err)
	}

	slog.Info("initializing file scanner...")
	scanner := filescanner.NewScanner(pg, dur, config.Scanner.InputDirectoryAbsolutePath, config.Scanner.OutputDirectoryAbsolutePath)
	scannerStop := make(chan struct{})
	scanner.Start(scannerStop)

	slog.Info("starting http server...")
	handler := v1.NewV1Handler(pg)
	s := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}
	go func() {
		s.ListenAndServe()
	}()

	slog.Info("configuring gracefull shutdown...")
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	<-shutdown

	slog.Info("shutting down the service...")
	scannerStop <- struct{}{}
	s.Shutdown(context.Background())

	slog.Info("exited succesfully!")
}
