package ldapauth

import (
	"context"
	"crypto/tls"
	"errors"
	"testing"
	"time"

	"github.com/go-ldap/ldap/v3"
	"github.com/stretchr/testify/require"
)

// fakeSettings is an in-memory Settings implementation keyed by "category.Name".
type fakeSettings struct {
	str map[string]string
	i64 map[string]int64
	b   map[string]bool
}

func (f fakeSettings) GetSettingsStringValue(category, key string) string {
	return f.str[category+"."+key]
}
func (f fakeSettings) GetSettingsInt64Value(category, key string) int64 {
	return f.i64[category+"."+key]
}
func (f fakeSettings) GetSettingsBoolValue(category, key string) bool {
	return f.b[category+"."+key]
}

// bindCall records a Bind invocation against the fake connection.
type bindCall struct {
	dn, password string
}

// fakeConn is a scripted directory connection used to exercise the verifier
// without a real LDAP server.
type fakeConn struct {
	startTLSErr    error
	startTLSCalled bool
	binds          []bindCall
	bindFunc       func(dn, password string) error
	searchReq      *ldap.SearchRequest
	searchFunc     func(req *ldap.SearchRequest) (*ldap.SearchResult, error)
	closed         bool
}

func (f *fakeConn) StartTLS(*tls.Config) error { //nolint:revive // satisfies conn
	f.startTLSCalled = true
	return f.startTLSErr
}

func (f *fakeConn) Bind(dn, password string) error {
	f.binds = append(f.binds, bindCall{dn, password})
	if f.bindFunc != nil {
		return f.bindFunc(dn, password)
	}
	return nil
}

func (f *fakeConn) Search(req *ldap.SearchRequest) (*ldap.SearchResult, error) {
	f.searchReq = req
	if f.searchFunc != nil {
		return f.searchFunc(req)
	}
	return &ldap.SearchResult{}, nil
}

func (f *fakeConn) Close() error {
	f.closed = true
	return nil
}

// invalidCreds builds the directory error returned for a wrong password.
func invalidCreds() error {
	return ldap.NewError(ldap.LDAPResultInvalidCredentials, errors.New("invalid credentials"))
}

// newVerifier wires a verifier to a fake dialer. When fc is nil the dialer
// returns dialErr; otherwise it returns fc and records that it was called.
func newVerifier(cfg Config, fc *fakeConn, dialErr error) (*Verifier, *bool) {
	dialed := false
	v := &Verifier{cfg: cfg, dial: func(Config) (conn, error) {
		dialed = true
		if fc == nil {
			return nil, dialErr
		}
		return fc, nil
	}}
	return v, &dialed
}

func templateConfig() Config {
	return Config{
		Enabled:        true,
		ServerURL:      "ldap://dc.example.com:389",
		BindMode:       BindModeTemplate,
		BindDNTemplate: "uid=%s,ou=people,dc=example,dc=com",
		Timeout:        5 * time.Second,
	}
}

func searchConfig() Config {
	return Config{
		Enabled:            true,
		ServerURL:          "ldaps://dc.example.com:636",
		BaseDN:             "dc=example,dc=com",
		BindMode:           BindModeSearch,
		SearchBindDN:       "cn=svc,dc=example,dc=com",
		SearchBindPassword: "svcpass",
		UserFilter:         "(uid=%s)",
		Timeout:            5 * time.Second,
	}
}

func TestLoadConfig(t *testing.T) {
	t.Run("defaults when unset", func(t *testing.T) {
		cfg := LoadConfig(fakeSettings{})
		require.False(t, cfg.Enabled)
		require.Equal(t, defaultTimeout, cfg.Timeout, "missing timeout falls back to default")
		require.Equal(t, BindModeTemplate, cfg.BindMode, "unset mode normalizes to template")
	})

	t.Run("populated", func(t *testing.T) {
		s := fakeSettings{
			str: map[string]string{
				"ldap.ServerURL":      "ldaps://dc.example.com:636",
				"ldap.BaseDN":         "dc=example,dc=com",
				"ldap.BindMode":       "search",
				"ldap.BindDNTemplate": "uid=%s,dc=example,dc=com",
				"ldap.SearchBindDN":   "cn=svc,dc=example,dc=com",
				"ldap.UserFilter":     "(sAMAccountName=%s)",
			},
			i64: map[string]int64{"ldap.Timeout": 9},
			b:   map[string]bool{"ldap.Enabled": true, "ldap.StartTLS": true},
		}
		cfg := LoadConfig(s)
		require.True(t, cfg.Enabled)
		require.True(t, cfg.StartTLS)
		require.Equal(t, BindModeSearch, cfg.BindMode)
		require.Equal(t, "ldaps://dc.example.com:636", cfg.ServerURL)
		require.Equal(t, "(sAMAccountName=%s)", cfg.UserFilter)
		require.Equal(t, 9*time.Second, cfg.Timeout)
	})

	t.Run("non-positive timeout falls back", func(t *testing.T) {
		cfg := LoadConfig(fakeSettings{i64: map[string]int64{"ldap.Timeout": -3}})
		require.Equal(t, defaultTimeout, cfg.Timeout)
	})
}

