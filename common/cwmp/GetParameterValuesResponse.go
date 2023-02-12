package cwmp

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/talkincode/toughradius/common/xmlx"
)

// GetParameterValuesResponse getParameterValues response
type GetParameterValuesResponse struct {
	ID     string
	Name   string
	Values map[string]string
}

// NewGetParameterValuesResponse create GetParameterValuesResponse object
func NewGetParameterValuesResponse() (m *GetParameterValuesResponse) {
	m = &GetParameterValuesResponse{}
	m.ID = m.GetID()
	m.Name = m.GetName()
	return m
}

type getParameterValuesResponseBodyStruct struct {
	Body getParameterValuesResponseStruct `xml:"cwmp:GetParameterValuesResponse"`
}

type getParameterValuesResponseStruct struct {
	Params ParameterListStruct `xml:"ParameterList"`
}

// GetName get type name
func (msg *GetParameterValuesResponse) GetName() string {
	return "GetParameterValuesResponse"
}

// GetID get msg id
func (msg *GetParameterValuesResponse) GetID() string {
	if len(msg.ID) < 1 {
		msg.ID = fmt.Sprintf("ID:intrnl.unset.id.%s%d.%d", msg.GetName(), time.Now().Unix(), time.Now().UnixNano())
	}
	return msg.ID
}

// CreateXML encode into xml
func (msg *GetParameterValuesResponse) CreateXML() []byte {
	env := Envelope{}
	id := IDStruct{"1", msg.GetID()}
	env.XmlnsEnv = "http://schemas.xmlsoap.org/soap/envelope/"
	env.XmlnsEnc = "http://schemas.xmlsoap.org/soap/encoding/"
	env.XmlnsXsd = "http://www.w3.org/2001/XMLSchema"
	env.XmlnsXsi = "http://www.w3.org/2001/XMLSchema-instance"
	env.XmlnsCwmp = "urn:dslforum-org:cwmp-1-0"
	env.Header = HeaderStruct{ID: id}

	paramLen := strconv.Itoa(len(msg.Values))
	params := ParameterListStruct{Type: "cwmp:ParameterValueStruct[" + paramLen + "]"}
	for k, v := range msg.Values {
		param := ParameterValueStruct{
			Name:  NodeStruct{Type: XsdString, Value: k},
			Value: NodeStruct{Type: XsdString, Value: v}}
		params.Params = append(params.Params, param)
	}
	info := getParameterValuesResponseStruct{Params: params}
	env.Body = getParameterValuesResponseBodyStruct{info}
	output, err := xml.MarshalIndent(env, "  ", "    ")
	// output, err := xml.Marshal(env)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	return output
}

// Parse decode from xml
func (msg *GetParameterValuesResponse) Parse(doc *xmlx.Document) {
	msg.ID = getDocNodeValue(doc, "*", "ID")
	paramsNode := doc.SelectNode("*", "ParameterList")
	if len(strings.TrimSpace(paramsNode.String())) > 0 {
		params := make(map[string]string)
		var name, value string
		for _, param := range paramsNode.Children {
			fmt.Println("param:", param)

			if len(strings.TrimSpace(param.String())) > 0 {
				name = param.SelectNode("", "Name").GetValue()
				value = param.SelectNode("", "Value").GetValue()
				params[name] = value
			}

		}
		msg.Values = params
	}
}
