package cwmp

import (
	"encoding/xml"
	"fmt"
	"time"

	"github.com/talkincode/toughradius/v8/common/xmlx"
)

// FactoryReset cpe
type FactoryReset struct {
	ID     string
	Name   string
	NoMore int
}

type factoryResetBodyStruct struct {
	Body factoryResetStruct `xml:"cwmp:FactoryReset"`
}

type factoryResetStruct struct {
}

// GetID get msg id
func (msg *FactoryReset) GetID() string {
	if len(msg.ID) < 1 {
		msg.ID = fmt.Sprintf("ID:intrnl.unset.id.%s%d.%d", msg.GetName(), time.Now().Unix(), time.Now().UnixNano())
	}
	return msg.ID
}

// GetName get msg name
func (msg *FactoryReset) GetName() string {
	return "FactoryReset"
}

// CreateXML encode into xml
func (msg *FactoryReset) CreateXML() []byte {
	env := Envelope{}
	env.XmlnsEnv = "http://schemas.xmlsoap.org/soap/envelope/"
	env.XmlnsEnc = "http://schemas.xmlsoap.org/soap/encoding/"
	env.XmlnsXsd = "http://www.w3.org/2001/XMLSchema"
	env.XmlnsXsi = "http://www.w3.org/2001/XMLSchema-instance"
	env.XmlnsCwmp = "urn:dslforum-org:cwmp-1-0"
	id := IDStruct{Attr: "1", Value: msg.GetID()}
	env.Header = HeaderStruct{ID: id, NoMore: msg.NoMore}
	body := factoryResetStruct{}
	env.Body = factoryResetBodyStruct{body}
	// output, err := xml.Marshal(env)
	output, err := xml.MarshalIndent(env, "  ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	return output
}

// Parse decode from xml
func (msg *FactoryReset) Parse(doc *xmlx.Document) {
	// TODO
}
