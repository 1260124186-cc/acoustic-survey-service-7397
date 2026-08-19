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

func TestClosedSurveyRejectsLateReading(t *testing.T) {
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

	request := func(method, path, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		return response
	}

	if response := request(http.MethodPost, "/v1/surveys", `{"id":"late-readings","area":"bay","band":"coastal-38khz"}`); response.Code != http.StatusCreated {
		t.Fatalf("create status = %d", response.Code)
	}
	if response := request(http.MethodPost, "/v1/surveys/late-readings/activate", ""); response.Code != http.StatusOK {
		t.Fatalf("activate status = %d", response.Code)
	}
	readingBody := `{"frequency_hz":38000,"echo_db":-48,"noise_db":-70,"depth_m":30}`
	if response := request(http.MethodPost, "/v1/surveys/late-readings/readings", readingBody); response.Code != http.StatusCreated {
		t.Fatalf("initial reading status = %d", response.Code)
	}
	if response := request(http.MethodPost, "/v1/surveys/late-readings/close", ""); response.Code != http.StatusOK {
		t.Fatalf("close status = %d", response.Code)
	}
	late := request(http.MethodPost, "/v1/surveys/late-readings/readings", readingBody)
	if late.Code == http.StatusCreated {
		t.Fatal("closed survey accepted a late reading")
	}

	summary := request(http.MethodGet, "/v1/surveys/late-readings/summary", "")
	if summary.Code != http.StatusOK {
		t.Fatalf("summary status = %d", summary.Code)
	}
	var value struct {
		TotalReadings int `json:"total_readings"`
	}
	if err := json.NewDecoder(summary.Body).Decode(&value); err != nil {
		t.Fatal(err)
	}
	if value.TotalReadings != 1 {
		t.Fatalf("total readings = %d, want 1", value.TotalReadings)
	}
}
