package iploc

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

const (
	indexBlockSize   = 7
	ipByteSize       = 4
	positionByteSize = 3

	terminatorFlag = 0x00
	redirectAll    = 0x01
	redirectPart   = 0x02
)

/*
	<qqwry.dat>
	Original download URL: http://www.cz88.net/fox/ipdat.shtml
	Original qqwry.dat, byte order: LittleEndian, encoding: GBK
	Use the iploc-conv tool to convert to UTF-8
	File layout, all offsets are absolute 3-byte values
	<Header 8 bytes|Data section|Index section multiple of 7 bytes>
	[Header]	Index start position (4 bytes)|Index end position (4 bytes)
	[Data section]	End IP (4 bytes)|Country and region data
			*Country and region data
				Typically country data (ends with 0x00)|Region data (ends with 0x00)
				When the first byte of country/region data is 0x01 or 0x02 it indicates redirect mode, followed by a 3-byte offset

				0x01	Both country and region redirected, offset (3 bytes)
				0x02	Country redirected, offset (3 bytes)|Region data
				No redirect	Country data (ends with 0x00)|Region data

				Region data might contain another redirect (starts with 0x02)

	[Index section] Start IP (4 bytes)|Offset (3 bytes) = Each index entry is 7 bytes
*/

type Parser struct {
	res   resReadCloser
	min   uint32
	max   uint32
	total uint32
	count uint32
	size  uint32
}

type indexIterator func(i int, start, end, pos uint32) bool

type LocationRaw struct {
	Text [2][]byte
	Pos  [2]uint32
	Mode [2]byte
}

func NewParser(qqwrySrc string, preload bool) (*Parser, error) {
	var (
		err  error
		size uint32
		fd   *os.File
		b    []byte
		res  resReadCloser
	)

	if preload {
		if b, err = ioutil.ReadFile(qqwrySrc); err != nil {
			return nil, err
		}
		size = uint32(len(b))
		res = &resource{data: b}
	} else {
		if fd, err = os.OpenFile(qqwrySrc, os.O_RDONLY, 0400); err != nil {
			return nil, err
		}
		fi, err := fd.Stat()
		if err != nil {
			return nil, err
		}
		size = uint32(fi.Size())
		res = fd
	}
	return NewParserRes(res, size)
}

func NewParserRes(res resReadCloser, size uint32) (*Parser, error) {
	if res == nil {
		return nil, fmt.Errorf("nil resource")
	}
	var (
		p             = &Parser{res: res}
		b             []byte
		n             int
		err           error
		errInvalidDat = fmt.Errorf("invalid IP dat file")
	)
	b = make([]byte, ipByteSize*2)
	if n, err = p.res.ReadAt(b, 0); err != nil || n != ipByteSize*2 {
		return nil, errInvalidDat
	}

	p.min = binary.LittleEndian.Uint32(b[:ipByteSize])
	p.max = binary.LittleEndian.Uint32(b[ipByteSize:])
	if (p.max-p.min)%indexBlockSize != 0 || size != p.max+indexBlockSize {
		return nil, errInvalidDat
	}
	p.total = (p.max - p.min) / indexBlockSize
	p.count = (p.max-p.min)/indexBlockSize + 1
	p.size = size
	return p, nil
}

func (p *Parser) Close() error {
	return p.res.Close()
}

func (p *Parser) Count() uint32 {
	return p.count
}

func (p *Parser) Size() uint32 {
	return p.size
}

func (p *Parser) Reader() io.Reader {
	return p.res
}

// (*Parser) ReadByte Read1byte(s)，to identify redirect mode
func (p *Parser) ReadByte(pos int64) byte {
	b := make([]byte, 1)
	n, err := p.res.ReadAt(b, pos)
	if err != nil || n != 1 {
		panic("ReadByte damaged DAT files, position: " + fmt.Sprint(pos))
	}
	return b[0]
}

// (*Parser) ReadBytes Readnbyte(s)and reverse
func (p *Parser) ReadBytes(pos, n int64) (b []byte) {
	b = make([]byte, n)
	i, err := p.res.ReadAt(b, pos)
	if err != nil || int64(i) != n {
		panic("ReadBytes damaged DAT files, position: " + fmt.Sprint(pos))
	}
	// reverse bytes
	for i, j := 0, len(b)-1; i < len(b)/2; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
	return
}

// (*Parser) ReadPosition Read3byte(s)offset position
func (p *Parser) ReadPosition(offset int64) int64 {
	b := p.ReadBytes(offset, positionByteSize)
	// left padding to the 32 bits
	if i := 4 - len(b); i > 0 {
		b = append(make([]byte, i), b...)
	}
	// bytes has been reversed, so it won't use binary.LittleEndian
	return int64(binary.BigEndian.Uint32(b))
}

