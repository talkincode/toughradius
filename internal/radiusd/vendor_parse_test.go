package radiusd

import (
	"fmt"
	"testing"
)

func TestParseVlansOfStd(t *testing.T) {
	s := "3/0/1:2814.727"
	fmt.Println(ParseVlanIds(s))
	s2 := "3/0/1:2814"
	fmt.Println(ParseVlanIds(s2))

	s3 := "slot=2;subslot=2;port=22;vlanid=503;"
	fmt.Println(ParseVlanIds(s3))
}
