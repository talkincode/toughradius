package handlers

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
)

// EAP-TTLS AVP codes and flags (RFC 5281 §10.1). Inner authentication is carried
// as a sequence of Diameter-style AVPs inside the TLS tunnel; non-vendor AVP
// codes are the corresponding RADIUS attribute types (RFC 5281 §10.1, RFC 2865).
const (
	// ttlsAVPCodeUserName is the User-Name AVP code (RADIUS attribute 1, RFC
	// 2865 §5.1): the identity the peer presents inside the tunnel.
	ttlsAVPCodeUserName = 1
	// ttlsAVPCodeUserPassword is the User-Password AVP code (RADIUS attribute 2,
	// RFC 2865 §5.2). For EAP-TTLS PAP its value is the cleartext password (not
	// RADIUS-obfuscated, since the TLS tunnel protects it), NUL-padded to a
	// 16-octet multiple to obfuscate length (RFC 5281 §11.2.5).
	ttlsAVPCodeUserPassword = 2

	// ttlsAVPFlagVendor (V) signals a 4-octet Vendor-ID follows the AVP Length
	// (RFC 5281 §10.1).
	ttlsAVPFlagVendor = 0x80
	// ttlsAVPFlagMandatory (M) marks support of the AVP as required by the
	// sender (RFC 5281 §10.1). Reserved bits are ignored on receipt.
	ttlsAVPFlagMandatory = 0x40

	// ttlsAVPHeaderLen is the AVP header length without a Vendor-ID: AVP Code
	// (4) + Flags (1) + AVP Length (3).
	ttlsAVPHeaderLen = 8
	// ttlsAVPVendorHeaderLen is the AVP header length including a Vendor-ID.
	ttlsAVPVendorHeaderLen = 12
)

// ttlsAVP is a decoded EAP-TTLS Diameter AVP (RFC 5281 §10.1).
type ttlsAVP struct {
	// Code is the AVP Code; for non-vendor AVPs it equals a RADIUS attribute
	// type.
	Code uint32
	// Mandatory reports whether the M (Mandatory) flag was set.
	Mandatory bool
	// VendorID is the SMI Network Management Private Enterprise Code when the V
	// flag is set, or 0 for a non-vendor AVP.
	VendorID uint32
	// Data is the AVP value with any four-octet boundary padding removed.
	Data []byte
}

// parseTTLSAVPs decodes a concatenated sequence of EAP-TTLS AVPs from the
// decrypted tunnel payload (RFC 5281 §10.2). Each AVP is laid out as
//
//	AVP Code (4) | Flags (1) | AVP Length (3) | [Vendor-ID (4)] | Data
//
// and is padded with zeros to the next four-octet boundary; the AVP Length
// counts the header and data but not the padding. The final AVP may omit its
// trailing padding. parseTTLSAVPs returns eap.ErrTTLSInnerProtocol for a
// truncated header, an AVP Length below its header size, or an AVP Length that
// overruns the remaining payload.
func parseTTLSAVPs(data []byte) ([]ttlsAVP, error) {
	var avps []ttlsAVP
	for len(data) > 0 {
		if len(data) < ttlsAVPHeaderLen {
			return nil, fmt.Errorf("%w: truncated AVP header (%d bytes remaining)", eap.ErrTTLSInnerProtocol, len(data))
		}

		code := binary.BigEndian.Uint32(data[0:4])
		flags := data[4]
		// AVP Length is a 3-octet field, so it fits in an int without overflow.
		length := int(data[5])<<16 | int(data[6])<<8 | int(data[7])

		headerLen := ttlsAVPHeaderLen
		var vendorID uint32
		if flags&ttlsAVPFlagVendor != 0 {
			if len(data) < ttlsAVPVendorHeaderLen {
				return nil, fmt.Errorf("%w: truncated vendor AVP header (%d bytes remaining)", eap.ErrTTLSInnerProtocol, len(data))
			}
			vendorID = binary.BigEndian.Uint32(data[8:12])
			headerLen = ttlsAVPVendorHeaderLen
		}

		if length < headerLen {
			return nil, fmt.Errorf("%w: AVP Length %d is below the %d-octet header", eap.ErrTTLSInnerProtocol, length, headerLen)
		}
		if len(data) < length {
			return nil, fmt.Errorf("%w: AVP Length %d overruns %d remaining bytes", eap.ErrTTLSInnerProtocol, length, len(data))
		}

		value := append([]byte(nil), data[headerLen:length]...)
		avps = append(avps, ttlsAVP{
			Code:      code,
			Mandatory: flags&ttlsAVPFlagMandatory != 0,
			VendorID:  vendorID,
			Data:      value,
		})

		// Advance past the AVP and its zero padding to the next four-octet
		// boundary (RFC 5281 §10.2). The final AVP may omit trailing padding, in
		// which case consume whatever remains. The AVP Length is at least the
		// header size (>= 8), so each iteration makes progress.
		advance := (length + 3) &^ 3
		if advance > len(data) {
			advance = len(data)
		}
		data = data[advance:]
	}
	return avps, nil
}

