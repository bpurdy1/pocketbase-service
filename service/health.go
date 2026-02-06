package service

import (
	"encoding/json"
	"net/http"
)

type HealthService struct{}

func NewHealthService() *HealthService {
	return &HealthService{}
}

func (s *HealthService) Name() string {
	return "health"
}

func (s *HealthService) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/custom/health", s.handleHealth)
}

func (s *HealthService) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}
