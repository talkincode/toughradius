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
	defer ri.Lock.Unlock()
	atomic.AddInt64(&ri.Rejects, 1)
	ri.LastReject = time.Now()
}

func (ri *RejectItem) IsOver(max int64) bool {
	ri.Lock.RLock()
	defer ri.Lock.RUnlock()
	if time.Since(ri.LastReject).Seconds() > 10 {
		ri.Lock.RUnlock()
		ri.Lock.Lock()
		defer ri.Lock.Unlock()
		if time.Since(ri.LastReject).Seconds() > 10 {
			atomic.StoreInt64(&ri.Rejects, 0)
		}
		return false
	}
	return atomic.LoadInt64(&ri.Rejects) > max
}

type RejectCache struct {
	Items map[string]*RejectItem
	Lock  sync.RWMutex
}

func (rc *RejectCache) GetItem(username string) *RejectItem {
	rc.Lock.RLock()
	defer rc.Lock.RUnlock()
	if len(rc.Items) >= 65535 {
		rc.Lock.RUnlock()
		rc.Lock.Lock()
		defer rc.Lock.Unlock()
		if len(rc.Items) >= 65535 {
			rc.Items = make(map[string]*RejectItem, 0)
		}
		return nil
	}
	return rc.Items[username]
}

func (rc *RejectCache) SetItem(username string) {
	rc.Lock.Lock()
	defer rc.Lock.Unlock()
	if _, ok := rc.Items[username]; !ok {
		rc.Items[username] = &RejectItem{
			Rejects:    1,
			LastReject: time.Now(),
		}
	}
}
