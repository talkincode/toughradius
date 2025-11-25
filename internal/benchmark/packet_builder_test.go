package benchmark

import (
	"net"
	"testing"

	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2866"
)

func TestPacketBuilderBuildAuthRequest(t *testing.T) {
	pb := NewPacketBuilder("secret", "benchmark-nas", "192.168.1.100")
	pkt, err := pb.BuildAuthRequest("alice", "password", "AA-BB-CC")
	if err != nil {
		t.Fatalf("BuildAuthRequest returned error: %v", err)
	}

	if pkt.Code != radius.CodeAccessRequest {
		t.Fatalf("expected Access-Request code, got %v", pkt.Code)
	}

	if got := rfc2865.UserName_GetString(pkt); got != "alice" {
		t.Fatalf("unexpected User-Name: %q", got)
	}

	if got := rfc2865.CallingStationID_GetString(pkt); got != "AA-BB-CC" {
		t.Fatalf("unexpected Calling-Station-ID: %q", got)
	}

	if got := rfc2865.NASIPAddress_Get(pkt); !got.Equal(net.ParseIP("192.168.1.100")) {
		t.Fatalf("unexpected NAS-IP-Address: %v", got)
	}
}

func TestPacketBuilderBuildAcctRequest(t *testing.T) {
	pb := NewPacketBuilder("secret", "benchmark-nas", "10.0.0.1")
	statusStart := rfc2866.AcctStatusType(1)
	pkt, err := pb.BuildAcctRequest("bob", "172.16.0.10", "DD-EE-FF", "session-1", statusStart)
	if err != nil {
		t.Fatalf("BuildAcctRequest returned error: %v", err)
	}

	if pkt.Code != radius.CodeAccountingRequest {
		t.Fatalf("expected Accounting-Request, got %v", pkt.Code)
	}

	if got := rfc2865.UserName_GetString(pkt); got != "bob" {
		t.Fatalf("unexpected User-Name: %q", got)
	}

	if got := rfc2865.FramedIPAddress_Get(pkt); !got.Equal(net.ParseIP("172.16.0.10")) {
		t.Fatalf("unexpected Framed-IP-Address: %v", got)
	}

	if got := rfc2866.AcctSessionID_GetString(pkt); got != "session-1" {
		t.Fatalf("unexpected Acct-Session-ID: %q", got)
	}

	if got := rfc2866.AcctStatusType_Get(pkt); got != statusStart {
		t.Fatalf("unexpected Acct-Status-Type: %v", got)
	}
}

func TestParseIPFallback(t *testing.T) {
	if got := parseIP("invalid"); !got.Equal(net.ParseIP("0.0.0.0")) {
		t.Fatalf("expected 0.0.0.0 fallback, got %v", got)
	}
}
