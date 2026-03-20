package validators

import (
	"errors"
	"testing"
	"time"

	"github.com/go-ldap/ldap/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/domain"
	radiuserrors "github.com/talkincode/toughradius/v9/internal/radiusd/errors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	"layeh.com/radius"
)

// mockLDAPConfig is a test double for ldapConfigReader backed by a simple key/value map.
type mockLDAPConfig struct {
	values map[string]string
}

func (m *mockLDAPConfig) GetString(category, name string) string {
	return m.values[category+"."+name]
}

// fakeLDAPClient is a test double for ldapClient that records calls and returns scripted responses.
type fakeLDAPClient struct {
	bindErrs   []error
	bindCalls  [][2]string
	searchErr  error
	searchDNs  []string
	searchBase string
	searchFilt string
}

func (f *fakeLDAPClient) Bind(username, password string) error {
	f.bindCalls = append(f.bindCalls, [2]string{username, password})
	if len(f.bindErrs) == 0 {
		return nil
	}
	err := f.bindErrs[0]
	f.bindErrs = f.bindErrs[1:]
	return err
}

func (f *fakeLDAPClient) Search(searchRequest *ldap.SearchRequest) (*ldap.SearchResult, error) {
	f.searchBase = searchRequest.BaseDN
	f.searchFilt = searchRequest.Filter
	if f.searchErr != nil {
		return nil, f.searchErr
	}

	entries := make([]*ldap.Entry, 0, len(f.searchDNs))
	for _, dn := range f.searchDNs {
		entries = append(entries, &ldap.Entry{DN: dn})
	}
	return &ldap.SearchResult{Entries: entries}, nil
}

func (f *fakeLDAPClient) Close() error { return nil }

func TestValidatePasswordWithLDAP_Disabled(t *testing.T) {
	authCtx := &auth.AuthContext{
		User: &domain.RadiusUser{Username: "alice"},
		Metadata: map[string]interface{}{
			"config_mgr": &mockLDAPConfig{values: map[string]string{
				"radius.LdapEnabled": "false",
			}},
		},
	}

	handled, err := validatePasswordWithLDAP(authCtx, "pwd")
	require.NoError(t, err)
	assert.False(t, handled)
}

func TestIsLDAPEnabled_TruthyValues(t *testing.T) {
	for _, v := range []string{"true", "1", "enabled", "yes", "on", " TRUE "} {
		assert.True(t, isLDAPEnabled(v), "value %q should be truthy", v)
	}
	for _, v := range []string{"false", "0", "disabled", "no", "off", ""} {
		assert.False(t, isLDAPEnabled(v), "value %q should be falsy", v)
	}
}

func TestValidatePasswordWithLDAP_MissingRequiredConfig(t *testing.T) {
	authCtx := &auth.AuthContext{
		User: &domain.RadiusUser{Username: "alice"},
		Metadata: map[string]interface{}{
			"config_mgr": &mockLDAPConfig{values: map[string]string{
				"radius.LdapEnabled": "true",
			}},
		},
	}

	handled, err := validatePasswordWithLDAP(authCtx, "pwd")
	require.Error(t, err)
	assert.True(t, handled)
	authErr, ok := radiuserrors.GetAuthError(err)
	require.True(t, ok)
	assert.Equal(t, app.MetricsRadiusRejectLdapError, authErr.MetricsKey())
}

func TestValidatePasswordWithLDAP_Success(t *testing.T) {
	origFactory := newLDAPClient
	t.Cleanup(func() { newLDAPClient = origFactory })

	client := &fakeLDAPClient{
		searchDNs: []string{"uid=alice,ou=users,dc=example,dc=com"},
	}
	newLDAPClient = func(serverURL string, timeout time.Duration) (ldapClient, error) {
		assert.Equal(t, "ldap://127.0.0.1:389", serverURL)
		assert.Equal(t, 3*time.Second, timeout)
		return client, nil
	}

	authCtx := &auth.AuthContext{
		User: &domain.RadiusUser{Username: "alice"},
		Request: &radius.Request{
			Packet: radius.New(radius.CodeAccessRequest, []byte("secret")),
		},
		Metadata: map[string]interface{}{
			"config_mgr": &mockLDAPConfig{values: map[string]string{
				"radius.LdapEnabled":        "true",
				"radius.LdapServer":         "ldap://127.0.0.1:389",
				"radius.LdapBaseDN":         "dc=example,dc=com",
				"radius.LdapUserFilter":     "(uid={username})",
				"radius.LdapBindDN":         "cn=admin,dc=example,dc=com",
				"radius.LdapBindPassword":   "adminpass",
				"radius.LdapTimeoutSeconds": "3",
			}},
		},
	}

	handled, err := validatePasswordWithLDAP(authCtx, "userpass")
	require.NoError(t, err)
	assert.True(t, handled)
	assert.Equal(t, "dc=example,dc=com", client.searchBase)
	assert.Equal(t, "(uid=alice)", client.searchFilt)
	require.Len(t, client.bindCalls, 2)
	assert.Equal(t, [2]string{"cn=admin,dc=example,dc=com", "adminpass"}, client.bindCalls[0])
	assert.Equal(t, [2]string{"uid=alice,ou=users,dc=example,dc=com", "userpass"}, client.bindCalls[1])
}

