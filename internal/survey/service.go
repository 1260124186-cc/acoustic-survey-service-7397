package survey

import (
	"strings"
	"time"

	"example.com/acoustic-survey-service/internal/catalog"
	"example.com/acoustic-survey-service/internal/model"
)

type Service struct {
	store   *Store
	catalog *catalog.Catalog
}

func NewService(store *Store, catalog *catalog.Catalog) *Service {
	return &Service{store: store, catalog: catalog}
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
	if !hasUsableReading(s.store.Readings(id)) {
		return model.Survey{}, model.NewError("invalid_state", "an active survey needs at least one usable reading before closing")
	}
	now := time.Now().UTC()
	value.State = model.Closed
	value.ClosedAt = &now
	if err := s.store.Save(value); err != nil {
		return model.Survey{}, err
	}
	return value, nil
}

// hasUsableReading 判断读数中是否存在可用于结束测线的有效读数。
func hasUsableReading(values []model.Reading) bool {
	for _, value := range values {
		if model.IsUsableReading(value) {
			return true
		}
	}
	return false
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
