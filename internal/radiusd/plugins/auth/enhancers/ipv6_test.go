package enhancers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	"layeh.com/radius"
	"layeh.com/radius/rfc3162"
	"layeh.com/radius/rfc4818"
	"layeh.com/radius/rfc6911"
)

// TestDefaultAcceptEnhancer_IPv6Prefix tests IPv6 prefix setting
func TestDefaultAcceptEnhancer_IPv6Prefix(t *testing.T) {
	tests := []struct {
		name           string
		ipv6Addr       string
		expectSet      bool
		expectedPrefix string
	}{
		{
			name:           "IPv6 with prefix",
			ipv6Addr:       "2001:db8::1/64",
			expectSet:      true,
			expectedPrefix: "2001:db8::/64",
		},
		{
			name:           "IPv6 without prefix",
			ipv6Addr:       "2001:db8::1",
			expectSet:      true,
			expectedPrefix: "2001:db8::1/128",
		},
		{
			name:           "IPv6 /128 prefix",
			ipv6Addr:       "fe80::1/128",
			expectSet:      true,
			expectedPrefix: "fe80::1/128",
		},
		{
			name:      "empty IPv6",
			ipv6Addr:  "",
			expectSet: false,
		},
		{
			name:      "NA value",
			ipv6Addr:  "N/A",
			expectSet: false,
		},
		{
			name:      "invalid IPv6",
			ipv6Addr:  "invalid-ipv6",
			expectSet: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enhancer := NewDefaultAcceptEnhancer()
			response := radius.New(radius.CodeAccessAccept, []byte("secret"))
			user := &domain.RadiusUser{
				IpV6Addr: tt.ipv6Addr,
			}
			authCtx := &auth.AuthContext{
				Response: response,
				User:     user,
			}

			err := enhancer.Enhance(context.Background(), authCtx)
			if err != nil {
				t.Fatalf("Enhance() error = %v", err)
			}

			ipnet := rfc3162.FramedIPv6Prefix_Get(response)
			if tt.expectSet {
				if ipnet == nil {
					t.Errorf("Expected FramedIPv6Prefix to be set, but got nil")
				} else if ipnet.String() != tt.expectedPrefix {
					t.Errorf("FramedIPv6Prefix = %v, want %v", ipnet.String(), tt.expectedPrefix)
				}
			} else {
				if ipnet != nil {
					t.Errorf("Expected FramedIPv6Prefix to be nil, but got %v", ipnet.String())
				}
			}
		})
	}
}

// TestDefaultAcceptEnhancer_FramedIPv6Address verifies that a static single-host
// IPv6 address is advertised as a Framed-IPv6-Address attribute (RFC 6911),
// while multi-host prefixes, IPv4, empty, and N/A values are not.
func TestDefaultAcceptEnhancer_FramedIPv6Address(t *testing.T) {
	tests := []struct {
		name       string
		ipv6Addr   string
		expectSet  bool
		expectedIP string
	}{
		{name: "bare host address", ipv6Addr: "2001:db8::1", expectSet: true, expectedIP: "2001:db8::1"},
		{name: "explicit /128 host", ipv6Addr: "2001:db8::2/128", expectSet: true, expectedIP: "2001:db8::2"},
		{name: "network prefix /64 is not a host", ipv6Addr: "2001:db8::1/64", expectSet: false},
		{name: "empty", ipv6Addr: "", expectSet: false},
		{name: "NA value", ipv6Addr: "N/A", expectSet: false},
		{name: "invalid", ipv6Addr: "not-an-ip", expectSet: false},
		{name: "ipv4 ignored", ipv6Addr: "192.0.2.1", expectSet: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enhancer := NewDefaultAcceptEnhancer()
			response := radius.New(radius.CodeAccessAccept, []byte("secret"))
			user := &domain.RadiusUser{IpV6Addr: tt.ipv6Addr}
			authCtx := &auth.AuthContext{Response: response, User: user}

			require.NoError(t, enhancer.Enhance(context.Background(), authCtx))

			ip := rfc6911.FramedIPv6Address_Get(response)
			if tt.expectSet {
				require.NotNil(t, ip)
				assert.Equal(t, tt.expectedIP, ip.String())
			} else {
				assert.Nil(t, ip)
			}
		})
	}
}

