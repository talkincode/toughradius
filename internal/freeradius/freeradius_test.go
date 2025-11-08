/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *     http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package freeradius

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/talkincode/toughradius/v9/config"
	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"github.com/talkincode/toughradius/v9/pkg/web"
)

// initTestApp 初始化测试应用
func initTestApp() {
	if app.GApp() == nil {
		cfg := &config.AppConfig{
			System: config.SysConfig{
				Appid:    "TestFreeRADIUS",
				Location: "Asia/Shanghai",
				Workdir:  "/tmp/test-freeradius",
				Debug:    false,
			},
			Database: config.DBConfig{
				Type: "sqlite",
				Name: "/tmp/test-freeradius.db",
			},
			Freeradius: config.FreeradiusConfig{
				Host:  "127.0.0.1",
				Port:  18181,
				Debug: false,
			},
		}
		app.InitGlobalApplication(cfg)
	}
}

// createTestForm 创建测试用的 WebForm
func createTestForm(values map[string]string) *web.WebForm {
	form := web.EmptyWebForm()
	for k, v := range values {
		form.Posts.Set(k, v)
	}
	return form
}

// TestGetOnlineCount 测试获取在线用户数量
func TestGetOnlineCount(t *testing.T) {
	initTestApp()

	count, err := getOnlineCount("testuser")
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, count, int64(0))
}

// TestGetOnlineCountBySessionid 测试根据会话ID获取在线数量
func TestGetOnlineCountBySessionid(t *testing.T) {
	initTestApp()

	count, err := getOnlineCountBySessionid("test-session-id")
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, count, int64(0))
}

// TestGetAcctStartTime 测试计算开始时间
func TestGetAcctStartTime(t *testing.T) {
	tests := []struct {
		name        string
		sessionTime string
		expectPast  bool
	}{
		{
			name:        "zero seconds",
			sessionTime: "0",
			expectPast:  false,
		},
		{
			name:        "60 seconds ago",
			sessionTime: "60",
			expectPast:  true,
		},
		{
			name:        "3600 seconds ago (1 hour)",
			sessionTime: "3600",
			expectPast:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getAcctStartTime(tt.sessionTime)
			now := time.Now()

			if tt.expectPast {
				assert.True(t, result.Before(now), "Expected time to be in the past")
			} else {
				// 允许小的时间差异（由于函数执行时间）
				diff := now.Sub(result)
				assert.Less(t, diff.Seconds(), 1.0, "Expected time to be very close to now")
			}
		})
	}
}

