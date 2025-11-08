package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSysConfig_TableName(t *testing.T) {
	model := SysConfig{}
	assert.Equal(t, "sys_config", model.TableName())
}

func TestSysOpr_TableName(t *testing.T) {
	model := SysOpr{}
	assert.Equal(t, "sys_opr", model.TableName())
}

func TestSysOprLog_TableName(t *testing.T) {
	model := SysOprLog{}
	assert.Equal(t, "sys_opr_log", model.TableName())
}

func TestNetNode_TableName(t *testing.T) {
	model := NetNode{}
	assert.Equal(t, "net_node", model.TableName())
}

func TestNetNas_TableName(t *testing.T) {
	model := NetNas{}
	assert.Equal(t, "net_nas", model.TableName())
}

func TestRadiusProfile_TableName(t *testing.T) {
	model := RadiusProfile{}
	assert.Equal(t, "radius_profile", model.TableName())
}

func TestRadiusUser_TableName(t *testing.T) {
	model := RadiusUser{}
	assert.Equal(t, "radius_user", model.TableName())
}

func TestRadiusOnline_TableName(t *testing.T) {
	model := RadiusOnline{}
	assert.Equal(t, "radius_online", model.TableName())
}

func TestRadiusAccounting_TableName(t *testing.T) {
	model := RadiusAccounting{}
	assert.Equal(t, "radius_accounting", model.TableName())
}

// TestAllModelsHaveTableName 验证所有在 Tables 列表中的模型都实现了 TableName 方法
func TestAllModelsHaveTableName(t *testing.T) {
	type tableNamer interface {
		TableName() string
	}

	for _, table := range Tables {
		t.Run("", func(t *testing.T) {
			_, ok := table.(tableNamer)
			assert.True(t, ok, "模型 %T 应该实现 TableName() 方法", table)
		})
	}
}

// TestTableNameUniqueness 验证所有表名都是唯一的
func TestTableNameUniqueness(t *testing.T) {
	type tableNamer interface {
		TableName() string
	}

	tableNames := make(map[string]bool)
	
	for _, table := range Tables {
		if tn, ok := table.(tableNamer); ok {
			name := tn.TableName()
			assert.False(t, tableNames[name], "表名 %s 重复", name)
			tableNames[name] = true
		}
	}

	// 验证所有表名都是蛇形命名（snake_case）
	expectedNames := map[string]bool{
		"sys_config":         true,
		"sys_opr":            true,
		"sys_opr_log":        true,
		"net_node":           true,
		"net_nas":            true,
		"radius_profile":     true,
		"radius_user":        true,
		"radius_online":      true,
		"radius_accounting":  true,
	}

	assert.Equal(t, len(expectedNames), len(tableNames), "表名数量应该匹配")
	
	for name := range tableNames {
		assert.True(t, expectedNames[name], "未预期的表名: %s", name)
	}
}
