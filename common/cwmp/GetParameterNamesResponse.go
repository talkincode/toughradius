package cwmp

import (
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"github.com/talkincode/toughradius/v8/common/xmlx"
)

// GetParameterNamesResponse GetParameterNames reponse
type GetParameterNamesResponse struct {
	ID     string
	Name   string
	Params []ParameterInfoStruct
}

type GetParameterNamesResponseBodyStruct struct {
	Body GetParameterNamesResponseStruct `xml:"cwmp:GetParameterNamesResponse"`
}

type GetParameterNamesResponseStruct struct {
	ParameterList ParameterListStruct `xml:"cwmp:ParameterList"`
}

// GetName get msg type
func (msg *GetParameterNamesResponse) GetName() string {
	return "GetParameterNamesResponse"
}

// GetID get msg id
func (msg *GetParameterNamesResponse) GetID() string {
	if len(msg.ID) < 1 {
		msg.ID = fmt.Sprintf("ID:intrnl.unset.id.%s%d.%d", msg.GetName(), time.Now().Unix(), time.Now().UnixNano())
	}
	return msg.ID
}

// CreateXML encode into xml
func (msg *GetParameterNamesResponse) CreateXML() []byte {
	env := Envelope{}
	id := IDStruct{"1", msg.GetID()}
	env.XmlnsEnv = "http://schemas.xmlsoap.org/soap/envelope/"
	env.XmlnsEnc = "http://schemas.xmlsoap.org/soap/encoding/"
	env.XmlnsXsd = "http://www.w3.org/2001/XMLSchema"
	env.XmlnsXsi = "http://www.w3.org/2001/XMLSchema-instance"
	env.XmlnsCwmp = "urn:dslforum-org:cwmp-1-0"
	env.Header = HeaderStruct{ID: id}
	body := GetParameterNamesResponseStruct{}
	env.Body = GetParameterNamesResponseBodyStruct{body}
	output, err := xml.MarshalIndent(env, "  ", "    ")
	// output, err := xml.Marshal(env)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	return output
}

// Parse decode from xml
func (msg *GetParameterNamesResponse) Parse(doc *xmlx.Document) {
	msg.ID = getDocNodeValue(doc, "*", "ID")
	paramList := doc.SelectNode("*", "ParameterList")
	if len(strings.TrimSpace(paramList.String())) > 0 {
		var params = make([]ParameterInfoStruct, 0)
		for _, param := range paramList.Children {
			if len(strings.TrimSpace(param.String())) > 0 {
				params = append(params, ParameterInfoStruct{
					Name:     param.SelectNode("", "Name").GetValue(),
					Writable: param.SelectNode("", "Writable").GetValue(),
				})
			}
		}
		msg.Params = params
	}
}
