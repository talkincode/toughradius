package radius

import (
	"context"
	"fmt"

	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2866"
)

func Disconnect(r *radius.Request, secret string, port int, username, nasrip string) error {
	packet := radius.New(radius.CodeDisconnectRequest, []byte(secret))
	sessionid := rfc2866.AcctSessionID_GetString(r.Packet)
	if sessionid == "" {
		return fmt.Errorf("sessionid is empty")
	}
	_ = rfc2865.UserName_SetString(packet, username)
	_ = rfc2866.AcctSessionID_Set(packet, []byte(sessionid))
	response, err := radius.Exchange(context.Background(), packet, fmt.Sprintf("%s:%d", nasrip, port))
	if err != nil {
		return err
	}
	if response.Code != radius.CodeDisconnectACK {
		return fmt.Errorf("disconnect failed, response code: %d", response.Code)
	}
	return nil
}
