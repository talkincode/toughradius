package tlsengine

import (
	"io"
	"net"
	"sync"
	"time"
)

// transport is an in-memory net.Conn that bridges the blocking crypto/tls
// handshake to the turn-based EAP delivery model.
//
//   - Write (called by crypto/tls) appends the records the server wants to send
//     to outbuf, which the driver drains each round.
//   - Read (called by crypto/tls) returns buffered peer bytes from inbuf. When
//     inbuf is empty it sets reading=true and blocks: that transition is the
//     signal to the driver that the TLS engine has consumed the current flight
//     and is waiting for the peer's next one.
//
// All fields are guarded by mu; cond coordinates the handshake goroutine and the
// driver (Engine.Process). The same mu/cond pair is shared with the Engine so a
// single Wait wakes on either transport progress or handshake completion.
type transport struct {
	mu     sync.Mutex
	cond   *sync.Cond
	inbuf  []byte // peer -> server bytes awaiting Read
	outbuf []byte // server -> peer bytes awaiting the driver
	closed bool
	// reading is true while crypto/tls is blocked in Read waiting for more peer
	// data (inbuf drained). It is the "flight complete" signal for the driver.
	reading bool
}

func newTransport() *transport {
	t := &transport{}
	t.cond = sync.NewCond(&t.mu)
	return t
}

// Read implements net.Conn. It returns available peer bytes, or blocks until
// more arrive or the transport is closed.
func (t *transport) Read(p []byte) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for len(t.inbuf) == 0 && !t.closed {
		// Signal the driver that the engine has drained all input and now needs
		// the next peer flight.
		t.reading = true
		t.cond.Broadcast()
		t.cond.Wait()
	}
	t.reading = false

	if len(t.inbuf) == 0 && t.closed {
		return 0, io.EOF
	}

	n := copy(p, t.inbuf)
	t.inbuf = t.inbuf[n:]
	if len(t.inbuf) == 0 {
		t.inbuf = nil
	}
	return n, nil
}

// Write implements net.Conn. It buffers the bytes crypto/tls wants to send and
// wakes the driver so it can collect this round's flight.
func (t *transport) Write(p []byte) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return 0, io.ErrClosedPipe
	}
	t.outbuf = append(t.outbuf, p...)
	t.cond.Broadcast()
	return len(p), nil
}

// takeOutLocked returns and clears the buffered outbound bytes. The caller must
// hold t.mu.
func (t *transport) takeOutLocked() []byte {
	if len(t.outbuf) == 0 {
		return nil
	}
	out := t.outbuf
	t.outbuf = nil
	return out
}

// close marks the transport closed and wakes any blocked goroutine.
func (t *transport) close() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.closed = true
	t.cond.Broadcast()
}

// Close implements net.Conn.
func (t *transport) Close() error {
	t.close()
	return nil
}

// The remaining net.Conn methods are unused by crypto/tls during a handshake
// over this synchronous in-memory transport, so they are satisfied with inert
// implementations.

type memAddr struct{}

func (memAddr) Network() string { return "eap-tls" }
func (memAddr) String() string  { return "eap-tls" }

func (t *transport) LocalAddr() net.Addr                { return memAddr{} }
func (t *transport) RemoteAddr() net.Addr               { return memAddr{} }
func (t *transport) SetDeadline(_ time.Time) error      { return nil }
func (t *transport) SetReadDeadline(_ time.Time) error  { return nil }
func (t *transport) SetWriteDeadline(_ time.Time) error { return nil }
