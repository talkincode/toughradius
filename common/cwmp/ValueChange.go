package cwmp

import (
	"github.com/talkincode/toughradius/v8/common/xmlx"
)

// ValueChange value change
type ValueChange struct {
	Sn    string `json:"sn"`
	Names []string
}

// GetName get msg type
func (msg *ValueChange) GetName() string {
	return "ValueChange"
}

// GetID get msg id
func (msg *ValueChange) GetID() string {
	return "ValueChange"
}

// CreateXML encode into xml
func (msg *ValueChange) CreateXML() (xml []byte) {
	return xml
}

// Parse decode from xml
func (msg *ValueChange) Parse(doc *xmlx.Document) {
	// TODO
}
