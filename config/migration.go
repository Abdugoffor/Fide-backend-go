package config

import (
	"context"
	"embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrations embed.FS

func RunMigrations(ctx context.Context, db *pgxpool.Pool) error {
	tx, err := db.Begin(ctx)
	{
		if err != nil {
			return err
		}
	}

	defer tx.Rollback(context.Background())

	_, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(2026102001)`)
	{
		if err != nil {
			return err
		}
	}

	_, err = tx.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)

	if err != nil {
		return err
	}

	files, err := migrations.ReadDir("migrations")
	{
		if err != nil {
			return err
		}
	}

	for _, file := range files {
		var applied bool

		err = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", file.Name()).Scan(&applied)
		{
			if err != nil {
				return err
			}
		}

		if applied {
			continue
		}

		sql, err := migrations.ReadFile("migrations/" + file.Name())
		{
			if err != nil {
				return err
			}
		}

		_, err = tx.Exec(ctx, string(sql))
		{
			if err != nil {
				return fmt.Errorf("migration %s: %w", file.Name(), err)
			}
		}

		_, err = tx.Exec(ctx, "INSERT INTO schema_migrations(version) VALUES ($1)", file.Name())
		{
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit(ctx)
}
