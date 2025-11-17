package domain

import (
	"golang.org/x/crypto/bcrypt"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Email は RFC に沿って検証されたメールアドレス。
type Email struct {
	value string
}

func NewEmail(value string) (Email, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return Email{}, ErrInvalidEmail
	}

	addr, err := mail.ParseAddress(trimmed)
	if err != nil {
		return Email{}, ErrInvalidEmail
	}

	return Email{value: addr.Address}, nil
}

func (e Email) String() string {
	return e.value
}

func (e Email) isZero() bool {
	return e.value == ""
}

// HashedPassword は bcrypt などでハッシュ化済みのパスワードを保持する。
type HashedPassword struct {
	value string
}

func NewHashedPassword(value string) (HashedPassword, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return HashedPassword{}, ErrInvalidPasswordHash
	}

	return HashedPassword{value: trimmed}, nil
}

func (p HashedPassword) String() string {
	return p.value
}

func (p HashedPassword) isZero() bool {
	return p.value == ""
}

func (p HashedPassword) Verify(plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(p.value), []byte(plain))
}

// UserRole は users.role の列挙を表す。
type UserRole string

const (
	UserRoleUser  UserRole = "user"
	UserRoleAdmin UserRole = "admin"
)

func NewUserRole(value string) (UserRole, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", ErrInvalidUserRole
	}

	role := UserRole(strings.ToLower(trimmed))
	if !role.valid() {
		return "", ErrInvalidUserRole
	}

	return role, nil
}

func (r UserRole) String() string {
	return string(r)
}

func (r UserRole) valid() bool {
	switch r {
	case UserRoleUser, UserRoleAdmin:
		return true
	default:
		return false
	}
}

func (r UserRole) isZero() bool {
	return r == ""
}

// User は users テーブルの行に対応するドメインエンティティ。
type User struct {
	id             uuid.UUID
	username       Name
	email          Email
	hashedPassword HashedPassword
	role           UserRole
	createdAt      time.Time
	updatedAt      time.Time
}

// NewUser は新規登録時に UUID とタイムスタンプを生成する。
func NewUser(username Name, email Email, hashedPassword HashedPassword, role UserRole, now time.Time) (User, error) {
	issued := now.UTC()
	if issued.IsZero() {
		return User{}, ErrInvalidUser
	}

	return buildUser(uuid.New(), username, email, hashedPassword, role, issued, issued)
}

// NewUserFromPersistence は永続化済みデータから再構築する。
func NewUserFromPersistence(id uuid.UUID, username Name, email Email, hashedPassword HashedPassword, role UserRole, createdAt, updatedAt time.Time) (User, error) {
	return buildUser(id, username, email, hashedPassword, role, createdAt, updatedAt)
}

func (u User) ID() uuid.UUID {
	return u.id
}

func (u User) Username() Name {
	return u.username
}

func (u User) Email() Email {
	return u.email
}

func (u User) HashedPassword() HashedPassword {
	return u.hashedPassword
}

func (u User) Role() UserRole {
	return u.role
}

func (u User) CreatedAt() time.Time {
	return u.createdAt
}

func (u User) UpdatedAt() time.Time {
	return u.updatedAt
}

func buildUser(id uuid.UUID, username Name, email Email, hashedPassword HashedPassword, role UserRole, createdAt, updatedAt time.Time) (User, error) {
	if id == uuid.Nil || username.String() == "" || email.isZero() || hashedPassword.isZero() || role.isZero() {
		return User{}, ErrInvalidUser
	}

	created := createdAt.UTC()
	updated := updatedAt.UTC()
	if created.IsZero() || updated.IsZero() || updated.Before(created) {
		return User{}, ErrInvalidUser
	}

	return User{
		id:             id,
		username:       username,
		email:          email,
		hashedPassword: hashedPassword,
		role:           role,
		createdAt:      created,
		updatedAt:      updated,
	}, nil
}
