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

	EAPCodeRequest = 1
    EAPCodeResponse = 2
    EAPCodeSuccess = 3
    EAPCodeFailure = 4

    EAPTypeIdentity = 1
    EAPTypeNotification = 2
    EAPTypeNak = 3 // Response only
    EAPTypeMD5Challenge = 4
    EAPTypeOTP = 5 // One-Time Password
    EAPTypeGTC = 6 // Generic Token Card
    // 7-9 Reserved
    EAPTypeTLS = 13
)

type EAPMessage struct {
	Code       uint8
	Identifier uint8
	Length     uint16
	Type       uint8
	Data       []byte
}

func (msg *EAPMessage) Encode() []byte {
    length := 5 // Code, Identifier, Length, Type 字段总共占用 5 字节
    if msg.Data != nil {
        length += len(msg.Data)
    }
    buffer := make([]byte, length)
    buffer[0] = msg.Code
    buffer[1] = msg.Identifier
    binary.BigEndian.PutUint16(buffer[2:4], uint16(length)) // Length 字段是 16 位的，所以我们使用 binary.BigEndian.PutUint16 来写入它
    buffer[4] = msg.Type
    if msg.Data != nil {
        copy(buffer[5:], msg.Data)
    }
    return buffer
}

// String()
// Returns a string representation of the EAP message.
func (e *EAPMessage) String() string {
	buff := strings.Builder{}
	buff.WriteString("EAPMessage{")
	buff.WriteString("Code=")
	buff.WriteString(strconv.FormatUint(uint64(e.Code), 10))
	buff.WriteString(", Identifier=")
	buff.WriteString(strconv.FormatUint(uint64(e.Identifier), 10))
	buff.WriteString(", Length=")
	buff.WriteString(strconv.FormatUint(uint64(e.Length), 10))
	buff.WriteString(", Type=")
	buff.WriteString(strconv.FormatUint(uint64(e.Type), 10))
	buff.WriteString(", Data=")
	buff.WriteString(fmt.Sprintf("%x", e.Data))
	buff.WriteString("}")
	return buff.String()
}


func parseEAPMessage(r *radius.Request) (*EAPMessage, error) {
	// 从RADIUS请求中获取EAP-Message属性
	attr, err := rfc2869.EAPMessage_Lookup(r.Packet)
	if err != nil {
		return nil, err
	}

	// 解析EAP消息
	eap := &EAPMessage{
		Code:       attr[0],
		Identifier: attr[1],
		Length:     binary.BigEndian.Uint16(attr[2:4]),
		Type:       attr[4],
		Data:       attr[5:],
	}

	return eap, nil
}

const ChallengeLength = 16

func createMD5Challenge() ([]byte, error) {
	challenge := make([]byte, ChallengeLength)
	_, err := rand.Read(challenge)
	if err != nil {
		return nil, err
	}
	return challenge, nil
}

func (s *AuthService) sendEAPRequest(w radius.ResponseWriter, r *radius.Request,  secret string) error {
	// 创建一个新的RADIUS响应
	var resp = r.Response(radius.CodeAccessChallenge)

	eapChallenge, err := createMD5Challenge()
	if err != nil {
		return err
	}

	state := common.UUID()
	s.AddEapState(state, rfc2865.UserName_GetString(r.Packet), eapChallenge)

	rfc2865.State_SetString(resp, state)

	// 创建EAP-Request/MD5-Challenge消息
	eapMessage := []byte{0x01, r.Identifier, 0x00, 0x16, 0x04, 0x10}
	eapMessage = append(eapMessage, eapChallenge...)

	// 设置EAP-Message属性
	rfc2869.EAPMessage_Set(resp, eapMessage)
	rfc2869.MessageAuthenticator_Set(resp, make([]byte, 16))

	authenticator := genMessageAuthenticator(resp, secret)
	// 设置Message-Authenticator属性
	rfc2869.MessageAuthenticator_Set(resp, authenticator)

	// debug message
	if app.GConfig().Radiusd.Debug {
		log.Info(FmtResponse(resp, r.RemoteAddr))
	}

	// 发送RADIUS响应
	return w.Write(resp)
}

func genMessageAuthenticator(r *radius.Packet, secret string) []byte {
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


func (s *AuthService)  verifyMD5Response(eapid uint8, password string, challenge, response []byte) bool {
    hash := md5.New()
	hash.Write([]byte{eapid})
    hash.Write([]byte(password))
    hash.Write(challenge)
    expectedResponse := hash.Sum(nil)
	return bytes.Equal(expectedResponse, response[1:])
}