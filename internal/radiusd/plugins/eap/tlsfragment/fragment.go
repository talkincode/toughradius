// Package tlsfragment implements EAP-TLS fragmentation and reassembly as
// defined in RFC 5216 §2.1.5 (EAP-TLS) and the general RADIUS fragmentation
// considerations in RFC 7499.
//
// An EAP-TLS payload (the EAP data field that follows the EAP Type octet) is:
//
//	0                   1                   2                   3
//	0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	|     Flags     |      TLS Message Length (optional, 4 octets) ...
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	|              TLS Data (optional) ...
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//
// Flags (RFC 5216 §3.1):
//
//	0 1 2 3 4 5 6 7
//	+-+-+-+-+-+-+-+-+
//	|L M S R R R R R|
//	+-+-+-+-+-+-+-+-+
//
//	L = Length included, M = More fragments, S = EAP-TLS Start, R = Reserved.
//
// The L flag indicates the presence of the four-octet TLS Message Length field,
// and MUST be set for the first fragment of a fragmented TLS message. The M flag
// is set on all but the last fragment. The S flag is only set within the
// EAP-TLS Start sent by the server.
package tlsfragment

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// EAP-TLS flag bits (RFC 5216 §3.1).
const (
	FlagLengthIncluded = 0x80 // L: TLS Message Length field present
	FlagMoreFragments  = 0x40 // M: more fragments follow
	FlagStart          = 0x20 // S: EAP-TLS Start
)

// DefaultMaxMessageLength is the default upper bound on the size of a single
// reassembled TLS message (or group of messages). RFC 5216 §2.1.5 notes that,
// to protect against reassembly lockup and denial-of-service attacks, an
// implementation may set a maximum size, and that 64 KB is a reasonable choice
// since a single certificate is rarely longer than a few thousand octets.
const DefaultMaxMessageLength = 64 * 1024

// Errors returned while parsing and reassembling EAP-TLS fragments.
var (
	// ErrEmptyPayload is returned when an EAP-TLS payload has no flags octet.
	ErrEmptyPayload = errors.New("eap-tls: empty payload (missing flags octet)")
	// ErrShortLength is returned when the L flag is set but the four-octet TLS
	// Message Length field is truncated.
	ErrShortLength = errors.New("eap-tls: length flag set but TLS Message Length field is truncated")
	// ErrLengthOnContinuation is returned when the L flag is set on a fragment
	// that is not the first one of a message (RFC 5216 §3.1: L MUST be set only
	// for the first fragment).
	ErrLengthOnContinuation = errors.New("eap-tls: length flag set on a continuation fragment")
	// ErrDeclaredLengthExceeded is returned when the reassembled data grows
	// beyond the TLS Message Length advertised in the first fragment.
	ErrDeclaredLengthExceeded = errors.New("eap-tls: reassembled data exceeds declared TLS Message Length")
	// ErrMaxLengthExceeded is returned when a declared or reassembled message
	// exceeds the configured maximum, guarding against reassembly DoS.
	ErrMaxLengthExceeded = errors.New("eap-tls: TLS message exceeds maximum allowed length")
	// ErrFinalLengthMismatch is returned when the last fragment arrives but the
	// total reassembled length does not match the declared TLS Message Length.
	ErrFinalLengthMismatch = errors.New("eap-tls: reassembled length does not match declared TLS Message Length")
)

// Packet is a parsed EAP-TLS payload (the EAP data field after the Type octet).
type Packet struct {
	Flags            uint8  // raw flags octet
	HasLength        bool   // whether the L flag was set and a length parsed
	TLSMessageLength uint32 // total message length advertised by the first fragment (valid when HasLength)
	Data             []byte // TLS record bytes carried by this fragment (may be empty for ACKs/Start)
}

// More reports whether the M (More fragments) flag is set.
func (p *Packet) More() bool { return p.Flags&FlagMoreFragments != 0 }

// Start reports whether the S (EAP-TLS Start) flag is set.
func (p *Packet) Start() bool { return p.Flags&FlagStart != 0 }

// IsACK reports whether the packet is a bare fragment acknowledgement: a
// payload carrying only a flags octet with no TLS data and none of the L/M/S
// flags set (RFC 5216 §2.1.5 fragment ACK).
func (p *Packet) IsACK() bool {
	return len(p.Data) == 0 && p.Flags&(FlagLengthIncluded|FlagMoreFragments|FlagStart) == 0
}

// Parse decodes an EAP-TLS payload (EAPMessage.Data) into a Packet. It validates
// the flag/length framing but does not interpret the TLS records themselves.
func Parse(payload []byte) (*Packet, error) {
	if len(payload) < 1 {
		return nil, ErrEmptyPayload
	}

	p := &Packet{Flags: payload[0]}
	rest := payload[1:]

	if p.Flags&FlagLengthIncluded != 0 {
		if len(rest) < 4 {
			return nil, ErrShortLength
		}
		p.HasLength = true
		p.TLSMessageLength = binary.BigEndian.Uint32(rest[:4])
		rest = rest[4:]
	}

	if len(rest) > 0 {
		p.Data = rest
	}
	return p, nil
}

