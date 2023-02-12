package snmp

const (
	VendorMikrotik = "14988"
	VendorH3C      = "25506"
	VendorHuawei   = "2011"
	VendorRuijie   = "4881"
)

func NewMikrotikDevice() *SnmpDevice {
	return &SnmpDevice{
		VendorName:         "Mikrotik",
		VendorCode:         VendorMikrotik,
		SoftwareIdOid:      ".1.3.6.1.4.1.14988.1.1.4.1.0",
		SnOid:              ".1.3.6.1.4.1.14988.1.1.7.3.0",
		SystemNameOid:      ".1.3.6.1.2.1.1.5.0",
		ModelOid:           ".1.3.6.1.2.1.1.1.0",
		UptimeOid:          ".1.3.6.1.2.1.1.3.0",
		IfOctetsMap:        nil,
	}
}

func NewH3CDevice() *SnmpDevice {
	return &SnmpDevice{
		VendorName:         "H3C",
		VendorCode:         VendorH3C,
		SoftwareIdOid:      ".1.3.6.1.4.1.25506.2.6.1.2.1.1.2.1",
		SnOid:              ".1.3.6.1.4.1.25506.2.6.1.2.1.1.2.1",
		SystemNameOid:      ".1.3.6.1.2.1.1.5.0",
		ModelOid:           ".1.3.6.1.2.1.1.1.0",
		UptimeOid:          ".1.3.6.1.2.1.1.3.0",
		IfOctetsMap:        nil,
	}
}

func NewHuaweiDevice() *SnmpDevice {
	return &SnmpDevice{
		VendorName:         "Huawei",
		VendorCode:         VendorHuawei,
		SoftwareIdOid:      ".1.3.6.1.4.1.2011.10.2.6.1.2.1.1.2",
		SnOid:              ".1.3.6.1.4.1.2011.10.2.6.1.2.1.1.2",
		SystemNameOid:      ".1.3.6.1.2.1.1.5.0",
		ModelOid:           ".1.3.6.1.2.1.1.1.0",
		UptimeOid:          ".1.3.6.1.2.1.1.3.0",
		IfOctetsMap:        nil,
	}
}

func NewRuijieDevice() *SnmpDevice {
	return &SnmpDevice{
		VendorName:         "Ruijie",
		VendorCode:         VendorRuijie,
		SoftwareIdOid:      ".1.3.6.1.4.1.4881.1.1.10.2.56.1.1.20.0",
		SnOid:              ".1.3.6.1.4.1.4881.1.1.10.2.56.1.1.20.0",
		SystemNameOid:      ".1.3.6.1.2.1.1.5.0",
		ModelOid:           ".1.3.6.1.2.1.1.1.0",
		UptimeOid:          ".1.3.6.1.2.1.1.3.0",
		IfOctetsMap:        nil,
	}
}
