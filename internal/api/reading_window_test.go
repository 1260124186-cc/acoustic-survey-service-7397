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

func TestReadingCapturedBeforeActivationIsRejected(t *testing.T) {
	bands := catalog.NewDefault()
	store := survey.NewStore()
	notices := alert.NewService()
	handler := api.NewServer(
		survey.NewService(store, bands),
		reading.NewService(bands, notices, store),
		notices,
		report.NewService(bands),
		bands,
	).Handler()

	create := httptest.NewRecorder()
	handler.ServeHTTP(create, httptest.NewRequest(
		http.MethodPost,
		"/v1/surveys",
		bytes.NewBufferString(`{"id":"window-line","area":"bay","band":"coastal-38khz"}`),
	))
	if create.Code != http.StatusCreated {
		t.Fatalf("create status=%d", create.Code)
	}
	activate := httptest.NewRecorder()
	handler.ServeHTTP(activate, httptest.NewRequest(http.MethodPost, "/v1/surveys/window-line/activate", nil))
	if activate.Code != http.StatusOK {
		t.Fatalf("activate status=%d", activate.Code)
	}

	capturedAt := time.Now().UTC().Add(-time.Minute).Format(time.RFC3339Nano)
	body, _ := json.Marshal(map[string]any{
		"frequency_hz": 38000,
		"echo_db":      -48,
		"noise_db":     -70,
		"depth_m":      30,
		"captured_at":  capturedAt,
	})
	readingResponse := httptest.NewRecorder()
	handler.ServeHTTP(readingResponse, httptest.NewRequest(
		http.MethodPost,
		"/v1/surveys/window-line/readings",
		bytes.NewReader(body),
	))
	if readingResponse.Code == http.StatusCreated {
		t.Fatal("reading captured before activation was accepted")
	}

	summaryResponse := httptest.NewRecorder()
	handler.ServeHTTP(summaryResponse, httptest.NewRequest(http.MethodGet, "/v1/surveys/window-line/summary", nil))
	var summary struct {
		TotalReadings int `json:"total_readings"`
	}
	if err := json.NewDecoder(summaryResponse.Body).Decode(&summary); err != nil {
		t.Fatal(err)
	}
	if summary.TotalReadings != 0 {
		t.Fatalf("summary total_readings=%d, want 0", summary.TotalReadings)
	}
}
