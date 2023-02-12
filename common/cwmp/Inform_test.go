package cwmp

import (
	_ "embed"
	"testing"
)

//go:embed inform_test.xml
var informTest []byte

func TestInform_Parse(t *testing.T) {
	msg, err := ParseXML(informTest)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(msg)
}
