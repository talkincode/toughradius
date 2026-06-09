// Package tlsengine drives a server-side TLS handshake for EAP-TLS
// (RFC 5216) across multiple EAP/RADIUS round-trips.
//
// EAP-TLS carries a normal TLS handshake inside EAP-Request/EAP-Response
// messages (RFC 5216 §2.1). Because the handshake spans several RADIUS
// round-trips, the TLS state machine must be suspended between rounds: the peer
// sends a flight of TLS records, the server feeds them to its TLS engine,
// collects the records the engine produces in response, and returns them to be
// transmitted in the next EAP-Request. This package bridges Go's blocking
// crypto/tls handshake to that turn-based model using an in-memory transport.
//
// Certificate validation (CA chain) is delegated to crypto/tls for EAP-TLS: the
// engine is configured with ClientAuth=RequireAndVerifyClientCert and a
// ClientCAs pool, so a handshake only completes when the peer presents a
// certificate that chains to a configured CA (RFC 5216 §2.2, §5.3). PEAP/TTLS
// server-only tunnels set Config.ServerOnly and authenticate the peer inside the
// protected tunnel instead. After a successful EAP-TLS handshake the validated
// peer identity is extracted from the client certificate per RFC 5216 §5.2.
package tlsengine

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"sync"
	"time"
)

// DefaultHandshakeTimeout bounds how long a single EAP-TLS handshake may take
// before the engine aborts it. EAP-TLS handshakes normally complete within
// seconds across a handful of RADIUS round-trips; the timeout guarantees that an
// abandoned handshake cannot leak the engine's background goroutine
// indefinitely.
const DefaultHandshakeTimeout = 30 * time.Second

// Errors returned by the engine.
var (
	// ErrNoConfig is returned when New is called without a usable config.
	ErrNoConfig = errors.New("tlsengine: nil config")
	// ErrNoServerCertificate is returned when the server cannot present a
	// usable certificate/private key for the TLS handshake.
	ErrNoServerCertificate = errors.New("tlsengine: missing server certificate")
	// ErrHandshakeIncomplete is returned by Identity when the handshake has not
	// completed successfully yet.
	ErrHandshakeIncomplete = errors.New("tlsengine: handshake not completed")
	// ErrNoPeerCertificate is returned when a completed handshake carries no
	// peer certificate (should not happen with RequireAndVerifyClientCert).
	ErrNoPeerCertificate = errors.New("tlsengine: no peer certificate presented")
)

// Config holds the materials needed to run the server side of an EAP-TLS
// handshake.
type Config struct {
	// ServerCertificate is the certificate (and private key) the RADIUS server
	// presents to the EAP-TLS peer.
	ServerCertificate tls.Certificate
	// ClientCAs is the set of certificate authorities used to verify the peer's
	// (client's) certificate chain. It MUST be non-nil: EAP-TLS requires the
	// peer to authenticate with a certificate (RFC 5216 §2.2).
	ClientCAs *x509.CertPool
	// MinVersion optionally pins the minimum TLS version (e.g. tls.VersionTLS12).
	// Zero lets crypto/tls choose its default.
	MinVersion uint16
	// ServerOnly disables client-certificate requests for tunneled EAP methods
	// such as PEAP/TTLS whose peers authenticate inside the protected tunnel
	// rather than with an outer TLS client certificate.
	ServerOnly bool
	// HandshakeTimeout bounds the total handshake duration. Zero selects
	// DefaultHandshakeTimeout.
	HandshakeTimeout time.Duration
}

// Engine runs one server-side TLS handshake, suspended between EAP rounds.
//
// An Engine is single-use: it drives exactly one handshake. It is safe to call
// its methods from one goroutine at a time (the EAP handler processes a given
// state serially); it is not designed for concurrent Process calls on the same
// engine.
type Engine struct {
	conn  *tls.Conn
	trans *transport

	mu        sync.Mutex
	done      bool
	hsErr     error
	finished  bool // handshake goroutine has returned
	timer     *time.Timer
	closeOnce sync.Once
}

