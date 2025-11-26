/*
 * Copyright (c) 2024-2025 TalkingCode
 * Licensed under the MIT License. See LICENSE file in the project root for details.
 */

package common

import (
	"testing"
)

func TestUUID(t *testing.T) {
	t.Log(UUID())
}

func TestUUIDint64(t *testing.T) {
	t.Log(UUIDint64())
}

func TestSha256HashWithSalt(t *testing.T) {
	hash := Sha256HashWithSalt("password", "salt")
	if hash == "" {
		t.Error("Expected non-empty hash")
	}
	t.Log(hash)
}

func TestInSlice(t *testing.T) {
	tests := []struct {
		name     string
		v        string
		sl       []string
		expected bool
	}{
		{"found", "a", []string{"a", "b", "c"}, true},
		{"not found", "d", []string{"a", "b", "c"}, false},
		{"empty slice", "a", []string{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InSlice(tt.v, tt.sl); got != tt.expected {
				t.Errorf("InSlice() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected bool
	}{
		{"empty string", "", true},
		{"non-empty string", "hello", false},
		{"zero int", 0, true},
		{"non-zero int", 1, false},
		{"nil", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsEmpty(tt.value); got != tt.expected {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsEmptyOrNA(t *testing.T) {
	tests := []struct {
		name     string
		val      string
		expected bool
	}{
		{"empty", "", true},
		{"NA", "N/A", true},
		{"value", "hello", false},
		{"whitespace", "  ", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsEmptyOrNA(tt.val); got != tt.expected {
				t.Errorf("IsEmptyOrNA() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIfEmptyStr(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		defval   string
		expected string
	}{
		{"empty returns default", "", "default", "default"},
		{"non-empty returns src", "value", "default", "value"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IfEmptyStr(tt.src, tt.defval); got != tt.expected {
				t.Errorf("IfEmptyStr() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestToJson(t *testing.T) {
	data := map[string]string{"key": "value"}
	result := ToJson(data)
	if result == "" {
		t.Error("Expected non-empty JSON string")
	}
	t.Log(result)
}

func TestTrimBytes(t *testing.T) {
	input := []byte("\xef\xbb\xbfhello")
	expected := []byte("hello")
	result := TrimBytes(input)
	if string(result) != string(expected) {
		t.Errorf("TrimBytes() = %v, want %v", result, expected)
	}
}
