package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/asoasis/pii-redaction-api/internal/detector"
	"github.com/asoasis/pii-redaction-api/internal/model"
	"github.com/asoasis/pii-redaction-api/internal/redactor"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

type RedactHandler struct {
	pipeline *detector.Pipeline
	redactor *redactor.Redactor
}

func NewRedactHandler(pipeline *detector.Pipeline, redactor *redactor.Redactor) *RedactHandler {
	return &RedactHandler{
		pipeline: pipeline,
		redactor: redactor,
	}
}

func (h *RedactHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	var req model.RedactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	detections, err := h.pipeline.Detect(r.Context(), req.DetectionRequest)
	if err != nil {
		http.Error(w, "Detection failed", http.StatusInternalServerError)
		return
	}

	res, err := h.redactor.Redact(r.Context(), req.Text, detections, req.Mode, req.TTL)
	if err != nil {
		http.Error(w, "Redaction failed", http.StatusInternalServerError)
		return
	}

	res.ProcessingTimeMs = time.Since(start).Milliseconds()
	res.RequestID, _ = gonanoid.New()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
