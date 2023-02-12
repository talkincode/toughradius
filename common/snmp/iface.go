package snmp

import (
	"strconv"
	"strings"

	"github.com/gosnmp/gosnmp"
	"github.com/talkincode/toughradius/common/snmp/mibs/hostresmib"
)

type DeviceIfOctets struct {
	Name      string `json:"name"`
	Index     int64  `json:"-"`
	Type      int64  `json:"-"`
	InOctets  int64  `json:"in_octets"`
	OutOctets int64  `json:"out_octets"`
}

// SNMP Get interface definition data
func (c *SnmpV2Client) queryInterfaces() (map[int64]*DeviceIfOctets, error) {
	// Get interface index
	indexs, err := c.Snmpc.BulkWalkAll(hostresmib.IF_MIB_ifIndex_OID)
	if err != nil {
		return nil, err
	}
	ifmap := make(map[int64]*DeviceIfOctets)
	for _, idx := range indexs {
		idxval := gosnmp.ToBigInt(idx.Value).Int64()
		ifmap[idxval] = &DeviceIfOctets{
			Index: gosnmp.ToBigInt(idx.Value).Int64(),
		}
	}
	names, err := c.Snmpc.BulkWalkAll(hostresmib.IF_MIB_ifDescr_OID)
	if err != nil {
		return nil, err
	}
	// 获取接口名称
	for _, name := range names {
		nameoid := name.Name
		idxval, err := strconv.ParseInt(nameoid[strings.LastIndex(nameoid, ".")+1:], 10, 64)
		if err != nil {
			continue
		}
		if _, flag := ifmap[idxval]; flag {
			ifmap[idxval].Name = string(name.Value.([]byte))
		}
	}

	types, err := c.Snmpc.BulkWalkAll(hostresmib.IF_MIB_ifType_OID)
	if err != nil {
		return nil, err
	}

	// Get interface name
	for _, _type := range types {
		nameoid := _type.Name
		idxval, err := strconv.ParseInt(nameoid[strings.LastIndex(nameoid, ".")+1:], 10, 64)
		if err != nil {
			continue
		}
		if _, flag := ifmap[idxval]; flag {
			ifmap[idxval].Type = gosnmp.ToBigInt(_type.Value).Int64()
		}
	}

	return ifmap, nil

}

func (c *SnmpV2Client) collectInterfacesInOctets(ifmap map[int64]*DeviceIfOctets) error {
	bytes, err := c.Snmpc.BulkWalkAll(hostresmib.IF_MIB_ifInOctets_OID)
	if err != nil {
		return err
	}
	// Get interface name
	for _, _item := range bytes {
		nameoid := _item.Name
		idxval, err := strconv.ParseInt(nameoid[strings.LastIndex(nameoid, ".")+1:], 10, 64)
		if err != nil {
			continue
		}
		if _, flag := ifmap[idxval]; flag {
			ifmap[idxval].InOctets = gosnmp.ToBigInt(_item.Value).Int64()
		}
	}
	return nil
}

func (c *SnmpV2Client) collectInterfacesOutOctets(ifmap map[int64]*DeviceIfOctets) error {
	bytes, err := c.Snmpc.BulkWalkAll(hostresmib.IF_MIB_ifOutOctets_OID)
	if err != nil {
		return err
	}
	// Get interface name
	for _, _item := range bytes {
		nameoid := _item.Name
		idxval, err := strconv.ParseInt(nameoid[strings.LastIndex(nameoid, ".")+1:], 10, 64)
		if err != nil {
			continue
		}
		if _, flag := ifmap[idxval]; flag {
			ifmap[idxval].OutOctets = gosnmp.ToBigInt(_item.Value).Int64()
		}
	}
	return nil
}
