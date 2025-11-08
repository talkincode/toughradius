package netutils

import (
	"testing"

	iplib "github.com/c-robinson/iplib"
)

// TestParseIpNet tests parsing IP networks
func TestParseIpNet(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "IPv4 with CIDR",
			input:   "192.168.1.0/24",
			wantErr: false,
		},
		{
			name:    "IPv4 without CIDR (default /32)",
			input:   "192.168.1.1",
			wantErr: false,
		},
		{
			name:    "IPv6 with CIDR",
			input:   "2001:db8::/32",
			wantErr: false,
		},
		{
			name:    "IPv6 without CIDR",
			input:   "2001:db8::1",
			wantErr: false,
		},
		{
			name:    "Localhost IPv4",
			input:   "127.0.0.1",
			wantErr: false,
		},
		{
			name:    "Full IPv4 network",
			input:   "10.0.0.0/8",
			wantErr: false,
		},
		{
			name:    "Invalid IP",
			input:   "invalid.ip.address",
			wantErr: true,
		},
		{
			name:    "Invalid CIDR",
			input:   "192.168.1.1/99",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inet, err := ParseIpNet(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if inet == nil {
				t.Error("expected non-nil network, got nil")
			}
		})
	}
}

// TestParseIpNet_DefaultCIDR tests that /32 is added when missing
func TestParseIpNet_DefaultCIDR(t *testing.T) {
	inet, err := ParseIpNet("192.168.1.100")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be treated as /32 (single host)
	if inet.IP().String() != "192.168.1.100" {
		t.Errorf("expected IP 192.168.1.100, got %s", inet.IP().String())
	}
}

// TestParseIpNet_IPv4 tests IPv4 network parsing
func TestParseIpNet_IPv4(t *testing.T) {
	tests := []struct {
		input       string
		expectedIP  string
		expectedNet string
	}{
		{"192.168.1.0/24", "192.168.1.0", "192.168.1.0/24"},
		{"10.0.0.0/8", "10.0.0.0", "10.0.0.0/8"},
		{"172.16.0.0/12", "172.16.0.0", "172.16.0.0/12"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			inet, err := ParseIpNet(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if inet.IP().String() != tt.expectedIP {
				t.Errorf("expected IP %s, got %s", tt.expectedIP, inet.IP().String())
			}
		})
	}
}

// TestParseIpNet_IPv6 tests IPv6 network parsing
func TestParseIpNet_IPv6(t *testing.T) {
	tests := []struct {
		input      string
		expectedIP string
	}{
		{"2001:db8::/32", "2001:db8::"},
		{"fe80::/10", "fe80::"},
		{"::1/128", "::1"}, // Added CIDR to match expected behavior
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			inet, err := ParseIpNet(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if inet.IP().String() != tt.expectedIP {
				t.Errorf("expected IP %s, got %s", tt.expectedIP, inet.IP().String())
			}
		})
	}
}

// TestContainsNetAddr tests checking if networks contain an address
func TestContainsNetAddr(t *testing.T) {
	// Setup test networks
	nets := []iplib.Net{}

	net1, _ := ParseIpNet("192.168.1.0/24")
	nets = append(nets, net1)

	net2, _ := ParseIpNet("10.0.0.0/8")
	nets = append(nets, net2)

	net3, _ := ParseIpNet("172.16.0.0/16")
	nets = append(nets, net3)

	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		{
			name:     "IP in first network",
			ip:       "192.168.1.100",
			expected: true,
		},
		{
			name:     "IP in second network",
			ip:       "10.5.5.5",
			expected: true,
		},
		{
			name:     "IP in third network",
			ip:       "172.16.100.1",
			expected: true,
		},
		{
			name:     "IP not in any network",
			ip:       "8.8.8.8",
			expected: false,
		},
		{
			name:     "IP at network boundary",
			ip:       "192.168.1.0",
			expected: true,
		},
		{
			name:     "IP at broadcast address",
			ip:       "192.168.1.255",
			expected: true,
		},
		{
			name:     "Invalid IP",
			ip:       "invalid.ip",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ContainsNetAddr(nets, tt.ip)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestContainsNetAddr_EmptyNetworks tests with empty network list
func TestContainsNetAddr_EmptyNetworks(t *testing.T) {
	nets := []iplib.Net{}

	result := ContainsNetAddr(nets, "192.168.1.1")
	if result != false {
		t.Error("empty network list should not contain any IP")
	}
}

// TestContainsNetAddr_SingleNetwork tests with single network
func TestContainsNetAddr_SingleNetwork(t *testing.T) {
	nets := []iplib.Net{}
	net1, _ := ParseIpNet("192.168.0.0/16")
	nets = append(nets, net1)

	tests := []struct {
		ip       string
		expected bool
	}{
		{"192.168.1.1", true},
		{"192.168.255.255", true},
		{"192.169.1.1", false},
		{"10.0.0.1", false},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			result := ContainsNetAddr(nets, tt.ip)
			if result != tt.expected {
				t.Errorf("for IP %s, expected %v, got %v", tt.ip, tt.expected, result)
			}
		})
	}
}

