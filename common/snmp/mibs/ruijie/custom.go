package ruijie

const (

	// ApItemCountOid AC 当前 AP 数量
	/**
	snmpwalk.exe -v 2c -c wehere 10.221.0.186:16101 .1.3.6.1.2.1.145.1.1.1.0
	SNMPv2-SMI::mib-2.145.1.1.1.0 = Gauge32: 3
	 */
	ApItemCountOid  = "1.3.6.1.2.1.145.1.1.1.0"

	// ApItemMaxCountOid AC 支持 AP 最大数量
	/**
	snmpwalk.exe -v 2c -c wehere 10.221.0.186:16101 1.3.6.1.2.1.145.1.1.2.0
	SNMPv2-SMI::mib-2.145.1.1.2.0 = Gauge32: 48
	 */
	ApItemMaxCountOid  = "1.3.6.1.2.1.145.1.1.1.0"


	// ApItemIpListOidPrefix AP IP 列表 OID
	/**
	snmpwalk.exe -v 2c -c wehere 10.221.0.186:16101 .1.3.6.1.2.1.145.1.2.2.1.3
	SNMPv2-SMI::mib-2.145.1.2.2.1.3.6.88.105.108.103.45.199 = Hex-STRING: AC 10 00 E2
	SNMPv2-SMI::mib-2.145.1.2.2.1.3.6.88.105.108.153.254.73 = Hex-STRING: AC 10 03 38
	SNMPv2-SMI::mib-2.145.1.2.2.1.3.6.88.105.108.206.10.46 = Hex-STRING: AC 10 03 36

	88.105.108.103.45.199 为ap设备编号
	 */
	ApItemIpListOidPrefix = "1.3.6.1.2.1.145.1.2.2.1.3.6"

	ApItemOidPrefix = "1.3.6.1.4.1.4881.1.1.10.2.1.1.39.1"


	// ApItemMacOidPrefix AP MAC 列表 OID
	/**
	snmpwalk.exe -v 2c -c wehere 10.221.0.186:16101 .1.3.6.1.4.1.4881.1.1.10.2.1.1.39.1.1
	SNMPv2-SMI::enterprises.4881.1.1.10.2.1.1.39.1.1.88.105.108.103.45.199 = Hex-STRING: 58 69 6C 67 2D C7
	SNMPv2-SMI::enterprises.4881.1.1.10.2.1.1.39.1.1.88.105.108.153.254.73 = Hex-STRING: 58 69 6C 99 FE 49
	SNMPv2-SMI::enterprises.4881.1.1.10.2.1.1.39.1.1.88.105.108.206.10.46 = Hex-STRING: 58 69 6C CE 0A 2E
	 */
	ApItemMacOidPrefix = "1.3.6.1.4.1.4881.1.1.10.2.1.1.39.1.1"

	// ApItemVersionOidPrefix AP Version 列表 OID
	/**
	snmpwalk.exe -v 2c -c wehere 10.221.0.186:16101 .1.3.6.1.4.1.4881.1.1.10.2.1.1.39.1.6
	SNMPv2-SMI::enterprises.4881.1.1.10.2.1.1.39.1.6.88.105.108.103.45.199 = STRING: "AP_RGOS 11.1(5)B8, Release(03142217)"
	SNMPv2-SMI::enterprises.4881.1.1.10.2.1.1.39.1.6.88.105.108.153.254.73 = STRING: "AP_RGOS 11.1(5)B83P3, Release(03241705)"
	SNMPv2-SMI::enterprises.4881.1.1.10.2.1.1.39.1.6.88.105.108.206.10.46 = STRING: "AP_RGOS 11.1(5)B6, Release(0219"
	 */
	ApItemVersionOidPrefix = "1.3.6.1.4.1.4881.1.1.10.2.1.1.39.1.6"

	// ApItemSnOidPrefix AP Sn 列表 OID
	/**
	snmpwalk.exe -v 2c -c wehere 10.221.0.186:16101 .1.3.6.1.4.1.4881.1.1.10.2.1.1.39.1.13
	SNMPv2-SMI::enterprises.4881.1.1.10.2.1.1.39.1.13.88.105.108.103.45.199 = STRING: "G1JDCEZ062230"
	SNMPv2-SMI::enterprises.4881.1.1.10.2.1.1.39.1.13.88.105.108.153.254.73 = STRING: "G1KD9U1723876"
	SNMPv2-SMI::enterprises.4881.1.1.10.2.1.1.39.1.13.88.105.108.206.10.46 = STRING: "G1KDB43018711"
	 */
	ApItemSnOidPrefix = "1.3.6.1.4.1.4881.1.1.10.2.1.1.39.1.13"

	// ApItemModelOidPrefix AP Name 列表 OID
	/**
	snmpwalk.exe -v 2c -c wehere 10.221.0.186:16101 .1.3.6.1.4.1.4881.1.1.10.2.1.1.39.1.14
	SNMPv2-SMI::enterprises.4881.1.1.10.2.1.1.39.1.14.88.105.108.103.45.199 = STRING: "AP130(L)"
	SNMPv2-SMI::enterprises.4881.1.1.10.2.1.1.39.1.14.88.105.108.153.254.73 = STRING: "AP120-W"
	SNMPv2-SMI::enterprises.4881.1.1.10.2.1.1.39.1.14.88.105.108.206.10.46 = STRING: "AP120-S"
	 */
	ApItemModelOidPrefix = "1.3.6.1.4.1.4881.1.1.10.2.1.1.39.1.14"


	// ApItemUptimeOidPrefix AP Uptime 列表 OID
	/**
	snmpwalk.exe -v 2c -c wehere 10.221.0.186:16101 .1.3.6.1.4.1.4881.1.1.10.2.1.1.39.1.16
	SNMPv2-SMI::enterprises.4881.1.1.10.2.1.1.39.1.16.88.105.108.103.45.199 = Timeticks: (87367500) 10 days, 2:41:15.00
	SNMPv2-SMI::enterprises.4881.1.1.10.2.1.1.39.1.16.88.105.108.153.254.73 = Timeticks: (87359500) 10 days, 2:39:55.00
	SNMPv2-SMI::enterprises.4881.1.1.10.2.1.1.39.1.16.88.105.108.206.10.46 = Timeticks: (87362900) 10 days, 2:40:29.00
	 */
	ApItemUptimeOidPrefix = "1.3.6.1.4.1.4881.1.1.10.2.1.1.39.1.16"


	// PoeMacListOidPrefix
	/**
	snmpwalk -Cc -v 2c  -c  wehere 10.203.0.10:16120 1.3.6.1.2.1.17.4.3.1.1
	SNMPv2-SMI::mib-2.17.4.3.1.1.40.87.190.190.77.57 = Hex-STRING: 28 57 BE BE 4D 39
	SNMPv2-SMI::mib-2.17.4.3.1.1.220.144.136.74.5.3 = Hex-STRING: DC 90 88 4A 05 03
	SNMPv2-SMI::mib-2.17.4.3.1.1.82.80.68.53.22.206 = Hex-STRING: 52 50 44 35 16 CE
	SNMPv2-SMI::mib-2.17.4.3.1.1.108.215.31.134.34.135 = Hex-STRING: 6C D7 1F 86 22 87
	 */
	PoeMacListOidPrefix = "1.3.6.1.2.1.17.4.3.1.1"

	// PoeBridgeListOidPrefix
	/**
	snmpwalk -Cc -v 2c  -c  wehere 10.203.0.10:16120 1.3.6.1.2.1.17.4.3.1.2
	SNMPv2-SMI::mib-2.17.4.3.1.2.40.87.190.190.77.57 = INTEGER: 24
	SNMPv2-SMI::mib-2.17.4.3.1.2.220.144.136.74.5.3 = INTEGER: 2
	SNMPv2-SMI::mib-2.17.4.3.1.2.82.80.68.53.22.206 = INTEGER: 9
	SNMPv2-SMI::mib-2.17.4.3.1.2.108.215.31.134.34.135 = INTEGER: 18
	SNMPv2-SMI::mib-2.17.4.3.1.2.88.105.108.102.69.36 = INTEGER: 5
	SNMPv2-SMI::mib-2.17.4.3.1.2.24.129.14.69.80.34 = INTEGER: 4
	 */
	PoeBridgeListOidPrefix = "1.3.6.1.2.1.17.4.3.1.2"

	// PoeBridgeInfaceOidPrefix
	/**
	snmpwalk -Cc -v 2c  -c  wehere 10.203.0.10:16120 1.3.6.1.2.1.17.1.4.1.2.14
	SNMPv2-SMI::mib-2.17.1.4.1.2.14 = INTEGER: 14
	 */
	PoeBridgeInfaceOidPrefix = "1.3.6.1.2.1.17.1.4.1.2"

	// ifMIB.ifMIBObjects.ifXTable.ifXEntry.ifName
	PoeInfaceNameOidPrefix                                     = "1.3.6.1.2.1.31.1.1.1.1"

)
