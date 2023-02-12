package snmp

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/montanaflynn/stats"
	"github.com/talkincode/toughradius/common"
	"github.com/talkincode/toughradius/common/snmp/mibs/hostresmib"

	"github.com/gosnmp/gosnmp"
)

type SnmpAuth struct {
	Ipaddr    string
	Port      int
	Community string
}

type SnmpDevice struct {
	SnmpAuth
	ShopId        int64
	ShopName      string
	VendorName    string
	VendorCode    string
	SoftwareIdOid string
	SoftwareId    string
	SnOid         string
	UseSnOid      bool
	Sn            string
	SystemNameOid string
	SystemName    string
	ModelOid      string
	Model         string
	CpuLoad       float64
	UptimeOid     string
	Uptime        int64
	MemTotal      int64
	MemUsage      int64
	MemPercent    float64
	DiskUsage     int64
	DiskTotal     int64
	DiskPercent   float64
	IfOctetsMap   map[int64]*DeviceIfOctets
}

func (dev *SnmpDevice) ToJson() string {
	if math.IsNaN(dev.CpuLoad) {
		dev.CpuLoad = 0
	}
	if math.IsNaN(dev.MemPercent) {
		dev.MemPercent = 0
	}
	if math.IsNaN(dev.DiskPercent) {
		dev.DiskPercent = 0
	}
	jsons, _ := json.MarshalIndent(dev, "", "\t")
	return string(jsons)
}

func (dev *SnmpDevice) SnmpDocment() (map[string]interface{}, error) {
	vmap := map[string]interface{}{
		"timestamp":    time.Now().Format(time.RFC3339),
		"name":         dev.SystemName,
		"sn":           dev.Sn,
		"model":        dev.Model,
		"shop":         dev.ShopName,
		"cpu_load":     common.ReplaceNaN(dev.CpuLoad, 0),
		"mem_usage":    dev.MemUsage,
		"mem_total":    dev.MemTotal,
		"mem_percent":  common.ReplaceNaN(dev.MemPercent, 0),
		"disk_usage":   dev.DiskUsage,
		"disk_total":   dev.DiskTotal,
		"disk_percent": common.ReplaceNaN(dev.DiskPercent, 0),
		"uptime":       dev.Uptime,
	}
	if dev.Sn == "" && dev.SoftwareId != "" {
		vmap["sn"] = dev.SoftwareId
	}
	return vmap, nil
}

func (dev *SnmpDevice) IfaceDocments() ([]map[string]interface{}, error) {
	ifaces := make([]map[string]interface{}, 0)
	for _, octets := range dev.IfOctetsMap {
		if octets.InOctets == 0 && octets.OutOctets == 0 {
			continue
		}
		vmap := map[string]interface{}{
			"timestamp":  time.Now().Format(time.RFC3339),
			"shop":       dev.ShopName,
			"name":       dev.SystemName,
			"sn":         dev.Sn,
			"iface":      octets.Name,
			"in_octets":  octets.InOctets,
			"out_octets": octets.OutOctets,
		}
		if dev.Sn == "" && dev.SoftwareId != "" {
			vmap["sn"] = dev.SoftwareId
		}
		ifaces = append(ifaces, vmap)
	}
	return ifaces, nil
}

type SnmpV2Client struct {
	Snmpc *gosnmp.GoSNMP
}

func NewSnmpV2Client(target string, port int, community string) (*SnmpV2Client, error) {
	c := &SnmpV2Client{
		Snmpc: &gosnmp.GoSNMP{
			Port:               uint16(port),
			Transport:          "udp",
			Community:          community,
			Version:            gosnmp.Version2c,
			Timeout:            time.Duration(3) * time.Second,
			Retries:            3,
			ExponentialTimeout: true,
			AppOpts:            map[string]interface{}{"c": true},
			MaxOids:            gosnmp.MaxOids,
			Target:             target,
		},
	}
	err := c.Snmpc.Connect()
	if err != nil {
		return nil, err
	}
	return c, nil
}

func NewSnmpV2Client2(auth SnmpAuth) (*SnmpV2Client, error) {
	c := &SnmpV2Client{
		Snmpc: &gosnmp.GoSNMP{
			Port:               uint16(auth.Port),
			Transport:          "udp",
			Community:          auth.Community,
			Version:            gosnmp.Version2c,
			Timeout:            time.Duration(3) * time.Second,
			Retries:            3,
			ExponentialTimeout: true,
			MaxOids:            gosnmp.MaxOids,
			Target:             auth.Ipaddr,
		},
	}
	err := c.Snmpc.Connect()
	if err != nil {
		return nil, err
	}
	return c, nil
}

// Close the connection
func (c *SnmpV2Client) Close() error {
	return c.Snmpc.Conn.Close()
}

func (c *SnmpV2Client) BulkWalkAll(oid string) (map[string]SnmpValue, error) {
	rss, err := c.Snmpc.BulkWalkAll(oid)
	if err != nil {
		return nil, err
	}
	values := make(map[string]SnmpValue, 0)
	for _, pdu := range rss {
		values[pdu.Name] = SnmpValue(pdu)
	}
	return values, err
}

