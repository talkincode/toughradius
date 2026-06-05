package radiusd

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/panjf2000/ants/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSubmitAcctTaskDropsOnOverload verifies that when the worker pool is
// saturated, an accounting task is dropped instead of spawning an unbounded
// goroutine, keeping goroutine usage bounded under overload.
func TestSubmitAcctTaskDropsOnOverload(t *testing.T) {
	pool, err := ants.NewPool(1, ants.WithNonblocking(true))
	require.NoError(t, err)
	defer pool.Release()

	acct := &AcctService{RadiusService: &RadiusService{TaskPool: pool}}

	started := make(chan struct{})
	block := make(chan struct{})

	// Occupy the only worker with a blocking task.
	require.True(t, acct.submitAcctTask(func() {
		close(started)
		<-block
	}, "busy"))
	<-started

	// With the single worker busy, the next task must be dropped (back-pressure),
	// not executed on a freshly spawned goroutine.
	var ran int32
	accepted := acct.submitAcctTask(func() {
		atomic.AddInt32(&ran, 1)
	}, "overflow")

	assert.False(t, accepted, "task should be dropped when pool is saturated")

	// Give any (incorrectly) spawned goroutine a chance to run before asserting.
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, int32(0), atomic.LoadInt32(&ran), "dropped task must not execute")

	close(block)
}
