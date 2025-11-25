package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockProfileCache implements ProfileCacheGetter for testing
type mockProfileCache struct {
	profiles map[int64]*RadiusProfile
	err      error
}

func newMockCache() *mockProfileCache {
	return &mockProfileCache{
		profiles: make(map[int64]*RadiusProfile),
	}
}

func (m *mockProfileCache) Get(profileID int64) (*RadiusProfile, error) {
	if m.err != nil {
		return nil, m.err
	}
	profile, ok := m.profiles[profileID]
	if !ok {
		return nil, errors.New("profile not found")
	}
	return profile, nil
}

func (m *mockProfileCache) SetProfile(id int64, profile *RadiusProfile) {
	m.profiles[id] = profile
}

func (m *mockProfileCache) SetError(err error) {
	m.err = err
}

func TestGetUpRate(t *testing.T) {
	cache := newMockCache()
	cache.SetProfile(1, &RadiusProfile{
		ID:     1,
		UpRate: 10240,
	})

	tests := []struct {
		name     string
		user     *RadiusUser
		cache    interface{}
		expected int
	}{
		{
			name: "user override takes priority",
			user: &RadiusUser{
				UpRate:          20480,
				ProfileId:       1,
				ProfileLinkMode: ProfileLinkModeDynamic,
			},
			cache:    cache,
			expected: 20480,
		},
		{
			name: "dynamic mode fetches from profile",
			user: &RadiusUser{
				UpRate:          0,
				ProfileId:       1,
				ProfileLinkMode: ProfileLinkModeDynamic,
			},
			cache:    cache,
			expected: 10240,
		},
		{
			name: "static mode returns user value",
			user: &RadiusUser{
				UpRate:          5120,
				ProfileId:       1,
				ProfileLinkMode: ProfileLinkModeStatic,
			},
			cache:    cache,
			expected: 5120,
		},
		{
			name: "static mode with zero returns zero",
			user: &RadiusUser{
				UpRate:          0,
				ProfileId:       1,
				ProfileLinkMode: ProfileLinkModeStatic,
			},
			cache:    cache,
			expected: 0,
		},
		{
			name: "nil cache returns user value",
			user: &RadiusUser{
				UpRate:          1024,
				ProfileId:       1,
				ProfileLinkMode: ProfileLinkModeDynamic,
			},
			cache:    nil,
			expected: 1024,
		},
		{
			name: "invalid cache type returns user value",
			user: &RadiusUser{
				UpRate:          2048,
				ProfileId:       1,
				ProfileLinkMode: ProfileLinkModeDynamic,
			},
			cache:    "invalid",
			expected: 2048,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.user.GetUpRate(tt.cache)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetUpRate_CacheError(t *testing.T) {
	cache := newMockCache()
	cache.SetError(errors.New("cache error"))

	user := &RadiusUser{
		UpRate:          5000,
		ProfileId:       1,
		ProfileLinkMode: ProfileLinkModeDynamic,
	}

	// Should fallback to user value on cache error
	result := user.GetUpRate(cache)
	assert.Equal(t, 5000, result)
}

func TestGetDownRate(t *testing.T) {
	cache := newMockCache()
	cache.SetProfile(1, &RadiusProfile{
		ID:       1,
		DownRate: 20480,
	})

	tests := []struct {
		name     string
		user     *RadiusUser
		cache    interface{}
		expected int
	}{
		{
			name: "user override takes priority",
			user: &RadiusUser{
				DownRate:        40960,
				ProfileId:       1,
				ProfileLinkMode: ProfileLinkModeDynamic,
			},
			cache:    cache,
			expected: 40960,
		},
		{
			name: "dynamic mode fetches from profile",
			user: &RadiusUser{
				DownRate:        0,
				ProfileId:       1,
				ProfileLinkMode: ProfileLinkModeDynamic,
			},
			cache:    cache,
			expected: 20480,
		},
		{
			name: "static mode returns user value",
			user: &RadiusUser{
				DownRate:        10240,
				ProfileId:       1,
				ProfileLinkMode: ProfileLinkModeStatic,
			},
			cache:    cache,
			expected: 10240,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.user.GetDownRate(tt.cache)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetActiveNum(t *testing.T) {
	cache := newMockCache()
	cache.SetProfile(1, &RadiusProfile{
		ID:        1,
		ActiveNum: 5,
	})

	tests := []struct {
		name     string
		user     *RadiusUser
		cache    interface{}
		expected int
	}{
		{
			name: "user override takes priority",
			user: &RadiusUser{
				ActiveNum:       10,
				ProfileId:       1,
				ProfileLinkMode: ProfileLinkModeDynamic,
			},
			cache:    cache,
			expected: 10,
		},
		{
			name: "dynamic mode fetches from profile",
			user: &RadiusUser{
				ActiveNum:       0,
				ProfileId:       1,
				ProfileLinkMode: ProfileLinkModeDynamic,
			},
			cache:    cache,
			expected: 5,
		},
		{
			name: "static mode returns user value",
			user: &RadiusUser{
				ActiveNum:       3,
				ProfileId:       1,
				ProfileLinkMode: ProfileLinkModeStatic,
			},
			cache:    cache,
			expected: 3,
		},
		{
			name: "zero means unlimited in static mode",
			user: &RadiusUser{
				ActiveNum:       0,
				ProfileId:       1,
				ProfileLinkMode: ProfileLinkModeStatic,
			},
			cache:    cache,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.user.GetActiveNum(tt.cache)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetAddrPool(t *testing.T) {
	cache := newMockCache()
	cache.SetProfile(1, &RadiusProfile{
		ID:       1,
		AddrPool: "pool-from-profile",
	})

	tests := []struct {
		name     string
		user     *RadiusUser
		cache    interface{}
		expected string
	}{
		{
			name: "user override takes priority",
			user: &RadiusUser{
				AddrPool:        "user-pool",
				ProfileId:       1,
				ProfileLinkMode: ProfileLinkModeDynamic,
			},
			cache:    cache,
			expected: "user-pool",
		},
		{
			name: "dynamic mode fetches from profile",
			user: &RadiusUser{
				AddrPool:        "",
				ProfileId:       1,
				ProfileLinkMode: ProfileLinkModeDynamic,
			},
			cache:    cache,
			expected: "pool-from-profile",
		},
		{
			name: "NA treated as empty",
			user: &RadiusUser{
				AddrPool:        "NA",
				ProfileId:       1,
				ProfileLinkMode: ProfileLinkModeDynamic,
			},
			cache:    cache,
			expected: "pool-from-profile",
		},
		{
			name: "static mode returns user value",
			user: &RadiusUser{
				AddrPool:        "static-pool",
				ProfileId:       1,
				ProfileLinkMode: ProfileLinkModeStatic,
			},
			cache:    cache,
			expected: "static-pool",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.user.GetAddrPool(tt.cache)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetDomain(t *testing.T) {
	cache := newMockCache()
	cache.SetProfile(1, &RadiusProfile{
		ID:     1,
		Domain: "profile.domain.com",
	})

	tests := []struct {
		name     string
		user     *RadiusUser
		cache    interface{}
		expected string
	}{
		{
			name: "user override takes priority",
			user: &RadiusUser{
				Domain:          "user.domain.com",
				ProfileId:       1,
				ProfileLinkMode: ProfileLinkModeDynamic,
			},
			cache:    cache,
			expected: "user.domain.com",
		},
		{
			name: "dynamic mode fetches from profile",
			user: &RadiusUser{
				Domain:          "",
				ProfileId:       1,
				ProfileLinkMode: ProfileLinkModeDynamic,
			},
			cache:    cache,
			expected: "profile.domain.com",
		},
		{
			name: "NA treated as empty",
			user: &RadiusUser{
				Domain:          "NA",
				ProfileId:       1,
				ProfileLinkMode: ProfileLinkModeDynamic,
			},
			cache:    cache,
			expected: "profile.domain.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.user.GetDomain(tt.cache)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetIPv6PrefixPool(t *testing.T) {
	cache := newMockCache()
	cache.SetProfile(1, &RadiusProfile{
		ID:             1,
		IPv6PrefixPool: "ipv6-pool-from-profile",
	})

	tests := []struct {
		name     string
		user     *RadiusUser
		cache    interface{}
		expected string
	}{
		{
			name: "user override takes priority",
			user: &RadiusUser{
				IPv6PrefixPool:  "user-ipv6-pool",
				ProfileId:       1,
				ProfileLinkMode: ProfileLinkModeDynamic,
			},
			cache:    cache,
			expected: "user-ipv6-pool",
		},
		{
			name: "dynamic mode fetches from profile",
			user: &RadiusUser{
				IPv6PrefixPool:  "",
				ProfileId:       1,
				ProfileLinkMode: ProfileLinkModeDynamic,
			},
			cache:    cache,
			expected: "ipv6-pool-from-profile",
		},
		{
			name: "NA treated as empty",
			user: &RadiusUser{
				IPv6PrefixPool:  "NA",
				ProfileId:       1,
				ProfileLinkMode: ProfileLinkModeDynamic,
			},
			cache:    cache,
			expected: "ipv6-pool-from-profile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.user.GetIPv6PrefixPool(tt.cache)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetBindMac(t *testing.T) {
	cache := newMockCache()
	cache.SetProfile(1, &RadiusProfile{
		ID:      1,
		BindMac: 1,
	})
	cache.SetProfile(2, &RadiusProfile{
		ID:      2,
		BindMac: 0,
	})

	tests := []struct {
		name     string
		user     *RadiusUser
		cache    interface{}
		expected int
	}{
		{
			name: "user BindMac=1 takes priority",
			user: &RadiusUser{
				BindMac:         1,
				MacAddr:         "",
				ProfileId:       1,
				ProfileLinkMode: ProfileLinkModeDynamic,
			},
			cache:    cache,
			expected: 1,
		},
		{
			name: "user BindMac=0 with MAC configured explicitly disables",
			user: &RadiusUser{
				BindMac:         0,
				MacAddr:         "00:11:22:33:44:55",
				ProfileId:       1,
				ProfileLinkMode: ProfileLinkModeDynamic,
			},
			cache:    cache,
			expected: 0,
		},
		{
			name: "dynamic mode with profile BindMac=1 and user MAC enforces binding",
			user: &RadiusUser{
				BindMac:         0,
				MacAddr:         "00:11:22:33:44:55",
				ProfileId:       1,
				ProfileLinkMode: ProfileLinkModeDynamic,
			},
			cache:    cache,
			expected: 0, // User explicitly disabled
		},
		{
			name: "dynamic mode fetches from profile when no MAC",
			user: &RadiusUser{
				BindMac:         0,
				MacAddr:         "",
				ProfileId:       1,
				ProfileLinkMode: ProfileLinkModeDynamic,
			},
			cache:    cache,
			expected: 1,
		},
		{
			name: "static mode with MAC but no BindMac enforces binding",
			user: &RadiusUser{
				BindMac:         0,
				MacAddr:         "00:11:22:33:44:55",
				ProfileId:       1,
				ProfileLinkMode: ProfileLinkModeStatic,
			},
			cache:    cache,
			expected: 0, // User explicitly disabled
		},
		{
			name: "static mode with no MAC returns user BindMac",
			user: &RadiusUser{
				BindMac:         0,
				MacAddr:         "",
				ProfileId:       1,
				ProfileLinkMode: ProfileLinkModeStatic,
			},
			cache:    cache,
			expected: 0,
		},
		{
			name: "NA MAC treated as empty",
			user: &RadiusUser{
				BindMac:         0,
				MacAddr:         "NA",
				ProfileId:       1,
				ProfileLinkMode: ProfileLinkModeStatic,
			},
			cache:    cache,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.user.GetBindMac(tt.cache)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetBindVlan(t *testing.T) {
	cache := newMockCache()
	cache.SetProfile(1, &RadiusProfile{
		ID:       1,
		BindVlan: 1,
	})
	cache.SetProfile(2, &RadiusProfile{
		ID:       2,
		BindVlan: 0,
	})

	tests := []struct {
		name     string
		user     *RadiusUser
		cache    interface{}
		expected int
	}{
		{
			name: "user BindVlan=1 takes priority",
			user: &RadiusUser{
				BindVlan:        1,
				Vlanid1:         0,
				Vlanid2:         0,
				ProfileId:       1,
				ProfileLinkMode: ProfileLinkModeDynamic,
			},
			cache:    cache,
			expected: 1,
		},
		{
			name: "user BindVlan=0 with VLAN configured explicitly disables",
			user: &RadiusUser{
				BindVlan:        0,
				Vlanid1:         100,
				Vlanid2:         200,
				ProfileId:       1,
				ProfileLinkMode: ProfileLinkModeDynamic,
			},
			cache:    cache,
			expected: 0,
		},
		{
			name: "dynamic mode fetches from profile when no VLAN",
			user: &RadiusUser{
				BindVlan:        0,
				Vlanid1:         0,
				Vlanid2:         0,
				ProfileId:       1,
				ProfileLinkMode: ProfileLinkModeDynamic,
			},
			cache:    cache,
			expected: 1,
		},
		{
			name: "static mode with VLAN but BindVlan=0 explicitly disables",
			user: &RadiusUser{
				BindVlan:        0,
				Vlanid1:         100,
				Vlanid2:         0,
				ProfileId:       1,
				ProfileLinkMode: ProfileLinkModeStatic,
			},
			cache:    cache,
			expected: 0,
		},
		{
			name: "static mode with no VLAN returns user BindVlan",
			user: &RadiusUser{
				BindVlan:        0,
				Vlanid1:         0,
				Vlanid2:         0,
				ProfileId:       1,
				ProfileLinkMode: ProfileLinkModeStatic,
			},
			cache:    cache,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.user.GetBindVlan(tt.cache)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetBindMac_DynamicModeWithProfileBinding(t *testing.T) {
	cache := newMockCache()
	cache.SetProfile(1, &RadiusProfile{
		ID:      1,
		BindMac: 1,
	})

	// User has MAC but no explicit BindMac, profile has BindMac=1
	// In dynamic mode with no explicit user BindMac setting, should fetch from profile
	user := &RadiusUser{
		BindMac:         0,
		MacAddr:         "",
		ProfileId:       1,
		ProfileLinkMode: ProfileLinkModeDynamic,
	}

	result := user.GetBindMac(cache)
	assert.Equal(t, 1, result) // Profile binding is enabled
}

func TestGetBindVlan_DynamicModeWithProfileBinding(t *testing.T) {
	cache := newMockCache()
	cache.SetProfile(1, &RadiusProfile{
		ID:       1,
		BindVlan: 1,
	})

	// User has no VLAN, profile has BindVlan=1
	user := &RadiusUser{
		BindVlan:        0,
		Vlanid1:         0,
		Vlanid2:         0,
		ProfileId:       1,
		ProfileLinkMode: ProfileLinkModeDynamic,
	}

	result := user.GetBindVlan(cache)
	assert.Equal(t, 1, result) // Profile binding is enabled
}

func TestProfileLinkModeConstants(t *testing.T) {
	assert.Equal(t, 0, ProfileLinkModeStatic)
	assert.Equal(t, 1, ProfileLinkModeDynamic)
}

func TestGetters_ProfileNotFound(t *testing.T) {
	cache := newMockCache()
	// Don't add profile to cache - simulates profile not found

	user := &RadiusUser{
		ID:              1,
		UpRate:          1024,
		DownRate:        2048,
		ActiveNum:       3,
		AddrPool:        "fallback-pool",
		Domain:          "fallback.domain",
		IPv6PrefixPool:  "fallback-ipv6",
		BindMac:         0,
		BindVlan:        0,
		ProfileId:       999, // Non-existent profile
		ProfileLinkMode: ProfileLinkModeDynamic,
	}

	// All getters should fall back to user values when profile not found
	assert.Equal(t, 1024, user.GetUpRate(cache))
	assert.Equal(t, 2048, user.GetDownRate(cache))
	assert.Equal(t, 3, user.GetActiveNum(cache))
	assert.Equal(t, "fallback-pool", user.GetAddrPool(cache))
	assert.Equal(t, "fallback.domain", user.GetDomain(cache))
	assert.Equal(t, "fallback-ipv6", user.GetIPv6PrefixPool(cache))
}

func TestGetters_AllAttributesFromProfile(t *testing.T) {
	cache := newMockCache()
	cache.SetProfile(1, &RadiusProfile{
		ID:             1,
		UpRate:         10240,
		DownRate:       20480,
		ActiveNum:      5,
		AddrPool:       "profile-pool",
		Domain:         "profile.domain.com",
		IPv6PrefixPool: "profile-ipv6-pool",
		BindMac:        1,
		BindVlan:       1,
	})

	// User in dynamic mode with all zero/empty values
	user := &RadiusUser{
		ID:              1,
		UpRate:          0,
		DownRate:        0,
		ActiveNum:       0,
		AddrPool:        "",
		Domain:          "",
		IPv6PrefixPool:  "",
		BindMac:         0,
		BindVlan:        0,
		MacAddr:         "",
		Vlanid1:         0,
		Vlanid2:         0,
		ProfileId:       1,
		ProfileLinkMode: ProfileLinkModeDynamic,
	}

	// All values should come from profile
	assert.Equal(t, 10240, user.GetUpRate(cache))
	assert.Equal(t, 20480, user.GetDownRate(cache))
	assert.Equal(t, 5, user.GetActiveNum(cache))
	assert.Equal(t, "profile-pool", user.GetAddrPool(cache))
	assert.Equal(t, "profile.domain.com", user.GetDomain(cache))
	assert.Equal(t, "profile-ipv6-pool", user.GetIPv6PrefixPool(cache))
	assert.Equal(t, 1, user.GetBindMac(cache))
	assert.Equal(t, 1, user.GetBindVlan(cache))
}

func TestGetters_MixedUserAndProfileValues(t *testing.T) {
	cache := newMockCache()
	cache.SetProfile(1, &RadiusProfile{
		ID:             1,
		UpRate:         10240,
		DownRate:       20480,
		ActiveNum:      5,
		AddrPool:       "profile-pool",
		Domain:         "profile.domain.com",
		IPv6PrefixPool: "profile-ipv6-pool",
		BindMac:        1,
		BindVlan:       1,
	})

	// User has some overrides
	user := &RadiusUser{
		ID:              1,
		UpRate:          5120,              // Override
		DownRate:        0,                 // From profile
		ActiveNum:       10,                // Override
		AddrPool:        "",                // From profile
		Domain:          "user.domain.com", // Override
		IPv6PrefixPool:  "",                // From profile
		BindMac:         0,
		BindVlan:        0,
		MacAddr:         "",
		Vlanid1:         0,
		Vlanid2:         0,
		ProfileId:       1,
		ProfileLinkMode: ProfileLinkModeDynamic,
	}

	// Mixed values - user overrides take priority
	assert.Equal(t, 5120, user.GetUpRate(cache))                        // User override
	assert.Equal(t, 20480, user.GetDownRate(cache))                     // From profile
	assert.Equal(t, 10, user.GetActiveNum(cache))                       // User override
	assert.Equal(t, "profile-pool", user.GetAddrPool(cache))            // From profile
	assert.Equal(t, "user.domain.com", user.GetDomain(cache))           // User override
	assert.Equal(t, "profile-ipv6-pool", user.GetIPv6PrefixPool(cache)) // From profile
}
