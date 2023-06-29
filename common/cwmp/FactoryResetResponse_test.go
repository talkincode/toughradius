package cwmp

import (
	"strings"
	"testing"

	"github.com/talkincode/toughradius/v8/common/xmlx"
)

func TestFactoryResetResponse(t *testing.T) {
	factoryResetResponse := &FactoryResetResponse{
		ID:   "test-id",
		Name: "test-name",
	}

	xml := factoryResetResponse.CreateXML()
	if strings.Contains(string(xml), "test-id") == false {
		t.Errorf("XML encoding failed: %s", xml)
	}

	doc := xmlx.New()
	if err := doc.LoadBytes(xml, nil); err != nil {
		t.Errorf("XML parse failed: %v", err)
	}

	factoryResetResponse.Parse(doc)
	if factoryResetResponse.ID != "test-id" {
		t.Errorf("XML parse failed: ID")
	}
	if factoryResetResponse.Name != "test-name" {
		t.Errorf("XML parse failed: Name")
	}
}
