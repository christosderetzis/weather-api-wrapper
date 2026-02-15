package dto

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse is the standard error response structure
type ErrorResponse struct {
	Message string `json:"message"`
}

func WriteErrorJSON(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResponse := ErrorResponse{
		Message: message,
	}

	_ = json.NewEncoder(w).Encode(errorResponse)
}
