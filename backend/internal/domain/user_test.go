package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewEmail(t *testing.T) {
	email, err := NewEmail("  alice@example.com  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if email.String() != "alice@example.com" {
		t.Fatalf("expected trimmed email, got %s", email.String())
	}
}

func TestNewEmail_Invalid(t *testing.T) {
	if _, err := NewEmail("not-an-email"); !errors.Is(err, ErrInvalidEmail) {
		t.Fatalf("expected ErrInvalidEmail, got %v", err)
	}

	if _, err := NewEmail("  "); !errors.Is(err, ErrInvalidEmail) {
		t.Fatalf("expected ErrInvalidEmail on blank input, got %v", err)
	}
}

func TestNewHashedPassword(t *testing.T) {
	password, err := NewHashedPassword("  hashed  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if password.String() != "hashed" {
		t.Fatalf("expected trimmed hash, got %s", password.String())
	}
}

func TestNewHashedPassword_Invalid(t *testing.T) {
	if _, err := NewHashedPassword("\n\t"); !errors.Is(err, ErrInvalidPasswordHash) {
		t.Fatalf("expected ErrInvalidPasswordHash, got %v", err)
	}
}

func TestNewUserRole(t *testing.T) {
	role, err := NewUserRole("ADMIN")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if role != UserRoleAdmin {
		t.Fatalf("expected admin role, got %s", role)
	}
}

func TestNewUserRole_Invalid(t *testing.T) {
	if _, err := NewUserRole("guest"); !errors.Is(err, ErrInvalidUserRole) {
		t.Fatalf("expected ErrInvalidUserRole, got %v", err)
	}
}

func TestNewUser(t *testing.T) {
	name, _ := NewName("Alice")
	email, _ := NewEmail("alice@example.com")
	password, _ := NewHashedPassword("hashed")
	role, _ := NewUserRole("user")
	now := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)

	user, err := NewUser(name, email, password, role, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.ID() == uuid.Nil {
		t.Fatalf("expected generated id")
	}

	if user.Username().String() != "Alice" {
		t.Fatalf("unexpected username: %s", user.Username())
	}

	if !user.CreatedAt().Equal(now) {
		t.Fatalf("expected created_at %v, got %v", now, user.CreatedAt())
	}

	if !user.UpdatedAt().Equal(now) {
		t.Fatalf("expected updated_at %v, got %v", now, user.UpdatedAt())
	}
}

func TestNewUser_InvalidInput(t *testing.T) {
	email, _ := NewEmail("alice@example.com")
	password, _ := NewHashedPassword("hashed")
	role, _ := NewUserRole("user")
	now := time.Now()

	cases := []struct {
		name string
		user User
		fn   func() (User, error)
	}{
		{
			name: "zero time",
			fn: func() (User, error) {
				return NewUser(Name{}, email, password, role, time.Time{})
			},
		},
		{
			name: "zero name",
			fn: func() (User, error) {
				return NewUser(Name{}, email, password, role, now)
			},
		},
	}

	for _, tc := range cases {
		if _, err := tc.fn(); !errors.Is(err, ErrInvalidUser) && !errors.Is(err, ErrEmptyName) {
			t.Fatalf("%s: unexpected error %v", tc.name, err)
		}
	}
}

func TestNewUserFromPersistence_Invalid(t *testing.T) {
	name, _ := NewName("Alice")
	email, _ := NewEmail("alice@example.com")
	password, _ := NewHashedPassword("hashed")
	role, _ := NewUserRole("user")
	created := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	updated := created.Add(-time.Minute)

	if _, err := NewUserFromPersistence(uuid.Nil, name, email, password, role, created, created); !errors.Is(err, ErrInvalidUser) {
		t.Fatalf("expected ErrInvalidUser for zero id, got %v", err)
	}

	if _, err := NewUserFromPersistence(uuid.New(), name, email, password, role, created, updated); !errors.Is(err, ErrInvalidUser) {
		t.Fatalf("expected ErrInvalidUser when updated<created, got %v", err)
	}
}
