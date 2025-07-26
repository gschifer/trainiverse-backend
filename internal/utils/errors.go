package utils

import (
	"encoding/json"
	"net/http"
)

func HandleError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	errorResponse := map[string]string{
		"error": message,
	}
	// Check if there's an error encoding the response
	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		// Log the error if encoding fails
		http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
	}
}
