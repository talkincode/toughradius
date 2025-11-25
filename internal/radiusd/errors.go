package radiusd

import "errors"

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
