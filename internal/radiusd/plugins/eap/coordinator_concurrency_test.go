package eap

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

// concurrencyProbeHandler reproduces the GetState → process → SetState sequence
// that real handlers run, and reports whether two goroutines ever execute it
// concurrently for the same state. A NAS retransmit of an EAP-Response carries
// the same RADIUS State, so without per-state serialization in the coordinator
// the two would overlap and could desync the shared *tlsengine.Engine.
type concurrencyProbeHandler struct {
	active    int32
	overlaps  int32
	processed int32
}

func (h *concurrencyProbeHandler) Name() string               { return "eap-probe" }
func (h *concurrencyProbeHandler) EAPType() uint8             { return TypeTLS }
func (h *concurrencyProbeHandler) CanHandle(*EAPContext) bool { return true }

func (h *concurrencyProbeHandler) HandleIdentity(*EAPContext) (bool, error) {
	return true, nil
}

func (h *concurrencyProbeHandler) HandleResponse(ctx *EAPContext) (bool, error) {
	stateID := rfc2865.State_GetString(ctx.Request.Packet)

	if _, err := ctx.StateManager.GetState(stateID); err != nil {
		return false, err
	}

	if atomic.AddInt32(&h.active, 1) > 1 {
		atomic.AddInt32(&h.overlaps, 1)
	}
	// Simulate the slow handler/engine work during which a retransmit might be
	// dispatched concurrently.
	time.Sleep(5 * time.Millisecond)
	atomic.AddInt32(&h.active, -1)
	atomic.AddInt32(&h.processed, 1)

	state := &EAPState{StateID: stateID, Method: "eap-probe"}
	if err := ctx.StateManager.SetState(stateID, state); err != nil {
		return false, err
	}
	return false, nil
}

// TestHandleEAPRequest_ConcurrentRetransmitSerialized replays a retransmitted
// EAP-Response for the same state while the original is still being processed,
// and asserts the coordinator serializes them so the handler never runs
// concurrently for one state ID.
func TestHandleEAPRequest_ConcurrentRetransmitSerialized(t *testing.T) {
	const stateID = "shared-state-id"

	probe := &concurrencyProbeHandler{}
	registry := newMockHandlerRegistry()
	registry.Register(probe)

	sm := newMockStateManager()
	sm.states[stateID] = &EAPState{StateID: stateID, Method: "eap-probe"}

	coordinator := NewCoordinator(sm, &mockPasswordProvider{}, registry, false)

	newReq := func() *radius.Request {
		packet := createEAPChallengeResponse(1, TypeTLS, []byte{1, 2, 3, 4})
		_ = rfc2865.State_SetString(packet, stateID) //nolint:errcheck
		return &radius.Request{Packet: packet}
	}

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := newReq()
			_, _, _ = coordinator.HandleEAPRequest( //nolint:errcheck
				&mockResponseWriter{}, req, &domain.RadiusUser{}, &domain.NetNas{},
				req.Response(radius.CodeAccessChallenge), "secret", false, "eap-tls",
			)
		}()
	}
	wg.Wait()

	if got := atomic.LoadInt32(&probe.overlaps); got != 0 {
		t.Fatalf("handler ran concurrently for the same state %d time(s); expected serialization", got)
	}
	if got := atomic.LoadInt32(&probe.processed); got != 8 {
		t.Fatalf("expected all 8 requests to be processed, got %d", got)
	}
}
