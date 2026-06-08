package mikrotik

import (
	"testing"

	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

// countSubAttr returns the number of Mikrotik vendor sub-attributes of the
// given type present across every Vendor-Specific attribute of p.
func countSubAttr(t *testing.T, p *radius.Packet, typ byte) int {
	t.Helper()
	count := 0
	for _, avp := range p.Attributes {
		if avp.Type != rfc2865.VendorSpecific_Type {
			continue
		}
		vendorID, vsa, err := radius.VendorSpecific(avp.Attribute)
		if err != nil || vendorID != _Mikrotik_VendorID {
			continue
		}
		for len(vsa) >= 3 {
			vsaTyp, vsaLen := vsa[0], vsa[1]
			if int(vsaLen) < 3 || int(vsaLen) > len(vsa) {
				t.Fatalf("malformed VSA payload: sub-attr length %d in %d remaining bytes", vsaLen, len(vsa))
			}
			if vsaTyp == typ {
				count++
			}
			vsa = vsa[int(vsaLen):]
		}
		if len(vsa) != 0 {
			t.Fatalf("VSA payload has %d trailing/stale byte(s)", len(vsa))
		}
	}
	return count
}

func vsaCount(p *radius.Packet) int {
	n := 0
	for _, avp := range p.Attributes {
		if avp.Type == rfc2865.VendorSpecific_Type {
			n++
		}
	}
	return n
}

// TestSetVendorReplaceSameAttribute is the regression test for
// talkincode/toughradius#324: setting the same vendor attribute twice used to
// panic ("slice bounds out of range") because the sub-attribute scan read a
// fixed offset (vsa[0]/vsa[1]) instead of the running cursor and advanced the
// cursor past the shrunken slice.
func TestSetVendorReplaceSameAttribute(t *testing.T) {
	p := &radius.Packet{}

	if err := MikrotikRecvLimit_Set(p, 1000); err != nil {
		t.Fatalf("first set: %v", err)
	}
	if err := MikrotikRecvLimit_Set(p, 2000); err != nil {
		t.Fatalf("second set: %v", err)
	}

	if got, err := MikrotikRecvLimit_Lookup(p); err != nil || got != 2000 {
		t.Fatalf("MikrotikRecvLimit_Lookup = %d, %v; want 2000, nil", got, err)
	}
	if n := countSubAttr(t, p, 1); n != 1 {
		t.Fatalf("RecvLimit sub-attribute count = %d; want 1", n)
	}
	if n := vsaCount(p); n != 1 {
		t.Fatalf("vendor-specific attribute count = %d; want 1", n)
	}
}

// TestSetVendorDeletesEmptyVSAAtNonZeroIndex exercises the empty-VSA removal
// branch when the matching attribute is not the first packet attribute. The
// old delete index "i+i" went out of range (panic) for i >= 2 and dropped the
// wrong elements; it must be "i+1".
func TestSetVendorDeletesEmptyVSAAtNonZeroIndex(t *testing.T) {
	p := &radius.Packet{}
	rfc2865.UserName_SetString(p, "alice")
	rfc2865.NASIdentifier_SetString(p, "nas-1")

	MikrotikRecvLimit_Set(p, 1000) // VSA now at index 2
	MikrotikRecvLimit_Set(p, 2000) // removes the empty VSA at index 2, re-adds

	if got, err := MikrotikRecvLimit_Lookup(p); err != nil || got != 2000 {
		t.Fatalf("MikrotikRecvLimit_Lookup = %d, %v; want 2000, nil", got, err)
	}
	if got := rfc2865.UserName_GetString(p); got != "alice" {
		t.Fatalf("UserName_GetString = %q; want %q", got, "alice")
	}
	if got := rfc2865.NASIdentifier_GetString(p); got != "nas-1" {
		t.Fatalf("NASIdentifier_GetString = %q; want %q", got, "nas-1")
	}
	if n := vsaCount(p); n != 1 {
		t.Fatalf("vendor-specific attribute count = %d; want 1", n)
	}
}

// TestSetVendorPreservesOtherSubAttributes verifies that re-setting one
// sub-attribute inside a VSA that packs several sub-attributes only removes the
// matching one, rewrites the VSA to its new (shorter) length without leaving
// stale trailing bytes, and preserves the other sub-attributes.
func TestSetVendorPreservesOtherSubAttributes(t *testing.T) {
	p := &radius.Packet{}

	// A single VSA holding two Mikrotik sub-attributes:
	//   type 1 (RecvLimit) = 1000, type 2 (XmitLimit) = 500.
	sub := []byte{
		1, 6, 0, 0, 0x03, 0xE8, // RecvLimit = 1000
		2, 6, 0, 0, 0x01, 0xF4, // XmitLimit = 500
	}
	vsa, err := radius.NewVendorSpecific(_Mikrotik_VendorID, sub)
	if err != nil {
		t.Fatalf("NewVendorSpecific: %v", err)
	}
	p.Add(rfc2865.VendorSpecific_Type, vsa)

	if err := MikrotikRecvLimit_Set(p, 2000); err != nil {
		t.Fatalf("set: %v", err)
	}

	if got, err := MikrotikRecvLimit_Lookup(p); err != nil || got != 2000 {
		t.Fatalf("MikrotikRecvLimit_Lookup = %d, %v; want 2000, nil", got, err)
	}
	if got, err := MikrotikXmitLimit_Lookup(p); err != nil || got != 500 {
		t.Fatalf("MikrotikXmitLimit_Lookup = %d, %v; want 500, nil", got, err)
	}
	if n := countSubAttr(t, p, 1); n != 1 {
		t.Fatalf("RecvLimit sub-attribute count = %d; want 1", n)
	}
	if n := countSubAttr(t, p, 2); n != 1 {
		t.Fatalf("XmitLimit sub-attribute count = %d; want 1 (stale bytes left behind?)", n)
	}
}
