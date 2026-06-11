package radiusd

import (
	"crypto/hmac"
	"crypto/md5"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2869"
)

// buildAccessRequest builds and re-parses an Access-Request that carries a valid
// RFC 3579 Message-Authenticator computed with the given secret. The packet is
// encoded then re-parsed so it mirrors what the server sees on the wire.
func buildAccessRequest(t *testing.T, secret []byte, withMessageAuth bool) *radius.Packet {
	t.Helper()
	p := radius.New(radius.CodeAccessRequest, secret)
	p.Identifier = 7
	require.NoError(t, rfc2865.UserName_SetString(p, "alice"))
	require.NoError(t, rfc2865.UserPassword_SetString(p, "secret-pass"))

	if withMessageAuth {
		// Zero the attribute, marshal, then fill in the HMAC, matching the way a
		// compliant NAS computes it before sending.
		require.NoError(t, rfc2869.MessageAuthenticator_Set(p, make([]byte, md5.Size)))
		wire, err := p.MarshalBinary()
		require.NoError(t, err)
		mac := computeMessageAuthenticator(wire, secret)
		require.NoError(t, rfc2869.MessageAuthenticator_Set(p, mac))
	}

	wire, err := p.Encode()
	require.NoError(t, err)
	parsed, err := radius.Parse(wire, secret)
	require.NoError(t, err)
	return parsed
}

func TestVerifyMessageAuthenticator(t *testing.T) {
	secret := []byte("correct-secret")
	svc := &RadiusService{}

	t.Run("valid", func(t *testing.T) {
		pkt := buildAccessRequest(t, secret, true)
		assert.Equal(t, msgAuthValid, svc.verifyMessageAuthenticator(pkt, secret))
	})

	t.Run("absent", func(t *testing.T) {
		pkt := buildAccessRequest(t, secret, false)
		assert.Equal(t, msgAuthAbsent, svc.verifyMessageAuthenticator(pkt, secret))
	})

	t.Run("wrong secret", func(t *testing.T) {
		pkt := buildAccessRequest(t, secret, true)
		assert.Equal(t, msgAuthInvalid, svc.verifyMessageAuthenticator(pkt, []byte("wrong-secret")))
	})

	t.Run("tampered attribute triggers mismatch", func(t *testing.T) {
		pkt := buildAccessRequest(t, secret, true)
		// Inject a different reply value: an on-path attacker editing attributes
		// without recomputing the Message-Authenticator must be detected.
		require.NoError(t, rfc2865.NASIdentifier_SetString(pkt, "rogue-nas"))
		assert.Equal(t, msgAuthInvalid, svc.verifyMessageAuthenticator(pkt, secret))
	})

	t.Run("empty secret", func(t *testing.T) {
		pkt := buildAccessRequest(t, secret, true)
		assert.Equal(t, msgAuthInvalid, svc.verifyMessageAuthenticator(pkt, nil))
	})

	t.Run("duplicate attribute rejected", func(t *testing.T) {
		pkt := buildAccessRequest(t, secret, true)
		// RFC 3579 §3.2 permits at most one Message-Authenticator.
		require.NoError(t, rfc2869.MessageAuthenticator_Add(pkt, make([]byte, md5.Size)))
		assert.Equal(t, msgAuthInvalid, svc.verifyMessageAuthenticator(pkt, secret))
	})

	t.Run("wrong length rejected", func(t *testing.T) {
		pkt := buildAccessRequest(t, secret, false)
		require.NoError(t, rfc2869.MessageAuthenticator_Set(pkt, make([]byte, 8)))
		assert.Equal(t, msgAuthInvalid, svc.verifyMessageAuthenticator(pkt, secret))
	})
}

func TestMessageAuthDecision(t *testing.T) {
	cases := []struct {
		name        string
		mode        string
		result      messageAuthResult
		discard     bool
		warnMissing bool
	}{
		{"disabled ignores everything", MsgAuthModeDisabled, msgAuthInvalid, false, false},
		{"warn valid", MsgAuthModeWarn, msgAuthValid, false, false},
		{"warn missing warns only", MsgAuthModeWarn, msgAuthAbsent, false, true},
		{"warn invalid discards", MsgAuthModeWarn, msgAuthInvalid, true, false},
		{"enforce valid", MsgAuthModeEnforce, msgAuthValid, false, false},
		{"enforce missing discards", MsgAuthModeEnforce, msgAuthAbsent, true, false},
		{"enforce invalid discards", MsgAuthModeEnforce, msgAuthInvalid, true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			discard, warnMissing := messageAuthDecision(tc.mode, tc.result)
			assert.Equal(t, tc.discard, discard)
			assert.Equal(t, tc.warnMissing, warnMissing)
		})
	}
}

