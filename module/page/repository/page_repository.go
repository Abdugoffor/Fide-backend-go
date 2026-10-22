package page_repository

import (
	"context"

	page_dto "auth-login/module/page/dto"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PageRepository interface {
	Create(ctx context.Context, req *page_dto.Create) (*page_dto.Response, error)
	List(ctx context.Context, filter *page_dto.Filter) ([]page_dto.Response, error)
	Form(ctx context.Context, tournamentID, categoryID int64) ([]page_dto.FormPage, error)
	Show(ctx context.Context, id int64) (*page_dto.Response, error)
	Update(ctx context.Context, id int64, req *page_dto.Update) (*page_dto.Response, error)
	Delete(ctx context.Context, id int64) error
}

type pageRepository struct {
	db *pgxpool.Pool
}

func NewPageRepository(db *pgxpool.Pool) PageRepository {
	return &pageRepository{db: db}
}

func (repository *pageRepository) Create(ctx context.Context, req *page_dto.Create) (*page_dto.Response, error) {
	var res page_dto.Response

	err := repository.db.QueryRow(ctx, `
		INSERT INTO pages(tournament_id, category_id, name, sort, is_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, tournament_id, category_id, name, sort, is_active, created_at, updated_at
	`, req.TournamentID, req.CategoryID, req.Name, req.Sort, req.IsActive).Scan(
		&res.ID,
		&res.TournamentID,
		&res.CategoryID,
		&res.Name,
		&res.Sort,
		&res.IsActive,
		&res.CreatedAt,
		&res.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (repository *pageRepository) List(ctx context.Context, filter *page_dto.Filter) ([]page_dto.Response, error) {
	rows, err := repository.db.Query(ctx, `
		SELECT id, tournament_id, category_id, name, sort, is_active, created_at, updated_at
		FROM pages
		WHERE ($1::BIGINT = 0 OR tournament_id = $1)
		  AND ($2::BIGINT = 0 OR category_id = $2)
		  AND ($3::TEXT = '' OR name ILIKE '%' || $3 || '%')
		  AND ($4::BOOLEAN IS NULL OR is_active = $4)
		ORDER BY tournament_id, category_id, sort, id
	`, filter.TournamentID, filter.CategoryID, filter.Name, filter.IsActive)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	result := make([]page_dto.Response, 0)

	for rows.Next() {
		var res page_dto.Response

		err := rows.Scan(
			&res.ID,
			&res.TournamentID,
			&res.CategoryID,
			&res.Name,
			&res.Sort,
			&res.IsActive,
			&res.CreatedAt,
			&res.UpdatedAt,
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

func (repository *pageRepository) Show(ctx context.Context, id int64) (*page_dto.Response, error) {
	var res page_dto.Response

	err := repository.db.QueryRow(ctx, `
		SELECT id, tournament_id, category_id, name, sort, is_active, created_at, updated_at
		FROM pages WHERE id = $1
	`, id).Scan(
		&res.ID,
		&res.TournamentID,
		&res.CategoryID,
		&res.Name,
		&res.Sort,
		&res.IsActive,
		&res.CreatedAt,
		&res.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (repository *pageRepository) Update(ctx context.Context, id int64, req *page_dto.Update) (*page_dto.Response, error) {
	var res page_dto.Response

	err := repository.db.QueryRow(ctx, `
		UPDATE pages
		SET tournament_id = $1, category_id = $2, name = $3, sort = $4, is_active = $5, updated_at = CURRENT_TIMESTAMP
		WHERE id = $6
		RETURNING id, tournament_id, category_id, name, sort, is_active, created_at, updated_at
	`, req.TournamentID, req.CategoryID, req.Name, req.Sort, req.IsActive, id).Scan(
		&res.ID,
		&res.TournamentID,
		&res.CategoryID,
		&res.Name,
		&res.Sort,
		&res.IsActive,
		&res.CreatedAt,
		&res.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (repository *pageRepository) Delete(ctx context.Context, id int64) error {
	result, err := repository.db.Exec(ctx, "DELETE FROM pages WHERE id = $1", id)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (repository *pageRepository) Form(ctx context.Context, tournamentID, categoryID int64) ([]page_dto.FormPage, error) {
	rows, err := repository.db.Query(ctx, `
		SELECT page.id, page.name,
		       COALESCE((
		           SELECT jsonb_agg(jsonb_build_object(
		               'id', field.id, 'type', field.type,
		               'options', field.options, 'name', field.name, 'validation', field.validation
		           ) ORDER BY field.sort ASC, field.id ASC)
		           FROM fields AS field
		           WHERE field.page_id = page.id AND field.is_active = TRUE
		       ), '[]'::JSONB) AS fields
		FROM pages AS page
		WHERE page.tournament_id = $1 AND page.category_id = $2 AND page.is_active = TRUE
		ORDER BY page.sort ASC, page.id ASC
	`, tournamentID, categoryID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	result := make([]page_dto.FormPage, 0)

	for rows.Next() {
		var res page_dto.FormPage

		err := rows.Scan(
			&res.ID,
			&res.PageName,
			&res.Fields,
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
