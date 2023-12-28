package toughradius

import (
	"github.com/talkincode/toughradius/v8/app"
	"github.com/talkincode/toughradius/v8/common"
	"github.com/talkincode/toughradius/v8/common/zaplog/log"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2869"
)

// sendEapMsChapV2Request
// 发送EAP-Request/MS-CHAPv2消息
func (s *AuthService) sendEapMsChapV2Request(w radius.ResponseWriter, r *radius.Request, secret string) error {
	// 创建一个新的RADIUS响应
	var resp = r.Response(radius.CodeAccessChallenge)

	name := "toughradius"
	eapChallenge, err := generateRandomBytes(16)
	if err != nil {
		return err
	}

	state := common.UUID()
	s.AddEapState(state, rfc2865.UserName_GetString(r.Packet), eapChallenge)

	rfc2865.State_SetString(resp, state)

	eapMessage := &EAPMessage{
		Code:       EAPCodeRequest,  // EAP code: Request
		Identifier: r.Identifier,    // EAP Identifier
		Type:       EAPTypeMSCHAPv2, // EAP Type: MS-CHAPv2
		Data: &MsChapV2EAPChallengeData{
			Name:      name,
			Challenge: eapChallenge,
		},
	}
	// 设置EAP-Message属性
	rfc2869.EAPMessage_Set(resp, eapMessage.Encode())
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