// findTTLSAVP returns the value of the first non-vendor AVP with the given code
// and whether such an AVP was present.
func findTTLSAVP(avps []ttlsAVP, code uint32) ([]byte, bool) {
	for i := range avps {
		if avps[i].Code == code && avps[i].VendorID == 0 {
			return avps[i].Data, true
		}
	}
	return nil, false
}

// findTTLSVendorAVP returns the value of the first AVP carrying the given
// Vendor-ID and Code (the V flag was set) and whether such an AVP was present.
// It is used for the Microsoft MS-CHAP AVPs that an inner MS-CHAP-V2 exchange
// tunnels (RFC 5281 §11.2.4, RFC 2548); use findTTLSAVP for base-protocol AVPs.
func findTTLSVendorAVP(avps []ttlsAVP, vendorID, code uint32) ([]byte, bool) {
	for i := range avps {
		if avps[i].VendorID == vendorID && avps[i].Code == code {
			return avps[i].Data, true
		}
	}
	return nil, false
}

// encodeTTLSAVP serializes a single EAP-TTLS AVP (RFC 5281 §10.1) and appends the
// zero padding to the next four-octet boundary. A non-zero vendorID sets the V
// flag and emits the 4-octet Vendor-ID; mandatory sets the M flag. It builds the
// AVPs the TTLS server tunnels back to the peer (for example MS-CHAP2-Success).
func encodeTTLSAVP(code, vendorID uint32, mandatory bool, data []byte) []byte {
	headerLen := ttlsAVPHeaderLen
	var flags byte
	if mandatory {
		flags |= ttlsAVPFlagMandatory
	}
	if vendorID != 0 {
		flags |= ttlsAVPFlagVendor
		headerLen = ttlsAVPVendorHeaderLen
	}

	length := headerLen + len(data)
	buf := make([]byte, (length+3)&^3)
	binary.BigEndian.PutUint32(buf[0:4], code)
	buf[4] = flags
	buf[5] = byte((length >> 16) & 0xFF)
	buf[6] = byte((length >> 8) & 0xFF)
	buf[7] = byte(length & 0xFF)

	off := ttlsAVPHeaderLen
	if vendorID != 0 {
		binary.BigEndian.PutUint32(buf[8:12], vendorID)
		off = ttlsAVPVendorHeaderLen
	}
	copy(buf[off:], data)
	return buf
}

// stripTTLSPasswordPadding removes the trailing NUL padding an EAP-TTLS client
// adds to the PAP password to obfuscate its length (RFC 5281 §11.2.5: the
// password SHOULD be NUL-padded to a 16-octet multiple). The password is carried
// in cleartext inside the TLS tunnel, so only the trailing zero octets are
// stripped before comparison.
func stripTTLSPasswordPadding(b []byte) []byte {
	return bytes.TrimRight(b, "\x00")
}
