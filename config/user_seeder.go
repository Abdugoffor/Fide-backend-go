package config

import (
	"context"

	"auth-login/helper"
	auth_service "auth-login/module/auth/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

func SeedSuperadmin(ctx context.Context, db *pgxpool.Pool) error {
	service := auth_service.NewAuthService(db, helper.ENV("JWT_SECRET"))

	return service.SeedSuperadmin(ctx)
}
