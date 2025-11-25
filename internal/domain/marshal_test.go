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
			name: "Serializing operation log",
			log: SysOprLog{
				ID:        1,
				OprName:   "admin",
				OprIp:     "192.168.1.100",
				OptAction: "login",
				OptDesc:   "User login",
				OptTime:   time.Date(2025, 11, 8, 14, 30, 0, 0, time.UTC),
			},
			wantErr: false,
			check: func(t *testing.T, data []byte) {
				var result map[string]interface{}
				err := json.Unmarshal(data, &result)
				require.NoError(t, err)

				// The ID field's json tag includes string, so it should serialize as a string
				assert.Equal(t, "1", result["id"])
				assert.Equal(t, "admin", result["opr_name"])
				assert.Equal(t, "192.168.1.100", result["opr_ip"])
				assert.Equal(t, "login", result["opt_action"])
				assert.Equal(t, "User login", result["opt_desc"])

				// Validate that the time format is RFC3339
				assert.Equal(t, "2025-11-08T14:30:00Z", result["opt_time"])
			},
		},
		{
			name: "Zero value time serialization",
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

				// Zero time should serialize as "0001-01-01T00:00:00Z"
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
			name: "Serializing user info",
			user: RadiusUser{
				ID:         1,
				NodeId:     100,
				ProfileId:  5,
				Realname:   "John Smith",
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
				assert.Equal(t, "John Smith", result["realname"])
				assert.Equal(t, "13800138000", result["mobile"])
				assert.Equal(t, "test001", result["username"])
				assert.Equal(t, "password123", result["password"])
				assert.Equal(t, "enabled", result["status"])

				// Validate the expire_time format is YYYY-MM-DD HH:MM:SS (2006-01-02 15:04:05)
				// Since the input time was UTC, but the output format doesn't include timezone,
				// we just check the string format.
				assert.Equal(t, "2025-12-31 23:59:59", result["expire_time"])

				// Validate the last_online format is YYYY-MM-DD HH:MM (2006-01-02 15:04)
				assert.Equal(t, "2025-11-08 14:30", result["last_online"])
			},
		},
		{
			name: "Zero value time serialization",
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

				// Check that zero times still produce formatted output
				assert.NotEmpty(t, result["expire_time"])
				assert.NotEmpty(t, result["last_online"])
			},
		},
		{
			name: "User with all fields",
			user: RadiusUser{
				ID:         3,
				NodeId:     200,
				ProfileId:  10,
				Realname:   "Jane Doe",
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
				Remark:     "Test account",
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
				assert.Equal(t, "Test account", result["remark"])
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
			name: "Deserializing user info",
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

				// Validate expire_time parsing (should be set to 23:59:59)
				// Since input string has no timezone, it is parsed as Local.
				// We construct expected time in Local timezone.
				expectedExpire := time.Date(2025, 12, 31, 23, 59, 59, 0, time.Local)
				assert.True(t, user.ExpireTime.Equal(expectedExpire),
					"ExpireTime should be %v but got %v", expectedExpire, user.ExpireTime)

				// Validate last_online parsing
				assert.Equal(t, 2025, user.LastOnline.Year())
				assert.Equal(t, time.November, user.LastOnline.Month())
				assert.Equal(t, 8, user.LastOnline.Day())
			},
		},
		{
			name: "Expire time with date only",
			json: `{
				"id": "2",
				"username": "test002",
				"expire_time": "2026-06-30",
				"last_online": "2025-11-07"
			}`,
			wantErr: false,
			check: func(t *testing.T, user *RadiusUser) {
				assert.Equal(t, "test002", user.Username)

				// Expiration time should be set to 23:59:59 on that day
				// dateparse seems to return UTC when parsing this format
				expectedExpire := time.Date(2026, 6, 30, 23, 59, 59, 0, time.UTC)
				assert.True(t, user.ExpireTime.Equal(expectedExpire),
					"ExpireTime should be %v but got %v", expectedExpire, user.ExpireTime)
			},
		},
		{
			name: "Formatted time string",
			json: `{
				"id": "3",
				"username": "test003",
				"expire_time": "2025-12-31 23:59:59",
				"last_online": "2025-11-08 14:30:00"
			}`,
			wantErr: false,
			check: func(t *testing.T, user *RadiusUser) {
				assert.Equal(t, "test003", user.Username)

				// Validate that the time values parsed correctly
				assert.Equal(t, 2025, user.ExpireTime.Year())
				assert.Equal(t, time.December, user.ExpireTime.Month())
				assert.Equal(t, 2025, user.LastOnline.Year())
				assert.Equal(t, time.November, user.LastOnline.Month())
			},
		},
		{
			name: "Full JSON payload",
			json: `{
				"id": "4",
				"node_id": "200",
				"profile_id": "10",
			"realname": "Jane Doe",
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
				"remark": "Complete test",
				"expire_time": "2027-12-31 23:59:59",
				"last_online": "2025-11-08 16:00:00"
			}`,
			wantErr: false,
			check: func(t *testing.T, user *RadiusUser) {
				assert.Equal(t, int64(4), user.ID)
				assert.Equal(t, "Jane Doe", user.Realname)
				assert.Equal(t, "13900139000", user.Mobile)
				assert.Equal(t, "test004", user.Username)
				assert.Equal(t, "pool2", user.AddrPool)
				assert.Equal(t, 3, user.ActiveNum)
				assert.Equal(t, 2048, user.UpRate)
				assert.Equal(t, 4096, user.DownRate)
				assert.Equal(t, "10.0.0.200", user.IpAddr)
				assert.Equal(t, "AA:BB:CC:DD:EE:FF", user.MacAddr)
				assert.Equal(t, "disabled", user.Status)
				assert.Equal(t, "Complete test", user.Remark)
			},
		},
		{
			name:    "Invalid JSON",
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
	// Test round-trip serialization and deserialization
	// Use Local time because JSON format doesn't include timezone, so Unmarshal assumes Local.
	original := RadiusUser{
		ID:         100,
		NodeId:     50,
		ProfileId:  10,
		Realname:   "Roundtrip Test",
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
		ExpireTime: time.Date(2025, 12, 31, 23, 59, 59, 0, time.Local),
		Status:     "enabled",
		Remark:     "Test remark",
		LastOnline: time.Date(2025, 11, 8, 10, 0, 0, 0, time.Local),
	}

	// Serialize
	data, err := json.Marshal(&original)
	require.NoError(t, err)
	t.Logf("Serialized JSON: %s", string(data))

	// Deserialize
	var decoded RadiusUser
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	// Validate key fields
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

	// Validate time fields (small discrepancies may appear due to formatting and parsing)
	assert.True(t, original.ExpireTime.Equal(decoded.ExpireTime),
		"ExpireTime mismatch: original=%v, decoded=%v", original.ExpireTime, decoded.ExpireTime)

	// LastOnline loses seconds because it is formatted with minute precision
	assert.Equal(t, original.LastOnline.Year(), decoded.LastOnline.Year())
	assert.Equal(t, original.LastOnline.Month(), decoded.LastOnline.Month())
	assert.Equal(t, original.LastOnline.Day(), decoded.LastOnline.Day())
	assert.Equal(t, original.LastOnline.Hour(), decoded.LastOnline.Hour())
	assert.Equal(t, original.LastOnline.Minute(), decoded.LastOnline.Minute())
}

