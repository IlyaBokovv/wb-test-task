package frontend

import (
	"net/http"
	"os"
	"wb_test_task/pkg/response"
)

type Handler struct{}

func NewHandler(router *http.ServeMux) {
	handler := &Handler{}
	router.HandleFunc("GET /", handler.GetOrder())
}

func (handler *Handler) GetOrder() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		page, err := os.ReadFile("web/index.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		response.WriteHtmlBlob(w, page)
	}
}
