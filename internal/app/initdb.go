package app

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"go.uber.org/zap"
)

func (a *Application) checkSuper() {
	var count int64
	a.gormDB.Model(&domain.SysOpr{}).Where("username='admin' and level = ?", "super").Count(&count)
	if count == 0 {
		a.gormDB.Create(&domain.SysOpr{
			ID:        common.UUIDint64(),
			Realname:  "administrator",
			Mobile:    "0000",
			Email:     "N/A",
			Username:  "admin",
			Password:  common.Sha256HashWithSalt("toughradius", common.SecretSalt),
			Level:     "super",
			Status:    "enabled",
			Remark:    "super",
			LastLogin: time.Now(),
		})
	}
}

func (a *Application) checkSettings() {
	// 从嵌入的 JSON 文件加载配置定义
	var schemasData ConfigSchemasJSON
	if err := json.Unmarshal(configSchemasData, &schemasData); err != nil {
		zap.L().Error("failed to load config schemas from JSON", zap.Error(err))
		return
	}

	// 遍历所有配置定义，检查并初始化缺失的配置
	for sortid, schema := range schemasData.Schemas {
		// 解析 key: "category.name" -> category, name
		parts := strings.SplitN(schema.Key, ".", 2)
		if len(parts) != 2 {
			zap.L().Warn("invalid config key format", zap.String("key", schema.Key))
			continue
		}

		category := parts[0]
		name := parts[1]

		// 检查配置是否存在
		var count int64
		a.gormDB.Model(&domain.SysConfig{}).
			Where("type = ? and name = ?", category, name).
			Count(&count)

		// 如果配置不存在，创建默认配置
		if count == 0 {
			a.gormDB.Create(&domain.SysConfig{
				ID:     0,
				Sort:   sortid,
				Type:   category,
				Name:   name,
				Value:  schema.Default,
				Remark: schema.Description,
			})
			zap.L().Info("initialized config",
				zap.String("key", schema.Key),
				zap.String("default", schema.Default))
		}
	}
}
