package response

import (
	"encoding/json"
	"net/http"
)

func WriteJsonResponse(w http.ResponseWriter, statusCode int, res any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(res)
	if err != nil {
		return
	}
}

func WriteHtmlBlob(w http.ResponseWriter, res []byte) {
	w.Header().Set("Content-Type", "text/html")
	_, err := w.Write(res)
	if err != nil {
		return
	}
}
