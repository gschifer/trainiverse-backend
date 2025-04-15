package utils

import (
	"encoding/json"
	"net/http"
)

func HandleError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	errorResponse := map[string]string{
		"error": message,
	}
	json.NewEncoder(w).Encode(errorResponse)
}
