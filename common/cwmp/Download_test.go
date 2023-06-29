package cwmp

import (
	"strings"
	"testing"

	"github.com/talkincode/toughradius/v8/common/xmlx"
)

func TestDownload(t *testing.T) {
	download := &Download{
		ID:             "test-id",
		Name:           "test-name",
		NoMore:         1,
		CommandKey:     "test-command",
		FileType:       "test-filetype",
		URL:            "http://test.com",
		Username:       "test-username",
		Password:       "test-password",
		FileSize:       1024,
		TargetFileName: "test-filename",
		DelaySeconds:   60,
		SuccessURL:     "http://success.com",
		FailureURL:     "http://failure.com",
	}

	xml := download.CreateXML()
	if strings.Contains(string(xml), "test-id") == false {
		t.Errorf("XML encoding failed: %s", xml)
	}

	doc := xmlx.New()
	if err := doc.LoadBytes(xml, nil); err != nil {
		t.Errorf("XML parse failed: %v", err)
	}

	download.Parse(doc)
	if download.ID != "test-id" {
		t.Errorf("XML parse failed: ID")
	}
	if download.Name != "test-name" {
		t.Errorf("XML parse failed: Name")
	}
	if download.NoMore != 1 {
		t.Errorf("XML parse failed: NoMore")
	}
	// ... 对于其他字段做类似的检查
}
