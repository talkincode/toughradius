package radiusd

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"strings"

	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2866"
	"layeh.com/radius/rfc2868"
	"layeh.com/radius/rfc2869"
	"layeh.com/radius/rfc3162"
	"layeh.com/radius/rfc3576"
	"layeh.com/radius/rfc4849"
)

type AttrFormatFunc = func(s []byte) string

var StringFormat = func(src []byte) string {
	return string(src)
}

var HexFormat = func(src []byte) string {
	return fmt.Sprintf("%x", src)
}

var UInt32Format = func(src []byte) string {
	return strconv.Itoa(int(binary.BigEndian.Uint32(src)))
}

var Ipv4Format = func(src []byte) string {
	return net.IPv4(src[0], src[1], src[2], src[3]).String()
}

var EapMessageFormat = func(attr []byte) string {
	// 解析EAP消息
	eap := &EAPMessage{
		EAPHeader: EAPHeader{
			Code:       attr[0],
			Identifier: attr[1],
			Length:     binary.BigEndian.Uint16(attr[2:4]),
		},
	}
	if len(attr) >= 5 {
		eap.Type = attr[4]
		eap.Data = attr[5:]
	}

	return eap.String()
}

var RadiusTypeMap = map[radius.Type]string{
	rfc2865.UserName_Type:               "UserName",
	rfc2865.UserPassword_Type:           "UserPassword",
	rfc2865.CHAPPassword_Type:           "CHAPPassword",
	rfc2865.NASIPAddress_Type:           "NASIPAddress",
	rfc2865.NASPort_Type:                "NASPort",
	rfc2865.ServiceType_Type:            "ServiceType",
	rfc2865.FramedProtocol_Type:         "FramedProtocol",
	rfc2865.FramedIPAddress_Type:        "FramedIPAddress",
	rfc2865.FramedIPNetmask_Type:        "FramedIPNetmask",
	rfc2865.FramedRouting_Type:          "FramedRouting",
	rfc2865.FilterID_Type:               "FilterID",
	rfc2865.FramedMTU_Type:              "FramedMTU",
	rfc2865.FramedCompression_Type:      "FramedCompression",
	rfc2865.LoginIPHost_Type:            "LoginIPHost",
	rfc2865.LoginService_Type:           "LoginService",
	rfc2865.LoginTCPPort_Type:           "LoginTCPPort",
	rfc2865.ReplyMessage_Type:           "ReplyMessage",
	rfc2865.CallbackNumber_Type:         "CallbackNumber",
	rfc2865.CallbackID_Type:             "CallbackID",
	rfc2865.FramedRoute_Type:            "FramedRoute",
	rfc2865.FramedIPXNetwork_Type:       "FramedIPXNetwork",
	rfc2865.State_Type:                  "State",
	rfc2865.Class_Type:                  "Class",
	rfc2865.VendorSpecific_Type:         "VendorSpecific",
	rfc2865.SessionTimeout_Type:         "SessionTimeout",
	rfc2865.IdleTimeout_Type:            "IdleTimeout",
	rfc2865.TerminationAction_Type:      "TerminationAction",
	rfc2865.CalledStationID_Type:        "CalledStationID",
	rfc2865.CallingStationID_Type:       "CallingStationID",
	rfc2865.NASIdentifier_Type:          "NASIdentifier",
	rfc2865.ProxyState_Type:             "ProxyState",
	rfc2865.LoginLATService_Type:        "LoginLATService",
	rfc2865.LoginLATNode_Type:           "LoginLATNode",
	rfc2865.LoginLATGroup_Type:          "LoginLATGroup",
	rfc2865.FramedAppleTalkLink_Type:    "FramedAppleTalkLink",
	rfc2865.FramedAppleTalkNetwork_Type: "FramedAppleTalkNetwork",
	rfc2865.FramedAppleTalkZone_Type:    "FramedAppleTalkZone",
	rfc2865.CHAPChallenge_Type:          "CHAPChallenge",
	rfc2865.NASPortType_Type:            "NASPortType",
	rfc2865.PortLimit_Type:              "PortLimit",
	rfc2865.LoginLATPort_Type:           "LoginLATPort",
	rfc2866.AcctStatusType_Type:         "AcctStatusType",
	rfc2866.AcctDelayTime_Type:          "AcctDelayTime",
	rfc2866.AcctInputOctets_Type:        "AcctInputOctets",
	rfc2866.AcctOutputOctets_Type:       "AcctOutputOctets",
	rfc2866.AcctSessionID_Type:          "AcctSessionID",
	rfc2866.AcctAuthentic_Type:          "AcctAuthentic",
	rfc2866.AcctSessionTime_Type:        "AcctSessionTime",
	rfc2866.AcctInputPackets_Type:       "AcctInputPackets",
	rfc2866.AcctOutputPackets_Type:      "AcctOutputPackets",
	rfc2866.AcctTerminateCause_Type:     "AcctTerminateCause",
	rfc2866.AcctMultiSessionID_Type:     "AcctMultiSessionID",
	rfc2866.AcctLinkCount_Type:          "AcctLinkCount",
	rfc2869.AcctInputGigawords_Type:     "AcctInputGigawords",
	rfc2869.AcctOutputGigawords_Type:    "AcctOutputGigawords",
	rfc2869.EventTimestamp_Type:         "EventTimestamp",
	rfc2869.ARAPPassword_Type:           "ARAPPassword",
	rfc2869.ARAPFeatures_Type:           "ARAPFeatures",
	rfc2869.ARAPZoneAccess_Type:         "ARAPZoneAccess",
	rfc2869.ARAPSecurity_Type:           "ARAPSecurity",
	rfc2869.ARAPSecurityData_Type:       "ARAPSecurityData",
	rfc2869.PasswordRetry_Type:          "PasswordRetry",
	rfc2869.Prompt_Type:                 "Prompt",
	rfc2869.ConnectInfo_Type:            "ConnectInfo",
	rfc2869.ConfigurationToken_Type:     "ConfigurationToken",
	rfc2869.EAPMessage_Type:             "EAPMessage",
	rfc2869.MessageAuthenticator_Type:   "MessageAuthenticator",
	rfc2869.ARAPChallengeResponse_Type:  "ARAPChallengeResponse",
	rfc2869.AcctInterimInterval_Type:    "AcctInterimInterval",
	rfc2869.NASPortID_Type:              "NASPortID",
	rfc2869.FramedPool_Type:             "FramedPool",
	rfc3162.NASIPv6Address_Type:         "NASIPv6Address",
	rfc3162.FramedInterfaceID_Type:      "FramedInterfaceID",
	rfc3162.FramedIPv6Prefix_Type:       "FramedIPv6Prefix",
	rfc3162.LoginIPv6Host_Type:          "LoginIPv6Host",
	rfc3162.FramedIPv6Route_Type:        "FramedIPv6Route",
	rfc3162.FramedIPv6Pool_Type:         "FramedIPv6Pool",
	rfc3576.ErrorCause_Type:             "ErrorCause",
	rfc4849.NASFilterRule_Type:          "NASFilterRule",
	rfc2868.TunnelType_Type:             "TunnelType",
	rfc2868.TunnelMediumType_Type:       "TunnelMediumType",
	rfc2868.TunnelClientEndpoint_Type:   "TunnelClientEndpoint",
	rfc2868.TunnelServerEndpoint_Type:   "TunnelServerEndpoint",
	rfc2868.TunnelPassword_Type:         "TunnelPassword",
	rfc2868.TunnelPrivateGroupID_Type:   "TunnelPrivateGroupID",
	rfc2868.TunnelAssignmentID_Type:     "TunnelAssignmentID",
	rfc2868.TunnelPreference_Type:       "TunnelPreference",
	rfc2868.TunnelClientAuthID_Type:     "TunnelClientAuthID",
	rfc2868.TunnelServerAuthID_Type:     "TunnelServerAuthID",
}

