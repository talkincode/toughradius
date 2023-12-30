package toughradius

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"github.com/talkincode/toughradius/v8/app"
	"github.com/talkincode/toughradius/v8/common"
	"github.com/talkincode/toughradius/v8/common/zaplog/log"
	"golang.org/x/exp/rand"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2869"
)

const (
	MSCHAPv2Challenge   = 1
	MSCHAPv2Response    = 2
	MSCHAPChallengeSize = 16
	MSCHAPv2Success     = 3
	MSCHAPv2Failure     = 4
)

// EAPMSCHAPv2Challenge represents an EAP-MSCHAPv2 Challenge message.
type EAPMSCHAPv2Challenge struct {
	EAPHeader
	Type           uint8
	OpCode         uint8
	MsIdentifier   uint8
	MsChapV2Length uint16
	ValueSize      uint8
	Challenge      [MSCHAPChallengeSize]byte
	Name           []byte
}

// NewEAPMSCHAPv2Challenge creates a new EAP-MSCHAPv2 Challenge packet with a random challenge.
func NewEAPMSCHAPv2Challenge(identifier uint8, name string) *EAPMSCHAPv2Challenge {
	var challenge [MSCHAPChallengeSize]byte
	rand.Seed(uint64(time.Now().UnixNano()))
	_, _ = rand.Read(challenge[:])

	eap := &EAPMSCHAPv2Challenge{
		EAPHeader: EAPHeader{
			Code:       EAPCodeRequest,
			Identifier: identifier,
			Length:     0, // Will be set later
		},
		Type:           EAPTypeMSCHAPv2,
		OpCode:         MSCHAPv2Challenge,
		MsIdentifier:   identifier, // Assuming the same as EAP Identifier
		MsChapV2Length: 0,          // Will be set later
		ValueSize:      MSCHAPChallengeSize,
		Challenge:      challenge,
		Name:           []byte(name),
	}

	// Calculate the MS-CHAPv2 Length
	msChapV2Length := uint16(1 + 1 + 2 + 1 + MSCHAPChallengeSize + len(eap.Name)) // OpCode (1 byte) + MsIdentifier (1 byte) + MsChapV2Length (2 bytes) + ValueSize (1 byte) + Challenge + Name
	eap.MsChapV2Length = msChapV2Length

	// Calculate the EAP Length
	eap.Length = uint16(4 + msChapV2Length) // EAP header (4 bytes) + MS-CHAPv2 data
	return eap
}

// Serialize serializes the EAP-MSCHAPv2 Challenge packet to bytes.
func (eap *EAPMSCHAPv2Challenge) Serialize() []byte {
	buffer := bytes.NewBuffer(nil)

	// Write EAP header
	_ = binary.Write(buffer, binary.BigEndian, eap.EAPHeader)

	// Write Type, OpCode, MsIdentifier, and MsChapV2Length
	buffer.WriteByte(eap.Type)
	buffer.WriteByte(eap.OpCode)
	buffer.WriteByte(eap.MsIdentifier)
	_ = binary.Write(buffer, binary.BigEndian, eap.MsChapV2Length)

	// Write ValueSize and Challenge
	buffer.WriteByte(eap.ValueSize)
	buffer.Write(eap.Challenge[:])

	// Write Name
	buffer.Write(eap.Name)

	return buffer.Bytes()
}

// EAPMSCHAPv2Response represents an EAP-MSCHAPv2 Response message.
type EAPMSCHAPv2Response struct {
	EAPHeader
	Type           uint8
	OpCode         uint8
	MsIdentifier   uint8
	MsChapV2Length uint16
	ValueSize      uint8
	PeerChallenge  [16]byte
	Reserved       [8]byte
	Response       [24]byte
	Flags          uint8
	Name           []byte
}

