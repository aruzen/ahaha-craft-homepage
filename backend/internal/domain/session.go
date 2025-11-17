package domain

import (
	"crypto/rand"
	"encoding/base64"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// DefaultLoginSessionTTL は login_sessions.expires_at のデフォルト(30分)に合わせる。
const DefaultLoginSessionTTL = 30 * time.Minute

const loginSessionTokenByteLength = 32

type LoginSessionToken struct {
	value string
}

type HashedLoginSessionToken struct {
	value string
}

func NewLoginSessionToken() (LoginSessionToken, error) {
	tokenBytes := make([]byte, loginSessionTokenByteLength)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		return LoginSessionToken{}, err
	}
	token := base64.RawURLEncoding.EncodeToString(tokenBytes)

	return LoginSessionToken{value: token}, nil
}

func ParseLoginSessionToken(value string) (LoginSessionToken, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return LoginSessionToken{}, ErrInvalidSessionToken
	}

	decoded, err := base64.RawURLEncoding.DecodeString(trimmed)
	if err != nil || len(decoded) != loginSessionTokenByteLength {
		return LoginSessionToken{}, ErrInvalidSessionToken
	}

	return LoginSessionToken{value: trimmed}, nil
}

func (t LoginSessionToken) String() string {
	return t.value
}

func (t LoginSessionToken) isZero() bool {
	return t.value == ""
}

func (t LoginSessionToken) Hash() (HashedLoginSessionToken, error) {
	result := HashedLoginSessionToken{}

	hashed, err := bcrypt.GenerateFromPassword([]byte(t.value), bcrypt.DefaultCost)
	if err != nil {
		return HashedLoginSessionToken{}, err
	}
	result.value = string(hashed)
	return result, nil
}

func ParseHashedLoginSessionToken(value string) (HashedLoginSessionToken, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return HashedLoginSessionToken{}, ErrInvalidSessionToken
	}
	return HashedLoginSessionToken{value: trimmed}, nil
}

func (h HashedLoginSessionToken) String() string {
	return h.value
}

// LoginSession はログイン済みユーザーのセッション状態を表す。
type LoginSession struct {
	id        uuid.UUID
	userID    uuid.UUID
	token     HashedLoginSessionToken
	expiresAt time.Time
	createdAt time.Time
}

// NewLoginSession はセッションを発行時間を基準に構築する。
func NewLoginSession(userID uuid.UUID, token HashedLoginSessionToken, issuedAt time.Time) (LoginSession, error) {
	issued := issuedAt.UTC()
	if issued.IsZero() {
		return LoginSession{}, ErrInvalidLoginSession
	}

	return buildLoginSession(uuid.New(), userID, token, issued, issued.Add(DefaultLoginSessionTTL))
}

// NewLoginSessionFromPersistence は既存レコードからセッションを再構築する。
func NewLoginSessionFromPersistence(id uuid.UUID, userID uuid.UUID, token HashedLoginSessionToken, expiresAt, createdAt time.Time) (LoginSession, error) {
	return buildLoginSession(id, userID, token, createdAt, expiresAt)
}

func (s LoginSession) ID() uuid.UUID {
	return s.id
}

func (s LoginSession) UserID() uuid.UUID {
	return s.userID
}

func (s LoginSession) Token() HashedLoginSessionToken {
	return s.token
}

func (s LoginSession) HashedToken() string {
	return s.token.String()
}

func (s LoginSession) Verify(token LoginSessionToken) error {
	return bcrypt.CompareHashAndPassword([]byte(s.token.value), []byte(token.value))
}

func (s LoginSession) CreatedAt() time.Time {
	return s.createdAt
}

func (s LoginSession) ExpiresAt() time.Time {
	return s.expiresAt
}

// IsExpired は参照時刻が有効期限に到達したかどうかを返す。
func (s LoginSession) IsExpired(at time.Time) bool {
	return !at.UTC().Before(s.expiresAt)
}

func buildLoginSession(id uuid.UUID, userID uuid.UUID, token HashedLoginSessionToken, createdAt, expiresAt time.Time) (LoginSession, error) {
	if id == uuid.Nil || userID == uuid.Nil {
		return LoginSession{}, ErrInvalidLoginSession
	}

	created := createdAt.UTC()
	expires := expiresAt.UTC()
	if created.IsZero() || expires.IsZero() || !expires.After(created) {
		return LoginSession{}, ErrInvalidLoginSession
	}

	return LoginSession{
		id:        id,
		userID:    userID,
		token:     token,
		createdAt: created,
		expiresAt: expires,
	}, nil
}

// SessionData は API へ返却する session-data-struct を表現する。
type SessionData struct {
	userID uuid.UUID
	token  LoginSessionToken
}

func NewSessionData(userID uuid.UUID, token LoginSessionToken) (SessionData, error) {
	if userID == uuid.Nil || token.isZero() {
		return SessionData{}, ErrInvalidSessionData
	}

	return SessionData{userID: userID, token: token}, nil
}

func (s SessionData) UserID() uuid.UUID {
	return s.userID
}

func (s SessionData) Token() LoginSessionToken {
	return s.token
}
