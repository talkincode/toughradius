package eap

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"encoding/binary"

	"go.uber.org/zap"
	"layeh.com/radius"
	"layeh.com/radius/rfc2869"
)

// ParseEAPMessage 从 RADIUS 包中解析 EAP 消息
func ParseEAPMessage(packet *radius.Packet) (*EAPMessage, error) {
	attr, err := rfc2869.EAPMessage_Lookup(packet)
	if err != nil {
		return nil, err
	}

	if len(attr) < 5 {
		return nil, ErrInvalidEAPMessage
	}

	eap := &EAPMessage{
		Code:       attr[0],
		Identifier: attr[1],
		Length:     binary.BigEndian.Uint16(attr[2:4]),
		Type:       attr[4],
		Data:       attr[5:],
	}
	return eap, nil
}

// EncodeEAPMessage 编码 EAP 消息为字节切片
func (msg *EAPMessage) Encode() []byte {
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

// EncodeEAPHeader 编码 EAP Header (Success/Failure)
func EncodeEAPHeader(code, identifier uint8) []byte {
	buffer := make([]byte, 4)
	buffer[0] = code
	buffer[1] = identifier
	binary.BigEndian.PutUint16(buffer[2:4], 4) // Length = 4 for Success/Failure
	return buffer
}

// GenerateRandomBytes 生成指定长度的随机字节数组
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		zap.L().Error("Failed to generate random bytes", zap.Error(err))
		return nil, err
	}
	return b, nil
}

// GenerateMessageAuthenticator 生成 Message-Authenticator 属性
func GenerateMessageAuthenticator(packet *radius.Packet, secret string) []byte {
	b, _ := packet.MarshalBinary()
	mac := hmac.New(md5.New, []byte(secret))
	mac.Write(b)
	return mac.Sum(nil)
}

// SetEAPMessageAndAuth 设置 EAP-Message 和 Message-Authenticator 属性
func SetEAPMessageAndAuth(response *radius.Packet, eapData []byte, secret string) {
	rfc2869.EAPMessage_Set(response, eapData)
	rfc2869.MessageAuthenticator_Set(response, make([]byte, 16))
	authenticator := GenerateMessageAuthenticator(response, secret)
	rfc2869.MessageAuthenticator_Set(response, authenticator)
}

// ComputeMD5Hash 计算 MD5 哈希
func ComputeMD5Hash(data []byte) []byte {
	hash := md5.New()
	hash.Write(data)
	return hash.Sum(nil)
}

// VerifyMD5Hash 验证 MD5 哈希
func VerifyMD5Hash(expected, actual []byte) bool {
	return bytes.Equal(expected, actual)
}
