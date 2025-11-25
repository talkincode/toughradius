package echarts

import (
	"encoding/json"
	"sync"
	"testing"
)

// TestNewSeriesObject tests creating a new SeriesObject
func TestNewSeriesObject(t *testing.T) {
	tests := []struct {
		name       string
		seriesType string
	}{
		{"Line chart", "line"},
		{"Bar chart", "bar"},
		{"Pie chart", "pie"},
		{"Scatter chart", "scatter"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			so := NewSeriesObject(tt.seriesType)
			if so == nil {
				t.Fatal("NewSeriesObject returned nil")
			}
			if so.Type != tt.seriesType {
				t.Errorf("expected Type %q, got %q", tt.seriesType, so.Type)
			}
			if so.attrs == nil {
				t.Error("attrs map should be initialized")
			}
		})
	}
}

// TestSeriesObject_SetAttr tests setting attributes on SeriesObject
func TestSeriesObject_SetAttr(t *testing.T) {
	so := NewSeriesObject("line")

	tests := []struct {
		key   string
		value interface{}
	}{
		{"name", "Sales"},
		{"smooth", true},
		{"lineWidth", 2},
		{"color", "#FF5733"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			so.SetAttr(tt.key, tt.value)
			if val, ok := so.attrs[tt.key]; !ok {
				t.Errorf("attribute %q not set", tt.key)
			} else if val != tt.value {
				t.Errorf("expected %v, got %v", tt.value, val)
			}
		})
	}
}

// TestSeriesObject_MarshalJSON tests JSON marshaling of SeriesObject
func TestSeriesObject_MarshalJSON(t *testing.T) {
	so := NewSeriesObject("line")
	so.Data = []int{1, 2, 3, 4, 5}
	so.SetAttr("name", "Test Series")
	so.SetAttr("smooth", true)

	data, err := json.Marshal(so)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	// Unmarshal to verify structure
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Verify type field
	if result["type"] != "line" {
		t.Errorf("expected type 'line', got %v", result["type"])
	}

	// Verify name attribute
	if result["name"] != "Test Series" {
		t.Errorf("expected name 'Test Series', got %v", result["name"])
	}

	// Verify smooth attribute
	if result["smooth"] != true {
		t.Errorf("expected smooth true, got %v", result["smooth"])
	}

	// Verify data field exists
	if _, ok := result["data"]; !ok {
		t.Error("data field missing")
	}
}

// TestSeries tests creating series map
func TestSeries(t *testing.T) {
	so1 := NewSeriesObject("line")
	so1.Data = []int{1, 2, 3}

	so2 := NewSeriesObject("bar")
	so2.Data = []int{4, 5, 6}

	result := Series(so1, so2)

	if len(result) != 1 {
		t.Errorf("expected 1 key in map, got %d", len(result))
	}

	series, ok := result["series"]
	if !ok {
		t.Fatal("'series' key not found in result")
	}

	if len(series) != 2 {
		t.Errorf("expected 2 series objects, got %d", len(series))
	}

	if series[0].Type != "line" {
		t.Errorf("expected first series type 'line', got %q", series[0].Type)
	}

	if series[1].Type != "bar" {
		t.Errorf("expected second series type 'bar', got %q", series[1].Type)
	}
}

// TestNewNameValuePair tests creating a new NameValuePair
func TestNewNameValuePair(t *testing.T) {
	tests := []struct {
		name     string
		pairName string
		value    int64
	}{
		{"Positive value", "Users", 100},
		{"Zero value", "Empty", 0},
		{"Negative value", "Deficit", -50},
		{"Large value", "BigNumber", 999999999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nvp := NewNameValuePair(tt.pairName, tt.value)
			if nvp == nil {
				t.Fatal("NewNameValuePair returned nil")
			}
			if nvp.Name != tt.pairName {
				t.Errorf("expected Name %q, got %q", tt.pairName, nvp.Name)
			}
			if nvp.Value != tt.value {
				t.Errorf("expected Value %d, got %d", tt.value, nvp.Value)
			}
		})
	}
}

// TestNameValuePair_Incr tests incrementing NameValuePair value
func TestNameValuePair_Incr(t *testing.T) {
	nvp := NewNameValuePair("Counter", 0)

	// Sequential increments
	for i := 1; i <= 10; i++ {
		nvp.Incr()
		if nvp.Value != int64(i) {
			t.Errorf("after %d increments, expected Value %d, got %d", i, i, nvp.Value)
		}
	}
}

