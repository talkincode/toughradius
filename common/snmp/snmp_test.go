package snmp

import (
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gosnmp/gosnmp"
	"github.com/hallidave/mibtool/smi"
	"github.com/talkincode/toughradius/common/snmp/mibs/hostresmib"
)

func TestWalk(t *testing.T) {
	var community = "publicSD2"
	target := "192.168.100.1"

	gosnmp.Default.Target = target
	gosnmp.Default.Community = community
	gosnmp.Default.Timeout = time.Duration(10 * time.Second) // Timeout better suited to walking
	err := gosnmp.Default.Connect()
	if err != nil {
		fmt.Printf("Connect err: %v\n", err)
		os.Exit(1)
	}
	defer gosnmp.Default.Conn.Close()

	rs, err := gosnmp.Default.BulkWalkAll(hostresmib.HOST_RESOURCES_MIB_hrProcessorLoad_OID)
	if err != nil {
		fmt.Printf("Walk Error: %v\n", err)
		os.Exit(1)
	}
	for _, pdu := range rs {
		fmt.Printf("name:%s = ", pdu.Name)

		switch pdu.Type {
		case gosnmp.OctetString:
			b := pdu.Value.([]byte)
			fmt.Printf("STRING: %s\n", string(b))
		default:
			fmt.Printf(" TYPE %d: %d\n", pdu.Type, gosnmp.ToBigInt(pdu.Value))
		}
	}
}

func TestWalkLoad(t *testing.T) {
	snmpc, _ := NewSnmpV2Client("192.168.100.1", 161, "publicSD2")
	defer snmpc.Close()

	rs, err := snmpc.Snmpc.Get([]string{
		".1.3.6.1.2.1.1.5.0",
		".1.3.6.1.2.1.1.6.0",
		".1.3.6.1.2.1.1.1.0",
		".1.3.6.1.4.1.14988.1.1.7.3.0",
		".1.3.6.1.4.1.14988.1.1.4.0",
		".1.3.6.1.2.1.25.2.1.2.0",
	})
	if err != nil {
		fmt.Printf("Walk Error: %v\n", err)
		os.Exit(1)
	}

	for _, pdu := range rs.Variables {
		switch pdu.Type {
		case gosnmp.OctetString:
			b := pdu.Value.([]byte)
			fmt.Printf("STRING: %s\n", string(b))
		default:
			fmt.Printf(" TYPE %d: %d\n", pdu.Type, gosnmp.ToBigInt(pdu.Value))
		}
	}
	// for _, pdu := range rs {
	// 	fmt.Printf("name:%s = ", pdu.Name)
	//
	// 	switch pdu.Type {
	// 	case gosnmp.OctetString:
	// 		b := pdu.Value.([]byte)
	// 		fmt.Printf("STRING: %s\n", string(b))
	// 	default:
	// 		fmt.Printf(" TYPE %d: %d\n", pdu.Type, gosnmp.ToBigInt(pdu.Value))
	// 	}
	// }
}

func TestMibParseIFMIB(t *testing.T) {
	mib := smi.NewMIB("/Users/wangjuntao/github/TeamsACS-HY/assets/mibs/ruijie")
	mib.Debug = true
	err := mib.LoadModules("DIFFSERV-MIB")
	if err != nil {
		log.Fatal(err)
	}

	// Walk all symbols in MIB
	mib.VisitSymbols(func(sym *smi.Symbol, oid smi.OID) {
		_s := strings.ReplaceAll(sym.String(), "-", "_")
		_s = strings.ReplaceAll(_s, "::", "_")
		fmt.Printf("%-45s = \"%s\"\n", _s+"_OID", oid)
		fmt.Printf("%-45s = \"%s\"\n", _s+"_NAME", sym.String())
	})
}
func TestMibParseUCDMIB(t *testing.T) {
	mib := smi.NewMIB("/usr/share/snmp/mibs")
	mib.Debug = true
	err := mib.LoadModules("ENTITY-MIB")
	if err != nil {
		log.Fatal(err)
	}

	// Walk all symbols in MIB
	mib.VisitSymbols(func(sym *smi.Symbol, oid smi.OID) {
		_s := strings.ReplaceAll(sym.String(), "-", "_")
		_s = strings.ReplaceAll(_s, "::", "_")
		fmt.Printf("%-45s = \"%s\"\n", _s+"_OID", oid)
		fmt.Printf("%-45s = \"%s\"\n", _s+"_NAME", sym.String())
	})
}

func TestMibParseMikrotik(t *testing.T) {
	mib := smi.NewMIB("/usr/share/snmp/mibs")
	mib.Debug = true
	err := mib.LoadModules("SNMPv2-MIB")
	if err != nil {
		log.Fatal(err)
	}

	// Walk all symbols in MIB
	mib.VisitSymbols(func(sym *smi.Symbol, oid smi.OID) {
		_s := strings.ReplaceAll(sym.String(), "-", "_")
		_s = strings.ReplaceAll(_s, "::", "_")
		fmt.Printf("%-45s = \"%s\"\n", _s+"_OID", oid)
		fmt.Printf("%-45s = \"%s\"\n", _s+"_NAME", sym.String())
	})
}

func TestXXX(t *testing.T) {
	oid := ".1.3.6.1.2.1.2.2.1.2.13"
	fmt.Println(oid[strings.LastIndex(oid, ".")+1:])
}

func TestSnmpV2Client_QueryInterfaces(t *testing.T) {
	sc, _ := NewSnmpV2Client("192.168.100.1", 161, "publicSD2")
	defer sc.Close()
	dev := &SnmpDevice{
		VendorCode:    "",
		SnOid:         "",
		Sn:            "",
		SystemNameOid: "",
		SystemName:    "",
		ModelOid:      "",
		Model:         "",
		CpuLoad:       0,
		MemUsage:      0,
		DiskUsage:     0,
		IfOctetsMap:   nil,
	}
	err := sc.CollectIfOctets(dev)
	fmt.Println(dev.ToJson())
	if err != nil {
		t.Fatal(err)
	}

}

func TestSnmpV2Client_CollectDevice(t *testing.T) {
	sc, _ := NewSnmpV2Client("192.168.100.1", 161, "publicSD2")
	defer sc.Close()
	dev := &SnmpDevice{
		VendorCode:    "14988",
		SoftwareIdOid: ".1.3.6.1.4.1.14988.1.1.4.1.0",
		SnOid:         ".1.3.6.1.4.1.14988.1.1.7.3.0",
		SystemNameOid: ".1.3.6.1.2.1.1.5.0",
		ModelOid:      ".1.3.6.1.2.1.1.1.0",
		UptimeOid:     ".1.3.6.1.2.1.1.3.0",
		IfOctetsMap:   nil,
	}
	sc.CollectDevice(dev)
	t.Log(dev.ToJson())
}
