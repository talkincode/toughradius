package radiusd

import (
	"bytes"
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
	"github.com/talkincode/toughradius/v9/config"
	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/domain"
	cachepkg "github.com/talkincode/toughradius/v9/internal/radiusd/cache"
	radiuserrors "github.com/talkincode/toughradius/v9/internal/radiusd/errors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/registry"
	"github.com/talkincode/toughradius/v9/internal/radiusd/repository"
	repogorm "github.com/talkincode/toughradius/v9/internal/radiusd/repository/gorm"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors/huawei"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2866"
	"layeh.com/radius/rfc2869"
	"layeh.com/radius/rfc3162"
	"layeh.com/radius/rfc4818"

	// Import vendor parsers for auto-registration
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	vendorparsers "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
	_ "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers/parsers"
)

const (
	RadiusRejectDelayTimes = 7
	RadiusAuthRateInterval = 1 // Original: 1 second rate limit

	// unknownNasSecret is a placeholder secret returned for packets from an
	// unrecognized NAS source IP. It is intentionally not a real credential; it
	// only lets the packet reach the request handler so the unauthorized NAS can
	// be logged and rejected with the correct metric.
	unknownNasSecret = "__unknown_nas__" //nolint:gosec // G101: placeholder, not a real secret
)

type VendorRequest struct {
	MacAddr string
	Vlanid1 int64
	Vlanid2 int64
}

type AuthRateUser struct {
	Username  string
	Starttime time.Time
}

type RadiusService struct {
	appCtx        app.AppContext // Use interface instead of concrete type
	AuthRateCache map[string]AuthRateUser
	TaskPool      *ants.Pool
	arclock       sync.Mutex
	nasCache      *cachepkg.TTLCache[*domain.NetNas]
	userCache     *cachepkg.TTLCache[*domain.RadiusUser]

	// New Repository Layer (v9 refactoring)
	UserRepo       repository.UserRepository
	SessionRepo    repository.SessionRepository
	AccountingRepo repository.AccountingRepository
	NasRepo        repository.NasRepository
}

func NewRadiusService(appCtx app.AppContext) *RadiusService {
	poolsize, err := strconv.Atoi(os.Getenv("TOUGHRADIUS_RADIUS_POOL"))
	if err != nil {
		poolsize = 1024
	}
	// Nonblocking: when all workers are busy, Submit returns ErrPoolOverload
	// immediately instead of blocking the caller. Accounting overflow is then
	// dropped and metered (see AcctService.submitAcctTask), which bounds the
	// number of goroutines under load rather than letting blocked submitters
	// accumulate without limit.
	pool, err := ants.NewPool(poolsize, ants.WithNonblocking(true))
	common.Must(err)

	// Initialize all repositories using injected context
	db := appCtx.DB()
	s := &RadiusService{
		appCtx:        appCtx,
		AuthRateCache: make(map[string]AuthRateUser),
		arclock:       sync.Mutex{},
		TaskPool:      pool,
		nasCache:      cachepkg.NewTTLCache[*domain.NetNas](time.Minute, 512),
		userCache:     cachepkg.NewTTLCache[*domain.RadiusUser](10*time.Second, 2048),
		// Initialize repository layer
		UserRepo:       repogorm.NewGormUserRepository(db),
		SessionRepo:    repogorm.NewGormSessionRepository(db),
		AccountingRepo: repogorm.NewGormAccountingRepository(db),
		NasRepo:        repogorm.NewGormNasRepository(db),
	}

	// Note: Plugin initialization is done externally after service creation
	// to avoid circular dependency. Call plugins.InitPlugins() from main.go.

	return s
}

// RADIUSSecret resolves the shared secret for an incoming UDP packet by looking
// up the originating NAS by its source IP address. When the NAS is unknown a
// non-empty placeholder is returned so the packet still reaches the request
// handler, where the unauthorized NAS is logged and rejected with the proper
// metric instead of being silently dropped by the server library.
func (s *RadiusService) RADIUSSecret(ctx context.Context, remoteAddr net.Addr) ([]byte, error) {
	ip := remoteAddr.String()
	if host, _, err := net.SplitHostPort(ip); err == nil {
		ip = host
	}
	nas, err := s.GetNas(ip, "")
	if err != nil {
		return []byte(unknownNasSecret), nil
	}
	return []byte(nas.Secret), nil
}

// GetNas looks up a NAS device by source IP (preferred) or identifier. Results
// are cached, and a missing record is mapped to an unauthorized-NAS error.
func (s *RadiusService) GetNas(ip, identifier string) (nas *domain.NetNas, err error) {
	cacheKey := fmt.Sprintf("%s|%s", ip, identifier)
	if cached, ok := s.nasCache.Get(cacheKey); ok {
		return cached, nil
	}
	// Adapter: delegate to repository layer
	nas, err = s.NasRepo.GetByIPOrIdentifier(context.Background(), ip, identifier)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, radiuserrors.NewUnauthorizedNasError(ip, identifier, err)
		}
		return nil, err
	}
	s.nasCache.Set(cacheKey, nas)
	return nas, nil
}

