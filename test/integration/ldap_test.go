//go:build integration

package integration

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"

	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/ldapauth"
	eap "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
	"github.com/talkincode/toughradius/v9/pkg/common"
)

type ldapTestEnv struct {
	url            string
	baseDN         string
	userFilter     string
	bindDNTemplate string
	adminDN        string
	adminPassword  string
	username       string
	password       string
}

// TestLDAPIntegrationAcceptance proves M14.6 against a real OpenLDAP service:
// LDAP bind, not the local RadiusUser password, authorizes bare PAP and
// EAP-TTLS/PAP; wrong passwords and non-PAP methods reject diagnostically; and
// a directory outage is counted as radus_reject_ldap_error.
func TestLDAPIntegrationAcceptance(t *testing.T) {
	ldapEnv := requireLDAPEnv(t)
	waitForOpenLDAP(t, ldapEnv)
	configureLDAP(t, ldapEnv)

	const secret = "it-ldap-secret"
	suffix := uniqueSuffix()
	nasIP := net.ParseIP("10.204.0.1")
	nasID := "it-ldap-nas-" + suffix

	require.NoError(t, h.appCtx.DB().Create(&domain.NetNas{
		ID:         common.UUIDint64(),
		Identifier: nasID,
		Ipaddr:     nasIP.String(),
		Secret:     secret,
		VendorCode: "0",
		Status:     common.ENABLED,
	}).Error)

	profileID := seedProfile(t, "it-ldap-profile-"+suffix)
	seedLDAPLocalUser(t, profileID, ldapEnv.username)

	serverAddr := h.radiusServerAddr()

	t.Run("bare PAP bind success ignores local password", func(t *testing.T) {
		resp := exchange(t, serverAddr, secret, ldapEnv.username, ldapEnv.password, nasID, nasIP.String())
		h.radiusSvc.ReleaseAuthRateLimit(ldapEnv.username)
		assert.Equalf(t, radius.CodeAccessAccept, resp.Code,
			"LDAP-backed PAP must authenticate through OpenLDAP bind, got %v (%q)",
			resp.Code, rfc2865.ReplyMessage_GetString(resp))
	})

	t.Run("bare PAP wrong password rejects as credential failure", func(t *testing.T) {
		before := app.GetRadiusMetrics(app.MetricsRadiusRejectPasswdError)
		resp := exchange(t, serverAddr, secret, ldapEnv.username, ldapEnv.password+"-wrong", nasID, nasIP.String())
		h.radiusSvc.ReleaseAuthRateLimit(ldapEnv.username)
		assert.Equalf(t, radius.CodeAccessReject, resp.Code, "wrong LDAP password must reject, got %v", resp.Code)
		assert.Equal(t, before+1, app.GetRadiusMetrics(app.MetricsRadiusRejectPasswdError))
	})

	t.Run("directory unavailable rejects as ldap error", func(t *testing.T) {
		port, err := freeTCPPort()
		require.NoError(t, err)
		setLDAPSettings(t, map[string]string{
			"ServerURL": fmt.Sprintf("ldap://127.0.0.1:%d", port),
			"Timeout":   "1",
		})

		before := app.GetRadiusMetrics(app.MetricsRadiusRejectLdapError)
		resp := exchange(t, serverAddr, secret, ldapEnv.username, ldapEnv.password, nasID, nasIP.String())
		h.radiusSvc.ReleaseAuthRateLimit(ldapEnv.username)
		assert.Equalf(t, radius.CodeAccessReject, resp.Code, "unreachable LDAP must reject, got %v", resp.Code)
		assert.Equal(t, before+1, app.GetRadiusMetrics(app.MetricsRadiusRejectLdapError))
		assert.Contains(t, strings.ToLower(rfc2865.ReplyMessage_GetString(resp)), "ldap backend unavailable")
	})

	t.Run("non-PAP request rejects explicitly while LDAP is active", func(t *testing.T) {
		before := app.GetRadiusMetrics(app.MetricsRadiusRejectLdapError)
		resp := exchangeCHAP(t, serverAddr, secret, ldapEnv.username, nasID, nasIP.String())
		h.radiusSvc.ReleaseAuthRateLimit(ldapEnv.username)
		assert.Equalf(t, radius.CodeAccessReject, resp.Code, "LDAP-backed CHAP must reject, got %v", resp.Code)
		assert.Equal(t, before+1, app.GetRadiusMetrics(app.MetricsRadiusRejectLdapError))
		reply := strings.ToLower(rfc2865.ReplyMessage_GetString(resp))
		assert.Contains(t, reply, "no pap password")
		assert.Contains(t, reply, "ldap bind")
	})

	t.Run("EAP-TTLS inner PAP bind success ignores local password", func(t *testing.T) {
		restoreEapMethod(t)
		ca := newEAPTLSTestCA(t, "IT LDAP TTLS Root CA "+suffix)
		serverCert := ca.issueServer(t, "radius.example.com")
		configureTTLS(t, serverCert)

		sup := &ttlsSupplicant{
			serverAddr: serverAddr,
			secret:     secret,
			username:   ldapEnv.username,
			password:   ldapEnv.password,
			nasID:      nasID,
			nasIP:      nasIP,
			clientCfg: &tls.Config{ //nolint:gosec // G402: TLS 1.2 pin matches the TTLS integration harness.
				RootCAs:    ca.pool(),
				ServerName: "radius.example.com",
				MinVersion: tls.VersionTLS12,
				MaxVersion: tls.VersionTLS12,
			},
			innerMethod: ttlsInnerPAP,
		}
		resp := sup.authenticate(t)
		h.radiusSvc.ReleaseAuthRateLimit(ldapEnv.username)
		require.Equalf(t, radius.CodeAccessAccept, resp.Code,
			"valid TTLS-PAP LDAP bind must authenticate, got %v (%q)",
			resp.Code, rfc2865.ReplyMessage_GetString(resp))
		assertEAPCode(t, resp, eap.CodeSuccess)
		assertMPPEKeys(t, resp, secret, sup.lastReqAuth)
	})
}

