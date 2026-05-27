package vendors

import (
	"sort"
	"sync"

	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
)

// VendorInfo holds metadata and handlers for a RADIUS vendor
type VendorInfo struct {
	Code        string
	Name        string
	Description string
	Parser      vendorparsers.VendorParser
	Builder     vendorparsers.VendorResponseBuilder
}

// VendorRegistry manages vendor registrations
type VendorRegistry struct {
	mu      sync.RWMutex
	vendors map[string]*VendorInfo
}

var globalRegistry = NewVendorRegistry()

// NewVendorRegistry creates a new vendor registry
func NewVendorRegistry() *VendorRegistry {
	return &VendorRegistry{
		vendors: make(map[string]*VendorInfo),
	}
}

// Register adds or updates a vendor in the registry
func (r *VendorRegistry) Register(info *VendorInfo) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if existing, exists := r.vendors[info.Code]; exists {
		// Merge with existing registration
		if info.Name != "" {
			existing.Name = info.Name
		}
		if info.Description != "" {
			existing.Description = info.Description
		}
		if info.Parser != nil {
			existing.Parser = info.Parser
		}
		if info.Builder != nil {
			existing.Builder = info.Builder
		}
		return nil
	}

	r.vendors[info.Code] = info
	return nil
}

// Get retrieves a vendor by code
func (r *VendorRegistry) Get(code string) (*VendorInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	info, ok := r.vendors[code]
	return info, ok
}

// List returns all registered vendors
func (r *VendorRegistry) List() []*VendorInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]*VendorInfo, 0, len(r.vendors))
	for _, info := range r.vendors {
		list = append(list, info)
	}

	// Sort by name for consistent output
	sort.Slice(list, func(i, j int) bool {
		return list[i].Name < list[j].Name
	})

	return list
}

// Global accessors

// Register adds a vendor to the global registry
func Register(info *VendorInfo) error {
	return globalRegistry.Register(info)
}

// Get retrieves a vendor from the global registry
func Get(code string) (*VendorInfo, bool) {
	return globalRegistry.Get(code)
}

// List returns all registered vendors from the global registry
func List() []*VendorInfo {
	return globalRegistry.List()
}

// ResetGlobalRegistry resets the global registry (for testing)
func ResetGlobalRegistry() {
	globalRegistry = NewVendorRegistry()
}
