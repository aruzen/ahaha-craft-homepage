package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
)

func NewConnection(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// TODO 消す
		dsn = "postgres://mnibackend:gatigatinopass@localhost:5432/my_name_is_backend?sslmode=disable"
	}

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	return pool, nil
}
