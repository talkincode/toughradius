package domain

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSysOprLog_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		log     SysOprLog
		wantErr bool
		check   func(t *testing.T, data []byte)
	}{
		{
			name: "正常序列化操作日志",
			log: SysOprLog{
				ID:        1,
				OprName:   "admin",
				OprIp:     "192.168.1.100",
				OptAction: "login",
				OptDesc:   "用户登录系统",
				OptTime:   time.Date(2025, 11, 8, 14, 30, 0, 0, time.UTC),
			},
			wantErr: false,
			check: func(t *testing.T, data []byte) {
				var result map[string]interface{}
				err := json.Unmarshal(data, &result)
				require.NoError(t, err)

				// ID 字段的 json tag 包含 string，所以序列化为字符串
				assert.Equal(t, "1", result["id"])
				assert.Equal(t, "admin", result["opr_name"])
				assert.Equal(t, "192.168.1.100", result["opr_ip"])
				assert.Equal(t, "login", result["opt_action"])
				assert.Equal(t, "用户登录系统", result["opt_desc"])

				// 验证时间格式为 RFC3339
				assert.Equal(t, "2025-11-08T14:30:00Z", result["opt_time"])
			},
		},
		{
			name: "零值时间序列化",
			log: SysOprLog{
				ID:      2,
				OprName: "system",
				OptTime: time.Time{},
			},
			wantErr: false,
			check: func(t *testing.T, data []byte) {
				var result map[string]interface{}
				err := json.Unmarshal(data, &result)
				require.NoError(t, err)

				// 零值时间应该序列化为 "0001-01-01T00:00:00Z"
				assert.Contains(t, result["opt_time"], "0001-01-01")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.log.MarshalJSON()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, data)

			if tt.check != nil {
				tt.check(t, data)
			}
		})
	}
}

