package radiusd

import (
	"encoding/binary"
	"net"
	"strings"
	"testing"

	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2866"
	"layeh.com/radius/rfc2869"
)

func TestStringType(t *testing.T) {
	tests := []struct {
		name     string
		attrType radius.Type
		expected string
	}{
		{
			name:     "UserName 类型",
			attrType: rfc2865.UserName_Type,
			expected: "UserName",
		},
		{
			name:     "UserPassword 类型",
			attrType: rfc2865.UserPassword_Type,
			expected: "UserPassword",
		},
		{
			name:     "NASIPAddress 类型",
			attrType: rfc2865.NASIPAddress_Type,
			expected: "NASIPAddress",
		},
		{
			name:     "AcctStatusType 类型",
			attrType: rfc2866.AcctStatusType_Type,
			expected: "AcctStatusType",
		},
		{
			name:     "EAPMessage 类型",
			attrType: rfc2869.EAPMessage_Type,
			expected: "EAPMessage",
		},
		{
			name:     "未知类型",
			attrType: radius.Type(255),
			expected: "255",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StringType(tt.attrType)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestFormatType(t *testing.T) {
	tests := []struct {
		name     string
		attrType radius.Type
		data     []byte
		expected string
	}{
		{
			name:     "字符串格式",
			attrType: rfc2865.UserName_Type,
			data:     []byte("testuser"),
			expected: "testuser",
		},
		{
			name:     "IPv4 格式",
			attrType: rfc2865.NASIPAddress_Type,
			data:     []byte{192, 168, 1, 1},
			expected: "192.168.1.1",
		},
		{
			name:     "UInt32 格式",
			attrType: rfc2865.SessionTimeout_Type,
			data: func() []byte {
				b := make([]byte, 4)
				binary.BigEndian.PutUint32(b, 3600)
				return b
			}(),
			expected: "3600",
		},
		{
			name:     "十六进制格式（未知类型）",
			attrType: radius.Type(255),
			data:     []byte{0xde, 0xad, 0xbe, 0xef},
			expected: "deadbeef",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatType(tt.attrType, tt.data)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestEapMessageFormat(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		contains []string
	}{
		{
			name: "EAP Request Identity",
			data: []byte{
				eap.CodeRequest, // Code
				0x01,            // Identifier
				0x00, 0x05,      // Length
				eap.TypeIdentity, // Type
			},
			contains: []string{"Code=1", "Type=1"},
		},
		{
			name: "EAP Response Identity",
			data: []byte{
				eap.CodeResponse, // Code
				0x02,             // Identifier
				0x00, 0x0a,       // Length
				eap.TypeIdentity,        // Type
				't', 'e', 's', 't', 'u', // Data
			},
			contains: []string{"Code=2", "Type=1"},
		},
		{
			name:     "短数据",
			data:     []byte{0x01, 0x02},
			contains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EapMessageFormat(tt.data)
			for _, substring := range tt.contains {
				if !strings.Contains(result, substring) {
					t.Errorf("expected result to contain %s, got: %s", substring, result)
				}
			}
		})
	}
}

func TestFmtRequest(t *testing.T) {
	// 创建测试 RADIUS 请求
	packet := radius.New(radius.CodeAccessRequest, []byte("secret"))
	rfc2865.UserName_SetString(packet, "testuser")
	rfc2865.NASIPAddress_Set(packet, net.IPv4(192, 168, 1, 1))

	req := &radius.Request{
		Packet:     packet,
		RemoteAddr: &net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 1812},
		LocalAddr:  &net.UDPAddr{IP: net.ParseIP("192.168.1.1"), Port: 1812},
	}

	result := FmtRequest(req)

	// 验证输出包含关键信息
	expectedStrings := []string{
		"RADIUS Request",
		"10.0.0.1",
		"192.168.1.1",
		"Identifier",
		"Code",
		"Attributes",
		"UserName: testuser",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(result, expected) {
			t.Errorf("expected output to contain '%s', got:\n%s", expected, result)
		}
	}
}

func TestFmtRequestNil(t *testing.T) {
	result := FmtRequest(nil)
	if result != "" {
		t.Errorf("expected empty string for nil request, got: %s", result)
	}
}

func TestFmtResponse(t *testing.T) {
	// 创建测试 RADIUS 响应
	packet := radius.New(radius.CodeAccessAccept, []byte("secret"))
	rfc2865.SessionTimeout_Set(packet, 3600)
	rfc2865.ReplyMessage_SetString(packet, "Welcome")

	remoteAddr := &net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 1812}

	result := FmtResponse(packet, remoteAddr)

	// 验证输出包含关键信息
	expectedStrings := []string{
		"RADIUS Response",
		"10.0.0.1",
		"Identifier",
		"Code",
		"Attributes",
		"SessionTimeout: 3600",
		"ReplyMessage: Welcome",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(result, expected) {
			t.Errorf("expected output to contain '%s', got:\n%s", expected, result)
		}
	}
}

func TestFmtResponseNil(t *testing.T) {
	remoteAddr := &net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 1812}
	result := FmtResponse(nil, remoteAddr)
	if result != "" {
		t.Errorf("expected empty string for nil packet, got: %s", result)
	}
}

func TestFmtPacket(t *testing.T) {
	// 创建测试包
	packet := radius.New(radius.CodeAccessRequest, []byte("secret"))
	rfc2865.UserName_SetString(packet, "testuser")
	rfc2866.AcctSessionID_SetString(packet, "session123")

	result := FmtPacket(packet)

	// 验证输出
	expectedStrings := []string{
		"RADIUS Packet",
		"Identifier",
		"Code",
		"Authenticator",
		"Attributes",
		"UserName: testuser",
		"AcctSessionID: session123",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(result, expected) {
			t.Errorf("expected output to contain '%s', got:\n%s", expected, result)
		}
	}
}

func TestFmtPacketNil(t *testing.T) {
	result := FmtPacket(nil)
	if result != "" {
		t.Errorf("expected empty string for nil packet, got: %s", result)
	}
}

func TestLength(t *testing.T) {
	tests := []struct {
		name   string
		packet *radius.Packet
		minLen int
	}{
		{
			name:   "空包",
			packet: radius.New(radius.CodeAccessRequest, []byte("secret")),
			minLen: 20, // 基础头部长度
		},
		{
			name: "带属性的包",
			packet: func() *radius.Packet {
				p := radius.New(radius.CodeAccessRequest, []byte("secret"))
				rfc2865.UserName_SetString(p, "testuser")
				rfc2865.NASIPAddress_Set(p, net.IPv4(192, 168, 1, 1))
				return p
			}(),
			minLen: 20, // 至少是头部长度
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			length := Length(tt.packet)
			if length < tt.minLen {
				t.Errorf("expected length >= %d, got %d", tt.minLen, length)
			}
		})
	}
}

