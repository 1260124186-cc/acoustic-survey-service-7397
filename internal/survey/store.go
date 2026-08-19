package survey

import (
	"sort"
	"sync"
	"time"

	"example.com/acoustic-survey-service/internal/model"
)

type Store struct {
	mu       sync.RWMutex
	surveys  map[string]model.Survey
	readings map[string][]model.Reading
}

func NewStore() *Store {
	return &Store{
		surveys:  make(map[string]model.Survey),
		readings: make(map[string][]model.Reading),
	}
}

func (s *Store) Create(value model.Survey) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.surveys[value.ID]; exists {
		return model.NewError("conflict", "survey %q already exists", value.ID)
	}
	s.surveys[value.ID] = value
	return nil
}

func (s *Store) Get(id string) (model.Survey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.surveys[id]
	if !ok {
		return model.Survey{}, model.NewError("not_found", "survey %q was not found", id)
	}
	return value, nil
}

func (s *Store) Save(value model.Survey) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.surveys[value.ID]; !exists {
		return model.NewError("not_found", "survey %q was not found", value.ID)
	}
	s.surveys[value.ID] = value
	return nil
}

func (s *Store) List() []model.Survey {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]model.Survey, 0, len(s.surveys))
	for _, value := range s.surveys {
		result = append(result, value)
	}
	sort.Slice(result, func(left int, right int) bool {
		return result[left].CreatedAt.Before(result[right].CreatedAt)
	})
	return result
}

func (s *Store) AppendReading(value model.Reading) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	survey, exists := s.surveys[value.SurveyID]
	if !exists {
		return model.NewError("not_found", "survey %q was not found", value.SurveyID)
	}
	if survey.State != model.Active {
		return model.NewError("invalid_state", "readings can only be added to active surveys")
	}
	s.readings[value.SurveyID] = append(s.readings[value.SurveyID], value)
	survey.ReadingCount = len(s.readings[value.SurveyID])
	s.surveys[value.SurveyID] = survey
	return nil
}

func (s *Store) Readings(id string) []model.Reading {
	s.mu.RLock()
	defer s.mu.RUnlock()
	source := s.readings[id]
	result := make([]model.Reading, len(source))
	copy(result, source)
	return result
}

func (s *Store) Close(id string) (model.Survey, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	value, exists := s.surveys[id]
	if !exists {
		return model.Survey{}, model.NewError("not_found", "survey %q was not found", id)
	}
	if value.State != model.Active {
		return model.Survey{}, model.NewError("invalid_state", "only active surveys can be closed")
	}
	if value.ReadingCount == 0 {
		return model.Survey{}, model.NewError("invalid_state", "an active survey needs at least one reading before closing")
	}
	now := time.Now().UTC()
	value.State = model.Closed
	value.ClosedAt = &now
	s.surveys[id] = value
	return value, nil
}
