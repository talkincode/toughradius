package radiusd

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewAuthRateLimiterRoundsToPowerOfTwo(t *testing.T) {
	cases := map[int]int{1: 1, 2: 2, 3: 4, 100: 128, 256: 256, 257: 512}
	for in, want := range cases {
		l := newAuthRateLimiter(in)
		if len(l.shards) != want {
			t.Errorf("newAuthRateLimiter(%d): got %d shards, want %d", in, len(l.shards), want)
		}
		if int(l.mask) != want-1 {
			t.Errorf("newAuthRateLimiter(%d): got mask %d, want %d", in, l.mask, want-1)
		}
	}
}

func TestAuthRateLimiterCheckReleaseGet(t *testing.T) {
	l := newAuthRateLimiter(defaultAuthRateShards)

	if err := l.check("u1", time.Second); err != nil {
		t.Fatalf("first check should pass: %v", err)
	}
	if err := l.check("u1", time.Second); err == nil {
		t.Fatal("second check within window should be limited")
	}
	if _, ok := l.get("u1"); !ok {
		t.Fatal("u1 should be present after check")
	}
	l.release("u1")
	if _, ok := l.get("u1"); ok {
		t.Fatal("u1 should be absent after release")
	}
	if err := l.check("u1", time.Second); err != nil {
		t.Fatalf("check after release should pass: %v", err)
	}
}

func TestAuthRateLimiterDistinctUsersAndLen(t *testing.T) {
	l := newAuthRateLimiter(defaultAuthRateShards)
	const n = 500
	for i := 0; i < n; i++ {
		if err := l.check(fmt.Sprintf("user-%d", i), time.Minute); err != nil {
			t.Fatalf("distinct user %d should pass: %v", i, err)
		}
	}
	if got := l.len(); got != n {
		t.Fatalf("len: got %d, want %d", got, n)
	}
}

func TestAuthRateLimiterExpiry(t *testing.T) {
	l := newAuthRateLimiter(defaultAuthRateShards)
	if err := l.check("u1", 50*time.Millisecond); err != nil {
		t.Fatalf("first check should pass: %v", err)
	}
	time.Sleep(80 * time.Millisecond)
	if err := l.check("u1", 50*time.Millisecond); err != nil {
		t.Fatalf("check after expiry should pass: %v", err)
	}
}

// globalRateLimiter is a single-mutex baseline used only to demonstrate, via
// benchmarks, the contention the sharded limiter removes. It mirrors the
// previous CheckAuthRateLimit/ReleaseAuthRateLimit implementation.
type globalRateLimiter struct {
	mu    sync.Mutex
	users map[string]AuthRateUser
}

func newGlobalRateLimiter() *globalRateLimiter {
	return &globalRateLimiter{users: make(map[string]AuthRateUser)}
}

func (g *globalRateLimiter) check(username string, interval time.Duration) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if val, ok := g.users[username]; ok {
		if time.Now().Before(val.Starttime.Add(interval)) {
			return errBenchLimited
		}
	}
	g.users[username] = AuthRateUser{Username: username, Starttime: time.Now()}
	return nil
}

func (g *globalRateLimiter) release(username string) {
	g.mu.Lock()
	delete(g.users, username)
	g.mu.Unlock()
}

var errBenchLimited = fmt.Errorf("limited")

// BenchmarkAuthRateLimiter_Sharded measures the check+release cycle for many
// distinct users running in parallel (the high-concurrency scenario from the
// issue). Compare against BenchmarkAuthRateLimiter_Global to see the win.
func BenchmarkAuthRateLimiter_Sharded(b *testing.B) {
	l := newAuthRateLimiter(defaultAuthRateShards)
	var seq uint64
	b.RunParallel(func(pb *testing.PB) {
		prefix := atomic.AddUint64(&seq, 1)
		var i uint64
		for pb.Next() {
			i++
			u := fmt.Sprintf("u-%d-%d", prefix, i)
			_ = l.check(u, time.Second)
			l.release(u)
		}
	})
}

func BenchmarkAuthRateLimiter_Global(b *testing.B) {
	g := newGlobalRateLimiter()
	var seq uint64
	b.RunParallel(func(pb *testing.PB) {
		prefix := atomic.AddUint64(&seq, 1)
		var i uint64
		for pb.Next() {
			i++
			u := fmt.Sprintf("u-%d-%d", prefix, i)
			_ = g.check(u, time.Second)
			g.release(u)
		}
	})
}