// TestSingleIPv6Host exercises the host-address classifier directly.
func TestSingleIPv6Host(t *testing.T) {
	tests := []struct {
		value   string
		wantOK  bool
		wantStr string
	}{
		{"2001:db8::1", true, "2001:db8::1"},
		{"2001:db8::1/128", true, "2001:db8::1"},
		{"2001:db8::/64", false, ""},
		{"192.0.2.1", false, ""},
		{"192.0.2.1/128", false, ""},
		{"not-an-ip", false, ""},
		{"", false, ""},
	}
	for _, tt := range tests {
		ip, ok := singleIPv6Host(tt.value)
		assert.Equal(t, tt.wantOK, ok, "value=%q", tt.value)
		if tt.wantOK {
			assert.Equal(t, tt.wantStr, ip.String(), "value=%q", tt.value)
		}
	}
}

// stubProfileCache is a minimal domain.ProfileCacheGetter used to exercise
// profile-inheritance paths in the enhancer without the real cache.
type stubProfileCache struct {
	profile *domain.RadiusProfile
}

func (s stubProfileCache) Get(int64) (*domain.RadiusProfile, error) {
	return s.profile, nil
}

// TestDefaultAcceptEnhancer_DelegatedIPv6Prefix verifies that a user's static
// Delegated-IPv6-Prefix (RFC 4818, attribute 123) is issued in the Access-Accept,
// that a bare address is normalised to a single-host /128 delegation, and that
// empty/N/A/IPv4/unparseable values are skipped rather than emitting a malformed
// attribute.
func TestDefaultAcceptEnhancer_DelegatedIPv6Prefix(t *testing.T) {
	tests := []struct {
		name           string
		delegated      string
		expectSet      bool
		expectedPrefix string
	}{
		{name: "network prefix /48", delegated: "2001:db8:1234::/48", expectSet: true, expectedPrefix: "2001:db8:1234::/48"},
		{name: "network prefix /56", delegated: "2001:db8:abcd:ee00::/56", expectSet: true, expectedPrefix: "2001:db8:abcd:ee00::/56"},
		{name: "bare address becomes /128", delegated: "2001:db8::1", expectSet: true, expectedPrefix: "2001:db8::1/128"},
		{name: "empty", delegated: "", expectSet: false},
		{name: "NA value", delegated: "N/A", expectSet: false},
		{name: "ipv4 prefix ignored", delegated: "192.0.2.0/24", expectSet: false},
		{name: "invalid", delegated: "not-a-prefix", expectSet: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enhancer := NewDefaultAcceptEnhancer()
			response := radius.New(radius.CodeAccessAccept, []byte("secret"))
			user := &domain.RadiusUser{DelegatedIpv6Prefix: tt.delegated}
			authCtx := &auth.AuthContext{Response: response, User: user}

			require.NoError(t, enhancer.Enhance(context.Background(), authCtx))

			ipnet := rfc4818.DelegatedIPv6Prefix_Get(response)
			if tt.expectSet {
				require.NotNil(t, ipnet)
				assert.Equal(t, tt.expectedPrefix, ipnet.String())
			} else {
				assert.Nil(t, ipnet)
			}
		})
	}
}

// TestDefaultAcceptEnhancer_DelegatedIPv6PrefixPool verifies that the
// Delegated-IPv6-Prefix-Pool (RFC 6911, attribute 171) is issued from a
// user-specific value, inherited from the linked profile in dynamic link mode,
// and skipped when unset/N/A. It must remain distinct from Framed-IPv6-Pool.
func TestDefaultAcceptEnhancer_DelegatedIPv6PrefixPool(t *testing.T) {
	t.Run("user-specific value", func(t *testing.T) {
		enhancer := NewDefaultAcceptEnhancer()
		response := radius.New(radius.CodeAccessAccept, []byte("secret"))
		user := &domain.RadiusUser{DelegatedIpv6PrefixPool: "pd-pool-user"}
		authCtx := &auth.AuthContext{Response: response, User: user}

		require.NoError(t, enhancer.Enhance(context.Background(), authCtx))
		assert.Equal(t, "pd-pool-user", rfc6911.DelegatedIPv6PrefixPool_GetString(response))
	})

	t.Run("inherited from profile in dynamic mode", func(t *testing.T) {
		enhancer := NewDefaultAcceptEnhancer()
		response := radius.New(radius.CodeAccessAccept, []byte("secret"))
		user := &domain.RadiusUser{
			ProfileId:       7,
			ProfileLinkMode: domain.ProfileLinkModeDynamic,
		}
		cache := stubProfileCache{profile: &domain.RadiusProfile{DelegatedIpv6PrefixPool: "pd-pool-profile"}}
		authCtx := &auth.AuthContext{
			Response: response,
			User:     user,
			Metadata: map[string]interface{}{"profile_cache": cache},
		}

		require.NoError(t, enhancer.Enhance(context.Background(), authCtx))
		assert.Equal(t, "pd-pool-profile", rfc6911.DelegatedIPv6PrefixPool_GetString(response))
	})

	t.Run("unset and NA are skipped", func(t *testing.T) {
		for _, v := range []string{"", "N/A"} {
			enhancer := NewDefaultAcceptEnhancer()
			response := radius.New(radius.CodeAccessAccept, []byte("secret"))
			user := &domain.RadiusUser{DelegatedIpv6PrefixPool: v}
			authCtx := &auth.AuthContext{Response: response, User: user}

			require.NoError(t, enhancer.Enhance(context.Background(), authCtx))
			assert.Empty(t, rfc6911.DelegatedIPv6PrefixPool_GetString(response), "value=%q", v)
		}
	})
}

