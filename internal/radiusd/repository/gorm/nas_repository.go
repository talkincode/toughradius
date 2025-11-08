package gorm

import (
	"context"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/repository"
	"gorm.io/gorm"
)

// GormNasRepository GORM实现的NAS仓储
type GormNasRepository struct {
	db *gorm.DB
}

// NewGormNasRepository 创建NAS仓储实例
func NewGormNasRepository(db *gorm.DB) repository.NasRepository {
	return &GormNasRepository{db: db}
}

func (r *GormNasRepository) GetByIP(ctx context.Context, ip string) (*domain.NetNas, error) {
	var nas domain.NetNas
	err := r.db.WithContext(ctx).Where("ipaddr = ?", ip).First(&nas).Error
	if err != nil {
		return nil, err
	}
	return &nas, nil
}

func (r *GormNasRepository) GetByIdentifier(ctx context.Context, identifier string) (*domain.NetNas, error) {
	var nas domain.NetNas
	err := r.db.WithContext(ctx).Where("identifier = ?", identifier).First(&nas).Error
	if err != nil {
		return nil, err
	}
	return &nas, nil
}

func (r *GormNasRepository) GetByIPOrIdentifier(ctx context.Context, ip, identifier string) (*domain.NetNas, error) {
	var nas domain.NetNas
	err := r.db.WithContext(ctx).
		Where("ipaddr = ? OR identifier = ?", ip, identifier).
		First(&nas).Error
	if err != nil {
		return nil, err
	}
	return &nas, nil
}
