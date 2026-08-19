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

func TestClosedSurveyRetainsRecordedReadingCount(t *testing.T) {
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
	if got := serve(http.MethodPost, "/v1/surveys", `{"id":"close-line","area":"outer-bay","band":"coastal-38khz"}`); got.Code != http.StatusCreated {
		t.Fatalf("create=%d", got.Code)
	}
	if got := serve(http.MethodPost, "/v1/surveys/close-line/activate", ""); got.Code != http.StatusOK {
		t.Fatalf("activate=%d", got.Code)
	}
	if got := serve(http.MethodPost, "/v1/surveys/close-line/readings", `{"frequency_hz":38000,"echo_db":-48,"noise_db":-80,"depth_m":30}`); got.Code != http.StatusCreated {
		t.Fatalf("reading=%d", got.Code)
	}
	if got := serve(http.MethodPost, "/v1/surveys/close-line/close", ""); got.Code != http.StatusOK {
		t.Fatalf("close=%d", got.Code)
	}
	response := serve(http.MethodGet, "/v1/surveys/close-line", "")
	if response.Code != http.StatusOK {
		t.Fatalf("get=%d", response.Code)
	}
	var value model.Survey
	if err := json.Unmarshal(response.Body.Bytes(), &value); err != nil {
		t.Fatal(err)
	}
	if value.ReadingCount != 1 {
		t.Fatalf("closed survey reading_count=%d, want 1", value.ReadingCount)
	}
}
