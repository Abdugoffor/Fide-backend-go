package application_repository

import (
	"context"
	"fmt"

	application_dto "auth-login/module/application/dto"
	tournament_dto "auth-login/module/tournament/dto"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ApplicationRepository interface {
	Create(ctx context.Context, req *application_dto.Create) (*application_dto.Response, error)
	List(ctx context.Context, filter *application_dto.Filter) ([]application_dto.DetailResponse, error)
	HasMore(ctx context.Context, filter *application_dto.Filter, cursor int64, direction string) (bool, error)
	Show(ctx context.Context, id int64) (*application_dto.DetailResponse, error)
	MarkEmailVerified(ctx context.Context, id int64) error
	Delete(ctx context.Context, id int64) error
	BeginData(ctx context.Context) (pgx.Tx, error)
	CountPages(ctx context.Context, applicationID int64, slugs []string) (int, error)
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

func (repository *applicationRepository) List(ctx context.Context, filter *application_dto.Filter) ([]application_dto.DetailResponse, error) {
	condition := "($1::BIGINT = 0 OR id < $1)"
	order := "DESC"

	if filter.Direction == tournament_dto.DirectionPrev {
		condition = "id > $1"
		order = "ASC"
	}

	query := fmt.Sprintf(`
		SELECT application.id, application.tournament_id, tournament.name,
		       application.category_id, category.name,
		       application.name, application.email, application.birth_date::TEXT,
		       application.email_verified_at, application.created_at, application.updated_at,
		       COALESCE(saved.data, '[]'::JSONB)
		FROM (
			SELECT id, tournament_id, category_id, name, email, birth_date,
			       email_verified_at, created_at, updated_at
			FROM applications
			WHERE %s
			  AND ($2::BIGINT = 0 OR tournament_id = $2)
			  AND ($3::BIGINT = 0 OR category_id = $3)
			  AND ($4::TEXT = '' OR email ILIKE '%%' || $4 || '%%')
			  AND ($5::TEXT = '' OR name ILIKE '%%' || $5 || '%%')
			ORDER BY id %s
			LIMIT $6
		) AS application
		JOIN tournaments AS tournament ON tournament.id = application.tournament_id
		JOIN categories AS category ON category.id = application.category_id
		                              AND category.tournament_id = application.tournament_id
		LEFT JOIN LATERAL (
			SELECT jsonb_agg(jsonb_build_object(
				'slug', data.slug,
				'name', data.name, 'value', data.value
			) ORDER BY data.page, data.id) AS data
			FROM applications_data AS data
			WHERE data.application_id = application.id
		) AS saved ON TRUE
		ORDER BY application.id %s
	`, condition, order, order)

	rows, err := repository.db.Query(ctx, query, filter.Cursor, filter.TournamentID, filter.CategoryID, filter.Email, filter.Name, filter.Limit+1)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	result := make([]application_dto.DetailResponse, 0, filter.Limit+1)

	for rows.Next() {
		var res application_dto.DetailResponse

		err := rows.Scan(
			&res.ID,
			&res.TournamentID,
			&res.TournamentName,
			&res.CategoryID,
			&res.CategoryName,
			&res.Name,
			&res.Email,
			&res.BirthDate,
			&res.EmailVerifiedAt,
			&res.CreatedAt,
			&res.UpdatedAt,
			&res.Data,
		)

		if err != nil {
			return nil, err
		}

		result = append(result, res)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (repository *applicationRepository) HasMore(ctx context.Context, filter *application_dto.Filter, cursor int64, direction string) (bool, error) {
	operator := "<"

	if direction == tournament_dto.DirectionPrev {
		operator = ">"
	}

	query := fmt.Sprintf(`
		SELECT EXISTS (
			SELECT 1 FROM applications
			WHERE id %s $1
			  AND ($2::BIGINT = 0 OR tournament_id = $2)
			  AND ($3::BIGINT = 0 OR category_id = $3)
			  AND ($4::TEXT = '' OR email ILIKE '%%' || $4 || '%%')
			  AND ($5::TEXT = '' OR name ILIKE '%%' || $5 || '%%')
		)
	`, operator)

	var exists bool

	err := repository.db.QueryRow(ctx, query, cursor, filter.TournamentID, filter.CategoryID, filter.Email, filter.Name).Scan(&exists)

	if err != nil {
		return false, err
	}

	return exists, nil
}

func (repository *applicationRepository) Show(ctx context.Context, id int64) (*application_dto.DetailResponse, error) {
	var res application_dto.DetailResponse

	err := repository.db.QueryRow(ctx, `
		SELECT application.id, application.tournament_id, tournament.name,
		       application.category_id, category.name,
		       application.name, application.email, application.birth_date::TEXT,
		       application.email_verified_at, application.created_at, application.updated_at,
		       COALESCE(saved.data, '[]'::JSONB)
		FROM applications AS application
		JOIN tournaments AS tournament ON tournament.id = application.tournament_id
		JOIN categories AS category ON category.id = application.category_id
		                              AND category.tournament_id = application.tournament_id
		LEFT JOIN LATERAL (
			SELECT jsonb_agg(jsonb_build_object(
				'slug', data.slug,
				'name', data.name, 'value', data.value
			) ORDER BY data.page, data.id) AS data
			FROM applications_data AS data
			WHERE data.application_id = application.id
		) AS saved ON TRUE
		WHERE application.id = $1
	`, id).Scan(
		&res.ID,
		&res.TournamentID,
		&res.TournamentName,
		&res.CategoryID,
		&res.CategoryName,
		&res.Name,
		&res.Email,
		&res.BirthDate,
		&res.EmailVerifiedAt,
		&res.CreatedAt,
		&res.UpdatedAt,
		&res.Data,
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

func (repository *applicationRepository) CountPages(ctx context.Context, applicationID int64, slugs []string) (int, error) {
	var count int

	err := repository.db.QueryRow(ctx, `SELECT COUNT(DISTINCT slug) FROM applications_data 
	WHERE application_id = $1 AND slug = ANY($2::TEXT[])
	`, applicationID, slugs).Scan(&count)

	if err != nil {
		return 0, err
	}

	return count, nil
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
			INSERT INTO applications_data(application_id, page, slug, field, name, value)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, applicationID, pageID, req.Slug, value.Field, value.Name, value.Value)

		if err != nil {
			return false, err
		}
	}

	return result.RowsAffected() == 0, nil
}
