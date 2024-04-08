package cyber

import (
	"encoding/json"
	"log"
	"net/http"
)

type ErrorResponse struct {
	//响应code
	Code string `json:"code"`
	//响应描述
	Message string `json:"message"`
}

func Success(w http.ResponseWriter, r *http.Request, StatusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(StatusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error JSONResponse: %v", err)
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
		return
	}
}
func Error(w http.ResponseWriter, r *http.Request, StatusCode int, code string, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(StatusCode)
	response := ErrorResponse{
		Code:    code,
		Message: message,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error JSONResponse: %v", err)
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
		return
	}
}
