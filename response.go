package cyber

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

type ErrorResponse struct {
	//响应code
	Code string `json:"code"`
	//响应描述
	Message string `json:"message"`
}

var jsonEncoderPool = &sync.Pool{
	New: func() interface{} {
		return json.NewEncoder(nil)
	},
}

func respondWithJSON(w http.ResponseWriter, r *http.Request, statusCode int, data interface{}) {
	enc := jsonEncoderPool.Get().(*json.Encoder)
	defer func() {
		err := enc.Encode(data)
		if err != nil {
			log.Printf("Error JSONResponse: %v", err)
			return
		}
		jsonEncoderPool.Put(enc)
	}()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := enc.Encode(data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func Success(w http.ResponseWriter, r *http.Request, StatusCode int, data interface{}) {
	respondWithJSON(w, r, StatusCode, data)
}
func Error(w http.ResponseWriter, r *http.Request, StatusCode int, code string, message string) {
	response := ErrorResponse{
		Code:    code,
		Message: message,
	}
	respondWithJSON(w, r, StatusCode, response)
}
