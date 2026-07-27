package main

import (
	"context"
	"log"
	"net/http"
	"wb_test_task/configs"
	"wb_test_task/internal/frontend"
	"wb_test_task/internal/order"
	"wb_test_task/pkg/db"
	"wb_test_task/pkg/kafka/consumer"
	"wb_test_task/pkg/kafka/producer"
)

func main() {
	conf := configs.LoadConfig()
	db := db.NewDb(conf)
	router := http.NewServeMux()

	orderRepository := order.NewOrderRepository(db)
	kafkaProducer := producer.NewProducer(conf)
	order.NewHandler(router, order.HandlerDeps{
		OrderRepository: orderRepository,
		Producer:        kafkaProducer,
	})
	frontend.NewHandler(router)

	kafkaConsumer := consumer.NewConsumer(conf, orderRepository)
	go kafkaConsumer.StartConsuming(context.Background())

	server := &http.Server{
		Addr:    ":8085",
		Handler: router,
	}

	log.Println("Server is listening on port 8085")
	server.ListenAndServe()
}
