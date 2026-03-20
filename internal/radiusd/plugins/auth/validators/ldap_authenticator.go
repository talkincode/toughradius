package validators

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-ldap/ldap/v3"
	"github.com/talkincode/toughradius/v9/internal/app"
	radiuserrors "github.com/talkincode/toughradius/v9/internal/radiusd/errors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	"layeh.com/radius/rfc2865"
)

type ldapConfigReader interface {
	GetString(category, name string) string
}

type ldapClient interface {
	Bind(username, password string) error
	Search(searchRequest *ldap.SearchRequest) (*ldap.SearchResult, error)
	Close() error
}

type ldapConn struct {
	*ldap.Conn
}

const defaultLDAPTimeoutSeconds int64 = 5

var newLDAPClient = func(serverURL string, timeout time.Duration) (ldapClient, error) {
	conn, err := ldap.DialURL(serverURL)
	if err != nil {
		return nil, err
	}
	conn.SetTimeout(timeout)
	return &ldapConn{Conn: conn}, nil
}

// validatePasswordWithLDAP validates PAP cleartext password against LDAP when enabled.
//
// Returns:
//   - handled=true when LDAP is enabled and an LDAP auth attempt was made (err indicates result)
//   - handled=false when LDAP is disabled or configuration is unavailable, allowing local password validation fallback
func validatePasswordWithLDAP(authCtx *auth.AuthContext, requestPassword string) (bool, error) {
	cfgMgr, ok := authCtx.Metadata["config_mgr"].(ldapConfigReader)
	if !ok || cfgMgr == nil {
		return false, nil
	}

	if !isLDAPEnabled(cfgMgr.GetString("radius", "LdapEnabled")) {
		return false, nil
	}

	serverURL := strings.TrimSpace(cfgMgr.GetString("radius", "LdapServer"))
	baseDN := strings.TrimSpace(cfgMgr.GetString("radius", "LdapBaseDN"))
	if serverURL == "" || baseDN == "" {
		return true, radiuserrors.NewAuthError(
			app.MetricsRadiusRejectLdapError,
			"ldap config incomplete: radius.LdapServer and radius.LdapBaseDN are required",
		)
	}

	filterTpl := strings.TrimSpace(cfgMgr.GetString("radius", "LdapUserFilter"))
	if filterTpl == "" {
		filterTpl = "(uid={username})"
	}

	username := ""
	if authCtx.User != nil {
		username = strings.TrimSpace(authCtx.User.Username)
	}
	if username == "" && authCtx.Request != nil && authCtx.Request.Packet != nil {
		username = strings.TrimSpace(rfc2865.UserName_GetString(authCtx.Request.Packet))
	}
	if username == "" {
		return true, radiuserrors.NewAuthError(app.MetricsRadiusRejectLdapError, "ldap auth failed: username is empty")
	}

	timeoutSeconds := defaultLDAPTimeoutSeconds
	if raw := strings.TrimSpace(cfgMgr.GetString("radius", "LdapTimeoutSeconds")); raw != "" {
		if parsed, err := strconv.ParseInt(raw, 10, 64); err == nil && parsed > 0 {
			timeoutSeconds = parsed
		}
	}

	client, err := newLDAPClient(serverURL, time.Duration(timeoutSeconds)*time.Second)
	if err != nil {
		return true, radiuserrors.NewAuthErrorWithCause(app.MetricsRadiusRejectLdapError, "ldap connect failed", err)
	}
	defer func() { _ = client.Close() }()

	bindDN := strings.TrimSpace(cfgMgr.GetString("radius", "LdapBindDN"))
	bindPassword := cfgMgr.GetString("radius", "LdapBindPassword")
	if bindDN != "" {
		if err := client.Bind(bindDN, bindPassword); err != nil {
			return true, radiuserrors.NewAuthErrorWithCause(app.MetricsRadiusRejectLdapError, "ldap bind account auth failed", err)
		}
	}

	filter := strings.ReplaceAll(filterTpl, "{username}", ldap.EscapeFilter(username))
	searchReq := ldap.NewSearchRequest(
		baseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		1,
		0,
		false,
		filter,
		[]string{"dn"},
		nil,
	)
	searchResult, err := client.Search(searchReq)
	if err != nil {
		return true, radiuserrors.NewAuthErrorWithCause(app.MetricsRadiusRejectLdapError, "ldap search failed", err)
	}
	if len(searchResult.Entries) != 1 {
		return true, radiuserrors.NewAuthError(
			app.MetricsRadiusRejectLdapError,
			fmt.Sprintf("ldap user search returned %d entries", len(searchResult.Entries)),
		)
	}

	if err := client.Bind(searchResult.Entries[0].DN, requestPassword); err != nil {
		return true, radiuserrors.NewAuthErrorWithCause(app.MetricsRadiusRejectLdapError, "ldap user auth failed", err)
	}

	return true, nil
}

// isLDAPEnabled parses configurable LDAP enable values.
// Accepted true values are: "true", "1", "enabled", "yes", "on" (case-insensitive).
func isLDAPEnabled(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true", "1", "enabled", "yes", "on":
		return true
	default:
		return false
	}
}