var RadiusTypeFmtMap = map[radius.Type]AttrFormatFunc{
	rfc2865.UserName_Type:               StringFormat,
	rfc2865.UserPassword_Type:           HexFormat,
	rfc2865.CHAPPassword_Type:           HexFormat,
	rfc2865.NASIPAddress_Type:           Ipv4Format,
	rfc2865.NASPort_Type:                UInt32Format,
	rfc2865.ServiceType_Type:            UInt32Format,
	rfc2865.FramedProtocol_Type:         UInt32Format,
	rfc2865.FramedIPAddress_Type:        Ipv4Format,
	rfc2865.FramedIPNetmask_Type:        Ipv4Format,
	rfc2865.FramedRouting_Type:          UInt32Format,
	rfc2865.FilterID_Type:               StringFormat,
	rfc2865.FramedMTU_Type:              UInt32Format,
	rfc2865.FramedCompression_Type:      UInt32Format,
	rfc2865.LoginIPHost_Type:            Ipv4Format,
	rfc2865.LoginService_Type:           UInt32Format,
	rfc2865.LoginTCPPort_Type:           UInt32Format,
	rfc2865.ReplyMessage_Type:           StringFormat,
	rfc2865.CallbackNumber_Type:         StringFormat,
	rfc2865.CallbackID_Type:             StringFormat,
	rfc2865.FramedRoute_Type:            StringFormat,
	rfc2865.FramedIPXNetwork_Type:       Ipv4Format,
	rfc2865.State_Type:                  StringFormat,
	rfc2865.Class_Type:                  StringFormat,
	rfc2865.VendorSpecific_Type:         HexFormat,
	rfc2865.SessionTimeout_Type:         UInt32Format,
	rfc2865.IdleTimeout_Type:            UInt32Format,
	rfc2865.TerminationAction_Type:      UInt32Format,
	rfc2865.CalledStationID_Type:        StringFormat,
	rfc2865.CallingStationID_Type:       StringFormat,
	rfc2865.NASIdentifier_Type:          StringFormat,
	rfc2865.ProxyState_Type:             StringFormat,
	rfc2865.LoginLATService_Type:        HexFormat,
	rfc2865.LoginLATNode_Type:           HexFormat,
	rfc2865.LoginLATGroup_Type:          HexFormat,
	rfc2865.FramedAppleTalkLink_Type:    HexFormat,
	rfc2865.FramedAppleTalkNetwork_Type: HexFormat,
	rfc2865.FramedAppleTalkZone_Type:    HexFormat,
	rfc2865.CHAPChallenge_Type:          HexFormat,
	rfc2865.NASPortType_Type:            UInt32Format,
	rfc2865.PortLimit_Type:              HexFormat,
	rfc2865.LoginLATPort_Type:           HexFormat,
	rfc2866.AcctStatusType_Type:         UInt32Format,
	rfc2866.AcctDelayTime_Type:          UInt32Format,
	rfc2866.AcctInputOctets_Type:        UInt32Format,
	rfc2866.AcctOutputOctets_Type:       UInt32Format,
	rfc2866.AcctSessionID_Type:          StringFormat,
	rfc2866.AcctAuthentic_Type:          UInt32Format,
	rfc2866.AcctSessionTime_Type:        UInt32Format,
	rfc2866.AcctInputPackets_Type:       UInt32Format,
	rfc2866.AcctOutputPackets_Type:      UInt32Format,
	rfc2866.AcctTerminateCause_Type:     UInt32Format,
	rfc2866.AcctMultiSessionID_Type:     StringFormat,
	rfc2866.AcctLinkCount_Type:          UInt32Format,
	rfc2869.AcctInputGigawords_Type:     UInt32Format,
	rfc2869.AcctOutputGigawords_Type:    UInt32Format,
	rfc2869.EventTimestamp_Type:         UInt32Format,
	rfc2869.ARAPPassword_Type:           HexFormat,
	rfc2869.ARAPFeatures_Type:           HexFormat,
	rfc2869.ARAPZoneAccess_Type:         HexFormat,
	rfc2869.ARAPSecurity_Type:           HexFormat,
	rfc2869.ARAPSecurityData_Type:       HexFormat,
	rfc2869.PasswordRetry_Type:          HexFormat,
	rfc2869.Prompt_Type:                 HexFormat,
	rfc2869.ConnectInfo_Type:            StringFormat,
	rfc2869.ConfigurationToken_Type:     StringFormat,
	rfc2869.EAPMessage_Type:             EapMessageFormat,
	rfc2869.MessageAuthenticator_Type:   HexFormat,
	rfc2869.ARAPChallengeResponse_Type:  HexFormat,
	rfc2869.AcctInterimInterval_Type:    UInt32Format,
	rfc2869.NASPortID_Type:              StringFormat,
	rfc2869.FramedPool_Type:             StringFormat,
	rfc3162.NASIPv6Address_Type:         HexFormat,
	rfc3162.FramedInterfaceID_Type:      HexFormat,
	rfc3162.FramedIPv6Prefix_Type:       HexFormat,
	rfc3162.LoginIPv6Host_Type:          HexFormat,
	rfc3162.FramedIPv6Route_Type:        HexFormat,
	rfc3162.FramedIPv6Pool_Type:         HexFormat,
	rfc3576.ErrorCause_Type:             UInt32Format,
	rfc4849.NASFilterRule_Type:          StringFormat,
	rfc2868.TunnelType_Type:             UInt32Format,
	rfc2868.TunnelMediumType_Type:       UInt32Format,
	rfc2868.TunnelClientEndpoint_Type:   StringFormat,
	rfc2868.TunnelServerEndpoint_Type:   StringFormat,
	rfc2868.TunnelPassword_Type:         StringFormat,
	rfc2868.TunnelPrivateGroupID_Type:   StringFormat,
	rfc2868.TunnelAssignmentID_Type:     HexFormat,
	rfc2868.TunnelPreference_Type:       HexFormat,
	rfc2868.TunnelClientAuthID_Type:     HexFormat,
	rfc2868.TunnelServerAuthID_Type:     HexFormat,
}