// NewEAPMSCHAPv2Response creates a new EAP-MSCHAPv2 Response packet.
func NewEAPMSCHAPv2Response(identifier uint8, peerChallenge, response []byte, name string) *EAPMSCHAPv2Response {
	eap := &EAPMSCHAPv2Response{
		EAPHeader: EAPHeader{
			Code:       EAPCodeResponse,
			Identifier: identifier,
			Length:     0, // Will be set later
		},
		Type:      EAPTypeMSCHAPv2,
		OpCode:    MSCHAPv2Challenge,
		ValueSize: 49, // PeerChallenge (16 bytes) + Reserved (8 bytes) + Response (24 bytes) + Flags (1 byte)
		Flags:     0,  // Typically 0 unless certain conditions are met
		Name:      []byte(name),
	}

	copy(eap.PeerChallenge[:], peerChallenge)
	copy(eap.Response[:], response)

	eap.Length = uint16(5 + 49 + len(eap.Name)) // EAP header (4 bytes) + Type (1 byte) + MSCHAPv2 Response fields + Name
	return eap
}

// Serialize serializes the EAP-MSCHAPv2 Response packet to bytes.
func (eap *EAPMSCHAPv2Response) Serialize() []byte {
	buffer := bytes.NewBuffer(nil)

	// Write EAP header
	binary.Write(buffer, binary.BigEndian, eap.EAPHeader)

	// Write Type, OpCode, and ValueSize
	buffer.WriteByte(eap.Type)
	buffer.WriteByte(eap.OpCode)
	buffer.WriteByte(eap.ValueSize)

	// Write PeerChallenge, Reserved, Response, and Flags
	buffer.Write(eap.PeerChallenge[:])
	buffer.Write(eap.Reserved[:])
	buffer.Write(eap.Response[:])
	buffer.WriteByte(eap.Flags)

	// Write Name
	buffer.Write(eap.Name)

	return buffer.Bytes()
}
func ParseEAPMSCHAPv2Response(packet *radius.Packet) (*EAPMSCHAPv2Response, error) {
	// Get the EAP-Message attribute from the RADIUS request
	attr, err := rfc2869.EAPMessage_Lookup(packet)
	if err != nil {
		return nil, err
	}

	var eap EAPMSCHAPv2Response
	buffer := bytes.NewBuffer(attr)

	// 读取 EAP 头部
	if err := binary.Read(buffer, binary.BigEndian, &eap.EAPHeader.Code); err != nil {
		return nil, fmt.Errorf("读取 EAP Code 失败: %w", err)
	}
	if err := binary.Read(buffer, binary.BigEndian, &eap.EAPHeader.Identifier); err != nil {
		return nil, fmt.Errorf("读取 EAP Identifier 失败: %w", err)
	}
	if err := binary.Read(buffer, binary.BigEndian, &eap.EAPHeader.Length); err != nil {
		return nil, fmt.Errorf("读取 EAP Length 失败: %w", err)
	}

	// Read Type, OpCode, and ValueSize
	eap.Type, _ = buffer.ReadByte()
	eap.OpCode, _ = buffer.ReadByte()
	eap.MsIdentifier, _ = buffer.ReadByte()

	if err := binary.Read(buffer, binary.BigEndian, &eap.MsChapV2Length); err != nil {
		return nil, fmt.Errorf("读取 EAP MsChapV2Length 失败: %w", err)
	}

	// 检查缓冲区是否有足够的字节来读取 MsChapV2Length
	eap.ValueSize, _ = buffer.ReadByte()

	// 检查缓冲区是否有足够的字节来读取 PeerChallenge, Reserved, Response 和 Flags
	if buffer.Len() < int(eap.ValueSize) {
		return nil, fmt.Errorf("缓冲区太短，无法读取 PeerChallenge, Reserved, Response 和 Flags")
	}

	// 读取 PeerChallenge, Reserved, Response 和 Flags
	copy(eap.PeerChallenge[:], buffer.Next(16))
	copy(eap.Reserved[:], buffer.Next(8))
	copy(eap.Response[:], buffer.Next(24))
	eap.Flags, _ = buffer.ReadByte()

	// 读取剩余的所有字节作为 Name
	eap.Name = buffer.Bytes()

	return &eap, nil
}

// EAPMSCHAPv2SuccessFailure represents an EAP-MSCHAPv2 Success or Failure message.
type EAPMSCHAPv2SuccessFailure struct {
	EAPHeader
	Type    uint8
	OpCode  uint8
	Message string
}

