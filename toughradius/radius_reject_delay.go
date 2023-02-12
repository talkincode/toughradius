package toughradius

import (
	"sync"
	"sync/atomic"
	"time"
)

type RejectItem struct {
	Rejects    int64
	LastReject time.Time
	Lock       sync.RWMutex
}

func (ri *RejectItem) Incr() {
	ri.Lock.Lock()
	atomic.AddInt64(&ri.Rejects, 1)
	ri.LastReject = time.Now()
	ri.Lock.Unlock()
}

func (ri *RejectItem) IsOver(max int64) bool {
	ri.Lock.RLock()
	defer ri.Lock.RUnlock()
	if time.Since(ri.LastReject).Seconds() > 10 {
		atomic.StoreInt64(&ri.Rejects, 0)
		return false
	}
	return ri.Rejects > max && time.Since(ri.LastReject).Seconds() < 10
}

type RejectCache struct {
	Items map[string]*RejectItem
	Lock  sync.Mutex
}

func (rc *RejectCache) GetItem(username string) *RejectItem {
	rc.Lock.Lock()
	defer rc.Lock.Unlock()
	if len(rc.Items) >= 65535 {
		rc.Items = make(map[string]*RejectItem, 0)
		return nil
	}
	return rc.Items[username]
}

func (rc *RejectCache) SetItem(username string) {
	rc.Lock.Lock()
	_, ok := rc.Items[username]
	if !ok {
		rc.Items[username] = &RejectItem{
			Rejects:    1,
			LastReject: time.Now(),
			Lock:       sync.RWMutex{},
		}
	}
	rc.Lock.Unlock()
}