func StringType(t radius.Type) string {
	v, ok := RadiusTypeMap[t]
	if !ok {
		return strconv.Itoa(int(t))
	}
	return v
}

func FormatType(t radius.Type, src radius.Attribute) string {
	vfunc, ok := RadiusTypeFmtMap[t]
	if !ok {
		return HexFormat(src)
	}
	return vfunc(src)
}

func FmtRequest(p *radius.Request) string {
	if p == nil {
		return ""
	}
	var buff = new(strings.Builder)
	buff.WriteString(fmt.Sprintf("RADIUS Request: %s => %s\n", p.RemoteAddr.String(), p.LocalAddr.String()))
	buff.WriteString(fmt.Sprintf("\tIdentifier: %v\n", p.Packet.Identifier))
	buff.WriteString(fmt.Sprintf("\tCode: %v\n", p.Packet.Code))
	buff.WriteString(fmt.Sprintf("\tAuthenticator: %v\n", p.Packet.Authenticator))
	buff.WriteString("\tAttributes:\n")
	for _, attribute := range p.Packet.Attributes {
		if attribute.Type == rfc2866.AcctStatusType_Type {
			buff.WriteString(fmt.Sprintf("\t\t%s: %s\n", StringType(attribute.Type),
				rfc2866.AcctStatusType(binary.BigEndian.Uint32(attribute.Attribute)).String()))
		} else if attribute.Type != rfc2865.VendorSpecific_Type {
			buff.WriteString(fmt.Sprintf("\t\t%s: %s\n", StringType(attribute.Type), FormatType(attribute.Type, attribute.Attribute)))
		} else {
			buff.WriteString(fmt.Sprintf("\t\t%s(%d:%d): %x\n",
				StringType(attribute.Type),
				binary.BigEndian.Uint16(attribute.Attribute[2:4]),
				attribute.Attribute[4:5][0],
				attribute.Attribute[6:]))
		}
	}
	return buff.String()
}

