package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Note: these tests require the validator package to be installed
// go get github.com/go-playground/validator/v10

// TestProfileValidation example: profile validation
func TestProfileValidation(t *testing.T) {
	// e.g., if the validator package is not installed, skip the test
	t.Skip("Please install go-playground/validator: go get github.com/go-playground/validator/v10")

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
			name: "Valid request",
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
			name: "Empty name - should fail",
			request: ProfileRequest{
				Name: "",
			},
			wantErr: true,
		},
		{
			name: "Invalid status value - should fail",
			request: ProfileRequest{
				Name:   "test",
				Status: "invalid",
			},
			wantErr: true,
		},
		{
			name: "Rate out of range - should fail",
			request: ProfileRequest{
				Name:   "test",
				UpRate: 99999999,
			},
			wantErr: true,
		},
		{
			name: "Concurrent sessions out of range - should fail",
			request: ProfileRequest{
				Name:      "test",
				ActiveNum: 200,
			},
			wantErr: true,
		},
		{
			name: "Invalid address pool format - should fail",
			request: ProfileRequest{
				Name:     "test",
				AddrPool: "192.168.1.0", // Missing subnet mask
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

// TestUserValidation example: user validation
func TestUserValidation(t *testing.T) {
	t.Skip("Please install go-playground/validator")

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
			name: "Valid user",
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
			name: "Username too short",
			request: UserRequest{
				Username: "ab",
				Password: "password123",
				Status:   "enabled",
			},
			wantErr: true,
		},
		{
			name: "Password too short",
			request: UserRequest{
				Username: "testuser",
				Password: "12345",
				Status:   "enabled",
			},
			wantErr: true,
		},
		{
			name: "Invalid email format",
			request: UserRequest{
				Username: "testuser",
				Password: "password123",
				Email:    "invalid-email",
				Status:   "enabled",
			},
			wantErr: true,
		},
		{
			name: "Invalid mobile length",
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

// TestCustomValidations tests custom validation rules
func TestCustomValidations(t *testing.T) {
	t.Skip("Please install go-playground/validator")

	validator := NewValidator()

	t.Run("Address pool validation", func(t *testing.T) {
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
			assert.NoError(t, err, "Address pool %s should be valid", pool)
		}

		invalidCases := []string{
			"192.168.1.0",
			"invalid",
			"192.168.1.0/",
		}

		for _, pool := range invalidCases {
			req := Request{AddrPool: pool}
			err := validator.Validate(&req)
			assert.Error(t, err, "Address pool %s should be invalid", pool)
		}
	})

	t.Run("RADIUS status validation", func(t *testing.T) {
		type Request struct {
			Status string `validate:"radiusstatus"`
		}

		validStatuses := []string{"enabled", "disabled", ""}
		for _, status := range validStatuses {
			req := Request{Status: status}
			err := validator.Validate(&req)
			assert.NoError(t, err, "Status %s should be valid", status)
		}

		invalidStatuses := []string{"active", "inactive", "pending"}
		for _, status := range invalidStatuses {
			req := Request{Status: status}
			err := validator.Validate(&req)
			assert.Error(t, err, "Status %s should be invalid", status)
		}
	})

	t.Run("Username validation", func(t *testing.T) {
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
			assert.NoError(t, err, "Username %s should be valid", username)
		}

		invalidUsernames := []string{
			"test user", // Contains a space
			"test#user", // Contains special characters
			"userΩ",     // Contains a non-Latin character
		}

		for _, username := range invalidUsernames {
			req := Request{Username: username}
			err := validator.Validate(&req)
			assert.Error(t, err, "Username %s should be invalid", username)
		}
	})
}
