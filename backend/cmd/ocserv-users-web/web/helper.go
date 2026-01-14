package web

import (
	"encoding/json"
	"net/http"
)

// Helper functions
func sendJSON(w http.ResponseWriter, code int, resp Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(resp)
}

func sendError(w http.ResponseWriter, code int, message string) {
	sendJSON(w, code, Response{Code: code, Message: message})
}

func sendSuccess(w http.ResponseWriter, message string, data interface{}) {
	sendJSON(w, http.StatusOK, Response{Code: 0, Message: message, Data: data})
}
