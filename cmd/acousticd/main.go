package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example.com/acoustic-survey-service/internal/alert"
	"example.com/acoustic-survey-service/internal/api"
	"example.com/acoustic-survey-service/internal/catalog"
	"example.com/acoustic-survey-service/internal/reading"
	"example.com/acoustic-survey-service/internal/report"
	"example.com/acoustic-survey-service/internal/survey"
)

func main() {
	address := os.Getenv("ACOUSTIC_ADDR")
	if address == "" {
		address = "127.0.0.1:18087"
	}
	bands := catalog.NewDefault()
	store := survey.NewStore()
	notices := alert.NewService()
	readings := reading.NewService(bands, notices, store)
	surveys := survey.NewService(store, bands)
	reports := report.NewService(bands)
	server := &http.Server{Addr: address, Handler: api.NewServer(surveys, readings, notices, reports, bands).Handler(), ReadHeaderTimeout: 5 * time.Second}
	go func() {
		log.Printf("acoustic survey service listening on %s", address)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	context, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	if err := server.Shutdown(context); err != nil {
		log.Printf("shutdown failed: %v", err)
	}
}
