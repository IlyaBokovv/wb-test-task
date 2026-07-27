package order

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCachePutThenGet(t *testing.T) {
	cache := NewCache()
	o := &Order{OrderUID: "abc123", TrackNumber: "WBILTEST"}

	cache.Put(o)

	got, ok := cache.Get("abc123")
	require.True(t, ok)
	assert.Equal(t, o.OrderUID, got.OrderUID)
	assert.Equal(t, o.TrackNumber, got.TrackNumber)
}

func TestCacheGetMiss(t *testing.T) {
	cache := NewCache()

	_, ok := cache.Get("doesnt-exists")

	assert.False(t, ok)
}

func TestCachePutOverwritesExistingEntry(t *testing.T) {
	cache := NewCache()
	cache.Put(&Order{OrderUID: "abc123", TrackNumber: "OLD"})

	cache.Put(&Order{OrderUID: "abc123", TrackNumber: "NEW"})

	got, ok := cache.Get("abc123")
	require.True(t, ok)
	assert.Equal(t, "NEW", got.TrackNumber)
}
