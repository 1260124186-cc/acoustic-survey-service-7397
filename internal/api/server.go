package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"example.com/acoustic-survey-service/internal/alert"
	"example.com/acoustic-survey-service/internal/catalog"
	"example.com/acoustic-survey-service/internal/model"
	"example.com/acoustic-survey-service/internal/reading"
	"example.com/acoustic-survey-service/internal/report"
	"example.com/acoustic-survey-service/internal/survey"
)

type Server struct {
	surveys  *survey.Service
	readings *reading.Service
	alerts   *alert.Service
	reports  *report.Service
	catalog  *catalog.Catalog
}

type createSurveyRequest struct {
	ID   string `json:"id"`
	Area string `json:"area"`
	Band string `json:"band"`
}

func NewServer(surveys *survey.Service, readings *reading.Service, alerts *alert.Service, reports *report.Service, catalog *catalog.Catalog) *Server {
	return &Server{surveys: surveys, readings: readings, alerts: alerts, reports: reports, catalog: catalog}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("GET /v1/bands", s.bands)
	mux.HandleFunc("GET /v1/surveys", s.listSurveys)
	mux.HandleFunc("POST /v1/surveys", s.createSurvey)
	mux.HandleFunc("GET /v1/surveys/{id}", s.getSurvey)
	mux.HandleFunc("POST /v1/surveys/{id}/activate", s.activateSurvey)
	mux.HandleFunc("POST /v1/surveys/{id}/readings", s.createReading)
	mux.HandleFunc("POST /v1/surveys/{id}/close", s.closeSurvey)
	mux.HandleFunc("GET /v1/surveys/{id}/summary", s.summary)
	mux.HandleFunc("GET /v1/surveys/{id}/alerts", s.listAlerts)
	return logging(mux)
}

func (s *Server) health(response http.ResponseWriter, request *http.Request) {
	writeJSON(response, http.StatusOK, map[string]string{"status": "ok", "service": "acoustic-survey"})
}

func (s *Server) bands(response http.ResponseWriter, request *http.Request) {
	writeJSON(response, http.StatusOK, map[string]any{"bands": s.catalog.List()})
}

func (s *Server) createSurvey(response http.ResponseWriter, request *http.Request) {
	var body createSurveyRequest
	if err := decode(request, &body); err != nil {
		writeJSON(response, http.StatusBadRequest, model.Error{Code: "invalid_json", Message: "request body must be valid JSON"})
		return
	}
	value, err := s.surveys.Create(body.ID, body.Area, body.Band)
	if err != nil {
		writeError(response, err)
		return
	}
	writeJSON(response, http.StatusCreated, value)
}

func (s *Server) listSurveys(response http.ResponseWriter, request *http.Request) {
	writeJSON(response, http.StatusOK, map[string]any{"surveys": s.surveys.List()})
}

func (s *Server) getSurvey(response http.ResponseWriter, request *http.Request) {
	value, err := s.surveys.Get(id(request))
	if err != nil {
		writeError(response, err)
		return
	}
	writeJSON(response, http.StatusOK, value)
}

func (s *Server) activateSurvey(response http.ResponseWriter, request *http.Request) {
	value, err := s.surveys.Activate(id(request))
	if err != nil {
		writeError(response, err)
		return
	}
	writeJSON(response, http.StatusOK, value)
}

func (s *Server) createReading(response http.ResponseWriter, request *http.Request) {
	var body model.ReadingInput
	if err := decode(request, &body); err != nil {
		writeJSON(response, http.StatusBadRequest, model.Error{Code: "invalid_json", Message: "request body must be valid JSON"})
		return
	}
	value, err := s.readings.Add(id(request), body)
	if err != nil {
		writeError(response, err)
		return
	}
	writeJSON(response, http.StatusCreated, value)
}

func (s *Server) closeSurvey(response http.ResponseWriter, request *http.Request) {
	value, err := s.surveys.Close(id(request))
	if err != nil {
		writeError(response, err)
		return
	}
	writeJSON(response, http.StatusOK, value)
}

func (s *Server) summary(response http.ResponseWriter, request *http.Request) {
	value, err := s.reports.Summary(s.surveys.Store(), s.alerts, id(request))
	if err != nil {
		writeError(response, err)
		return
	}
	writeJSON(response, http.StatusOK, value)
}

func (s *Server) listAlerts(response http.ResponseWriter, request *http.Request) {
	if _, err := s.surveys.Get(id(request)); err != nil {
		writeError(response, err)
		return
	}
	writeJSON(response, http.StatusOK, map[string]any{"alerts": s.alerts.List(id(request))})
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		started := time.Now()
		next.ServeHTTP(response, request)
		log.Printf("%s %s completed in %s", request.Method, request.URL.Path, time.Since(started).Round(time.Millisecond))
	})
}

func id(request *http.Request) string { return strings.TrimSpace(request.PathValue("id")) }

func decode(request *http.Request, target any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(nil, request.Body, 1<<20))
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func writeJSON(response http.ResponseWriter, status int, value any) {
	response.Header().Set("Content-Type", "application/json; charset=utf-8")
	response.WriteHeader(status)
	_ = json.NewEncoder(response).Encode(value)
}

func writeError(response http.ResponseWriter, err error) {
	value, ok := err.(*model.Error)
	if !ok {
		writeJSON(response, http.StatusInternalServerError, model.Error{Code: "internal", Message: "internal server error"})
		return
	}
	status := http.StatusBadRequest
	if value.Code == "not_found" {
		status = http.StatusNotFound
	}
	if value.Code == "conflict" {
		status = http.StatusConflict
	}
	writeJSON(response, status, value)
}
