package order

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"wb_test_task/pkg/response"
)

type Producer interface {
	ProduceMessage(message []byte) error
	SampleOrder() []byte
}

type HandlerDeps struct {
	OrderRepository Repository
	Producer        Producer
}

type Handler struct {
	OrderRepository Repository
	Producer        Producer
}

func NewHandler(router *http.ServeMux, deps HandlerDeps) {
	handler := &Handler{
		OrderRepository: deps.OrderRepository,
		Producer:        deps.Producer,
	}
	router.HandleFunc("GET /order/{order_uid}", handler.GetOrder())
	router.HandleFunc("POST /order/", handler.AddRandomOrder())
}

func (handler *Handler) GetOrder() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		uid := r.PathValue("order_uid")
		order, err := handler.OrderRepository.GetByUidCached(uid)
		if err != nil {
			http.Error(w, "record not found", http.StatusNotFound)
			return
		}
		response.WriteJsonResponse(w, http.StatusOK, order)
	}
}

func (handler *Handler) AddRandomOrder() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var order Order
		err := json.Unmarshal(handler.Producer.SampleOrder(), &order)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		order.OrderUID = generateRandomUid(10)

		updatedBody, err := json.Marshal(order)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := handler.Producer.ProduceMessage(updatedBody); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response.WriteJsonResponse(w, http.StatusOK, order)
	}
}

func generateRandomUid(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = CharacterSet[rand.Intn(len(CharacterSet))]
	}
	return string(b)
}

const CharacterSet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
