package domain

import (
	"go.uber.org/zap"
)

// Profile link mode constants
const (
	ProfileLinkModeStatic  = 0 // Static mode: user attributes are snapshot from profile at creation
	ProfileLinkModeDynamic = 1 // Dynamic mode: user attributes are fetched from profile in real-time
)

// ProfileCacheGetter defines interface for profile cache to avoid circular dependency
type ProfileCacheGetter interface {
	Get(profileID int64) (*RadiusProfile, error)
}

// GetUpRate returns the upload rate, respecting profile link mode
// In dynamic mode, fetches from profile; in static mode, returns user's own value
// User-specific overrides (non-zero values) always take precedence
// cache parameter should be a ProfileCacheGetter implementation (e.g., *app.ProfileCache)
func (u *RadiusUser) GetUpRate(cache interface{}) int {
	// User-specific override has highest priority
	if u.UpRate > 0 {
		return u.UpRate
	}

	// Dynamic mode: fetch from profile
	if u.ProfileLinkMode == ProfileLinkModeDynamic && cache != nil {
		if cacheGetter, ok := cache.(ProfileCacheGetter); ok {
			profile, err := cacheGetter.Get(u.ProfileId)
			if err != nil {
				zap.L().Error("failed to get profile from cache",
					zap.Int64("user_id", u.ID),
					zap.Int64("profile_id", u.ProfileId),
					zap.Error(err))
				return u.UpRate // Fallback to user's stored value
			}
			return profile.UpRate
		}
	}

	// Static mode: return user's stored value
	return u.UpRate
}

// GetDownRate returns the download rate, respecting profile link mode
func (u *RadiusUser) GetDownRate(cache interface{}) int {
	// User-specific override has highest priority
	if u.DownRate > 0 {
		return u.DownRate
	}

	// Dynamic mode: fetch from profile
	if u.ProfileLinkMode == ProfileLinkModeDynamic && cache != nil {
		if cacheGetter, ok := cache.(ProfileCacheGetter); ok {
			profile, err := cacheGetter.Get(u.ProfileId)
			if err != nil {
				zap.L().Error("failed to get profile from cache",
					zap.Int64("user_id", u.ID),
					zap.Int64("profile_id", u.ProfileId),
					zap.Error(err))
				return u.DownRate
			}
			return profile.DownRate
		}
	}

	return u.DownRate
}

// GetActiveNum returns the max concurrent sessions, respecting profile link mode
func (u *RadiusUser) GetActiveNum(cache interface{}) int {
	// User-specific override has highest priority
	if u.ActiveNum > 0 {
		return u.ActiveNum
	}

	// Dynamic mode: fetch from profile
	if u.ProfileLinkMode == ProfileLinkModeDynamic && cache != nil {
		if cacheGetter, ok := cache.(ProfileCacheGetter); ok {
			profile, err := cacheGetter.Get(u.ProfileId)
			if err != nil {
				zap.L().Error("failed to get profile from cache",
					zap.Int64("user_id", u.ID),
					zap.Int64("profile_id", u.ProfileId),
					zap.Error(err))
				return u.ActiveNum
			}
			return profile.ActiveNum
		}
	}

	return u.ActiveNum
}

// GetAddrPool returns the address pool name, respecting profile link mode
func (u *RadiusUser) GetAddrPool(cache interface{}) string {
	// User-specific override has highest priority (non-empty string)
	if u.AddrPool != "" && u.AddrPool != "NA" {
		return u.AddrPool
	}

	// Dynamic mode: fetch from profile
	if u.ProfileLinkMode == ProfileLinkModeDynamic && cache != nil {
		if cacheGetter, ok := cache.(ProfileCacheGetter); ok {
			profile, err := cacheGetter.Get(u.ProfileId)
			if err != nil {
				zap.L().Error("failed to get profile from cache",
					zap.Int64("user_id", u.ID),
					zap.Int64("profile_id", u.ProfileId),
					zap.Error(err))
				return u.AddrPool
			}
			return profile.AddrPool
		}
	}

	return u.AddrPool
}

