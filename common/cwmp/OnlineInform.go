package cwmp

import (
	"github.com/talkincode/toughradius/common/xmlx"
)

// OnlineInform online client
type OnlineInform struct {
	Sn    string `json:"sn"`
	Hosts []host
}

type host struct {
	Mac      string `json:"mac"`
	HostName string `json:"hostname"`
}

// GetName get msg type
func (msg *OnlineInform) GetName() string {
	return "OnlineInform"
}

// GetID get msg id
func (msg *OnlineInform) GetID() string {
	return "OnlineInform"
}

// CreateXML encode into xml
func (msg *OnlineInform) CreateXML() (xml []byte) {
	return xml
}

// Parse parse from xml
func (msg *OnlineInform) Parse(doc *xmlx.Document) {
	// TODO
}
