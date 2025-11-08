package radiusd

import (
	"fmt"
	"testing"
)

func TestParseVlanIds(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		vlanid1 int64
		vlanid2 int64
	}{
		{
			name:    "标准格式1 - 带端口和VLAN",
			input:   "3/0/1:2814.727",
			vlanid1: 2814,
			vlanid2: 727,
		},
		{
			name:    "标准格式1 - 仅VLAN",
			input:   "3/0/1:2814",
			vlanid1: 2814,
			vlanid2: 0,
		},
		{
			name:    "标准格式2 - slot/subslot/port",
			input:   "slot=2;subslot=2;port=22;vlanid=503;",
			vlanid1: 503,
			vlanid2: 0,
		},
		{
			name:    "标准格式2 - 双VLAN",
			input:   "slot=2;subslot=2;port=22;vlanid=503;vlanid2=100;",
			vlanid1: 503,
			vlanid2: 100,
		},
		{
			name:    "空字符串",
			input:   "",
			vlanid1: 0,
			vlanid2: 0,
		},
		{
			name:    "无效格式",
			input:   "invalid-format",
			vlanid1: 0,
			vlanid2: 0,
		},
		{
			name:    "仅端口无VLAN",
			input:   "3/0/1:",
			vlanid1: 0,
			vlanid2: 0,
		},
		{
			name:    "大VLAN ID",
			input:   "1/0/1:4094.4093",
			vlanid1: 4094,
			vlanid2: 4093,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vlanid1, vlanid2 := ParseVlanIds(tt.input)
			if vlanid1 != tt.vlanid1 {
				t.Errorf("expected vlanid1=%d, got %d", tt.vlanid1, vlanid1)
			}
			if vlanid2 != tt.vlanid2 {
				t.Errorf("expected vlanid2=%d, got %d", tt.vlanid2, vlanid2)
			}
		})
	}
}

func TestParseVlanIdsOutput(t *testing.T) {
	// 保留原有的输出测试
	s := "3/0/1:2814.727"
	vlanid1, vlanid2 := ParseVlanIds(s)
	fmt.Printf("Input: %s, VLAN1: %d, VLAN2: %d\n", s, vlanid1, vlanid2)

	s2 := "3/0/1:2814"
	vlanid1, vlanid2 = ParseVlanIds(s2)
	fmt.Printf("Input: %s, VLAN1: %d, VLAN2: %d\n", s2, vlanid1, vlanid2)

	s3 := "slot=2;subslot=2;port=22;vlanid=503;"
	vlanid1, vlanid2 = ParseVlanIds(s3)
	fmt.Printf("Input: %s, VLAN1: %d, VLAN2: %d\n", s3, vlanid1, vlanid2)
}

func TestParseVlanIdsEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "带空格的输入",
			input:   " 3/0/1:2814.727 ",
			wantErr: false,
		},
		{
			name:    "多个分隔符",
			input:   "3/0/1:2814.727.100",
			wantErr: false,
		},
		{
			name:    "负数VLAN（无效）",
			input:   "3/0/1:-1",
			wantErr: false, // 解析失败会返回0
		},
		{
			name:    "超大VLAN ID",
			input:   "3/0/1:99999",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vlanid1, vlanid2 := ParseVlanIds(tt.input)
			// 只验证不会 panic
			t.Logf("Input: %s, VLAN1: %d, VLAN2: %d", tt.input, vlanid1, vlanid2)
		})
	}
}

func TestParseVlanIdsRegexMatch(t *testing.T) {
	// 测试正则表达式匹配的不同格式
	inputs := []string{
		"1/0/1:100",
		"2/1/10:200.300",
		"GigabitEthernet 1/0/1:100",
		"vlanid=100;",
		"vlanid=100;vlanid2=200;",
		"slot=1;subslot=0;port=1;vlanid=100;",
	}

	for _, input := range inputs {
		t.Run("Input_"+input, func(t *testing.T) {
			vlanid1, vlanid2 := ParseVlanIds(input)
			t.Logf("Input: %s => VLAN1: %d, VLAN2: %d", input, vlanid1, vlanid2)

			// 验证结果合理性
			if vlanid1 < 0 || vlanid1 > 4094 {
				if vlanid1 != 0 { // 0 表示未匹配，是合法的
					t.Errorf("VLAN1 out of valid range (0-4094): %d", vlanid1)
				}
			}
			if vlanid2 < 0 || vlanid2 > 4094 {
				if vlanid2 != 0 {
					t.Errorf("VLAN2 out of valid range (0-4094): %d", vlanid2)
				}
			}
		})
	}
}
