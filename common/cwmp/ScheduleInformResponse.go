package cwmp

import (
	"encoding/xml"
	"fmt"
	"time"

	"github.com/talkincode/toughradius/v8/common/xmlx"
)

// ScheduleInformResponse resp
type ScheduleInformResponse struct {
	ID   string
	Name string
}

type scheduleInformResponseBodyStruct struct {
	Body scheduleInformResponseStruct `xml:"cwmp:ScheduleInformResponse"`
}

type scheduleInformResponseStruct struct {
}

// GetID get msg id
func (msg *ScheduleInformResponse) GetID() string {
	if len(msg.ID) < 1 {
		msg.ID = fmt.Sprintf("ID:intrnl.unset.id.%s%d.%d", msg.GetName(), time.Now().Unix(), time.Now().UnixNano())
	}
	return msg.ID
}

// GetName get msg type
func (msg *ScheduleInformResponse) GetName() string {
	return "ScheduleInformResponse"
}

// CreateXML encode into xml
func (msg *ScheduleInformResponse) CreateXML() []byte {
	env := Envelope{}
	env.XmlnsEnv = "http://schemas.xmlsoap.org/soap/envelope/"
	env.XmlnsEnc = "http://schemas.xmlsoap.org/soap/encoding/"
	env.XmlnsXsd = "http://www.w3.org/2001/XMLSchema"
	env.XmlnsXsi = "http://www.w3.org/2001/XMLSchema-instance"
	env.XmlnsCwmp = "urn:dslforum-org:cwmp-1-0"
	id := IDStruct{Attr: "1", Value: msg.GetID()}
	env.Header = HeaderStruct{ID: id}
	body := scheduleInformResponseStruct{}
	env.Body = scheduleInformResponseBodyStruct{body}
	// output, err := xml.Marshal(env)
	output, err := xml.MarshalIndent(env, "  ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	return output
}

// Parse decode from xml
func (msg *ScheduleInformResponse) Parse(doc *xmlx.Document) {
	msg.ID = getDocNodeValue(doc, "*", "ID")
}
