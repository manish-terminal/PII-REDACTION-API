package detector

import (
	"context"

	"github.com/asoasis/pii-redaction-api/internal/model"
)

type RegexDetector struct {
	defaultLocale string
}

func NewRegexDetector(locale string) *RegexDetector {
	if locale == "" {
		locale = "en-US"
	}
	return &RegexDetector{defaultLocale: locale}
}

func (d *RegexDetector) Detect(ctx context.Context, text string, locale string) ([]model.Detection, error) {
	if locale == "" {
		locale = d.defaultLocale
	}

	patterns, ok := localePatterns[locale]
	if !ok {
		// Fallback to en-US if locale not found
		patterns = localePatterns["en-US"]
	}

	var detections []model.Detection
	for _, p := range patterns {
		matches := p.Pattern.FindAllStringIndex(text, -1)
		for _, m := range matches {
			matchText := text[m[0]:m[1]]
			if p.Validator != nil && !p.Validator(matchText) {
				continue
			}

			detections = append(detections, model.Detection{
				EntityType:      p.Name,
				Text:            matchText,
				Start:           m[0],
				End:             m[1],
				Confidence:      p.Confidence,
				DetectionMethod: "regex",
			})
		}
	}

	return detections, nil
}
