package tlsengine

import (
	"errors"
	"time"
)

// DefaultAppReadTimeout bounds a single ReadApplication call. A tunneled inner
// EAP method (PEAP/TTLS) exchanges one small application record per RADIUS
// round-trip, so a completed flight decrypts promptly; the timeout only guards
// against a peer that supplies an incomplete TLS record, ensuring the engine's
// internal Read cannot block a RADIUS worker goroutine indefinitely.
const DefaultAppReadTimeout = 10 * time.Second

// maxAppRecord bounds a single decrypted application-data read. A TLS record's
// plaintext is at most 16 KiB (RFC 8446 §5.1); tunneled inner EAP packets are
// far smaller, so this comfortably holds one inner EAP message.
const maxAppRecord = 16 * 1024

// Errors returned by the application-data and key-export methods.
var (
	// ErrHandshakeNotComplete is returned by the post-handshake methods
	// (WriteApplication, ReadApplication, ExportKey) when called before the
	// TLS handshake has completed successfully.
	ErrHandshakeNotComplete = errors.New("tlsengine: handshake not complete")
	// ErrEngineClosed is returned when an application-data operation is
	// attempted after the engine's transport has been closed.
	ErrEngineClosed = errors.New("tlsengine: engine closed")
	// ErrAppReadTimeout is returned when ReadApplication does not obtain a full
	// application record within DefaultAppReadTimeout.
	ErrAppReadTimeout = errors.New("tlsengine: application read timeout")
)

// WriteApplication encrypts plaintext as TLS application data and returns the
// resulting TLS records for transmission to the peer.
//
// It is used by tunneled EAP methods (PEAP/TTLS) to send an inner EAP request
// through the protected tunnel: the caller fragments the returned records into
// EAP-Request messages exactly as it does for handshake flights. The handshake
// must have completed (Process reported done=true); otherwise it returns
// ErrHandshakeNotComplete.
func (e *Engine) WriteApplication(plaintext []byte) ([]byte, error) {
	if !e.HandshakeComplete() {
		return nil, ErrHandshakeNotComplete
	}

	// conn.Write encrypts plaintext into one or more TLS records and hands them
	// to the in-memory transport, whose Write never blocks. The handshake
	// goroutine has already returned, so this call runs to completion here.
	if _, err := e.conn.Write(plaintext); err != nil {
		return nil, err
	}

	e.trans.mu.Lock()
	out := e.trans.takeOutLocked()
	e.trans.mu.Unlock()
	return out, nil
}

// ReadApplication feeds the peer's reassembled inbound TLS records (one inner
// EAP flight) and returns the decrypted application plaintext.
//
// The caller is responsible for reassembling EAP-TLS fragments into the
// complete TLS record(s) before calling. The handshake must have completed.
// The read runs under DefaultAppReadTimeout so a truncated or malicious flight
// cannot wedge the calling goroutine: on timeout the transport is closed and
// ErrAppReadTimeout is returned.
func (e *Engine) ReadApplication(records []byte) ([]byte, error) {
	if !e.HandshakeComplete() {
		return nil, ErrHandshakeNotComplete
	}

	e.trans.mu.Lock()
	if e.trans.closed {
		e.trans.mu.Unlock()
		return nil, ErrEngineClosed
	}
	e.trans.inbuf = append(e.trans.inbuf, records...)
	e.trans.cond.Broadcast()
	e.trans.mu.Unlock()

	type readResult struct {
		data []byte
		err  error
	}
	ch := make(chan readResult, 1)
	go func() {
		buf := make([]byte, maxAppRecord)
		n, err := e.conn.Read(buf)
		ch <- readResult{data: append([]byte(nil), buf[:n]...), err: err}
	}()

	timer := time.NewTimer(e.readTimeout())
	defer timer.Stop()
	select {
	case r := <-ch:
		if r.err != nil {
			return nil, r.err
		}
		return r.data, nil
	case <-timer.C:
		// Unblock the pending Read so its goroutine exits instead of leaking.
		e.trans.close()
		return nil, ErrAppReadTimeout
	}
}

// ExportKey returns exported keying material derived from the completed TLS
// session per RFC 5705, used by EAP methods to derive the MSK and the
// MS-MPPE-Send-Key / MS-MPPE-Recv-Key (RFC 2548). For PEAP/TTLS and EAP-TLS
// with TLS 1.2 the label is "client EAP encryption" with a 64-octet length
// (RFC 5216 §2.3); EAP-TLS with TLS 1.3 uses "EXPORTER_EAP_TLS_Key_Material"
// with the EAP type octet as context and the full 128-octet length (RFC 9190
// §2.3). In both layouts octets 0..31 of the MSK form the MS-MPPE-Recv-Key
// and octets 32..63 the MS-MPPE-Send-Key.
//
// The handshake must have completed. crypto/tls requires TLS 1.3 or the
// Extended Master Secret extension (RFC 7627) to export keying material; both
// are negotiated by default with modern peers.
func (e *Engine) ExportKey(label string, context []byte, length int) ([]byte, error) {
	if !e.HandshakeComplete() {
		return nil, ErrHandshakeNotComplete
	}
	state := e.conn.ConnectionState()
	return state.ExportKeyingMaterial(label, context, length)
}

// HandshakeComplete reports whether the TLS handshake has completed
// successfully, i.e. the engine is ready for application-data exchange and key
// export.
func (e *Engine) HandshakeComplete() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.done
}

// NegotiatedVersion returns the TLS protocol version negotiated by the
// completed handshake (e.g. tls.VersionTLS12 or tls.VersionTLS13). It returns 0
// if the handshake has not completed, so callers can branch on the negotiated
// version only after Process reports done=true. EAP-TLS uses it to apply the
// TLS 1.3-only protected success indication (RFC 9190 §2.1.1) while keeping the
// TLS 1.2 flow byte-identical.
func (e *Engine) NegotiatedVersion() uint16 {
	if !e.HandshakeComplete() {
		return 0
	}
	return e.conn.ConnectionState().Version
}

// readTimeout returns the configured ReadApplication timeout, defaulting to
// DefaultAppReadTimeout when unset.
func (e *Engine) readTimeout() time.Duration {
	if e.appReadTimeout > 0 {
		return e.appReadTimeout
	}
	return DefaultAppReadTimeout
}
