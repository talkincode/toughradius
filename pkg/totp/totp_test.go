package totp

import (
	"fmt"
	"testing"
)

// 开启二次认证
func _initAuth(user string) (secret, code string) {

	ng := NewGoogleAuth()
	// 秘钥
	secret = ng.GetSecret()
	fmt.Println("Secret:", secret)

	// 动态码(每隔30s会动态生成一个6位数的数字)
	code, err := ng.GetCode(secret)
	fmt.Println("Code:", code, err)

	// 用户名
	qrCode := ng.GetQrcode(user, code, "ToughDemo")
	fmt.Println("Qrcode", qrCode)

	// 打印二维码地址
	qrCodeUrl := ng.GetQrcodeUrl(user, secret, "ToughDemo")
	fmt.Println("QrcodeUrl", qrCodeUrl)

	return
}

func TestOTP(t *testing.T) {
	// fmt.Println("-----------------开启二次认证----------------------")
	user := "testxxx@qq.com"
	secret, code := _initAuth(user)
	fmt.Println(secret, code)

	fmt.Println("-----------------信息校验----------------------")

	// secret最好持久化保存在
	// 验证,动态码(从谷歌验证器获取或者freeotp获取)
	bool, err := NewGoogleAuth().VerifyCode(secret, code)
	if bool {
		fmt.Println("√")
	} else {
		fmt.Println("X", err)
	}
}
