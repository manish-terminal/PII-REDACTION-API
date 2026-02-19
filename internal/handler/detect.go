package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/asoasis/pii-redaction-api/internal/detector"
	"github.com/asoasis/pii-redaction-api/internal/model"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

type DetectHandler struct {
	pipeline *detector.Pipeline
}

func NewDetectHandler(pipeline *detector.Pipeline) *DetectHandler {
	return &DetectHandler{pipeline: pipeline}
}

func (h *DetectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	var req model.DetectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	detections, err := h.pipeline.Detect(r.Context(), req)
	if err != nil {
		http.Error(w, "Detection failed", http.StatusInternalServerError)
		return
	}

	res := model.DetectionResponse{
		EntitiesFound:    len(detections),
		Detections:       detections,
		ProcessingTimeMs: time.Since(start).Milliseconds(),
		RequestID:        "",
	}
	res.RequestID, _ = gonanoid.New()

	// Calculate RiskSummary
	for _, d := range detections {
		switch d.EntityType {
		case "SSN", "PHONE_US", "PERSON", "DATE":
			res.RiskSummary.HIPAARelevant++
			res.RiskSummary.GDPRRelevant++
		case "EMAIL", "LOCATION":
			res.RiskSummary.GDPRRelevant++
		case "CREDIT_CARD":
			res.RiskSummary.PCIRelevant++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
