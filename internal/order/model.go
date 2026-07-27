package order

import (
	"time"
	"wb_test_task/internal/delivery"
	"wb_test_task/internal/item"
	"wb_test_task/internal/payment"
)

type Order struct {
	OrderUID          string            `json:"order_uid" gorm:"primaryKey;size:64"`
	TrackNumber       string            `json:"track_number"`
	Entry             string            `json:"entry"`
	Delivery          delivery.Delivery `json:"delivery" gorm:"foreignKey:OrderUID;references:OrderUID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Payment           payment.Payment   `json:"payment" gorm:"foreignKey:OrderUID;references:OrderUID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Items             []item.Item       `json:"items" gorm:"foreignKey:OrderUID;references:OrderUID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Locale            string            `json:"locale"`
	InternalSignature string            `json:"internal_signature"`
	CustomerID        string            `json:"customer_id"`
	DeliveryService   string            `json:"delivery_service"`
	ShardKey          string            `json:"shardkey"`
	SMID              int               `json:"sm_id"`
	DateCreated       time.Time         `json:"date_created"`
	OOFShard          string            `json:"oof_shard"`
}
