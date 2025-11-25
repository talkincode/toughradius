package checkers

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/errors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
)

func TestExpireChecker_Name(t *testing.T) {
	checker := &ExpireChecker{}
	assert.Equal(t, "expire", checker.Name())
}

func TestExpireChecker_Order(t *testing.T) {
	checker := &ExpireChecker{}
	assert.Equal(t, 10, checker.Order())
}

func TestExpireChecker_Check(t *testing.T) {
	checker := &ExpireChecker{}
	ctx := context.Background()

	tests := []struct {
		name        string
		expireTime  time.Time
		expectError bool
	}{
		{
			name:        "user not expired",
			expireTime:  time.Now().Add(24 * time.Hour),
			expectError: false,
		},
		{
			name:        "user expired",
			expireTime:  time.Now().Add(-24 * time.Hour),
			expectError: true,
		},
		{
			name:        "user expires exactly now (edge case)",
			expireTime:  time.Now().Add(-1 * time.Second),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &domain.RadiusUser{
				Username:   "testuser",
				ExpireTime: tt.expireTime,
			}

			authCtx := &auth.AuthContext{
				User: user,
			}

			err := checker.Check(ctx, authCtx)

			if tt.expectError {
				require.Error(t, err)
				authErr, ok := errors.GetAuthError(err)
				assert.True(t, ok)
				assert.Contains(t, authErr.Message, "expired")
			} else {
				require.NoError(t, err)
			}
		})
	}
}
