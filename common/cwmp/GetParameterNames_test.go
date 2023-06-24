package cwmp

import (
	_ "embed"
	"testing"

	"github.com/talkincode/toughradius/v8/common"
)

//go:embed GetParameterNames.xml
var getParameterNamesXml []byte

//go:embed GetParameterNamesResponse.xml
var getParameterNamesResponseXml []byte

func TestGetParameterNames_CreateXML(t *testing.T) {
	msg := GetParameterNames{
		ID:            "",
		Name:          "",
		NoMore:        0,
		ParameterPath: "InternetGatewayDevice.DeviceInfo",
		NextLevel:     "false",
	}
	bytes := msg.CreateXML()
	t.Log(string(bytes))
}

func TestGetParameterNames_Parse(t *testing.T) {
	msg, err := ParseXML(getParameterNamesResponseXml)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(common.ToJson(msg))
}
