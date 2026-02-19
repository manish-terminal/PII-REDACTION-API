package handler

import (
	"encoding/json"
	"net/http"

	"github.com/asoasis/pii-redaction-api/internal/redactor"
)

type DetokenizeHandler struct {
	redactor *redactor.Redactor
}

func NewDetokenizeHandler(redactor *redactor.Redactor) *DetokenizeHandler {
	return &DetokenizeHandler{redactor: redactor}
}

func (h *DetokenizeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Text   string   `json:"text"`
		Tokens []string `json:"tokens"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	detokenized, err := h.redactor.Detokenize(r.Context(), req.Text, req.Tokens)
	if err != nil {
		http.Error(w, "Detokenization failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"detokenized_text": detokenized,
	})
}
