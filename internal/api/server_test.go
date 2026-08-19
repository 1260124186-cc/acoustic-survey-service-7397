package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"example.com/acoustic-survey-service/internal/alert"
	"example.com/acoustic-survey-service/internal/api"
	"example.com/acoustic-survey-service/internal/catalog"
	"example.com/acoustic-survey-service/internal/reading"
	"example.com/acoustic-survey-service/internal/report"
	"example.com/acoustic-survey-service/internal/survey"
)

func TestLifecycle(t *testing.T) {
	bands := catalog.NewDefault()
	store := survey.NewStore()
	notices := alert.NewService()
	handler := api.NewServer(survey.NewService(store, bands), reading.NewService(bands, notices, store), notices, report.NewService(bands), bands).Handler()
	create := httptest.NewRequest(http.MethodPost, "/v1/surveys", bytes.NewBufferString(`{"id":"line-a","area":"bay","band":"coastal-38khz"}`))
	create.Header.Set("Content-Type", "application/json")
	created := httptest.NewRecorder()
	handler.ServeHTTP(created, create)
	if created.Code != http.StatusCreated {
		t.Fatalf("create status = %d", created.Code)
	}
	active := httptest.NewRecorder()
	handler.ServeHTTP(active, httptest.NewRequest(http.MethodPost, "/v1/surveys/line-a/activate", nil))
	if active.Code != http.StatusOK {
		t.Fatalf("activate status = %d", active.Code)
	}
	readingRequest := httptest.NewRequest(http.MethodPost, "/v1/surveys/line-a/readings", bytes.NewBufferString(`{"frequency_hz":38000,"echo_db":-48,"noise_db":-70,"depth_m":30}`))
	read := httptest.NewRecorder()
	handler.ServeHTTP(read, readingRequest)
	if read.Code != http.StatusCreated {
		t.Fatalf("reading status = %d", read.Code)
	}
}

func TestAlertsAreOrderedByCaptureTimeAfterLateArrival(t *testing.T) {
	bands := catalog.NewDefault()
	store := survey.NewStore()
	notices := alert.NewService()
	handler := api.NewServer(survey.NewService(store, bands), reading.NewService(bands, notices, store), notices, report.NewService(bands), bands).Handler()
	for index, body := range []string{`{"id":"alert-order","area":"bay","band":"coastal-38khz"}`, ""} {
		path := "/v1/surveys"
		if index == 1 {
			path = "/v1/surveys/alert-order/activate"
		}
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body)))
		if response.Code != http.StatusCreated && response.Code != http.StatusOK {
			t.Fatalf("setup status = %d", response.Code)
		}
	}
	for _, body := range []string{
		`{"frequency_hz":42000,"echo_db":-48,"noise_db":-70,"depth_m":30,"captured_at":"2026-08-19T10:20:00Z"}`,
		`{"frequency_hz":42000,"echo_db":-48,"noise_db":-70,"depth_m":30,"captured_at":"2026-08-19T10:00:00Z"}`,
	} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/v1/surveys/alert-order/readings", bytes.NewBufferString(body)))
		if response.Code != http.StatusCreated {
			t.Fatalf("reading status = %d", response.Code)
		}
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/surveys/alert-order/alerts", nil))
	var result struct {
		Alerts []struct {
			CreatedAt time.Time `json:"created_at"`
		} `json:"alerts"`
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if len(result.Alerts) < 2 {
		t.Fatalf("alerts = %d, want at least 2", len(result.Alerts))
	}
	if result.Alerts[1].CreatedAt.Before(result.Alerts[0].CreatedAt) {
		t.Fatal("alerts are not chronological")
	}
}
