package checkers

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
	radiusErrors "github.com/talkincode/toughradius/v9/internal/radiusd/errors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
)

// mockSessionRepository simulates a SessionRepository
type mockSessionRepository struct {
	countByUsername func(ctx context.Context, username string) (int, error)
}

func (m *mockSessionRepository) CountByUsername(ctx context.Context, username string) (int, error) {
	if m.countByUsername != nil {
		return m.countByUsername(ctx, username)
	}
	return 0, nil
}

// Implement other interface methods (unused in tests)
func (m *mockSessionRepository) Create(ctx context.Context, session *domain.RadiusOnline) error {
	return nil
}

func (m *mockSessionRepository) Update(ctx context.Context, session *domain.RadiusOnline) error {
	return nil
}

func (m *mockSessionRepository) Delete(ctx context.Context, sessionId string) error {
	return nil
}

func (m *mockSessionRepository) GetBySessionId(ctx context.Context, sessionId string) (*domain.RadiusOnline, error) {
	return nil, nil
}

func (m *mockSessionRepository) Exists(ctx context.Context, sessionId string) (bool, error) {
	return false, nil
}

func (m *mockSessionRepository) BatchDelete(ctx context.Context, ids []string) error {
	return nil
}

func (m *mockSessionRepository) BatchDeleteByNas(ctx context.Context, nasAddr, nasId string) error {
	return nil
}

func TestOnlineCountChecker_Name(t *testing.T) {
	checker := NewOnlineCountChecker(&mockSessionRepository{})
	assert.Equal(t, "online_count", checker.Name())
}

func TestOnlineCountChecker_Order(t *testing.T) {
	checker := NewOnlineCountChecker(&mockSessionRepository{})
	assert.Equal(t, 30, checker.Order())
}

func TestOnlineCountChecker_Check(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		activeNum     int
		onlineCount   int
		countError    error
		expectError   bool
		errorContains string
	}{
		{
			name:        "no limit (activeNum = 0)",
			activeNum:   0,
			onlineCount: 100,
			expectError: false,
		},
		{
			name:        "under limit",
			activeNum:   5,
			onlineCount: 3,
			expectError: false,
		},
		{
			name:        "at limit",
			activeNum:   5,
			onlineCount: 5,
			expectError: true,
		},
		{
			name:          "over limit",
			activeNum:     5,
			onlineCount:   10,
			expectError:   true,
			errorContains: "exceeded",
		},
		{
			name:        "repository error",
			activeNum:   5,
			onlineCount: 0,
			countError:  errors.New("database error"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockSessionRepository{
				countByUsername: func(ctx context.Context, username string) (int, error) {
					if tt.countError != nil {
						return 0, tt.countError
					}
					return tt.onlineCount, nil
				},
			}

			checker := NewOnlineCountChecker(mockRepo)

			user := &domain.RadiusUser{
				Username:  "testuser",
				ActiveNum: tt.activeNum,
			}

			authCtx := &auth.AuthContext{
				User: user,
			}

			err := checker.Check(ctx, authCtx)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				// Check whether it's an authentication error (excluding repository errors)
				if tt.countError == nil {
					authErr, ok := radiusErrors.GetAuthError(err)
					assert.True(t, ok)
					assert.NotNil(t, authErr)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
