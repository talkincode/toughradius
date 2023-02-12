package cwmp

import (
	"encoding/xml"
	"fmt"
	"time"

	"github.com/talkincode/toughradius/common/xmlx"
)

// Upload tr069 upload msg
type Upload struct {
	ID           string
	Name         string
	NoMore       int
	CommandKey   string
	FileType     string
	URL          string
	Username     string
	Password     string
	DelaySeconds int
}

type uploadBodyStruct struct {
	Body uploadStruct `xml:"cwmp:Upload"`
}

type uploadStruct struct {
	CommandKey   string
	FileType     string
	URL          string
	Username     string
	Password     string
	DelaySeconds int
}

// GetID get upload msg id(tr069 msg id)
func (msg *Upload) GetID() string {
	if len(msg.ID) < 1 {
		msg.ID = fmt.Sprintf("ID:intrnl.unset.id.%s%d.%d", msg.GetName(), time.Now().Unix(), time.Now().UnixNano())
	}
	return msg.ID
}

// GetName name is msg object type, use for decode
func (msg *Upload) GetName() string {
	return "Upload"
}

// CreateXML encode xml
func (msg *Upload) CreateXML() []byte {
	env := Envelope{}
	env.XmlnsEnv = "http://schemas.xmlsoap.org/soap/envelope/"
	env.XmlnsEnc = "http://schemas.xmlsoap.org/soap/encoding/"
	env.XmlnsXsd = "http://www.w3.org/2001/XMLSchema"
	env.XmlnsXsi = "http://www.w3.org/2001/XMLSchema-instance"
	env.XmlnsCwmp = "urn:dslforum-org:cwmp-1-0"
	id := IDStruct{Attr: "1", Value: msg.GetID()}
	env.Header = HeaderStruct{ID: id, NoMore: msg.NoMore}
	body := uploadStruct{
		CommandKey:   msg.CommandKey,
		FileType:     msg.FileType,
		URL:          msg.URL,
		Username:     msg.Username,
		Password:     msg.Password,
		DelaySeconds: msg.DelaySeconds}
	env.Body = uploadBodyStruct{body}
	// output, err := xml.Marshal(env)
	output, err := xml.MarshalIndent(env, "  ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	return output
}

// Parse parse from xml
func (msg *Upload) Parse(doc *xmlx.Document) {
	// TODO
}