func TestRadiusUser_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		user    RadiusUser
		wantErr bool
		check   func(t *testing.T, data []byte)
	}{
		{
			name: "正常序列化用户信息",
			user: RadiusUser{
				ID:         1,
				NodeId:     100,
				ProfileId:  5,
				Realname:   "张三",
				Mobile:     "13800138000",
				Username:   "test001",
				Password:   "password123",
				ActiveNum:  1,
				UpRate:     10240,
				DownRate:   20480,
				Status:     "enabled",
				ExpireTime: time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
				LastOnline: time.Date(2025, 11, 8, 14, 30, 0, 0, time.UTC),
				CreatedAt:  time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			wantErr: false,
			check: func(t *testing.T, data []byte) {
				var result map[string]interface{}
				err := json.Unmarshal(data, &result)
				require.NoError(t, err)

				assert.Equal(t, "1", result["id"])
				assert.Equal(t, "100", result["node_id"])
				assert.Equal(t, "5", result["profile_id"])
				assert.Equal(t, "张三", result["realname"])
				assert.Equal(t, "13800138000", result["mobile"])
				assert.Equal(t, "test001", result["username"])
				assert.Equal(t, "password123", result["password"])
				assert.Equal(t, "enabled", result["status"])

				// 验证过期时间格式 YYYYMMDDHHMMSS (2006-01-02 15:04:05)
				assert.Equal(t, "2025-12-31 23:59:59", result["expire_time"])

				// 验证最后在线时间格式 YYYYMMDDHHMM (2006-01-02 15:04)
				assert.Equal(t, "2025-11-08 14:30", result["last_online"])
			},
		},
		{
			name: "零值时间序列化",
			user: RadiusUser{
				ID:         2,
				Username:   "test002",
				ExpireTime: time.Time{},
				LastOnline: time.Time{},
			},
			wantErr: false,
			check: func(t *testing.T, data []byte) {
				var result map[string]interface{}
				err := json.Unmarshal(data, &result)
				require.NoError(t, err)

				// 零值时间格式化后的结果
				assert.NotEmpty(t, result["expire_time"])
				assert.NotEmpty(t, result["last_online"])
			},
		},
		{
			name: "包含完整字段的用户",
			user: RadiusUser{
				ID:         3,
				NodeId:     200,
				ProfileId:  10,
				Realname:   "李四",
				Mobile:     "13900139000",
				Username:   "test003",
				Password:   "pass456",
				AddrPool:   "pool1",
				ActiveNum:  2,
				UpRate:     5120,
				DownRate:   10240,
				Vlanid1:    100,
				Vlanid2:    200,
				IpAddr:     "10.0.0.100",
				MacAddr:    "00:11:22:33:44:55",
				BindVlan:   1,
				BindMac:    1,
				ExpireTime: time.Date(2026, 6, 30, 23, 59, 59, 0, time.UTC),
				Status:     "disabled",
				Remark:     "测试账号",
				LastOnline: time.Date(2025, 11, 7, 10, 15, 0, 0, time.UTC),
			},
			wantErr: false,
			check: func(t *testing.T, data []byte) {
				var result map[string]interface{}
				err := json.Unmarshal(data, &result)
				require.NoError(t, err)

				assert.Equal(t, "test003", result["username"])
				assert.Equal(t, "pool1", result["addr_pool"])
				assert.Equal(t, "10.0.0.100", result["ip_addr"])
				assert.Equal(t, "00:11:22:33:44:55", result["mac_addr"])
				assert.Equal(t, "disabled", result["status"])
				assert.Equal(t, "测试账号", result["remark"])
				assert.Equal(t, "2026-06-30 23:59:59", result["expire_time"])
				assert.Equal(t, "2025-11-07 10:15", result["last_online"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.user.MarshalJSON()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, data)

			if tt.check != nil {
				tt.check(t, data)
			}
		})
	}
}

func TestRadiusUser_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
		check   func(t *testing.T, user *RadiusUser)
	}{
		{
			name: "正常反序列化用户信息",
			json: `{
				"id": "1",
				"node_id": "100",
				"profile_id": "5",
				"username": "test001",
				"password": "password123",
				"status": "enabled",
				"expire_time": "2025-12-31 23:59:59",
				"last_online": "2025-11-08 14:30:00"
			}`,
			wantErr: false,
			check: func(t *testing.T, user *RadiusUser) {
				assert.Equal(t, int64(1), user.ID)
				assert.Equal(t, int64(100), user.NodeId)
				assert.Equal(t, int64(5), user.ProfileId)
				assert.Equal(t, "test001", user.Username)
				assert.Equal(t, "password123", user.Password)
				assert.Equal(t, "enabled", user.Status)

				// 验证过期时间解析（应该被设置为 23:59:59）
				expectedExpire := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)
				assert.True(t, user.ExpireTime.Equal(expectedExpire),
					"ExpireTime 应该是 %v，但得到 %v", expectedExpire, user.ExpireTime)

				// 验证最后在线时间解析
				assert.Equal(t, 2025, user.LastOnline.Year())
				assert.Equal(t, time.November, user.LastOnline.Month())
				assert.Equal(t, 8, user.LastOnline.Day())
			},
		},
		{
			name: "仅日期的过期时间",
			json: `{
				"id": "2",
				"username": "test002",
				"expire_time": "2026-06-30",
				"last_online": "2025-11-07"
			}`,
			wantErr: false,
			check: func(t *testing.T, user *RadiusUser) {
				assert.Equal(t, "test002", user.Username)

				// 过期时间应该被设置为当天的 23:59:59
				expectedExpire := time.Date(2026, 6, 30, 23, 59, 59, 0, time.UTC)
				assert.True(t, user.ExpireTime.Equal(expectedExpire),
					"ExpireTime 应该是 %v，但得到 %v", expectedExpire, user.ExpireTime)
			},
		},
		{
			name: "格式化的时间字符串",
			json: `{
				"id": "3",
				"username": "test003",
				"expire_time": "2025-12-31 23:59:59",
				"last_online": "2025-11-08 14:30:00"
			}`,
			wantErr: false,
			check: func(t *testing.T, user *RadiusUser) {
				assert.Equal(t, "test003", user.Username)

				// 验证时间正确解析
				assert.Equal(t, 2025, user.ExpireTime.Year())
				assert.Equal(t, time.December, user.ExpireTime.Month())
				assert.Equal(t, 2025, user.LastOnline.Year())
				assert.Equal(t, time.November, user.LastOnline.Month())
			},
		},
		{
			name: "包含完整字段的 JSON",
			json: `{
				"id": "4",
				"node_id": "200",
				"profile_id": "10",
				"realname": "李四",
				"mobile": "13900139000",
				"username": "test004",
				"password": "pass789",
				"addr_pool": "pool2",
				"active_num": 3,
				"up_rate": 2048,
				"down_rate": 4096,
				"vlanid1": 10,
				"vlanid2": 20,
				"ip_addr": "10.0.0.200",
				"mac_addr": "AA:BB:CC:DD:EE:FF",
				"bind_vlan": 1,
				"bind_mac": 0,
				"status": "disabled",
				"remark": "完整测试",
				"expire_time": "2027-12-31 23:59:59",
				"last_online": "2025-11-08 16:00:00"
			}`,
			wantErr: false,
			check: func(t *testing.T, user *RadiusUser) {
				assert.Equal(t, int64(4), user.ID)
				assert.Equal(t, "李四", user.Realname)
				assert.Equal(t, "13900139000", user.Mobile)
				assert.Equal(t, "test004", user.Username)
				assert.Equal(t, "pool2", user.AddrPool)
				assert.Equal(t, 3, user.ActiveNum)
				assert.Equal(t, 2048, user.UpRate)
				assert.Equal(t, 4096, user.DownRate)
				assert.Equal(t, "10.0.0.200", user.IpAddr)
				assert.Equal(t, "AA:BB:CC:DD:EE:FF", user.MacAddr)
				assert.Equal(t, "disabled", user.Status)
				assert.Equal(t, "完整测试", user.Remark)
			},
		},
		{
			name:    "无效的 JSON",
			json:    `{invalid json}`,
			wantErr: true,
			check:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var user RadiusUser
			err := json.Unmarshal([]byte(tt.json), &user)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.check != nil {
				tt.check(t, &user)
			}
		})
	}
}

