package cwmp

import (
	"encoding/xml"

	"github.com/talkincode/toughradius/v8/common/xmlx"
)

const (
	// XsdString string type
	XsdString string = "xsd:string"
	// XsdUnsignedint uint type
	XsdUnsignedint string = "xsd:unsignedInt"
)

const (
	// SoapArray array type
	SoapArray string = "soap-enc:Array"
)

const (
	// EventBootStrap first connection
	EventBootStrap string = "0 BOOTSTRAP"
	// EventBoot reset or power on
	EventBoot string = "1 BOOT"
	// EventPeriodic periodic inform
	EventPeriodic string = "2 PERIODIC"
	// EventScheduled scheduled infrorm
	EventScheduled string = "3 SCHEDULED"
	// EventValueChange value change event
	EventValueChange string = "4 VALUE CHANGE"
	// EventKicked acs notify cpe
	EventKicked string = "5 KICKED"
	// EventConnectionRequest cpe request connection
	EventConnectionRequest string = "6 CONNECTION REQUEST"
	// EventTransferComplete download complete
	EventTransferComplete string = "7 TRANSFER COMPLETE"
	// EventClientChange custom event client online/offline
	EventClientChange string = "8 CLIENT CHANGE"
)

// Message tr069 msg interface
type Message interface {
	Parse(doc *xmlx.Document)
	CreateXML() []byte
	GetName() string
	GetID() string
}

// Envelope tr069 body
type Envelope struct {
	XMLName   xml.Name    `xml:"soap-env:Envelope"`
	XmlnsEnv  string      `xml:"xmlns:soap-env,attr"`
	XmlnsEnc  string      `xml:"xmlns:soap-enc,attr"`
	XmlnsXsd  string      `xml:"xmlns:xsd,attr"`
	XmlnsXsi  string      `xml:"xmlns:xsi,attr"`
	XmlnsCwmp string      `xml:"xmlns:cwmp,attr"`
	Header    interface{} `xml:"soap-env:Header"`
	Body      interface{} `xml:"soap-env:Body"`
}

// HeaderStruct tr069 header
type HeaderStruct struct {
	ID     IDStruct    `xml:"cwmp:ID"`
	NoMore interface{} `xml:"cwmp:NoMoreRequests,ommitempty"`
}

// IDStruct msg id
type IDStruct struct {
	Attr  string `xml:"soap-env:mustUnderstand,attr,ommitempty"`
	Value string `xml:",chardata"`
}

// NodeStruct node
type NodeStruct struct {
	Type  interface{} `xml:"xsi:type,attr"`
	Value string      `xml:",chardata"`
}

// EventStruct event
type EventStruct struct {
	Type   string            `xml:"soap-enc:arrayType,attr"`
	Events []EventNodeStruct `xml:"EventStruct"`
}

// EventNodeStruct event node
type EventNodeStruct struct {
	EventCode  NodeStruct `xml:"EventCode"`
	CommandKey string     `xml:"CommandKey"`
}

// ParameterListStruct param list
type ParameterListStruct struct {
	Type   string                 `xml:"soap-enc:arrayType,attr"`
	Params []ParameterValueStruct `xml:"ParameterValueStruct"`
}

// ParameterValueStruct param value
type ParameterValueStruct struct {
	Name  NodeStruct `xml:"Name"`
	Value NodeStruct `xml:"Value"`
}

type ParameterInfoStruct struct {
	Name     string `xml:"Name" json:"name"`
	Writable string `xml:"Writable" json:"writable"`
}

// FaultStruct error
type FaultStruct struct {
	FaultCode   int
	FaultString string
}

// ValueStruct value
type ValueStruct struct {
	Type  string
	Value string
}
