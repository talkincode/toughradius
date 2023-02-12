package iploc

import (
	"encoding/gob"

	"github.com/google/btree"
)

type dataItem = btree.Item

// 索引时
// 0: ip start
// 1: ip end
// 2: country position
// 3: region position
// 无索引时
// 2: ip end position
// 3: n/a
type indexItem [4]uint32

func (i indexItem) Less(than btree.Item) bool {
	// 默认降序 1:End < 0:Start
	if v, ok := than.(indexItem); ok {
		return i[1] < v[0]

	}
	return i[1] < than.(indexItemAscend)[0]
}

// indexItem 升序
type indexItemAscend [4]uint32

func (i indexItemAscend) Less(than btree.Item) bool {
	// 升序 0:End < 0:Start
	if v, ok := than.(indexItem); ok {
		return i[0] < v[0]

	}
	return i[0] < than.(indexItemAscend)[0]
}

type indexes struct {
	index     *btree.BTree
	indexMid  indexItem
	locations map[uint32][]byte
}

func (idx *indexes) indexOf(u uint32) (hit indexItem) {
	// 对比中间值决定高低顺序，提升查询速度
	if u > idx.indexMid[1] {
		idx.index.DescendLessOrEqual(indexItem{0, u}, func(i btree.Item) bool {
			hit = i.(indexItem)
			return false
		})
	} else if u < idx.indexMid[0] {
		idx.index.AscendGreaterOrEqual(indexItemAscend{u}, func(i btree.Item) bool {
			hit = i.(indexItem)
			return false
		})
	} else {
		hit = idx.indexMid
	}
	return
}

func (idx *indexes) getLocation(i, j uint32) Location {
	return parseLocation(idx.locations[i], idx.locations[j])
}

func newIndexes(p *Parser) *indexes {
	idx := &indexes{
		index: btree.New(10),
	}
	idx.locations = make(map[uint32][]byte)

	var (
		item indexItem
		raw  LocationRaw
		mid  = int(p.Count()) >> 1
		has  bool
	)

	p.IndexRange(func(i int, start, end, pos uint32) bool {
		item = indexItem{start, end, pos}
		raw = p.ReadLocationRaw(int64(pos))
		if raw.Text[0] != nil {
			if _, has = idx.locations[raw.Pos[0]]; !has {
				idx.locations[raw.Pos[0]] = raw.Text[0]
			}
		}
		if raw.Text[1] != nil {
			if _, has = idx.locations[raw.Pos[1]]; !has {
				idx.locations[raw.Pos[1]] = raw.Text[1]
			}
		}
		item[2] = raw.Pos[0]
		item[3] = raw.Pos[1]
		if i == mid {
			idx.indexMid = item
		}
		idx.index.ReplaceOrInsert(item)
		return true
	})
	return idx
}

func init() {
	gob.Register([][4]uint32{})
	gob.Register(map[uint32][]byte{})
}