// CollectOids 批量采集OID
func (c *SnmpV2Client) CollectOids(oids []string) (map[string]SnmpValue, error) {
	rs, err := c.Snmpc.Get(oids)
	if err != nil {
		return nil, err
	}
	values := make(map[string]SnmpValue, 0)
	for _, pdu := range rs.Variables {
		for _, oid := range oids {
			if pdu.Name == oid {
				values[oid] = SnmpValue(pdu)
			}
		}
	}
	return values, err
}

func (c *SnmpV2Client) CollectIfOctets(device *SnmpDevice) (err error) {
	device.IfOctetsMap, err = c.queryInterfaces()
	if err != nil {
		return err
	}
	err = c.collectInterfacesInOctets(device.IfOctetsMap)
	if err != nil {
		return err
	}
	err = c.collectInterfacesOutOctets(device.IfOctetsMap)
	if err != nil {
		return err
	}
	return
}

// CollectDeviceSystemInfo 采集设备系统信息
func (c *SnmpV2Client) CollectDeviceSystemInfo(device *SnmpDevice) (err error) {
	oids := []string{
		device.SystemNameOid,
		device.ModelOid,
		device.UptimeOid,
	}
	if device.UseSnOid {
		oids = append(oids, device.SnOid)
	}
	rs, err := c.CollectOids(oids)
	if err != nil {
		return err
	}

	for _, val := range rs {
		switch val.Name {
		case device.SystemNameOid:
			device.SystemName = val.StringValue()
		case device.ModelOid:
			device.Model = val.StringValue()
		case device.UptimeOid:
			device.Uptime = val.Int64Value()
		case device.SnOid:
			device.Sn = val.StringValue()
		}
	}
	return nil
}

// CollectDevice 采集设备信息
func (c *SnmpV2Client) CollectDevice(device *SnmpDevice) error {
	err := c.CollectDeviceSystemInfo(device)
	if err != nil {
		return err
	}

	err = c.collectDeviceCpuUse(device)
	if err != nil {
		return err
	}

	err = c.collectDeviceStorage(device)
	if err != nil {
		return err
	}

	err = c.CollectIfOctets(device)
	if err != nil {
		return err
	}
	return nil
}

// collectDeviceCpuUse 采集CPU使用
func (c *SnmpV2Client) collectDeviceCpuUse(device *SnmpDevice) (err error) {
	switch device.VendorCode {
	default:
		rs, err := c.Snmpc.BulkWalkAll(hostresmib.HOST_RESOURCES_MIB_hrProcessorLoad_OID)
		if err != nil {
			return err
		}
		var totalLoad int64 = 0
		coreCount := int64(len(rs))
		for _, pdu := range rs {
			totalLoad = totalLoad + SnmpValue(pdu).Int64Value()
		}
		device.CpuLoad, _ = stats.Round(float64(totalLoad)/float64(coreCount), 2)
	}
	return nil
}

// collectDeviceStorage 采集磁盘存储
func (c *SnmpV2Client) collectDeviceStorage(device *SnmpDevice) (err error) {
	switch device.VendorCode {
	default:
		var collectStorage = func(oidext string) (used int64, size int64) {
			hrStorageUsed := fmt.Sprintf(".1.3.6.1.2.1.25.2.3.1.6%s", oidext)
			hrStorageSize := fmt.Sprintf(".1.3.6.1.2.1.25.2.3.1.5%s", oidext)
			unit := fmt.Sprintf(".1.3.6.1.2.1.25.2.3.1.4%s", oidext)
			_rs, err := c.CollectOids([]string{
				hrStorageUsed,
				hrStorageSize,
				unit,
			})
			if err != nil {
				log.Printf("collectMemory error %s", err.Error())
				return
			}
			var _unitSize int64 = 1
			for _, pdu := range _rs {
				switch pdu.Name {
				case hrStorageUsed:
					used = pdu.Int64Value()
				case hrStorageSize:
					size = pdu.Int64Value()
				case unit:
					_unitSize = pdu.Int64Value()
				}
			}
			if _unitSize != 0 {
				used = used * _unitSize
				size = size * _unitSize
			}
			return
		}

		var memUsed int64 = 0
		var memTotal int64 = 0
		var diskUsed int64 = 0
		var diskTotal int64 = 0
		rs, err := c.Snmpc.BulkWalkAll(hostresmib.HOST_RESOURCES_MIB_hrStorageDescr_OID)
		if err != nil {
			return err
		}

		for _, _pdu := range rs {
			val := SnmpValue(_pdu)
			switch {
			case strings.Contains(val.StringValue(), "memory"):
				_used, _size := collectStorage(val.Name[strings.LastIndex(val.Name, "."):])
				memUsed += _used
				memTotal += _size
			case strings.Contains(val.StringValue(), "disk"):
				_used2, _size2 := collectStorage(val.Name[strings.LastIndex(val.Name, "."):])
				diskUsed += _used2
				diskTotal += _size2
			}
		}
		device.MemUsage = memUsed
		device.MemTotal = memTotal
		device.MemPercent, _ = stats.Round(float64(memUsed)/float64(memTotal)*100, 2)
		device.DiskUsage = diskUsed
		device.DiskTotal = diskTotal
		device.DiskPercent, _ = stats.Round(float64(diskUsed)/float64(diskTotal)*100, 2)
		return nil
	}
}
