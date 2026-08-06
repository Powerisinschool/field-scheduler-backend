package handlers

import (
	"encoding/json"
	"net/http"
)

// writeJSON centralizes the header setting and encoding process
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// You can log this encoding error internally if necessary
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

// writeError standardizes how your API returns error messages
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, BasicErrorResponse{Error: message})
}

func writeSuccess(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, BasicSuccessResponse{Message: message})
}
