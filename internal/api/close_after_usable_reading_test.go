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

// TestCloseAllowedAfterUsableReading 补采到一条可用读数后即可结束测线，
// 并且结束后测线被锁定、不再接受新读数。
func TestCloseAllowedAfterUsableReading(t *testing.T) {
	bands := catalog.NewDefault()
	store := survey.NewStore()
	notices := alert.NewService()
	handler := api.NewServer(survey.NewService(store, bands), reading.NewService(bands, notices, store), notices, report.NewService(bands), bands).Handler()

	request := func(method string, path string, body string) *httptest.ResponseRecorder {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(method, path, bytes.NewBufferString(body)))
		return recorder
	}

	if response := request(http.MethodPost, "/v1/surveys", `{"id":"quality-recollect","area":"bay","band":"coastal-38khz"}`); response.Code != http.StatusCreated {
		t.Fatalf("create status = %d", response.Code)
	}
	if response := request(http.MethodPost, "/v1/surveys/quality-recollect/activate", ""); response.Code != http.StatusOK {
		t.Fatalf("activate status = %d", response.Code)
	}

	// 设备调频失误：频率超出标定范围，读数不可用。
	if response := request(http.MethodPost, "/v1/surveys/quality-recollect/readings", `{"frequency_hz":42000,"echo_db":-48,"noise_db":-70,"depth_m":30}`); response.Code != http.StatusCreated {
		t.Fatalf("first reading status = %d", response.Code)
	}
	// 此时全部读数不可靠，结束测线应被拒绝，保持可继续补采。
	if response := request(http.MethodPost, "/v1/surveys/quality-recollect/close", ""); response.Code != http.StatusBadRequest {
		t.Fatalf("close before usable status = %d, want %d", response.Code, http.StatusBadRequest)
	}

	// 补采一条落在标定范围内的可用读数。
	if response := request(http.MethodPost, "/v1/surveys/quality-recollect/readings", `{"frequency_hz":38000,"echo_db":-48,"noise_db":-70,"depth_m":30}`); response.Code != http.StatusCreated {
		t.Fatalf("usable reading status = %d", response.Code)
	}

	// 至少有一条可用读数后，结束测线成功。
	if response := request(http.MethodPost, "/v1/surveys/quality-recollect/close", ""); response.Code != http.StatusOK {
		t.Fatalf("close after usable status = %d, want %d", response.Code, http.StatusOK)
	}

	// 结束后测线锁定，不再接受新读数。
	if response := request(http.MethodPost, "/v1/surveys/quality-recollect/readings", `{"frequency_hz":38000,"echo_db":-48,"noise_db":-70,"depth_m":30}`); response.Code != http.StatusBadRequest {
		t.Fatalf("reading after close status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}
