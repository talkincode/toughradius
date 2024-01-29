package toughradius

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"testing"
	"time"

	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2869"
)

func TestEAPOTP(t *testing.T) {
    // RADIUS 服务器配置
    address := "127.0.0.1:1812" // 替换为 RADIUS 服务器地址
    secret := []byte("secret")  // 替换为 RADIUS 共享密钥

    // 创建 RADIUS 客户端
    client := radius.Client{}

    // 创建 RADIUS 认证请求
    request := radius.New(radius.CodeAccessRequest, secret)
	rfc2865.CallingStationID_SetString(request, "10.10.10.10")
	rfc2865.NASIdentifier_Set(request, []byte("tradtest"))
	rfc2865.NASIPAddress_Set(request, net.ParseIP("127.0.0.1"))
	rfc2865.NASPort_Set(request, 0)
	rfc2865.NASPortType_Set(request, 0)
	rfc2869.NASPortID_Set(request, []byte("slot=2;subslot=2;port=22;vlanid=100;"))
	rfc2865.CalledStationID_SetString(request, "11:11:11:11:11:11")
	rfc2865.CallingStationID_SetString(request, "11:11:11:11:11:11")

    // 设置用户名
    username := "test01"
    rfc2865.UserName_SetString(request, username)

    // 构建 EAP-Response/Identity 消息
    eapIdentity := []byte{2, 1} // 2-Response, 1-Id
    eapIdentity = append(eapIdentity, []byte{0x00, 0x00}...) // Length, will be set later
    eapIdentity = append(eapIdentity, 1) // EAP Type = Identity
    eapIdentity = append(eapIdentity, []byte(username)...)

    // Set the length
    binary.BigEndian.PutUint16(eapIdentity[2:4], uint16(len(eapIdentity)))

    rfc2869.EAPMessage_Set(request, eapIdentity)

    // 发送 RADIUS 请求，并接收响应
    fmt.Println(FmtPacket(request))
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()  
    response, err := client.Exchange(ctx, request, address)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    fmt.Println(FmtPacket(response))

    // 检查响应类型
    if response.Code != radius.CodeAccessChallenge {
        fmt.Println("Unexpected response code:", response.Code)
        return
    }

    stateId := rfc2865.State_GetString(response)

    // 处理 EAP-Request/OTP 挑战
    eapChallenge := response.Attributes.Get(rfc2869.EAPMessage_Type)
    if eapChallenge == nil {
        fmt.Println("EAP-Request/OTP challenge not received")
        return
    }

    // 提取 EAP-Request/OTP 挑战中的标识符
    eapID := eapChallenge[1]

    // 构建 EAP-Response/OTP 消息
    otp := "123456" // 模拟的 OTP 值
    eapResponse := []byte{2, eapID} // 2-Response, EapID
    eapResponse = append(eapResponse, []byte{0x00, 0x00}...) // Length, will be set later
    eapResponse = append(eapResponse, 5) // EAP Type = OTP
    eapResponse = append(eapResponse, []byte(otp)...)

    // Set the length
    binary.BigEndian.PutUint16(eapResponse[2:4], uint16(len(eapResponse)))

    // 创建新的 RADIUS 请求以响应 OTP 挑战
    request = radius.New(radius.CodeAccessRequest, secret)
	rfc2865.CallingStationID_SetString(request, "10.10.10.10")
	rfc2865.NASIdentifier_Set(request, []byte("tradtest"))
	rfc2865.NASIPAddress_Set(request, net.ParseIP("127.0.0.1"))
	rfc2865.NASPort_Set(request, 0)
	rfc2865.NASPortType_Set(request, 0)
    rfc2865.State_SetString(request, stateId)
	rfc2869.NASPortID_Set(request, []byte("slot=2;subslot=2;port=22;vlanid=100;"))
	rfc2865.CalledStationID_SetString(request, "11:11:11:11:11:11")
	rfc2865.CallingStationID_SetString(request, "11:11:11:11:11:11")

    rfc2865.UserName_SetString(request, username)
    rfc2869.EAPMessage_Set(request, eapResponse)

    // 再次发送 RADIUS 请求，并接收响应
    fmt.Println(FmtPacket(request))

    ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()  
    response, err = client.Exchange(ctx, request, address)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    fmt.Println(FmtPacket(response))

    // 检查最终的 RADIUS 响应
    if response.Code == radius.CodeAccessAccept {
        fmt.Println("Authentication successful")
    } else if response.Code == radius.CodeAccessReject {
        fmt.Println("Authentication failed")
    } else {
        fmt.Println("Unexpected response code:", response.Code)
    }
}