// GetDomain returns the domain attribute, respecting profile link mode
func (u *RadiusUser) GetDomain(cache interface{}) string {
	// User-specific override has highest priority
	if u.Domain != "" && u.Domain != "NA" {
		return u.Domain
	}

	// Dynamic mode: fetch from profile
	if u.ProfileLinkMode == ProfileLinkModeDynamic && cache != nil {
		if cacheGetter, ok := cache.(ProfileCacheGetter); ok {
			profile, err := cacheGetter.Get(u.ProfileId)
			if err != nil {
				zap.L().Error("failed to get profile from cache",
					zap.Int64("user_id", u.ID),
					zap.Int64("profile_id", u.ProfileId),
					zap.Error(err))
				return u.Domain
			}
			return profile.Domain
		}
	}

	return u.Domain
}

// GetIPv6PrefixPool returns the IPv6 prefix pool name, respecting profile link mode
func (u *RadiusUser) GetIPv6PrefixPool(cache interface{}) string {
	// User-specific override has highest priority
	if u.IPv6PrefixPool != "" && u.IPv6PrefixPool != "NA" {
		return u.IPv6PrefixPool
	}

	// Dynamic mode: fetch from profile
	if u.ProfileLinkMode == ProfileLinkModeDynamic && cache != nil {
		if cacheGetter, ok := cache.(ProfileCacheGetter); ok {
			profile, err := cacheGetter.Get(u.ProfileId)
			if err != nil {
				zap.L().Error("failed to get profile from cache",
					zap.Int64("user_id", u.ID),
					zap.Int64("profile_id", u.ProfileId),
					zap.Error(err))
				return u.IPv6PrefixPool
			}
			return profile.IPv6PrefixPool
		}
	}

	return u.IPv6PrefixPool
}

// GetBindMac returns the MAC binding flag, respecting profile link mode
// Note: User-specific binding always applies (e.g., specific MAC address in MacAddr field)
func (u *RadiusUser) GetBindMac(cache interface{}) int {
	// User-specific override has priority
	if u.BindMac > 0 {
		return u.BindMac
	}

	// If user explicitly disabled binding (BindMac=0), respect that even if MAC is configured
	if u.BindMac == 0 && u.MacAddr != "" && u.MacAddr != "NA" {
		// User has MAC configured but explicitly disabled binding
		return 0
	}

	// Dynamic mode: fetch from profile
	if u.ProfileLinkMode == ProfileLinkModeDynamic && cache != nil {
		if cacheGetter, ok := cache.(ProfileCacheGetter); ok {
			profile, err := cacheGetter.Get(u.ProfileId)
			if err != nil {
				zap.L().Error("failed to get profile from cache",
					zap.Int64("user_id", u.ID),
					zap.Int64("profile_id", u.ProfileId),
					zap.Error(err))
				return u.BindMac
			}
			// If profile binding is enabled and user has specific MAC, enforce binding
			if profile.BindMac > 0 && u.MacAddr != "" && u.MacAddr != "NA" {
				return 1
			}
			return profile.BindMac
		}
	}

	// Static mode: if user has specific MAC but no explicit binding flag, enforce binding
	if u.MacAddr != "" && u.MacAddr != "NA" {
		return 1
	}

	return u.BindMac
}

// GetBindVlan returns the VLAN binding flag, respecting profile link mode
func (u *RadiusUser) GetBindVlan(cache interface{}) int {
	// User-specific override has priority
	if u.BindVlan > 0 {
		return u.BindVlan
	}

	// If user explicitly disabled binding (BindVlan=0), respect that even if VLANs are configured
	if u.BindVlan == 0 && (u.Vlanid1 > 0 || u.Vlanid2 > 0) {
		// User has VLANs configured but explicitly disabled binding
		return 0
	}

	// Dynamic mode: fetch from profile
	if u.ProfileLinkMode == ProfileLinkModeDynamic && cache != nil {
		if cacheGetter, ok := cache.(ProfileCacheGetter); ok {
			profile, err := cacheGetter.Get(u.ProfileId)
			if err != nil {
				zap.L().Error("failed to get profile from cache",
					zap.Int64("user_id", u.ID),
					zap.Int64("profile_id", u.ProfileId),
					zap.Error(err))
				return u.BindVlan
			}
			// If profile binding is enabled and user has specific VLANs, enforce binding
			if profile.BindVlan > 0 && (u.Vlanid1 > 0 || u.Vlanid2 > 0) {
				return 1
			}
			return profile.BindVlan
		}
	}

	// Static mode: if user has specific VLANs but no explicit binding flag, enforce binding
	if u.Vlanid1 > 0 || u.Vlanid2 > 0 {
		return 1
	}

	return u.BindVlan
}