func TestSysOprLog_MarshalUnmarshal_RoundTrip(t *testing.T) {
	// Test round-trip serialization and deserialization for operation logs
	original := SysOprLog{
		ID:        123,
		OprName:   "admin",
		OprIp:     "192.168.1.1",
		OptAction: "create_user",
		OptDesc:   "Created new user test001",
		OptTime:   time.Date(2025, 11, 8, 15, 30, 45, 0, time.UTC),
	}

	// Serialize
	data, err := json.Marshal(&original)
	require.NoError(t, err)
	t.Logf("Serialized JSON: %s", string(data))

	// Deserialize
	var decoded SysOprLog
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	// Validate fields
	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.OprName, decoded.OprName)
	assert.Equal(t, original.OprIp, decoded.OprIp)
	assert.Equal(t, original.OptAction, decoded.OptAction)
	assert.Equal(t, original.OptDesc, decoded.OptDesc)

	// Validate time (RFC3339 format should be preserved)
	assert.True(t, original.OptTime.Equal(decoded.OptTime),
		"OptTime mismatch: original=%v, decoded=%v", original.OptTime, decoded.OptTime)
}

func BenchmarkRadiusUser_MarshalJSON(b *testing.B) {
	user := RadiusUser{
		ID:         1,
		NodeId:     100,
		ProfileId:  5,
		Realname:   "Performance Test",
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
		OptDesc:   "User login",
		OptTime:   time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(&log)
	}
}
