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
	// Note: 这个测试不能完全测试 Enhance 方法，因为它依赖全局 app 实例
	// 我们只测试基本的 nil 安全性
	enhancer := NewDefaultAcceptEnhancer()
	ctx := context.Background()

	t.Run("nil safety", func(t *testing.T) {
		// 测试 nil context
		err := enhancer.Enhance(ctx, nil)
		require.NoError(t, err)

		// 测试 nil response
		err = enhancer.Enhance(ctx, &auth.AuthContext{
			User: &domain.RadiusUser{},
		})
		require.NoError(t, err)

		// 测试 nil user
		err = enhancer.Enhance(ctx, &auth.AuthContext{
			Response: radius.New(radius.CodeAccessAccept, []byte("secret")),
		})
		require.NoError(t, err)
	})

	// TODO: 完整的功能测试需要初始化全局 app 实例
	// 或者重构代码以支持依赖注入
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
