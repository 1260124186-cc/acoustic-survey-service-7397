package survey

import (
	"testing"

	"example.com/acoustic-survey-service/internal/catalog"
)

func TestEmptyActiveSurveyCannotClose(t *testing.T) {
	bands := catalog.NewDefault()
	service := NewService(NewStore(), bands)
	value, err := service.Create("empty-line", "bay", "coastal-38khz")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Activate(value.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Close(value.ID); err == nil {
		t.Fatal("close unexpectedly succeeded")
	}
}
