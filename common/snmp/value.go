package snmp

import (
	"github.com/gosnmp/gosnmp"
	"net"
	"strconv"
	"strings"
)

type SnmpValue gosnmp.SnmpPDU

func (v SnmpValue) StringValue() string {
	switch v.Type {
	case gosnmp.OctetString:
		b := v.Value.([]byte)
		return string(b)
	case gosnmp.NoSuchObject:
		return ""
	default:
		return gosnmp.ToBigInt(v.Value).String()
	}
}

func (v SnmpValue) Int64Value() int64 {
	switch v.Type {
	case gosnmp.OctetString:
		b := v.Value.([]byte)
		v, _ := strconv.ParseInt(string(b), 10, 64)
		return v
	case gosnmp.NoSuchObject:
		return 0
	default:
		return gosnmp.ToBigInt(v.Value).Int64()
	}
}

func (v SnmpValue) IpAddress() net.IP {
	switch v.Type {
	case gosnmp.IPAddress:
		b := v.Value.([]byte)
		return net.IP(b)
	case gosnmp.OctetString:
		b := v.Value.([]byte)
		return net.IP(b)
	default:
		return net.ParseIP("0.0.0.0")
	}
}

func (v SnmpValue) MacAddr() net.HardwareAddr{
	switch v.Type {
	case gosnmp.IPAddress:
		b := v.Value.([]byte)
		return net.HardwareAddr(b)
	case gosnmp.OctetString:
		b := v.Value.([]byte)
		return net.HardwareAddr(b)
	default:
		_mac, _ := net.ParseMAC("00:00:00:00:00:00")
		return _mac
	}
}

func (v SnmpValue) IsOid(oprefix string) bool{
	if !strings.HasPrefix(oprefix,"."){
		oprefix = "."+oprefix+"."
	}
	return strings.HasPrefix(v.Name, oprefix)
}

func (v SnmpValue) String() string {
	switch v.Type {
	case gosnmp.OctetString:
		b := v.Value.([]byte)
		return string(b)
	case gosnmp.IPAddress:
		b := v.Value.([]byte)
		return net.IP(b).String()
	case gosnmp.NoSuchObject:
		return ""
	default:
		return gosnmp.ToBigInt(v.Value).String()
	}
}
