package handlers

import (
	"crypto/tls"
	"fmt"

	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap/tlsengine"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap/tlsfragment"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

// tlsTunnel drives the TLS record exchange shared by EAP-TLS and PEAP outer
// tunnels.
//
// RFC 5216 §2.1.5 and §3.1 define the Length/More-Fragments/Start framing used
// by EAP-TLS. PEAPv0 has no formal RFC, but Microsoft [MS-PEAP] carries the
// outer TLS handshake using the same EAP-TLS-derived flags with the low bits as
// the PEAP version field. The eapType parameter keeps the shared state machine
// byte-identical for EAP-TLS while emitting type 25 for PEAP.
type tlsTunnel struct {
	eapType             uint8
	maxFragment         int
	configProvider      TLSConfigProvider
	onHandshakeComplete func(ctx *eap.EAPContext, state *eap.EAPState, engine *tlsengine.Engine) (bool, error)
	// clientSpeaksFirst marks a tunneled method whose phase 2 is peer-initiated:
	// after the outer handshake the supplicant sends the first inner flight
	// (application data) rather than the bare ACK that EAP-TLS/PEAP expect.
	// EAP-TTLS sets this (RFC 5281 §7.3); EAP-TLS and PEAP (server speaks first)
	// leave it false, preserving their existing behavior.
	clientSpeaksFirst bool
	// protectedSuccess enables the RFC 9190 §2.1.1 protected success indication
	// for TLS 1.3 handshakes: after its final handshake flight the server sends
	// one octet of 0x00 as encrypted TLS application data and only then, on the
	// peer's acknowledgement, an EAP-Success. This authenticates the success
	// result inside the tunnel (a plain EAP-Success is unprotected and a
	// compliant TLS 1.3 peer would reject it, §2.5). Only plain EAP-TLS sets
	// this; PEAP/TTLS follow their own tunneled success semantics and are
	// unaffected. TLS 1.2 handshakes never send the indication (RFC 5216 has no
	// such message), preserving the existing byte-identical 1.2 flow.
	protectedSuccess bool
	// onCommit validates the authentication decision before the protected
	// success indication is emitted. RFC 9190 §2.1.1 defines the 0x00 octet as
	// a success commitment, so any policy that could still reject the peer
	// (e.g. EAP-TLS certificate identity binding, RFC 5216 §5.2) must run
	// first: a server must never commit success and then send EAP-Failure.
	// A non-nil error rejects the handshake without sending the indication.
	// Nil onCommit skips the pre-commit check.
	onCommit func(ctx *eap.EAPContext, state *eap.EAPState, engine *tlsengine.Engine) error
	// onApplicationData drives a tunneled inner EAP method (PEAP phase 2) once
	// the outer TLS tunnel is established. It is given the decrypted inbound
	// inner EAP bytes (nil on the very first call, which asks it to produce the
	// opening inner request) and returns the next inner EAP bytes to send back
	// through the tunnel (reply), whether authentication has succeeded, or an
	// error to reject. It is nil for plain EAP-TLS, which never enters the inner
	// phase.
	onApplicationData func(ctx *eap.EAPContext, state *eap.EAPState, engine *tlsengine.Engine, inner []byte) (reply []byte, success bool, err error)
}

// HandleResponse handles EAP-Response messages carrying TLS handshake bytes.
func (t *tlsTunnel) HandleResponse(ctx *eap.EAPContext) (bool, error) {
	stateID := rfc2865.State_GetString(ctx.Request.Packet)
	if stateID == "" {
		return false, eap.ErrStateNotFound
	}

	state, err := ctx.StateManager.GetState(stateID)
	if err != nil {
		return false, err
	}

	frag, err := tlsfragment.Parse(ctx.EAPMessage.Data)
	if err != nil {
		return false, err
	}

	// Once the outer tunnel is established the conversation switches to the
	// inner EAP method carried as TLS application data (PEAP phase 2).
	if getBool(state, stateKeyInnerActive) {
		return t.handleInnerRound(ctx, state, frag)
	}

	if getBool(state, stateKeyPendingSuccess) {
		// EAP-TTLS makes phase 2 peer-initiated (RFC 5281 §7.3): the supplicant
		// sends its inner AVP flight immediately after the outer handshake,
		// rather than the bare ACK that EAP-TLS/PEAP expect here. Route that
		// flight into the inner phase instead of rejecting it as an unexpected
		// fragment. A bare ACK still falls through to onHandshakeComplete below.
		if t.clientSpeaksFirst && !frag.IsACK() {
			clearKey(state, stateKeyPendingSuccess)
			setBool(state, stateKeyInnerActive, true)
			return t.handleInnerRound(ctx, state, frag)
		}
		if !frag.IsACK() {
			closeEngine(state)
			return false, eap.ErrTLSUnexpectedFragment
		}
		engine, err := t.engineFor(state)
		if err != nil {
			return false, err
		}
		return t.onHandshakeComplete(ctx, state, engine)
	}

	if t.hasQueuedFragments(state) {
		return t.sendNextFragment(ctx, state, frag)
	}

	reassembler := loadReassembler(state)
	complete, err := reassembler.Accept(frag)
	if err != nil {
		closeEngine(state)
		return false, err
	}

	if !complete {
		saveReassembler(state, reassembler)
		if err := ctx.StateManager.SetState(stateID, state); err != nil {
			closeEngine(state)
			return false, err
		}
		return false, t.writeChallenge(ctx, stateID, t.buildFragmentACK(ctx.EAPMessage.Identifier+1))
	}

	tlsInput := reassembler.Buffer()
	resetReassembler(state)
	return t.advanceHandshake(ctx, state, tlsInput)
}

func (t *tlsTunnel) advanceHandshake(ctx *eap.EAPContext, state *eap.EAPState, tlsInput []byte) (bool, error) {
	engine, err := t.engineFor(state)
	if err != nil {
		return false, err
	}

	out, done, hsErr := engine.Process(tlsInput)
	if hsErr != nil {
		closeEngine(state)
		return false, fmt.Errorf("%w: %v", eap.ErrTLSHandshakeFailed, hsErr)
	}

	// RFC 9190 §2.1.1: with TLS 1.3 the server commits to the handshake result
	// by sending one octet of 0x00 as protected application data after its
	// final handshake message; EAP-Success may only follow the peer's
	// acknowledgement of that record. Session tickets are disabled in the
	// engine, so this commitment record is appended to whatever handshake bytes
	// remain (typically none: the TLS 1.3 server finishes on the peer's flight)
	// and rides the normal flight -> pending-success -> ACK state machine.
	if done && t.protectedSuccess && engine.NegotiatedVersion() == tls.VersionTLS13 {
		// The 0x00 octet is a success commitment, so run the remaining
		// authentication policy (certificate identity binding) first and
		// reject without the indication if it fails.
		if t.onCommit != nil {
			if cerr := t.onCommit(ctx, state, engine); cerr != nil {
				closeEngine(state)
				return false, cerr
			}
		}
		commit, werr := engine.WriteApplication([]byte{0x00})
		if werr != nil {
			closeEngine(state)
			return false, fmt.Errorf("%w: %v", eap.ErrTLSHandshakeFailed, werr)
		}
		out = append(out, commit...)
	}

	if len(out) == 0 {
		if done {
			return t.onHandshakeComplete(ctx, state, engine)
		}
		closeEngine(state)
		return false, eap.ErrTLSUnexpectedFragment
	}

	return t.startFlight(ctx, state, out, done)
}

func (t *tlsTunnel) startFlight(ctx *eap.EAPContext, state *eap.EAPState, tlsData []byte, done bool) (bool, error) {
	packets, err := tlsfragment.Fragment(tlsData, t.maxFragment)
	if err != nil {
		return false, err
	}

	first, rest := packets[0], packets[1:]
	setBool(state, stateKeyHandshakeDone, done)

	if len(rest) > 0 {
		setFragments(state, rest)
	} else if done {
		setBool(state, stateKeyPendingSuccess, true)
	}

	if err := ctx.StateManager.SetState(state.StateID, state); err != nil {
		closeEngine(state)
		return false, err
	}
	return false, t.writeChallenge(ctx, state.StateID, t.buildEAPRequest(ctx.EAPMessage.Identifier+1, first))
}

func (t *tlsTunnel) sendNextFragment(ctx *eap.EAPContext, state *eap.EAPState, frag *tlsfragment.Packet) (bool, error) {
	if !frag.IsACK() {
		closeEngine(state)
		return false, eap.ErrTLSUnexpectedFragment
	}

	queue := getFragments(state)
	next := queue[0]
	queue = queue[1:]

	if len(queue) > 0 {
		setFragments(state, queue)
	} else {
		clearKey(state, stateKeyOutFrags)
		if getBool(state, stateKeyHandshakeDone) {
			setBool(state, stateKeyPendingSuccess, true)
		}
	}

	if err := ctx.StateManager.SetState(state.StateID, state); err != nil {
		closeEngine(state)
		return false, err
	}
	return false, t.writeChallenge(ctx, state.StateID, t.buildEAPRequest(ctx.EAPMessage.Identifier+1, next))
}

// startInner transitions the tunnel from the completed outer handshake into the
// inner EAP phase (PEAP phase 2). It marks the inner phase active and asks the
// inner callback to produce the opening inner request (inner == nil), which is
// then encrypted and sent through the tunnel. It is invoked from a PEAP
// onHandshakeComplete callback in place of granting/rejecting.
func (t *tlsTunnel) startInner(ctx *eap.EAPContext, state *eap.EAPState, engine *tlsengine.Engine) (bool, error) {
	setBool(state, stateKeyInnerActive, true)
	return t.driveInner(ctx, state, engine, nil)
}

// handleInnerRound processes one inner-phase EAP-Response: it drains any queued
// outbound application fragments, reassembles the inbound TLS records, decrypts
// them into an inner EAP message, and routes it through driveInner.
func (t *tlsTunnel) handleInnerRound(ctx *eap.EAPContext, state *eap.EAPState, frag *tlsfragment.Packet) (bool, error) {
	if t.hasQueuedFragments(state) {
		return t.sendNextAppFragment(ctx, state, frag)
	}

	engine, err := t.engineFor(state)
	if err != nil {
		return false, err
	}

	// A client-speaks-first method (EAP-TTLS) signals a terminal inner result
	// with an empty EAP-TTLS frame: after an MS-CHAP2-Success is tunneled the
	// peer acknowledges with a zero-length Data field (RFC 5281 §11.2.4). With no
	// outbound fragments queued and no inbound reassembly in progress, route that
	// bare ACK into the inner handler as an empty flight (inner == nil) rather
	// than blocking ReadApplication on application records the peer will never
	// send. EAP-TLS/PEAP leave clientSpeaksFirst false and are unaffected.
	if t.clientSpeaksFirst && frag.IsACK() && !reassemblyInProgress(state) {
		return t.driveInner(ctx, state, engine, nil)
	}

	reassembler := loadReassembler(state)
	complete, err := reassembler.Accept(frag)
	if err != nil {
		closeEngine(state)
		return false, err
	}
	if !complete {
		saveReassembler(state, reassembler)
		if err := ctx.StateManager.SetState(state.StateID, state); err != nil {
			closeEngine(state)
			return false, err
		}
		return false, t.writeChallenge(ctx, state.StateID, t.buildFragmentACK(ctx.EAPMessage.Identifier+1))
	}

	records := reassembler.Buffer()
	resetReassembler(state)

	inner, err := engine.ReadApplication(records)
	if err != nil {
		closeEngine(state)
		return false, err
	}
	return t.driveInner(ctx, state, engine, inner)
}

// driveInner invokes the inner EAP callback and acts on its decision: grant on
// success (the callback has already populated the Access-Accept attributes such
// as the MS-MPPE keys), reject on error, or encrypt and send the next inner
// request otherwise. The engine is kept alive across inner rounds and closed
// only on a terminal outcome.
func (t *tlsTunnel) driveInner(ctx *eap.EAPContext, state *eap.EAPState, engine *tlsengine.Engine, inner []byte) (bool, error) {
	if t.onApplicationData == nil {
		closeEngine(state)
		return false, eap.ErrPEAPInnerNotImplemented
	}

	reply, success, err := t.onApplicationData(ctx, state, engine, inner)
	if err != nil {
		closeEngine(state)
		return false, err
	}
	if success {
		closeEngine(state)
		state.Success = true
		if serr := ctx.StateManager.SetState(state.StateID, state); serr != nil {
			return false, serr
		}
		return true, nil
	}

	out, err := engine.WriteApplication(reply)
	if err != nil {
		closeEngine(state)
		return false, err
	}
	return t.startAppFlight(ctx, state, out)
}

// startAppFlight fragments an outbound inner-EAP application record and sends
// the first fragment. Unlike startFlight it never sets the handshake
// pending-success terminal: the inner phase ends only when onApplicationData
// reports success or an error, so after the final fragment the tunnel simply
// awaits the peer's next inner response.
func (t *tlsTunnel) startAppFlight(ctx *eap.EAPContext, state *eap.EAPState, tlsData []byte) (bool, error) {
	packets, err := tlsfragment.Fragment(tlsData, t.maxFragment)
	if err != nil {
		closeEngine(state)
		return false, err
	}

	first, rest := packets[0], packets[1:]
	if len(rest) > 0 {
		setFragments(state, rest)
	}

	if err := ctx.StateManager.SetState(state.StateID, state); err != nil {
		closeEngine(state)
		return false, err
	}
	return false, t.writeChallenge(ctx, state.StateID, t.buildEAPRequest(ctx.EAPMessage.Identifier+1, first))
}

// sendNextAppFragment sends the next queued inner-phase application fragment in
// response to a peer ACK. Like startAppFlight it omits the pending-success
// terminal used by the handshake path.
func (t *tlsTunnel) sendNextAppFragment(ctx *eap.EAPContext, state *eap.EAPState, frag *tlsfragment.Packet) (bool, error) {
	if !frag.IsACK() {
		closeEngine(state)
		return false, eap.ErrTLSUnexpectedFragment
	}

	queue := getFragments(state)
	next := queue[0]
	queue = queue[1:]
	if len(queue) > 0 {
		setFragments(state, queue)
	} else {
		clearKey(state, stateKeyOutFrags)
	}

	if err := ctx.StateManager.SetState(state.StateID, state); err != nil {
		closeEngine(state)
		return false, err
	}
	return false, t.writeChallenge(ctx, state.StateID, t.buildEAPRequest(ctx.EAPMessage.Identifier+1, next))
}

func (t *tlsTunnel) engineFor(state *eap.EAPState) (*tlsengine.Engine, error) {
	if state.Data != nil {
		if e, ok := state.Data[stateKeyEngine].(*tlsengine.Engine); ok && e != nil {
			return e, nil
		}
	}

	if t.configProvider == nil {
		return nil, eap.ErrTLSNotConfigured
	}
	cfg, err := t.configProvider()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", eap.ErrTLSNotConfigured, err)
	}
	if cfg == nil {
		return nil, eap.ErrTLSNotConfigured
	}

	engine, err := tlsengine.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", eap.ErrTLSNotConfigured, err)
	}
	if state.Data == nil {
		state.Data = make(map[string]interface{})
	}
	state.Data[stateKeyEngine] = engine
	return engine, nil
}

func (t *tlsTunnel) hasQueuedFragments(state *eap.EAPState) bool {
	return len(getFragments(state)) > 0
}

func (t *tlsTunnel) buildFragmentACK(identifier uint8) []byte {
	return t.buildEAPRequest(identifier, &tlsfragment.Packet{})
}

func (t *tlsTunnel) buildEAPRequest(identifier uint8, frag *tlsfragment.Packet) []byte {
	msg := &eap.EAPMessage{
		Code:       eap.CodeRequest,
		Identifier: identifier,
		Type:       t.eapType,
		Data:       frag.Encode(),
	}
	return msg.Encode()
}

func (t *tlsTunnel) writeChallenge(ctx *eap.EAPContext, stateID string, eapData []byte) error {
	response := ctx.Request.Response(radius.CodeAccessChallenge)
	_ = rfc2865.State_SetString(response, stateID) //nolint:errcheck
	eap.SetEAPMessageAndAuth(response, eapData, ctx.Secret)
	return ctx.ResponseWriter.Write(response)
}
