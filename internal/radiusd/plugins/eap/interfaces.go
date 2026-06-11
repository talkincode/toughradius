package eap

import (
	"context"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"layeh.com/radius"
)

// EAP code constants
const (
	CodeRequest  = 1 // EAP Request message
	CodeResponse = 2 // EAP Response message
	CodeSuccess  = 3 // Indicates successful authentication
	CodeFailure  = 4 // Indicates failed authentication
)

// EAP type constants
const (
	TypeIdentity     = 1  // Identity
	TypeNotification = 2  // Notification
	TypeNak          = 3  // Response only (Legacy Nak)
	TypeMD5Challenge = 4  // MD5-Challenge
	TypeOTP          = 5  // One-Time Password
	TypeGTC          = 6  // Generic Token Card
	TypeTLS          = 13 // EAP-TLS
	TypeTTLS         = 21 // EAP-TTLS (Tunneled TLS), IANA EAP method type 21
	TypePEAP         = 25 // Protected EAP (PEAP), IANA EAP method type 25
	TypeMSCHAPv2     = 26 // EAP-MSCHAPv2
)

// EAPState holds EAP status data
type EAPState struct {
	Username  string                 // Username
	Challenge []byte                 // Challenge data
	StateID   string                 // StateID (RADIUS State attribute value)
	Method    string                 // EAP method name (eap-md5, eap-mschapv2, etc.)
	Success   bool                   // whether authentication succeeded
	Data      map[string]interface{} // Additional data storage
}

// EAPContext holds EAP authentication context
type EAPContext struct {
	Context        context.Context
	Request        *radius.Request
	ResponseWriter radius.ResponseWriter // RADIUS response writer
	Response       *radius.Packet
	User           *domain.RadiusUser
	NAS            *domain.NetNas
	EAPMessage     *EAPMessage
	EAPState       *EAPState
	IsMacAuth      bool
	Secret         string // RADIUS Secret
	StateManager   EAPStateManager
	PwdProvider    PasswordProvider
	// Verifier is an optional external password authority (e.g. LDAP). When set
	// and Active it verifies inner PAP by binding instead of comparing against a
	// local password; it is nil when no such backend is configured, in which
	// case handlers fall back to PwdProvider.
	Verifier CredentialVerifier
}

// EAPMessage represents the EAP message structure
type EAPMessage struct {
	Code       uint8  // EAP Code
	Identifier uint8  // EAP Identifier
	Length     uint16 // EAP Length
	Type       uint8  // EAP Type
	Data       []byte // EAP Data
}

// EAPHandler defines the EAP authentication handler interface
type EAPHandler interface {
	// Name Returnshandlernames (e.g., "eap-md5", "eap-mschapv2")
	Name() string

	// EAPType returns the EAP type code this handler handles
	EAPType() uint8

	// CanHandle determines whether this handler can process the EAP message
	CanHandle(ctx *EAPContext) bool

	// HandleIdentity Handle EAP-Response/Identity，Send Challenge
	// Returns true if handled and a response was sent; otherwise false
	HandleIdentity(ctx *EAPContext) (bool, error)

	// HandleResponse Handle EAP-Response (Challenge Response)
	// Returns true if authentication succeeded, false otherwise
	HandleResponse(ctx *EAPContext) (bool, error)
}

// EAPStateManager defines the EAP state manager interface
type EAPStateManager interface {
	// GetState get EAP Status
	GetState(stateID string) (*EAPState, error)

	// SetState stores the EAP status
	SetState(stateID string, state *EAPState) error

	// DeleteState Delete EAP Status
	DeleteState(stateID string) error
}

// PasswordProvider defines how to retrieve passwords
type PasswordProvider interface {
	// GetPassword retrieves the user's password (plain or encrypted)
	GetPassword(user *domain.RadiusUser, isMacAuth bool) (string, error)
}

// CredentialVerifier is an optional capability a PasswordProvider may also
// implement: an external password authority (for example an LDAP/AD directory)
// that verifies a presented cleartext password by binding rather than by
// disclosing the stored password.
//
// When Active reports true the directory governs passwords, so only PAP-family
// inner methods (the inner PAP of EAP-TTLS, RFC 5281 §11.2.5) can be served;
// challenge/response inner methods must reject because the server holds no
// secret with which to recompute their response.
type CredentialVerifier interface {
	// Active reports whether external verification is enabled.
	Active() bool
	// VerifyCleartext authenticates username with the presented cleartext
	// password, returning nil on success.
	VerifyCleartext(ctx context.Context, username, password string) error
}
