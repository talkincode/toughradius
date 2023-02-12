package cwmp

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"time"

	"github.com/talkincode/toughradius/common/xmlx"
)

// SetParameterValues set param
type SetParameterValues struct {
	ID           string
	Name         string
	NoMore       int
	Params       map[string]ValueStruct
	ParameterKey string
}

type setParameterValuesBodyStruct struct {
	Body setParameterValuesStruct `xml:"cwmp:SetParameterValues"`
}

type setParameterValuesStruct struct {
	ParamList    ParameterListStruct `xml:"ParameterList"`
	ParameterKey string
}

// GetName get msg type
func (msg *SetParameterValues) GetName() string {
	return "SetParameterValues"
}

// GetID get msg id
func (msg *SetParameterValues) GetID() string {
	if len(msg.ID) < 1 {
		msg.ID = fmt.Sprintf("ID:intrnl.unset.id.%s%d.%d", msg.GetName(), time.Now().Unix(), time.Now().UnixNano())
	}
	return msg.ID
}

// CreateXML encode into xml
func (msg *SetParameterValues) CreateXML() []byte {
	env := Envelope{}
	id := IDStruct{"1", msg.GetID()}
	env.XmlnsEnv = "http://schemas.xmlsoap.org/soap/envelope/"
	env.XmlnsEnc = "http://schemas.xmlsoap.org/soap/encoding/"
	env.XmlnsXsd = "http://www.w3.org/2001/XMLSchema"
	env.XmlnsXsi = "http://www.w3.org/2001/XMLSchema-instance"
	env.XmlnsCwmp = "urn:dslforum-org:cwmp-1-0"
	env.Header = HeaderStruct{ID: id, NoMore: msg.NoMore}

	paramLen := strconv.Itoa(len(msg.Params))
	paramList := ParameterListStruct{Type: "cwmp:ParameterValueStruct[" + paramLen + "]"}
	for k, v := range msg.Params {
		param := ParameterValueStruct{
			Name:  NodeStruct{Value: k},
			Value: NodeStruct{Type: v.Type, Value: v.Value}}
		paramList.Params = append(paramList.Params, param)
	}
	body := setParameterValuesStruct{
		ParamList:    paramList,
		ParameterKey: msg.ParameterKey,
	}
	env.Body = setParameterValuesBodyStruct{body}
	output, err := xml.MarshalIndent(env, "  ", "    ")
	// output, err := xml.Marshal(env)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	return output
}

// Parse decode from xml
func (msg *SetParameterValues) Parse(doc *xmlx.Document) {
	// TODO
}
