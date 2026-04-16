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

type TopEvent struct {
	Title       string `bson:"title" json:"title"`
	Severity    string `bson:"severity" json:"severity"`
	Country     string `bson:"country" json:"country"`
	Location    string `bson:"location" json:"location"`
	Description string `bson:"description" json:"description"`
}

type DailySummary struct {
	ID                bson.ObjectID     `bson:"_id,omitempty" json:"id"`
	SummaryDate       string            `bson:"summary_date" json:"summary_date"`
	Status            string            `bson:"status" json:"status"` // completed, failed, no_events
	EventCount        int               `bson:"event_count" json:"event_count"`
	IncidentCount     int               `bson:"incident_count" json:"incident_count"` // after dedup
	Title             string            `bson:"title" json:"title"`
	Content           string            `bson:"content" json:"content"`
	TopEvents         []TopEvent        `bson:"top_events" json:"top_events"`
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

// Country to region mapping using FIPS country codes (GDELT standard)
var CountryToRegion = map[string]string{
	// Eastern Europe (FIPS)
	"UP": "Eastern Europe", // Ukraine
	"RS": "Eastern Europe", // Russia
	"BO": "Eastern Europe", // Belarus
	"MD": "Eastern Europe", // Moldova
	"PL": "Eastern Europe", // Poland
	"RO": "Eastern Europe", // Romania
	"HU": "Eastern Europe", // Hungary

	// Middle East (FIPS)
	"IS": "Middle East", // Israel
	"GZ": "Middle East", // Gaza Strip
	"WE": "Middle East", // West Bank
	"IR": "Middle East", // Iran
	"IZ": "Middle East", // Iraq
	"SY": "Middle East", // Syria
	"YM": "Middle East", // Yemen
	"LE": "Middle East", // Lebanon
	"JO": "Middle East", // Jordan
	"SA": "Middle East", // Saudi Arabia
	"AE": "Middle East", // UAE
	"QA": "Middle East", // Qatar
	"TU": "Middle East", // Turkey

	// South Asia (FIPS)
	"AF": "South Asia", // Afghanistan
	"PK": "South Asia", // Pakistan
	"IN": "South Asia", // India
	"CE": "South Asia", // Sri Lanka
	"NP": "South Asia", // Nepal
	"BG": "South Asia", // Bangladesh
	"BM": "South Asia", // Burma/Myanmar

	// Southeast Asia (FIPS)
	"TH": "Southeast Asia", // Thailand
	"RP": "Southeast Asia", // Philippines
	"ID": "Southeast Asia", // Indonesia
	"MY": "Southeast Asia", // Malaysia
	"VM": "Southeast Asia", // Vietnam
	"CB": "Southeast Asia", // Cambodia

	// East Asia (FIPS)
	"CH": "East Asia", // China
	"KN": "East Asia", // North Korea
	"KS": "East Asia", // South Korea
	"TW": "East Asia", // Taiwan
	"JA": "East Asia", // Japan

	// West Africa (FIPS)
	"NI": "West Africa", // Nigeria
	"ML": "West Africa", // Mali
	"UV": "West Africa", // Burkina Faso
	"NG": "West Africa", // Niger
	"GH": "West Africa", // Ghana
	"SG": "West Africa", // Senegal
	"CM": "West Africa", // Cameroon

	// East Africa (FIPS)
	"SO": "East Africa", // Somalia
	"ET": "East Africa", // Ethiopia
	"KE": "East Africa", // Kenya
	"SU": "East Africa", // Sudan
	"OD": "East Africa", // South Sudan
	"UG": "East Africa", // Uganda
	"CG": "East Africa", // Congo (DRC)
	"RW": "East Africa", // Rwanda

	// North Africa (FIPS)
	"LY": "North Africa", // Libya
	"EG": "North Africa", // Egypt
	"TS": "North Africa", // Tunisia
	"AG": "North Africa", // Algeria
	"MO": "North Africa", // Morocco

	// Central Africa (FIPS)
	"CT": "Central Africa", // Central African Republic
	"CF": "Central Africa", // Congo (Brazzaville)

	// Southern Africa (FIPS)
	"MZ": "Southern Africa", // Mozambique
	"SF": "Southern Africa", // South Africa
	"ZI": "Southern Africa", // Zimbabwe

	// North America (FIPS)
	"US": "North America", // United States
	"CA": "North America", // Canada
	"MX": "North America", // Mexico

	// Central America & Caribbean (FIPS)
	"HO": "Central America", // Honduras
	"GT": "Central America", // Guatemala
	"ES": "Central America", // El Salvador
	"NU": "Central America", // Nicaragua
	"HA": "Central America", // Haiti
	"CU": "Central America", // Cuba

	// South America (FIPS)
	"CO": "South America", // Colombia
	"VE": "South America", // Venezuela
	"BR": "South America", // Brazil
	"PE": "South America", // Peru
	"EC": "South America", // Ecuador

	// Western Europe (FIPS)
	"FR": "Western Europe", // France
	"GM": "Western Europe", // Germany
	"UK": "Western Europe", // United Kingdom
	"SP": "Western Europe", // Spain
	"IT": "Western Europe", // Italy
	"BE": "Western Europe", // Belgium
	"NL": "Western Europe", // Netherlands
	"DA": "Western Europe", // Denmark
	"SW": "Western Europe", // Sweden
	"NO": "Western Europe", // Norway

	// Caucasus & Central Asia (FIPS)
	"AM": "Caucasus & Central Asia", // Armenia
	"AJ": "Caucasus & Central Asia", // Azerbaijan
	"GG": "Caucasus & Central Asia", // Georgia
	"KZ": "Caucasus & Central Asia", // Kazakhstan
	"UZ": "Caucasus & Central Asia", // Uzbekistan
	"TI": "Caucasus & Central Asia", // Tajikistan

	// Oceania (FIPS)
	"AS": "Oceania", // Australia
}

func RegionForCountry(countryCode string) string {
	if region, ok := CountryToRegion[countryCode]; ok {
		return region
	}
	return "Other"
}
