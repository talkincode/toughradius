package gorm

import (
	"context"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/repository"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"gorm.io/gorm"
)

// GormSessionRepository GORM实现的会话仓储
type GormSessionRepository struct {
	db *gorm.DB
}

// NewGormSessionRepository 创建会话仓储实例
func NewGormSessionRepository(db *gorm.DB) repository.SessionRepository {
	return &GormSessionRepository{db: db}
}

func (r *GormSessionRepository) Create(ctx context.Context, session *domain.RadiusOnline) error {
	if session.ID == 0 {
		session.ID = common.UUIDint64()
	}
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *GormSessionRepository) Update(ctx context.Context, session *domain.RadiusOnline) error {
	param := map[string]interface{}{
		"acct_input_total":    session.AcctInputTotal,
		"acct_output_total":   session.AcctOutputTotal,
		"acct_input_packets":  session.AcctInputPackets,
		"acct_output_packets": session.AcctOutputPackets,
		"acct_session_time":   session.AcctSessionTime,
		"last_update":         time.Now(),
	}
	return r.db.WithContext(ctx).
		Model(&domain.RadiusOnline{}).
		Where("acct_session_id = ?", session.AcctSessionId).
		Updates(param).Error
}

func (r *GormSessionRepository) Delete(ctx context.Context, sessionId string) error {
	return r.db.WithContext(ctx).
		Where("acct_session_id = ?", sessionId).
		Delete(&domain.RadiusOnline{}).Error
}

func (r *GormSessionRepository) GetBySessionId(ctx context.Context, sessionId string) (*domain.RadiusOnline, error) {
	var session domain.RadiusOnline
	err := r.db.WithContext(ctx).
		Where("acct_session_id = ?", sessionId).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *GormSessionRepository) CountByUsername(ctx context.Context, username string) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.RadiusOnline{}).
		Where("username = ?", username).
		Count(&count).Error
	return int(count), err
}

func (r *GormSessionRepository) Exists(ctx context.Context, sessionId string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.RadiusOnline{}).
		Where("acct_session_id = ?", sessionId).
		Count(&count).Error
	return count > 0, err
}

func (r *GormSessionRepository) BatchDelete(ctx context.Context, ids []string) error {
	return r.db.WithContext(ctx).
		Where("id IN (?)", ids).
		Delete(&domain.RadiusOnline{}).Error
}

func (r *GormSessionRepository) BatchDeleteByNas(ctx context.Context, nasAddr, nasId string) error {
	if nasAddr != "" {
		if err := r.db.WithContext(ctx).Where("nas_addr = ?", nasAddr).Delete(&domain.RadiusOnline{}).Error; err != nil {
			return err
		}
	}
	if nasId != "" {
		if err := r.db.WithContext(ctx).Where("nas_id = ?", nasId).Delete(&domain.RadiusOnline{}).Error; err != nil {
			return err
		}
	}
	return nil
}
