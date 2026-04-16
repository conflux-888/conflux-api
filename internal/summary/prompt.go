package summary

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/conflux-888/conflux-api/internal/common/gemini"
	"github.com/conflux-888/conflux-api/internal/event"
	genailib "github.com/google/generative-ai-go/genai"
	"github.com/rs/zerolog/log"
)

const modelName = "gemini-2.5-flash"

const systemInstruction = `You are a senior conflict correspondent writing for an international news wire service like Reuters or BBC World Service. You produce daily conflict briefings that are factual, authoritative, and written in clear professional English.

Your audience cares about INTERNATIONAL conflicts and geopolitical threats — wars, military operations, cross-border violence, terrorism, and insurgencies. Domestic crime, local police incidents, and routine law enforcement are NOT relevant.

Rules:
- Focus on the 3-5 most significant INTERNATIONAL developments only
- Do NOT attempt to cover every incident — be selective and editorial
- Write in third person, past tense
- Lead with the single most important story
- Each development should get 1-2 paragraphs with context
- End with a single sentence noting total incident count and most affected regions
- Do NOT use bullet points, headers, or lists — write continuous prose
- Do NOT speculate, but DO provide brief geopolitical context where obvious (e.g. "part of the ongoing Russia-Ukraine conflict")
- Ignore domestic US police/crime incidents unless they have international significance
- Ignore incidents where the actor or context is unclear/nonsensical (data noise)
- Length: 300-500 words. Quality over quantity.
- For top_events: pick the 5 most internationally significant events with 1-2 sentence descriptions`

type LLMOutput struct {
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	TopEvents []TopEvent `json:"top_events"`
}

var summarySchema = &genailib.Schema{
	Type: genailib.TypeObject,
	Properties: map[string]*genailib.Schema{
		"title":   {Type: genailib.TypeString, Description: "News headline for the daily briefing"},
		"content": {Type: genailib.TypeString, Description: "Full news article in continuous prose, 400-800 words"},
		"top_events": {
			Type:        genailib.TypeArray,
			Description: "Top 5 most critical events of the day",
			Items: &genailib.Schema{
				Type: genailib.TypeObject,
				Properties: map[string]*genailib.Schema{
					"title":       {Type: genailib.TypeString, Description: "Short event title"},
					"severity":    {Type: genailib.TypeString, Description: "critical, high, medium, or low"},
					"country":     {Type: genailib.TypeString, Description: "Country code"},
					"location":    {Type: genailib.TypeString, Description: "Location name"},
					"description": {Type: genailib.TypeString, Description: "1-2 sentence factual description"},
				},
				Required: []string{"title", "severity", "country", "location", "description"},
			},
		},
	},
	Required: []string{"title", "content", "top_events"},
}

type generateResult struct {
	Output          *LLMOutput
	IncidentCount   int
	PromptTokens    int
	CompletionTokens int
}

func generateSummary(ctx context.Context, client *gemini.Client, date string, events []event.Event) (*generateResult, error) {
	// Step 1: Deduplicate events into incidents
	incidents := deduplicateEvents(events)
	log.Info().Int("raw_events", len(events)).Int("incidents", len(incidents)).Msg("[summary.prompt] events deduplicated")

	// Step 2: Rank by significance
	rankIncidents(incidents)

	// Step 3: Build compact prompt
	prompt := buildPrompt(date, len(events), incidents)

	var output LLMOutput
	resp, err := client.GenerateJSON(ctx, gemini.GenerateRequest{
		Model:             modelName,
		SystemInstruction: systemInstruction,
		Prompt:            prompt,
		Temperature:       0.2,
		ResponseSchema:    summarySchema,
	}, &output)
	if err != nil {
		return nil, err
	}

	return &generateResult{
		Output:           &output,
		IncidentCount:    len(incidents),
		PromptTokens:     resp.PromptTokens,
		CompletionTokens: resp.CompletionTokens,
	}, nil
}

// incident represents a deduplicated group of events at the same location with the same type
type incident struct {
	LocationName  string
	Country       string
	Region        string
	EventType     string
	EventRootCode string
	MaxSeverity   string
	Actors        map[string]bool
	TotalSources  int
	TotalArticles int
	EventCount    int
	Score         float64 // significance score
}

var severityWeight = map[string]float64{
	"critical": 4,
	"high":     3,
	"medium":   2,
	"low":      1,
}

