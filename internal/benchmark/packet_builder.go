package benchmark

import (
	"fmt"
	"net"

	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2866"
	"layeh.com/radius/rfc2869"
)

// PacketBuilder provides methods to construct RADIUS packets for testing.
// It encapsulates common packet construction logic and ensures consistency
// across different test scenarios.
type PacketBuilder struct {
	Secret        []byte
	NASIdentifier string
	NASIP         net.IP
	NASPort       uint32
	NASPortType   uint32
}

// NewPacketBuilder creates a new PacketBuilder with the given secret and NAS configuration.
//
// Parameters:
//   - secret: RADIUS shared secret for packet authentication
//   - nasIdentifier: NAS-Identifier attribute value (e.g., "benchmark-test")
//   - nasIP: NAS-IP-Address attribute value
//
// Returns:
//   - *PacketBuilder: Configured packet builder instance
func NewPacketBuilder(secret, nasIdentifier string, nasIP string) *PacketBuilder {
	return &PacketBuilder{
		Secret:        []byte(secret),
		NASIdentifier: nasIdentifier,
		NASIP:         parseIP(nasIP),
		NASPort:       0,
		NASPortType:   0,
	}
}

// BuildAuthRequest constructs an Access-Request packet for authentication testing.
//
// Currently supports PAP authentication only. The packet includes standard
// attributes: User-Name, User-Password, NAS-Identifier, NAS-IP-Address,
// NAS-Port, NAS-Port-Type, NAS-Port-ID, Called-Station-ID, Calling-Station-ID.
//
// Parameters:
//   - username: User-Name attribute value
//   - password: User-Password attribute value (will be encrypted)
//   - callingStationID: Calling-Station-ID (typically MAC address)
//
// Returns:
//   - *radius.Packet: Constructed Access-Request packet
//   - error: Error if packet construction fails
func (pb *PacketBuilder) BuildAuthRequest(username, password, callingStationID string) (*radius.Packet, error) {
	pkt := radius.New(radius.CodeAccessRequest, pb.Secret)

	if err := rfc2865.UserName_SetString(pkt, username); err != nil {
		return nil, fmt.Errorf("failed to set User-Name: %w", err)
	}

	if err := rfc2865.UserPassword_SetString(pkt, password); err != nil {
		return nil, fmt.Errorf("failed to set User-Password: %w", err)
	}

	if err := rfc2865.NASIdentifier_Set(pkt, []byte(pb.NASIdentifier)); err != nil {
		return nil, fmt.Errorf("failed to set NAS-Identifier: %w", err)
	}

	if err := rfc2865.NASIPAddress_Set(pkt, pb.NASIP); err != nil {
		return nil, fmt.Errorf("failed to set NAS-IP-Address: %w", err)
	}

	if err := rfc2865.NASPort_Set(pkt, rfc2865.NASPort(pb.NASPort)); err != nil {
		return nil, fmt.Errorf("failed to set NAS-Port: %w", err)
	}

	if err := rfc2865.NASPortType_Set(pkt, rfc2865.NASPortType(pb.NASPortType)); err != nil {
		return nil, fmt.Errorf("failed to set NAS-Port-Type: %w", err)
	}

	if err := rfc2869.NASPortID_Set(pkt, []byte("slot=2;subslot=2;port=22;vlanid=100;")); err != nil {
		return nil, fmt.Errorf("failed to set NAS-Port-ID: %w", err)
	}

	if err := rfc2865.CalledStationID_SetString(pkt, "11:11:11:11:11:11"); err != nil {
		return nil, fmt.Errorf("failed to set Called-Station-ID: %w", err)
	}

	if err := rfc2865.CallingStationID_SetString(pkt, callingStationID); err != nil {
		return nil, fmt.Errorf("failed to set Calling-Station-ID: %w", err)
	}

	return pkt, nil
}

