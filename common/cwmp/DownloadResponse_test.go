package cwmp

import (
	"strings"
	"testing"

	"github.com/talkincode/toughradius/v8/common/xmlx"
)

func TestDownloadResponse(t *testing.T) {
	downloadResponse := &DownloadResponse{
		ID:           "test-id",
		Name:         "test-name",
		Status:       200,
		StartTime:    "2023-07-01T00:00:00",
		CompleteTime: "2023-07-01T00:01:00",
	}

	xml := downloadResponse.CreateXML()
	if strings.Contains(string(xml), "test-id") == false {
		t.Errorf("XML encoding failed: %s", xml)
	}

	doc := xmlx.New()
	if err := doc.LoadBytes(xml, nil); err != nil {
		t.Errorf("XML parse failed: %v", err)
	}

	downloadResponse.Parse(doc)
	if downloadResponse.ID != "test-id" {
		t.Errorf("XML parse failed: ID")
	}
	if downloadResponse.Name != "test-name" {
		t.Errorf("XML parse failed: Name")
	}
	if downloadResponse.Status != 200 {
		t.Errorf("XML parse failed: Status")
	}
	if downloadResponse.StartTime != "2023-07-01T00:00:00" {
		t.Errorf("XML parse failed: StartTime")
	}
	if downloadResponse.CompleteTime != "2023-07-01T00:01:00" {
		t.Errorf("XML parse failed: CompleteTime")
	}
}
