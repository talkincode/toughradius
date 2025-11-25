package echarts

import (
	"testing"
)

// TestAvgMetricLine tests averaging metric line data
func TestAvgMetricLine(t *testing.T) {
	tests := []struct {
		name     string
		input    []MetricLineItem
		expected int // expected number of output items
	}{
		{
			name: "Single time group",
			input: []MetricLineItem{
				{Id: 1, Time: "2024-01-01", Value: 100},
				{Id: 2, Time: "2024-01-01", Value: 200},
				{Id: 3, Time: "2024-01-01", Value: 300},
			},
			expected: 1,
		},
		{
			name: "Multiple time groups",
			input: []MetricLineItem{
				{Id: 1, Time: "2024-01-01", Value: 100},
				{Id: 2, Time: "2024-01-01", Value: 200},
				{Id: 3, Time: "2024-01-02", Value: 300},
				{Id: 4, Time: "2024-01-02", Value: 400},
			},
			expected: 2,
		},
		{
			name:     "Empty input",
			input:    []MetricLineItem{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AvgMetricLine(tt.input)

			if len(result) != tt.expected {
				t.Errorf("expected %d items, got %d", tt.expected, len(result))
			}

			// Verify result has proper structure
			for i, item := range result {
				if item.Time == "" {
					t.Errorf("item %d has empty Time", i)
				}
				if item.Value < 0 {
					t.Errorf("item %d has negative Value: %f", i, item.Value)
				}
			}
		})
	}
}

// TestAvgMetricLine_AverageCalculation tests correct average calculation
func TestAvgMetricLine_AverageCalculation(t *testing.T) {
	input := []MetricLineItem{
		{Id: 1, Time: "2024-01-01", Value: 100},
		{Id: 2, Time: "2024-01-01", Value: 200},
		{Id: 3, Time: "2024-01-01", Value: 300},
	}

	result := AvgMetricLine(input)

	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}

	// Average of 100, 200, 300 should be 200 (rounded to int64 then float64)
	expected := float64(200)
	if result[0].Value != expected {
		t.Errorf("expected average %f, got %f", expected, result[0].Value)
	}

	if result[0].Time != "2024-01-01" {
		t.Errorf("expected Time '2024-01-01', got %q", result[0].Time)
	}
}

// TestSumMetricLine tests summing metric line data
func TestSumMetricLine(t *testing.T) {
	tests := []struct {
		name     string
		input    []MetricLineItem
		expected int // expected number of output items
	}{
		{
			name: "Single time group",
			input: []MetricLineItem{
				{Id: 1, Time: "2024-01-01", Value: 100},
				{Id: 2, Time: "2024-01-01", Value: 200},
				{Id: 3, Time: "2024-01-01", Value: 300},
			},
			expected: 1,
		},
		{
			name: "Multiple time groups",
			input: []MetricLineItem{
				{Id: 1, Time: "2024-01-01", Value: 100},
				{Id: 2, Time: "2024-01-01", Value: 200},
				{Id: 3, Time: "2024-01-02", Value: 300},
				{Id: 4, Time: "2024-01-02", Value: 400},
			},
			expected: 2,
		},
		{
			name:     "Empty input",
			input:    []MetricLineItem{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SumMetricLine(tt.input)

			if len(result) != tt.expected {
				t.Errorf("expected %d items, got %d", tt.expected, len(result))
			}

			// Verify result has proper structure
			for i, item := range result {
				if item.Time == "" {
					t.Errorf("item %d has empty Time", i)
				}
				if item.Value < 0 {
					t.Errorf("item %d has negative Value: %f", i, item.Value)
				}
			}
		})
	}
}

// TestSumMetricLine_SumCalculation tests correct sum calculation
func TestSumMetricLine_SumCalculation(t *testing.T) {
	input := []MetricLineItem{
		{Id: 1, Time: "2024-01-01", Value: 100},
		{Id: 2, Time: "2024-01-01", Value: 200},
		{Id: 3, Time: "2024-01-01", Value: 300},
	}

	result := SumMetricLine(input)

	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}

	// Sum of 100, 200, 300 should be 600
	expected := float64(600)
	if result[0].Value != expected {
		t.Errorf("expected sum %f, got %f", expected, result[0].Value)
	}

	if result[0].Time != "2024-01-01" {
		t.Errorf("expected Time '2024-01-01', got %q", result[0].Time)
	}
}

// TestSumMetricLine_MultipleGroups tests sum across multiple time groups
func TestSumMetricLine_MultipleGroups(t *testing.T) {
	input := []MetricLineItem{
		{Id: 1, Time: "2024-01-01", Value: 100},
		{Id: 2, Time: "2024-01-01", Value: 200},
		{Id: 3, Time: "2024-01-02", Value: 300},
		{Id: 4, Time: "2024-01-02", Value: 400},
		{Id: 5, Time: "2024-01-03", Value: 500},
	}

	result := SumMetricLine(input)

	if len(result) != 3 {
		t.Fatalf("expected 3 results, got %d", len(result))
	}

	// Verify sums
	expectedSums := map[string]float64{
		"2024-01-01": 300,
		"2024-01-02": 700,
		"2024-01-03": 500,
	}

	for _, item := range result {
		expected, ok := expectedSums[item.Time]
		if !ok {
			t.Errorf("unexpected time group: %q", item.Time)
			continue
		}
		if item.Value != expected {
			t.Errorf("for time %q, expected sum %f, got %f", item.Time, expected, item.Value)
		}
	}
}

// TestAvgMetricLine_MultipleGroups tests average across multiple time groups
func TestAvgMetricLine_MultipleGroups(t *testing.T) {
	input := []MetricLineItem{
		{Id: 1, Time: "2024-01-01", Value: 100},
		{Id: 2, Time: "2024-01-01", Value: 200},
		{Id: 3, Time: "2024-01-02", Value: 300},
		{Id: 4, Time: "2024-01-02", Value: 400},
		{Id: 5, Time: "2024-01-02", Value: 500},
	}

	result := AvgMetricLine(input)

	if len(result) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result))
	}

	// Verify averages
	expectedAvgs := map[string]float64{
		"2024-01-01": 150, // (100+200)/2
		"2024-01-02": 400, // (300+400+500)/3
	}

	for _, item := range result {
		expected, ok := expectedAvgs[item.Time]
		if !ok {
			t.Errorf("unexpected time group: %q", item.Time)
			continue
		}
		if item.Value != expected {
			t.Errorf("for time %q, expected average %f, got %f", item.Time, expected, item.Value)
		}
	}
}

// BenchmarkAvgMetricLine benchmarks averaging operation
func BenchmarkAvgMetricLine(b *testing.B) {
	input := make([]MetricLineItem, 1000)
	for i := 0; i < 1000; i++ {
		input[i] = MetricLineItem{
			Id:    i,
			Time:  "2024-01-01",
			Value: float64(i * 10),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = AvgMetricLine(input)
	}
}

// BenchmarkSumMetricLine benchmarks summing operation
func BenchmarkSumMetricLine(b *testing.B) {
	input := make([]MetricLineItem, 1000)
	for i := 0; i < 1000; i++ {
		input[i] = MetricLineItem{
			Id:    i,
			Time:  "2024-01-01",
			Value: float64(i * 10),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SumMetricLine(input)
	}
}
