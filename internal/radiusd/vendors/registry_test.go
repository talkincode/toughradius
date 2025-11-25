package vendors

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
	"layeh.com/radius"
)

// MockParser implements VendorParser for testing
type MockParser struct {
	code string
	name string
}

func (p *MockParser) VendorCode() string { return p.code }
func (p *MockParser) VendorName() string { return p.name }
func (p *MockParser) Parse(r *radius.Request) (*vendorparsers.VendorRequest, error) {
	return &vendorparsers.VendorRequest{}, nil
}

func TestVendorRegistry(t *testing.T) {
	registry := NewVendorRegistry()

	t.Run("Register and Get", func(t *testing.T) {
		info := &VendorInfo{
			Code:        "2011",
			Name:        "Huawei",
			Description: "Huawei Vendor",
			Parser:      &MockParser{code: "2011", name: "Huawei"},
		}

		err := registry.Register(info)
		assert.NoError(t, err)

		got, ok := registry.Get("2011")
		assert.True(t, ok)
		assert.Equal(t, info, got)
	})

	t.Run("Register Duplicate (Merge)", func(t *testing.T) {
		info := &VendorInfo{
			Code: "2011",
			Name: "Huawei Updated",
		}
		err := registry.Register(info)
		assert.NoError(t, err)

		got, ok := registry.Get("2011")
		assert.True(t, ok)
		assert.Equal(t, "Huawei Updated", got.Name)
		assert.Equal(t, "Huawei Vendor", got.Description) // Should preserve description from previous test
	})

	t.Run("Get Non-Existent", func(t *testing.T) {
		_, ok := registry.Get("9999")
		assert.False(t, ok)
	})

	t.Run("List", func(t *testing.T) {
		registry := NewVendorRegistry()
		_ = registry.Register(&VendorInfo{Code: "1", Name: "A"}) //nolint:errcheck
		_ = registry.Register(&VendorInfo{Code: "2", Name: "B"}) //nolint:errcheck

		list := registry.List()
		assert.Len(t, list, 2)
	})
}

func TestGlobalRegistry(t *testing.T) {
	ResetGlobalRegistry()

	info := &VendorInfo{
		Code: "2011",
		Name: "Huawei",
	}

	err := Register(info)
	assert.NoError(t, err)

	got, ok := Get("2011")
	assert.True(t, ok)
	assert.Equal(t, info, got)

	list := List()
	assert.Len(t, list, 1)
}
