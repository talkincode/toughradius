package checkers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/errors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	"github.com/talkincode/toughradius/v9/pkg/common"
)

func TestStatusChecker_Name(t *testing.T) {
	checker := &StatusChecker{}
	assert.Equal(t, "status", checker.Name())
}

func TestStatusChecker_Order(t *testing.T) {
	checker := &StatusChecker{}
	assert.Equal(t, 5, checker.Order())
}

func TestStatusChecker_Check(t *testing.T) {
	checker := &StatusChecker{}
	ctx := context.Background()

	tests := []struct {
		name        string
		userStatus  string
		expectError bool
	}{
		{
			name:        "enabled user",
			userStatus:  common.ENABLED,
			expectError: false,
		},
		{
			name:        "disabled user",
			userStatus:  common.DISABLED,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &domain.RadiusUser{
				Username: "testuser",
				Status:   tt.userStatus,
			}

			authCtx := &auth.AuthContext{
				User: user,
			}

			err := checker.Check(ctx, authCtx)

			if tt.expectError {
				require.Error(t, err)
				authErr, ok := errors.GetAuthError(err)
				assert.True(t, ok)
				assert.Contains(t, authErr.Message, "disabled")
			} else {
				require.NoError(t, err)
			}
		})
	}
}
