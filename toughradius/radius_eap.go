package toughradius

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"

	"github.com/talkincode/toughradius/v8/app"
	"github.com/talkincode/toughradius/v8/common"
	"github.com/talkincode/toughradius/v8/common/zaplog/log"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2869"
)

const (
	EapMd5Method      = "eap-md5"
	EapMschapv2Method = "eap-mschapv2"
	EapTlsMethod      = "eap-tls"
	EapPeapMethod     = "eap-peap"
	EapTtlsMethod     = "eap-ttls"
	EapGtcMethod      = "eap-gtc"
	EapSimMethod      = "eap-sim"
	EapAkaMethod      = "eap-aka"
	EapFastMethod     = "eap-fast"
	EapPaxMethod      = "eap-pax"
	EapPskMethod      = "eap-psk"
	EapSakeMethod     = "eap-sake"
	EapIkev2Method    = "eap-ikev2"
	EapTncMethod      = "eap-tnc"
)

const (
	EAPCodeRequest      = 1  // EAP Request message
	EAPCodeResponse     = 2  // EAP Response message
	EAPCodeSuccess      = 3  // Indicates successful authentication
	EAPCodeFailure      = 4  // Indicates failed authentication
	EAPCodeNakNak       = 5  // Used by the peer to negotiate the authentication method (Response only)
	EAPCodeMD5Challenge = 6  // MD5-Challenge EAP method
	EAPCodeOTP          = 7  // One-Time Password (OTP) EAP method
	EAPCodeGTC          = 8  // Generic Token Card (GTC) EAP method
	EAPCodeTLSv1        = 13 // EAP-TLS method, using TLSv1
	EAPCodeMSCHAPv2     = 26 // EAP method for Microsoft Challenge Handshake Authentication Protocol version 2
	EAPCodeSIM          = 18 // EAP-SIM method for GSM networks
	EAPCodeAKA          = 23 // EAP-AKA method for UMTS authentication and key agreement
	EAPCodePEAP         = 25 // Protected EAP (PEAP), a method that creates an encrypted channel to protect transmitted information
	EAPCodeTTLS         = 21 // Tunneled Transport Layer Security (TTLS) EAP method
	EAPCodeFAST         = 43 // Flexible Authentication via Secure Tunneling (EAP-FAST) method
	EAPCodePAX          = 46 // Password Authenticated Exchange (EAP-PAX) method
	EAPCodePSK          = 47 // Pre-Shared Key (EAP-PSK) method
	EAPCodeSAKE         = 48 // SIM Authentication Key Exchange (EAP-SAKE) method
	EAPCodeIKEv2        = 49 // EAP method based on Internet Key Exchange version 2 (EAP-IKEv2)
)

const (
	EAPTypeIdentity     = 1
	EAPTypeNotification = 2
	EAPTypeNak          = 3 // Response only
	EAPTypeMD5Challenge = 4
	EAPTypeOTP          = 5 // One-Time Password
	EAPTypeGTC          = 6 // Generic Token Card
	// EAPTypeTLS 7-9 Reserved
	EAPTypeTLS      = 13
	EAPTypeMSCHAPv2 = 26
)

type EAPHeader struct {
	Code       uint8
	Identifier uint8
	Length     uint16
}

// NewEAPSuccess creates a new EAP-Success packet.
func NewEAPSuccess(identifier uint8) *EAPHeader {
	return &EAPHeader{
		Code:       EAPCodeSuccess,
		Identifier: identifier,
		Length:     4, // EAP header is always 4 bytes for Success/Failure
	}
}

// NewEAPFailure creates a new EAP-Failure packet.
func NewEAPFailure(identifier uint8) *EAPHeader {
	return &EAPHeader{
		Code:       EAPCodeFailure,
		Identifier: identifier,
		Length:     4, // EAP header is always 4 bytes for Success/Failure
	}
}

// Serialize serializes the EAP-Success or EAP-Failure packet to bytes.
func (eap *EAPHeader) Serialize() []byte {
	buffer := bytes.NewBuffer(nil)

	// Write EAP header
	binary.Write(buffer, binary.BigEndian, eap)

	return buffer.Bytes()
}

type EAPMessage struct {
	EAPHeader
	Type uint8
	Data []byte
}

