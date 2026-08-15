package app

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	defaultSuperUsername = "admin"

	// WellKnownBootstrapPassword is the historical first-start password.
	// It must never be persisted or accepted (CWE-1392 / GHSA-2gwm-6gf5-8699).
	WellKnownBootstrapPassword = "toughradius"

	// AdminPasswordEnv is the optional bootstrap password for a fresh super
	// admin. It is read only when creating or rotating the initial account.
	AdminPasswordEnv = "TOUGHRADIUS_ADMIN_PASSWORD"

	bootstrapPasswordLen = 20
	bootstrapLetters     = "abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ"
	bootstrapDigits      = "23456789"
	bootstrapAlphabet    = bootstrapLetters + bootstrapDigits
)

// IsWellKnownBootstrapPassword reports whether password is the historical
// built-in credential that must not be stored or accepted.
func IsWellKnownBootstrapPassword(password string) bool {
	return password == WellKnownBootstrapPassword
}

func (a *Application) checkSuper() {
	var operator domain.SysOpr
	err := a.gormDB.Where("username = ?", defaultSuperUsername).First(&operator).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		a.createBootstrapSuper()
		return
	case err != nil:
		zap.L().Error("failed to query super admin", zap.Error(err))
		return
	}

	a.rotateInsecureSuperPassword(&operator)
}

func (a *Application) createBootstrapSuper() {
	var supers int64
	if err := a.gormDB.Model(&domain.SysOpr{}).Where("level = ?", "super").Count(&supers).Error; err != nil {
		zap.L().Error("failed to count super operators", zap.Error(err))
		return
	}
	if supers > 0 {
		zap.L().Info("skipped default admin bootstrap because a super operator already exists")
		return
	}

	password, source, err := resolveBootstrapAdminPassword()
	if err != nil {
		a.failInsecureBootstrap("failed to generate bootstrap super admin password", zap.Error(err))
		return
	}
	hashedPassword, err := common.HashPassword(password)
	if err != nil {
		a.failInsecureBootstrap("failed to hash bootstrap super admin password", zap.Error(err))
		return
	}

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
		return
	}

	fields := []zap.Field{
		zap.String("username", defaultSuperUsername),
		zap.String("source", source),
		zap.String("action", "store this password now; it will not be shown again"),
	}
	if source == "generated" {
		fields = append(fields, zap.String("password", password))
	}
	zap.L().Info("initialized bootstrap super admin account", fields...)
}

func (a *Application) rotateInsecureSuperPassword(operator *domain.SysOpr) {
	if operator == nil {
		return
	}
	if strings.TrimSpace(operator.Password) != "" &&
		!common.VerifyPassword(WellKnownBootstrapPassword, operator.Password) {
		return
	}

	password, source, err := resolveBootstrapAdminPassword()
	if err != nil {
		a.failInsecureBootstrap("failed to generate replacement for insecure admin password", zap.Error(err))
		return
	}
	hashedPassword, err := common.HashPassword(password)
	if err != nil {
		a.failInsecureBootstrap("failed to hash replacement admin password", zap.Error(err))
		return
	}

	if err := a.gormDB.Model(&domain.SysOpr{}).Where("id = ?", operator.ID).Updates(map[string]interface{}{
		"password":   hashedPassword,
		"updated_at": time.Now(),
	}).Error; err != nil {
		a.failInsecureBootstrap("failed to rotate insecure admin password", zap.Error(err))
		return
	}

	fields := []zap.Field{
		zap.String("username", operator.Username),
		zap.String("source", source),
		zap.String("action", "the historical default password is no longer valid"),
	}
	if source == "generated" {
		fields = append(fields, zap.String("password", password))
	}
	zap.L().Warn("rotated insecure super admin password", fields...)
}

func (a *Application) failInsecureBootstrap(msg string, fields ...zap.Field) {
	if isProductionRuntime(a.appConfig) {
		zap.L().Fatal(msg, fields...)
		return
	}
	zap.L().Error(msg, fields...)
}

func resolveBootstrapAdminPassword() (password, source string, err error) {
	if env := strings.TrimSpace(os.Getenv(AdminPasswordEnv)); env != "" {
		if IsWellKnownBootstrapPassword(env) {
			zap.L().Warn("TOUGHRADIUS_ADMIN_PASSWORD uses the historical default and was ignored")
		} else {
			return env, "env", nil
		}
	}
	generated, err := generateBootstrapAdminPassword()
	if err != nil {
		return "", "", err
	}
	return generated, "generated", nil
}

func generateBootstrapAdminPassword() (string, error) {
	buf := make([]byte, bootstrapPasswordLen)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	out := make([]byte, bootstrapPasswordLen)
	for i, b := range buf {
		out[i] = bootstrapAlphabet[int(b)%len(bootstrapAlphabet)]
	}
	// Guarantee the letter+digit mix required by operator password policy.
	out[0] = bootstrapLetters[int(buf[0])%len(bootstrapLetters)]
	out[1] = bootstrapDigits[int(buf[1])%len(bootstrapDigits)]
	return string(out), nil
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
