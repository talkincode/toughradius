package tlsfragment

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse_FlagsOnly(t *testing.T) {
	// A bare Start (S bit) payload: single flags octet, no length, no data.
	p, err := Parse([]byte{FlagStart})
	require.NoError(t, err)
	assert.True(t, p.Start())
	assert.False(t, p.More())
	assert.False(t, p.HasLength)
	assert.Empty(t, p.Data)
}

func TestParse_LengthAndData(t *testing.T) {
	// L + M set, declared length 0x00000102, then 3 data bytes.
	payload := []byte{FlagLengthIncluded | FlagMoreFragments, 0x00, 0x00, 0x01, 0x02, 0xAA, 0xBB, 0xCC}
	p, err := Parse(payload)
	require.NoError(t, err)
	assert.True(t, p.HasLength)
	assert.True(t, p.More())
	assert.Equal(t, uint32(0x102), p.TLSMessageLength)
	assert.Equal(t, []byte{0xAA, 0xBB, 0xCC}, p.Data)
}

func TestParse_Errors(t *testing.T) {
	_, err := Parse(nil)
	assert.ErrorIs(t, err, ErrEmptyPayload)

	// L flag set but fewer than 4 length octets.
	_, err = Parse([]byte{FlagLengthIncluded, 0x00, 0x01})
	assert.ErrorIs(t, err, ErrShortLength)
}

func TestParse_IsACK(t *testing.T) {
	p, err := Parse([]byte{0x00})
	require.NoError(t, err)
	assert.True(t, p.IsACK())

	p, err = Parse([]byte{FlagMoreFragments})
	require.NoError(t, err)
	assert.False(t, p.IsACK(), "M bit set is not an ACK")
}

func TestEncode_RoundTrip(t *testing.T) {
	cases := []*Packet{
		{Flags: FlagStart},
		{Data: []byte{1, 2, 3}},
		{HasLength: true, TLSMessageLength: 300, Flags: FlagMoreFragments, Data: bytes.Repeat([]byte{9}, 10)},
		{HasLength: true, TLSMessageLength: 0},
	}
	for _, in := range cases {
		out, err := Parse(in.Encode())
		require.NoError(t, err)
		assert.Equal(t, in.HasLength, out.HasLength)
		assert.Equal(t, in.TLSMessageLength, out.TLSMessageLength)
		assert.Equal(t, in.More(), out.More())
		assert.Equal(t, in.Data, out.Data)
	}
}

func TestEncode_SetsLengthFlag(t *testing.T) {
	p := &Packet{HasLength: true, TLSMessageLength: 4, Data: []byte{1, 2, 3, 4}}
	enc := p.Encode()
	assert.NotZero(t, enc[0]&FlagLengthIncluded, "L bit must be set when HasLength")
	assert.Equal(t, uint32(4), uint32(enc[1])<<24|uint32(enc[2])<<16|uint32(enc[3])<<8|uint32(enc[4]))
}

func TestReassembler_SingleFragment(t *testing.T) {
	r := NewReassembler(0)
	complete, err := r.Accept(&Packet{Data: []byte("hello")})
	require.NoError(t, err)
	assert.True(t, complete)
	assert.Equal(t, []byte("hello"), r.Buffer())
}

func TestReassembler_MultiFragment(t *testing.T) {
	r := NewReassembler(0)

	complete, err := r.Accept(&Packet{HasLength: true, TLSMessageLength: 9, Flags: FlagMoreFragments, Data: []byte("abc")})
	require.NoError(t, err)
	assert.False(t, complete)
	declared, ok := r.Declared()
	assert.True(t, ok)
	assert.Equal(t, uint32(9), declared)

	complete, err = r.Accept(&Packet{Flags: FlagMoreFragments, Data: []byte("def")})
	require.NoError(t, err)
	assert.False(t, complete)

	complete, err = r.Accept(&Packet{Data: []byte("ghi")})
	require.NoError(t, err)
	assert.True(t, complete)
	assert.Equal(t, []byte("abcdefghi"), r.Buffer())
}

func TestReassembler_LengthOnContinuation(t *testing.T) {
	r := NewReassembler(0)
	_, err := r.Accept(&Packet{Flags: FlagMoreFragments, Data: []byte("abc")})
	require.NoError(t, err)
	// A continuation fragment must not carry the L flag.
	_, err = r.Accept(&Packet{HasLength: true, TLSMessageLength: 6, Data: []byte("def")})
	assert.ErrorIs(t, err, ErrLengthOnContinuation)
}

