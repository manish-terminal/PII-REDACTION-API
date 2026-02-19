package model

// Detection represents a single piece of PII found in text.
type Detection struct {
	EntityType      string  `json:"entity_type"`
	Text            string  `json:"text,omitempty"` // Original text match
	Start           int     `json:"start"`          // Start offset in original text
	End             int     `json:"end"`            // End offset in original text
	Confidence      float64 `json:"confidence"`
	DetectionMethod string  `json:"detection_method"`
}

// DetectionRequest represents the input for PII detection.
type DetectionRequest struct {
	Text                string   `json:"text"`
	Locale              string   `json:"locale,omitempty"`
	EntityTypes         []string `json:"entity_types,omitempty"`
	ConfidenceThreshold float64  `json:"confidence_threshold,omitempty"`
}

// DetectionResponse represents the output of PII detection.
type DetectionResponse struct {
	EntitiesFound    int         `json:"entities_found"`
	Detections       []Detection `json:"detections"`
	RiskSummary      RiskSummary `json:"risk_summary,omitempty"`
	ProcessingTimeMs int64       `json:"processing_time_ms"`
	RequestID        string      `json:"request_id"`
}

// RiskSummary provides a high-level view of compliance risk.
type RiskSummary struct {
	HIPAARelevant int `json:"hipaa_relevant"`
	GDPRRelevant  int `json:"gdpr_relevant"`
	PCIRelevant   int `json:"pci_relevant"`
}
