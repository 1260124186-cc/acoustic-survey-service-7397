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

func TestLateArrivalDoesNotCreateAbruptChangeAlert(t *testing.T) {
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
	if got := serve(http.MethodPost, "/v1/surveys", `{"id":"late-line","area":"outer-bay","band":"coastal-38khz"}`); got.Code != http.StatusCreated {
		t.Fatalf("create=%d", got.Code)
	}
	if got := serve(http.MethodPost, "/v1/surveys/late-line/activate", ""); got.Code != http.StatusOK {
		t.Fatalf("activate=%d", got.Code)
	}
	readings := []string{
		`{"frequency_hz":38000,"echo_db":-75,"noise_db":-100,"depth_m":10,"captured_at":"2026-08-19T10:00:00Z"}`,
		`{"frequency_hz":38000,"echo_db":-55,"noise_db":-100,"depth_m":10,"captured_at":"2026-08-19T10:02:00Z"}`,
		`{"frequency_hz":38000,"echo_db":-65,"noise_db":-100,"depth_m":10,"captured_at":"2026-08-19T10:01:00Z"}`,
	}
	for _, body := range readings {
		if got := serve(http.MethodPost, "/v1/surveys/late-line/readings", body); got.Code != http.StatusCreated {
			t.Fatalf("reading=%d", got.Code)
		}
	}
	response := serve(http.MethodGet, "/v1/surveys/late-line/alerts", "")
	if response.Code != http.StatusOK {
		t.Fatalf("alerts=%d", response.Code)
	}
	var body struct {
		Alerts []model.Alert `json:"alerts"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	for _, notice := range body.Alerts {
		if notice.Rule == "abrupt_normalized_change" {
			t.Fatalf("arrival order created a false abrupt alert: %+v", notice)
		}
	}
}
