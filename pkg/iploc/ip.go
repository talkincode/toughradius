package iploc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
)

const (
	IPLen = 4
)

type IP [IPLen]byte

func (ip IP) Bytes() []byte {
	return ip[:]
}

func (ip IP) ReverseBytes() []byte {
	var b [4]byte
	copy(b[:], ip[:])
	for i, j := 0, len(b)-1; i < len(b)/2; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
	return b[:]
}

func (ip IP) String() string {
	return fmt.Sprintf("%d.%d.%d.%d", ip[0], ip[1], ip[2], ip[3])
}

func (ip IP) Uint() uint32 {
	return binary.BigEndian.Uint32(ip[:])
}

// Compare like bytes.Compare
// The result will be 0 if == a, -1 if < a, and +1 if > a.
func (ip IP) Compare(a IP) int {
	return bytes.Compare(ip[:], a[:])
}

func ParseIP(s string) (ip IP, err error) {
	var b = make([]byte, 0, IPLen)
	var d uint64
	for i, v := range strings.Split(s, ".") {
		if d, err = strconv.ParseUint(v, 10, 8); err != nil {
			err = fmt.Errorf("invalid IP address %s", s)
			return
		}
		b = append(b, byte(d))
		if i == 3 {
			break
		}
	}
	if len(b) == 0 {
		err = fmt.Errorf("invalid IP address %s", s)
		return
	}

	// filling，e.g. 127.1 -> 127.0.0.1
	// copy to array, right padding
	copy(ip[:], b)
	if padding := IPLen - len(b); padding > 0 {
		// left padding
		if lastIndex := len(b) - 1; b[lastIndex] > 0 {
			ip[lastIndex], ip[3] = 0, ip[lastIndex]
		}
	}
	return
}

func ParseBytesIP(b []byte) (ip IP) {
	copy(ip[:], b)
	return
}

func ParseUintIP(u uint32) (ip IP) {
	binary.BigEndian.PutUint32(ip[:], u)
	return
}
