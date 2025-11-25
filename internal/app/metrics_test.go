package app

import "testing"

func TestGetSystemMetrics(t *testing.T) {
	metrics := GetSystemMetrics()

	if metrics == nil {
		t.Error("GetSystemMetrics returned nil")
	}

	// Currently returns empty map, verify type is correct
	if _, ok := metrics["test"]; ok {
		t.Error("Empty metrics should not contain any keys")
	}

	// Verify it's a proper map
	metrics["test"] = 123
	if metrics["test"] != 123 {
		t.Error("Map assignment failed")
	}
}
