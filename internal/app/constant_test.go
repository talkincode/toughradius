package app

import "testing"

func TestConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"SystemTitle", ConfigSystemTitle, "SystemTitle"},
		{"SystemTheme", ConfigSystemTheme, "SystemTheme"},
		{"SystemLoginRemark", ConfigSystemLoginRemark, "SystemLoginRemark"},
		{"SystemLoginSubtitle", ConfigSystemLoginSubtitle, "SystemLoginSubtitle"},
		{"RadiusIgnorePwd", ConfigRadiusIgnorePwd, "RadiusIgnorePwd"},
		{"AccountingHistoryDays", ConfigRadiusAccountingHistoryDays, "AccountingHistoryDays"},
		{"AcctInterimInterval", ConfigRadiusAcctInterimInterval, "AcctInterimInterval"},
		{"RadiusEapMethod", ConfigRadiusEapMethod, "RadiusEapMethod"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("Expected constant '%s', got '%s'", tt.expected, tt.constant)
			}
		})
	}
}

func TestConfigConstantsArray(t *testing.T) {
	expectedLength := 8
	if len(ConfigConstants) != expectedLength {
		t.Errorf("Expected ConfigConstants to have %d elements, got %d", expectedLength, len(ConfigConstants))
	}

	// Verify all constants are present
	expectedConstants := []string{
		ConfigSystemTitle,
		ConfigSystemTheme,
		ConfigSystemLoginRemark,
		ConfigSystemLoginSubtitle,
		ConfigRadiusIgnorePwd,
		ConfigRadiusAccountingHistoryDays,
		ConfigRadiusAcctInterimInterval,
		ConfigRadiusEapMethod,
	}

	for i, expected := range expectedConstants {
		if i >= len(ConfigConstants) {
			t.Errorf("ConfigConstants missing element at index %d", i)
			continue
		}
		if ConfigConstants[i] != expected {
			t.Errorf("At index %d, expected '%s', got '%s'", i, expected, ConfigConstants[i])
		}
	}
}

func TestConfigConstantsUniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for _, constant := range ConfigConstants {
		if seen[constant] {
			t.Errorf("Duplicate constant found: '%s'", constant)
		}
		seen[constant] = true
	}
}
