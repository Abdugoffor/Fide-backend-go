package application_repository

import (
	"context"

	application_dto "auth-login/module/application/dto"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ApplicationRepository interface {
	Create(ctx context.Context, req *application_dto.Create) (*application_dto.Response, error)
	Show(ctx context.Context, id int64) (*application_dto.Response, error)
	MarkEmailVerified(ctx context.Context, id int64) error
	Delete(ctx context.Context, id int64) error
}

type applicationRepository struct {
	db *pgxpool.Pool
}

func NewApplicationRepository(db *pgxpool.Pool) ApplicationRepository {
	return &applicationRepository{db: db}
}

func (repository *applicationRepository) Create(ctx context.Context, req *application_dto.Create) (*application_dto.Response, error) {
	var res application_dto.Response

	err := repository.db.QueryRow(ctx, `
		INSERT INTO applications(tournament_id, category_id, name, email, birth_date)
		VALUES ($1, $2, $3, $4, $5::TEXT::DATE)
		RETURNING id, tournament_id, category_id, name, email, birth_date::TEXT,
		          email_verified_at, created_at, updated_at
	`, req.TournamentID, req.CategoryID, req.Name, req.Email, req.BirthDate).Scan(
		&res.ID,
		&res.TournamentID,
		&res.CategoryID,
		&res.Name,
		&res.Email,
		&res.BirthDate,
		&res.EmailVerifiedAt,
		&res.CreatedAt,
		&res.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (repository *applicationRepository) Show(ctx context.Context, id int64) (*application_dto.Response, error) {
	var res application_dto.Response

	err := repository.db.QueryRow(ctx, `
		SELECT id, tournament_id, category_id, name, email, birth_date::TEXT,
		       email_verified_at, created_at, updated_at
		FROM applications
		WHERE id = $1
	`, id).Scan(
		&res.ID,
		&res.TournamentID,
		&res.CategoryID,
		&res.Name,
		&res.Email,
		&res.BirthDate,
		&res.EmailVerifiedAt,
		&res.CreatedAt,
		&res.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (repository *applicationRepository) MarkEmailVerified(ctx context.Context, id int64) error {
	result, err := repository.db.Exec(ctx, `
		UPDATE applications
		SET email_verified_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND email_verified_at IS NULL
	`, id)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (repository *applicationRepository) Delete(ctx context.Context, id int64) error {
	result, err := repository.db.Exec(ctx, "DELETE FROM applications WHERE id = $1", id)
	{
		if err != nil {
			return err
		}
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}