func requireLDAPEnv(t *testing.T) ldapTestEnv {
	t.Helper()
	url := os.Getenv("TEST_LDAP_URL")
	if url == "" {
		if integrationRequired() {
			t.Fatal("integration tests required but TEST_LDAP_URL is not configured")
		}
		t.Skip("skipping LDAP integration tests: TEST_LDAP_URL not configured")
	}
	return ldapTestEnv{
		url:            url,
		baseDN:         getenvDefault("TEST_LDAP_BASE_DN", "dc=example,dc=org"),
		userFilter:     getenvDefault("TEST_LDAP_USER_FILTER", "(cn=%s)"),
		bindDNTemplate: getenvDefault("TEST_LDAP_BIND_DN_TEMPLATE", "cn=%s,ou=users,dc=example,dc=org"),
		adminDN:        getenvDefault("TEST_LDAP_ADMIN_DN", "cn=admin,dc=example,dc=org"),
		adminPassword:  getenvDefault("TEST_LDAP_ADMIN_PASSWORD", "adminpassword"),
		username:       getenvDefault("TEST_LDAP_USER", "alice"),
		password:       getenvDefault("TEST_LDAP_PASSWORD", "alice-password"),
	}
}

func getenvDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func waitForOpenLDAP(t *testing.T, env ldapTestEnv) {
	t.Helper()
	cfg := env.ldapConfig()
	deadline := time.Now().Add(30 * time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		err := ldapauth.New(cfg).Verify(ctx, env.username, env.password)
		cancel()
		if err == nil {
			return
		}
		lastErr = err
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatalf("OpenLDAP did not become ready: %v", lastErr)
}

func (e ldapTestEnv) ldapConfig() ldapauth.Config {
	return ldapauth.Config{
		Enabled:            true,
		ServerURL:          e.url,
		BaseDN:             e.baseDN,
		BindMode:           ldapauth.BindModeSearch,
		BindDNTemplate:     e.bindDNTemplate,
		SearchBindDN:       e.adminDN,
		SearchBindPassword: e.adminPassword,
		UserFilter:         e.userFilter,
		Timeout:            3 * time.Second,
	}
}

func configureLDAP(t *testing.T, env ldapTestEnv) {
	t.Helper()
	setLDAPSettings(t, map[string]string{
		"Enabled":            "true",
		"ServerURL":          env.url,
		"BindMode":           "search",
		"BindDNTemplate":     env.bindDNTemplate,
		"BaseDN":             env.baseDN,
		"UserFilter":         env.userFilter,
		"SearchBindDN":       env.adminDN,
		"SearchBindPassword": env.adminPassword,
		"StartTLS":           "false",
		"TLSSkipVerify":      "false",
		"Timeout":            "3",
	})
}

func setLDAPSettings(t *testing.T, values map[string]string) {
	t.Helper()
	cm := h.appCtx.ConfigMgr()
	previous := make(map[string]string, len(values))
	for key := range values {
		previous[key] = cm.GetString("ldap", key)
	}
	for key, value := range values {
		require.NoErrorf(t, cm.Set("ldap", key, value), "set ldap.%s", key)
	}
	t.Cleanup(func() {
		for key, value := range previous {
			_ = cm.Set("ldap", key, value)
		}
	})
}

func seedLDAPLocalUser(t *testing.T, profileID int64, username string) {
	t.Helper()
	require.NoError(t, h.appCtx.DB().Create(&domain.RadiusUser{
		ID:         common.UUIDint64(),
		ProfileId:  profileID,
		Username:   username,
		Password:   "local-password-must-not-be-used",
		Status:     common.ENABLED,
		ExpireTime: time.Now().AddDate(1, 0, 0),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}).Error)
}

func exchangeCHAP(t *testing.T, serverAddr, secret, username, nasID, nasIP string) *radius.Packet {
	t.Helper()
	packet := radius.New(radius.CodeAccessRequest, []byte(secret))
	_ = rfc2865.UserName_SetString(packet, username)
	_ = rfc2865.NASIdentifier_SetString(packet, nasID)
	_ = rfc2865.NASIPAddress_Set(packet, net.ParseIP(nasIP))
	_ = rfc2865.CHAPPassword_Set(packet, append([]byte{1}, make([]byte, 16)...))
	_ = rfc2865.CHAPChallenge_Set(packet, []byte("0123456789abcdef"))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := radius.Exchange(ctx, packet, serverAddr)
	require.NoError(t, err)
	return resp
}
