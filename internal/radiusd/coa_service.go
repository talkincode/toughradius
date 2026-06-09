package radiusd

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net"
	"strconv"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2866"
	"layeh.com/radius/rfc2869"
	"layeh.com/radius/rfc3576"
)

// DefaultCoAPort is the IANA-assigned UDP port for RADIUS Dynamic Authorization
// (CoA-Request / Disconnect-Request) defined in RFC 5176 §2.1. It is used as a
// fallback when a NAS record does not configure an explicit CoA port.
const DefaultCoAPort = 3799

const (
	// defaultCoATimeout bounds how long a single CoA/Disconnect exchange waits
	// for an ACK/NAK before the attempt is considered timed out.
	defaultCoATimeout = 5 * time.Second
	// defaultCoARetries is the number of additional retransmissions performed
	// after the first attempt times out. RFC 5176 §2.3 recommends retransmitting
	// the same packet (same Identifier) so the NAS can detect duplicates.
	defaultCoARetries = 2
)

// ErrSessionNotFound is returned by the *Session helpers when no online session
// matches the supplied Acct-Session-Id.
var ErrSessionNotFound = errors.New("coa: online session not found")

// ErrNoTarget is returned when a CoA/Disconnect request is built without a
// destination NAS address.
var ErrNoTarget = errors.New("coa: target NAS address is required")

// CoAAction identifies the RADIUS Dynamic Authorization operation being sent.
type CoAAction string

const (
	// CoAActionDisconnect corresponds to a Disconnect-Request (RFC 5176 §2.1).
	CoAActionDisconnect CoAAction = "disconnect"
	// CoAActionCoA corresponds to a CoA-Request (RFC 5176 §2.2).
	CoAActionCoA CoAAction = "coa"
)

// CoATarget identifies the NAS (acting as the Dynamic Authorization Server) that
// a CoA/Disconnect request is delivered to.
type CoATarget struct {
	// Addr is the destination IP address or hostname of the NAS.
	Addr string
	// Secret is the RADIUS shared secret configured for the NAS.
	Secret string
	// Port is the NAS CoA/Disconnect UDP port. When zero or negative,
	// DefaultCoAPort (3799) is used.
	Port int
}

// endpoint renders the "host:port" destination, applying the default CoA port
// when none is configured.
func (t CoATarget) endpoint() string {
	port := t.Port
	if port <= 0 {
		port = DefaultCoAPort
	}
	return net.JoinHostPort(t.Addr, strconv.Itoa(port))
}

// CoATargetFromNas builds a CoATarget from a NAS record. The CoA request is sent
// to the NAS configured IP address; the CoA port falls back to DefaultCoAPort
// when the record leaves it unset.
func CoATargetFromNas(nas *domain.NetNas) CoATarget {
	return CoATarget{Addr: nas.Ipaddr, Secret: nas.Secret, Port: nas.CoaPort}
}

// SessionIdentity carries the RFC 5176 §3 NAS and session identification
// attributes the NAS uses to match the session(s) targeted by a CoA/Disconnect
// request. Empty fields are omitted from the packet.
type SessionIdentity struct {
	Username       string  // User-Name (#1)
	NasIP          string  // NAS-IP-Address (#4)
	NasIdentifier  string  // NAS-Identifier (#32)
	AcctSessionID  string  // Acct-Session-Id (#44)
	FramedIP       string  // Framed-IP-Address (#8)
	CallingStation string  // Calling-Station-Id (#31)
	NasPort        *uint32 // NAS-Port (#5); nil omits the attribute (0 is a valid port)
	NasPortID      string  // NAS-Port-Id (#87)
}

// SessionIdentityFromOnline derives the identification attributes from a stored
// online session. NAS-Port is only populated when the stored value is positive,
// since the schema uses 0 to mean "unset".
func SessionIdentityFromOnline(o *domain.RadiusOnline) SessionIdentity {
	id := SessionIdentity{
		Username:       o.Username,
		NasIP:          o.NasAddr,
		NasIdentifier:  o.NasId,
		AcctSessionID:  o.AcctSessionId,
		FramedIP:       o.FramedIpaddr,
		CallingStation: o.MacAddr,
		NasPortID:      o.NasPortId,
	}
	if o.NasPort > 0 && o.NasPort <= math.MaxUint32 {
		p := uint32(o.NasPort)
		id.NasPort = &p
	}
	return id
}

