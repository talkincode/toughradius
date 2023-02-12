package cwmp

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"time"

	"github.com/talkincode/toughradius/common/xmlx"
)

// UploadResponse upload response
type UploadResponse struct {
	ID           string
	Name         string
	Status       int
	StartTime    string
	CompleteTime string
}

type uploadResponseBodyStruct struct {
	UploadResp uploadResponseStruct `xml:"cwmp:UploadResponse"`
}

type uploadResponseStruct struct {
	Status       int
	StartTime    string
	CompleteTime string
}

// GetID tr069 msg id
func (msg *UploadResponse) GetID() string {
	if len(msg.ID) < 1 {
		msg.ID = fmt.Sprintf("ID:intrnl.unset.id.%s%d.%d", msg.GetName(), time.Now().Unix(), time.Now().UnixNano())
	}
	return msg.ID
}

// GetName msg type name
func (msg *UploadResponse) GetName() string {
	return "UploadResponse"
}

// CreateXML encode into xml
func (msg *UploadResponse) CreateXML() []byte {
	env := Envelope{}
	env.XmlnsEnv = "http://schemas.xmlsoap.org/soap/envelope/"
	env.XmlnsEnc = "http://schemas.xmlsoap.org/soap/encoding/"
	env.XmlnsXsd = "http://www.w3.org/2001/XMLSchema"
	env.XmlnsXsi = "http://www.w3.org/2001/XMLSchema-instance"
	env.XmlnsCwmp = "urn:dslforum-org:cwmp-1-0"
	id := IDStruct{Attr: "1", Value: msg.GetID()}
	env.Header = HeaderStruct{ID: id}
	body := uploadResponseStruct{
		StartTime:    msg.StartTime,
		CompleteTime: msg.CompleteTime,
		Status:       msg.Status,
	}
	env.Body = uploadResponseBodyStruct{body}
	// output, err := xml.Marshal(env)
	output, err := xml.MarshalIndent(env, "  ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	return output
}

// Parse decode into struct
func (msg *UploadResponse) Parse(doc *xmlx.Document) {
	msg.ID = getDocNodeValue(doc, "*", "ID")
	statusNode := doc.SelectNode("*", "Status")
	if statusNode != nil {
		var err error
		msg.Status, err = strconv.Atoi(statusNode.GetValue())
		if err != nil {
			fmt.Printf("error: %v\n", err)
		}
	}

	msg.StartTime = doc.SelectNode("*", "StartTime").GetValue()
	msg.CompleteTime = doc.SelectNode("*", "CompleteTime").GetValue()
}
