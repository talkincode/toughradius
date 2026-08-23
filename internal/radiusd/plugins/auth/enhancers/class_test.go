package enhancers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

func TestDefaultAcceptEnhancer_Class(t *testing.T) {
	tests := []struct {
		name          string
		userClass     string
		profileClass  string
		linkMode      int
		expectSet     bool
		expectedClass string
	}{
		{
			name:          "user class is emitted as-is",
			userClass:     "OU=vpn-users",
			expectSet:     true,
			expectedClass: "OU=vpn-users",
		},
		{
			name:          "dynamic profile class is inherited",
			profileClass:  "OU=profile-group;staff",
			linkMode:      domain.ProfileLinkModeDynamic,
			expectSet:     true,
			expectedClass: "OU=profile-group;staff",
		},
		{
			name:          "user class overrides dynamic profile",
			userClass:     "OU=user-group",
			profileClass:  "OU=profile-group",
			linkMode:      domain.ProfileLinkModeDynamic,
			expectSet:     true,
			expectedClass: "OU=user-group",
		},
		{
			name:         "static empty class is omitted",
			userClass:    "",
			profileClass: "OU=profile-group",
			linkMode:     domain.ProfileLinkModeStatic,
			expectSet:    false,
		},
		{
			name:      "empty class is omitted",
			userClass: "",
			expectSet: false,
		},
		{
			name:      "N/A class is omitted",
			userClass: "N/A",
			expectSet: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enhancer := NewDefaultAcceptEnhancer()
			response := radius.New(radius.CodeAccessAccept, []byte("secret"))
			user := &domain.RadiusUser{
				RadiusClass:     tt.userClass,
				ProfileId:       1,
				ProfileLinkMode: tt.linkMode,
			}
			authCtx := &auth.AuthContext{
				Response: response,
				User:     user,
			}
			if tt.profileClass != "" {
				cache := &staticProfileCache{profile: &domain.RadiusProfile{
					ID:          1,
					RadiusClass: tt.profileClass,
				}}
				authCtx.Metadata = map[string]interface{}{
					"profile_cache": cache,
				}
			}

			require.NoError(t, enhancer.Enhance(context.Background(), authCtx))

			got := rfc2865.Class_GetString(response)
			if tt.expectSet {
				assert.Equal(t, tt.expectedClass, got)
			} else {
				assert.Empty(t, got)
			}
		})
	}
}

type staticProfileCache struct {
	profile *domain.RadiusProfile
}

func (c *staticProfileCache) Get(profileID int64) (*domain.RadiusProfile, error) {
	if c.profile == nil || c.profile.ID != profileID {
		return nil, assert.AnError
	}
	return c.profile, nil
}