// CoAResult is the structured, audit-ready outcome of a single CoA/Disconnect
// exchange. M2.1 returns it to the caller; durable persistence is handled by a
// later milestone (M2.3).
type CoAResult struct {
	Action         CoAAction     // disconnect or coa
	Target         string        // "host:port" the request was sent to
	Username       string        // User-Name targeted (for audit correlation)
	AcctSessionID  string        // Acct-Session-Id targeted
	Identifier     int           // RADIUS packet Identifier (stable across retries)
	Success        bool          // true when the NAS replied with an ACK
	ResponseCode   string        // e.g. "Disconnect-ACK"/"CoA-NAK"; empty on timeout
	ErrorCause     int           // RFC 3576/5176 Error-Cause value (0 = none)
	ErrorCauseText string        // human-readable Error-Cause
	Attempts       int           // number of transmissions performed
	RTT            time.Duration // round-trip time of the answered attempt
	TimedOut       bool          // true when every attempt timed out
	Err            string        // transport/build error, if any
	SentAt         time.Time     // wall-clock time of the first transmission
}

// AttributeSetter mutates a CoA-Request to add an authorization-change attribute
// (for example a new Session-Timeout or Filter-Id). It is only used with CoA
// requests; a Disconnect-Request must carry identification attributes only
// (RFC 5176 §3).
type AttributeSetter func(*radius.Packet) error

// WithSessionTimeout returns an AttributeSetter that sets Session-Timeout (#27),
// the most common CoA authorization change for re-limiting a live session.
func WithSessionTimeout(seconds uint32) AttributeSetter {
	return func(p *radius.Packet) error {
		return rfc2865.SessionTimeout_Set(p, rfc2865.SessionTimeout(seconds))
	}
}

// WithFilterID returns an AttributeSetter that sets Filter-Id (#11), used to
// apply a named access-control filter to a live session via CoA.
func WithFilterID(name string) AttributeSetter {
	return func(p *radius.Packet) error {
		return rfc2865.FilterID_SetString(p, name)
	}
}

// CoAService implements the RFC 5176 Dynamic Authorization Client (DAC) role: it
// sends CoA-Request and Disconnect-Request packets to a NAS, retransmits on
// timeout up to a bounded budget, and reports a structured CoAResult.
//
// The embedded *RadiusService is optional. The explicit Disconnect/CoA methods
// only require a target and identity and can be used standalone; the
// DisconnectSession/CoASession helpers additionally use the session and NAS
// repositories to resolve a live session, and therefore require a non-nil
// RadiusService.
type CoAService struct {
	*RadiusService
	timeout time.Duration
	retries int
	// exchange performs the actual UDP round-trip. It is overridable in tests;
	// by default it is the Exchange method of a dedicated radius.Client whose
	// internal Retry is disabled (CoAService manages retransmission itself).
	exchange func(ctx context.Context, packet *radius.Packet, addr string) (*radius.Packet, error)
}

// CoAOption customizes a CoAService.
type CoAOption func(*CoAService)

// WithCoATimeout overrides the per-attempt exchange timeout.
func WithCoATimeout(d time.Duration) CoAOption {
	return func(s *CoAService) {
		if d > 0 {
			s.timeout = d
		}
	}
}

// WithCoARetries overrides the number of retransmissions performed after the
// first attempt times out. Negative values are clamped to zero.
func WithCoARetries(n int) CoAOption {
	return func(s *CoAService) {
		if n < 0 {
			n = 0
		}
		s.retries = n
	}
}

