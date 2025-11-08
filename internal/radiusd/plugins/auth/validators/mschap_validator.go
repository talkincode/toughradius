package validators

import (
	"bytes"
	"context"

	"github.com/talkincode/toughradius/v9/internal/radiusd/errors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors/microsoft"
	"layeh.com/radius/rfc2759"
	"layeh.com/radius/rfc3079"
)

// MSCHAPValidator MSCHAP 密码验证器（非EAP模式）
type MSCHAPValidator struct{}

func (v *MSCHAPValidator) Name() string {
	return "mschap"
}

func (v *MSCHAPValidator) CanHandle(authCtx *auth.AuthContext) bool {
	challenge := microsoft.MSCHAPChallenge_Get(authCtx.Request.Packet)
	response := microsoft.MSCHAP2Response_Get(authCtx.Request.Packet)
	return challenge != nil && response != nil
}

func (v *MSCHAPValidator) Validate(ctx context.Context, authCtx *auth.AuthContext, password string) error {
	challenge := microsoft.MSCHAPChallenge_Get(authCtx.Request.Packet)
	response := microsoft.MSCHAP2Response_Get(authCtx.Request.Packet)

	if len(challenge) != 16 || len(response) != 50 {
		return errors.NewAuthError("radus_reject_passwd_error",
			"user mschap challenge len or response len error")
	}

	ident := response[0]
	peerChallenge := response[2:18]
	peerResponse := response[26:50]

	return v.validateMSCHAPv2(authCtx, password, challenge, ident, peerChallenge, peerResponse)
}

func (v *MSCHAPValidator) validateMSCHAPv2(
	authCtx *auth.AuthContext,
	password string,
	challenge []byte,
	ident byte,
	peerChallenge,
	peerResponse []byte,
) error {
	username := authCtx.User.Username
	byteUser := []byte(username)
	bytePwd := []byte(password)

	ntResponse, err := rfc2759.GenerateNTResponse(challenge, peerChallenge, byteUser, bytePwd)
	if err != nil {
		return errors.NewAuthError("radus_reject_passwd_error",
			"user mschap cannot generate ntResponse")
	}

	if !bytes.Equal(ntResponse, peerResponse) {
		return errors.NewPasswordMismatchError()
	}

	// 生成加密密钥
	recvKey, err := rfc3079.MakeKey(ntResponse, bytePwd, false)
	if err != nil {
		return errors.NewAuthError("radus_reject_passwd_error",
			"user mschap cannot make recvKey")
	}

	sendKey, err := rfc3079.MakeKey(ntResponse, bytePwd, true)
	if err != nil {
		return errors.NewAuthError("radus_reject_passwd_error",
			"user mschap cannot make sendKey")
	}

	authenticatorResponse, err := rfc2759.GenerateAuthenticatorResponse(challenge, peerChallenge, ntResponse, byteUser, bytePwd)
	if err != nil {
		return errors.NewAuthError("radus_reject_passwd_error",
			"user mschap cannot generate authenticator response")
	}

	success := make([]byte, 43)
	success[0] = ident
	copy(success[1:], authenticatorResponse)

	// 添加响应属性
	microsoft.MSCHAP2Success_Add(authCtx.Response, []byte(success))
	microsoft.MSMPPERecvKey_Add(authCtx.Response, recvKey)
	microsoft.MSMPPESendKey_Add(authCtx.Response, sendKey)
	microsoft.MSMPPEEncryptionPolicy_Add(authCtx.Response, microsoft.MSMPPEEncryptionPolicy_Value_EncryptionAllowed)
	microsoft.MSMPPEEncryptionTypes_Add(authCtx.Response, microsoft.MSMPPEEncryptionTypes_Value_RC440or128BitAllowed)

	return nil
}
