package response

import (
	"encoding/json"
	"net/http"
)

func WriteJsonResponse(w http.ResponseWriter, statusCode int, res any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(res)
}

func WriteHtmlBlob(w http.ResponseWriter, res []byte) {
	w.Header().Set("Content-Type", "text/html")
	w.Write(res)
}
