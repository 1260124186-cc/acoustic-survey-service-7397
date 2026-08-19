package reading

import (
	"testing"

	"example.com/acoustic-survey-service/internal/alert"
	"example.com/acoustic-survey-service/internal/catalog"
	"example.com/acoustic-survey-service/internal/model"
	"example.com/acoustic-survey-service/internal/survey"
)

func TestOutOfRangeReadingIsNotUsable(t *testing.T) {
	bands := catalog.NewDefault()
	store := survey.NewStore()
	surveys := survey.NewService(store, bands)
	created, err := surveys.Create("range-line", "bay", "coastal-38khz")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = surveys.Activate(created.ID); err != nil {
		t.Fatal(err)
	}
	value, err := NewService(bands, alert.NewService(), store).Add(created.ID, model.ReadingInput{FrequencyHz: 42000, EchoDB: -48, NoiseDB: -70, DepthM: 20})
	if err != nil {
		t.Fatal(err)
	}
	if IsUsable(value) {
		t.Fatalf("quality=%d, wanted unusable", value.QualityScore)
	}
}