// NewCoAService constructs a CoAService. radiusService may be nil when only the
// explicit Disconnect/CoA methods are used (for example in unit tests or callers
// that resolve the NAS themselves).
func NewCoAService(radiusService *RadiusService, opts ...CoAOption) *CoAService {
	// A dedicated client with Retry disabled: CoAService performs its own
	// bounded retransmission loop, so the library must not also retransmit on
	// its 1-second default ticker. This keeps result.Attempts equal to the
	// number of packets actually put on the wire. InsecureSkipVerify stays
	// false so the NAS reply's Response Authenticator is verified.
	client := &radius.Client{
		Retry:           0,
		MaxPacketErrors: 10,
	}
	s := &CoAService{
		RadiusService: radiusService,
		timeout:       defaultCoATimeout,
		retries:       defaultCoARetries,
		exchange:      client.Exchange,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Disconnect sends a Disconnect-Request (RFC 5176 §2.1) targeting the session
// described by id. Per RFC 5176 §3 the packet carries only NAS and session
// identification attributes. It returns a structured CoAResult; a NAK or timeout
// is reported in the result (Success=false) rather than as an error. A non-nil
// error indicates the request could not be built or sent at all.
func (s *CoAService) Disconnect(ctx context.Context, target CoATarget, id SessionIdentity) (*CoAResult, error) {
	if target.Addr == "" {
		return nil, ErrNoTarget
	}
	packet := radius.New(radius.CodeDisconnectRequest, []byte(target.Secret))
	if err := applyIdentity(packet, id); err != nil {
		return nil, fmt.Errorf("coa: build disconnect request: %w", err)
	}
	return s.send(ctx, CoAActionDisconnect, target, id, packet), nil
}

// CoA sends a CoA-Request (RFC 5176 §2.2) targeting the session described by id
// and applying the supplied authorization-change attributes. It returns a
// structured CoAResult; a NAK or timeout is reported in the result rather than as
// an error.
func (s *CoAService) CoA(ctx context.Context, target CoATarget, id SessionIdentity, changes ...AttributeSetter) (*CoAResult, error) {
	if target.Addr == "" {
		return nil, ErrNoTarget
	}
	packet := radius.New(radius.CodeCoARequest, []byte(target.Secret))
	if err := applyIdentity(packet, id); err != nil {
		return nil, fmt.Errorf("coa: build coa request: %w", err)
	}
	for _, change := range changes {
		if change == nil {
			continue
		}
		if err := change(packet); err != nil {
			return nil, fmt.Errorf("coa: apply authorization change: %w", err)
		}
	}
	return s.send(ctx, CoAActionCoA, target, id, packet), nil
}

// DisconnectSession resolves the online session identified by acctSessionID,
// looks up its NAS, and sends a Disconnect-Request. It returns ErrSessionNotFound
// when no online session matches. Requires a non-nil embedded RadiusService.
func (s *CoAService) DisconnectSession(ctx context.Context, acctSessionID string) (*CoAResult, error) {
	target, id, err := s.resolveSession(ctx, acctSessionID)
	if err != nil {
		return nil, err
	}
	return s.Disconnect(ctx, target, id)
}

// CoASession resolves the online session identified by acctSessionID, looks up
// its NAS, and sends a CoA-Request applying the supplied changes. It returns
// ErrSessionNotFound when no online session matches. Requires a non-nil embedded
// RadiusService.
func (s *CoAService) CoASession(ctx context.Context, acctSessionID string, changes ...AttributeSetter) (*CoAResult, error) {
	target, id, err := s.resolveSession(ctx, acctSessionID)
	if err != nil {
		return nil, err
	}
	return s.CoA(ctx, target, id, changes...)
}

// resolveSession loads the online session and its NAS, mapping them to a target
// and identity for a CoA/Disconnect request.
func (s *CoAService) resolveSession(ctx context.Context, acctSessionID string) (CoATarget, SessionIdentity, error) {
	if s.RadiusService == nil {
		return CoATarget{}, SessionIdentity{}, errors.New("coa: RadiusService is required for session lookups")
	}
	online, err := s.SessionRepo.GetBySessionId(ctx, acctSessionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return CoATarget{}, SessionIdentity{}, fmt.Errorf("%w: %s", ErrSessionNotFound, acctSessionID)
		}
		return CoATarget{}, SessionIdentity{}, fmt.Errorf("coa: load session %s: %w", acctSessionID, err)
	}
	nas, err := s.GetNas(online.NasAddr, online.NasId)
	if err != nil {
		return CoATarget{}, SessionIdentity{}, fmt.Errorf("coa: resolve nas for session %s: %w", acctSessionID, err)
	}
	return CoATargetFromNas(nas), SessionIdentityFromOnline(online), nil
}

// send transmits packet to the target with a bounded timeout and retry budget,
// then classifies the reply into a CoAResult. The same packet (and therefore the
// same Identifier) is reused across retransmissions per RFC 5176 §2.3 so the NAS
// can deduplicate.
func (s *CoAService) send(ctx context.Context, action CoAAction, target CoATarget, id SessionIdentity, packet *radius.Packet) *CoAResult {
	endpoint := target.endpoint()
	result := &CoAResult{
		Action:        action,
		Target:        endpoint,
		Username:      id.Username,
		AcctSessionID: id.AcctSessionID,
		Identifier:    int(packet.Identifier),
		SentAt:        time.Now(),
	}

	var lastErr error
	for attempt := 0; attempt <= s.retries; attempt++ {
		if err := ctx.Err(); err != nil {
			lastErr = err
			break
		}
		result.Attempts++

		attemptCtx, cancel := context.WithTimeout(ctx, s.timeout)
		start := time.Now()
		resp, err := s.exchange(attemptCtx, packet, endpoint)
		rtt := time.Since(start)
		cancel()

		if err == nil {
			result.RTT = rtt
			classifyResponse(result, resp)
			s.logResult(result, packet, resp)
			return result
		}

		lastErr = err
		// Only timeouts are retransmitted; other transport errors (dial/parse)
		// will not improve on retry and are surfaced immediately.
		if !isTimeoutErr(err) {
			break
		}
	}

	if lastErr != nil {
		result.Err = lastErr.Error()
	}
	result.TimedOut = isTimeoutErr(lastErr)
	s.logResult(result, packet, nil)
	return result
}

// classifyResponse maps a NAS reply onto the result, extracting the Error-Cause
// from NAK responses.
func classifyResponse(result *CoAResult, resp *radius.Packet) {
	if resp == nil {
		return
	}
	result.ResponseCode = resp.Code.String()
	switch resp.Code {
	case radius.CodeDisconnectACK, radius.CodeCoAACK:
		result.Success = true
	case radius.CodeDisconnectNAK, radius.CodeCoANAK:
		result.Success = false
		if cause, err := rfc3576.ErrorCause_Lookup(resp); err == nil {
			result.ErrorCause = int(cause)
			result.ErrorCauseText = cause.String()
		}
	default:
		result.Success = false
	}
}

// applyIdentity sets the RFC 5176 §3 identification attributes that are present
// in id onto packet.
func applyIdentity(packet *radius.Packet, id SessionIdentity) error {
	if id.Username != "" {
		if err := rfc2865.UserName_SetString(packet, id.Username); err != nil {
			return fmt.Errorf("set User-Name: %w", err)
		}
	}
	if id.NasIP != "" {
		if ip := net.ParseIP(id.NasIP); ip != nil {
			if err := rfc2865.NASIPAddress_Set(packet, ip); err != nil {
				return fmt.Errorf("set NAS-IP-Address: %w", err)
			}
		}
	}
	if id.NasIdentifier != "" {
		if err := rfc2865.NASIdentifier_Set(packet, []byte(id.NasIdentifier)); err != nil {
			return fmt.Errorf("set NAS-Identifier: %w", err)
		}
	}
	if id.AcctSessionID != "" {
		if err := rfc2866.AcctSessionID_SetString(packet, id.AcctSessionID); err != nil {
			return fmt.Errorf("set Acct-Session-Id: %w", err)
		}
	}
	if id.FramedIP != "" {
		if ip := net.ParseIP(id.FramedIP); ip != nil {
			if err := rfc2865.FramedIPAddress_Set(packet, ip); err != nil {
				return fmt.Errorf("set Framed-IP-Address: %w", err)
			}
		}
	}
	if id.CallingStation != "" {
		if err := rfc2865.CallingStationID_SetString(packet, id.CallingStation); err != nil {
			return fmt.Errorf("set Calling-Station-Id: %w", err)
		}
	}
	if id.NasPort != nil {
		if err := rfc2865.NASPort_Set(packet, rfc2865.NASPort(*id.NasPort)); err != nil {
			return fmt.Errorf("set NAS-Port: %w", err)
		}
	}
	if id.NasPortID != "" {
		if err := rfc2869.NASPortID_Set(packet, []byte(id.NasPortID)); err != nil {
			return fmt.Errorf("set NAS-Port-Id: %w", err)
		}
	}
	return nil
}

// isTimeoutErr reports whether err represents a request timeout (a context
// deadline or a net.Error timeout), which is the condition that warrants a
// retransmission.
func isTimeoutErr(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout()
	}
	return false
}

// logResult emits a structured zap record for the exchange, mirroring the
// logging conventions used elsewhere in the radius package.
func (s *CoAService) logResult(result *CoAResult, request, response *radius.Packet) {
	fields := []zap.Field{
		zap.String("namespace", "radius"),
		zap.String("action", string(result.Action)),
		zap.String("target", result.Target),
		zap.String("username", result.Username),
		zap.String("acct_session_id", result.AcctSessionID),
		zap.Int("identifier", result.Identifier),
		zap.Int("attempts", result.Attempts),
		zap.Duration("rtt", result.RTT),
	}
	switch {
	case result.Success:
		fields = append(fields, zap.String("response", result.ResponseCode))
		zap.L().Info("radius coa request acknowledged", fields...)
	case result.TimedOut:
		fields = append(fields, zap.String("error", result.Err))
		zap.L().Warn("radius coa request timed out", fields...)
	default:
		fields = append(fields,
			zap.String("response", result.ResponseCode),
			zap.Int("error_cause", result.ErrorCause),
			zap.String("error_cause_text", result.ErrorCauseText),
		)
		if result.Err != "" {
			fields = append(fields, zap.String("error", result.Err))
		}
		zap.L().Warn("radius coa request rejected", fields...)
	}
}
