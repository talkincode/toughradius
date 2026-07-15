// Package radback provides a hand-maintained minimal RADIUS dictionary of the
// Vendor-Specific Attributes (RFC 2865 §5.26) registered under SMI Network
// Management Private Enterprise Code 2352 (the "radback"/Redback vendor
// namespace).
//
// Unlike the other vendor packages, this dictionary is deliberately pruned to
// the attributes actually consumed by the radback request parser
// (internal/radiusd/plugins/vendorparsers/parsers/radback_parser.go):
//
//   - Mac-Addr (attribute 145, string)
//   - Bind-Dot1q-Vlan-Tag-Id (attribute 54, integer)
//
// The full radius-dict-gen output for this vendor was ~15k lines with only the
// two accessors above referenced anywhere in the tree, so the generated file
// was replaced by this minimal hand-written equivalent. The accessor names and
// wire behavior are identical to the generated code; the round-trip tests in
// radback_test.go pin the wire contract. If a wider attribute surface is ever
// needed, regenerate the full dictionary with radius-dict-gen instead of
// extending this file attribute by attribute.
package radback

import (
	"fmt"

	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

// _Redback_VendorID is the IANA private enterprise number of the
// radback/Redback vendor namespace.
const _Redback_VendorID = 2352

// Attribute numbers of the VSAs kept in this pruned dictionary.
const (
	_Redback_MacAddr_Type            = 145
	_Redback_BindDot1qVlanTagID_Type = 54
)

func _Redback_AddVendor(p *radius.Packet, typ byte, attr radius.Attribute) (err error) {
	// RFC 2865 §5: an attribute (here the inner vendor TLV) carries a one-byte
	// length, so the value must fit 255-2 bytes; guard before the byte cast.
	if len(attr) > 253 {
		return fmt.Errorf("radback: VSA value too long (%d bytes)", len(attr))
	}
	var vsa radius.Attribute
	vendor := make(radius.Attribute, 2+len(attr))
	vendor[0] = typ
	vendor[1] = byte(len(vendor)) //nolint:gosec // G115: len(vendor) ≤ 255 guarded above
	copy(vendor[2:], attr)
	vsa, err = radius.NewVendorSpecific(_Redback_VendorID, vendor)
	if err != nil {
		return
	}
	p.Add(rfc2865.VendorSpecific_Type, vsa)
	return
}

func _Redback_LookupVendor(p *radius.Packet, typ byte) (attr radius.Attribute, ok bool) {
	for _, avp := range p.Attributes {
		if avp.Type != rfc2865.VendorSpecific_Type {
			continue
		}
		attr := avp.Attribute
		vendorID, vsa, err := radius.VendorSpecific(attr)
		if err != nil || vendorID != _Redback_VendorID {
			continue
		}
		for len(vsa) >= 3 {
			vsaTyp, vsaLen := vsa[0], vsa[1]
			if int(vsaLen) > len(vsa) || vsaLen < 3 {
				break
			}
			if vsaTyp == typ {
				return vsa[2:int(vsaLen)], true
			}
			vsa = vsa[int(vsaLen):]
		}
	}
	return
}

func _Redback_SetVendor(p *radius.Packet, typ byte, attr radius.Attribute) (err error) {
	for i := 0; i < len(p.Attributes); {
		avp := p.Attributes[i]
		if avp.Type != rfc2865.VendorSpecific_Type {
			i++
			continue
		}
		vendorID, vsa, err := radius.VendorSpecific(avp.Attribute)
		if err != nil || vendorID != _Redback_VendorID {
			i++
			continue
		}
		for j := 0; len(vsa[j:]) >= 3; {
			vsaTyp, vsaLen := vsa[j], vsa[j+1]
			if int(vsaLen) > len(vsa[j:]) || vsaLen < 3 {
				break
			}
			if vsaTyp == typ {
				vsa = append(vsa[:j], vsa[j+int(vsaLen):]...)
			} else {
				j += int(vsaLen)
			}
		}
		if len(vsa) > 0 {
			p.Attributes[i].Attribute = append(avp.Attribute[:4:4], vsa...)
			i++
		} else {
			p.Attributes = append(p.Attributes[:i], p.Attributes[i+1:]...)
		}
	}
	return _Redback_AddVendor(p, typ, attr)
}

func _Redback_DelVendor(p *radius.Packet, typ byte) {
vsaLoop:
	for i := 0; i < len(p.Attributes); {
		avp := p.Attributes[i]
		if avp.Type != rfc2865.VendorSpecific_Type {
			i++
			continue
		}
		vendorID, vsa, err := radius.VendorSpecific(avp.Attribute)
		if err != nil || vendorID != _Redback_VendorID {
			i++
			continue
		}
		offset := 0
		for len(vsa[offset:]) >= 3 {
			vsaTyp, vsaLen := vsa[offset], vsa[offset+1]
			if int(vsaLen) > len(vsa) || vsaLen < 3 {
				continue vsaLoop
			}
			if vsaTyp == typ {
				copy(vsa[offset:], vsa[offset+int(vsaLen):])
				vsa = vsa[:len(vsa)-int(vsaLen)]
			} else {
				offset += int(vsaLen)
			}
		}
		if offset == 0 {
			p.Attributes = append(p.Attributes[:i], p.Attributes[i+1:]...)
		} else {
			i++
		}
	}
}

// MacAddr_LookupString returns the first Redback Mac-Addr (145) VSA on p as a
// string, or radius.ErrNoAttribute if the packet carries none.
func MacAddr_LookupString(p *radius.Packet) (value string, err error) {
	a, ok := _Redback_LookupVendor(p, _Redback_MacAddr_Type)
	if !ok {
		err = radius.ErrNoAttribute
		return
	}
	value = radius.String(a)
	return
}

// MacAddr_GetString returns the first Redback Mac-Addr (145) VSA on p as a
// string, or the empty string if the packet carries none.
func MacAddr_GetString(p *radius.Packet) (value string) {
	value, _ = MacAddr_LookupString(p)
	return
}

// MacAddr_SetString replaces any Redback Mac-Addr (145) VSAs on p with the
// given string value.
func MacAddr_SetString(p *radius.Packet, value string) (err error) {
	var a radius.Attribute
	a, err = radius.NewString(value)
	if err != nil {
		return
	}
	return _Redback_SetVendor(p, _Redback_MacAddr_Type, a)
}

// MacAddr_Del removes all Redback Mac-Addr (145) VSAs from p.
func MacAddr_Del(p *radius.Packet) {
	_Redback_DelVendor(p, _Redback_MacAddr_Type)
}

// BindDot1qVlanTagID is the value of the Redback Bind-Dot1q-Vlan-Tag-Id (54)
// integer VSA, carrying the 802.1Q VLAN tag bound to the subscriber session.
type BindDot1qVlanTagID uint32

// BindDot1qVlanTagID_Lookup returns the first Redback Bind-Dot1q-Vlan-Tag-Id
// (54) VSA on p, or radius.ErrNoAttribute if the packet carries none.
func BindDot1qVlanTagID_Lookup(p *radius.Packet) (value BindDot1qVlanTagID, err error) {
	a, ok := _Redback_LookupVendor(p, _Redback_BindDot1qVlanTagID_Type)
	if !ok {
		err = radius.ErrNoAttribute
		return
	}
	var i uint32
	i, err = radius.Integer(a)
	if err != nil {
		return
	}
	value = BindDot1qVlanTagID(i)
	return
}

// BindDot1qVlanTagID_Get returns the first Redback Bind-Dot1q-Vlan-Tag-Id (54)
// VSA on p, or zero if the packet carries none.
func BindDot1qVlanTagID_Get(p *radius.Packet) (value BindDot1qVlanTagID) {
	value, _ = BindDot1qVlanTagID_Lookup(p)
	return
}

// BindDot1qVlanTagID_Set replaces any Redback Bind-Dot1q-Vlan-Tag-Id (54) VSAs
// on p with the given value.
func BindDot1qVlanTagID_Set(p *radius.Packet, value BindDot1qVlanTagID) (err error) {
	a := radius.NewInteger(uint32(value))
	return _Redback_SetVendor(p, _Redback_BindDot1qVlanTagID_Type, a)
}
