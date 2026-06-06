package radiusd

import (
	"context"
	"encoding/binary"
	"errors"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"layeh.com/radius"
)

// radsecHandlerFunc adapts a function to the RadsecHandler interface.
type radsecHandlerFunc func(w radius.ResponseWriter, r *radius.Request)

func (f radsecHandlerFunc) ServeRADIUS(w radius.ResponseWriter, r *radius.Request) { f(w, r) }

// radsecSecretSourceFunc adapts a function to the radius.SecretSource interface.
type radsecSecretSourceFunc func(ctx context.Context, remoteAddr net.Addr) ([]byte, error)

func (f radsecSecretSourceFunc) RADIUSSecret(ctx context.Context, remoteAddr net.Addr) ([]byte, error) {
	return f(ctx, remoteAddr)
}

// radsecTestPacket builds a minimal valid RADIUS-over-TCP packet: a 4-byte
// header (Code, Identifier, Length) followed by the 16-byte authenticator and
// no attributes. Length is 20 (4 header + 16 authenticator).
func radsecTestPacket(id byte) []byte {
	buf := make([]byte, 20)
	buf[0] = 1 // Code = Access-Request
	buf[1] = id
	binary.BigEndian.PutUint16(buf[2:4], 20)
	return buf
}

// TestRadsecPacketServer_BoundsConcurrentHandlers verifies that the read loop
// never spawns more concurrent handler goroutines than RadsecWorker, even when
// many packets arrive while handlers are slow. This is the regression guard for
// the unbounded-goroutine issue: the worker slot must be acquired before the
// handler goroutine is spawned.
func TestRadsecPacketServer_BoundsConcurrentHandlers(t *testing.T) {
	const workers = 4
	const total = 16

	var running int32
	var maxRunning int32
	release := make(chan struct{})
	started := make(chan struct{}, total)

	handler := radsecHandlerFunc(func(w radius.ResponseWriter, r *radius.Request) {
		cur := atomic.AddInt32(&running, 1)
		for {
			old := atomic.LoadInt32(&maxRunning)
			if cur <= old || atomic.CompareAndSwapInt32(&maxRunning, old, cur) {
				break
			}
		}
		started <- struct{}{}
		<-release
		atomic.AddInt32(&running, -1)
	})

	server := &RadsecPacketServer{
		Handler: handler,
		SecretSource: radsecSecretSourceFunc(func(context.Context, net.Addr) ([]byte, error) {
			return []byte("testing123"), nil
		}),
		RadsecWorker: workers,
	}

	clientConn, serverConn := net.Pipe()
	go func() { _ = server.Serve(serverConn) }()

	// Feed packets with distinct identifiers so the per-(IP,Identifier) dedup
	// does not collapse them. net.Pipe is unbounded backpressure: writes block
	// until the server reads, so a stalled read loop naturally stalls writes.
	go func() {
		for i := 0; i < total; i++ {
			if _, err := clientConn.Write(radsecTestPacket(byte(i))); err != nil {
				return
			}
		}
	}()

	// Wait until the pool is fully occupied.
	for i := 0; i < workers; i++ {
		select {
		case <-started:
		case <-time.After(3 * time.Second):
			t.Fatalf("timed out waiting for handler %d to start", i)
		}
	}

	// With the pool saturated, no additional handler must start until a slot
	// frees. If a 5th handler starts here, the goroutine was spawned before
	// acquiring a worker slot (the bug).
	select {
	case <-started:
		t.Fatal("a handler started while the worker pool was saturated; goroutines are not bounded")
	case <-time.After(250 * time.Millisecond):
	}

	if got := atomic.LoadInt32(&running); got != workers {
		t.Fatalf("expected exactly %d concurrent handlers, got %d", workers, got)
	}

	// Release everything and drain the remaining packets.
	close(release)
	for i := workers; i < total; i++ {
		select {
		case <-started:
		case <-time.After(3 * time.Second):
			t.Fatalf("timed out draining handler %d", i)
		}
	}

	if got := atomic.LoadInt32(&maxRunning); got > workers {
		t.Fatalf("max concurrent handlers %d exceeded worker limit %d", got, workers)
	}

	_ = clientConn.Close()
	_ = server.Shutdown(context.Background())
}