// BuildAcctRequest constructs an Accounting-Request packet for accounting testing.
//
// The packet type is determined by the acctStatusType parameter (Start/Stop/Interim-Update).
// Standard attributes are included: User-Name, NAS-Identifier, NAS-IP-Address, NAS-Port,
// NAS-Port-Type, NAS-Port-ID, Called-Station-ID, Calling-Station-ID, Acct-Session-ID,
// Acct-Status-Type, Framed-IP-Address, and traffic counters.
//
// Parameters:
//   - username: User-Name attribute value
//   - framedIP: Framed-IP-Address (user's assigned IP)
//   - callingStationID: Calling-Station-ID (typically MAC address)
//   - sessionID: Acct-Session-ID (unique session identifier)
//   - acctStatusType: Acct-Status-Type (Start/Stop/Interim-Update)
//
// Returns:
//   - *radius.Packet: Constructed Accounting-Request packet
//   - error: Error if packet construction fails
func (pb *PacketBuilder) BuildAcctRequest(username, framedIP, callingStationID, sessionID string, acctStatusType rfc2866.AcctStatusType) (*radius.Packet, error) {
	pkt := radius.New(radius.CodeAccountingRequest, pb.Secret)

	if err := rfc2865.UserName_SetString(pkt, username); err != nil {
		return nil, fmt.Errorf("failed to set User-Name: %w", err)
	}

	if err := rfc2865.NASIdentifier_Set(pkt, []byte(pb.NASIdentifier)); err != nil {
		return nil, fmt.Errorf("failed to set NAS-Identifier: %w", err)
	}

	if err := rfc2865.NASIPAddress_Set(pkt, pb.NASIP); err != nil {
		return nil, fmt.Errorf("failed to set NAS-IP-Address: %w", err)
	}

	if err := rfc2865.NASPort_Set(pkt, rfc2865.NASPort(pb.NASPort)); err != nil {
		return nil, fmt.Errorf("failed to set NAS-Port: %w", err)
	}

	if err := rfc2865.NASPortType_Set(pkt, rfc2865.NASPortType(pb.NASPortType)); err != nil {
		return nil, fmt.Errorf("failed to set NAS-Port-Type: %w", err)
	}

	if err := rfc2869.NASPortID_Set(pkt, []byte("slot=2;subslot=2;port=22;vlanid=100;")); err != nil {
		return nil, fmt.Errorf("failed to set NAS-Port-ID: %w", err)
	}

	if err := rfc2865.CalledStationID_Set(pkt, []byte("11:11:11:11:11:11")); err != nil {
		return nil, fmt.Errorf("failed to set Called-Station-ID: %w", err)
	}

	if err := rfc2865.CallingStationID_Set(pkt, []byte(callingStationID)); err != nil {
		return nil, fmt.Errorf("failed to set Calling-Station-ID: %w", err)
	}

	if err := rfc2866.AcctSessionID_SetString(pkt, sessionID); err != nil {
		return nil, fmt.Errorf("failed to set Acct-Session-ID: %w", err)
	}

	if err := rfc2866.AcctInputOctets_Set(pkt, 0); err != nil {
		return nil, fmt.Errorf("failed to set Acct-Input-Octets: %w", err)
	}

	if err := rfc2866.AcctOutputOctets_Set(pkt, 0); err != nil {
		return nil, fmt.Errorf("failed to set Acct-Output-Octets: %w", err)
	}

	if err := rfc2866.AcctInputPackets_Set(pkt, 0); err != nil {
		return nil, fmt.Errorf("failed to set Acct-Input-Packets: %w", err)
	}

	if err := rfc2866.AcctOutputPackets_Set(pkt, 0); err != nil {
		return nil, fmt.Errorf("failed to set Acct-Output-Packets: %w", err)
	}

	if err := rfc2865.FramedIPAddress_Set(pkt, parseIP(framedIP)); err != nil {
		return nil, fmt.Errorf("failed to set Framed-IP-Address: %w", err)
	}

	if err := rfc2866.AcctStatusType_Set(pkt, acctStatusType); err != nil {
		return nil, fmt.Errorf("failed to set Acct-Status-Type: %w", err)
	}

	return pkt, nil
}

// parseIP parses an IP address string and returns a net.IP.
// Returns 0.0.0.0 if parsing fails.
func parseIP(ipStr string) net.IP {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return net.ParseIP("0.0.0.0")
	}
	return ip
}
