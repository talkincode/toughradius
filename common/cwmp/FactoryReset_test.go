package cwmp

import (
	"strings"
	"testing"

	"github.com/talkincode/toughradius/v8/common/xmlx"
)

func TestFactoryReset(t *testing.T) {
	factoryReset := &FactoryReset{
		ID:     "test-id",
		Name:   "test-name",
		NoMore: 1,
	}

	xml := factoryReset.CreateXML()
	if strings.Contains(string(xml), "test-id") == false {
		t.Errorf("XML encoding failed: %s", xml)
	}

	doc := xmlx.New()
	if err := doc.LoadBytes(xml, nil); err != nil {
		t.Errorf("XML parse failed: %v", err)
	}

	factoryReset.Parse(doc)
	if factoryReset.ID != "test-id" {
		t.Errorf("XML parse failed: ID")
	}
	if factoryReset.Name != "test-name" {
		t.Errorf("XML parse failed: Name")
	}
	if factoryReset.NoMore != 1 {
		t.Errorf("XML parse failed: NoMore")
	}
}
