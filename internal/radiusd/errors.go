package radiusd

import (
	"errors"

	"github.com/talkincode/toughradius/v9/internal/app"
)

type AuthError struct {
	Type string
	Err  error
}

func NewAuthError(errType string, err string) *AuthError {
	return &AuthError{Type: errType, Err: errors.New(err)}
}

func (e *AuthError) Error() string {
	return e.Err.Error()
}

func (s *RadiusService) CheckRadAuthError(username, nasip string, err error) {
	if err != nil {
		rjuser := s.RejectCache.GetItem(username)
		if rjuser == nil {
			s.RejectCache.SetItem(username)
		} else {
			if rjuser.IsOver(RadiusRejectDelayTimes) {
				panic(NewAuthError(app.MetricsRadiusRejectLimit, err.Error()))
			}
			rjuser.Incr()
		}
		panic(err)
	}
}