// NewEAPMSCHAPv2SuccessFailure creates a new EAP-MSCHAPv2 Success or Failure packet.
func NewEAPMSCHAPv2SuccessFailure(code uint8, identifier uint8, opCode uint8, message string) *EAPMSCHAPv2SuccessFailure {
	eap := &EAPMSCHAPv2SuccessFailure{
		EAPHeader: EAPHeader{
			Code:       code,
			Identifier: identifier,
			Length:     0, // Will be set later
		},
		Type:    EAPTypeMSCHAPv2,
		OpCode:  opCode,
		Message: message,
	}

	eap.Length = uint16(5 + len(eap.Message)) // EAP header (4 bytes) + Type (1 byte) + Message
	return eap
}

// Serialize serializes the EAP-MSCHAPv2 Success or Failure packet to bytes.
func (eap *EAPMSCHAPv2SuccessFailure) Serialize() []byte {
	buffer := bytes.NewBuffer(nil)

	// Write EAP header
	_ = binary.Write(buffer, binary.BigEndian, eap.EAPHeader)

	// Write Type and OpCode
	buffer.WriteByte(eap.Type)
	buffer.WriteByte(eap.OpCode)

	// Write Message
	buffer.WriteString(eap.Message)

	return buffer.Bytes()
}

func parseEAPMSCHAPv2SuccessFailure(packet *radius.Packet) (*EAPMSCHAPv2SuccessFailure, error) {
	// 从RADIUS请求中获取EAP-Message属性
	attr, err := rfc2869.EAPMessage_Lookup(packet)
	if err != nil {
		return nil, err
	}

	var eap EAPMSCHAPv2SuccessFailure
	buffer := bytes.NewBuffer(attr)

	// Read EAP header
	if err := binary.Read(buffer, binary.BigEndian, &eap.EAPHeader); err != nil {
		return nil, err
	}

	// Read Type and OpCode
	eap.Type, _ = buffer.ReadByte()
	eap.OpCode, _ = buffer.ReadByte()

	// Read Message
	eap.Message = string(buffer.Bytes())

	return &eap, nil
}

// parseEAPMSCHAPv2OpCode extracts the OpCode from an EAP-MSCHAPv2 message within a RADIUS packet.
func parseEAPMSCHAPv2OpCode(packet *radius.Packet) (uint8, error) {
	// Retrieve the EAP-Message attribute from the RADIUS packet
	attr, err := rfc2869.EAPMessage_Lookup(packet)
	if err != nil {
		return 0, err
	}

	// Ensure that the EAP-Message attribute is present
	if attr == nil {
		return 0, errors.New("EAP-Message attribute is missing")
	}

	// The EAP-Message attribute is a slice of bytes, so we need to check its length
	if len(attr) < 6 {
		return 0, errors.New("EAP-Message attribute is too short to contain an OpCode")
	}

	// OpCode is the fifth byte in the EAP-MSCHAPv2 message
	opCode := attr[5]

	return opCode, nil
}

// sendEapMsChapV2Request
// 发送EAP-Request/MS-CHAPv2消息
func (s *AuthService) sendEapMsChapV2Request(w radius.ResponseWriter, r *radius.Request, secret string) error {
	// 创建一个新的RADIUS响应
	var resp = r.Response(radius.CodeAccessChallenge)

	name := "toughradius"
	eapMessage := NewEAPMSCHAPv2Challenge(r.Identifier, name)

	state := common.UUID()
	s.AddEapState(state, rfc2865.UserName_GetString(r.Packet), eapMessage.Challenge[:], EapMschapv2Method)

	rfc2865.State_SetString(resp, state)

	// 设置EAP-Message属性
	_ = rfc2869.EAPMessage_Set(resp, eapMessage.Serialize())
	_ = rfc2869.MessageAuthenticator_Set(resp, make([]byte, 16))

	authenticator := generateMessageAuthenticator(resp, secret)
	// 设置Message-Authenticator属性
	_ = rfc2869.MessageAuthenticator_Set(resp, authenticator)

	// debug message
	if app.GConfig().Radiusd.Debug {
		log.Info(FmtResponse(resp, r.RemoteAddr))
	}

	// 发送RADIUS响应
	return w.Write(resp)
}
