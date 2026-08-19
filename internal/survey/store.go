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

func (s *Store) PreviousReading(id string, capturedAt time.Time) *model.Reading {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var previous *model.Reading
	for _, value := range s.readings[id] {
		if !value.CapturedAt.Before(capturedAt) {
			continue
		}
		if previous == nil || previous.CapturedBefore(value) {
			copy := value
			previous = &copy
		}
	}
	return previous
}
