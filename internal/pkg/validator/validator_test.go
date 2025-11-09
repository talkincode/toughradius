package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// 注意：这些测试需要先安装 validator 包
// go get github.com/go-playground/validator/v10

// TestProfileValidation 示例：Profile 验证测试
func TestProfileValidation(t *testing.T) {
	// 如果没有安装 validator，跳过测试
	t.Skip("需要先安装 go-playground/validator: go get github.com/go-playground/validator/v10")

	validator := NewValidator()

	type ProfileRequest struct {
		Name      string `validate:"required,min=1,max=100"`
		Status    string `validate:"omitempty,oneof=enabled disabled"`
		AddrPool  string `validate:"omitempty,addrpool"`
		ActiveNum int    `validate:"gte=0,lte=100"`
		UpRate    int    `validate:"gte=0,lte=10000000"`
		DownRate  int    `validate:"gte=0,lte=10000000"`
	}

	tests := []struct {
		name    string
		request ProfileRequest
		wantErr bool
	}{
		{
			name: "有效的请求",
			request: ProfileRequest{
				Name:      "test-profile",
				Status:    "enabled",
				AddrPool:  "192.168.1.0/24",
				ActiveNum: 1,
				UpRate:    1024,
				DownRate:  2048,
			},
			wantErr: false,
		},
		{
			name: "名称为空 - 应该失败",
			request: ProfileRequest{
				Name: "",
			},
			wantErr: true,
		},
		{
			name: "状态值无效 - 应该失败",
			request: ProfileRequest{
				Name:   "test",
				Status: "invalid",
			},
			wantErr: true,
		},
		{
			name: "速率超出范围 - 应该失败",
			request: ProfileRequest{
				Name:   "test",
				UpRate: 99999999,
			},
			wantErr: true,
		},
		{
			name: "并发数超出范围 - 应该失败",
			request: ProfileRequest{
				Name:      "test",
				ActiveNum: 200,
			},
			wantErr: true,
		},
		{
			name: "地址池格式错误 - 应该失败",
			request: ProfileRequest{
				Name:     "test",
				AddrPool: "192.168.1.0", // 缺少掩码
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(&tt.request)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestUserValidation 示例：User 验证测试
func TestUserValidation(t *testing.T) {
	t.Skip("需要先安装 go-playground/validator")

	validator := NewValidator()

	type UserRequest struct {
		Username string `validate:"required,min=3,max=50,username"`
		Password string `validate:"required,min=6,max=100"`
		Email    string `validate:"omitempty,email"`
		Mobile   string `validate:"omitempty,len=11,numeric"`
		Status   string `validate:"required,oneof=enabled disabled"`
	}

	tests := []struct {
		name    string
		request UserRequest
		wantErr bool
	}{
		{
			name: "有效的用户",
			request: UserRequest{
				Username: "testuser",
				Password: "password123",
				Email:    "test@example.com",
				Mobile:   "13800138000",
				Status:   "enabled",
			},
			wantErr: false,
		},
		{
			name: "用户名太短",
			request: UserRequest{
				Username: "ab",
				Password: "password123",
				Status:   "enabled",
			},
			wantErr: true,
		},
		{
			name: "密码太短",
			request: UserRequest{
				Username: "testuser",
				Password: "12345",
				Status:   "enabled",
			},
			wantErr: true,
		},
		{
			name: "邮箱格式错误",
			request: UserRequest{
				Username: "testuser",
				Password: "password123",
				Email:    "invalid-email",
				Status:   "enabled",
			},
			wantErr: true,
		},
		{
			name: "手机号长度错误",
			request: UserRequest{
				Username: "testuser",
				Password: "password123",
				Mobile:   "138001380",
				Status:   "enabled",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(&tt.request)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestCustomValidations 测试自定义验证规则
func TestCustomValidations(t *testing.T) {
	t.Skip("需要先安装 go-playground/validator")

	validator := NewValidator()

	t.Run("地址池验证", func(t *testing.T) {
		type Request struct {
			AddrPool string `validate:"addrpool"`
		}

		validCases := []string{
			"192.168.1.0/24",
			"10.0.0.0/8",
			"172.16.0.0/12",
		}

		for _, pool := range validCases {
			req := Request{AddrPool: pool}
			err := validator.Validate(&req)
			assert.NoError(t, err, "地址池 %s 应该是有效的", pool)
		}

		invalidCases := []string{
			"192.168.1.0",
			"invalid",
			"192.168.1.0/",
		}

		for _, pool := range invalidCases {
			req := Request{AddrPool: pool}
			err := validator.Validate(&req)
			assert.Error(t, err, "地址池 %s 应该是无效的", pool)
		}
	})

	t.Run("RADIUS 状态验证", func(t *testing.T) {
		type Request struct {
			Status string `validate:"radiusstatus"`
		}

		validStatuses := []string{"enabled", "disabled", ""}
		for _, status := range validStatuses {
			req := Request{Status: status}
			err := validator.Validate(&req)
			assert.NoError(t, err, "状态 %s 应该是有效的", status)
		}

		invalidStatuses := []string{"active", "inactive", "pending"}
		for _, status := range invalidStatuses {
			req := Request{Status: status}
			err := validator.Validate(&req)
			assert.Error(t, err, "状态 %s 应该是无效的", status)
		}
	})

	t.Run("用户名验证", func(t *testing.T) {
		type Request struct {
			Username string `validate:"username"`
		}

		validUsernames := []string{
			"testuser",
			"test_user",
			"test-user",
			"test@example.com",
			"user123",
		}

		for _, username := range validUsernames {
			req := Request{Username: username}
			err := validator.Validate(&req)
			assert.NoError(t, err, "用户名 %s 应该是有效的", username)
		}

		invalidUsernames := []string{
			"test user",  // 包含空格
			"test#user",  // 包含特殊字符
			"测试用户",     // 包含中文
		}

		for _, username := range invalidUsernames {
			req := Request{Username: username}
			err := validator.Validate(&req)
			assert.Error(t, err, "用户名 %s 应该是无效的", username)
		}
	})
}