func TestStageMessageAuthenticatorWarnMode(t *testing.T) {
	secret := []byte("correct-secret")
	// nil ConfigMgr => warn mode.
	svc := &AuthService{RadiusService: &RadiusService{appCtx: &mockAppContext{}}}
	require.Equal(t, MsgAuthModeWarn, svc.MessageAuthenticatorMode())
	nas := &domain.NetNas{Secret: string(secret)}

	t.Run("valid does not stop", func(t *testing.T) {
		ctx := &AuthPipelineContext{Request: &radius.Request{Packet: buildAccessRequest(t, secret, true)}, NAS: nas}
		require.NoError(t, svc.stageMessageAuthenticator(ctx))
		assert.False(t, ctx.IsStopped())
	})

	t.Run("absent does not stop in warn mode", func(t *testing.T) {
		ctx := &AuthPipelineContext{Request: &radius.Request{Packet: buildAccessRequest(t, secret, false)}, NAS: nas}
		require.NoError(t, svc.stageMessageAuthenticator(ctx))
		assert.False(t, ctx.IsStopped())
	})

	t.Run("mismatch silently discards", func(t *testing.T) {
		pkt := buildAccessRequest(t, secret, true)
		require.NoError(t, rfc2865.NASIdentifier_SetString(pkt, "rogue-nas"))
		ctx := &AuthPipelineContext{Request: &radius.Request{Packet: pkt}, NAS: nas}
		require.NoError(t, svc.stageMessageAuthenticator(ctx))
		assert.True(t, ctx.IsStopped())
	})
}

func TestAddResponseMessageAuthenticator(t *testing.T) {
	secret := "correct-secret"
	// mockAppContext returns a nil ConfigMgr, so the mode defaults to warn and
	// signing is enabled.
	svc := &RadiusService{appCtx: &mockAppContext{}}
	require.Equal(t, MsgAuthModeWarn, svc.MessageAuthenticatorMode())

	req := radius.New(radius.CodeAccessRequest, []byte(secret))
	req.Identifier = 9
	resp := req.Response(radius.CodeAccessAccept)
	require.NoError(t, rfc2865.ReplyMessage_SetString(resp, "ok"))

	svc.addResponseMessageAuthenticator(resp, secret)

	// The signed response must carry exactly one Message-Authenticator that a
	// client can independently verify (NAS-side BlastRADIUS protection).
	got, err := rfc2869.MessageAuthenticator_Lookup(resp)
	require.NoError(t, err)
	require.Len(t, got, md5.Size)

	zeroed := req.Response(radius.CodeAccessAccept)
	require.NoError(t, rfc2865.ReplyMessage_SetString(zeroed, "ok"))
	require.NoError(t, rfc2869.MessageAuthenticator_Set(zeroed, make([]byte, md5.Size)))
	wire, err := zeroed.MarshalBinary()
	require.NoError(t, err)
	want := computeMessageAuthenticator(wire, []byte(secret))
	assert.True(t, hmac.Equal(got, want))
}

func TestAddResponseMessageAuthenticatorSkips(t *testing.T) {
	svc := &RadiusService{appCtx: &mockAppContext{}}
	req := radius.New(radius.CodeAccessRequest, []byte("s"))

	t.Run("empty secret", func(t *testing.T) {
		resp := req.Response(radius.CodeAccessReject)
		svc.addResponseMessageAuthenticator(resp, "")
		_, err := rfc2869.MessageAuthenticator_Lookup(resp)
		assert.Error(t, err)
	})

	t.Run("unknown nas placeholder", func(t *testing.T) {
		resp := req.Response(radius.CodeAccessReject)
		svc.addResponseMessageAuthenticator(resp, unknownNasSecret)
		_, err := rfc2869.MessageAuthenticator_Lookup(resp)
		assert.Error(t, err)
	})
}
