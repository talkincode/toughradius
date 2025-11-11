package totp

import (
	"fmt"
	"testing"
)

// Enable two-factor authentication
func _initAuth(user string) (secret, code string) {

	ng := NewGoogleAuth()
	// Secret
	secret = ng.GetSecret()
	fmt.Println("Secret:", secret)

	// Dynamic code(Every30sdynamically generate a6digit number)
	code, err := ng.GetCode(secret)
	fmt.Println("Code:", code, err)

	// Username
	qrCode := ng.GetQrcode(user, code, "ToughDemo")
	fmt.Println("Qrcode", qrCode)

	// Print QR code URL
	qrCodeUrl := ng.GetQrcodeUrl(user, secret, "ToughDemo")
	fmt.Println("QrcodeUrl", qrCodeUrl)

	return
}

func TestOTP(t *testing.T) {
	// fmt.Println("-----------------Enable two-factor authentication----------------------")
	user := "testxxx@qq.com"
	secret, code := _initAuth(user)
	fmt.Println(secret, code)

	fmt.Println("-----------------Info validation----------------------")

	// secretBest to persist in
	// Validate,Dynamic code(Get from Google Authenticator orfreeotpget)
	bool, err := NewGoogleAuth().VerifyCode(secret, code)
	if bool {
		fmt.Println("√")
	} else {
		fmt.Println("X", err)
	}
}
