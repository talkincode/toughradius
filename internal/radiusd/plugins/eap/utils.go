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

// ParseEAPMessage extracts the EAP message from a RADIUS packet
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

// EncodeEAPMessage encodes an EAP message into bytes
func (msg *EAPMessage) Encode() []byte {
	length := 5 // Code, Identifier, Length, and Type fields occupy 5 bytes
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

// EncodeEAPHeader encodes an EAP header (Success/Failure)
func EncodeEAPHeader(code, identifier uint8) []byte {
	buffer := make([]byte, 4)
	buffer[0] = code
	buffer[1] = identifier
	binary.BigEndian.PutUint16(buffer[2:4], 4) // Length = 4 for Success/Failure
	return buffer
}

// GenerateRandomBytes generates a random byte array of the specified length
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		zap.L().Error("Failed to generate random bytes", zap.Error(err))
		return nil, err
	}
	return b, nil
}

// GenerateMessageAuthenticator generates the Message-Authenticator attribute
func GenerateMessageAuthenticator(packet *radius.Packet, secret string) []byte {
	b, _ := packet.MarshalBinary()
	mac := hmac.New(md5.New, []byte(secret))
	mac.Write(b)
	return mac.Sum(nil)
}

// SetEAPMessageAndAuth sets the EAP-Message and Message-Authenticator attributes
func SetEAPMessageAndAuth(response *radius.Packet, eapData []byte, secret string) {
	_ = rfc2869.EAPMessage_Set(response, eapData)                    //nolint:errcheck
	_ = rfc2869.MessageAuthenticator_Set(response, make([]byte, 16)) //nolint:errcheck
	authenticator := GenerateMessageAuthenticator(response, secret)
	_ = rfc2869.MessageAuthenticator_Set(response, authenticator) //nolint:errcheck
}

// ComputeMD5Hash computes the MD5 hash
func ComputeMD5Hash(data []byte) []byte {
	hash := md5.New()
	hash.Write(data)
	return hash.Sum(nil)
}

// VerifyMD5Hash validates the MD5 hash
func VerifyMD5Hash(expected, actual []byte) bool {
	return bytes.Equal(expected, actual)
}
