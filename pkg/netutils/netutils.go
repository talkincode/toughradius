package netutils

import (
	"fmt"
	"strings"

	iplib "github.com/c-robinson/iplib"
)

func ParseIpNet(d string) (inet iplib.Net, err error) {
	if !strings.Contains(d, "/") {
		d = d + "/32"
	}
	_net := iplib.Net4FromStr(d)
	if _net.IP() == nil {
		_net6 := iplib.Net6FromStr(d)
		if _net6.IP() == nil {
			return nil, fmt.Errorf("error ip  %s", d)
		}
		return _net6, nil
	}
	return _net, nil
}

func ContainsNetAddr(ns []iplib.Net, ipstr string) bool {
	inet, err := ParseIpNet(ipstr)
	if err != nil {
		return false
	}
	for _, n := range ns {
		if n.ContainsNet(inet) {
			return true
		}
	}
	return false
}
