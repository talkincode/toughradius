package handlers

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
)

// buildTTLSAVP encodes a single EAP-TTLS AVP in the RFC 5281 §10.1 wire format:
// AVP Code (4) | Flags (1) | AVP Length (3) | [Vendor-ID (4)] | Data, zero-padded
// to the next four-octet boundary. It is the test-side counterpart of
// parseTTLSAVPs and delegates to the production encoder so the two never drift.
func buildTTLSAVP(code uint32, mandatory bool, vendorID uint32, value []byte) []byte {
	return encodeTTLSAVP(code, vendorID, mandatory, value)
}

// padTTLSPassword NUL-pads a PAP password to a 16-octet multiple, as an EAP-TTLS
// client SHOULD to obfuscate the password length (RFC 5281 §11.2.5).
func padTTLSPassword(password string) []byte {
	b := []byte(password)
	pad := (16 - len(b)%16) % 16
	return append(b, make([]byte, pad)...)
}

// encodeTTLSPAP builds the inner PAP AVP flight an EAP-TTLS client tunnels: a
// User-Name AVP followed by a NUL-padded User-Password AVP (RFC 5281 §11.2.5).
func encodeTTLSPAP(username, password string) []byte {
	var buf []byte
	buf = append(buf, buildTTLSAVP(ttlsAVPCodeUserName, true, 0, []byte(username))...)
	buf = append(buf, buildTTLSAVP(ttlsAVPCodeUserPassword, true, 0, padTTLSPassword(password))...)
	return buf
}

func TestParseTTLSAVPs_UserNameAndPassword(t *testing.T) {
	flight := encodeTTLSPAP("alice", "S3cr3t!")

	avps, err := parseTTLSAVPs(flight)
	require.NoError(t, err)
	require.Len(t, avps, 2)

	name, ok := findTTLSAVP(avps, ttlsAVPCodeUserName)
	require.True(t, ok, "User-Name AVP must be present")
	assert.Equal(t, "alice", string(name))

	rawPwd, ok := findTTLSAVP(avps, ttlsAVPCodeUserPassword)
	require.True(t, ok, "User-Password AVP must be present")
	// The encoded password is NUL-padded to a 16-octet multiple.
	assert.Equal(t, 16, len(rawPwd))
	assert.Equal(t, "S3cr3t!", string(stripTTLSPasswordPadding(rawPwd)))

	// Both AVPs carry the Mandatory flag and no Vendor-ID.
	for _, a := range avps {
		assert.True(t, a.Mandatory)
		assert.Zero(t, a.VendorID)
	}
}

func TestParseTTLSAVPs_NonMultipleOfFourIsPadded(t *testing.T) {
	// "abc" (3 bytes) makes the User-Name AVP 11 bytes, which must be padded to
	// 12 so the following User-Password AVP still parses on a four-octet
	// boundary (RFC 5281 §10.2).
	flight := append(buildTTLSAVP(ttlsAVPCodeUserName, false, 0, []byte("abc")),
		buildTTLSAVP(ttlsAVPCodeUserPassword, false, 0, []byte("pw"))...)

	avps, err := parseTTLSAVPs(flight)
	require.NoError(t, err)
	require.Len(t, avps, 2)

	name, ok := findTTLSAVP(avps, ttlsAVPCodeUserName)
	require.True(t, ok)
	assert.Equal(t, "abc", string(name))
	pw, ok := findTTLSAVP(avps, ttlsAVPCodeUserPassword)
	require.True(t, ok)
	assert.Equal(t, "pw", string(pw))
}

func TestParseTTLSAVPs_VendorAVPNotMatchedByNonVendorLookup(t *testing.T) {
	// A vendor AVP that happens to reuse the User-Password code must not satisfy
	// a non-vendor User-Password lookup.
	flight := buildTTLSAVP(ttlsAVPCodeUserPassword, false, 99, []byte("vendor"))

	avps, err := parseTTLSAVPs(flight)
	require.NoError(t, err)
	require.Len(t, avps, 1)
	assert.Equal(t, uint32(99), avps[0].VendorID)
	assert.Equal(t, "vendor", string(avps[0].Data))

	_, ok := findTTLSAVP(avps, ttlsAVPCodeUserPassword)
	assert.False(t, ok, "a vendor AVP must not match a non-vendor lookup")
}

func TestParseTTLSAVPs_FinalAVPWithoutPaddingTolerated(t *testing.T) {
	// Hand-build a single AVP with a 3-byte value and omit the trailing padding
	// byte; the parser tolerates a missing final pad (RFC 5281 §10.2).
	const code = 7
	value := []byte("abc")
	length := ttlsAVPHeaderLen + len(value) // 11
	buf := make([]byte, length)
	binary.BigEndian.PutUint32(buf[0:4], code)
	buf[4] = 0
	buf[5] = byte((length >> 16) & 0xFF)
	buf[6] = byte((length >> 8) & 0xFF)
	buf[7] = byte(length & 0xFF)
	copy(buf[ttlsAVPHeaderLen:], value)

	avps, err := parseTTLSAVPs(buf)
	require.NoError(t, err)
	require.Len(t, avps, 1)
	assert.Equal(t, uint32(code), avps[0].Code)
	assert.Equal(t, "abc", string(avps[0].Data))
}

func TestParseTTLSAVPs_Empty(t *testing.T) {
	avps, err := parseTTLSAVPs(nil)
	require.NoError(t, err)
	assert.Empty(t, avps)
}

func TestParseTTLSAVPs_Malformed(t *testing.T) {
	cases := map[string][]byte{
		"truncated header":      {0x00, 0x00, 0x00, 0x01, 0x40, 0x00}, // 6 bytes < 8
		"length below header":   {0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x04},
		"length overruns input": {0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0xFF, 0x00, 0x00, 0x00, 0x00},
		"truncated vendor": {
			0x00, 0x00, 0x00, 0x01, // code
			ttlsAVPFlagVendor, // V flag set
			0x00, 0x00, 0x0C,  // length 12
			0x00, 0x00, 0x00, // only 3 of 4 vendor-id bytes
		},
	}
	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := parseTTLSAVPs(data)
			require.Error(t, err)
			assert.ErrorIs(t, err, eap.ErrTTLSInnerProtocol)
		})
	}
}

func TestStripTTLSPasswordPadding(t *testing.T) {
	assert.Equal(t, "pass", string(stripTTLSPasswordPadding([]byte("pass\x00\x00\x00\x00"))))
	assert.Equal(t, "pass", string(stripTTLSPasswordPadding([]byte("pass"))))
	assert.Empty(t, stripTTLSPasswordPadding([]byte{0x00, 0x00, 0x00}))
	assert.True(t, bytes.Equal([]byte("a\x00b"), stripTTLSPasswordPadding([]byte("a\x00b\x00\x00"))),
		"only trailing NULs are stripped")
}
