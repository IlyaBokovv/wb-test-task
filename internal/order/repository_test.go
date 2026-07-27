package order

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"wb_test_task/internal/delivery"
	"wb_test_task/internal/item"
	"wb_test_task/internal/payment"
	"wb_test_task/pkg/db"
)

func newTestRepository(t *testing.T) *OrderRepository {
	t.Helper()

	name := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	// spin up a new instance for every test
	// sqllite is single threaded so rows could be shared between tests
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", name)
	gormDB, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	require.NoError(t, gormDB.AutoMigrate(&Order{}, &delivery.Delivery{}, &payment.Payment{}, &item.Item{}))

	return &OrderRepository{Database: &db.Db{DB: gormDB}, Cache: NewCache()}
}

func sampleOrder(uid string, createdAt time.Time) Order {
	return Order{
		OrderUID:    uid,
		TrackNumber: "WBILMTESTTRACK",
		Entry:       "WBIL",
		Delivery: delivery.Delivery{
			Name:    "Test Testov",
			Phone:   "+9720000000",
			Zip:     "2639809",
			City:    "Kiryat Mozkin",
			Address: "Ploshad Mira 15",
			Region:  "Kraiot",
			Email:   "test@gmail.com",
		},
		Payment: payment.Payment{
			Transaction:  uid,
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       1817,
			PaymentDT:    1637907727,
			Bank:         "alpha",
			DeliveryCost: 1500,
			GoodsTotal:   317,
		},
		Items: []item.Item{
			{ChrtID: 9934930, TrackNumber: "WBILMTESTTRACK", Price: 453, RID: "ab4219087a764ae0btest", Name: "Mascaras", Sale: 30, Size: "0", TotalPrice: 317, NMID: 2389212, Brand: "Vivienne Sabo", Status: 202},
		},
		Locale:      "en",
		CustomerID:  "test",
		DateCreated: createdAt,
	}
}

func TestRepositorySaveAndGetByUid(t *testing.T) {
	repo := newTestRepository(t)
	want := sampleOrder("order-1", time.Now())

	require.NoError(t, repo.Save(want))

	got, err := repo.GetByUid("order-1")
	require.NoError(t, err)
	assert.Equal(t, want.OrderUID, got.OrderUID)
	assert.Equal(t, want.Delivery.Name, got.Delivery.Name)
	assert.Equal(t, want.Payment.Amount, got.Payment.Amount)
	assert.Equal(t, want.Items[0].RID, got.Items[0].RID)
}

func TestRepositoryGetByUidNotFound(t *testing.T) {
	repo := newTestRepository(t)

	_, err := repo.GetByUid("missing")

	assert.Error(t, err)
}

func TestRepositoryGetByUidCachedFillsCacheOnMiss(t *testing.T) {
	repo := newTestRepository(t)
	require.NoError(t, repo.Save(sampleOrder("order-1", time.Now())))

	_, ok := repo.Cache.Get("order-1")
	require.False(t, ok)

	got, err := repo.GetByUidCached("order-1")
	require.NoError(t, err)
	assert.Equal(t, "order-1", got.OrderUID)

	_, ok = repo.Cache.Get("order-1")
	assert.True(t, ok)
}

func TestRepository_FillCache_OrdersByDateCreatedDescending(t *testing.T) {
	repo := newTestRepository(t)
	now := time.Now()
	require.NoError(t, repo.Save(sampleOrder("oldest", now.Add(-2*time.Hour))))
	require.NoError(t, repo.Save(sampleOrder("newest", now)))
	require.NoError(t, repo.Save(sampleOrder("middle", now.Add(-1*time.Hour))))

	require.NoError(t, repo.FillCache())

	newest, ok := repo.Cache.Get("newest")
	require.True(t, ok)
	assert.Equal(t, "newest", newest.OrderUID)

	orders, err := repo.GetLastNRows(CacheInitialValue)
	require.NoError(t, err)
	require.Len(t, orders, 3)
	assert.Equal(t, "newest", orders[0].OrderUID)
	assert.Equal(t, "middle", orders[1].OrderUID)
	assert.Equal(t, "oldest", orders[2].OrderUID)
}
