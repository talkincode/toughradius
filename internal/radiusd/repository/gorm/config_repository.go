package gorm

import (
	"context"
	"strconv"

	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/radiusd/repository"
	"gorm.io/gorm"
)

// GormConfigRepository GORM实现的配置仓储
type GormConfigRepository struct {
	db *gorm.DB
}

// NewGormConfigRepository 创建配置仓储实例
func NewGormConfigRepository(db *gorm.DB) repository.ConfigRepository {
	return &GormConfigRepository{db: db}
}

func (r *GormConfigRepository) GetString(ctx context.Context, category, key string) string {
	return app.GApp().GetSettingsStringValue(category, key)
}

func (r *GormConfigRepository) GetInt(ctx context.Context, category, key string, defaultVal int64) int64 {
	cval := app.GApp().GetSettingsStringValue(category, key)
	ival, err := strconv.ParseInt(cval, 10, 64)
	if err != nil {
		return defaultVal
	}
	return ival
}

func (r *GormConfigRepository) GetBool(ctx context.Context, category, key string) bool {
	val := app.GApp().GetSettingsStringValue(category, key)
	return val == "enabled" || val == "true" || val == "1"
}