func TestConfigValidate(t *testing.T) {
	require.ErrorIs(t, Config{Enabled: false}.Validate(), ErrDisabled)
	require.ErrorIs(t, Config{Enabled: true}.Validate(), ErrNotConfigured)
	require.ErrorIs(t, Config{Enabled: true, ServerURL: "ldap://x", BindMode: BindModeTemplate}.Validate(), ErrNotConfigured)
	require.ErrorIs(t, Config{Enabled: true, ServerURL: "ldap://x", BindMode: BindModeSearch, BaseDN: "dc=x"}.Validate(), ErrNotConfigured)
	require.ErrorIs(t, Config{Enabled: true, ServerURL: "ldap://x", BindMode: BindMode("bogus")}.Validate(), ErrNotConfigured)
	require.NoError(t, templateConfig().Validate())
	require.NoError(t, searchConfig().Validate())
}

func TestVerify_ShortCircuitsWithoutDial(t *testing.T) {
	t.Run("disabled", func(t *testing.T) {
		v, dialed := newVerifier(Config{Enabled: false}, &fakeConn{}, nil)
		require.ErrorIs(t, v.Verify(context.Background(), "alice", "pw"), ErrDisabled)
		require.False(t, *dialed, "must not dial when disabled")
	})

	t.Run("empty password", func(t *testing.T) {
		v, dialed := newVerifier(templateConfig(), &fakeConn{}, nil)
		require.ErrorIs(t, v.Verify(context.Background(), "alice", ""), ErrEmptyCredential)
		require.False(t, *dialed, "must not dial with empty credentials (RFC 4513 §5.1.2)")
	})

	t.Run("empty username", func(t *testing.T) {
		v, dialed := newVerifier(templateConfig(), &fakeConn{}, nil)
		require.ErrorIs(t, v.Verify(context.Background(), "", "pw"), ErrEmptyCredential)
		require.False(t, *dialed)
	})

	t.Run("canceled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		v, dialed := newVerifier(templateConfig(), &fakeConn{}, nil)
		require.ErrorIs(t, v.Verify(ctx, "alice", "pw"), context.Canceled)
		require.False(t, *dialed)
	})
}

func TestVerify_TemplateMode(t *testing.T) {
	t.Run("success binds rendered DN", func(t *testing.T) {
		fc := &fakeConn{}
		v, _ := newVerifier(templateConfig(), fc, nil)
		require.NoError(t, v.Verify(context.Background(), "alice", "secret"))
		require.Len(t, fc.binds, 1)
		require.Equal(t, "uid=alice,ou=people,dc=example,dc=com", fc.binds[0].dn)
		require.Equal(t, "secret", fc.binds[0].password)
		require.True(t, fc.closed, "connection must be closed")
	})

	t.Run("wrong password maps to ErrInvalidCredentials", func(t *testing.T) {
		fc := &fakeConn{bindFunc: func(string, string) error { return invalidCreds() }}
		v, _ := newVerifier(templateConfig(), fc, nil)
		require.ErrorIs(t, v.Verify(context.Background(), "alice", "wrong"), ErrInvalidCredentials)
	})

	t.Run("DN special chars are escaped", func(t *testing.T) {
		fc := &fakeConn{}
		v, _ := newVerifier(templateConfig(), fc, nil)
		require.NoError(t, v.Verify(context.Background(), "a,b+c", "pw"))
		require.Len(t, fc.binds, 1)
		// EscapeDN escapes the comma and plus so they cannot alter the DN structure.
		require.Equal(t, `uid=a\,b\+c,ou=people,dc=example,dc=com`, fc.binds[0].dn)
	})
}

