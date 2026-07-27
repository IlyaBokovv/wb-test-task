package order

import (
	"log"
	"wb_test_task/internal/item"
	"wb_test_task/pkg/db"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	CacheInitialValue = 10
)

type Repository interface {
	Save(order Order) error
	GetByUid(uid string) (*Order, error)
	GetByUidCached(uid string) (*Order, error)
	PutCache(order *Order)
}

type OrderRepository struct {
	Database *db.Db
	Cache    *Cache
}

func NewOrderRepository(database *db.Db) *OrderRepository {
	repo := &OrderRepository{
		Database: database,
		Cache:    NewCache(),
	}
	err := repo.FillCache()
	if err != nil {
		log.Fatalf("Failed to fill cache %v ", err.Error())
	}
	return repo
}

func (repo *OrderRepository) Save(order Order) error {
	order.Delivery.OrderUID, order.Payment.OrderUID = order.OrderUID, order.OrderUID
	for index := range order.Items {
		order.Items[index].OrderUID = order.OrderUID
		order.Items[index].ID = 0
	}
	return repo.Database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit(clause.Associations).Create(&order).Error; err != nil {
			return err
		}
		if err := tx.
			Create(&order.Delivery).Error; err != nil {
			return err
		}
		if err := tx.
			Create(&order.Payment).Error; err != nil {
			return err
		}
		if err := tx.
			Where("order_uid = ?", order.OrderUID).
			Delete(&item.Item{}).Error; err != nil {
			return err
		}
		if len(order.Items) > 0 {
			return tx.Create(&order.Items).Error
		}
		return nil
	})
}

func (repo *OrderRepository) GetByUid(uid string) (*Order, error) {
	var order Order
	err := repo.Database.DB.
		Preload("Delivery").
		Preload("Payment").
		Preload("Items").
		First(&order, "order_uid = ?", uid).Error
	return &order, err
}

func (repo *OrderRepository) GetByUidCached(uid string) (*Order, error) {
	order, ok := repo.Cache.Get(uid)
	if !ok {
		order, err := repo.GetByUid(uid)
		if err != nil {
			return nil, err
		}
		log.Printf("order uid not found in cache, searching in db %v", order.OrderUID)
		repo.Cache.Put(order)
		return order, nil
	}
	log.Printf("order uid cache hit %v", order.OrderUID)
	return order, nil
}

func (repo *OrderRepository) PutCache(order *Order) {
	repo.Cache.Put(order)
}

func (repo *OrderRepository) GetLastNRows(limit int) ([]Order, error) {
	var orders []Order
	err := repo.Database.DB.
		Preload("Delivery").
		Preload("Payment").
		Preload("Items").
		Order("date_created DESC").
		Limit(limit).
		Find(&orders).Error
	return orders, err
}

func (repo *OrderRepository) FillCache() error {
	orders, err := repo.GetLastNRows(CacheInitialValue)
	if err != nil {
		return err
	}

	for _, order := range orders {
		repo.Cache.Put(&order)
	}
	log.Printf("Cache filled with %d item(s)\n", len(orders))
	return nil
}
