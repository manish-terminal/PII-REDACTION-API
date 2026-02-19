package model

import "time"

// RedactionMode defines how PII should be masked.
type RedactionMode string

const (
	MaskMode     RedactionMode = "mask"
	ReplaceMode  RedactionMode = "replace"
	HashMode     RedactionMode = "hash"
	TokenizeMode RedactionMode = "tokenize"
)

// RedactionRequest represents the input for PII redaction.
type RedactionRequest struct {
	DetectionRequest
	Mode RedactionMode `json:"mode"`
	TTL  int           `json:"ttl,omitempty"` // TTL for tokens in hours (default 24h)
}

// RedactionResponse represents the output of PII redaction.
type RedactionResponse struct {
	RedactedText     string            `json:"redacted_text"`
	EntitiesFound    int               `json:"entities_found"`
	Detections       []RedactionDetail `json:"detections"`
	ProcessingTimeMs int64             `json:"processing_time_ms"`
	RequestID        string            `json:"request_id"`
}

// RedactionDetail provides info about each redacted entity.
type RedactionDetail struct {
	EntityType      string  `json:"entity_type"`
	OriginalStart   int     `json:"original_start"`
	OriginalEnd     int     `json:"original_end"`
	RedactedValue   string  `json:"redacted_value"`
	Confidence      float64 `json:"confidence"`
	DetectionMethod string  `json:"detection_method"`
}

// TokenMapping maps a token to its original PII value.
type TokenMapping struct {
	Token      string    `json:"token"`
	EntityType string    `json:"entity_type"`
	Value      string    `json:"value,omitempty"`
	ExpiresAt  time.Time `json:"expires_at"`
}

// DetokenizeRequest restores values from tokens.
type DetokenizeRequest struct {
	Text   string   `json:"text"`
	Tokens []string `json:"tokens"`
}