func TestVerify_SearchMode(t *testing.T) {
	t.Run("success: service bind, search, user bind", func(t *testing.T) {
		userDN := "cn=Alice Example,ou=people,dc=example,dc=com"
		fc := &fakeConn{
			searchFunc: func(*ldap.SearchRequest) (*ldap.SearchResult, error) {
				return &ldap.SearchResult{Entries: []*ldap.Entry{{DN: userDN}}}, nil
			},
		}
		v, _ := newVerifier(searchConfig(), fc, nil)
		require.NoError(t, v.Verify(context.Background(), "alice", "secret"))
		require.Len(t, fc.binds, 2)
		require.Equal(t, "cn=svc,dc=example,dc=com", fc.binds[0].dn, "first bind is the service account")
		require.Equal(t, "svcpass", fc.binds[0].password)
		require.Equal(t, userDN, fc.binds[1].dn, "second bind is the discovered user DN")
		require.Equal(t, "secret", fc.binds[1].password)
	})

	t.Run("user not found", func(t *testing.T) {
		fc := &fakeConn{searchFunc: func(*ldap.SearchRequest) (*ldap.SearchResult, error) {
			return &ldap.SearchResult{}, nil
		}}
		v, _ := newVerifier(searchConfig(), fc, nil)
		require.ErrorIs(t, v.Verify(context.Background(), "ghost", "pw"), ErrUserNotFound)
		require.Len(t, fc.binds, 1, "only the service bind happens; no user bind")
	})

	t.Run("ambiguous match rejected", func(t *testing.T) {
		fc := &fakeConn{searchFunc: func(*ldap.SearchRequest) (*ldap.SearchResult, error) {
			return &ldap.SearchResult{Entries: []*ldap.Entry{{DN: "a"}, {DN: "b"}}}, nil
		}}
		v, _ := newVerifier(searchConfig(), fc, nil)
		require.ErrorIs(t, v.Verify(context.Background(), "dup", "pw"), ErrUserNotFound)
		require.Len(t, fc.binds, 1, "ambiguous result must not trigger a user bind")
	})

	t.Run("filter injection is escaped", func(t *testing.T) {
		fc := &fakeConn{searchFunc: func(*ldap.SearchRequest) (*ldap.SearchResult, error) {
			return &ldap.SearchResult{Entries: []*ldap.Entry{{DN: "uid=x,dc=example,dc=com"}}}, nil
		}}
		v, _ := newVerifier(searchConfig(), fc, nil)
		require.NoError(t, v.Verify(context.Background(), "*)(uid=*", "pw"))
		require.NotNil(t, fc.searchReq)
		require.Equal(t, `(uid=\2a\29\28uid=\2a)`, fc.searchReq.Filter, "metacharacters must be escaped")
	})

	t.Run("user bind wrong password", func(t *testing.T) {
		fc := &fakeConn{
			searchFunc: func(*ldap.SearchRequest) (*ldap.SearchResult, error) {
				return &ldap.SearchResult{Entries: []*ldap.Entry{{DN: "uid=x,dc=example,dc=com"}}}, nil
			},
			bindFunc: func(dn, _ string) error {
				if dn == "uid=x,dc=example,dc=com" {
					return invalidCreds()
				}
				return nil // service bind succeeds
			},
		}
		v, _ := newVerifier(searchConfig(), fc, nil)
		require.ErrorIs(t, v.Verify(context.Background(), "x", "wrong"), ErrInvalidCredentials)
	})

	t.Run("service bind failure is unavailable", func(t *testing.T) {
		fc := &fakeConn{bindFunc: func(string, string) error { return errors.New("connection refused") }}
		v, _ := newVerifier(searchConfig(), fc, nil)
		require.ErrorIs(t, v.Verify(context.Background(), "x", "pw"), ErrUnavailable)
	})
}

func TestVerify_DialErrorIsUnavailable(t *testing.T) {
	v, dialed := newVerifier(templateConfig(), nil, errors.New("no route to host"))
	require.ErrorIs(t, v.Verify(context.Background(), "alice", "pw"), ErrUnavailable)
	require.True(t, *dialed)
}

func TestVerify_StartTLS(t *testing.T) {
	t.Run("invoked when enabled", func(t *testing.T) {
		cfg := templateConfig()
		cfg.StartTLS = true
		fc := &fakeConn{}
		v, _ := newVerifier(cfg, fc, nil)
		require.NoError(t, v.Verify(context.Background(), "alice", "pw"))
		require.True(t, fc.startTLSCalled)
	})

	t.Run("failure is unavailable and skips bind", func(t *testing.T) {
		cfg := templateConfig()
		cfg.StartTLS = true
		fc := &fakeConn{startTLSErr: errors.New("tls handshake failed")}
		v, _ := newVerifier(cfg, fc, nil)
		require.ErrorIs(t, v.Verify(context.Background(), "alice", "pw"), ErrUnavailable)
		require.Empty(t, fc.binds, "no bind after StartTLS failure")
	})
}
