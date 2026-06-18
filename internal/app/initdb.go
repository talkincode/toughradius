package app

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	defaultSuperUsername = "admin"
	defaultSuperPassword = "toughradius"
)

func (a *Application) checkSuper() {
	hashedPassword, err := common.HashPassword(defaultSuperPassword)
	if err != nil {
		zap.L().Error("failed to hash default super admin password", zap.Error(err))
		return
	}

	var operator domain.SysOpr
	err = a.gormDB.Where("username = ?", defaultSuperUsername).First(&operator).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		if err := a.gormDB.Create(&domain.SysOpr{
			ID:        common.UUIDint64(),
			Realname:  "administrator",
			Mobile:    "0000",
			Email:     "N/A",
			Username:  defaultSuperUsername,
			Password:  hashedPassword,
			Level:     "super",
			Status:    common.ENABLED,
			Remark:    "super",
			LastLogin: time.Now(),
		}).Error; err != nil {
			zap.L().Error("failed to create default super admin", zap.Error(err))
		} else {
			zap.L().Info("initialized default super admin account", zap.String("username", defaultSuperUsername))
			warnDefaultSuperPassword("created")
		}
		return
	case err != nil:
		zap.L().Error("failed to query super admin", zap.Error(err))
		return
	}

	resetPassword := strings.TrimSpace(operator.Password) == ""
	resetLevel := !strings.EqualFold(operator.Level, "super")
	resetStatus := !strings.EqualFold(operator.Status, common.ENABLED)

	if !resetPassword && !resetLevel && !resetStatus {
		if common.VerifyPassword(defaultSuperPassword, operator.Password) {
			warnDefaultSuperPassword("existing")
		}
		return
	}

	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}
	if resetPassword {
		updates["password"] = hashedPassword
	}
	if resetLevel {
		updates["level"] = "super"
	}
	if resetStatus {
		updates["status"] = common.ENABLED
	}

	if err := a.gormDB.Model(&domain.SysOpr{}).Where("id = ?", operator.ID).Updates(updates).Error; err != nil {
		zap.L().Error("failed to repair super admin account", zap.Error(err))
		return
	}

	zap.L().Warn("repaired default super admin account",
		zap.String("username", defaultSuperUsername),
		zap.Bool("passwordReset", resetPassword),
		zap.Bool("levelReset", resetLevel),
		zap.Bool("statusEnabled", resetStatus))
	if resetPassword {
		warnDefaultSuperPassword("repaired")
	}
}

func warnDefaultSuperPassword(source string) {
	zap.L().Warn("default super admin account uses the built-in password",
		zap.String("username", defaultSuperUsername),
		zap.String("source", source),
		zap.String("action", "change the admin password before exposing the admin API"))
}

func (a *Application) checkSettings() {
	// Load configuration definitions from the embedded JSON file
	var schemasData ConfigSchemasJSON
	if err := json.Unmarshal(configSchemasData, &schemasData); err != nil {
		zap.L().Error("failed to load config schemas from JSON", zap.Error(err))
		return
	}

	// Iterate over all configuration definitions, checking and initializing missing entries
	for sortid, schema := range schemasData.Schemas {
		// Parse key: "category.name" -> category, name
		parts := strings.SplitN(schema.Key, ".", 2)
		if len(parts) != 2 {
			zap.L().Warn("invalid config key format", zap.String("key", schema.Key))
			continue
		}

		category := parts[0]
		name := parts[1]

		// Check whether the configuration already exists
		var count int64
		a.gormDB.Model(&domain.SysConfig{}).
			Where("type = ? and name = ?", category, name).
			Count(&count)

		// e.g., if the configuration does not exist, create the default configuration
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
