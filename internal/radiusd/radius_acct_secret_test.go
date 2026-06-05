package radiusd

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2866"
)

// buildAccountingRequest builds and re-parses an Accounting-Request packet so it
// carries a valid keyed Request Authenticator computed with the given secret.
func buildAccountingRequest(t *testing.T, secret []byte) *radius.Packet {
	t.Helper()
	p := radius.New(radius.CodeAccountingRequest, secret)
	p.Identifier = 1
	require.NoError(t, rfc2865.UserName_SetString(p, "testuser"))
	require.NoError(t, rfc2866.AcctStatusType_Set(p, rfc2866.AcctStatusType_Value_Start))
	wire, err := p.Encode()
	require.NoError(t, err)
	parsed, err := radius.Parse(wire, secret)
	require.NoError(t, err)
	return parsed
}

func TestCheckRequestSecretAccounting(t *testing.T) {
	secret := []byte("correct-secret")
	pkt := buildAccountingRequest(t, secret)

	svc := &RadiusService{}

	// Correct secret validates successfully.
	require.NoError(t, svc.CheckRequestSecret(pkt, secret))

	// Wrong secret is rejected: this is the forged-accounting protection.
	err := svc.CheckRequestSecret(pkt, []byte("wrong-secret"))
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrSecretMismatch)

	// Empty secret is rejected.
	assert.ErrorIs(t, svc.CheckRequestSecret(pkt, nil), ErrSecretEmpty)
}

func TestRADIUSSecretResolvesPerNas(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping DB-backed test in short mode")
	}
	appCtx, _ := setupTestEnv(t)
	defer appCtx.Release()

	svc := NewRadiusService(appCtx)
	defer svc.Release()

	nas := &domain.NetNas{
		ID:         1,
		Identifier: "nas-1",
		Ipaddr:     "10.9.9.9",
		Secret:     "per-nas-secret",
		VendorCode: "0",
		Status:     common.ENABLED,
	}
	require.NoError(t, appCtx.DB().Create(nas).Error)

	addr := &net.UDPAddr{IP: net.ParseIP("10.9.9.9"), Port: 5000}
	secret, err := svc.RADIUSSecret(context.Background(), addr)
	require.NoError(t, err)
	assert.Equal(t, "per-nas-secret", string(secret))

	// Unknown NAS returns a non-empty placeholder (so the handler can log it),
	// never the hardcoded legacy "mysecret".
	unknownAddr := &net.UDPAddr{IP: net.ParseIP("172.31.0.1"), Port: 5000}
	placeholder, err := svc.RADIUSSecret(context.Background(), unknownAddr)
	require.NoError(t, err)
	assert.NotEmpty(t, placeholder)
	assert.NotEqual(t, "mysecret", string(placeholder))
	assert.NotEqual(t, "per-nas-secret", string(placeholder))
}