// Encode 编码 EAP 消息为字节切片
func (msg *EAPMessage) Encode() []byte {
	// 初始化 EAP 消息的基础长度
	length := 5 // Code, Identifier, Length, Type 字段总共占用 5 字节
	var data []byte
	if msg.Data != nil {
		data = msg.Data
		length += len(data)
	}
	buffer := make([]byte, length)
	buffer[0] = msg.Code
	buffer[1] = msg.Identifier
	binary.BigEndian.PutUint16(buffer[2:4], uint16(length))
	buffer[4] = msg.Type

	if len(data) > 0 {
		copy(buffer[5:], data)
	}
	return buffer
}

// String()
// Returns a string representation of the EAP message.
func (msg *EAPMessage) String() string {
	buff := strings.Builder{}
	buff.WriteString("EAPMessage{")
	buff.WriteString("Code=")
	buff.WriteString(strconv.FormatUint(uint64(msg.Code), 10))
	buff.WriteString(", Identifier=")
	buff.WriteString(strconv.FormatUint(uint64(msg.Identifier), 10))
	buff.WriteString(", Length=")
	buff.WriteString(strconv.FormatUint(uint64(msg.Length), 10))
	buff.WriteString(", Type=")
	buff.WriteString(strconv.FormatUint(uint64(msg.Type), 10))
	buff.WriteString(", Data=")
	buff.WriteString(fmt.Sprintf("%x", msg.Data))
	buff.WriteString("}")
	return buff.String()
}

func parseEAPMessage(packet *radius.Packet) (*EAPMessage, error) {
	// 从RADIUS请求中获取EAP-Message属性
	attr, err := rfc2869.EAPMessage_Lookup(packet)
	if err != nil {
		return nil, err
	}

	// 解析EAP消息
	eap := &EAPMessage{
		EAPHeader: EAPHeader{
			Code:       attr[0],
			Identifier: attr[1],
			Length:     binary.BigEndian.Uint16(attr[2:4]),
		},
		Type: attr[4],
		Data: attr[5:],
	}
	return eap, nil
}

// GenerateRandomBytes 生成一个指定长度的随机字节数组
func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// 注意这里返回的 n 是读取的字节数
	if err != nil {
		return nil, err
	}
	return b, nil
}

func generateMessageAuthenticator(r *radius.Packet, secret string) []byte {
	// 创建一个新的MD5哈希
	b, _ := r.MarshalBinary()
	// 创建一个新的MD5哈希
	mac := hmac.New(md5.New, []byte(secret))
	// 写入RADIUS包
	mac.Write(b)
	// 计算Message-Authenticator属性的值
	authenticator := mac.Sum(nil)
	return authenticator
}

func (s *AuthService) sendEapMD5ChallengeRequest(w radius.ResponseWriter, r *radius.Request, secret string) error {
	// 创建一个新的RADIUS响应
	var resp = r.Response(radius.CodeAccessChallenge)

	eapChallenge, err := generateRandomBytes(16)
	if err != nil {
		return err
	}

	state := common.UUID()
	s.AddEapState(state, rfc2865.UserName_GetString(r.Packet), eapChallenge, EapMd5Method)

	rfc2865.State_SetString(resp, state)

	// 创建EAP-Request/MD5-Challenge消息
	eapMessage := []byte{0x01, r.Identifier, 0x00, 0x16, 0x04, 0x10}
	eapMessage = append(eapMessage, eapChallenge...)

	// 设置EAP-Message属性
	rfc2869.EAPMessage_Set(resp, eapMessage)
	rfc2869.MessageAuthenticator_Set(resp, make([]byte, 16))

	authenticator := generateMessageAuthenticator(resp, secret)
	// 设置Message-Authenticator属性
	rfc2869.MessageAuthenticator_Set(resp, authenticator)

	// debug message
	if app.GConfig().Radiusd.Debug {
		log.Info(FmtResponse(resp, r.RemoteAddr))
	}

	// 发送RADIUS响应
	return w.Write(resp)
}

func (s *AuthService) verifyEapMD5Response(eapid uint8, password string, challenge, response []byte) bool {
	hash := md5.New()
	hash.Write([]byte{eapid})
	hash.Write([]byte(password))
	hash.Write(challenge)
	expectedResponse := hash.Sum(nil)
	return bytes.Equal(expectedResponse, response[1:])
}
