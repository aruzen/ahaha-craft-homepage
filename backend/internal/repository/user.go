package repository

import (
	"backend/internal/domain"
	"context"
	"time"

	"github.com/google/uuid"
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

type rowScanner interface {
	Scan(dest ...interface{}) error
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
