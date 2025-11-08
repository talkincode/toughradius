package radiusd

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"testing"
	"time"

	"go.uber.org/zap"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2866"
	"layeh.com/radius/rfc2869"
)

func getAuthPacket() *radius.Packet {
	packet := radius.New(radius.CodeAccessRequest, []byte(`secret`))
	rfc2865.UserName_SetString(packet, "test01")
	rfc2865.UserPassword_SetString(packet, "111111")
	rfc2865.CallingStationID_SetString(packet, "10.10.10.10")
	rfc2865.NASIdentifier_Set(packet, []byte("tradtest"))
	rfc2865.NASIPAddress_Set(packet, net.ParseIP("127.0.0.1"))
	rfc2865.NASPort_Set(packet, 0)
	rfc2865.NASPortType_Set(packet, 0)
	rfc2869.NASPortID_Set(packet, []byte("slot=2;subslot=2;port=22;vlanid=100;"))
	rfc2865.CalledStationID_SetString(packet, "11:11:11:11:11:11")
	rfc2865.CallingStationID_SetString(packet, "11:11:11:11:11:11")
	return packet
}

func TestAuth(t *testing.T) {
	packet := radius.New(radius.CodeAccessRequest, []byte(`secret`))
	rfc2865.UserName_SetString(packet, "test01")
	rfc2865.UserPassword_SetString(packet, "111111")
	rfc2865.CallingStationID_SetString(packet, "10.10.10.10")
	rfc2865.NASIdentifier_Set(packet, []byte("tradtest"))
	rfc2865.NASIPAddress_Set(packet, net.ParseIP("127.0.0.1"))
	rfc2865.NASPort_Set(packet, 0)
	rfc2865.NASPortType_Set(packet, 0)
	rfc2869.NASPortID_Set(packet, []byte("slot=2;subslot=2;port=22;vlanid=100;"))
	rfc2865.CalledStationID_SetString(packet, "11:11:11:11:11:11")
	rfc2865.CallingStationID_SetString(packet, "11:11:11:11:11:11")
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	fmt.Println(packet)
	cli := radius.Client{
		Net:                "udp",
		Retry:              0,
		MaxPacketErrors:    10,
		InsecureSkipVerify: true,
	}
	response, err := cli.Exchange(ctx, packet, "127.0.0.1:1812")
	if err != nil {
		zap.S().Fatal(err)
	}

	zap.S().Info("Code:", response.Code)
}

func getAcctPacket(sessionid, ipaddr string, acctType rfc2866.AcctStatusType) *radius.Packet {
	packet := radius.New(radius.CodeAccountingRequest, []byte(`secret`))
	rfc2865.UserName_SetString(packet, "test01")
	rfc2865.NASIdentifier_Set(packet, []byte("tradtest"))
	rfc2865.NASIPAddress_Set(packet, net.ParseIP("127.0.0.1"))
	rfc2865.NASPort_Set(packet, 0)
	rfc2865.NASPortType_Set(packet, 0)
	rfc2869.NASPortID_Set(packet, []byte("slot=2;subslot=2;port=22;vlanid=100;"))
	rfc2865.CalledStationID_Set(packet, []byte("11:11:11:11:11:11"))
	rfc2865.CallingStationID_Set(packet, []byte("11:11:11:11:11:11"))
	rfc2866.AcctSessionID_SetString(packet, sessionid)
	rfc2865.FramedIPAddress_Set(packet, net.ParseIP(ipaddr))
	rfc2866.AcctStatusType_Set(packet, acctType)
	return packet
}

func TestAcctStart(t *testing.T) {
	start := getAcctPacket("123456", "10.10.10.2", rfc2866.AcctStatusType_Value_Start)
	rfc2866.AcctInputOctets_Set(start, 0)
	rfc2866.AcctOutputOctets_Set(start, 0)
	ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
	fmt.Println(start)
	cli := radius.Client{
		Net:                "udp",
		Retry:              0,
		MaxPacketErrors:    10,
		InsecureSkipVerify: true,
	}
	response, err := cli.Exchange(ctx, start, "127.0.0.1:1813")
	if err != nil {
		zap.S().Fatal(err)
	}

	zap.S().Info("Code:", response.Code)
}

func TestAcctUpdate(t *testing.T) {
	start := getAcctPacket("123456", "10.10.10.2", rfc2866.AcctStatusType_Value_InterimUpdate)
	rfc2866.AcctInputOctets_Set(start, 102400)
	rfc2866.AcctOutputOctets_Set(start, 102400)
	ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
	fmt.Println(start)
	cli := radius.Client{
		Net:                "udp",
		Retry:              0,
		MaxPacketErrors:    10,
		InsecureSkipVerify: true,
	}
	response, err := cli.Exchange(ctx, start, "127.0.0.1:1813")
	if err != nil {
		zap.S().Fatal(err)
	}

	zap.S().Info("Code:", response.Code)
}

func TestAcctStop(t *testing.T) {
	start := getAcctPacket("123456", "10.10.10.2", rfc2866.AcctStatusType_Value_Stop)
	rfc2866.AcctInputOctets_Set(start, 409600)
	rfc2866.AcctOutputOctets_Set(start, 409600)
	ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
	fmt.Println(start)
	cli := radius.Client{
		Net:                "udp",
		Retry:              0,
		MaxPacketErrors:    10,
		InsecureSkipVerify: true,
	}
	response, err := cli.Exchange(ctx, start, "127.0.0.1:1813")
	if err != nil {
		zap.S().Fatal(err)
	}

	zap.S().Info("Code:", response.Code)
}

func TestTlsClient(t *testing.T) {
	cert, err := tls.LoadX509KeyPair("../assets/client.tls.crt", "../assets/client.tls.key")
	if err != nil {
		t.Fatal(err)
	}
	conn, err := tls.Dial("tcp", "127.0.0.1:2083", &tls.Config{
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{cert},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(conn)
	pkt := getAuthPacket()
	bs, err := pkt.Encode()
	if err != nil {
		t.Fatal(err)
	}
	_, err = conn.Write(bs)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second * 3)
}
