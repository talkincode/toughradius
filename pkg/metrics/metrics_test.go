package metrics

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitMetrics(t *testing.T) {
	err := InitMetrics("")
	assert.NoError(t, err)
	assert.NotNil(t, GetStore())
}

func TestCounter(t *testing.T) {
	err := InitMetrics("")
	assert.NoError(t, err)

	store := GetStore()
	counter := store.Counter("test_counter")

	assert.Equal(t, int64(0), counter.Value())

	counter.Inc()
	assert.Equal(t, int64(1), counter.Value())

	counter.Add(5)
	assert.Equal(t, int64(6), counter.Value())

	counter.Reset()
	assert.Equal(t, int64(0), counter.Value())
}

func TestGauge(t *testing.T) {
	err := InitMetrics("")
	assert.NoError(t, err)

	store := GetStore()
	gauge := store.Gauge("test_gauge")

	assert.Equal(t, int64(0), gauge.Value())

	gauge.Set(100)
	assert.Equal(t, int64(100), gauge.Value())

	gauge.Set(50)
	assert.Equal(t, int64(50), gauge.Value())
}

func TestConvenienceFunctions(t *testing.T) {
	err := InitMetrics("")
	assert.NoError(t, err)

	Inc("requests")
	Inc("requests")
	Inc("requests")
	assert.Equal(t, int64(3), GetCounter("requests"))

	Add("bytes", 1024)
	assert.Equal(t, int64(1024), GetCounter("bytes"))

	SetGauge("connections", 42)
	assert.Equal(t, int64(42), GetStore().GetGaugeValue("connections"))
}

func TestGetAllCounters(t *testing.T) {
	err := InitMetrics("")
	assert.NoError(t, err)

	store := GetStore()
	store.Counter("counter1").Add(10)
	store.Counter("counter2").Add(20)

	all := store.GetAllCounters()
	assert.Equal(t, int64(10), all["counter1"])
	assert.Equal(t, int64(20), all["counter2"])
}

func TestGetAllGauges(t *testing.T) {
	err := InitMetrics("")
	assert.NoError(t, err)

	store := GetStore()
	store.Gauge("gauge1").Set(100)
	store.Gauge("gauge2").Set(200)

	all := store.GetAllGauges()
	assert.Equal(t, int64(100), all["gauge1"])
	assert.Equal(t, int64(200), all["gauge2"])
}

func TestConcurrentAccess(t *testing.T) {
	err := InitMetrics("")
	assert.NoError(t, err)

	store := GetStore()
	counter := store.Counter("concurrent_counter")

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Inc()
		}()
	}
	wg.Wait()

	assert.Equal(t, int64(100), counter.Value())
}

func TestClose(t *testing.T) {
	err := InitMetrics("")
	assert.NoError(t, err)

	err = Close()
	assert.NoError(t, err)
}
