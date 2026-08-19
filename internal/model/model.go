package model

import (
	"fmt"
	"math"
	"strings"
	"time"
)

type SurveyState string

const (
	Draft  SurveyState = "draft"
	Active SurveyState = "active"
	Closed SurveyState = "closed"
)

type Survey struct {
	ID           string      `json:"id"`
	Area         string      `json:"area"`
	Band         string      `json:"band"`
	State        SurveyState `json:"state"`
	CreatedAt    time.Time   `json:"created_at"`
	ActivatedAt  *time.Time  `json:"activated_at,omitempty"`
	ClosedAt     *time.Time  `json:"closed_at,omitempty"`
	ReadingCount int         `json:"reading_count"`
}

type ReadingInput struct {
	FrequencyHz float64    `json:"frequency_hz"`
	EchoDB      float64    `json:"echo_db"`
	NoiseDB     float64    `json:"noise_db"`
	DepthM      float64    `json:"depth_m"`
	CapturedAt  *time.Time `json:"captured_at,omitempty"`
}

type Reading struct {
	ID           string    `json:"id"`
	SurveyID     string    `json:"survey_id"`
	FrequencyHz  float64   `json:"frequency_hz"`
	EchoDB       float64   `json:"echo_db"`
	NoiseDB      float64   `json:"noise_db"`
	DepthM       float64   `json:"depth_m"`
	NormalizedDB float64   `json:"normalized_db"`
	QualityScore int       `json:"quality_score"`
	CapturedAt   time.Time `json:"captured_at"`
}

type AlertSeverity string

const (
	AlertInfo     AlertSeverity = "info"
	AlertWarning  AlertSeverity = "warning"
	AlertCritical AlertSeverity = "critical"
)

type Alert struct {
	ID        string        `json:"id"`
	SurveyID  string        `json:"survey_id"`
	ReadingID string        `json:"reading_id"`
	Rule      string        `json:"rule"`
	Severity  AlertSeverity `json:"severity"`
	Message   string        `json:"message"`
	CreatedAt time.Time     `json:"created_at"`
	Resolved  bool          `json:"resolved"`
}

type Band struct {
	ID          string  `json:"id"`
	Label       string  `json:"label"`
	CenterHz    float64 `json:"center_hz"`
	MinimumHz   float64 `json:"minimum_hz"`
	MaximumHz   float64 `json:"maximum_hz"`
	ReferenceDB float64 `json:"reference_db"`
	MinimumSNR  float64 `json:"minimum_snr"`
}

type Summary struct {
	SurveyID            string         `json:"survey_id"`
	State               SurveyState    `json:"state"`
	TotalReadings       int            `json:"total_readings"`
	ValidReadings       int            `json:"valid_readings"`
	AverageNormalizedDB float64        `json:"average_normalized_db"`
	MinimumDepthM       float64        `json:"minimum_depth_m"`
	MaximumDepthM       float64        `json:"maximum_depth_m"`
	OpenAlerts          int            `json:"open_alerts"`
	BandCounts          map[string]int `json:"band_counts"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return e.Message
}

func NewError(code string, format string, values ...any) *Error {
	return &Error{Code: code, Message: fmt.Sprintf(format, values...)}
}

func ValidateSurvey(id string, area string, band string) error {
	if len(strings.TrimSpace(id)) < 3 {
		return NewError("invalid_survey", "survey id must contain at least three characters")
	}
	if len(strings.TrimSpace(area)) < 3 {
		return NewError("invalid_survey", "area must contain at least three characters")
	}
	if strings.TrimSpace(band) == "" {
		return NewError("invalid_survey", "band is required")
	}
	return nil
}

func ValidateReading(input ReadingInput) error {
	if input.FrequencyHz <= 0 || math.IsNaN(input.FrequencyHz) {
		return NewError("invalid_reading", "frequency must be positive")
	}
	if input.DepthM <= 0 || math.IsNaN(input.DepthM) {
		return NewError("invalid_reading", "depth must be positive")
	}
	if input.EchoDB < -130 || input.EchoDB > 20 {
		return NewError("invalid_reading", "echo level is outside the supported range")
	}
	if input.NoiseDB < -150 || input.NoiseDB > 20 {
		return NewError("invalid_reading", "noise level is outside the supported range")
	}
	if input.NoiseDB > input.EchoDB {
		return NewError("invalid_reading", "noise level cannot exceed echo level")
	}
	return nil
}

// UsableQualityThreshold 是读数被判定为可用的最低质量分。
const UsableQualityThreshold = 70

// IsUsableReading 判断读数是否可用于结束测线与摘要统计。
func IsUsableReading(value Reading) bool {
	return value.QualityScore >= UsableQualityThreshold
}

func Round(value float64, places int) float64 {
	factor := math.Pow10(places)
	return math.Round(value*factor) / factor
}

func ClampInt(value int, minimum int, maximum int) int {
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}
