package sync

import "github.com/conflux-888/conflux-api/internal/event"

func ClassifySeverity(goldsteinScale float64) string {
	switch {
	case goldsteinScale <= -7.0:
		return event.SeverityCritical
	case goldsteinScale <= -5.0:
		return event.SeverityHigh
	case goldsteinScale <= -2.0:
		return event.SeverityMedium
	default:
		return event.SeverityLow
	}
}
