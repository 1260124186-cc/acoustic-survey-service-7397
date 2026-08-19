package reading

import (
	"sync"
	"testing"

	"example.com/acoustic-survey-service/internal/alert"
	"example.com/acoustic-survey-service/internal/catalog"
	"example.com/acoustic-survey-service/internal/model"
	"example.com/acoustic-survey-service/internal/survey"
)

func TestConcurrentReadingsReceiveDistinctIDs(t *testing.T) {
	bands := catalog.NewDefault()
	store := survey.NewStore()
	surveys := survey.NewService(store, bands)
	created, err := surveys.Create("parallel-line", "outer-bay", "coastal-38khz")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = surveys.Activate(created.ID); err != nil {
		t.Fatal(err)
	}

	const workers = 32
	service := NewService(bands, alert.NewService(), store)
	start := make(chan struct{})
	results := make(chan model.Reading, workers)
	errors := make(chan error, workers)
	var group sync.WaitGroup
	for i := 0; i < workers; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			<-start
			value, addErr := service.Add(created.ID, model.ReadingInput{FrequencyHz: 38000, EchoDB: -48, NoiseDB: -80, DepthM: 30})
			if addErr != nil {
				errors <- addErr
				return
			}
			results <- value
		}()
	}
	close(start)
	group.Wait()
	close(results)
	close(errors)
	for addErr := range errors {
		t.Fatalf("concurrent add failed: %v", addErr)
	}

	ids := make(map[string]bool)
	for value := range results {
		if ids[value.ID] {
			t.Fatalf("duplicate reading id returned: %s", value.ID)
		}
		ids[value.ID] = true
	}
	if len(ids) != workers {
		t.Fatalf("distinct ids = %d, want %d", len(ids), workers)
	}
	if got := len(store.Readings(created.ID)); got != workers {
		t.Fatalf("stored readings = %d, want %d", got, workers)
	}
}
