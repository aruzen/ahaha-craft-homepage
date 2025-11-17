package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestParseLoginSessionToken(t *testing.T) {
	generated, err := NewLoginSessionToken()
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	parsed, err := ParseLoginSessionToken(generated.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if parsed.String() != generated.String() {
		t.Fatalf("expected %s got %s", generated.String(), parsed.String())
	}

	if _, err := ParseLoginSessionToken("invalid-base64"); !errors.Is(err, ErrInvalidSessionToken) {
		t.Fatalf("expected ErrInvalidSessionToken, got %v", err)
	}
}

func TestParseHashedLoginSessionToken(t *testing.T) {
	if _, err := ParseHashedLoginSessionToken("  "); !errors.Is(err, ErrInvalidSessionToken) {
		t.Fatalf("expected ErrInvalidSessionToken, got %v", err)
	}

	hashed, err := ParseHashedLoginSessionToken("  hashed-value  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hashed.String() != "hashed-value" {
		t.Fatalf("expected trimmed hash, got %s", hashed.String())
	}
}

func TestNewLoginSession(t *testing.T) {
	userID := uuid.New()
	token, err := NewLoginSessionToken()
	if err != nil {
		t.Fatalf("token error: %v", err)
	}
	hashed, err := token.Hash()
	if err != nil {
		t.Fatalf("hash error: %v", err)
	}

	issuedAt := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	session, err := NewLoginSession(userID, hashed, issuedAt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if session.ID() == uuid.Nil {
		t.Fatalf("expected non-nil session id")
	}

	if session.UserID() != userID {
		t.Fatalf("unexpected user id: %v", session.UserID())
	}

	if err := session.Verify(token); err != nil {
		t.Fatalf("unexpected hashed token: %s", err)
	}

	if !session.CreatedAt().Equal(issuedAt) {
		t.Fatalf("expected created_at %v, got %v", issuedAt, session.CreatedAt())
	}

	expectedExpiry := issuedAt.Add(DefaultLoginSessionTTL)
	if !session.ExpiresAt().Equal(expectedExpiry) {
		t.Fatalf("expected expiry %v, got %v", expectedExpiry, session.ExpiresAt())
	}
}

func TestNewLoginSession_InvalidInput(t *testing.T) {
	token, err := NewLoginSessionToken()
	if err != nil {
		t.Fatalf("token error: %v", err)
	}

	cases := []struct {
		name   string
		userID uuid.UUID
		token  LoginSessionToken
		issued time.Time
	}{
		{"zero user", uuid.Nil, token, time.Now()},
		{"zero issued", uuid.New(), token, time.Time{}},
	}

	for _, tc := range cases {
		hashed, err := tc.token.Hash()
		if err != nil {
			t.Fatalf("hash error: %v", err)
		}
		if _, err := NewLoginSession(tc.userID, hashed, tc.issued); !errors.Is(err, ErrInvalidLoginSession) {
			t.Fatalf("%s: expected ErrInvalidLoginSession, got %v", tc.name, err)
		}
	}
}

func TestNewLoginSessionFromPersistence_InvalidTimes(t *testing.T) {
	token, err := NewLoginSessionToken()
	if err != nil {
		t.Fatalf("token error: %v", err)
	}
	hashed, err := token.Hash()
	if err != nil {
		t.Fatalf("hash error: %v", err)
	}

	created := time.Now().UTC()
	if _, err := NewLoginSessionFromPersistence(uuid.New(), uuid.New(), hashed, created, created); !errors.Is(err, ErrInvalidLoginSession) {
		t.Fatalf("expected ErrInvalidLoginSession when expiry <= created_at, got %v", err)
	}

	if _, err := NewLoginSessionFromPersistence(uuid.Nil, uuid.New(), hashed, created.Add(time.Minute), created); !errors.Is(err, ErrInvalidLoginSession) {
		t.Fatalf("expected ErrInvalidLoginSession for zero id, got %v", err)
	}
}

func TestLoginSession_IsExpired(t *testing.T) {
	token, err := NewLoginSessionToken()
	if err != nil {
		t.Fatalf("token error: %v", err)
	}
	hashed, err := token.Hash()
	if err != nil {
		t.Fatalf("hash error: %v", err)
	}

	created := time.Date(2025, 2, 3, 4, 5, 6, 0, time.UTC)
	expires := created.Add(time.Minute)
	session, err := NewLoginSessionFromPersistence(uuid.New(), uuid.New(), hashed, expires, created)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if session.IsExpired(created.Add(30 * time.Second)) {
		t.Fatalf("session should still be valid")
	}

	if !session.IsExpired(expires.Add(time.Nanosecond)) {
		t.Fatalf("session should be expired")
	}
}

func TestNewSessionData(t *testing.T) {
	userID := uuid.New()
	token, err := NewLoginSessionToken()
	if err != nil {
		t.Fatalf("token error: %v", err)
	}

	data, err := NewSessionData(userID, token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if data.UserID() != userID {
		t.Fatalf("unexpected user id")
	}

	if data.Token().String() != token.String() {
		t.Fatalf("unexpected token")
	}

	if _, err := NewSessionData(uuid.Nil, token); !errors.Is(err, ErrInvalidSessionData) {
		t.Fatalf("expected ErrInvalidSessionData, got %v", err)
	}

	if _, err := NewSessionData(userID, LoginSessionToken{}); !errors.Is(err, ErrInvalidSessionData) {
		t.Fatalf("expected ErrInvalidSessionData, got %v", err)
	}
}
