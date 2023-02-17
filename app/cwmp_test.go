package app

import (
	"testing"

	"github.com/talkincode/toughradius/models"
)

func TestMatchTaskTags(t *testing.T) {
	tests := []struct {
		cpeTaskTags    string
		configTaskTags string
		expectedMatch  bool
	}{
		// Test cases with matching tags
		{"tag1, tag2, tag3", "tag1", true},
		{"tag1, tag2, tag3", "tag2", true},
		{"tag1, tag2, tag3", "tag3", true},
		{"tag1, tag2, tag3", "tag1,tag2", true},
		{"tag1, tag2, tag3", "tag2,tag3", true},
		{"tag1, tag2, tag3", "tag1,tag2,tag3", true},
		{"tag1, tag2, tag3", "tag4,tag2", true},
		{"tag1, tag2, tag3", "", true},
		{"tag1, tag2, tag3", " ", true},

		// Test cases with non-matching tags
		{"tag1, tag2, tag3", "tag4", false},
		{"tag1, tag2, tag3", "tag4,tag5", false},
		{"", "tag1", false},
	}

	for _, test := range tests {
		result := MatchTaskTags(test.cpeTaskTags, test.configTaskTags)
		if result != test.expectedMatch {
			t.Errorf("MatchTaskTags(%q, %q) returned %v, but expected %v", test.cpeTaskTags, test.configTaskTags, result, test.expectedMatch)
		}
	}
}

func TestMatchDevice(t *testing.T) {
	c := models.NetCpe{
		Oui:             "test-oui",
		ProductClass:    "test-product-class",
		SoftwareVersion: "test-software-version",
	}

	tests := []struct {
		oui             string
		productClass    string
		softwareVersion string
		expectedMatch   bool
	}{
		// Test cases with matching values
		{"test-oui", "test-product-class", "test-software-version", true},
		{"test-oui", "", "", true},
		{"", "test-product-class", "", true},
		{"", "", "test-software-version", true},
		{"any", "test-product-class", "test-software-version", true},
		{"test-oui,test-oui2", "", "", true},
		{"", "test-product-class,test-product-class2", "", true},
		{"", "", "test-software-version,test-software-version2", true},

		// Test cases with non-matching values
		{"test-oui2", "", "", false},
		{"", "test-product-class2", "", false},
		{"", "", "test-software-version2", false},
		{"test-oui2", "test-product-class", "test-software-version", false},
		{"test-oui", "test-product-class2", "test-software-version", false},
		{"test-oui", "test-product-class", "test-software-version2", false},
	}

	for _, test := range tests {
		result := MatchDevice(c, test.oui, test.productClass, test.softwareVersion)
		if result != test.expectedMatch {
			t.Errorf("MatchDevice(%v, %q, %q, %q) returned %v, but expected %v", c, test.oui, test.productClass, test.softwareVersion, result, test.expectedMatch)
		}
	}
}
