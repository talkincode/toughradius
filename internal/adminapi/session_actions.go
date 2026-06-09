package adminapi

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// sessionCoAService builds the RFC 5176 Dynamic Authorization client used by the
// operator-initiated session action endpoints. It is a package variable (rather
// than an inline constructor) so tests can substitute a fast-timeout client to
// exercise the timeout path without waiting on real retransmission budgets; this
// mirrors the testOperatorResolver seam used elsewhere in this package.
//
// The default bounds a single exchange to at most two transmissions of 3s each
// (~6s worst case) so the synchronous HTTP request returns promptly while still
// retransmitting once on packet loss.
var sessionCoAService = func() *radiusd.CoAService {
	return radiusd.NewCoAService(nil,
		radiusd.WithCoATimeout(3*time.Second),
		radiusd.WithCoARetries(1))
}

// coaChangePayload is the request body for POST /sessions/:id/coa. Every field is
// optional, but at least one must be supplied so the CoA-Request carries a real
// authorization change (RFC 5176 §2.2); a CoA-Request with only identification
// attributes is rejected with 400 NO_CHANGES.
type coaChangePayload struct {
	// SessionTimeout sets Session-Timeout (#27) in seconds. A pointer is used so
	// that an explicit 0 (terminate at next checkpoint) is distinguishable from
	// "field omitted".
	SessionTimeout *uint32 `json:"session_timeout"`
	// FilterID sets Filter-Id (#11), applying a named NAS access-control filter to
	// the live session. RADIUS string attributes are capped at 253 octets.
	FilterID *string `json:"filter_id" validate:"omitempty,max=253"`
}

// attributeSetters converts the supplied changes into radiusd.AttributeSetter
// values in a stable order.
func (p coaChangePayload) attributeSetters() []radiusd.AttributeSetter {
	var setters []radiusd.AttributeSetter
	if p.SessionTimeout != nil {
		setters = append(setters, radiusd.WithSessionTimeout(*p.SessionTimeout))
	}
	if p.FilterID != nil {
		setters = append(setters, radiusd.WithFilterID(*p.FilterID))
	}
	return setters
}

// coaActionResponse is the snake_case JSON projection of radiusd.CoAResult
// returned to the management UI. It deliberately decouples the wire format from
// the internal result struct so the admin API stays consistent with the rest of
// the package.
type coaActionResponse struct {
	Action         string `json:"action"`
	Target         string `json:"target"`
	Username       string `json:"username"`
	AcctSessionID  string `json:"acct_session_id"`
	Identifier     int    `json:"identifier"`
	Success        bool   `json:"success"`
	ResponseCode   string `json:"response_code,omitempty"`
	ErrorCause     int    `json:"error_cause,omitempty"`
	ErrorCauseText string `json:"error_cause_text,omitempty"`
	Attempts       int    `json:"attempts"`
	RTTMillis      int64  `json:"rtt_ms"`
	TimedOut       bool   `json:"timed_out"`
	Error          string `json:"error,omitempty"`
}

func newCoAActionResponse(r *radiusd.CoAResult) coaActionResponse {
	return coaActionResponse{
		Action:         string(r.Action),
		Target:         r.Target,
		Username:       r.Username,
		AcctSessionID:  r.AcctSessionID,
		Identifier:     r.Identifier,
		Success:        r.Success,
		ResponseCode:   r.ResponseCode,
		ErrorCause:     r.ErrorCause,
		ErrorCauseText: r.ErrorCauseText,
		Attempts:       r.Attempts,
		RTTMillis:      r.RTT.Milliseconds(),
		TimedOut:       r.TimedOut,
		Error:          r.Err,
	}
}

// resolveSessionAndNAS loads the online session identified by id and the NAS that
// hosts it, returning a CoA target and session identity for a Dynamic
// Authorization exchange. On failure it writes the appropriate error response and
// returns ok=false; the caller then returns the (already written) response.
func resolveSessionAndNAS(c echo.Context, id int64) (radiusd.CoATarget, radiusd.SessionIdentity, bool) {
	var session domain.RadiusOnline
	if err := GetDB(c).First(&session, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			_ = fail(c, http.StatusNotFound, "NOT_FOUND", "Session not found", nil) //nolint:errcheck
		} else {
			_ = fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to load session", err.Error()) //nolint:errcheck
		}
		return radiusd.CoATarget{}, radiusd.SessionIdentity{}, false
	}

	var nas domain.NetNas
	if err := GetDB(c).Where("ipaddr = ?", session.NasAddr).First(&nas).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			_ = fail(c, http.StatusUnprocessableEntity, "NAS_NOT_FOUND", //nolint:errcheck
				"NAS device for this session is not configured; cannot send request", nil)
		} else {
			_ = fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to load NAS", err.Error()) //nolint:errcheck
		}
		return radiusd.CoATarget{}, radiusd.SessionIdentity{}, false
	}

	return radiusd.CoATargetFromNas(&nas), radiusd.SessionIdentityFromOnline(&session), true
}

