package cwmp

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"time"

	"github.com/talkincode/toughradius/v8/common/xmlx"
)

// TransferComplete download complete
type TransferComplete struct {
	ID           string
	Name         string
	CommandKey   string
	StartTime    string
	CompleteTime string
	FaultCode    int
	FaultString  string
}

type transferCompleteBodyStruct struct {
	Body transferCompleteStruct `xml:"cwmp:TransferComplete"`
}

type transferCompleteStruct struct {
	CommandKey   string
	StartTime    string
	CompleteTime string
	Fault        interface{} `xml:"FaultStruct,ommitempty"`
}

// GetID get msg id
func (msg *TransferComplete) GetID() string {
	if len(msg.ID) < 1 {
		msg.ID = fmt.Sprintf("ID:intrnl.unset.id.%s%d.%d", msg.GetName(), time.Now().Unix(), time.Now().UnixNano())
	}
	return msg.ID
}

// GetName get msg type
func (msg *TransferComplete) GetName() string {
	return "TransferComplete"
}

// CreateXML encode into mxl
func (msg *TransferComplete) CreateXML() []byte {
	env := Envelope{}
	env.XmlnsEnv = "http://schemas.xmlsoap.org/soap/envelope/"
	env.XmlnsEnc = "http://schemas.xmlsoap.org/soap/encoding/"
	env.XmlnsXsd = "http://www.w3.org/2001/XMLSchema"
	env.XmlnsXsi = "http://www.w3.org/2001/XMLSchema-instance"
	env.XmlnsCwmp = "urn:dslforum-org:cwmp-1-0"
	id := IDStruct{Attr: "1", Value: msg.GetID()}
	env.Header = HeaderStruct{ID: id}
	var body transferCompleteStruct
	if len(msg.FaultString) > 0 {
		fault := FaultStruct{FaultCode: msg.FaultCode, FaultString: msg.FaultString}
		body = transferCompleteStruct{
			CommandKey:   msg.CommandKey,
			StartTime:    msg.StartTime,
			CompleteTime: msg.CompleteTime,
			Fault:        fault,
		}
	} else {
		body = transferCompleteStruct{
			CommandKey:   msg.CommandKey,
			StartTime:    msg.StartTime,
			CompleteTime: msg.CompleteTime,
		}
	}

	env.Body = transferCompleteBodyStruct{body}
	// output, err := xml.Marshal(env)
	output, err := xml.MarshalIndent(env, "  ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	return output
}

// Parse decode from xml
func (msg *TransferComplete) Parse(doc *xmlx.Document) {
	msg.ID = getDocNodeValue(doc, "*", "ID")
	msg.CommandKey = doc.SelectNode("*", "CommandKey").GetValue()
	msg.CompleteTime = doc.SelectNode("*", "CompleteTime").GetValue()
	msg.StartTime = doc.SelectNode("*", "StartTime").GetValue()
	msg.FaultString = doc.SelectNode("*", "FaultString").GetValue()
	faultCode, err := strconv.Atoi(doc.SelectNode("*", "FaultCode").GetValue())
	if err != nil {
		fmt.Printf("falutCode error %v\n", err)
	}
	msg.FaultCode = faultCode

}
