package radiusd

import (
	"bytes"
	"net"
	"sync"
	"testing"

	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2866"
)

// buildBenchAcctPacket builds a realistic, valid Accounting-Request packet whose
// Request Authenticator matches the shared secret, so CheckRequestSecret takes
// its success path (the common case in production). The packet is round-tripped
// through Encode/Parse so its on-wire form is what the hot path actually sees.
func buildBenchAcctPacket(tb testing.TB) (*radius.Packet, []byte) {
	tb.Helper()
	secret := []byte("testing123")
	p := radius.New(radius.CodeAccountingRequest, secret)
	set := func(err error) {
		tb.Helper()
		if err != nil {
			tb.Fatalf("set attribute: %v", err)
		}
	}
	set(rfc2865.UserName_SetString(p, "user0001"))
	set(rfc2865.NASIPAddress_Set(p, net.IPv4(10, 0, 0, 1)))
	set(rfc2865.CallingStationID_SetString(p, "00:11:22:33:44:55"))
	set(rfc2865.CalledStationID_SetString(p, "AP-Office-01"))
	set(rfc2866.AcctSessionID_SetString(p, "sess-0000000000000001"))
	set(rfc2866.AcctStatusType_Set(p, rfc2866.AcctStatusType_Value_InterimUpdate))
	set(rfc2866.AcctInputOctets_Set(p, 123456))
	set(rfc2866.AcctOutputOctets_Set(p, 654321))
	set(rfc2866.AcctSessionTime_Set(p, 3600))

	encoded, err := p.Encode()
	if err != nil {
		tb.Fatalf("encode: %v", err)
	}
	parsed, err := radius.Parse(encoded, secret)
	if err != nil {
		tb.Fatalf("parse: %v", err)
	}
	return parsed, secret
}

// BenchmarkCheckRequestSecret measures allocations on the accounting
// authenticator-verification path, which runs for every accounting packet.
func BenchmarkCheckRequestSecret(b *testing.B) {
	svc := &RadiusService{}
	pkt, secret := buildBenchAcctPacket(b)

	if err := svc.CheckRequestSecret(pkt, secret); err != nil {
		b.Fatalf("expected valid secret, got %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := svc.CheckRequestSecret(pkt, secret); err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

// BenchmarkParseTcpPacket measures allocations on the RadSec per-packet parse
// path. A single bytes.Reader is reset each iteration so only parseTcpPacket's
// own allocations (the data slice, attribute parsing, the packet) are counted.
func BenchmarkParseTcpPacket(b *testing.B) {
	pkt, secret := buildBenchAcctPacket(b)
	encoded, err := pkt.Encode()
	if err != nil {
		b.Fatalf("encode: %v", err)
	}

	rd := bytes.NewReader(nil)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rd.Reset(encoded)
		if _, err := parseTcpPacket(rd, secret); err != nil {
			b.Fatalf("parseTcpPacket: %v", err)
		}
	}
}

// BenchmarkPacketLength measures the per-packet response length calculation.
func BenchmarkPacketLength(b *testing.B) {
	pkt, _ := buildBenchAcctPacket(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Length(pkt)
	}
}

// TestParseTcpPacket_RoundTrip verifies that a packet parsed from the pooled
// buffer reproduces the original code, identifier, attributes and authenticator.
func TestParseTcpPacket_RoundTrip(t *testing.T) {
	pkt, secret := buildBenchAcctPacket(t)
	encoded, err := pkt.Encode()
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	got, err := parseTcpPacket(bytes.NewReader(encoded), secret)
	if err != nil {
		t.Fatalf("parseTcpPacket: %v", err)
	}
	if got.Code != pkt.Code || got.Identifier != pkt.Identifier {
		t.Fatalf("header mismatch: got code=%v id=%v", got.Code, got.Identifier)
	}
	if name := rfc2865.UserName_GetString(got); name != "user0001" {
		t.Fatalf("attribute mismatch: UserName=%q", name)
	}
	if got.Authenticator != pkt.Authenticator {
		t.Fatalf("authenticator mismatch")
	}
}

// TestParseTcpPacket_RejectsMalformedLength verifies the length guards added for
// safe buffer pooling: a length below the header size and a payload shorter than
// the 16-byte authenticator are both rejected instead of underflowing or panicking.
func TestParseTcpPacket_RejectsMalformedLength(t *testing.T) {
	belowHeader := []byte{1, 2, 0, 3} // Length 3 < 4-byte header
	if _, err := parseTcpPacket(bytes.NewReader(belowHeader), []byte("s")); err == nil {
		t.Fatal("expected error for length below header size")
	}

	shortPayload := []byte{1, 2, 0, 18} // Length 18 => dataLen 14 < 16
	if _, err := parseTcpPacket(bytes.NewReader(shortPayload), []byte("s")); err == nil {
		t.Fatal("expected error for payload shorter than authenticator")
	}

	// Length above the RFC 2865 maximum must be rejected before allocation.
	overMax := []byte{1, 2, 0xFF, 0xFF} // Length 65535 > 4096
	if _, err := parseTcpPacket(bytes.NewReader(overMax), []byte("s")); err == nil {
		t.Fatal("expected error for length above RFC 2865 maximum")
	}
}

// TestParseTcpPacket_ConcurrentReuseSafe exercises the shared buffer pool from
// many goroutines; under -race this fails if a pooled buffer is reused while
// still referenced. It guards against attribute corruption from pool misuse.
func TestParseTcpPacket_ConcurrentReuseSafe(t *testing.T) {
	pkt, secret := buildBenchAcctPacket(t)
	encoded, err := pkt.Encode()
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	want := rfc2865.UserName_GetString(pkt)

	var wg sync.WaitGroup
	for g := 0; g < 16; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				got, err := parseTcpPacket(bytes.NewReader(encoded), secret)
				if err != nil {
					t.Errorf("parseTcpPacket: %v", err)
					return
				}
				if rfc2865.UserName_GetString(got) != want {
					t.Errorf("attribute corrupted under concurrent pool reuse")
					return
				}
			}
		}()
	}
	wg.Wait()
}
