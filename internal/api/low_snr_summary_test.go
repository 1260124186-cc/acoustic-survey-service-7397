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
	"example.com/acoustic-survey-service/internal/model"
	"example.com/acoustic-survey-service/internal/reading"
	"example.com/acoustic-survey-service/internal/report"
	"example.com/acoustic-survey-service/internal/survey"
)

func TestSummaryExcludesLowSignalToNoiseReadingFromValidCount(t *testing.T) {
	bands := catalog.NewDefault()
	store := survey.NewStore()
	notices := alert.NewService()
	handler := api.NewServer(survey.NewService(store, bands), reading.NewService(bands, notices, store), notices, report.NewService(bands), bands).Handler()
	serve := func(method, path, body string) *httptest.ResponseRecorder {
		request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		return response
	}
	if got := serve(http.MethodPost, "/v1/surveys", `{"id":"snr-line","area":"outer-bay","band":"coastal-38khz"}`); got.Code != http.StatusCreated {
		t.Fatalf("create=%d", got.Code)
	}
	if got := serve(http.MethodPost, "/v1/surveys/snr-line/activate", ""); got.Code != http.StatusOK {
		t.Fatalf("activate=%d", got.Code)
	}
	if got := serve(http.MethodPost, "/v1/surveys/snr-line/readings", `{"frequency_hz":38000,"echo_db":-60,"noise_db":-65,"depth_m":30}`); got.Code != http.StatusCreated {
		t.Fatalf("reading=%d", got.Code)
	}
	response := serve(http.MethodGet, "/v1/surveys/snr-line/summary", "")
	if response.Code != http.StatusOK {
		t.Fatalf("summary=%d", response.Code)
	}
	var summary model.Summary
	if err := json.Unmarshal(response.Body.Bytes(), &summary); err != nil {
		t.Fatal(err)
	}
	if summary.ValidReadings != 0 {
		t.Fatalf("valid readings=%d, want 0 for low SNR", summary.ValidReadings)
	}
	if summary.OpenAlerts == 0 {
		t.Fatalf("low SNR should remain visible as an alert")
	}
}
