package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/imightbuyaboat/SOCKS5-Proxy/client/internal/user"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgreStorage struct {
	pool *pgxpool.Pool
	ctx  context.Context
}

func NewPostgresStorage(ctx context.Context, url string) (*PostgreStorage, error) {
	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}
	config.MaxConns = 100
	config.MinConns = 10

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	if err = pool.Ping(ctx); err != nil {
		return nil, err
	}

	query := `create table if not exists users (
				id serial primary key,
				username text,
				password_hash text not null
				);`

	if _, err = pool.Exec(ctx, query); err != nil {
		return nil, err
	}

	return &PostgreStorage{pool, ctx}, nil
}

func (s *PostgreStorage) CheckUser(u *user.User) error {
	query := `select password_hash
				from users
				where username = @username;`

	args := pgx.NamedArgs{
		"username": u.Username,
	}

	var hashFromDB string
	if err := s.pool.QueryRow(s.ctx, query, args).Scan(&hashFromDB); err != nil {
		if err == pgx.ErrNoRows {
			return ErrUserNotExists
		}
		return err
	}

	passwordHash := sha256.Sum256([]byte(u.Password))
	if hex.EncodeToString(passwordHash[:]) != hashFromDB {
		return ErrIncorrectPassword
	}

	return nil
}
