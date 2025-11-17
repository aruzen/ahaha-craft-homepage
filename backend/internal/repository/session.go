package repository

import (
	"backend/internal/domain"
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// LoginSessionRepository は login_sessions テーブルを扱う。
type LoginSessionRepository struct {
	db *pgxpool.Pool
}

func NewLoginSessionRepository(db *pgxpool.Pool) *LoginSessionRepository {
	return &LoginSessionRepository{db: db}
}

// Create はセッションを永続化する。
func (r *LoginSessionRepository) Create(ctx context.Context, session domain.LoginSession) error {
	const query = `
		INSERT INTO login_sessions (id, user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.Exec(ctx, query,
		session.ID(),
		session.UserID(),
		session.HashedToken(),
		session.ExpiresAt(),
		session.CreatedAt(),
	)
	return err
}

// Find は指定ユーザーのセッション群から入力トークンと一致するハッシュを探索する。
func (r *LoginSessionRepository) Find(ctx context.Context, userID uuid.UUID, token domain.LoginSessionToken) (domain.LoginSession, error) {
	const query = `
		SELECT id, user_id, token, expires_at, created_at
		FROM login_sessions
		WHERE user_id = $1
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return domain.LoginSession{}, err
	}
	defer rows.Close()

	for rows.Next() {
		session, err := scanLoginSession(rows)
		if err != nil {
			return domain.LoginSession{}, err
		}

		if bcrypt.CompareHashAndPassword([]byte(session.HashedToken()), []byte(token.String())) == nil {
			return session, nil
		}
	}

	if err := rows.Err(); err != nil {
		return domain.LoginSession{}, err
	}

	return domain.LoginSession{}, pgx.ErrNoRows
}

// DeleteByID は指定したセッションを削除する。
func (r *LoginSessionRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	const query = `
		DELETE FROM login_sessions
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, id)
	return err
}

func scanLoginSession(row rowScanner) (domain.LoginSession, error) {
	var (
		id        uuid.UUID
		userID    uuid.UUID
		token     string
		expiresAt time.Time
		createdAt time.Time
	)

	if err := row.Scan(&id, &userID, &token, &expiresAt, &createdAt); err != nil {
		return domain.LoginSession{}, err
	}

	hashedToken, err := domain.NewHashedLoginSessionTokenFromPersistence(token)
	if err != nil {
		return domain.LoginSession{}, err
	}

	return domain.NewLoginSessionFromPersistence(id, userID, hashedToken, expiresAt, createdAt)
}
