package consumer

import (
	"errors"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wb_test_task/internal/order"
)

type fakeStore struct {
	saveErr error
	saved   []order.Order
	cached  []*order.Order
}

func (f *fakeStore) Save(o order.Order) error {
	f.saved = append(f.saved, o)
	return f.saveErr
}

func (f *fakeStore) PutCache(o *order.Order) {
	f.cached = append(f.cached, o)
}

func newTestConsumer(store OrderStore) *Consumer {
	return &Consumer{orderRepository: store, validate: validator.New()}
}

const validOrderJSON = `{
  "order_uid": "b563feb7b2b84b6test",
  "track_number": "WBILMTESTTRACK",
  "entry": "WBIL",
  "delivery": {
    "name": "Test Testov",
    "phone": "+9720000000",
    "zip": "2639809",
    "city": "Kiryat Mozkin",
    "address": "Ploshad Mira 15",
    "region": "Kraiot",
    "email": "test@gmail.com"
  },
  "payment": {
    "transaction": "b563feb7b2b84b6test",
    "request_id": "",
    "currency": "USD",
    "provider": "wbpay",
    "amount": 1817,
    "payment_dt": 1637907727,
    "bank": "alpha",
    "delivery_cost": 1500,
    "goods_total": 317,
    "custom_fee": 0
  },
  "items": [
    {
      "chrt_id": 9934930,
      "track_number": "WBILMTESTTRACK",
      "price": 453,
      "rid": "ab4219087a764ae0btest",
      "name": "Mascaras",
      "sale": 30,
      "size": "0",
      "total_price": 317,
      "nm_id": 2389212,
      "brand": "Vivienne Sabo",
      "status": 202
    }
  ],
  "locale": "en",
  "internal_signature": "",
  "customer_id": "test",
  "delivery_service": "meest",
  "shardkey": "9",
  "sm_id": 99,
  "date_created": "2021-11-26T06:22:19Z",
  "oof_shard": "1"
}`

func TestProcessMessageValidOrderSavesAndCaches(t *testing.T) {
	store := &fakeStore{}
	c := newTestConsumer(store)

	commit, err := c.processMessage([]byte(validOrderJSON))

	require.NoError(t, err)
	assert.True(t, commit)
	require.Len(t, store.saved, 1)
	assert.Equal(t, "b563feb7b2b84b6test", store.saved[0].OrderUID)
	require.Len(t, store.cached, 1)
	assert.Equal(t, "b563feb7b2b84b6test", store.cached[0].OrderUID)
}

func TestProcessMessageMalformedJSONIsDiscardedAndNotSaved(t *testing.T) {
	store := &fakeStore{}
	c := newTestConsumer(store)

	commit, err := c.processMessage([]byte(`{not valid json`))

	assert.Error(t, err)
	assert.True(t, commit)
	assert.Empty(t, store.saved)
	assert.Empty(t, store.cached)
}

func TestProcessMessageFailedValidationIsDiscardedAndNotSaved(t *testing.T) {
	store := &fakeStore{}
	c := newTestConsumer(store)

	commit, err := c.processMessage([]byte(`{"order_uid": "abc"}`))

	assert.Error(t, err)
	assert.True(t, commit)
	assert.Empty(t, store.saved)
	assert.Empty(t, store.cached)
}

func TestProcessMessageDuplicateKeyIsTreatedAsAlreadyHandled(t *testing.T) {
	store := &fakeStore{saveErr: &pgconn.PgError{Code: "23505"}}
	c := newTestConsumer(store)

	commit, err := c.processMessage([]byte(validOrderJSON))

	require.NoError(t, err)
	assert.True(t, commit)
}

func TestProcessMessageGenericDbErrorLeavesMessageUncommitted(t *testing.T) {
	store := &fakeStore{saveErr: errors.New("connection refused")}
	c := newTestConsumer(store)

	commit, err := c.processMessage([]byte(validOrderJSON))

	assert.Error(t, err)
	assert.False(t, commit)
	assert.Empty(t, store.cached)
}
