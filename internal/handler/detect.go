package handler

import (
	"net/http"

	"github.com/asoasis/pii-redaction-api/internal/detector"
)

type DetectHandler struct {
	pipeline *detector.Pipeline
}

func NewDetectHandler(pipeline *detector.Pipeline) *DetectHandler {
	return &DetectHandler{pipeline: pipeline}
}

func (h *DetectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message":"Detection endpoint reached"}`))
}
