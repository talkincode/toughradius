package iploc

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io/ioutil"
	"runtime"
)

const Version = "1.0"

type rangeIterator func(i int, start, end IP) bool

// Find shorthand for iploc.Open.Find
// not be preload file and without indexed
func Find(qqwrySrc, rawIP string) (*Detail, error) {
	loc, err := OpenWithoutIndexes(qqwrySrc)
	if err != nil {
		return nil, err
	}
	defer loc.parser.Close()
	return loc.Find(rawIP), nil
}

// Open 生成索引，查询速度快
func Open(qqwrySrc string) (loc *Locator, err error) {
	loc = &Locator{}
	var parser *Parser
	if parser, err = NewParser(qqwrySrc, true); err != nil {
		return nil, err
	}
	loc.indexes = newIndexes(parser)
	loc.count = int(parser.count)
	return
}

// OpenWithoutIndexes 无索引，不预载文件，打开速度快，但查询速度慢
func OpenWithoutIndexes(qqwrySrc string) (loc *Locator, err error) {
	loc = &Locator{}
	if loc.parser, err = NewParser(qqwrySrc, false); err != nil {
		return nil, err
	}
	loc.count = int(loc.parser.count)
	return
}

func Load(b []byte) (loc *Locator, err error) {
	buf := bytes.NewReader(b)
	r, err := zlib.NewReader(buf)
	if err != nil {
		return nil, err
	}
	b, err = ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	loc = &Locator{}
	res := &resource{data: b}
	var parser *Parser
	if parser, err = NewParserRes(res, uint32(len(res.data))); err != nil {
		return nil, err
	}
	loc.indexes = newIndexes(parser)
	loc.count = int(parser.count)
	return
}

func LoadWithoutIndexes(b []byte) (loc *Locator, err error) {
	buf := bytes.NewReader(b)
	r, err := zlib.NewReader(buf)
	if err != nil {
		return nil, err
	}
	b, err = ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	loc = &Locator{}
	res := &resource{data:b}
	if loc.parser, err = NewParserRes(res, uint32(len(res.data))); err != nil {
		return nil, err
	}
	loc.count = int(loc.parser.count)
	return
}

type Locator struct {
	parser  *Parser
	indexes *indexes
	count   int
}

// Close close the file descriptor, if there is no preload
func (loc *Locator) Close() error {
	if loc.parser != nil {
		return loc.parser.Close()
	}
	return nil
}

func (loc *Locator) Count() int {
	return loc.count
}

func (loc *Locator) FindIP(ip IP) *Detail {
	return loc.find(ip)
}

func (loc *Locator) FindUint(uintIP uint32) *Detail {
	return loc.find(ParseUintIP(uintIP))
}

func (loc *Locator) Find(rawIP string) *Detail {
	ip, err := ParseIP(rawIP)
	if err != nil {
		return nil
	}
	return loc.find(ip)
}

func (loc *Locator) Range(iterator rangeIterator) {
	if loc.indexes != nil {
		var n int
		loc.indexes.index.AscendRange(nil, nil, func(i dataItem) bool {
			n++
			return iterator(n, ParseUintIP(i.(indexItem)[0]), ParseUintIP(i.(indexItem)[1]))
		})
	} else {
		loc.parser.IndexRange(func(i int, start, end, pos uint32) bool {
			return iterator(i, ParseUintIP(start), ParseUintIP(end))
		})
	}
}

func (loc *Locator) find(ip IP) *Detail {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 1<<16)
			buf = buf[:runtime.Stack(buf, false)]
			fmt.Printf("iploc > panic: %v\n%s", r, buf)
		}
	}()

	hit := loc.seek(ip)
	detail := &Detail{
		IP:       ip,
		Start:    ParseUintIP(hit[0]),
		End:      ParseUintIP(hit[1]),
		Location: loc.getLocation(hit),
	}
	return detail.fill()
}

func (loc *Locator) seek(ip IP) (hit indexItem) {
	if loc.indexes != nil {
		hit = loc.indexes.indexOf(ip.Uint())
	} else {
		var low, mid, high uint32
		var index int64
		var start, end IP
		high = loc.parser.Count() - 1
		for low <= high {
			mid = (low + high) >> 1
			index = int64(loc.parser.min + mid*indexBlockSize)
			start = ParseBytesIP(loc.parser.ReadBytes(index, ipByteSize))
			if ip.Compare(start) < 0 {
				high = mid - 1
			} else {
				index = loc.parser.ReadPosition(index + ipByteSize)
				end = ParseBytesIP(loc.parser.ReadBytes(index, ipByteSize))
				if ip.Compare(end) > 0 {
					low = mid + 1
				} else {
					hit = indexItem{start.Uint(), end.Uint(), uint32(index), 0}
					break
				}
			}
		}
	}
	return
}

func (loc *Locator) getLocation(item indexItem) Location {
	if loc.indexes != nil {
		return loc.indexes.getLocation(item[2], item[3])
	}
	return loc.parser.digLocation(int64(item[2]))
}
