package cwmp

import (
	"encoding/xml"
	"fmt"
	"time"

	"github.com/talkincode/toughradius/common/xmlx"
)

// GetParameterNames get rpc methods
type GetParameterNames struct {
	ID            string
	Name          string
	NoMore        int
	ParameterPath string
	NextLevel     string
}

type getParameterNamesBodyStruct struct {
	Body getParameterNamesStruct `xml:"cwmp:GetParameterNames"`
}

type getParameterNamesStruct struct {
	ParameterPath string `xml:"ParameterPath"`
	NextLevel     string `xml:"NextLevel"`
}

// GetID get msg id
func (msg *GetParameterNames) GetID() string {
	if len(msg.ID) < 1 {
		msg.ID = fmt.Sprintf("ID:intrnl.unset.id.%s%d.%d", msg.GetName(), time.Now().Unix(), time.Now().UnixNano())
	}
	return msg.ID
}

// GetName get type name
func (msg *GetParameterNames) GetName() string {
	return "GetParameterNames"
}

// CreateXML encode into xml
func (msg *GetParameterNames) CreateXML() []byte {
	env := Envelope{}
	env.XmlnsEnv = "http://schemas.xmlsoap.org/soap/envelope/"
	env.XmlnsEnc = "http://schemas.xmlsoap.org/soap/encoding/"
	env.XmlnsXsd = "http://www.w3.org/2001/XMLSchema"
	env.XmlnsXsi = "http://www.w3.org/2001/XMLSchema-instance"
	env.XmlnsCwmp = "urn:dslforum-org:cwmp-1-0"
	id := IDStruct{Attr: "1", Value: msg.GetID()}
	env.Header = HeaderStruct{ID: id, NoMore: msg.NoMore}
	respBody := getParameterNamesStruct{
		ParameterPath: msg.ParameterPath,
		NextLevel:     msg.NextLevel,
	}
	env.Body = getParameterNamesBodyStruct{respBody}
	// output, err := xml.Marshal(env)
	output, err := xml.MarshalIndent(env, "  ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	return output
}

// Parse decode from xml
func (msg *GetParameterNames) Parse(doc *xmlx.Document) {
	// TODO
}