// GetValidUser retrieves a user by username (or MAC address for MAC auth),
// caches the result, and rejects disabled or expired accounts.
func (s *RadiusService) GetValidUser(usernameOrMac string, macauth bool) (user *domain.RadiusUser, err error) {
	cacheKey := fmt.Sprintf("%t|%s", macauth, usernameOrMac)
	if cached, ok := s.userCache.Get(cacheKey); ok {
		return cached, nil
	}
	// Adapter: delegate to repository layer
	ctx := context.Background()
	if macauth {
		user, err = s.UserRepo.GetByMacAddr(ctx, usernameOrMac)
	} else {
		user, err = s.UserRepo.GetByUsername(ctx, usernameOrMac)
	}

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, radiuserrors.NewUserNotExistsError()
		}
		return nil, err
	}

	// Keep original validation logic for backward compatibility
	if user.Status == common.DISABLED {
		return nil, radiuserrors.NewUserDisabledError()
	}

	if user.ExpireTime.Before(time.Now()) {
		return nil, radiuserrors.NewUserExpiredError()
	}
	s.userCache.Set(cacheKey, user)
	return user, nil
}

// UpdateUserMac persists the most recently seen MAC address for a user.
func (s *RadiusService) UpdateUserMac(username string, macaddr string) {
	_ = s.UserRepo.UpdateMacAddr(context.Background(), username, macaddr)
}

// UpdateUserLastOnline records the user's last-online timestamp.
func (s *RadiusService) UpdateUserLastOnline(username string) {
	_ = s.UserRepo.UpdateLastOnline(context.Background(), username)
}

func (s *RadiusService) GetEapMethod() string {
	// Read directly from the ConfigManager (already in memory)
	return s.appCtx.ConfigMgr().GetString("radius", "EapMethod")
}

// Config returns the application configuration
func (s *RadiusService) Config() *config.AppConfig {
	return s.appCtx.Config()
}

// AppContext returns the application context
func (s *RadiusService) AppContext() app.AppContext {
	return s.appCtx
}

func GetFramedIpv6Address(r *radius.Request, nas *domain.NetNas) string {
	switch nas.VendorCode {
	case vendors.CodeHuawei:
		return common.IfEmptyStr(huawei.HuaweiFramedIPv6Address_Get(r.Packet).String(), common.NA)
	default:
		return ""
	}
}

func GetNetRadiusOnlineFromRequest(r *radius.Request, vr *VendorRequest, nas *domain.NetNas, nasrip string) domain.RadiusOnline {
	acctInputOctets := int(rfc2866.AcctInputOctets_Get(r.Packet))
	acctInputGigawords := int(rfc2869.AcctInputGigawords_Get(r.Packet))
	acctOutputOctets := int(rfc2866.AcctOutputOctets_Get(r.Packet))
	acctOutputGigawords := int(rfc2869.AcctOutputGigawords_Get(r.Packet))

	getAcctStartTime := func(sessionTime int) time.Time {
		m, _ := time.ParseDuration(fmt.Sprintf("-%ds", sessionTime))
		return time.Now().Add(m)
	}
	return domain.RadiusOnline{
		ID:                  0,
		Username:            rfc2865.UserName_GetString(r.Packet),
		NasId:               common.IfEmptyStr(rfc2865.NASIdentifier_GetString(r.Packet), common.NA),
		NasAddr:             nas.Ipaddr,
		NasPaddr:            nasrip,
		SessionTimeout:      int(rfc2865.SessionTimeout_Get(r.Packet)),
		FramedIpaddr:        common.IfEmptyStr(rfc2865.FramedIPAddress_Get(r.Packet).String(), common.NA),
		FramedNetmask:       common.IfEmptyStr(rfc2865.FramedIPNetmask_Get(r.Packet).String(), common.NA),
		FramedIpv6Address:   GetFramedIpv6Address(r, nas),
		FramedIpv6Prefix:    common.IfEmptyStr(rfc3162.FramedIPv6Prefix_Get(r.Packet).String(), common.NA),
		DelegatedIpv6Prefix: common.IfEmptyStr(rfc4818.DelegatedIPv6Prefix_Get(r.Packet).String(), common.NA),
		MacAddr:             common.IfEmptyStr(vr.MacAddr, common.NA),
		NasPort:             0,
		NasClass:            common.NA,
		NasPortId:           common.IfEmptyStr(rfc2869.NASPortID_GetString(r.Packet), common.NA),
		NasPortType:         0,
		ServiceType:         0,
		AcctSessionId:       rfc2866.AcctSessionID_GetString(r.Packet),
		AcctSessionTime:     int(rfc2866.AcctSessionTime_Get(r.Packet)),
		AcctInputTotal:      int64(acctInputOctets) + int64(acctInputGigawords)*4*1024*1024*1024,
		AcctOutputTotal:     int64(acctOutputOctets) + int64(acctOutputGigawords)*4*1024*1024*1024,
		AcctInputPackets:    int(rfc2866.AcctInputPackets_Get(r.Packet)),
		AcctOutputPackets:   int(rfc2866.AcctOutputPackets_Get(r.Packet)),
		AcctStartTime:       getAcctStartTime(int(rfc2866.AcctSessionTime_Get(r.Packet))),
		LastUpdate:          time.Now(),
	}

}

