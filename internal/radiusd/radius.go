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
	"strings"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
	"github.com/talkincode/toughradius/v9/config"
	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/registry"
	"github.com/talkincode/toughradius/v9/internal/radiusd/repository"
	repogorm "github.com/talkincode/toughradius/v9/internal/radiusd/repository/gorm"
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
	VendorMikrotik = "14988"
	VendorIkuai    = "10055"
	VendorHuawei   = "2011"
	VendorZte      = "3902"
	VendorH3c      = "25506"
	VendorRadback  = "2352"
	VendorCisco    = "9"

	RadiusRejectDelayTimes = 7
	RadiusAuthRateInterval = 1
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

type EapState struct {
	Username  string
	Challenge []byte
	StateID   string
	EapMethad string
	Success   bool
}

type RadiusService struct {
	appCtx        app.AppContext // Use interface instead of concrete type
	AuthRateCache map[string]AuthRateUser
	EapStateCache map[string]EapState
	TaskPool      *ants.Pool
	arclock       sync.Mutex
	eaplock       sync.Mutex

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
	pool, err := ants.NewPool(poolsize)
	common.Must(err)

	// Initialize all repositories using injected context
	db := appCtx.DB()
	s := &RadiusService{
		appCtx:        appCtx,
		AuthRateCache: make(map[string]AuthRateUser),
		EapStateCache: make(map[string]EapState),
		arclock:       sync.Mutex{},
		TaskPool:      pool,
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

func (s *RadiusService) RADIUSSecret(ctx context.Context, remoteAddr net.Addr) ([]byte, error) {
	return []byte("mysecret"), nil
}

// GetNas looks up a NAS device, preferring IP before ID
// Deprecated: Use NasRepo.GetByIPOrIdentifier instead
func (s *RadiusService) GetNas(ip, identifier string) (nas *domain.NetNas, err error) {
	// Adapter: delegate to repository layer
	nas, err = s.NasRepo.GetByIPOrIdentifier(context.Background(), ip, identifier)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, NewAuthError(app.MetricsRadiusRejectUnauthorized,
				fmt.Sprintf("unauthorized access to device, Ip=%s, Identifier=%s, %s",
					ip, identifier, err.Error()))
		}
		return nil, err
	}
	return nas, nil
}

// GetValidUser retrieves a valid user and performs initial checks
// Deprecated: Use UserRepo methods with plugin-based validation instead
func (s *RadiusService) GetValidUser(usernameOrMac string, macauth bool) (user *domain.RadiusUser, err error) {
	// Adapter: delegate to repository layer
	ctx := context.Background()
	if macauth {
		user, err = s.UserRepo.GetByMacAddr(ctx, usernameOrMac)
	} else {
		user, err = s.UserRepo.GetByUsername(ctx, usernameOrMac)
	}

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, NewAuthError(app.MetricsRadiusRejectNotExists, "user not exists")
		}
		return nil, err
	}

	// Keep original validation logic for backward compatibility
	if user.Status == common.DISABLED {
		return nil, NewAuthError(app.MetricsRadiusRejectDisable, "user status is disabled")
	}

	if user.ExpireTime.Before(time.Now()) {
		return nil, NewAuthError(app.MetricsRadiusRejectExpire, "user expire")
	}
	return user, nil
}

// GetUserForAcct fetches the user without validating expiration or status
// Deprecated: Use UserRepo.GetByUsername instead
func (s *RadiusService) GetUserForAcct(username string) (user *domain.RadiusUser, err error) {
	// Adapter: delegate to repository layer
	user, err = s.UserRepo.GetByUsername(context.Background(), username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, NewAuthError(app.MetricsRadiusRejectNotExists, "user not exists")
		}
		return nil, err
	}
	return user, nil
}

// Deprecated: Use UserRepo.UpdateField instead
func (s *RadiusService) UpdateUserField(username string, field string, value interface{}) {
	err := s.UserRepo.UpdateField(context.Background(), username, field, value)
	if err != nil {
		zap.L().Error(fmt.Sprintf("update user %s error", field), zap.Error(err), zap.String("namespace", "radius"))
	}
}

// Deprecated: Use UserRepo.UpdateMacAddr instead
func (s *RadiusService) UpdateUserMac(username string, macaddr string) {
	_ = s.UserRepo.UpdateMacAddr(context.Background(), username, macaddr)
}

// Deprecated: Use UserRepo.UpdateVlanId instead
func (s *RadiusService) UpdateUserVlanid1(username string, vlanid1 int) {
	_ = s.UserRepo.UpdateVlanId(context.Background(), username, vlanid1, 0)
}

// Deprecated: Use UserRepo.UpdateVlanId instead
func (s *RadiusService) UpdateUserVlanid2(username string, vlanid2 int) {
	_ = s.UserRepo.UpdateVlanId(context.Background(), username, 0, vlanid2)
}

// Deprecated: Use UserRepo.UpdateLastOnline instead
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
	case VendorHuawei:
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
			return NewAuthError(app.MetricsRadiusRejectLimit, "there is a authentication still in process")
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

// Deprecated: Use SessionRepo.Create instead
func (s *RadiusService) AddRadiusOnline(ol domain.RadiusOnline) error {
	ol.ID = common.UUIDint64()
	return s.SessionRepo.Create(context.Background(), &ol)
}