func TestValidatePasswordWithLDAP_SearchResultNotSingleEntry(t *testing.T) {
	origFactory := newLDAPClient
	t.Cleanup(func() { newLDAPClient = origFactory })

	newLDAPClient = func(serverURL string, timeout time.Duration) (ldapClient, error) {
		return &fakeLDAPClient{searchDNs: []string{}}, nil
	}

	authCtx := &auth.AuthContext{
		User: &domain.RadiusUser{Username: "alice"},
		Metadata: map[string]interface{}{
			"config_mgr": &mockLDAPConfig{values: map[string]string{
				"radius.LdapEnabled": "true",
				"radius.LdapServer":  "ldap://127.0.0.1:389",
				"radius.LdapBaseDN":  "dc=example,dc=com",
			}},
		},
	}

	handled, err := validatePasswordWithLDAP(authCtx, "userpass")
	require.Error(t, err)
	assert.True(t, handled)
	authErr, ok := radiuserrors.GetAuthError(err)
	require.True(t, ok)
	assert.Equal(t, app.MetricsRadiusRejectLdapError, authErr.MetricsKey())
}

func TestValidatePasswordWithLDAP_ConnectionError(t *testing.T) {
	origFactory := newLDAPClient
	t.Cleanup(func() { newLDAPClient = origFactory })

	newLDAPClient = func(serverURL string, timeout time.Duration) (ldapClient, error) {
		return nil, errors.New("dial failed")
	}

	authCtx := &auth.AuthContext{
		User: &domain.RadiusUser{Username: "alice"},
		Metadata: map[string]interface{}{
			"config_mgr": &mockLDAPConfig{values: map[string]string{
				"radius.LdapEnabled": "true",
				"radius.LdapServer":  "ldap://127.0.0.1:389",
				"radius.LdapBaseDN":  "dc=example,dc=com",
			}},
		},
	}

	handled, err := validatePasswordWithLDAP(authCtx, "userpass")
	require.Error(t, err)
	assert.True(t, handled)
}

func TestValidatePasswordWithLDAP_InvalidTimeoutFallsBackToDefault(t *testing.T) {
	origFactory := newLDAPClient
	t.Cleanup(func() { newLDAPClient = origFactory })

	client := &fakeLDAPClient{
		searchDNs: []string{"uid=alice,ou=users,dc=example,dc=com"},
	}
	newLDAPClient = func(serverURL string, timeout time.Duration) (ldapClient, error) {
		assert.Equal(t, time.Duration(defaultLDAPTimeoutSeconds)*time.Second, timeout)
		return client, nil
	}

	authCtx := &auth.AuthContext{
		User: &domain.RadiusUser{Username: "alice"},
		Metadata: map[string]interface{}{
			"config_mgr": &mockLDAPConfig{values: map[string]string{
				"radius.LdapEnabled":        "true",
				"radius.LdapServer":         "ldap://127.0.0.1:389",
				"radius.LdapBaseDN":         "dc=example,dc=com",
				"radius.LdapTimeoutSeconds": "invalid",
			}},
		},
	}

	handled, err := validatePasswordWithLDAP(authCtx, "userpass")
	require.NoError(t, err)
	assert.True(t, handled)
}

func TestValidatePasswordWithLDAP_MultipleSearchResults(t *testing.T) {
	origFactory := newLDAPClient
	t.Cleanup(func() { newLDAPClient = origFactory })

	newLDAPClient = func(serverURL string, timeout time.Duration) (ldapClient, error) {
		return &fakeLDAPClient{
			searchDNs: []string{
				"uid=alice,ou=users,dc=example,dc=com",
				"uid=alice2,ou=users,dc=example,dc=com",
			},
		}, nil
	}

	authCtx := &auth.AuthContext{
		User: &domain.RadiusUser{Username: "alice"},
		Metadata: map[string]interface{}{
			"config_mgr": &mockLDAPConfig{values: map[string]string{
				"radius.LdapEnabled": "true",
				"radius.LdapServer":  "ldap://127.0.0.1:389",
				"radius.LdapBaseDN":  "dc=example,dc=com",
			}},
		},
	}

	handled, err := validatePasswordWithLDAP(authCtx, "userpass")
	require.Error(t, err)
	assert.True(t, handled)
	assert.Contains(t, err.Error(), "multiple entries")
}
