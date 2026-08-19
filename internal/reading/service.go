package reading

import (
	"fmt"
	"sort"
	"time"

	"example.com/acoustic-survey-service/internal/alert"
	"example.com/acoustic-survey-service/internal/catalog"
	"example.com/acoustic-survey-service/internal/model"
	"example.com/acoustic-survey-service/internal/survey"
)

type Service struct {
	catalog *catalog.Catalog
	alerts  *alert.Service
	store   *survey.Store
}

func NewService(catalog *catalog.Catalog, alerts *alert.Service, store *survey.Store) *Service {
	return &Service{catalog: catalog, alerts: alerts, store: store}
}

func (s *Service) Add(surveyID string, input model.ReadingInput) (model.Reading, error) {
	if err := model.ValidateReading(input); err != nil {
		return model.Reading{}, err
	}
	parent, err := s.store.Get(surveyID)
	if err != nil {
		return model.Reading{}, err
	}
	if parent.State != model.Active {
		return model.Reading{}, model.NewError("invalid_state", "readings can only be added to active surveys")
	}
	band, ok := s.catalog.Find(parent.Band)
	if !ok {
		return model.Reading{}, model.NewError("invalid_band", "survey band %q is unavailable", parent.Band)
	}
	captured := time.Now().UTC()
	if input.CapturedAt != nil {
		captured = input.CapturedAt.UTC()
	}
	value := model.Reading{
		ID:           fmt.Sprintf("%s-%03d", surveyID, parent.ReadingCount+1),
		SurveyID:     surveyID,
		FrequencyHz:  input.FrequencyHz,
		EchoDB:       input.EchoDB,
		NoiseDB:      input.NoiseDB,
		DepthM:       input.DepthM,
		NormalizedDB: catalog.NormalizeEcho(band, input.EchoDB, input.DepthM),
		QualityScore: qualityScore(band, input),
		CapturedAt:   captured,
	}
	if err := s.store.AppendReading(value); err != nil {
		return model.Reading{}, err
	}
	// 读数可能因测线设备断链补传而乱序到达，相邻跳变告警必须按采样时间
	// 而非网络抵达顺序判断，因此每次到达都按 CapturedAt 重排后整体重建。
	chronological := sortedByCapture(s.store.Readings(surveyID))
	s.alerts.Evaluate(band, value)
	s.alerts.ReevaluateNeighbors(chronological)
	return value, nil
}

func qualityScore(band model.Band, input model.ReadingInput) int {
	score := 100
	if !catalog.InCalibrationRange(band, input.FrequencyHz) {
		score -= 45
	}
	if catalog.SignalToNoise(input.EchoDB, input.NoiseDB) < band.MinimumSNR {
		score -= 30
	}
	if input.DepthM > 500 {
		score -= 10
	}
	return model.ClampInt(score, 0, 100)
}

func IsUsable(value model.Reading) bool {
	return value.QualityScore >= 70
}

// sortedByCapture 按采样时间升序返回读数副本，采样时间相同的读数保持稳定顺序。
func sortedByCapture(values []model.Reading) []model.Reading {
	result := make([]model.Reading, len(values))
	copy(result, values)
	sort.SliceStable(result, func(left int, right int) bool {
		return result[left].CapturedAt.Before(result[right].CapturedAt)
	})
	return result
}
