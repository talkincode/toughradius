package guards

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
	radiusErrors "github.com/talkincode/toughradius/v9/internal/radiusd/errors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

func TestRejectDelayGuard_Name(t *testing.T) {
	guard := NewRejectDelayGuard()
	assert.Equal(t, "reject-delay", guard.Name())
}

func TestRejectDelayGuard_OnError_NoError(t *testing.T) {
	guard := NewRejectDelayGuard()
	ctx := context.Background()

	err := guard.OnError(ctx, &auth.AuthContext{}, "test", nil)
	require.NoError(t, err)
}

func TestRejectDelayGuard_OnError_RejectLimit(t *testing.T) {
	guard := NewRejectDelayGuard()
	ctx := context.Background()

	packet := radius.New(radius.CodeAccessRequest, []byte("secret"))
	rfc2865.UserName_SetString(packet, "testuser")

	authCtx := &auth.AuthContext{
		Request: &radius.Request{Packet: packet},
	}

	testErr := errors.New("auth failed")

	// 测试拒绝次数累计
	for i := int64(0); i <= guard.maxRejects; i++ {
		err := guard.OnError(ctx, authCtx, "test", testErr)
		require.NoError(t, err)
	}

	// 超过限制后应该返回限流错误
	err := guard.OnError(ctx, authCtx, "test", testErr)
	require.Error(t, err)
	authErr, ok := radiusErrors.GetAuthError(err)
	require.True(t, ok)
	assert.Contains(t, authErr.MetricsType, "limit")
}

func TestRejectDelayGuard_OnError_ResetWindow(t *testing.T) {
	guard := NewRejectDelayGuard()
	guard.resetAfter = 100 * time.Millisecond // 缩短重置时间以加快测试
	ctx := context.Background()

	packet := radius.New(radius.CodeAccessRequest, []byte("secret"))
	rfc2865.UserName_SetString(packet, "testuser")

	authCtx := &auth.AuthContext{
		Request: &radius.Request{Packet: packet},
	}

	testErr := errors.New("auth failed")

	// 累计拒绝次数到限制
	for i := int64(0); i <= guard.maxRejects; i++ {
		_ = guard.OnError(ctx, authCtx, "test", testErr)
	}

	// 超过限制
	err := guard.OnError(ctx, authCtx, "test", testErr)
	require.Error(t, err)

	// 等待重置窗口过期
	time.Sleep(guard.resetAfter + 50*time.Millisecond)

	// 应该可以再次认证（计数器被重置）
	err = guard.OnError(ctx, authCtx, "test", testErr)
	require.NoError(t, err)
}

func TestRejectDelayGuard_OnError_DifferentUsers(t *testing.T) {
	guard := NewRejectDelayGuard()
	ctx := context.Background()

	testErr := errors.New("auth failed")

	// 用户1累计到限制
	packet1 := radius.New(radius.CodeAccessRequest, []byte("secret"))
	rfc2865.UserName_SetString(packet1, "user1")
	authCtx1 := &auth.AuthContext{
		Request: &radius.Request{Packet: packet1},
	}

	for i := int64(0); i <= guard.maxRejects; i++ {
		_ = guard.OnError(ctx, authCtx1, "test", testErr)
	}

	// 用户1超过限制
	err := guard.OnError(ctx, authCtx1, "test", testErr)
	require.Error(t, err)

	// 用户2应该不受影响
	packet2 := radius.New(radius.CodeAccessRequest, []byte("secret"))
	rfc2865.UserName_SetString(packet2, "user2")
	authCtx2 := &auth.AuthContext{
		Request: &radius.Request{Packet: packet2},
	}

	err = guard.OnError(ctx, authCtx2, "test", testErr)
	require.NoError(t, err)
}

func TestRejectDelayGuard_ResolveUsername(t *testing.T) {
	guard := NewRejectDelayGuard()

	tests := []struct {
		name     string
		authCtx  *auth.AuthContext
		expected string
	}{
		{
			name:     "nil context",
			authCtx:  nil,
			expected: "",
		},
		{
			name: "from metadata",
			authCtx: &auth.AuthContext{
				Metadata: map[string]interface{}{
					"username": "meta_user",
				},
			},
			expected: "meta_user",
		},
		{
			name: "from user object",
			authCtx: &auth.AuthContext{
				User: &domain.RadiusUser{
					Username: "user_obj",
				},
			},
			expected: "user_obj",
		},
		{
			name: "from request packet",
			authCtx: &auth.AuthContext{
				Request: &radius.Request{
					Packet: func() *radius.Packet {
						p := radius.New(radius.CodeAccessRequest, []byte("secret"))
						rfc2865.UserName_SetString(p, "packet_user")
						return p
					}(),
				},
			},
			expected: "packet_user",
		},
		{
			name: "priority: metadata > user > packet",
			authCtx: &auth.AuthContext{
				Metadata: map[string]interface{}{
					"username": "meta_user",
				},
				User: &domain.RadiusUser{
					Username: "user_obj",
				},
				Request: &radius.Request{
					Packet: func() *radius.Packet {
						p := radius.New(radius.CodeAccessRequest, []byte("secret"))
						rfc2865.UserName_SetString(p, "packet_user")
						return p
					}(),
				},
			},
			expected: "meta_user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			username := guard.resolveUsername(tt.authCtx)
			assert.Equal(t, tt.expected, username)
		})
	}
}

func TestRejectDelayGuard_OnError_AnonymousUser(t *testing.T) {
	guard := NewRejectDelayGuard()
	ctx := context.Background()

	// 无用户名的请求
	authCtx := &auth.AuthContext{}
	testErr := errors.New("auth failed")

	// 应该使用 "anonymous" 作为用户名
	for i := int64(0); i <= guard.maxRejects; i++ {
		_ = guard.OnError(ctx, authCtx, "test", testErr)
	}

	err := guard.OnError(ctx, authCtx, "test", testErr)
	require.Error(t, err)
}

func TestRejectDelayGuard_CacheLimit(t *testing.T) {
	guard := NewRejectDelayGuard()
	ctx := context.Background()

	testErr := errors.New("auth failed")

	// 添加大量不同的用户，测试缓存清理
	for i := 0; i < maxCachedRejectItems+100; i++ {
		packet := radius.New(radius.CodeAccessRequest, []byte("secret"))
		username := "user_" + string(rune(i))
		rfc2865.UserName_SetString(packet, username)

		authCtx := &auth.AuthContext{
			Request: &radius.Request{Packet: packet},
		}

		guard.OnError(ctx, authCtx, "test", testErr)
	}

	// 验证缓存大小不超过限制
	guard.mu.RLock()
	cacheSize := len(guard.items)
	guard.mu.RUnlock()

	assert.LessOrEqual(t, cacheSize, maxCachedRejectItems)
}