// deduplicateEvents groups events by country + location_name + event_root_code
func deduplicateEvents(events []event.Event) []incident {
	key := func(e event.Event) string {
		return e.Country + "|" + e.LocationName + "|" + e.EventRootCode
	}

	groups := map[string]*incident{}
	for _, e := range events {
		k := key(e)
		inc, exists := groups[k]
		if !exists {
			inc = &incident{
				LocationName:  e.LocationName,
				Country:       e.Country,
				Region:        RegionForCountry(e.Country),
				EventType:     e.EventType,
				EventRootCode: e.EventRootCode,
				MaxSeverity:   e.Severity,
				Actors:        map[string]bool{},
			}
			groups[k] = inc
		}

		inc.TotalSources += e.NumSources
		inc.TotalArticles += e.NumArticles
		inc.EventCount++

		for _, a := range e.Actors {
			if a != "" {
				inc.Actors[a] = true
			}
		}

		if severityWeight[e.Severity] > severityWeight[inc.MaxSeverity] {
			inc.MaxSeverity = e.Severity
		}
	}

	result := make([]incident, 0, len(groups))
	for _, inc := range groups {
		inc.Score = float64(inc.TotalArticles) * severityWeight[inc.MaxSeverity]
		result = append(result, *inc)
	}
	return result
}

// rankIncidents sorts by significance score descending
func rankIncidents(incidents []incident) {
	sort.Slice(incidents, func(i, j int) bool {
		return incidents[i].Score > incidents[j].Score
	})
}

func buildPrompt(date string, totalRawEvents int, incidents []incident) string {
	// Count severity across incidents
	var critical, high, medium, low int
	for _, inc := range incidents {
		switch inc.MaxSeverity {
		case "critical":
			critical++
		case "high":
			high++
		case "medium":
			medium++
		case "low":
			low++
		}
	}

	// Regional stats
	regionStats := map[string]struct{ total, critical, high int }{}
	for _, inc := range incidents {
		stats := regionStats[inc.Region]
		stats.total++
		if inc.MaxSeverity == "critical" {
			stats.critical++
		}
		if inc.MaxSeverity == "high" {
			stats.high++
		}
		regionStats[inc.Region] = stats
	}

	// Sort regions by total incidents
	type regionStat struct {
		name                   string
		total, critical, high int
	}
	var regions []regionStat
	for name, s := range regionStats {
		regions = append(regions, regionStat{name, s.total, s.critical, s.high})
	}
	sort.Slice(regions, func(i, j int) bool {
		return regions[i].total > regions[j].total
	})

	var b strings.Builder

	// Header
	fmt.Fprintf(&b, "Date: %s\n", date)
	fmt.Fprintf(&b, "Total: %d raw events → %d unique incidents across %d regions\n",
		totalRawEvents, len(incidents), len(regions))
	fmt.Fprintf(&b, "Incidents by severity: Critical: %d, High: %d, Medium: %d, Low: %d\n\n",
		critical, high, medium, low)

	// Top 50 incidents by significance
	topN := 50
	if len(incidents) < topN {
		topN = len(incidents)
	}
	b.WriteString("Top incidents by significance:\n")
	for i, inc := range incidents[:topN] {
		actors := actorList(inc.Actors)
		fmt.Fprintf(&b, "%d. [%s] %s in %s, %s — Actors: %s — %d sources, %d articles (score: %.0f)\n",
			i+1, strings.ToUpper(inc.MaxSeverity), inc.EventType, inc.LocationName, inc.Country,
			actors, inc.TotalSources, inc.TotalArticles, inc.Score)
	}
	if len(incidents) > topN {
		fmt.Fprintf(&b, "... and %d more incidents\n", len(incidents)-topN)
	}

	// Regional overview
	b.WriteString("\nRegional overview:\n")
	for _, r := range regions {
		fmt.Fprintf(&b, "- %s: %d incidents (%d critical, %d high)\n",
			r.name, r.total, r.critical, r.high)
	}

	b.WriteString("\nWrite a daily briefing covering the most significant developments.\n")
	b.WriteString("Also identify the top 5 most critical events with brief factual descriptions.\n")

	return b.String()
}

func actorList(actors map[string]bool) string {
	if len(actors) == 0 {
		return "Unknown"
	}
	list := make([]string, 0, len(actors))
	for a := range actors {
		list = append(list, a)
	}
	sort.Strings(list)
	return strings.Join(list, ", ")
}
