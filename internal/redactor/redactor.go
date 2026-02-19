package redactor

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/asoasis/pii-redaction-api/internal/model"
	"github.com/asoasis/pii-redaction-api/internal/store"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

type Redactor struct {
	store *store.DynamoDBStore
}

func NewRedactor(store *store.DynamoDBStore) *Redactor {
	return &Redactor{store: store}
}

func (r *Redactor) Redact(ctx context.Context, text string, detections []model.Detection, mode model.RedactionMode, ttlHours int) (model.RedactionResponse, error) {
	if ttlHours == 0 {
		ttlHours = 24
	}

	res := model.RedactionResponse{
		EntitiesFound: len(detections),
		Detections:    make([]model.RedactionDetail, 0, len(detections)),
	}

	// Work backwards to maintain offsets
	redactedText := text
	for i := len(detections) - 1; i >= 0; i-- {
		det := detections[i]
		redactedValue, err := r.applyMode(ctx, det, mode, ttlHours)
		if err != nil {
			return res, err
		}

		redactedText = redactedText[:det.Start] + redactedValue + redactedText[det.End:]

		res.Detections = append(res.Detections, model.RedactionDetail{
			EntityType:      det.EntityType,
			OriginalStart:   det.Start,
			OriginalEnd:     det.End,
			RedactedValue:   redactedValue,
			Confidence:      det.Confidence,
			DetectionMethod: det.DetectionMethod,
		})
	}

	res.RedactedText = redactedText
	return res, nil
}

func (r *Redactor) applyMode(ctx context.Context, det model.Detection, mode model.RedactionMode, ttlHours int) (string, error) {
	switch mode {
	case model.MaskMode:
		return strings.Repeat("*", len(det.Text)), nil
	case model.ReplaceMode:
		return "[" + det.EntityType + "]", nil
	case model.HashMode:
		hash := sha256.Sum256([]byte(det.Text))
		return hex.EncodeToString(hash[:8]), nil
	case model.TokenizeMode:
		token, err := gonanoid.New()
		if err != nil {
			return "", err
		}
		token = "tok_" + token
		err = r.store.StoreToken(ctx, model.TokenMapping{
			Token:      token,
			EntityType: det.EntityType,
			Value:      det.Text,
			ExpiresAt:  time.Now().Add(time.Duration(ttlHours) * time.Hour),
		})
		if err != nil {
			return "", err
		}
		return token, nil
	default:
		return "[" + det.EntityType + "]", nil
	}
}

func (r *Redactor) Detokenize(ctx context.Context, text string, tokens []string) (string, error) {
	detokenizedText := text
	for _, token := range tokens {
		mapping, err := r.store.GetToken(ctx, token)
		if err != nil {
			// Skip if token not found or expired
			continue
		}
		detokenizedText = strings.ReplaceAll(detokenizedText, token, mapping.Value)
	}
	return detokenizedText, nil
}
