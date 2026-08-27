package alert

import (
	"fmt"
	"sync"

	"example.com/acoustic-survey-service/internal/catalog"
	"example.com/acoustic-survey-service/internal/model"
)

type Service struct {
	mu     sync.RWMutex
	alerts map[string][]model.Alert
}

func NewService() *Service {
	return &Service{alerts: make(map[string][]model.Alert)}
}

func (s *Service) Evaluate(band model.Band, reading model.Reading, previous *model.Reading) {
	notices := make([]model.Alert, 0, 3)
	if !catalog.InCalibrationRange(band, reading.FrequencyHz) {
		notices = append(notices, s.newAlert(reading, "frequency_outside_calibration", model.AlertCritical, fmt.Sprintf("frequency %.0f Hz is outside the calibration interval", reading.FrequencyHz)))
	}
	if catalog.SignalToNoise(reading.EchoDB, reading.NoiseDB) < band.MinimumSNR {
		notices = append(notices, s.newAlert(reading, "low_signal_to_noise", model.AlertWarning, "signal-to-noise ratio is below the band threshold"))
	}
	if previous != nil && absolute(reading.NormalizedDB-previous.NormalizedDB) > 18 {
		notices = append(notices, s.newAlert(reading, "abrupt_normalized_change", model.AlertWarning, "normalized echo changed abruptly from the prior reading"))
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.alerts[reading.SurveyID] = append(s.alerts[reading.SurveyID], notices...)
}

func (s *Service) List(surveyID string) []model.Alert {
	s.mu.RLock()
	defer s.mu.RUnlock()
	source := s.alerts[surveyID]
	if source == nil {
		return []model.Alert{}
	}
	result := make([]model.Alert, len(source))
	copy(result, source)
	return result
}

func (s *Service) OpenCount(surveyID string) int {
	count := 0
	for _, value := range s.List(surveyID) {
		if !value.Resolved {
			count++
		}
	}
	return count
}

func (s *Service) newAlert(reading model.Reading, rule string, severity model.AlertSeverity, message string) model.Alert {
	return model.Alert{
		ID:        reading.ID + ":" + rule,
		SurveyID:  reading.SurveyID,
		ReadingID: reading.ID,
		Rule:      rule,
		Severity:  severity,
		Message:   message,
		CreatedAt: reading.CapturedAt,
	}
}

func absolute(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}
