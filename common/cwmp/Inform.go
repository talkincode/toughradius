package cwmp

import (
	"encoding/xml"
	"fmt"
	// "github.com/coraldane/godom"
	"strconv"
	"strings"
	"time"

	"github.com/talkincode/toughradius/v8/common/xmlx"
)

// Inform tr069 inform (heartbeat)
type Inform struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Manufacturer string            `json:"manufacturer"`
	OUI          string            `json:"oui"`
	ProductClass string            `json:"productClass"`
	Sn           string            `json:"sn"`
	Events       map[string]string `json:"events"`
	MaxEnvelopes int               `json:"maxEnvelopes"`
	CurrentTime  string            `json:"currentTime"`
	RetryCount   int               `json:"retryCount"`
	CommandKey   string            `json:"commandKey"`
	Params       map[string]string `json:"params"`
}

type informBodyStruct struct {
	Body informStruct `xml:"cwmp:Inform"`
}

type informStruct struct {
	DeviceID     deviceIDStruct      `xml:"DeviceId"`
	Event        EventStruct         `xml:"Event"`
	MaxEnvelopes NodeStruct          `xml:"MaxEnvelopes"`
	CurrentTime  NodeStruct          `xml:"CurrentTime"`
	RetryCount   NodeStruct          `xml:"RetryCount"`
	CommandKey   NodeStruct          `xml:"CommandKey"`
	Params       ParameterListStruct `xml:"ParameterList"`
}

type deviceIDStruct struct {
	Type         string     `xml:"xsi:type,attr"`
	Manufacturer NodeStruct `xml:"Manufacturer"`
	OUI          NodeStruct `xml:"OUI"`
	ProductClass NodeStruct `xml:"ProductClass"`
	SerialNumber NodeStruct `xml:"SerialNumber"`
}

func NewInform() *Inform {
	inform := new(Inform)
	inform.Events = make(map[string]string)
	inform.Params = make(map[string]string)
	return inform
}

// GetName get msg type
func (msg *Inform) GetName() string {
	return "Inform"
}

// GetID get msg id
func (msg *Inform) GetID() string {
	if len(msg.ID) < 1 {
		msg.ID = fmt.Sprintf("ID:intrnl.unset.id.%s%d.%d", msg.GetName(), time.Now().Unix(), time.Now().UnixNano())
	}
	return msg.ID
}

// CreateXML encode into xml
func (msg *Inform) CreateXML() []byte {
	env := Envelope{}
	id := IDStruct{"1", msg.GetID()}
	env.XmlnsEnv = "http://schemas.xmlsoap.org/soap/envelope/"
	env.XmlnsEnc = "http://schemas.xmlsoap.org/soap/encoding/"
	env.XmlnsXsd = "http://www.w3.org/2001/XMLSchema"
	env.XmlnsXsi = "http://www.w3.org/2001/XMLSchema-instance"
	env.XmlnsCwmp = "urn:dslforum-org:cwmp-1-0"
	env.Header = HeaderStruct{ID: id}
	manufacturer := NodeStruct{Type: XsdString, Value: msg.Manufacturer}
	oui := NodeStruct{Type: XsdString, Value: msg.OUI}
	productClass := NodeStruct{Type: XsdString, Value: msg.ProductClass}
	serialNumber := NodeStruct{Type: XsdString, Value: msg.Sn}
	deviceID := deviceIDStruct{Type: "cwmp:DeviceIdStruct", Manufacturer: manufacturer, OUI: oui, ProductClass: productClass, SerialNumber: serialNumber}
	eventLen := strconv.Itoa(len(msg.Events))
	event := EventStruct{Type: "cwmp:EventStruct[" + eventLen + "]"}
	for k, v := range msg.Events {
		eventCode := NodeStruct{Type: XsdString, Value: k}
		event.Events = append(event.Events, EventNodeStruct{EventCode: eventCode, CommandKey: v})
	}

	maxEnv := strconv.Itoa(msg.MaxEnvelopes)
	maxEnvelopes := NodeStruct{Type: XsdString, Value: maxEnv}
	currentTime := NodeStruct{Type: XsdString, Value: msg.CurrentTime}
	trys := strconv.Itoa(msg.RetryCount)
	retryCount := NodeStruct{Type: XsdString, Value: trys}
	commandKey := NodeStruct{Type: XsdString, Value: msg.CommandKey}
	paramLen := strconv.Itoa(len(msg.Params))
	paramList := ParameterListStruct{Type: "cwmp:ParameterValueStruct[" + paramLen + "]"}
	for k, v := range msg.Params {
		param := ParameterValueStruct{
			Name:  NodeStruct{Type: XsdString, Value: k},
			Value: NodeStruct{Type: XsdString, Value: v}}
		paramList.Params = append(paramList.Params, param)
	}
	info := informStruct{DeviceID: deviceID, Event: event, MaxEnvelopes: maxEnvelopes,
		CurrentTime: currentTime, RetryCount: retryCount, CommandKey: commandKey, Params: paramList}
	env.Body = informBodyStruct{info}
	output, err := xml.MarshalIndent(env, "  ", "    ")
	// output, err := xml.Marshal(env)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	return output
}

