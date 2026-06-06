package gorm

import (
	"context"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	cachepkg "github.com/talkincode/toughradius/v9/internal/radiusd/cache"
	"github.com/talkincode/toughradius/v9/internal/radiusd/repository"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GormSessionRepository is the GORM implementation of the session repository
type GormSessionRepository struct {
	db         *gorm.DB
	countCache *cachepkg.TTLCache[int]
}

// NewGormSessionRepository creates a session repository instance
func NewGormSessionRepository(db *gorm.DB) repository.SessionRepository {
	return &GormSessionRepository{
		db:         db,
		countCache: cachepkg.NewTTLCache[int](2*time.Second, 4096),
	}
}

// Create inserts the online session idempotently. The insert relies on the
// unique index on acct_session_id together with an ON CONFLICT DO NOTHING
// clause, so a retransmitted Accounting-Start can never produce a duplicate
// online row. It returns created=false when the row already existed.
func (r *GormSessionRepository) Create(ctx context.Context, session *domain.RadiusOnline) (bool, error) {
	if session.ID == 0 {
		session.ID = common.UUIDint64()
	}
	result := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "acct_session_id"}},
			DoNothing: true,
		}).
		Create(session)
	if result.Error != nil {
		return false, result.Error
	}
	created := result.RowsAffected > 0
	if created {
		r.invalidate(session.Username)
	}
	return created, nil
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
	username := r.lookupUsernameBySession(ctx, sessionId)
	err := r.db.WithContext(ctx).
		Where("acct_session_id = ?", sessionId).
		Delete(&domain.RadiusOnline{}).Error
	if err == nil {
		r.invalidate(username)
	}
	return err
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
	if username != "" {
		if cached, ok := r.countCache.Get(username); ok {
			return cached, nil
		}
	}
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.RadiusOnline{}).
		Where("username = ?", username).
		Count(&count).Error
	if err == nil && username != "" {
		r.countCache.Set(username, int(count))
	}
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
	err := r.db.WithContext(ctx).
		Where("id IN (?)", ids).
		Delete(&domain.RadiusOnline{}).Error
	if err == nil {
		r.countCache.Clear()
	}
	return err
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
	r.countCache.Clear()
	return nil
}

func (r *GormSessionRepository) invalidate(username string) {
	if username == "" {
		return
	}
	r.countCache.Delete(username)
}

func (r *GormSessionRepository) lookupUsernameBySession(ctx context.Context, sessionId string) string {
	if sessionId == "" {
		return ""
	}
	var result struct {
		Username string
	}
	if err := r.db.WithContext(ctx).
		Model(&domain.RadiusOnline{}).
		Select("username").
		Where("acct_session_id = ?", sessionId).
		Limit(1).
		Take(&result).Error; err != nil {
		return ""
	}
	return result.Username
}
