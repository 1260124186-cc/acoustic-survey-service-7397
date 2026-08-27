package reading

import (
	"fmt"
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
	// 读数采样时间必须落在测线采样窗口内：不早于启动时间、不晚于结束时间。
	// 设备断链补传的缓存读数若早于启动时间则属于上一条测线的残留，应被拒绝。
	if parent.ActivatedAt != nil && captured.Before(*parent.ActivatedAt) {
		return model.Reading{}, model.NewError("invalid_reading", "reading was captured before the survey started")
	}
	if parent.ClosedAt != nil && captured.After(*parent.ClosedAt) {
		return model.Reading{}, model.NewError("invalid_reading", "reading was captured after the survey ended")
	}
	prior := latest(s.store.Readings(surveyID))
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
	s.alerts.Evaluate(band, value, prior)
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

func latest(values []model.Reading) *model.Reading {
	if len(values) == 0 {
		return nil
	}
	value := values[len(values)-1]
	return &value
}
