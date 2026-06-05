package adminapi

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFlexibleTime(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"rfc3339", "2025-11-01T21:16:00Z", false},
		{"datetime-local", "2025-11-01T21:16", false},
		{"date only", "2025-11-01", false},
		{"empty", "", true},
		{"garbage", "not-a-time", true},
		{"partial", "2025-13-40", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFlexibleTime(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				assert.True(t, got.IsZero())
				return
			}
			require.NoError(t, err)
			assert.False(t, got.IsZero())
		})
	}
}

func TestParseFlexibleTime_RFC3339Value(t *testing.T) {
	got, err := parseFlexibleTime("2025-11-01T21:16:00Z")
	require.NoError(t, err)
	assert.Equal(t, time.Date(2025, 11, 1, 21, 16, 0, 0, time.UTC), got.UTC())
}

func TestEscapeLikePattern(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"plain", "plain"},
		{"50%", "50\\%"},
		{"a_b", "a\\_b"},
		{"back\\slash", "back\\\\slash"},
		{"%_\\", "\\%\\_\\\\"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, escapeLikePattern(tt.input))
		})
	}
}