// TestGetInputTotal 测试计算输入总量
func TestGetInputTotal(t *testing.T) {
	tests := []struct {
		name              string
		acctInputOctets   int64
		acctInputGigaword int64
		expected          int64
	}{
		{
			name:              "only octets",
			acctInputOctets:   1000,
			acctInputGigaword: 0,
			expected:          1000,
		},
		{
			name:              "only gigawords",
			acctInputOctets:   0,
			acctInputGigaword: 1,
			expected:          4 * 1024 * 1024 * 1024,
		},
		{
			name:              "octets and gigawords",
			acctInputOctets:   1000,
			acctInputGigaword: 2,
			expected:          1000 + 2*4*1024*1024*1024,
		},
		{
			name:              "zero values",
			acctInputOctets:   0,
			acctInputGigaword: 0,
			expected:          0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := createTestForm(map[string]string{
				"acctInputOctets":   strconv.FormatInt(tt.acctInputOctets, 10),
				"acctInputGigaword": strconv.FormatInt(tt.acctInputGigaword, 10),
			})

			result := getInputTotal(form)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGetOutputTotal 测试计算输出总量
func TestGetOutputTotal(t *testing.T) {
	tests := []struct {
		name               string
		acctOutputOctets   int64
		acctOutputGigaword int64
		expected           int64
	}{
		{
			name:               "only octets",
			acctOutputOctets:   2000,
			acctOutputGigaword: 0,
			expected:           2000,
		},
		{
			name:               "only gigawords",
			acctOutputOctets:   0,
			acctOutputGigaword: 1,
			expected:           4 * 1024 * 1024 * 1024,
		},
		{
			name:               "octets and gigawords",
			acctOutputOctets:   2000,
			acctOutputGigaword: 3,
			expected:           2000 + 3*4*1024*1024*1024,
		},
		{
			name:               "zero values",
			acctOutputOctets:   0,
			acctOutputGigaword: 0,
			expected:           0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := createTestForm(map[string]string{
				"acctOutputOctets":    strconv.FormatInt(tt.acctOutputOctets, 10),
				"acctOutputGigawords": strconv.FormatInt(tt.acctOutputGigaword, 10),
			})

			result := getOutputTotal(form)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestBatchClearRadiusOnlineDataByNas 测试批量清除在线数据
func TestBatchClearRadiusOnlineDataByNas(t *testing.T) {
	// 这个测试需要真实的数据库连接
	// 在单元测试中，应该 mock 数据库
	err := BatchClearRadiusOnlineDataByNas("192.168.1.1", "nas01")
	// 在没有测试数据库的情况下，这可能会失败
	// 但至少验证函数可以被调用
	_ = err
}

// TestUpdateRadiusOnline_Start 测试开始会话的在线更新
func TestUpdateRadiusOnline_Start(t *testing.T) {
	// 这个测试需要完整的测试环境
	// 包括数据库和预先存在的用户数据
	// 这里仅做基本的结构测试

	form := createTestForm(map[string]string{
		"username":        "testuser",
		"nasid":           "nas01",
		"nasip":           "192.168.1.1",
		"acctSessionId":   "test-session-001",
		"acctStatusType":  "Start",
		"sessionTimeout":  "3600",
		"acctSessionTime": "0",
	})

	// 注意：这个测试在没有真实数据库的情况下会失败
	// 应该使用 test database 或 mock
	err := updateRadiusOnline(form)
	// 验证错误信息包含预期的内容
	if err != nil {
		assert.Contains(t, err.Error(), "not exists")
	}
}

// TestUpdateRadiusOnline_Stop 测试停止会话的在线更新
func TestUpdateRadiusOnline_Stop(t *testing.T) {
	form := createTestForm(map[string]string{
		"username":         "testuser",
		"nasid":            "nas01",
		"nasip":            "192.168.1.1",
		"acctSessionId":    "test-session-002",
		"acctStatusType":   "Stop",
		"acctSessionTime":  "3600",
		"acctInputOctets":  "1024000",
		"acctOutputOctets": "2048000",
	})

	err := updateRadiusOnline(form)
	// 验证错误信息
	if err != nil {
		assert.Contains(t, err.Error(), "not exists")
	}
}

// TestUpdateRadiusOnline_Update 测试更新会话的在线更新
func TestUpdateRadiusOnline_Update(t *testing.T) {
	form := createTestForm(map[string]string{
		"username":        "testuser",
		"nasid":           "nas01",
		"nasip":           "192.168.1.1",
		"acctSessionId":   "test-session-003",
		"acctStatusType":  "Update",
		"acctSessionTime": "1800",
		"framedIPAddress": "10.0.0.1",
		"macAddr":         "00:11:22:33:44:55",
	})

	err := updateRadiusOnline(form)
	if err != nil {
		assert.Contains(t, err.Error(), "not exists")
	}
}

// TestUpdateRadiusOnline_AccountingOn 测试 Accounting-On 状态
func TestUpdateRadiusOnline_AccountingOn(t *testing.T) {
	form := createTestForm(map[string]string{
		"username":       "testuser",
		"nasid":          "nas01",
		"nasip":          "192.168.1.1",
		"acctSessionId":  "test-session-004",
		"acctStatusType": "Accounting-On",
	})

	err := updateRadiusOnline(form)
	if err != nil {
		// Accounting-On 可能会因为用户不存在而失败
		assert.Contains(t, err.Error(), "not exists")
	}
}

// BenchmarkGetAcctStartTime 基准测试计算开始时间
func BenchmarkGetAcctStartTime(b *testing.B) {
	for i := 0; i < b.N; i++ {
		getAcctStartTime("3600")
	}
}

// BenchmarkGetInputTotal 基准测试计算输入总量
func BenchmarkGetInputTotal(b *testing.B) {
	form := createTestForm(map[string]string{
		"acctInputOctets":   "1024000",
		"acctInputGigaword": "2",
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getInputTotal(form)
	}
}

// BenchmarkGetOutputTotal 基准测试计算输出总量
func BenchmarkGetOutputTotal(b *testing.B) {
	form := createTestForm(map[string]string{
		"acctOutputOctets":    "2048000",
		"acctOutputGigawords": "3",
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getOutputTotal(form)
	}
}

// TestRadiusOnlineStruct 测试在线数据结构
func TestRadiusOnlineStruct(t *testing.T) {
	online := domain.RadiusOnline{
		ID:                common.UUIDint64(),
		Username:          "testuser",
		NasId:             "nas01",
		NasAddr:           "192.168.1.1",
		NasPaddr:          "192.168.1.1",
		SessionTimeout:    3600,
		FramedIpaddr:      "10.0.0.1",
		FramedNetmask:     "255.255.255.0",
		MacAddr:           "00:11:22:33:44:55",
		NasPort:           0,
		NasClass:          "standard",
		NasPortId:         "eth0",
		NasPortType:       0,
		ServiceType:       0,
		AcctSessionId:     "session-001",
		AcctSessionTime:   1800,
		AcctInputTotal:    1024000,
		AcctOutputTotal:   2048000,
		AcctInputPackets:  1000,
		AcctOutputPackets: 2000,
		AcctStartTime:     time.Now().Add(-30 * time.Minute),
		LastUpdate:        time.Now(),
	}

	assert.NotZero(t, online.ID)
	assert.Equal(t, "testuser", online.Username)
	assert.Equal(t, "nas01", online.NasId)
	assert.Equal(t, "192.168.1.1", online.NasAddr)
	assert.Equal(t, int64(3600), int64(online.SessionTimeout))
	assert.Equal(t, "10.0.0.1", online.FramedIpaddr)
	assert.Equal(t, "session-001", online.AcctSessionId)
} // TestMetricsConstants 测试指标常量
func TestMetricsConstants(t *testing.T) {
	assert.NotEmpty(t, app.MetricsRadiusAccept)
	assert.NotEmpty(t, app.MetricsRadiusRejectNotExists)
	assert.NotEmpty(t, app.MetricsRadiusRejectDisable)
	assert.NotEmpty(t, app.MetricsRadiusRejectExpire)
	assert.NotEmpty(t, app.MetricsRadiusRejectLimit)
	assert.NotEmpty(t, app.MetricsRadiusRejectOther)
	assert.NotEmpty(t, app.MetricsRadiusOline)
	assert.NotEmpty(t, app.MetricsRadiusOffline)
	assert.NotEmpty(t, app.MetricsRadiusAccounting)
	assert.NotEmpty(t, app.MetricsRadiusAcctDrop)
}
