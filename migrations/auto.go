package main

import (
	"os"
	"wb_test_task/configs"
	"wb_test_task/internal/delivery"
	"wb_test_task/internal/item"
	"wb_test_task/internal/order"
	"wb_test_task/internal/payment"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	err := godotenv.Load(configs.DefaultEnvFileName)
	if err != nil {
		panic(err)
	}
	db, err := gorm.Open(postgres.Open(os.Getenv("DSN")), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		panic(err)
	}
	err = db.AutoMigrate(
		&order.Order{},
		&delivery.Delivery{},
		&payment.Payment{},
		&item.Item{},
	)
	if err != nil {
		panic("Migration failed: " + err.Error())
	}
}
