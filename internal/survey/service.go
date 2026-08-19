package survey

import (
	"strings"
	"time"

	"example.com/acoustic-survey-service/internal/alert"
	"example.com/acoustic-survey-service/internal/catalog"
	"example.com/acoustic-survey-service/internal/model"
)

type Service struct {
	store   *Store
	catalog *catalog.Catalog
	alerts  *alert.Service
}

func NewService(store *Store, catalog *catalog.Catalog) *Service {
	return &Service{store: store, catalog: catalog}
}

// WithAlerts 注入告警服务，用于在结束测线时阻断未解决的严重告警。
func (s *Service) WithAlerts(alerts *alert.Service) *Service {
	s.alerts = alerts
	return s
}

func (s *Service) Create(id string, area string, bandID string) (model.Survey, error) {
	id = strings.TrimSpace(id)
	area = strings.TrimSpace(area)
	bandID = strings.TrimSpace(bandID)
	if err := model.ValidateSurvey(id, area, bandID); err != nil {
		return model.Survey{}, err
	}
	if _, ok := s.catalog.Find(bandID); !ok {
		return model.Survey{}, model.NewError("invalid_band", "band %q is not available", bandID)
	}
	value := model.Survey{
		ID:        id,
		Area:      area,
		Band:      bandID,
		State:     model.Draft,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.store.Create(value); err != nil {
		return model.Survey{}, err
	}
	return value, nil
}

func (s *Service) Activate(id string) (model.Survey, error) {
	value, err := s.store.Get(id)
	if err != nil {
		return model.Survey{}, err
	}
	if value.State != model.Draft {
		return model.Survey{}, model.NewError("invalid_state", "only draft surveys can be activated")
	}
	now := time.Now().UTC()
	value.State = model.Active
	value.ActivatedAt = &now
	if err := s.store.Save(value); err != nil {
		return model.Survey{}, err
	}
	return value, nil
}

func (s *Service) Close(id string) (model.Survey, error) {
	value, err := s.store.Get(id)
	if err != nil {
		return model.Survey{}, err
	}
	if value.State != model.Active {
		return model.Survey{}, model.NewError("invalid_state", "only active surveys can be closed")
	}
	if value.ReadingCount == 0 {
		return model.Survey{}, model.NewError("invalid_state", "an active survey needs at least one reading before closing")
	}
	if s.alerts != nil && s.alerts.HasOpenCritical(id) {
		return model.Survey{}, model.NewError("unresolved_alerts", "cannot close a survey with unresolved critical alerts")
	}
	now := time.Now().UTC()
	value.State = model.Closed
	value.ClosedAt = &now
	if err := s.store.Save(value); err != nil {
		return model.Survey{}, err
	}
	return value, nil
}

func (s *Service) Get(id string) (model.Survey, error) {
	return s.store.Get(id)
}

func (s *Service) List() []model.Survey {
	return s.store.List()
}

func (s *Service) Store() *Store {
	return s.store
}
