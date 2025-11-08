package radiusd

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"strings"

	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors/microsoft"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"layeh.com/radius"
	"layeh.com/radius/rfc2759"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc3079"
)

func (s *AuthService) GetLocalPassword(user *domain.RadiusUser, isMacAuth bool) (string, error) {
	if isMacAuth {
		return user.MacAddr, nil
	}
	return user.Password, nil
}

// CheckPassword
// passward 不为空为 PAP 认证
// chapPassword 不为空为 Chap 认证
func (s *AuthService) CheckPassword(r *radius.Request, username, localpassword string, radAccept *radius.Packet, isMacAuth bool) error {
	ignoreChk := s.GetStringConfig(app.ConfigRadiusIgnorePwd, common.DISABLED) == common.DISABLED
	password := rfc2865.UserPassword_GetString(r.Packet)
	challenge := microsoft.MSCHAPChallenge_Get(r.Packet)
	response := microsoft.MSCHAP2Response_Get(r.Packet)

	if ignoreChk && challenge == nil {
		return nil
	}
	// mschap 认证
	if challenge != nil && response != nil {
		return s.CheckMsChapPassword(username, localpassword, challenge, response, radAccept)
	}

	chapPassword := rfc2865.CHAPPassword_Get(r.Packet)
	if chapPassword != nil && !ignoreChk && !isMacAuth {
		chapChallenge := rfc2865.CHAPChallenge_Get(r.Packet)
		if len(chapPassword) != 17 {
			return NewAuthError(app.MetricsRadiusRejectPasswdError, "user chap password must be 17 bytes")
		}

		if len(chapChallenge) != 16 {
			return NewAuthError(app.MetricsRadiusRejectPasswdError, fmt.Sprintf("user chap challenge (len=%d) must be 16 bytes", len(chapChallenge)))
		}

		w := md5.New()
		w.Write([]byte{chapPassword[0]})
		w.Write([]byte(localpassword))
		w.Write(chapChallenge)
		md5r := w.Sum(nil)
		for i := 0; i < 16; i++ {
			if md5r[i] != chapPassword[i+1] {
				return NewAuthError(app.MetricsRadiusRejectPasswdError, "user chap password error")
			}
		}

		return nil
	}

	if strings.TrimSpace(password) != "" &&
		!ignoreChk && !isMacAuth &&
		strings.TrimSpace(password) != localpassword {
		return NewAuthError(app.MetricsRadiusRejectPasswdError, "user pap password is not match")
	}

	return nil
}

// CheckMsChapPassword 非 EAP 模式的验证
func (s *AuthService) CheckMsChapPassword(
	username, password string,
	challenge, response []byte,
	radAccept *radius.Packet,
) error {
	if len(challenge) != 16 || len(response) != 50 {
		return NewAuthError(app.MetricsRadiusRejectPasswdError,
			"user mschap access reject challenge len or response len error")
	}
	ident := response[0]
	peerChallenge := response[2:18]
	peerResponse := response[26:50]
	return s.CheckMsChapV2Password(username, password, challenge, ident, peerChallenge, peerResponse, radAccept)
}

// CheckMsChapV2Password EAP 模式的验证
func (s *AuthService) CheckMsChapV2Password(
	username,
	password string,
	challenge []byte,
	ident byte,
	peerChallenge,
	peerResponse []byte,
	radAccept *radius.Packet,
) error {
	byteUser := []byte(username)
	bytePwd := []byte(password)
	ntResponse, err := rfc2759.GenerateNTResponse(challenge, peerChallenge, byteUser, bytePwd)
	if err != nil {
		return NewAuthError(app.MetricsRadiusRejectPasswdError,
			"user mschap access mschap access cannot generate ntResponse")
	}

	if bytes.Equal(ntResponse, peerResponse) {
		recvKey, err := rfc3079.MakeKey(ntResponse, bytePwd, false)
		if err != nil {
			return NewAuthError(app.MetricsRadiusRejectPasswdError,
				"user mschap access cannot make recvKey")
		}

		sendKey, err := rfc3079.MakeKey(ntResponse, bytePwd, true)
		if err != nil {
			return NewAuthError(app.MetricsRadiusRejectPasswdError,
				"user mschap access cannot make sendKey")
		}

		authenticatorResponse, err := rfc2759.GenerateAuthenticatorResponse(challenge, peerChallenge, ntResponse, byteUser, bytePwd)
		if err != nil {
			return NewAuthError(app.MetricsRadiusRejectPasswdError,
				"user mschap access  cannot generate authenticator response")
		}

		success := make([]byte, 43)
		success[0] = ident
		copy(success[1:], authenticatorResponse)

		microsoft.MSCHAP2Success_Add(radAccept, []byte(success))
		microsoft.MSMPPERecvKey_Add(radAccept, recvKey)
		microsoft.MSMPPESendKey_Add(radAccept, sendKey)
		microsoft.MSMPPEEncryptionPolicy_Add(radAccept, microsoft.MSMPPEEncryptionPolicy_Value_EncryptionAllowed)
		microsoft.MSMPPEEncryptionTypes_Add(radAccept, microsoft.MSMPPEEncryptionTypes_Value_RC440or128BitAllowed)
		return nil
	}

	return NewAuthError(app.MetricsRadiusRejectPasswdError,
		"user mschap access reject password error")

}