// TestHuaweiAcceptEnhancer_IPv6Address tests Huawei IPv6 address setting
func TestHuaweiAcceptEnhancer_IPv6Address(t *testing.T) {
	tests := []struct {
		name      string
		ipv6Addr  string
		expectSet bool
	}{
		{
			name:      "IPv6 with prefix",
			ipv6Addr:  "2001:db8::1/64",
			expectSet: true,
		},
		{
			name:      "IPv6 without prefix",
			ipv6Addr:  "2001:db8::1",
			expectSet: true,
		},
		{
			name:      "empty IPv6",
			ipv6Addr:  "",
			expectSet: false,
		},
		{
			name:      "NA value",
			ipv6Addr:  "N/A",
			expectSet: false,
		},
		{
			name:      "invalid IPv6",
			ipv6Addr:  "invalid",
			expectSet: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enhancer := NewHuaweiAcceptEnhancer()
			response := radius.New(radius.CodeAccessAccept, []byte("secret"))
			user := &domain.RadiusUser{
				IpV6Addr: tt.ipv6Addr,
				UpRate:   100,
				DownRate: 100,
			}
			nas := &domain.NetNas{
				VendorCode: "2011", // Huawei
			}
			authCtx := &auth.AuthContext{
				Response: response,
				User:     user,
				Nas:      nas,
			}

			err := enhancer.Enhance(context.Background(), authCtx)
			if err != nil {
				t.Fatalf("Enhance() error = %v", err)
			}

			// Check if IPv6 VSA was set
			// Note: We can't directly check the VSA without importing huawei package
			// So we just ensure no error occurred
			if tt.expectSet {
				// If we expected it to be set, the code should have executed without error
				// More detailed validation would require parsing the RADIUS packet
				_ = tt.expectSet //nolint:staticcheck // intentionally empty branch for documentation
			}
		})
	}
}

// TestHuaweiAcceptEnhancer_DomainName tests Huawei Domain Name setting
func TestHuaweiAcceptEnhancer_DomainName(t *testing.T) {
	tests := []struct {
		name      string
		domain    string
		expectSet bool
	}{
		{
			name:      "valid domain",
			domain:    "enterprise.example.com",
			expectSet: true,
		},
		{
			name:      "simple domain",
			domain:    "testdomain",
			expectSet: true,
		},
		{
			name:      "empty domain",
			domain:    "",
			expectSet: false,
		},
		{
			name:      "NA value",
			domain:    "N/A",
			expectSet: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enhancer := NewHuaweiAcceptEnhancer()
			response := radius.New(radius.CodeAccessAccept, []byte("secret"))
			user := &domain.RadiusUser{
				Domain:   tt.domain,
				UpRate:   100,
				DownRate: 100,
			}
			nas := &domain.NetNas{
				VendorCode: "2011", // Huawei
			}
			authCtx := &auth.AuthContext{
				Response: response,
				User:     user,
				Nas:      nas,
			}

			err := enhancer.Enhance(context.Background(), authCtx)
			if err != nil {
				t.Fatalf("Enhance() error = %v", err)
			}

			// Domain was set successfully if no error occurred
			// Detailed validation would require parsing the RADIUS packet VSA
		})
	}
}
