package handlers

import (
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

	if getBool(state, stateKeyPendingSuccess) {
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
