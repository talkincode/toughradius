package validators

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"

	"github.com/talkincode/toughradius/v9/internal/radiusd/errors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	"layeh.com/radius/rfc2865"
)

// CHAPValidator CHAP密码验证器
type CHAPValidator struct{}

func (v *CHAPValidator) Name() string {
	return "chap"
}

func (v *CHAPValidator) CanHandle(authCtx *auth.AuthContext) bool {
	chapPassword := rfc2865.CHAPPassword_Get(authCtx.Request.Packet)
	return chapPassword != nil
}

func (v *CHAPValidator) Validate(ctx context.Context, authCtx *auth.AuthContext, password string) error {
	chapPassword := rfc2865.CHAPPassword_Get(authCtx.Request.Packet)
	if len(chapPassword) != 17 {
		return errors.NewAuthError("radus_reject_passwd_error",
			"user chap password must be 17 bytes")
	}

	chapChallenge := rfc2865.CHAPChallenge_Get(authCtx.Request.Packet)
	if len(chapChallenge) != 16 {
		return errors.NewAuthError("radus_reject_passwd_error",
			fmt.Sprintf("user chap challenge (len=%d) must be 16 bytes", len(chapChallenge)))
	}

	w := md5.New()
	w.Write([]byte{chapPassword[0]})
	w.Write([]byte(password))
	w.Write(chapChallenge)
	md5r := w.Sum(nil)

	if !bytes.Equal(md5r, chapPassword[1:17]) {
		return errors.NewPasswordMismatchError()
	}

	return nil
}
