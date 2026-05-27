package gorm

import (
	"context"
	"fmt"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/repository"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"gorm.io/gorm"
)

// GormAccountingRepository is the GORM implementation of the accounting repository
type GormAccountingRepository struct {
	db *gorm.DB
}

// NewGormAccountingRepository creates an accounting repository instance
func NewGormAccountingRepository(db *gorm.DB) repository.AccountingRepository {
	return &GormAccountingRepository{db: db}
}

func (r *GormAccountingRepository) Create(ctx context.Context, accounting *domain.RadiusAccounting) error {
	if accounting.ID == 0 {
		accounting.ID = common.UUIDint64()
	}
	return r.db.WithContext(ctx).Create(accounting).Error
}

func (r *GormAccountingRepository) UpdateStop(ctx context.Context, sessionId string, accounting *domain.RadiusAccounting) error {
	param := map[string]interface{}{
		"acct_stop_time":      time.Now(),
		"acct_input_total":    accounting.AcctInputTotal,
		"acct_output_total":   accounting.AcctOutputTotal,
		"acct_input_packets":  accounting.AcctInputPackets,
		"acct_output_packets": accounting.AcctOutputPackets,
		"acct_session_time":   accounting.AcctSessionTime,
	}

	result := r.db.WithContext(ctx).
		Model(&domain.RadiusAccounting{}).
		Where("acct_session_id = ?", sessionId).
		Updates(param)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no records found with acct_session_id = %v", sessionId)
	}

	return nil
}