// TestContainsNetAddr_OverlappingNetworks tests with overlapping networks
func TestContainsNetAddr_OverlappingNetworks(t *testing.T) {
	nets := []iplib.Net{}

	net1, _ := ParseIpNet("192.168.0.0/16")
	nets = append(nets, net1)

	net2, _ := ParseIpNet("192.168.1.0/24")
	nets = append(nets, net2)

	// IP in both networks should return true
	result := ContainsNetAddr(nets, "192.168.1.100")
	if !result {
		t.Error("IP should be found in overlapping networks")
	}
}

// TestContainsNetAddr_IPv6 tests with IPv6 networks
func TestContainsNetAddr_IPv6(t *testing.T) {
	nets := []iplib.Net{}

	net1, _ := ParseIpNet("2001:db8::/32")
	nets = append(nets, net1)

	tests := []struct {
		ip       string
		expected bool
	}{
		{"2001:db8::1", true},
		{"2001:db8:1::1", true},
		{"2001:db9::1", false},
		{"::1", false},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			result := ContainsNetAddr(nets, tt.ip)
			if result != tt.expected {
				t.Errorf("for IP %s, expected %v, got %v", tt.ip, tt.expected, result)
			}
		})
	}
}

// TestContainsNetAddr_MixedIPv4IPv6 tests with mixed IPv4 and IPv6 networks
func TestContainsNetAddr_MixedIPv4IPv6(t *testing.T) {
	nets := []iplib.Net{}

	net4, _ := ParseIpNet("192.168.1.0/24")
	nets = append(nets, net4)

	net6, _ := ParseIpNet("2001:db8::/32")
	nets = append(nets, net6)

	tests := []struct {
		ip       string
		expected bool
	}{
		{"192.168.1.100", true},
		{"2001:db8::1", true},
		{"10.0.0.1", false},
		{"2001:db9::1", false},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			result := ContainsNetAddr(nets, tt.ip)
			if result != tt.expected {
				t.Errorf("for IP %s, expected %v, got %v", tt.ip, tt.expected, result)
			}
		})
	}
}

// BenchmarkParseIpNet benchmarks IP network parsing
func BenchmarkParseIpNet(b *testing.B) {
	testIPs := []string{
		"192.168.1.0/24",
		"10.0.0.0/8",
		"172.16.0.0/12",
		"2001:db8::/32",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseIpNet(testIPs[i%len(testIPs)])
	}
}

// BenchmarkContainsNetAddr benchmarks network containment check
func BenchmarkContainsNetAddr(b *testing.B) {
	nets := []iplib.Net{}
	net1, _ := ParseIpNet("192.168.0.0/16")
	nets = append(nets, net1)
	net2, _ := ParseIpNet("10.0.0.0/8")
	nets = append(nets, net2)
	net3, _ := ParseIpNet("172.16.0.0/12")
	nets = append(nets, net3)

	testIPs := []string{
		"192.168.1.1",
		"10.5.5.5",
		"172.20.1.1",
		"8.8.8.8",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ContainsNetAddr(nets, testIPs[i%len(testIPs)])
	}
}
