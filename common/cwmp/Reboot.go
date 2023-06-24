package cwmp

import (
	"encoding/xml"
	"fmt"
	"time"

	"github.com/talkincode/toughradius/v8/common/xmlx"
)

// Reboot reboot cpe
type Reboot struct {
	ID         string
	Name       string
	NoMore     int
	CommandKey string
}

type rebootBodyStruct struct {
	Body rebootStruct `xml:"cwmp:Reboot"`
}

type rebootStruct struct {
	CommandKey string
}

// GetID get msg id
func (msg *Reboot) GetID() string {
	if len(msg.ID) < 1 {
		msg.ID = fmt.Sprintf("ID:intrnl.unset.id.%s%d.%d", msg.GetName(), time.Now().Unix(), time.Now().UnixNano())
	}
	return msg.ID
}

// GetName get msg name
func (msg *Reboot) GetName() string {
	return "Reboot"
}

// CreateXML encode into xml
func (msg *Reboot) CreateXML() []byte {
	env := Envelope{}
	env.XmlnsEnv = "http://schemas.xmlsoap.org/soap/envelope/"
	env.XmlnsEnc = "http://schemas.xmlsoap.org/soap/encoding/"
	env.XmlnsXsd = "http://www.w3.org/2001/XMLSchema"
	env.XmlnsXsi = "http://www.w3.org/2001/XMLSchema-instance"
	env.XmlnsCwmp = "urn:dslforum-org:cwmp-1-0"
	id := IDStruct{Attr: "1", Value: msg.GetID()}
	env.Header = HeaderStruct{ID: id, NoMore: msg.NoMore}
	body := rebootStruct{CommandKey: msg.CommandKey}
	env.Body = rebootBodyStruct{body}
	// output, err := xml.Marshal(env)
	output, err := xml.MarshalIndent(env, "  ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	return output
}

// Parse decode from xml
func (msg *Reboot) Parse(doc *xmlx.Document) {
	// TODO
}
