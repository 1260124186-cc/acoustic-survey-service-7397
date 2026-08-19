package catalog

import (
	"math"
	"sort"

	"example.com/acoustic-survey-service/internal/model"
)

type Catalog struct {
	bands map[string]model.Band
}

func NewDefault() *Catalog {
	all := make([]model.Band, 0, 250)
	all = append(all, coreBands()...)
	all = append(all, profileBands()...)
	indexed := make(map[string]model.Band, len(all))
	for _, band := range all {
		indexed[band.ID] = band
	}
	return &Catalog{bands: indexed}
}

func (c *Catalog) Find(id string) (model.Band, bool) {
	value, ok := c.bands[id]
	return value, ok
}

func (c *Catalog) List() []model.Band {
	result := make([]model.Band, 0, len(c.bands))
	for _, value := range c.bands {
		result = append(result, value)
	}
	sort.Slice(result, func(left int, right int) bool {
		return result[left].CenterHz < result[right].CenterHz
	})
	return result
}

func NormalizeEcho(band model.Band, echoDB float64, depthM float64) float64 {
	spreading := 20 * math.Log10(math.Max(depthM, 1))
	return model.Round(echoDB+spreading-band.ReferenceDB, 2)
}

func InCalibrationRange(band model.Band, frequency float64) bool {
	return frequency >= band.MinimumHz && frequency <= band.MaximumHz
}

func SignalToNoise(echoDB float64, noiseDB float64) float64 {
	return model.Round(echoDB-noiseDB, 2)
}

func coreBands() []model.Band {
	return []model.Band{
		{
			ID:          "coastal-38khz",
			Label:       "Coastal 38 kHz",
			CenterHz:    38000,
			MinimumHz:   37000,
			MaximumHz:   39000,
			ReferenceDB: -55,
			MinimumSNR:  12,
		},
		{
			ID:          "shelf-120khz",
			Label:       "Shelf 120 kHz",
			CenterHz:    120000,
			MinimumHz:   117000,
			MaximumHz:   123000,
			ReferenceDB: -49,
			MinimumSNR:  10,
		},
		{
			ID:          "deep-200khz",
			Label:       "Deep 200 kHz",
			CenterHz:    200000,
			MinimumHz:   196000,
			MaximumHz:   204000,
			ReferenceDB: -45,
			MinimumSNR:  9,
		},
	}
}

func profileBands() []model.Band {
	result := make([]model.Band, 0, 240)
	result = append(result, sector01Profiles...)
	result = append(result, sector02Profiles...)
	result = append(result, sector03Profiles...)
	result = append(result, sector04Profiles...)
	result = append(result, sector05Profiles...)
	result = append(result, sector06Profiles...)
	result = append(result, sector07Profiles...)
	result = append(result, sector08Profiles...)
	result = append(result, sector09Profiles...)
	result = append(result, sector10Profiles...)
	result = append(result, sector11Profiles...)
	result = append(result, sector12Profiles...)
	result = append(result, sector13Profiles...)
	result = append(result, sector14Profiles...)
	result = append(result, sector15Profiles...)
	result = append(result, sector16Profiles...)
	result = append(result, sector17Profiles...)
	result = append(result, sector18Profiles...)
	result = append(result, sector19Profiles...)
	result = append(result, sector20Profiles...)
	return result
}