func TestRadiusUser_MarshalUnmarshal_RoundTrip(t *testing.T) {
	// 测试序列化和反序列化的往返转换
	original := RadiusUser{
		ID:         100,
		NodeId:     50,
		ProfileId:  10,
		Realname:   "往返测试",
		Mobile:     "13812345678",
		Username:   "roundtrip",
		Password:   "test123",
		AddrPool:   "pool1",
		ActiveNum:  2,
		UpRate:     1024,
		DownRate:   2048,
		Vlanid1:    10,
		Vlanid2:    20,
		IpAddr:     "192.168.1.100",
		MacAddr:    "11:22:33:44:55:66",
		BindVlan:   1,
		BindMac:    1,
		ExpireTime: time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
		Status:     "enabled",
		Remark:     "测试备注",
		LastOnline: time.Date(2025, 11, 8, 10, 0, 0, 0, time.UTC),
	}

	// 序列化
	data, err := json.Marshal(&original)
	require.NoError(t, err)
	t.Logf("序列化后的 JSON: %s", string(data))

	// 反序列化
	var decoded RadiusUser
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	// 验证关键字段
	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.NodeId, decoded.NodeId)
	assert.Equal(t, original.ProfileId, decoded.ProfileId)
	assert.Equal(t, original.Username, decoded.Username)
	assert.Equal(t, original.Password, decoded.Password)
	assert.Equal(t, original.Realname, decoded.Realname)
	assert.Equal(t, original.Mobile, decoded.Mobile)
	assert.Equal(t, original.Status, decoded.Status)
	assert.Equal(t, original.AddrPool, decoded.AddrPool)
	assert.Equal(t, original.IpAddr, decoded.IpAddr)
	assert.Equal(t, original.MacAddr, decoded.MacAddr)
	assert.Equal(t, original.Remark, decoded.Remark)

	// 验证时间字段（由于格式化和解析，可能有微小差异）
	assert.True(t, original.ExpireTime.Equal(decoded.ExpireTime),
		"ExpireTime 不匹配: 原始=%v, 解码=%v", original.ExpireTime, decoded.ExpireTime)

	// LastOnline 由于格式化为分钟精度，秒数会丢失
	assert.Equal(t, original.LastOnline.Year(), decoded.LastOnline.Year())
	assert.Equal(t, original.LastOnline.Month(), decoded.LastOnline.Month())
	assert.Equal(t, original.LastOnline.Day(), decoded.LastOnline.Day())
	assert.Equal(t, original.LastOnline.Hour(), decoded.LastOnline.Hour())
	assert.Equal(t, original.LastOnline.Minute(), decoded.LastOnline.Minute())
}

func TestSysOprLog_MarshalUnmarshal_RoundTrip(t *testing.T) {
	// 测试操作日志的序列化和反序列化往返
	original := SysOprLog{
		ID:        123,
		OprName:   "admin",
		OprIp:     "192.168.1.1",
		OptAction: "create_user",
		OptDesc:   "创建新用户 test001",
		OptTime:   time.Date(2025, 11, 8, 15, 30, 45, 0, time.UTC),
	}

	// 序列化
	data, err := json.Marshal(&original)
	require.NoError(t, err)
	t.Logf("序列化后的 JSON: %s", string(data))

	// 反序列化
	var decoded SysOprLog
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	// 验证字段
	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.OprName, decoded.OprName)
	assert.Equal(t, original.OprIp, decoded.OprIp)
	assert.Equal(t, original.OptAction, decoded.OptAction)
	assert.Equal(t, original.OptDesc, decoded.OptDesc)

	// 验证时间（RFC3339 格式应该精确保留）
	assert.True(t, original.OptTime.Equal(decoded.OptTime),
		"OptTime 不匹配: 原始=%v, 解码=%v", original.OptTime, decoded.OptTime)
}

func BenchmarkRadiusUser_MarshalJSON(b *testing.B) {
	user := RadiusUser{
		ID:         1,
		NodeId:     100,
		ProfileId:  5,
		Realname:   "性能测试",
		Mobile:     "13800138000",
		Username:   "benchmark",
		Password:   "password123",
		ActiveNum:  1,
		UpRate:     10240,
		DownRate:   20480,
		Status:     "enabled",
		ExpireTime: time.Now().AddDate(1, 0, 0),
		LastOnline: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(&user)
	}
}

func BenchmarkRadiusUser_UnmarshalJSON(b *testing.B) {
	jsonData := []byte(`{
		"id": "1",
		"node_id": "100",
		"profile_id": "5",
		"username": "benchmark",
		"password": "password123",
		"status": "enabled",
		"expire_time": "2026-11-08 23:59:59",
		"last_online": "2025-11-08 14:30:00"
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var user RadiusUser
		_ = json.Unmarshal(jsonData, &user)
	}
}

func BenchmarkSysOprLog_MarshalJSON(b *testing.B) {
	log := SysOprLog{
		ID:        1,
		OprName:   "admin",
		OprIp:     "192.168.1.100",
		OptAction: "login",
		OptDesc:   "用户登录系统",
		OptTime:   time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(&log)
	}
}