// TestNameValuePair_Incr_Concurrent tests concurrent increments (atomic operation)
func TestNameValuePair_Incr_Concurrent(t *testing.T) {
	nvp := NewNameValuePair("ConcurrentCounter", 0)
	iterations := 1000
	goroutines := 10

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				nvp.Incr()
			}
		}()
	}

	wg.Wait()

	expected := int64(goroutines * iterations)
	if nvp.Value != expected {
		t.Errorf("expected Value %d after concurrent increments, got %d", expected, nvp.Value)
	}
}

// TestNameValuePair_JSON tests JSON marshaling of NameValuePair
func TestNameValuePair_JSON(t *testing.T) {
	nvp := NewNameValuePair("TestMetric", 42)

	data, err := json.Marshal(nvp)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	var result NameValuePair
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if result.Name != nvp.Name {
		t.Errorf("expected Name %q, got %q", nvp.Name, result.Name)
	}

	if result.Value != nvp.Value {
		t.Errorf("expected Value %d, got %d", nvp.Value, result.Value)
	}
}

// TestNewTimeValues tests creating a new TimeValues
func TestNewTimeValues(t *testing.T) {
	tv := NewTimeValues()
	if tv == nil {
		t.Fatal("NewTimeValues returned nil")
	}
	// Note: value slice is initialized as nil, not empty slice
	if len(tv.value) != 0 {
		t.Errorf("expected empty value slice, got length %d", len(tv.value))
	}
}

// TestTimeValues_AddData tests adding data to TimeValues
func TestTimeValues_AddData(t *testing.T) {
	tv := NewTimeValues()

	tests := []struct {
		time  int64
		value interface{}
	}{
		{1609459200000, 100},
		{1609545600000, 200.5},
		{1609632000000, "300"},
		{1609718400000, true},
	}

	for i, tt := range tests {
		tv.AddData(tt.time, tt.value)
		if len(tv.value) != i+1 {
			t.Errorf("expected %d items, got %d", i+1, len(tv.value))
		}

		lastItem := tv.value[i]
		if len(lastItem) != 2 {
			t.Errorf("expected item with 2 elements, got %d", len(lastItem))
		}

		if lastItem[0] != tt.time {
			t.Errorf("expected time %d, got %v", tt.time, lastItem[0])
		}

		if lastItem[1] != tt.value {
			t.Errorf("expected value %v, got %v", tt.value, lastItem[1])
		}
	}
}

// TestTimeValues_MarshalJSON tests JSON marshaling of TimeValues
func TestTimeValues_MarshalJSON(t *testing.T) {
	tv := NewTimeValues()
	tv.AddData(1609459200000, 100)
	tv.AddData(1609545600000, 200)
	tv.AddData(1609632000000, 300)

	data, err := json.Marshal(tv)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	// Should produce: [[timestamp1, value1], [timestamp2, value2], ...]
	var result [][]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("expected 3 items, got %d", len(result))
	}

	// Verify first item
	if len(result[0]) != 2 {
		t.Error("expected each item to have 2 elements")
	}

	// Note: JSON unmarshals numbers as float64
	if result[0][0].(float64) != 1609459200000 {
		t.Errorf("expected time 1609459200000, got %v", result[0][0])
	}

	if result[0][1].(float64) != 100 {
		t.Errorf("expected value 100, got %v", result[0][1])
	}
}

// TestTimeValues_Empty tests marshaling empty TimeValues
func TestTimeValues_Empty(t *testing.T) {
	tv := NewTimeValues()

	data, err := json.Marshal(tv)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	expected := "null"
	if string(data) != expected {
		t.Errorf("expected %q, got %q", expected, string(data))
	}
}

// BenchmarkSeriesObject_MarshalJSON benchmarks JSON marshaling
func BenchmarkSeriesObject_MarshalJSON(b *testing.B) {
	so := NewSeriesObject("line")
	so.Data = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	so.SetAttr("name", "Benchmark")
	so.SetAttr("smooth", true)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(so)
	}
}

// BenchmarkNameValuePair_Incr benchmarks atomic increment
func BenchmarkNameValuePair_Incr(b *testing.B) {
	nvp := NewNameValuePair("Counter", 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		nvp.Incr()
	}
}

// BenchmarkTimeValues_AddData benchmarks adding data
func BenchmarkTimeValues_AddData(b *testing.B) {
	tv := NewTimeValues()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tv.AddData(int64(i), i*100)
	}
}
