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

func TestCalibrationBoundaryDoesNotCreateAlert(t *testing.T) {
	bands := catalog.NewDefault()
	store := survey.NewStore()
	notices := alert.NewService()
	handler := api.NewServer(survey.NewService(store, bands), reading.NewService(bands, notices, store), notices, report.NewService(bands), bands).Handler()
	for index, body := range []string{`{"id":"boundary-line","area":"bay","band":"coastal-38khz"}`, ""} {
		path := "/v1/surveys"
		if index == 1 {
			path = "/v1/surveys/boundary-line/activate"
		}
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body)))
		if response.Code != http.StatusCreated && response.Code != http.StatusOK {
			t.Fatalf("setup status = %d", response.Code)
		}
	}
	readingResponse := httptest.NewRecorder()
	handler.ServeHTTP(readingResponse, httptest.NewRequest(http.MethodPost, "/v1/surveys/boundary-line/readings", bytes.NewBufferString(`{"frequency_hz":37000,"echo_db":-48,"noise_db":-70,"depth_m":30}`)))
	if readingResponse.Code != http.StatusCreated {
		t.Fatalf("reading status = %d", readingResponse.Code)
	}
	alertsResponse := httptest.NewRecorder()
	handler.ServeHTTP(alertsResponse, httptest.NewRequest(http.MethodGet, "/v1/surveys/boundary-line/alerts", nil))
	var result struct {
		Alerts []json.RawMessage `json:"alerts"`
	}
	if err := json.NewDecoder(alertsResponse.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if len(result.Alerts) != 0 {
		t.Fatalf("alerts = %d, want 0", len(result.Alerts))
	}
}
