package cwmp

import (
	"encoding/xml"
	"fmt"
	"time"

	"github.com/talkincode/toughradius/common/xmlx"
)

// InformResponse infrom response
type InformResponse struct {
	ID           string
	Name         string
	NoMore       int
	MaxEnvelopes int
}

type informResponseBodyStruct struct {
	Body informResponseStruct `xml:"cwmp:InformResponse"`
}

type informResponseStruct struct {
	MaxEnvelopes int `xml:"MaxEnvelopes"`
}

// GetID get msg id
func (msg *InformResponse) GetID() string {
	if len(msg.ID) < 1 {
		msg.ID = fmt.Sprintf("ID:intrnl.unset.id.%s%d.%d", msg.GetName(), time.Now().Unix(), time.Now().UnixNano())
	}
	return msg.ID
}

// GetName get msg type
func (msg *InformResponse) GetName() string {
	return "InformResponse"
}

// CreateXML encode into xml
func (msg *InformResponse) CreateXML() []byte {
	env := Envelope{}
	env.XmlnsEnv = "http://schemas.xmlsoap.org/soap/envelope/"
	env.XmlnsEnc = "http://schemas.xmlsoap.org/soap/encoding/"
	env.XmlnsXsd = "http://www.w3.org/2001/XMLSchema"
	env.XmlnsXsi = "http://www.w3.org/2001/XMLSchema-instance"
	env.XmlnsCwmp = "urn:dslforum-org:cwmp-1-0"
	id := IDStruct{Attr: "1", Value: msg.GetID()}
	env.Header = HeaderStruct{ID: id, NoMore: msg.NoMore}
	respBody := informResponseStruct{MaxEnvelopes: msg.MaxEnvelopes}
	env.Body = informResponseBodyStruct{respBody}
	// output, err := xml.Marshal(env)
	output, err := xml.MarshalIndent(env, "  ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	return output
}

// Parse decode from xml
func (msg *InformResponse) Parse(doc *xmlx.Document) {
	// TODO
}
