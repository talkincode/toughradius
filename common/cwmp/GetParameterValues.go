package cwmp

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"time"

	"github.com/talkincode/toughradius/v8/common/xmlx"
)

// GetParameterValues get paramvalues
type GetParameterValues struct {
	ID             string
	Name           string
	NoMore         int
	ParameterNames []string
}

type getParameterValuesBodyStruct struct {
	Body getParameterValuesStruct `xml:"cwmp:GetParameterValues"`
}

type getParameterValuesStruct struct {
	Params parameterNamesStruct `xml:"ParameterNames"`
}

type parameterNamesStruct struct {
	Type       string   `xml:"soap-enc:arrayType,attr"`
	ParamNames []string `xml:"string"`
}

// GetName get type name
func (msg *GetParameterValues) GetName() string {
	return "GetParameterValues"
}

// GetID get tr069 msg id
func (msg *GetParameterValues) GetID() string {
	if len(msg.ID) < 1 {
		msg.ID = fmt.Sprintf("ID:intrnl.unset.id.%s%d.%d", msg.GetName(), time.Now().Unix(), time.Now().UnixNano())
	}
	return msg.ID
}

// CreateXML encode into xml
func (msg *GetParameterValues) CreateXML() []byte {
	env := Envelope{}
	id := IDStruct{"1", msg.GetID()}
	env.XmlnsEnv = "http://schemas.xmlsoap.org/soap/envelope/"
	env.XmlnsEnc = "http://schemas.xmlsoap.org/soap/encoding/"
	env.XmlnsXsd = "http://www.w3.org/2001/XMLSchema"
	env.XmlnsXsi = "http://www.w3.org/2001/XMLSchema-instance"
	env.XmlnsCwmp = "urn:dslforum-org:cwmp-1-0"
	env.Header = HeaderStruct{ID: id, NoMore: msg.NoMore}
	paramLen := strconv.Itoa(len(msg.ParameterNames))
	paramNames := parameterNamesStruct{
		Type: XsdString + "[" + paramLen + "]",
	}
	for _, v := range msg.ParameterNames {
		paramNames.ParamNames = append(paramNames.ParamNames, v)
	}
	body := getParameterValuesStruct{paramNames}
	env.Body = getParameterValuesBodyStruct{body}
	output, err := xml.MarshalIndent(env, "  ", "    ")
	// output, err := xml.Marshal(env)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	return output
}

// Parse decode from xml
func (msg *GetParameterValues) Parse(doc *xmlx.Document) {
	// TODO
}
