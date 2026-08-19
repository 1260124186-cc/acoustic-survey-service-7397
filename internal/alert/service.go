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

const abruptRule = "abrupt_normalized_change"

// Evaluate 对单条读数评估不依赖相邻关系的质量规则（校准范围、信噪比）。
// 相邻跳变告警由 ReevaluateNeighbors 按采样时间统一维护，避免乱序到达时
// 把网络抵达顺序误当成采样相邻关系。
func (s *Service) Evaluate(band model.Band, reading model.Reading) {
	notices := make([]model.Alert, 0, 2)
	if !catalog.InCalibrationRange(band, reading.FrequencyHz) {
		notices = append(notices, s.newAlert(reading, "frequency_outside_calibration", model.AlertCritical, fmt.Sprintf("frequency %.0f Hz is outside the calibration interval", reading.FrequencyHz)))
	}
	if catalog.SignalToNoise(reading.EchoDB, reading.NoiseDB) < band.MinimumSNR {
		notices = append(notices, s.newAlert(reading, "low_signal_to_noise", model.AlertWarning, "signal-to-noise ratio is below the band threshold"))
	}
	if len(notices) == 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.alerts[reading.SurveyID] = append(s.alerts[reading.SurveyID], notices...)
}

// ReevaluateNeighbors 基于按采样时间升序排列的读数序列重建相邻跳变告警。
// readings 必须按 CapturedAt 升序传入；乱序补传的读数会落到正确位置，受影响的
// 相邻关系随之重算，此前按抵达顺序产生的过期跳变告警会被丢弃。
func (s *Service) ReevaluateNeighbors(readings []model.Reading) {
	if len(readings) == 0 {
		return
	}
	surveyID := readings[0].SurveyID
	notices := make([]model.Alert, 0)
	for index := 1; index < len(readings); index++ {
		current := readings[index]
		previous := readings[index-1]
		if absolute(current.NormalizedDB-previous.NormalizedDB) > 18 {
			notices = append(notices, s.newAlert(current, abruptRule, model.AlertWarning, "normalized echo changed abruptly from the prior reading"))
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	kept := make([]model.Alert, 0, len(s.alerts[surveyID])+len(notices))
	for _, existing := range s.alerts[surveyID] {
		if existing.Rule == abruptRule {
			continue
		}
		kept = append(kept, existing)
	}
	s.alerts[surveyID] = append(kept, notices...)
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
