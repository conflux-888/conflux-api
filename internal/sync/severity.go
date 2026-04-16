package sync

import "github.com/conflux-888/conflux-api/internal/event"

// ClassifySeverity determines severity based on num_articles (media attention)
// and event_root_code (type of violence).
// GoldsteinScale is not used because root codes 18-20 always score -9 to -10
// with no meaningful variation.
func ClassifySeverity(eventRootCode string, numArticles int) string {
	// Military force (root 20) with significant coverage
	if eventRootCode == "20" && numArticles >= 5 {
		return event.SeverityCritical
	}

	// High media attention = significant event
	if numArticles >= 15 {
		return event.SeverityCritical
	}
	if numArticles >= 8 {
		return event.SeverityHigh
	}
	if numArticles >= 3 {
		return event.SeverityMedium
	}

	return event.SeverityLow
}
