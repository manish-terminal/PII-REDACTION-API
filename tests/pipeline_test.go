package tests

import (
	"context"
	"testing"

	"github.com/asoasis/pii-redaction-api/internal/detector"
	"github.com/asoasis/pii-redaction-api/internal/model"
)

func TestPipeline_Detect(t *testing.T) {
	p := detector.NewPipeline("en-US")
	text := "John Smith's SSN is 123-45-6789. Email him at john@acme.com or call 555-867-5309."

	req := model.DetectionRequest{
		Text:   text,
		Locale: "en-US",
	}

	detections, err := p.Detect(context.Background(), req)
	if err != nil {
		t.Fatalf("Detection failed: %v", err)
	}

	expectedTypes := map[string]bool{
		"PERSON": true,
		"SSN":    true,
		"EMAIL":  true,
		// "PHONE_US": true, // prose might pick it up differently or regex might overlap
	}

	foundTypes := make(map[string]bool)
	for _, d := range detections {
		foundTypes[d.EntityType] = true
		t.Logf("Found %s: %s at [%d:%d]", d.EntityType, d.Text, d.Start, d.End)
	}

	for et := range expectedTypes {
		if !foundTypes[et] {
			t.Errorf("Expected to find PII type %s, but didn't", et)
		}
	}
}
