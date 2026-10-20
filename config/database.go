package config

import (
	"context"
	"fmt"
	"net"
	"net/url"

	"auth-login/helper"

	"github.com/jackc/pgx/v5/pgxpool"
)

func DBConnect(ctx context.Context) (*pgxpool.Pool, error) {
	// Ulanish ma'lumotlari .env faylidagi DB_* qiymatlaridan olinadi.
	dsn := fmt.Sprintf(
		"postgres://%s@%s/%s?sslmode=%s&timezone=%s",
		url.UserPassword(helper.ENV("DB_USER"), helper.ENV("DB_PASSWORD")),
		net.JoinHostPort(helper.ENV("DB_HOST"), helper.ENV("DB_PORT")),
		url.PathEscape(helper.ENV("DB_NAME")),
		url.QueryEscape(helper.ENV("DB_SSLMODE")),
		url.QueryEscape(helper.ENV("DB_TIMEZONE")),
	)

	db, err := pgxpool.New(ctx, dsn)
	{
		if err != nil {
			return nil, fmt.Errorf("connect to PostgreSQL: %w", err)
		}

		if err := db.Ping(ctx); err != nil {
			db.Close()
			return nil, fmt.Errorf("ping PostgreSQL: %w", err)
		}
	}

	return db, nil
}