// Parse decode from xml
func (msg *Inform) Parse(doc *xmlx.Document) {
	msg.ID = getDocNodeValue(doc, "*", "ID")
	deviceNode := doc.SelectNode("*", "DeviceId")
	if deviceNode != nil && len(strings.TrimSpace(deviceNode.String())) > 0 {
		msg.Manufacturer = getNodeValue(deviceNode, "", "Manufacturer")
		msg.OUI = getNodeValue(deviceNode, "", "OUI")
		msg.ProductClass = getNodeValue(deviceNode, "", "ProductClass")
		msg.Sn = getNodeValue(deviceNode, "", "SerialNumber")
	}
	informNode := doc.SelectNode("*", "Inform")
	if informNode != nil && len(strings.TrimSpace(informNode.String())) > 0 {
		var err error
		msg.CommandKey = getNodeValue(informNode, "", "CommandKey")
		msg.CurrentTime = getNodeValue(informNode, "", "CurrentTime")
		msg.MaxEnvelopes, err = strconv.Atoi(getNodeValue(informNode, "", "MaxEnvelopes"))
		if err != nil {
			fmt.Printf("error: %v\n", err)
		}
		msg.RetryCount, err = strconv.Atoi(getNodeValue(informNode, "", "RetryCount"))
		if err != nil {
			fmt.Printf("error: %v\n", err)
		}
	}

	eventNode := doc.SelectNode("*", "Event")
	if eventNode != nil && len(strings.TrimSpace(eventNode.String())) > 0 {
		// msg.Events = make(map[string]string)
		var code string
		for _, event := range eventNode.Children {
			if event != nil && len(strings.TrimSpace(event.String())) > 0 {
				code = getNodeValue(event, "", "EventCode")
				msg.Events[code] = getNodeValue(event, "", "CommandKey")
			}
		}
	}

	paramsNode := doc.SelectNode("*", "ParameterList")
	if paramsNode != nil && len(strings.TrimSpace(paramsNode.String())) > 0 {
		// msg.Params = make(map[string]string)
		var name string
		for _, param := range paramsNode.Children {
			if param != nil && len(strings.TrimSpace(param.String())) > 0 {
				name = getNodeValue(param, "", "Name")
				msg.Params[name] = getNodeValue(param, "", "Value")
			}
		}
	}
}

// IsEvent is a connect request or others
func (msg *Inform) IsEvent(event string) bool {
	/*
		for k,_:= range msg.Events {
			if k == event {
				return true
			}
		}
	*/
	if _, ok := msg.Events[event]; ok {
		return true
	}
	return false
}

// GetParam get param in inform
func (msg *Inform) GetParam(name string) (value string) {
	/*
		for k, v := range msg.Params {
			if k == name {
				value = v
				break
			}
		}
	*/
	value = msg.Params[name]
	return
}

// GetConfigVersion get current config version
func (msg *Inform) GetConfigVersion() (version string) {
	version = msg.GetParam("InternetGatewayDevice.DeviceConfig.ConfigVersion")
	return
}

func (msg *Inform) GetSoftwareVersion() (version string) {
	version = msg.GetParam("Device.DeviceInfo.SoftwareVersion")
	return
}
