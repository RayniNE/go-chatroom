package utils

import (
	"encoding/json"
	"net/http"

	"github.com/raynine/go-chatroom/models"
)

func EncodeErrorResponse(w http.ResponseWriter, err models.CustomError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Code)
	json.NewEncoder(w).Encode(err.Message)
}
