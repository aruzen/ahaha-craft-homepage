package domain

import (
	"golang.org/x/crypto/bcrypt"
	"strings"
)

type AdminCredential struct {
	name           Name
	password       string
	hashedPassword string
}

func NewAdminCredential(name Name, rawPassword string) (AdminCredential, error) {
	trimmed := strings.TrimSpace(rawPassword)
	if trimmed == "" {
		return AdminCredential{}, ErrInvalidCredential
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(trimmed), bcrypt.DefaultCost)
	if err != nil {
		return AdminCredential{}, ErrInvalidCredential
	}

	return AdminCredential{
		name:           name,
		password:       trimmed,
		hashedPassword: string(hashed),
	}, nil
}

func (c AdminCredential) Name() Name {
	return c.name
}

func (c AdminCredential) Password() string {
	return c.password
}

func (c AdminCredential) HashedPassword() string {
	return c.hashedPassword
}
