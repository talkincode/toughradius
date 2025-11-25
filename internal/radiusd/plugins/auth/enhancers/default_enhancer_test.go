package enhancers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	"layeh.com/radius"
)

func TestDefaultAcceptEnhancer_Name(t *testing.T) {
	enhancer := NewDefaultAcceptEnhancer()
	assert.Equal(t, "default-accept", enhancer.Name())
}

func TestDefaultAcceptEnhancer_Enhance(t *testing.T) {
	// Note: this test cannot fully exercise Enhance because it depends on the global app instance
	// We only verify basic nil-safety
	enhancer := NewDefaultAcceptEnhancer()
	ctx := context.Background()

	t.Run("nil safety", func(t *testing.T) {
		// Test nil context
		err := enhancer.Enhance(ctx, nil)
		require.NoError(t, err)

		// Test nil response
		err = enhancer.Enhance(ctx, &auth.AuthContext{
			User: &domain.RadiusUser{},
		})
		require.NoError(t, err)

		// Test nil user
		err = enhancer.Enhance(ctx, &auth.AuthContext{
			Response: radius.New(radius.CodeAccessAccept, []byte("secret")),
		})
		require.NoError(t, err)
	})

	// TODO: Full functionality tests need the global app instance initialized
	// or refactor the code to support dependency injection
}

func TestDefaultAcceptEnhancer_Enhance_NilSafety(t *testing.T) {
	enhancer := NewDefaultAcceptEnhancer()
	ctx := context.Background()

	tests := []struct {
		name    string
		authCtx *auth.AuthContext
	}{
		{
			name:    "nil context",
			authCtx: nil,
		},
		{
			name: "nil response",
			authCtx: &auth.AuthContext{
				User: &domain.RadiusUser{},
			},
		},
		{
			name: "nil user",
			authCtx: &auth.AuthContext{
				Response: radius.New(radius.CodeAccessAccept, []byte("secret")),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := enhancer.Enhance(ctx, tt.authCtx)
			require.NoError(t, err)
		})
	}
}
