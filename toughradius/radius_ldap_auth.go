package toughradius

import (
	"crypto/tls"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-ldap/ldap/v3"
	"github.com/talkincode/toughradius/v8/app"
	"github.com/talkincode/toughradius/v8/common"
	"github.com/talkincode/toughradius/v8/common/timeutil"
	"github.com/talkincode/toughradius/v8/models"
	"github.com/talkincode/toughradius/v8/toughradius/vendors/microsoft"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

type LdapRadisProfile struct {
	Status          string
	MfaSecret       string
	MfaStatus       string
	Domain          string
	AddrPool        string
	MacAddr         string
	IpAddr          string
	ActiveNum       int
	LimitPolicy     string
	UpLimitPolicy   string
	DownLimitPolicy string
	UpRate          int
	DownRate        int
	ExpireTime      time.Time
}

func (s *AuthService) LdapUserAuth(rw radius.ResponseWriter, r *radius.Request,
	username string, ldapNode *models.NetLdapServer, radAccept *radius.Packet, vreq *VendorRequest) (*LdapRadisProfile, error) {
	ignoreChk := s.GetStringConfig(app.ConfigRadiusIgnorePwd, common.DISABLED) == common.DISABLED

	var checkType = "pap"
	// mschapv2
	challenge := microsoft.MSCHAPChallenge_Get(r.Packet)
	if challenge != nil {
		checkType = "mschapv2"
	}

	// chap
	chapPassword := rfc2865.CHAPPassword_Get(r.Packet)
	if chapPassword != nil {
		checkType = "chap"
	}

	// connect ldap
	ld, err := ldap.Dial("tcp", ldapNode.Address)
	if err != nil {
		return nil, NewAuthError(app.MetricsRadiusRejectLdapError, "username ldap auth error, ldap connect error "+err.Error())
	}
	defer ld.Close()

	// start tls
	if ldapNode.Istls == common.ENABLED {
		err = ld.StartTLS(&tls.Config{InsecureSkipVerify: true})
		if err != nil {
			return nil, NewAuthError(app.MetricsRadiusRejectLdapError, "username ldap auth error, ldap tls error "+err.Error())
		}
	}

	// ldapPwd, err := aes.DecryptFromB64(ldapNode.Password, constant.AesKey())
	// if err != nil {
	// 	return nil, fmt.Errorf("username:%s ldap auth error, ldap:%s password format error", username, ldapNode.Name)
	// }

	err = ld.Bind(ldapNode.Basedn, ldapNode.Password)
	if err != nil {
		return nil, NewAuthError(app.MetricsRadiusRejectLdapError, "username ldap auth error, ldap bind auth error "+err.Error())
	}

	searchRequest := ldap.NewSearchRequest(
		ldapNode.Searchdn,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf(ldapNode.UserFilter, username),
		[]string{"dn", "radiusReplyItem", "radiusCallingStationId"},
		nil,
	)

	sr, err := ld.Search(searchRequest)
	if err != nil {
		return nil, NewAuthError(app.MetricsRadiusRejectLdapError, "username ldap auth error, ldap search error "+err.Error())
	}

	if len(sr.Entries) == 0 && !ignoreChk {
		return nil, NewAuthError(app.MetricsRadiusRejectNotExists, "username ldap auth error, user not exists")
	}

	// parse ldap radius attr
	var userProfile = new(LdapRadisProfile)
	userProfile.ExpireTime = time.Now().Add(time.Hour * 24)
	userProfile.parseLdapRadiusAttrs(sr.Entries[0].GetAttributeValues("radiusReplyItem"))
	userProfile.MacAddr = sr.Entries[0].GetAttributeValue("radiusCallingStationId")

	// check status
	if userProfile.Status == common.DISABLED {
		return nil, NewAuthError(app.MetricsRadiusRejectDisable, "ldap user is disabled")
	}

	// check expire
	if userProfile.ExpireTime.Before(time.Now()) {
		return nil, NewAuthError(app.MetricsRadiusRejectExpire, "user Ldap is expire")
	}

	// mac auth check
	if vreq.MacAddr == username {
		return userProfile, nil
	}

	// 如果是 PAP 验证， 直接校验 Ldap 密码
	if !ignoreChk && checkType == "pap" {
		password := rfc2865.UserPassword_GetString(r.Packet)
		userdn := sr.Entries[0].DN
		err = ld.Bind(userdn, password)
		if err != nil {
			return nil, NewAuthError(app.MetricsRadiusRejectPasswdError, "username ldap auth error, user password check error")
		}
	}

	if !ignoreChk && checkType == "chap" {
		return nil, NewAuthError(app.MetricsRadiusRejectPasswdError, "user Ldap chap password is not support")
	}

	// check online
	err = s.CheckOnlineCount(username, userProfile.ActiveNum)
	if err != nil {
		return nil, err
	}

	return userProfile, nil
}

func (p *LdapRadisProfile) parseLdapRadiusAttrs(values []string) {
	for _, value := range values {
		kv := strings.Split(value, "=")
		if len(kv) != 2 {
			continue
		}
		switch strings.TrimSpace(kv[0]) {
		case "Status":
			p.Status = strings.TrimSpace(kv[1])
		case "MfaSecret":
			p.MfaSecret = strings.TrimSpace(kv[1])
		case "MfaStatus":
			p.MfaStatus = strings.TrimSpace(kv[1])
		case "Domain":
			p.Domain = strings.TrimSpace(kv[1])
		case "AddrPool":
			p.AddrPool = strings.TrimSpace(kv[1])
		case "IpAddr":
			p.IpAddr = strings.TrimSpace(kv[1])
		case "LimitPolicy":
			p.LimitPolicy = strings.TrimSpace(kv[1])
		case "UpLimitPolicy":
			p.UpLimitPolicy = strings.TrimSpace(kv[1])
		case "DownLimitPolicy":
			p.DownLimitPolicy = strings.TrimSpace(kv[1])
		case "ActiveNum":
			_ActiveNum, err := strconv.ParseInt(kv[1], 10, 64)
			if err == nil {
				p.ActiveNum = int(_ActiveNum)
			}
		case "UpRate":
			_UpRate, err := strconv.ParseInt(kv[1], 10, 64)
			if err == nil {
				p.UpRate = int(_UpRate)
			}
		case "DownRate":
			_DownRate, err := strconv.ParseInt(kv[1], 10, 64)
			if err == nil {
				p.DownRate = int(_DownRate)
			}
		case "ExpireTime":
			if kv[1] != "" {
				_ExpireTime, err := time.Parse(timeutil.YYYYMMDD_LAYOUT, kv[1])
				if err == nil {
					p.ExpireTime = _ExpireTime
				}
			}
		}
	}
}