// CheckAuthRateLimit
// Authentication frequency detection, each user can only authenticate once every few seconds
func (s *RadiusService) CheckAuthRateLimit(username string) error {
	s.arclock.Lock()
	defer s.arclock.Unlock()
	val, ok := s.AuthRateCache[username]
	if ok {
		if time.Now().Before(val.Starttime.Add(time.Duration(RadiusAuthRateInterval) * time.Second)) {
			return radiuserrors.NewOnlineLimitError("there is a authentication still in process")
		}
		delete(s.AuthRateCache, username)
	}
	s.AuthRateCache[username] = AuthRateUser{
		Username:  username,
		Starttime: time.Now(),
	}
	return nil
}

func (s *RadiusService) ReleaseAuthRateLimit(username string) {
	s.arclock.Lock()
	defer s.arclock.Unlock()
	delete(s.AuthRateCache, username)
}

func (s *RadiusService) Release() {
	s.TaskPool.Running()
	_ = s.TaskPool.ReleaseTimeout(time.Second * 5)
}

// ErrSecretEmpty indicates an empty RADIUS secret
var ErrSecretEmpty = errors.New("secret is empty")

// ErrSecretMismatch indicates a RADIUS secret mismatch
var ErrSecretMismatch = errors.New("secret mismatch")

// CheckRequestSecret validates the RADIUS packet authenticator against the shared secret.
// Returns an error if validation fails, nil on success.
func (s *RadiusService) CheckRequestSecret(r *radius.Packet, secret []byte) error {
	request, err := r.MarshalBinary()
	if err != nil {
		return fmt.Errorf("failed to marshal packet: %w", err)
	}

	if len(secret) == 0 {
		return ErrSecretEmpty
	}

	hash := md5.New()
	hash.Write(request[:4])
	var nul [16]byte
	hash.Write(nul[:])
	hash.Write(request[20:])
	hash.Write(secret)
	var sum [md5.Size]byte
	if !bytes.Equal(hash.Sum(sum[:0]), request[4:20]) {
		return ErrSecretMismatch
	}
	return nil
}

func (s *AuthService) GetLocalPassword(user *domain.RadiusUser, isMacAuth bool) (string, error) {
	if isMacAuth {
		return user.MacAddr, nil
	}
	return user.Password, nil
}

func (s *AuthService) UpdateBind(user *domain.RadiusUser, vendorReq *VendorRequest) {
	if user.MacAddr != vendorReq.MacAddr {
		s.UpdateUserMac(user.Username, vendorReq.MacAddr)
	}
	reqvid1 := int(vendorReq.Vlanid1)
	reqvid2 := int(vendorReq.Vlanid2)
	// UpdateVlanId writes both vlanid columns at once, so persist them together
	// when either differs. Updating them via the single-field helpers would zero
	// out the other column (and the old code also wrote vlanid1 into vlanid2).
	if user.Vlanid1 != reqvid1 || user.Vlanid2 != reqvid2 {
		_ = s.UserRepo.UpdateVlanId(context.Background(), user.Username, reqvid1, reqvid2)
	}
}

// ApplyAcceptEnhancers delivers user profile configuration via plugins
func (s *AuthService) ApplyAcceptEnhancers(
	user *domain.RadiusUser,
	nas *domain.NetNas,
	vendorReq *vendorparsers.VendorRequest,
	radAccept *radius.Packet,
) {
	authCtx := &auth.AuthContext{
		User:          user,
		Nas:           nas,
		VendorRequest: vendorReq,
		Response:      radAccept,
	}

	ctx := context.Background()
	for _, enhancer := range registry.GetResponseEnhancers() {
		if err := enhancer.Enhance(ctx, authCtx); err != nil {
			zap.L().Warn("response enhancer failed",
				zap.String("enhancer", enhancer.Name()),
				zap.Error(err))
		}
	}
}

func (s *RadiusService) DoAcctDisconnect(r *radius.Request, nas *domain.NetNas, username, nasrip string) {
	packet := radius.New(radius.CodeDisconnectRequest, []byte(nas.Secret))
	sessionid := rfc2866.AcctSessionID_GetString(r.Packet)
	if sessionid == "" {
		return
	}
	_ = rfc2865.UserName_SetString(packet, username)
	_ = rfc2866.AcctSessionID_Set(packet, []byte(sessionid))
	response, err := radius.Exchange(context.Background(), packet, fmt.Sprintf("%s:%d", nasrip, nas.CoaPort))
	if err != nil {
		zap.L().Error("radius disconnect error",
			zap.String("namespace", "radius"),
			zap.String("username", username),
			zap.Error(err),
		)
		return
	}
	zap.L().Info("radius disconnect done",
		zap.String("namespace", "radius"),
		zap.String("nasip", nasrip),
		zap.Int("coaport", nas.CoaPort),
		zap.String("request", FmtPacket(packet)),
		zap.String("response", FmtPacket(response)),
	)
}
