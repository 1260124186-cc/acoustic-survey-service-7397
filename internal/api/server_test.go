package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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

func TestLateArrivalUsesChronologicalPredecessorForAlerts(t *testing.T) {
	bands := catalog.NewDefault()
	store := survey.NewStore()
	notices := alert.NewService()
	handler := api.NewServer(survey.NewService(store, bands), reading.NewService(bands, notices, store), notices, report.NewService(bands), bands).Handler()
	for index, body := range []string{`{"id":"timing-line","area":"bay","band":"coastal-38khz"}`, ""} {
		path := "/v1/surveys"
		if index == 1 {
			path = "/v1/surveys/timing-line/activate"
		}
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body)))
		if response.Code != http.StatusCreated && response.Code != http.StatusOK {
			t.Fatalf("setup status = %d", response.Code)
		}
	}
	for _, body := range []string{
		`{"frequency_hz":38000,"echo_db":-48,"noise_db":-70,"depth_m":30,"captured_at":"2026-08-19T10:00:00Z"}`,
		`{"frequency_hz":38000,"echo_db":-20,"noise_db":-70,"depth_m":30,"captured_at":"2026-08-19T10:20:00Z"}`,
		`{"frequency_hz":38000,"echo_db":-47,"noise_db":-70,"depth_m":30,"captured_at":"2026-08-19T10:10:00Z"}`,
	} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/v1/surveys/timing-line/readings", bytes.NewBufferString(body)))
		if response.Code != http.StatusCreated {
			t.Fatalf("reading status = %d", response.Code)
		}
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/surveys/timing-line/alerts", nil))
	var result struct {
		Alerts []json.RawMessage `json:"alerts"`
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if len(result.Alerts) != 1 {
		t.Fatalf("alerts = %d, want 1", len(result.Alerts))
	}
	var alert struct {
		Rule string `json:"rule"`
	}
	if err := json.Unmarshal(result.Alerts[0], &alert); err != nil {
		t.Fatal(err)
	}
	if alert.Rule != "abrupt_normalized_change" {
		t.Fatalf("rule = %s, want abrupt_normalized_change", alert.Rule)
	}
}
