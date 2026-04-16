package summary

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/conflux-888/conflux-api/internal/common/gemini"
	"github.com/conflux-888/conflux-api/internal/event"
	genailib "github.com/google/generative-ai-go/genai"
)

const modelName = "gemini-2.5-flash"

const systemInstruction = `You are a senior conflict correspondent writing for an international news wire service like Reuters or BBC World Service. You produce daily conflict briefings that are factual, authoritative, and written in clear professional English.

Rules:
- Write in third person, past tense
- Lead with the most significant developments
- Group related events by region, flowing naturally between paragraphs
- Include specific details: locations, actor names, number of incidents
- End with a brief statistical context paragraph
- Do NOT use bullet points, headers, or lists — write continuous prose
- Do NOT speculate or editorialize — report only what the data shows
- Length: 400-800 words depending on event volume`

type LLMOutput struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

var summarySchema = &genailib.Schema{
	Type: genailib.TypeObject,
	Properties: map[string]*genailib.Schema{
		"title":   {Type: genailib.TypeString, Description: "News headline for the daily briefing"},
		"content": {Type: genailib.TypeString, Description: "Full news article in continuous prose, 400-800 words"},
	},
	Required: []string{"title", "content"},
}

func generateSummary(ctx context.Context, client *gemini.Client, date string, events []event.Event) (*LLMOutput, int, int, error) {
	prompt := buildPrompt(date, events)

	var output LLMOutput
	resp, err := client.GenerateJSON(ctx, gemini.GenerateRequest{
		Model:             modelName,
		SystemInstruction: systemInstruction,
		Prompt:            prompt,
		Temperature:       0.2,
		ResponseSchema:    summarySchema,
	}, &output)
	if err != nil {
		return nil, 0, 0, err
	}

	return &output, resp.PromptTokens, resp.CompletionTokens, nil
}

func buildPrompt(date string, events []event.Event) string {
	var critical, high, medium, low int
	for _, e := range events {
		switch e.Severity {
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

	regionEvents := map[string][]event.Event{}
	for _, e := range events {
		region := RegionForCountry(e.Country)
		regionEvents[region] = append(regionEvents[region], e)
	}

	type regionEntry struct {
		name   string
		events []event.Event
	}
	var regions []regionEntry
	for name, evts := range regionEvents {
		regions = append(regions, regionEntry{name, evts})
	}
	sort.Slice(regions, func(i, j int) bool {
		return len(regions[i].events) > len(regions[j].events)
	})

	var b strings.Builder
	fmt.Fprintf(&b, "Date: %s\n", date)
	fmt.Fprintf(&b, "Total: %d events | Critical: %d, High: %d, Medium: %d, Low: %d\n\n", len(events), critical, high, medium, low)

	for _, r := range regions {
		fmt.Fprintf(&b, "Region: %s (%d events)\n", r.name, len(r.events))

		sort.Slice(r.events, func(i, j int) bool {
			return severityOrder(r.events[i].Severity) < severityOrder(r.events[j].Severity)
		})

		limit := 20
		if len(r.events) < limit {
			limit = len(r.events)
		}
		for _, e := range r.events[:limit] {
			actors := strings.Join(e.Actors, ", ")
			if actors == "" {
				actors = "Unknown"
			}
			fmt.Fprintf(&b, "- [%s] %s — Actors: %s — %d sources, %d articles\n",
				strings.ToUpper(e.Severity), e.Title, actors, e.NumSources, e.NumArticles)
		}
		if len(r.events) > limit {
			fmt.Fprintf(&b, "- ... and %d more events\n", len(r.events)-limit)
		}
		b.WriteString("\n")
	}

	return b.String()
}

func severityOrder(s string) int {
	switch s {
	case "critical":
		return 0
	case "high":
		return 1
	case "medium":
		return 2
	default:
		return 3
	}
}
