package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/imightbuyaboat/SOCKS5-Proxy/client/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	checkUserQuery = `
		select password_hash
		from users
		where username = $1`

	checkUserQueryName = "check_user"
)

type PostgreStorage struct {
	pool *pgxpool.Pool
}

func NewPostgresStorage(ctx context.Context, url, migrationsPath string) (models.Storage, error) {
	if err := migrateDB(url, migrationsPath); err != nil {
		return nil, err
	}

	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}
	config.MaxConns = 100
	config.MinConns = 10
	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		if _, err := conn.Prepare(ctx, checkUserQueryName, checkUserQuery); err != nil {
			return err
		}
		return nil
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	if err = pool.Ping(ctx); err != nil {
		return nil, err
	}

	ps := &PostgreStorage{
		pool: pool,
	}

	go func() {
		<-ctx.Done()
		ps.pool.Close()
	}()

	return ps, nil
}

func (s *PostgreStorage) CheckUser(ctx context.Context, u *models.User) error {
	var hashFromDB string
	if err := s.pool.QueryRow(ctx, checkUserQueryName, u.Username).Scan(&hashFromDB); err != nil {
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
