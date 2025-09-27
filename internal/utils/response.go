package utils

import (
	"encoding/json"
	"net/http"

	"github.com/AlbertGhazaly/Steganography-on-Audio-Files-with-Multiple-LSB-Method/internal/models"
)

func SendResponse(w http.ResponseWriter, success bool, message string, data any) {
	w.Header().Set("Content-Type", "application/json")
	response := models.Response{
		Success: success,
		Message: message,
		Data:    data,
	}
	json.NewEncoder(w).Encode(response)
}

func SendError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	SendResponse(w, false, message, nil)
}
