package survey

import (
	"fmt"
	"sort"
	"sync"

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

func (s *Store) AppendReading(value model.Reading) (model.Reading, *model.Reading, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	parent, exists := s.surveys[value.SurveyID]
	if !exists {
		return model.Reading{}, nil, model.NewError("not_found", "survey %q was not found", value.SurveyID)
	}
	if parent.State != model.Active {
		return model.Reading{}, nil, model.NewError("invalid_state", "readings can only be added to active surveys")
	}
	values := s.readings[value.SurveyID]
	var previous *model.Reading
	if len(values) > 0 {
		copy := values[len(values)-1]
		previous = &copy
	}
	value.Sequence = len(values) + 1
	value.ID = fmt.Sprintf("%s-%03d", value.SurveyID, value.Sequence)
	s.readings[value.SurveyID] = append(values, value)
	parent.ReadingCount = value.Sequence
	s.surveys[value.SurveyID] = parent
	return value, previous, nil
}

func (s *Store) Readings(id string) []model.Reading {
	s.mu.RLock()
	defer s.mu.RUnlock()
	source := s.readings[id]
	result := make([]model.Reading, len(source))
	copy(result, source)
	return result
}
