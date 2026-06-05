package vendorparsers

import "testing"

func TestParseVlanIDs(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		vlanid1 int64
		vlanid2 int64
	}{
		{"port and dual vlan", "3/0/1:2814.727", 2814, 727},
		{"port and single vlan", "3/0/1:2814", 2814, 0},
		{"kv single vlan", "slot=2;subslot=2;port=22;vlanid=503;", 503, 0},
		{"kv dual vlan", "slot=2;subslot=2;port=22;vlanid=503;vlanid2=100;", 503, 100},
		{"empty", "", 0, 0},
		{"invalid", "invalid-format", 0, 0},
		{"port only no vlan", "3/0/1:", 0, 0},
		{"large vlan", "1/0/1:4094.4093", 4094, 4093},
		{"leading interface name", "GigabitEthernet 1/0/1:100", 100, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v1, v2 := ParseVlanIDs(tt.input)
			if v1 != tt.vlanid1 || v2 != tt.vlanid2 {
				t.Errorf("ParseVlanIDs(%q) = (%d, %d), want (%d, %d)",
					tt.input, v1, v2, tt.vlanid1, tt.vlanid2)
			}
		})
	}
}
