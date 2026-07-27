package order

import (
	"sync"
)

type Cache struct {
	mu     sync.RWMutex
	orders map[string]Order
}

func NewCache() *Cache {
	orders := make(map[string]Order)

	return &Cache{
		orders: orders,
	}
}

func (c *Cache) Get(id string) (*Order, bool) {
	o, ok := c.orders[id]
	return &o, ok
}

func (c *Cache) Put(order *Order) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.orders[order.OrderUID] = *order
}