func TestReassembler_DeclaredLengthExceeded(t *testing.T) {
	r := NewReassembler(0)
	_, err := r.Accept(&Packet{HasLength: true, TLSMessageLength: 4, Flags: FlagMoreFragments, Data: []byte("ab")})
	require.NoError(t, err)
	_, err = r.Accept(&Packet{Data: []byte("cdef")}) // 2+4 = 6 > 4
	assert.ErrorIs(t, err, ErrDeclaredLengthExceeded)
}

func TestReassembler_FinalLengthMismatch(t *testing.T) {
	r := NewReassembler(0)
	// Declares 5 but the final fragment delivers only 3 octets in total.
	_, err := r.Accept(&Packet{HasLength: true, TLSMessageLength: 5, Data: []byte("abc")})
	assert.ErrorIs(t, err, ErrFinalLengthMismatch)
}

func TestReassembler_MaxLengthDeclared(t *testing.T) {
	r := NewReassembler(8)
	_, err := r.Accept(&Packet{HasLength: true, TLSMessageLength: 100, Data: []byte("x")})
	assert.ErrorIs(t, err, ErrMaxLengthExceeded)
}

func TestReassembler_MaxLengthAccumulated(t *testing.T) {
	r := NewReassembler(4)
	// No declared length; cap enforced on accumulated data.
	_, err := r.Accept(&Packet{Flags: FlagMoreFragments, Data: []byte("abcde")})
	assert.ErrorIs(t, err, ErrMaxLengthExceeded)
}

func TestReassembler_Resume(t *testing.T) {
	// Simulate persisting and reloading reassembly state across EAP rounds.
	r1 := NewReassembler(0)
	_, err := r1.Accept(&Packet{HasLength: true, TLSMessageLength: 6, Flags: FlagMoreFragments, Data: []byte("ab")})
	require.NoError(t, err)
	declared, ok := r1.Declared()
	require.True(t, ok)

	r2 := LoadReassembler(r1.Buffer(), declared, ok, 0)
	complete, err := r2.Accept(&Packet{Data: []byte("cdef")})
	require.NoError(t, err)
	assert.True(t, complete)
	assert.Equal(t, []byte("abcdef"), r2.Buffer())
}

func TestFragment_SingleFragment(t *testing.T) {
	packets, err := Fragment([]byte("hello"), 100)
	require.NoError(t, err)
	require.Len(t, packets, 1)
	assert.True(t, packets[0].HasLength)
	assert.Equal(t, uint32(5), packets[0].TLSMessageLength)
	assert.False(t, packets[0].More())
	assert.Equal(t, []byte("hello"), packets[0].Data)
}

func TestFragment_MultipleFragments(t *testing.T) {
	data := []byte("abcdefghij") // 10 bytes, fragment size 4 -> 4 + 4 + 2
	packets, err := Fragment(data, 4)
	require.NoError(t, err)
	require.Len(t, packets, 3)

	assert.True(t, packets[0].HasLength)
	assert.Equal(t, uint32(10), packets[0].TLSMessageLength)
	assert.True(t, packets[0].More())
	assert.False(t, packets[1].HasLength, "only the first fragment carries the L flag")
	assert.True(t, packets[1].More())
	assert.False(t, packets[2].More(), "the last fragment clears the M flag")
}

func TestFragment_Empty(t *testing.T) {
	packets, err := Fragment(nil, 16)
	require.NoError(t, err)
	require.Len(t, packets, 1)
	assert.True(t, packets[0].HasLength)
	assert.Equal(t, uint32(0), packets[0].TLSMessageLength)
	assert.Empty(t, packets[0].Data)
}

// TestFragmentReassembleRoundTrip fragments a message and reassembles it back,
// exercising the full outbound -> inbound path against several sizes.
func TestFragmentReassembleRoundTrip(t *testing.T) {
	sizes := []int{1, 7, 64, 255, 4096}
	for _, n := range sizes {
		data := make([]byte, n)
		for i := range data {
			data[i] = byte(i)
		}
		packets, err := Fragment(data, 100)
		require.NoError(t, err)

		r := NewReassembler(0)
		var complete bool
		for _, p := range packets {
			// Round-trip each fragment through the wire encoding too.
			parsed, perr := Parse(p.Encode())
			require.NoError(t, perr)
			complete, err = r.Accept(parsed)
			require.NoError(t, err)
		}
		assert.True(t, complete, "size %d should reassemble fully", n)
		assert.True(t, bytes.Equal(data, r.Buffer()), "size %d round trip mismatch", n)
	}
}

func TestErrorsAreDistinct(t *testing.T) {
	all := []error{
		ErrEmptyPayload, ErrShortLength, ErrLengthOnContinuation,
		ErrDeclaredLengthExceeded, ErrMaxLengthExceeded, ErrFinalLengthMismatch,
	}
	for i := range all {
		for j := range all {
			if i != j {
				assert.False(t, errors.Is(all[i], all[j]))
			}
		}
	}
}
