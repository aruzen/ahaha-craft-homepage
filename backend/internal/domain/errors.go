package domain

import "errors"

var (
	ErrEmptyName           = errors.New("domain: empty name")
	ErrInvalidChoice       = errors.New("domain: invalid choice")
	ErrInvalidRange        = errors.New("domain: invalid record range")
	ErrInvalidToken        = errors.New("domain: invalid token")
	ErrExpiredToken        = errors.New("domain: expired token")
	ErrInvalidCredential   = errors.New("domain: invalid credential")
	ErrInvalidPassword     = errors.New("domain: invalid password")
	ErrInvalidSessionToken = errors.New("domain: invalid login session token")
	ErrInvalidLoginSession = errors.New("domain: invalid login session")
	ErrInvalidSessionData  = errors.New("domain: invalid session data")
	ErrInvalidEmail        = errors.New("domain: invalid email")
	ErrInvalidPasswordHash = errors.New("domain: invalid password hash")
	ErrInvalidUserRole     = errors.New("domain: invalid user role")
	ErrInvalidUser         = errors.New("domain: invalid user")
	ErrDuplicateUsername   = errors.New("domain: duplicate username")
	ErrDuplicateEmail      = errors.New("domain: duplicate email")
	ErrInvalidAPIError     = errors.New("domain: invalid api error")
)
