package radiusd

import (
	"sync"
	"time"

	radiuserrors "github.com/talkincode/toughradius/v9/internal/radiusd/errors"
)

// defaultAuthRateShards is the number of shards used by the authentication
// rate limiter. It must be a power of two so the shard index can be computed
// with a bitmask. A higher count lowers lock contention at a negligible memory
// cost (one mutex + one map header per shard).
const defaultAuthRateShards = 256

// maxAuthRateShards caps the shard count so the power-of-two rounding stays
// within uint32 range and avoids pathologically large allocations.
const maxAuthRateShards = 1 << 16

// authRateShard is a single partition of the rate-limit state, guarded by its
// own mutex so that requests for usernames hashing to different shards never
// contend on the same lock.
type authRateShard struct {
	mu    sync.Mutex
	users map[string]AuthRateUser
}

// authRateLimiter implements per-user authentication rate limiting backed by a
// fixed set of independently locked shards. Replacing the previous single
// global mutex removes the serialization point that capped authentication
// throughput on multi-core machines: distinct users now contend only when they
// happen to hash to the same shard.
type authRateLimiter struct {
	shards []*authRateShard
	mask   uint32
}

// newAuthRateLimiter creates a rate limiter with the given number of shards.
// shardCount is rounded up to the next power of two (minimum 1, capped at
// maxAuthRateShards).
func newAuthRateLimiter(shardCount int) *authRateLimiter {
	if shardCount < 1 {
		shardCount = 1
	}
	n := 1
	for n < shardCount && n < maxAuthRateShards {
		n <<= 1
	}
	shards := make([]*authRateShard, n)
	for i := range shards {
		shards[i] = &authRateShard{users: make(map[string]AuthRateUser)}
	}
	// n is a power of two in [1, maxAuthRateShards], so n-1 fits in uint32.
	return &authRateLimiter{shards: shards, mask: uint32(n - 1)} //nolint:gosec // G115: n is bounded by maxAuthRateShards
}

// fnv1a computes a 32-bit FNV-1a hash without allocating, suitable for the hot
// authentication path.
func fnv1a(s string) uint32 {
	const (
		offset32 = 2166136261
		prime32  = 16777619
	)
	h := uint32(offset32)
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= prime32
	}
	return h
}

func (l *authRateLimiter) shardFor(username string) *authRateShard {
	return l.shards[fnv1a(username)&l.mask]
}

// check records an authentication attempt for username. It returns a
// rate-limit error if a previous attempt is still within interval; otherwise it
// stores the new attempt timestamp and returns nil.
func (l *authRateLimiter) check(username string, interval time.Duration) error {
	sh := l.shardFor(username)
	sh.mu.Lock()
	defer sh.mu.Unlock()
	if val, ok := sh.users[username]; ok {
		if time.Now().Before(val.Starttime.Add(interval)) {
			return radiuserrors.NewOnlineLimitError("there is a authentication still in process")
		}
	}
	sh.users[username] = AuthRateUser{Username: username, Starttime: time.Now()}
	return nil
}

// release removes any rate-limit state for username, allowing an immediate
// subsequent authentication.
func (l *authRateLimiter) release(username string) {
	sh := l.shardFor(username)
	sh.mu.Lock()
	delete(sh.users, username)
	sh.mu.Unlock()
}

// get returns the stored rate-limit entry for username, used for introspection
// and testing.
func (l *authRateLimiter) get(username string) (AuthRateUser, bool) {
	sh := l.shardFor(username)
	sh.mu.Lock()
	defer sh.mu.Unlock()
	v, ok := sh.users[username]
	return v, ok
}

// len returns the total number of tracked users across all shards.
func (l *authRateLimiter) len() int {
	total := 0
	for _, sh := range l.shards {
		sh.mu.Lock()
		total += len(sh.users)
		sh.mu.Unlock()
	}
	return total
}
