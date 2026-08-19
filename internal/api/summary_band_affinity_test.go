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

func TestSummaryKeepsSurveyBandWhenCatalogRangesOverlap(t *testing.T) {
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
	if got := serve(http.MethodPost, "/v1/surveys", `{"id":"overlap-line","area":"outer-bay","band":"coastal-38khz"}`); got.Code != http.StatusCreated {
		t.Fatalf("create=%d", got.Code)
	}
	if got := serve(http.MethodPost, "/v1/surveys/overlap-line/activate", ""); got.Code != http.StatusOK {
		t.Fatalf("activate=%d", got.Code)
	}
	if got := serve(http.MethodPost, "/v1/surveys/overlap-line/readings", `{"frequency_hz":37500,"echo_db":-48,"noise_db":-80,"depth_m":30}`); got.Code != http.StatusCreated {
		t.Fatalf("reading=%d", got.Code)
	}
	response := serve(http.MethodGet, "/v1/surveys/overlap-line/summary", "")
	if response.Code != http.StatusOK {
		t.Fatalf("summary=%d", response.Code)
	}
	var summary model.Summary
	if err := json.Unmarshal(response.Body.Bytes(), &summary); err != nil {
		t.Fatal(err)
	}
	if got := summary.BandCounts["coastal-38khz"]; got != 1 {
		t.Fatalf("coastal band count=%d, summary=%+v", got, summary.BandCounts)
	}
	if got := summary.BandCounts["sector-02-window-10"]; got != 0 {
		t.Fatalf("overlapping profile was selected: %+v", summary.BandCounts)
	}
}
