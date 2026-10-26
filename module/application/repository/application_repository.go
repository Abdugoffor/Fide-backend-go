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
	BeginData(ctx context.Context) (pgx.Tx, error)
	SaveData(ctx context.Context, tx pgx.Tx, req *application_dto.SaveData, pageID int64, values []application_dto.DataValue) (bool, error)
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

func (repository *applicationRepository) BeginData(ctx context.Context) (pgx.Tx, error) {
	return repository.db.Begin(ctx)
}

func (repository *applicationRepository) SaveData(ctx context.Context, tx pgx.Tx, req *application_dto.SaveData, pageID int64, values []application_dto.DataValue) (bool, error) {
	var applicationID int64

	// Serialize replacements, including the first submission with no existing data.
	err := tx.QueryRow(ctx, `
		SELECT id FROM applications
		WHERE id = $1 AND tournament_id = $2 AND category_id = $3
		FOR UPDATE
	`, req.ApplicationID, req.TournamentID, req.CategoryID).Scan(&applicationID)

	if err != nil {
		return false, err
	}

	result, err := tx.Exec(ctx, `
		DELETE FROM applications_data WHERE application_id = $1 AND page = $2
	`, applicationID, pageID)

	if err != nil {
		return false, err
	}

	for _, value := range values {
		_, err = tx.Exec(ctx, `
			INSERT INTO applications_data(application_id, page, slug, field, value)
			VALUES ($1, $2, $3, $4, $5)
		`, applicationID, pageID, req.Slug, value.Field, value.Value)

		if err != nil {
			return false, err
		}
	}

	return result.RowsAffected() == 0, nil
}
