package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

type ErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Log     string `json:"log,omitempty"`
	} `json:"error"`
}

func RespondWithError(w http.ResponseWriter, statusCode int, userMessage string, logMessage string) {
	log.Println("[API ERROR]", logMessage)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{}
	response.Error.Message = userMessage
	response.Error.Log = logMessage
	log.Println("[API ERROR]", logMessage)
	json.NewEncoder(w).Encode(response)
}