func FmtResponse(p *radius.Packet, RemoteAddr net.Addr) string {
	if p == nil {
		return ""
	}
	var buff = new(strings.Builder)
	buff.WriteString(fmt.Sprintf("RADIUS Response: => %s\n", RemoteAddr.String()))
	buff.WriteString(fmt.Sprintf("\tIdentifier: %v\n", p.Identifier))
	buff.WriteString(fmt.Sprintf("\tCode: %v\n", p.Code))
	buff.WriteString(fmt.Sprintf("\tAuthenticator: %v\n", p.Authenticator))
	buff.WriteString("\tAttributes:\n")
	for _, attribute := range p.Attributes {
		if attribute.Type != rfc2865.VendorSpecific_Type {
			buff.WriteString(fmt.Sprintf("\t\t%s: %s\n", StringType(attribute.Type), FormatType(attribute.Type, attribute.Attribute)))
		} else {
			buff.WriteString(fmt.Sprintf("\t\t%s(%d:%d): %x\n",
				StringType(attribute.Type),
				binary.BigEndian.Uint16(attribute.Attribute[2:4]),
				attribute.Attribute[4:5][0],
				attribute.Attribute[6:]))
		}
	}
	return buff.String()
}

func FmtPacket(p *radius.Packet) string {
	if p == nil {
		return ""
	}
	var buff = new(strings.Builder)
	buff.WriteString("RADIUS Packet: \n")
	buff.WriteString(fmt.Sprintf("\tIdentifier: %v\n", p.Identifier))
	buff.WriteString(fmt.Sprintf("\tCode: %v\n", p.Code))
	buff.WriteString(fmt.Sprintf("\tAuthenticator: %s\n", HexFormat(p.Authenticator[:])))
	buff.WriteString("\tAttributes:\n")
	for _, attribute := range p.Attributes {
		if attribute.Type != rfc2865.VendorSpecific_Type {
			buff.WriteString(fmt.Sprintf("\t\t%s: %s\n", StringType(attribute.Type), FormatType(attribute.Type, attribute.Attribute)))
		} else {
			buff.WriteString(fmt.Sprintf("\t\t%s(%d:%d): %x\n",
				StringType(attribute.Type),
				binary.BigEndian.Uint16(attribute.Attribute[2:4]),
				attribute.Attribute[4:5][0],
				attribute.Attribute[6:]))
		}
	}
	return buff.String()
}

func Length(p *radius.Packet) int {
	if p == nil {
		return 0
	}
	var l = 20
	for _, a := range p.Attributes {
		l += len(a.Attribute)
	}
	return l
}
