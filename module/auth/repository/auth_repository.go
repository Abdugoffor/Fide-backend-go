package auth_repository

import (
	"context"
	"time"

	auth_dto "auth-login/module/auth/dto"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	auth_dto.Response
	PasswordHash string `json:"-"`
}

type AuthRepository interface {
	SeedSuperadmin(ctx context.Context, passwordHash string) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	CreateSession(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error
	FindBySession(ctx context.Context, userID int64, tokenHash string) (*auth_dto.Response, error)
	DeleteSession(ctx context.Context, userID int64, tokenHash string) error
}

type authRepository struct {
	db *pgxpool.Pool
}

func NewAuthRepository(db *pgxpool.Pool) AuthRepository {
	return &authRepository{
		db: db,
	}
}

func (repository *authRepository) SeedSuperadmin(ctx context.Context, passwordHash string) error {
	_, err := repository.db.Exec(ctx, `
		INSERT INTO users(name, email, password_hash, role, status)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (email) DO NOTHING
	`, "Superadmin", "admin@gmail.com", passwordHash, auth_dto.RoleSuperadmin, auth_dto.StatusActive)

	if err != nil {
		return err
	}

	return nil
}

func (repository *authRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	var user User

	err := repository.db.QueryRow(ctx, `
		SELECT id, name, email, password_hash, role, status, created_at, updated_at
		FROM users
		WHERE email = $1
	`, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (repository *authRepository) CreateSession(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error {
	result, err := repository.db.Exec(ctx, `
		INSERT INTO user_sessions(token_hash, user_id, expires_at)
		SELECT $1, id, $2 FROM users WHERE id = $3 AND status = $4
	`, tokenHash, expiresAt, userID, auth_dto.StatusActive)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (repository *authRepository) FindBySession(ctx context.Context, userID int64, tokenHash string) (*auth_dto.Response, error) {
	var user auth_dto.Response

	err := repository.db.QueryRow(ctx, `
		SELECT u.id, u.name, u.email, u.role, u.status, u.created_at, u.updated_at
		FROM user_sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.token_hash = $1 AND s.user_id = $2 AND s.expires_at > NOW()
	`, tokenHash, userID).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (repository *authRepository) DeleteSession(ctx context.Context, userID int64, tokenHash string) error {
	result, err := repository.db.Exec(ctx, `
		DELETE FROM user_sessions WHERE token_hash = $1 AND user_id = $2 AND expires_at > NOW()
	`, tokenHash, userID)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}
