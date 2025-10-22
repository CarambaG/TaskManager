package handlers

import (
	"encoding/json"
	"net/http"
)

// Простой health check
func (a *App) HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	response := struct {
		Status string `json:"status"`
	}{
		Status: "OK",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