// TestRadsecPacketServer_AcquireWorkerSlotShutdown verifies that a blocked slot
// acquisition is released when the server shuts down, so Serve cannot hang on a
// saturated pool during shutdown.
func TestRadsecPacketServer_AcquireWorkerSlotShutdown(t *testing.T) {
	s := &RadsecPacketServer{RadsecWorker: 1}
	s.mu.Lock()
	s.initLocked()
	s.mu.Unlock()

	if !s.acquireWorkerSlot() {
		t.Fatal("first acquire on an empty pool should succeed")
	}

	// Pool is now full. A second acquire would block; trigger shutdown and
	// ensure it returns false instead of blocking forever.
	s.ctxDone()

	done := make(chan bool, 1)
	go func() { done <- s.acquireWorkerSlot() }()

	select {
	case ok := <-done:
		if ok {
			t.Fatal("acquire should fail after shutdown when the pool is saturated")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("acquireWorkerSlot blocked after shutdown")
	}
}

// TestRadsecPacketServer_AcquireWorkerSlotShutdownFreeSlot verifies that once the
// server is shutting down, acquireWorkerSlot returns false even when the pool
// still has free capacity, so no new handlers are started during shutdown.
func TestRadsecPacketServer_AcquireWorkerSlotShutdownFreeSlot(t *testing.T) {
	s := &RadsecPacketServer{RadsecWorker: 4}
	s.mu.Lock()
	s.initLocked()
	s.mu.Unlock()

	s.ctxDone()

	if s.acquireWorkerSlot() {
		t.Fatal("acquire should fail after shutdown even with free pool capacity")
	}
}

// TestRadsecPacketServer_KeepSlotUnlessShutdownReleases verifies the race-closing
// re-check: when shutdown is signaled after a slot was acquired, the slot is
// released and false is returned, so a handler is not started during shutdown.
func TestRadsecPacketServer_KeepSlotUnlessShutdownReleases(t *testing.T) {
	s := &RadsecPacketServer{RadsecWorker: 1}
	s.mu.Lock()
	s.initLocked()
	s.mu.Unlock()

	// Simulate a successful non-blocking send that raced with shutdown.
	s.workerPool <- struct{}{}
	s.ctxDone()

	if s.keepSlotUnlessShutdown() {
		t.Fatal("keepSlotUnlessShutdown should return false during shutdown")
	}
	if len(s.workerPool) != 0 {
		t.Fatalf("slot should be released on shutdown, occupied=%d", len(s.workerPool))
	}
}

// TestRadsecPacketServer_DefaultWorkerPool verifies that an unconfigured
// RadsecWorker falls back to a bounded, buffered pool instead of producing an
// unbuffered (deadlock-prone) channel.
func TestRadsecPacketServer_DefaultWorkerPool(t *testing.T) {
	s := &RadsecPacketServer{}
	s.mu.Lock()
	s.initLocked()
	s.mu.Unlock()

	if cap(s.workerPool) != defaultRadsecWorkers {
		t.Fatalf("expected default worker pool capacity %d, got %d", defaultRadsecWorkers, cap(s.workerPool))
	}
}

// TestRadsecPacketServer_MalformedFrameClosesConnection verifies that an
// over-max length prefix is treated as a fatal framing error: Serve returns
// (closing the connection) instead of continuing, which would misread the
// frame's unconsumed body as subsequent packet headers and desynchronize the
// stream. A valid packet sent before the malformed frame must still be served,
// and the bytes embedded in the malformed frame's body must NOT be processed.
func TestRadsecPacketServer_MalformedFrameClosesConnection(t *testing.T) {
	var handled int32
	served := make(chan struct{}, 8)
	handler := radsecHandlerFunc(func(w radius.ResponseWriter, r *radius.Request) {
		atomic.AddInt32(&handled, 1)
		served <- struct{}{}
	})

	server := &RadsecPacketServer{
		Handler: handler,
		SecretSource: radsecSecretSourceFunc(func(context.Context, net.Addr) ([]byte, error) {
			return []byte("testing123"), nil
		}),
		RadsecWorker: 4,
	}

	clientConn, serverConn := net.Pipe()
	serveErr := make(chan error, 1)
	go func() { serveErr <- server.Serve(serverConn) }()

	// A valid packet must be served normally.
	go func() { _, _ = clientConn.Write(radsecTestPacket(1)) }()
	select {
	case <-served:
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for the valid packet to be served")
	}

	// Now send a frame advertising an over-max Length, followed by a valid
	// 20-byte packet as its body. Without the framing guard the read loop would
	// continue and misread that body as the next packet's header.
	overMax := []byte{1, 9, 0xFF, 0xFF}
	overMax = append(overMax, radsecTestPacket(2)...)
	go func() { _, _ = clientConn.Write(overMax) }()

	select {
	case err := <-serveErr:
		if !errors.Is(err, errMalformedFrame) {
			t.Fatalf("expected Serve to return errMalformedFrame, got %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Serve did not close the connection on a malformed frame")
	}

	// The embedded body must not have been processed as a second packet.
	if got := atomic.LoadInt32(&handled); got != 1 {
		t.Fatalf("expected exactly 1 served packet, got %d (stream desynchronized)", got)
	}

	// Serve must have actually closed its side of the connection, not merely
	// returned. With net.Pipe, the peer observes io.EOF once the other end is
	// closed.
	_ = clientConn.SetReadDeadline(time.Now().Add(2 * time.Second))
	if _, err := clientConn.Read(make([]byte, 1)); err == nil {
		t.Fatal("expected the connection to be closed by Serve, but read succeeded")
	}

	_ = clientConn.Close()
	_ = server.Shutdown(context.Background())
}
