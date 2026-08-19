package api_test

import (
	"bytes"
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

func TestCloseRejectsSurveyWithBlockingAlert(t *testing.T) {
	bands := catalog.NewDefault()
	store := survey.NewStore()
	notices := alert.NewService()
	handler := api.NewServer(survey.NewService(store, bands), reading.NewService(bands, notices, store), notices, report.NewService(bands), bands).Handler()
	request := func(method, path, body string) *httptest.ResponseRecorder {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(method, path, bytes.NewBufferString(body)))
		return recorder
	}
	if response := request(http.MethodPost, "/v1/surveys", `{"id":"blocked-close","area":"bay","band":"coastal-38khz"}`); response.Code != http.StatusCreated { t.Fatalf("create status = %d", response.Code) }
	if response := request(http.MethodPost, "/v1/surveys/blocked-close/activate", ""); response.Code != http.StatusOK { t.Fatalf("activate status = %d", response.Code) }
	if response := request(http.MethodPost, "/v1/surveys/blocked-close/readings", `{"frequency_hz":42000,"echo_db":-48,"noise_db":-70,"depth_m":30}`); response.Code != http.StatusCreated { t.Fatalf("reading status = %d", response.Code) }
	if response := request(http.MethodPost, "/v1/surveys/blocked-close/close", ""); response.Code != http.StatusBadRequest { t.Fatalf("close status = %d", response.Code) }
}
