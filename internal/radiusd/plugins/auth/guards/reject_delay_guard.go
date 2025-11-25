package guards

import (
	"context"
	"sync"
	"time"

	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/radiusd/errors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	"layeh.com/radius/rfc2865"
)

const (
	defaultRejectLimit   = 7
	defaultResetWindow   = 10 * time.Second
	maxCachedRejectItems = 65535
)

// rejectItem stores rejection info for a single username
type rejectItem struct {
	mu         sync.Mutex
	rejects    int64
	lastReject time.Time
}

func (ri *rejectItem) exceeded(limit int64, window time.Duration) bool {
	ri.mu.Lock()
	defer ri.mu.Unlock()

	if time.Since(ri.lastReject) > window {
		ri.rejects = 0
	}

	if ri.rejects > limit {
		return true
	}

	ri.rejects++
	ri.lastReject = time.Now()
	return false
}

// RejectDelayGuard blocks requests when consecutive rejection counts exceed the threshold
type RejectDelayGuard struct {
	maxRejects int64
	resetAfter time.Duration

	configGetter configInt64Getter

	mu    sync.RWMutex
	items map[string]*rejectItem
}

type configInt64Getter interface {
	GetInt64(category, name string) int64
}

// NewRejectDelayGuard Create RejectDelayGuard
func NewRejectDelayGuard(getter configInt64Getter) *RejectDelayGuard {
	return &RejectDelayGuard{
		maxRejects:   defaultRejectLimit,
		resetAfter:   defaultResetWindow,
		configGetter: getter,
		items:        make(map[string]*rejectItem),
	}
}

func (g *RejectDelayGuard) Name() string {
	return "reject-delay"
}

// OnError tracks rejection counts and returns a rate-limit error when the threshold is exceeded
func (g *RejectDelayGuard) OnError(ctx context.Context, authCtx *auth.AuthContext, stage string, err error) error {
	if err == nil {
		return nil
	}

	username := g.resolveUsername(authCtx)
	if username == "" {
		username = "anonymous"
	}

	item := g.getItem(username)
	if item.exceeded(g.currentMaxRejects(), g.currentResetWindow()) {
		return errors.NewAuthError(app.MetricsRadiusRejectLimit, err.Error())
	}

	return nil
}

func (g *RejectDelayGuard) currentMaxRejects() int64 {
	if g.configGetter == nil {
		return g.maxRejects
	}
	if val := g.configGetter.GetInt64("radius", app.ConfigRadiusRejectDelayMaxRejects); val > 0 {
		return val
	}
	return g.maxRejects
}

func (g *RejectDelayGuard) currentResetWindow() time.Duration {
	if g.configGetter == nil {
		return g.resetAfter
	}
	if seconds := g.configGetter.GetInt64("radius", app.ConfigRadiusRejectDelayWindowSecond); seconds > 0 {
		return time.Duration(seconds) * time.Second
	}
	return g.resetAfter
}

func (g *RejectDelayGuard) resolveUsername(ctx *auth.AuthContext) string {
	if ctx == nil {
		return ""
	}
	if ctx.Metadata != nil {
		if v, ok := ctx.Metadata["username"].(string); ok && v != "" {
			return v
		}
	}
	if ctx.User != nil && ctx.User.Username != "" {
		return ctx.User.Username
	}
	if ctx.Request != nil {
		if v := rfc2865.UserName_GetString(ctx.Request.Packet); v != "" {
			return v
		}
	}
	return ""
}

func (g *RejectDelayGuard) getItem(username string) *rejectItem {
	g.mu.RLock()
	item, ok := g.items[username]
	g.mu.RUnlock()
	if ok {
		return item
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if item, ok = g.items[username]; ok {
		return item
	}

	if len(g.items) >= maxCachedRejectItems {
		g.items = make(map[string]*rejectItem)
	}

	item = &rejectItem{}
	g.items[username] = item
	return item
}
