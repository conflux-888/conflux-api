package summary

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type SeverityBreakdown struct {
	Critical int `bson:"critical" json:"critical"`
	High     int `bson:"high" json:"high"`
	Medium   int `bson:"medium" json:"medium"`
	Low      int `bson:"low" json:"low"`
}

type DailySummary struct {
	ID                bson.ObjectID     `bson:"_id,omitempty" json:"id"`
	SummaryDate       string            `bson:"summary_date" json:"summary_date"`
	Status            string            `bson:"status" json:"status"` // completed, failed, no_events
	EventCount        int               `bson:"event_count" json:"event_count"`
	Title             string            `bson:"title" json:"title"`
	Content           string            `bson:"content" json:"content"`
	SeverityBreakdown SeverityBreakdown `bson:"severity_breakdown" json:"severity_breakdown"`
	Model             string            `bson:"model" json:"model"`
	PromptTokens      int               `bson:"prompt_tokens" json:"prompt_tokens"`
	CompletionTokens  int               `bson:"completion_tokens" json:"completion_tokens"`
	GenerationNumber  int               `bson:"generation_number" json:"generation_number"`
	GeneratedAt       time.Time         `bson:"generated_at" json:"generated_at"`
	CreatedAt         time.Time         `bson:"created_at" json:"created_at"`
	UpdatedAt         time.Time         `bson:"updated_at" json:"updated_at"`
	ErrorMessage      string            `bson:"error_message,omitempty" json:"error_message,omitempty"`
}

type TriggerRequest struct {
	Date string `json:"date" binding:"required"`
}

// Country to region mapping for pre-processing events before LLM input
var CountryToRegion = map[string]string{
	// Eastern Europe
	"UA": "Eastern Europe", "RU": "Eastern Europe", "BY": "Eastern Europe",
	"MD": "Eastern Europe", "GG": "Eastern Europe", "PL": "Eastern Europe",

	// Middle East
	"IL": "Middle East", "PS": "Middle East", "IR": "Middle East",
	"IQ": "Middle East", "SY": "Middle East", "YE": "Middle East",
	"LB": "Middle East", "JO": "Middle East", "SA": "Middle East",

	// South Asia
	"AF": "South Asia", "PK": "South Asia", "IN": "South Asia",
	"LK": "South Asia", "NP": "South Asia", "BD": "South Asia",
	"BM": "South Asia",

	// Southeast Asia
	"MM": "Southeast Asia", "TH": "Southeast Asia", "PH": "Southeast Asia",
	"ID": "Southeast Asia", "MY": "Southeast Asia",

	// East Asia
	"CN": "East Asia", "KN": "East Asia", "KS": "East Asia",
	"TW": "East Asia", "JA": "East Asia",

	// West Africa
	"NG": "West Africa", "ML": "West Africa", "BF": "West Africa",
	"NI": "West Africa", "GH": "West Africa", "SN": "West Africa",
	"CM": "West Africa",

	// East Africa
	"SO": "East Africa", "ET": "East Africa", "KE": "East Africa",
	"SD": "East Africa", "SS": "East Africa", "UG": "East Africa",
	"CD": "East Africa", "RW": "East Africa",

	// North Africa
	"LY": "North Africa", "EG": "North Africa", "TN": "North Africa",
	"AG": "North Africa", "MO": "North Africa",

	// Central Africa
	"CF": "Central Africa", "CG": "Central Africa", "CB": "Central Africa",

	// Southern Africa
	"MZ": "Southern Africa", "SF": "Southern Africa", "ZI": "Southern Africa",

	// Central America & Caribbean
	"MX": "Central America", "HO": "Central America", "GT": "Central America",
	"ES": "Central America", "NU": "Central America", "HA": "Central America",
	"CU": "Central America",

	// South America
	"CO": "South America", "VE": "South America", "BR": "South America",
	"PE": "South America", "EC": "South America",

	// Europe (Western)
	"FR": "Western Europe", "GM": "Western Europe", "UK": "Western Europe",
	"SP": "Western Europe", "IT": "Western Europe",

	// Caucasus & Central Asia
	"AM": "Caucasus & Central Asia", "AJ": "Caucasus & Central Asia",
	"KZ": "Caucasus & Central Asia",
	"UZ": "Caucasus & Central Asia", "TI": "Caucasus & Central Asia",
}

func RegionForCountry(countryCode string) string {
	if region, ok := CountryToRegion[countryCode]; ok {
		return region
	}
	return "Other"
}
