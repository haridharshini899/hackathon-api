package api

import (
	"encoding/json"
	"hackathon-api/pkg/aggregator"
	"net/http"
	"time"
)

type Handler struct {
	service *aggregator.Service
}

func NewHandler(service *aggregator.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetHackathons(w http.ResponseWriter, r *http.Request) {

	mode := r.URL.Query().Get("mode")
	if mode == "" {
		mode = r.URL.Query().Get("type")
	}
	platform := r.URL.Query().Get("platform")

	data, err := h.service.GetHackathons(r.Context(), mode, platform)
	if err != nil {
		http.Error(w, "Failed to fetch data", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status":       "success",
		"count":        len(data),
		"generated_at": time.Now().Format(time.RFC3339),
		"data":         data,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Data-Sources", "devnovate,devpost,mlh,devfolio,unstop,hack2skill,reskill,whereuelevate,hackerearth")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) GetHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}