// (*Parser) ReadText Read country/region data，ending with0x00terminator
func (p *Parser) ReadText(offset int64) ([]byte, int) {
	if uint32(offset) >= p.min {
		return nil, 0
	}
	var s []byte
	var b byte
	for {
		b = p.ReadByte(offset)
		if b != terminatorFlag {
			s = append(s, b)
		} else if len(s) > 0 {
			break
		}
		offset++
	}
	return s, len(s)
}

func (p *Parser) ReadString(offset int64) (string, int) {
	s, n := p.ReadText(offset)
	return string(s), n
}

// (*Parser) ReadRegion Read region data，handle possible redirect
func (p *Parser) ReadRegion(offset int64) (s []byte) {
	switch p.ReadByte(offset) {
	case redirectPart:
		s, _ = p.ReadText(p.ReadPosition(offset + 1))
	default:
		s, _ = p.ReadText(offset)
	}
	return
}

func (p *Parser) ReadRegionString(offset int64) string {
	return string(p.ReadRegion(offset))
}

func (p *Parser) digLocation(offset int64) (location Location) {
	var n int
	switch p.ReadByte(offset + ipByteSize) {
	case redirectAll:
		offset = p.ReadPosition(offset + ipByteSize + 1)
		switch p.ReadByte(offset) {
		case redirectPart:
			location.Country, _ = p.ReadString(p.ReadPosition(offset + 1))
			location.Region = p.ReadRegionString(offset + 1 + positionByteSize)
		default:
			location.Country, n = p.ReadString(offset)
			// +1, skip 1 bytes 0x00
			location.Region = p.ReadRegionString(offset + 1 + int64(n))
		}
	case redirectPart:
		location.Country, _ = p.ReadString(p.ReadPosition(offset + ipByteSize + 1))
		location.Region = p.ReadRegionString(offset + ipByteSize + 1 + positionByteSize)
	default:
		location.Country, n = p.ReadString(offset + ipByteSize)
		// +1, skip 1 bytes 0x00
		location.Region = p.ReadRegionString(offset + ipByteSize + 1 + int64(n))
	}
	location.raw = location.Country
	if location.Region != "" {
		location.raw += " " + location.Region
	}
	return
}

func (p *Parser) readRegionRaw(offset int64) (s []byte, pos uint32, mode byte) {
	switch p.ReadByte(offset) {
	case redirectPart:
		pos = uint32(p.ReadPosition(offset + 1))
		mode = redirectPart
	default:
		s, _ = p.ReadText(offset)
	}
	return
}

// ReadLocationRaw Used for export or indexing
func (p *Parser) ReadLocationRaw(offset int64) (raw LocationRaw) {
	var n int
	raw.Mode[0] = p.ReadByte(offset + ipByteSize)
	switch raw.Mode[0] {
	case redirectAll:
		offset = p.ReadPosition(offset + ipByteSize + 1)
		switch p.ReadByte(offset) {
		case redirectPart:
			raw.Mode[0] = redirectPart
			raw.Pos[0] = uint32(p.ReadPosition(offset + 1))
			raw.Text[1], raw.Pos[1], raw.Mode[1] = p.readRegionRaw(offset + 1 + positionByteSize)
			if raw.Text[1] != nil {
				raw.Pos[1] = uint32(offset + 1 + positionByteSize)
			}
		default:
			raw.Pos[0] = uint32(offset)
			_, n = p.ReadText(offset)
			_, raw.Pos[1], _ = p.readRegionRaw(offset + 1 + int64(n))
			if raw.Pos[1] == 0 {
				raw.Pos[1] = uint32(offset + 1 + int64(n))
			}
		}
	case redirectPart:
		raw.Pos[0] = uint32(p.ReadPosition(offset + ipByteSize + 1))
		raw.Text[1], raw.Pos[1], raw.Mode[1] = p.readRegionRaw(offset + ipByteSize + 1 + positionByteSize)
		if raw.Text[1] != nil {
			raw.Pos[1] = uint32(offset + ipByteSize + 1 + positionByteSize)
		}
	default:
		raw.Pos[0] = uint32(offset + ipByteSize)
		raw.Mode[0] = 0x00
		raw.Text[0], n = p.ReadText(offset + ipByteSize)
		raw.Text[1], raw.Pos[1], raw.Mode[1] = p.readRegionRaw(offset + ipByteSize + 1 + int64(n))
		if raw.Text[1] != nil {
			raw.Pos[1] = uint32(offset + ipByteSize + 1 + int64(n))
		}
	}
	return
}

// (*Parser) IndexRange
// calls the iterator for every index within the range (i, start, end, Pos)
// until iterator returns false.
func (p *Parser) IndexRange(iterator indexIterator) {
	var (
		count      = int(p.count)
		index, pos int64
	)
	for i := 0; i < count; i++ {
		index = int64(p.min) + indexBlockSize*int64(i)
		pos = p.ReadPosition(index + 4)
		if !iterator(
			i+1,
			ParseBytesIP(p.ReadBytes(index, ipByteSize)).Uint(),
			ParseBytesIP(p.ReadBytes(pos, ipByteSize)).Uint(),
			uint32(pos),
		) {
			break
		}
	}
}
