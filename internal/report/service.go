package report

import (
	"example.com/acoustic-survey-service/internal/alert"
	"example.com/acoustic-survey-service/internal/catalog"
	"example.com/acoustic-survey-service/internal/model"
	"example.com/acoustic-survey-service/internal/reading"
	"example.com/acoustic-survey-service/internal/survey"
)

type Service struct {
	catalog *catalog.Catalog
}

func NewService(catalog *catalog.Catalog) *Service {
	return &Service{catalog: catalog}
}

func (s *Service) Summary(store *survey.Store, alerts *alert.Service, surveyID string) (model.Summary, error) {
	parent, err := store.Get(surveyID)
	if err != nil {
		return model.Summary{}, err
	}
	values := store.Readings(surveyID)
	eligible := make([]model.Reading, 0, len(values))
	for _, value := range values {
		if model.IsReadingWithinWindow(value.CapturedAt, parent.ActivatedAt, parent.ClosedAt) {
			eligible = append(eligible, value)
		}
	}
	summary := model.Summary{
		SurveyID:      parent.ID,
		State:         parent.State,
		TotalReadings: len(eligible),
		OpenAlerts:    alerts.OpenCount(surveyID),
		BandCounts:    make(map[string]int),
	}
	if len(eligible) == 0 {
		return summary, nil
	}
	minimumDepth := eligible[0].DepthM
	maximumDepth := eligible[0].DepthM
	normalizedTotal := 0.0
	for _, value := range eligible {
		if value.DepthM < minimumDepth {
			minimumDepth = value.DepthM
		}
		if value.DepthM > maximumDepth {
			maximumDepth = value.DepthM
		}
		summary.BandCounts[s.frequencyBand(value.FrequencyHz)]++
		if reading.IsUsable(value) {
			summary.ValidReadings++
			normalizedTotal += value.NormalizedDB
		}
	}
	summary.MinimumDepthM = model.Round(minimumDepth, 2)
	summary.MaximumDepthM = model.Round(maximumDepth, 2)
	if summary.ValidReadings > 0 {
		summary.AverageNormalizedDB = model.Round(normalizedTotal/float64(summary.ValidReadings), 2)
	}
	return summary, nil
}

func (s *Service) frequencyBand(frequency float64) string {
	for _, band := range s.catalog.List() {
		if catalog.InCalibrationRange(band, frequency) {
			return band.ID
		}
	}
	return "outside-catalog"
}