// Deprecated: Use AccountingRepo.Create instead
func (s *RadiusService) AddRadiusAccounting(ol domain.RadiusOnline, start bool) error {
	accounting := domain.RadiusAccounting{
		ID:                  common.UUIDint64(),
		Username:            ol.Username,
		AcctSessionId:       ol.AcctSessionId,
		NasId:               ol.NasId,
		NasAddr:             ol.NasAddr,
		NasPaddr:            ol.NasPaddr,
		SessionTimeout:      ol.SessionTimeout,
		FramedIpaddr:        ol.FramedIpaddr,
		FramedNetmask:       ol.FramedNetmask,
		FramedIpv6Prefix:    ol.FramedIpv6Prefix,
		FramedIpv6Address:   ol.FramedIpv6Address,
		DelegatedIpv6Prefix: ol.DelegatedIpv6Prefix,
		MacAddr:             ol.MacAddr,
		NasPort:             ol.NasPort,
		NasClass:            ol.NasClass,
		NasPortId:           ol.NasPortId,
		NasPortType:         ol.NasPortType,
		ServiceType:         ol.ServiceType,
		AcctSessionTime:     ol.AcctSessionTime,
		AcctInputTotal:      ol.AcctInputTotal,
		AcctOutputTotal:     ol.AcctOutputTotal,
		AcctInputPackets:    ol.AcctInputPackets,
		AcctOutputPackets:   ol.AcctOutputPackets,
		LastUpdate:          time.Now(),
		AcctStartTime:       ol.AcctStartTime,
		AcctStopTime:        time.Time{},
	}

	if !start {
		accounting.AcctStopTime = time.Now()
	}
	return s.AccountingRepo.Create(context.Background(), &accounting)
}

// Deprecated: Use SessionRepo.CountByUsername instead
func (s *RadiusService) GetRadiusOnlineCount(username string) int {
	count, _ := s.SessionRepo.CountByUsername(context.Background(), username)
	return count
}

// Deprecated: Use SessionRepo.Exists instead
func (s *RadiusService) ExistRadiusOnline(sessionId string) bool {
	exists, _ := s.SessionRepo.Exists(context.Background(), sessionId)
	return exists
}

// Deprecated: Use SessionRepo.Update instead
func (s *RadiusService) UpdateRadiusOnlineData(data domain.RadiusOnline) error {
	return s.SessionRepo.Update(context.Background(), &data)
}

// Deprecated: Use AccountingRepo.UpdateStop instead
func (s *RadiusService) EndRadiusAccounting(online domain.RadiusOnline) error {
	accounting := domain.RadiusAccounting{
		AcctSessionId:     online.AcctSessionId,
		AcctSessionTime:   online.AcctSessionTime,
		AcctInputTotal:    online.AcctInputTotal,
		AcctOutputTotal:   online.AcctOutputTotal,
		AcctInputPackets:  online.AcctInputPackets,
		AcctOutputPackets: online.AcctOutputPackets,
	}
	return s.AccountingRepo.UpdateStop(context.Background(), online.AcctSessionId, &accounting)
}

// Deprecated: Use SessionRepo.Delete instead
func (s *RadiusService) RemoveRadiusOnline(sessionId string) error {
	return s.SessionRepo.Delete(context.Background(), sessionId)
}

// Deprecated: Use SessionRepo.BatchDelete instead
func (s *RadiusService) BatchClearRadiusOnline(ids string) error {
	return s.SessionRepo.BatchDelete(context.Background(), strings.Split(ids, ","))
}

// Deprecated: Use SessionRepo.BatchDeleteByNas instead
func (s *RadiusService) BatchClearRadiusOnlineByNas(nasip, nasid string) {
	_ = s.SessionRepo.BatchDeleteByNas(context.Background(), nasip, nasid)
}

func (s *RadiusService) Release() {
	s.TaskPool.Running()
	_ = s.TaskPool.ReleaseTimeout(time.Second * 5)
}

var secretError = errors.New("secret error")

func (s *RadiusService) CheckRequestSecret(r *radius.Packet, secret []byte) {
	request, err := r.MarshalBinary()
	if err != nil {
		panic(err)
	}

	if len(secret) == 0 {
		panic(secretError)
	}

	hash := md5.New()
	hash.Write(request[:4])
	var nul [16]byte
	hash.Write(nul[:])
	hash.Write(request[20:])
	hash.Write(secret)
	var sum [md5.Size]byte
	if !bytes.Equal(hash.Sum(sum[:0]), request[4:20]) {
		panic(secretError)
	}
}

// State add
func (s *RadiusService) AddEapState(stateid, username string, challenge []byte, eapMethad string) {
	s.eaplock.Lock()
	defer s.eaplock.Unlock()
	s.EapStateCache[stateid] = EapState{
		Username:  username,
		StateID:   stateid,
		Challenge: challenge,
		EapMethad: eapMethad,
		Success:   false,
	}
}

// State get
func (s *RadiusService) GetEapState(stateid string) (state *EapState, err error) {
	s.eaplock.Lock()
	defer s.eaplock.Unlock()
	val, ok := s.EapStateCache[stateid]
	if ok {
		return &val, nil
	}
	return nil, errors.New("state not found")
}

// State delete
func (s *RadiusService) DeleteEapState(stateid string) {
	s.eaplock.Lock()
	defer s.eaplock.Unlock()
	delete(s.EapStateCache, stateid)
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
	if user.Vlanid1 != reqvid1 {
		s.UpdateUserVlanid2(user.Username, reqvid1)
	}
	if user.Vlanid2 != reqvid2 {
		s.UpdateUserVlanid2(user.Username, reqvid2)
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