func TestLengthNil(t *testing.T) {
	result := Length(nil)
	if result != 0 {
		t.Errorf("expected length 0 for nil packet, got: %d", result)
	}
}

func TestFmtRequestWithAcctStatusType(t *testing.T) {
	// 测试带有 AcctStatusType 的请求（特殊格式化）
	packet := radius.New(radius.CodeAccountingRequest, []byte("secret"))
	rfc2866.AcctStatusType_Set(packet, rfc2866.AcctStatusType_Value_Start)
	rfc2865.UserName_SetString(packet, "testuser")

	req := &radius.Request{
		Packet:     packet,
		RemoteAddr: &net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 1813},
		LocalAddr:  &net.UDPAddr{IP: net.ParseIP("192.168.1.1"), Port: 1813},
	}

	result := FmtRequest(req)

	// 验证 AcctStatusType 被特殊格式化
	if !strings.Contains(result, "AcctStatusType") {
		t.Errorf("expected output to contain AcctStatusType, got:\n%s", result)
	}
	if !strings.Contains(result, "Start") {
		t.Errorf("expected output to contain 'Start' status, got:\n%s", result)
	}
}

func TestFmtRequestWithVendorSpecific(t *testing.T) {
	// 测试带有厂商特定属性的请求
	packet := radius.New(radius.CodeAccessRequest, []byte("secret"))

	// 创建厂商特定属性
	vendorAttr := make([]byte, 10)
	binary.BigEndian.PutUint16(vendorAttr[2:4], 9) // Vendor ID (Cisco)
	vendorAttr[4] = 1                              // Vendor Type
	vendorAttr[5] = 4                              // Length

	packet.Add(rfc2865.VendorSpecific_Type, vendorAttr)
	rfc2865.UserName_SetString(packet, "testuser")

	req := &radius.Request{
		Packet:     packet,
		RemoteAddr: &net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 1812},
		LocalAddr:  &net.UDPAddr{IP: net.ParseIP("192.168.1.1"), Port: 1812},
	}

	result := FmtRequest(req)

	// 验证厂商特定属性被正确格式化
	if !strings.Contains(result, "VendorSpecific") {
		t.Errorf("expected output to contain VendorSpecific, got:\n%s", result)
	}
}

func TestStringFormat(t *testing.T) {
	data := []byte("teststring")
	result := StringFormat(data)
	if result != "teststring" {
		t.Errorf("expected 'teststring', got %s", result)
	}
}

func TestHexFormat(t *testing.T) {
	data := []byte{0xde, 0xad, 0xbe, 0xef}
	result := HexFormat(data)
	if result != "deadbeef" {
		t.Errorf("expected 'deadbeef', got %s", result)
	}
}

func TestUInt32Format(t *testing.T) {
	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, 12345)
	result := UInt32Format(data)
	if result != "12345" {
		t.Errorf("expected '12345', got %s", result)
	}
}

func TestIpv4Format(t *testing.T) {
	data := []byte{192, 168, 1, 100}
	result := Ipv4Format(data)
	if result != "192.168.1.100" {
		t.Errorf("expected '192.168.1.100', got %s", result)
	}
}
