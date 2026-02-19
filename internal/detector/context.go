package detector

import (
	"context"
	"strings"

	"github.com/asoasis/pii-redaction-api/internal/model"
)

type ContextAnalyzer struct{}

func NewContextAnalyzer() *ContextAnalyzer {
	return &ContextAnalyzer{}
}

func (a *ContextAnalyzer) Refine(ctx context.Context, text string, detections []model.Detection) []model.Detection {
	var refined []model.Detection
	for _, det := range detections {
		// Example refinement: Boost confidence for SSN if keywords are nearby
		if det.EntityType == "SSN" {
			surrounding := getSurroundingText(text, det.Start, 30)
			if containsAny(surrounding, "ssn", "social security", "tax id") {
				det.Confidence = min(det.Confidence*1.1, 1.0)
			}
		}

		// Disambiguate: "John Deere" (org) vs "Dear John" (person)
		if det.EntityType == "PERSON" {
			if strings.Contains(strings.ToLower(det.Text), "deere") || strings.Contains(strings.ToLower(det.Text), "inc") {
				det.EntityType = "ORGANIZATION"
			}
		}

		// Filter out very low confidence
		if det.Confidence >= 0.60 {
			refined = append(refined, det)
		}
	}
	return refined
}

func getSurroundingText(text string, offset int, window int) string {
	start := maxInt(0, offset-window)
	end := minInt(len(text), offset+window)
	return strings.ToLower(text[start:end])
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func containsAny(text string, keywords ...string) bool {
	for _, k := range keywords {
		if strings.Contains(text, k) {
			return true
		}
	}
	return false
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