// Encode serializes the Packet back into an EAP-TLS payload (EAPMessage.Data).
// The L flag and TLS Message Length field are emitted when HasLength is set.
func (p *Packet) Encode() []byte {
	flags := p.Flags
	size := 1
	if p.HasLength {
		flags |= FlagLengthIncluded
		size += 4
	} else {
		flags &^= FlagLengthIncluded
	}
	size += len(p.Data)

	buf := make([]byte, size)
	buf[0] = flags
	off := 1
	if p.HasLength {
		binary.BigEndian.PutUint32(buf[off:off+4], p.TLSMessageLength)
		off += 4
	}
	copy(buf[off:], p.Data)
	return buf
}

// Reassembler accumulates inbound EAP-TLS fragments into a complete TLS message.
// The zero value is not ready for use; create one with NewReassembler.
type Reassembler struct {
	buf       []byte
	declared  uint32
	hasLength bool
	maxLength int
}

// NewReassembler creates a Reassembler. A non-positive maxLength selects
// DefaultMaxMessageLength.
func NewReassembler(maxLength int) *Reassembler {
	if maxLength <= 0 {
		maxLength = DefaultMaxMessageLength
	}
	return &Reassembler{maxLength: maxLength}
}

// LoadReassembler reconstructs a Reassembler from previously persisted state so
// that reassembly can continue across EAP rounds. buf is the data accumulated so
// far; declared is the advertised TLS Message Length (0 when unknown); hasLength
// reports whether declared was advertised.
func LoadReassembler(buf []byte, declared uint32, hasLength bool, maxLength int) *Reassembler {
	r := NewReassembler(maxLength)
	r.buf = buf
	r.declared = declared
	r.hasLength = hasLength
	return r
}

// Buffer returns the data accumulated so far. The slice is owned by the
// Reassembler and must not be modified by the caller.
func (r *Reassembler) Buffer() []byte { return r.buf }

// Declared returns the advertised TLS Message Length and whether it is known.
func (r *Reassembler) Declared() (uint32, bool) { return r.declared, r.hasLength }

// Accept folds the fragment p into the reassembly buffer. It returns complete=true
// once the final fragment (M flag clear) has been accepted, at which point
// Buffer holds the full TLS message.
func (r *Reassembler) Accept(p *Packet) (complete bool, err error) {
	first := len(r.buf) == 0 && !r.hasLength

	if p.HasLength {
		if !first {
			return false, ErrLengthOnContinuation
		}
		if int64(p.TLSMessageLength) > int64(r.maxLength) {
			return false, fmt.Errorf("%w: declared %d > max %d", ErrMaxLengthExceeded, p.TLSMessageLength, r.maxLength)
		}
		r.declared = p.TLSMessageLength
		r.hasLength = true
	}

	if len(p.Data) > 0 {
		if len(r.buf)+len(p.Data) > r.maxLength {
			return false, fmt.Errorf("%w: accumulated %d > max %d", ErrMaxLengthExceeded, len(r.buf)+len(p.Data), r.maxLength)
		}
		if r.hasLength && uint64(len(r.buf)+len(p.Data)) > uint64(r.declared) {
			return false, ErrDeclaredLengthExceeded
		}
		r.buf = append(r.buf, p.Data...)
	}

	if p.More() {
		return false, nil
	}

	if r.hasLength && uint64(len(r.buf)) != uint64(r.declared) {
		return false, fmt.Errorf("%w: reassembled %d != declared %d", ErrFinalLengthMismatch, len(r.buf), r.declared)
	}
	return true, nil
}

// Fragment splits a complete TLS message into ordered EAP-TLS fragments, each
// carrying at most maxFragment octets of TLS data. The first fragment sets the L
// flag and the TLS Message Length field; every fragment except the last sets the
// M flag (RFC 5216 §2.1.5). A non-positive maxFragment selects DefaultMaxMessageLength.
//
// An empty tlsData yields a single fragment with the L flag and a zero TLS
// Message Length, matching the framing peers expect for an empty handshake
// payload.
func Fragment(tlsData []byte, maxFragment int) ([]*Packet, error) {
	if maxFragment <= 0 {
		maxFragment = DefaultMaxMessageLength
	}
	total := len(tlsData)
	if int64(total) > int64(^uint32(0)) {
		return nil, fmt.Errorf("%w: %d octets", ErrMaxLengthExceeded, total)
	}

	if total == 0 {
		return []*Packet{{HasLength: true, TLSMessageLength: 0}}, nil
	}

	var packets []*Packet
	for off := 0; off < total; off += maxFragment {
		end := off + maxFragment
		if end > total {
			end = total
		}
		chunk := make([]byte, end-off)
		copy(chunk, tlsData[off:end])

		p := &Packet{Data: chunk}
		if off == 0 {
			p.HasLength = true
			p.TLSMessageLength = uint32(total) //nolint:gosec // G115: bounded above by uint32 max
		}
		if end < total {
			p.Flags |= FlagMoreFragments
		}
		packets = append(packets, p)
	}
	return packets, nil
}
