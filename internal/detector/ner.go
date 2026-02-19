package detector

import (
	"context"
	"fmt"
	"strings"

	"github.com/asoasis/pii-redaction-api/internal/model"
	"github.com/jdkato/prose/v2"
)

type NERDetector struct{}

func NewNERDetector() *NERDetector {
	return &NERDetector{}
}

func (d *NERDetector) Detect(ctx context.Context, text string) ([]model.Detection, error) {
	doc, err := prose.NewDocument(text)
	if err != nil {
		return nil, fmt.Errorf("NER processing failed: %w", err)
	}

	var detections []model.Detection
	for _, ent := range doc.Entities() {
		piiType := mapProseEntityToPII(ent.Label)
		if piiType == "" {
			continue
		}

		// prose doesn't provide offsets in a straightforward way in doc.Entities()
		// We'll need to find the offset in the text. This is a bit naive but works for v1.
		// A better way would be to iterate through tokens if prose provided token offsets.
		start := strings.Index(text, ent.Text)
		if start == -1 {
			continue
		}

		detections = append(detections, model.Detection{
			EntityType:      piiType,
			Text:            ent.Text,
			Start:           start,
			End:             start + len(ent.Text),
			Confidence:      0.85, // NER base confidence
			DetectionMethod: "ner",
		})
	}
	return detections, nil
}

func mapProseEntityToPII(label string) string {
	switch label {
	case "GPE", "LOC":
		return "LOCATION"
	case "PERSON":
		return "PERSON"
	case "ORG":
		return "ORGANIZATION"
	case "DATE":
		return "DATE"
	default:
		return ""
	}
}