// New creates an Engine and starts the server handshake in the background. The
// handshake immediately blocks waiting for the peer's first flight (ClientHello),
// which is delivered through Process.
func New(cfg *Config) (*Engine, error) {
	if cfg == nil {
		return nil, ErrNoConfig
	}
	if !cfg.ServerOnly && cfg.ClientCAs == nil {
		return nil, fmt.Errorf("%w: ClientCAs is required for EAP-TLS peer authentication", ErrNoConfig)
	}
	if len(cfg.ServerCertificate.Certificate) == 0 || cfg.ServerCertificate.PrivateKey == nil {
		return nil, ErrNoServerCertificate
	}

	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{cfg.ServerCertificate},
		MinVersion:   cfg.MinVersion,
	}
	if cfg.ServerOnly {
		tlsCfg.ClientAuth = tls.NoClientCert
	} else {
		tlsCfg.ClientAuth = tls.RequireAndVerifyClientCert
		tlsCfg.ClientCAs = cfg.ClientCAs
	}

	trans := newTransport()
	e := &Engine{
		conn:  tls.Server(trans, tlsCfg),
		trans: trans,
	}

	timeout := cfg.HandshakeTimeout
	if timeout <= 0 {
		timeout = DefaultHandshakeTimeout
	}
	// On timeout, close the transport so the blocked handshake goroutine
	// unblocks with an error and exits instead of leaking.
	e.timer = time.AfterFunc(timeout, func() { trans.close() })

	go e.runHandshake()
	return e, nil
}

// runHandshake performs the blocking TLS handshake and records its outcome.
func (e *Engine) runHandshake() {
	err := e.conn.Handshake()

	e.trans.mu.Lock()
	e.mu.Lock()
	e.finished = true
	if err != nil {
		e.hsErr = err
	} else {
		e.done = true
	}
	e.mu.Unlock()
	// Wake any Process call waiting for the handshake to progress.
	e.trans.cond.Broadcast()
	e.trans.mu.Unlock()
}

// Process feeds the peer's inbound TLS bytes (the reassembled EAP-TLS data for
// this round, possibly empty for the opening round) into the handshake and
// returns the TLS bytes the server wants to send back, whether the handshake has
// completed, and any handshake error.
//
// It blocks only until the TLS engine has consumed all the supplied input and is
// either finished or waiting for more peer data, so it always returns promptly
// once a flight has been produced.
func (e *Engine) Process(in []byte) (out []byte, done bool, err error) {
	t := e.trans

	t.mu.Lock()
	if len(in) > 0 {
		t.inbuf = append(t.inbuf, in...)
	}
	t.cond.Broadcast()

	for {
		e.mu.Lock()
		finished := e.finished
		hsErr := e.hsErr
		ok := e.done
		e.mu.Unlock()

		if finished {
			out = t.takeOutLocked()
			t.mu.Unlock()
			if hsErr != nil {
				return out, false, hsErr
			}
			return out, ok, nil
		}

		// The handshake goroutine has drained all input and is blocked waiting
		// for the next peer flight: this round's output is complete.
		if len(t.inbuf) == 0 && t.reading {
			out = t.takeOutLocked()
			t.mu.Unlock()
			return out, false, nil
		}

		t.cond.Wait()
	}
}

// Identity returns the validated peer identity extracted from the client
// certificate. It must only be called after Process reports done=true.
func (e *Engine) Identity() (*PeerIdentity, error) {
	e.mu.Lock()
	done := e.done
	e.mu.Unlock()
	if !done {
		return nil, ErrHandshakeIncomplete
	}

	state := e.conn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return nil, ErrNoPeerCertificate
	}
	return identityFromCertificate(state.PeerCertificates[0]), nil
}

// Close releases the engine's resources and unblocks the handshake goroutine if
// it is still running. It is safe to call multiple times.
func (e *Engine) Close() error {
	e.closeOnce.Do(func() {
		if e.timer != nil {
			e.timer.Stop()
		}
		e.trans.close()
		_ = e.conn.Close() //nolint:errcheck // best-effort cleanup
	})
	return nil
}