// DisconnectOnlineSession sends a RADIUS Disconnect-Request (RFC 5176 §2.1) to the
// NAS hosting the session, forcing the user offline, and returns the structured
// exchange result (ACK/NAK/timeout, Error-Cause, attempts, RTT).
//
// Unlike DELETE /sessions/:id — which removes the local record and best-effort
// notifies the NAS asynchronously — this endpoint performs a synchronous,
// retry-bounded exchange and reports the NAS's actual response, giving the
// operator immediate confirmation. The local session record is left untouched;
// the NAS is expected to emit an Accounting-Stop that clears it.
//
// A NAK or timeout is reported with HTTP 200 and success=false in the body (the
// exchange genuinely ran; its outcome is the payload). Build/transport setup
// failures return 4xx/5xx.
//
// Authorization: admin/super only (requireAdmin), since it disrupts live users.
//
// @Summary Force a session offline via RADIUS Disconnect-Request
// @Tags OnlineSession
// @Param id path int true "Session ID"
// @Success 200 {object} Response
// @Router /api/v1/sessions/{id}/disconnect [post]
func DisconnectOnlineSession(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid Session ID", nil)
	}

	target, identity, resolved := resolveSessionAndNAS(c, id)
	if !resolved {
		return nil
	}

	result, err := sessionCoAService().Disconnect(c.Request().Context(), target, identity)
	if err != nil {
		return fail(c, http.StatusUnprocessableEntity, "COA_REQUEST_FAILED", "Failed to build Disconnect request", err.Error())
	}

	logSessionAction(c, result)
	if err := persistSessionActionAudit(c, id, identity, result); err != nil {
		return fail(c, http.StatusInternalServerError, "AUDIT_PERSIST_FAILED",
			"Disconnect request completed but failed to persist audit record", err.Error())
	}
	return ok(c, newCoAActionResponse(result))
}

// ChangeOnlineSessionAuthorization sends a RADIUS CoA-Request (RFC 5176 §2.2) to
// the session's NAS to change the live session's authorization (for example a new
// Session-Timeout or Filter-Id) without disconnecting it, returning the structured
// exchange result. At least one change must be supplied.
//
// As with DisconnectOnlineSession, a NAK or timeout is reported with HTTP 200 and
// success=false; the structured body carries the Error-Cause and timing.
//
// Authorization: admin/super only (requireAdmin).
//
// @Summary Change a live session's authorization via RADIUS CoA-Request
// @Tags OnlineSession
// @Param id path int true "Session ID"
// @Param body body coaChangePayload true "Authorization changes"
// @Success 200 {object} Response
// @Router /api/v1/sessions/{id}/coa [post]
func ChangeOnlineSessionAuthorization(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid Session ID", nil)
	}

	var payload coaChangePayload
	if err := c.Bind(&payload); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse CoA parameters", nil)
	}
	if err := c.Validate(&payload); err != nil {
		return handleValidationError(c, err)
	}

	changes := payload.attributeSetters()
	if len(changes) == 0 {
		return fail(c, http.StatusBadRequest, "NO_CHANGES",
			"At least one authorization change (session_timeout, filter_id) is required", nil)
	}

	target, identity, resolved := resolveSessionAndNAS(c, id)
	if !resolved {
		return nil
	}

	result, err := sessionCoAService().CoA(c.Request().Context(), target, identity, changes...)
	if err != nil {
		return fail(c, http.StatusUnprocessableEntity, "COA_REQUEST_FAILED", "Failed to build CoA request", err.Error())
	}

	logSessionAction(c, result)
	if err := persistSessionActionAudit(c, id, identity, result); err != nil {
		return fail(c, http.StatusInternalServerError, "AUDIT_PERSIST_FAILED",
			"CoA request completed but failed to persist audit record", err.Error())
	}
	return ok(c, newCoAActionResponse(result))
}

// persistSessionActionAudit writes a durable per-action record required by M2.3,
// capturing who triggered the action, which session was targeted, and the
// structured CoA/Disconnect outcome.
func persistSessionActionAudit(c echo.Context, sessionID int64, identity radiusd.SessionIdentity, r *radiusd.CoAResult) error {
	operatorID := int64(0)
	operatorName := "unknown"
	if op, err := resolveOperatorFromContext(c); err == nil && op != nil {
		operatorID = op.ID
		operatorName = op.Username
	}

	triggeredAt := r.SentAt
	if triggeredAt.IsZero() {
		triggeredAt = time.Now()
	}

	record := domain.RadiusSessionActionAudit{
		SessionID:      sessionID,
		AcctSessionID:  r.AcctSessionID,
		Action:         string(r.Action),
		Username:       r.Username,
		NasAddr:        identity.NasIP,
		Target:         r.Target,
		OperatorID:     operatorID,
		OperatorName:   operatorName,
		OperatorIP:     c.RealIP(),
		Success:        r.Success,
		ResponseCode:   r.ResponseCode,
		ErrorCause:     r.ErrorCause,
		ErrorCauseText: r.ErrorCauseText,
		Attempts:       r.Attempts,
		RTTMillis:      r.RTT.Milliseconds(),
		TimedOut:       r.TimedOut,
		Error:          r.Err,
		TriggeredAt:    triggeredAt,
	}

	return GetDB(c).Create(&record).Error
}

// logSessionAction emits an audit log line recording which operator triggered a
// CoA/Disconnect and its outcome. Durable, queryable audit persistence is
// additionally written via persistSessionActionAudit; this log line provides
// immediate operational traceability.
func logSessionAction(c echo.Context, r *radiusd.CoAResult) {
	operator := "unknown"
	if op, err := resolveOperatorFromContext(c); err == nil && op != nil {
		operator = op.Username
	}
	fields := []zap.Field{
		zap.String("namespace", "adminapi"),
		zap.String("operator", operator),
		zap.String("action", string(r.Action)),
		zap.String("target", r.Target),
		zap.String("username", r.Username),
		zap.String("acct_session_id", r.AcctSessionID),
		zap.Bool("success", r.Success),
		zap.String("response_code", r.ResponseCode),
		zap.Int("attempts", r.Attempts),
		zap.Int("error_cause", r.ErrorCause),
	}
	if r.Success {
		zap.L().Info("operator triggered radius dynamic authorization", fields...)
	} else {
		zap.L().Warn("operator triggered radius dynamic authorization failed", fields...)
	}
}
