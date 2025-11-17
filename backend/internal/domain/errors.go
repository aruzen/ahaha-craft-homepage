package domain

import "errors"

var (
	ErrEmptyName           = errors.New("domain: empty name")
	ErrInvalidChoice       = errors.New("domain: invalid choice")
	ErrInvalidRange        = errors.New("domain: invalid record range")
	ErrInvalidToken        = errors.New("domain: invalid token")
	ErrInvalidCredential   = errors.New("domain: invalid credential")
	ErrInvalidSessionToken = errors.New("domain: invalid login session token")
	ErrInvalidLoginSession = errors.New("domain: invalid login session")
)
