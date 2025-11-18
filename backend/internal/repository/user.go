package repository

import (
	"context"
	"errors"
	"time"

	"backend/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepository は users テーブルを読み書きする。
type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// FindByID は primary key でユーザーを検索する。
func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	const query = `
		SELECT id, username, email, hashed_password, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	row := r.db.QueryRow(ctx, query, id)
	return scanUser(row)
}

// FindByEmail はメールアドレスでユーザーを検索し、見つからなければ pgx.ErrNoRows を返す。
func (r *UserRepository) FindByEmail(ctx context.Context, email domain.Email) (domain.User, error) {
	const query = `
		SELECT id, username, email, hashed_password, role, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	row := r.db.QueryRow(ctx, query, email.String())
	return scanUser(row)
}

// FindByName は username 列をユニークキーとして検索する。
func (r *UserRepository) FindByName(ctx context.Context, name domain.Name) (domain.User, error) {
	const query = `
		SELECT id, username, email, hashed_password, role, created_at, updated_at
		FROM users
		WHERE username = $1
	`

	row := r.db.QueryRow(ctx, query, name.String())
	return scanUser(row)
}

// Create はユーザーを挿入し、ユニーク制約違反をドメインエラーへ変換する。
func (r *UserRepository) Create(ctx context.Context, user domain.User) error {
	const query = `
		INSERT INTO users (id, username, email, hashed_password, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Exec(ctx, query,
		user.ID(),
		user.Username().String(),
		user.Email().String(),
		user.HashedPassword().String(),
		user.Role().String(),
		user.CreatedAt(),
		user.UpdatedAt(),
	)
	if err != nil {
		return translateUserConstraintError(err)
	}
	return nil
}

func scanUser(row rowScanner) (domain.User, error) {
	var (
		id        uuid.UUID
		username  string
		email     string
		hash      string
		role      string
		createdAt time.Time
		updatedAt time.Time
	)

	if err := row.Scan(&id, &username, &email, &hash, &role, &createdAt, &updatedAt); err != nil {
		return domain.User{}, err
	}

	name, err := domain.NewName(username)
	if err != nil {
		return domain.User{}, err
	}

	domainEmail, err := domain.NewEmail(email)
	if err != nil {
		return domain.User{}, err
	}

	password, err := domain.NewHashedPassword(hash)
	if err != nil {
		return domain.User{}, err
	}

	userRole, err := domain.NewUserRole(role)
	if err != nil {
		return domain.User{}, err
	}

	return domain.NewUserFromPersistence(id, name, domainEmail, password, userRole, createdAt, updatedAt)
}

const (
	uniqueViolationCode   = "23505"
	usernameConstraintKey = "users_username_key"
	emailConstraintKey    = "users_email_key"
)

func translateUserConstraintError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}

	if pgErr.Code != uniqueViolationCode {
		return err
	}

	switch pgErr.ConstraintName {
	case usernameConstraintKey:
		return domain.ErrDuplicateUsername
	case emailConstraintKey:
		return domain.ErrDuplicateEmail
	default:
		return err
	}
}
