package category_repository

import (
	"context"

	category_dto "auth-login/module/category/dto"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CategoryRepository interface {
	Create(ctx context.Context, req *category_dto.Create) (*category_dto.Response, error)
	List(ctx context.Context, filter *category_dto.Filter) ([]category_dto.Response, error)
	WithChildren(ctx context.Context, tournamentID int64) ([]*category_dto.WithChildrenResponse, error)
	Show(ctx context.Context, id int64) (*category_dto.Response, error)
	Update(ctx context.Context, id int64, req *category_dto.Update) (*category_dto.Response, error)
	Delete(ctx context.Context, id int64) error
}

type categoryRepository struct {
	db *pgxpool.Pool
}

func NewCategoryRepository(db *pgxpool.Pool) CategoryRepository {
	return &categoryRepository{db: db}
}

func (repository *categoryRepository) Create(ctx context.Context, req *category_dto.Create) (*category_dto.Response, error) {
	var res category_dto.Response

	err := repository.db.QueryRow(ctx, `
		INSERT INTO categories(tournament_id, parent_id, name, is_active)
		VALUES ($1, $2, $3, $4)
		RETURNING id, tournament_id, parent_id, name, is_active, created_at, updated_at
	`, req.TournamentID, req.ParentID, req.Name, req.IsActive).Scan(
		&res.ID,
		&res.TournamentID,
		&res.ParentID,
		&res.Name,
		&res.IsActive,
		&res.CreatedAt,
		&res.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (repository *categoryRepository) List(ctx context.Context, filter *category_dto.Filter) ([]category_dto.Response, error) {
	rows, err := repository.db.Query(ctx, `
		SELECT id, tournament_id, parent_id, name, is_active, created_at, updated_at
		FROM categories
		WHERE ($1::BIGINT = 0 OR id < $1)
		  AND ($2::BIGINT = 0 OR tournament_id = $2)
		  AND ((COALESCE($3::BIGINT, 0) = 0 AND parent_id IS NULL) OR ($3 > 0 AND parent_id = $3))
		  AND ($4::TEXT = '' OR name ILIKE '%' || $4 || '%')
		  AND ($5::BOOLEAN IS NULL OR is_active = $5)
		ORDER BY id DESC
		LIMIT $6
	`, filter.Cursor, filter.TournamentID, filter.ParentID, filter.Name, filter.IsActive, filter.Limit+1)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	result := make([]category_dto.Response, 0, filter.Limit+1)

	for rows.Next() {
		var res category_dto.Response

		err := rows.Scan(
			&res.ID,
			&res.TournamentID,
			&res.ParentID,
			&res.Name,
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

func (repository *categoryRepository) WithChildren(ctx context.Context, tournamentID int64) ([]*category_dto.WithChildrenResponse, error) {
	rows, err := repository.db.Query(ctx, `
		SELECT category.id, category.tournament_id, category.parent_id, category.name,
		       category.is_active, category.created_at, category.updated_at,
		       category_children_json(category.tournament_id, category.id) AS children
		FROM categories AS category
		WHERE category.tournament_id = $1
		  AND category.parent_id IS NULL
		ORDER BY category.id DESC
	`, tournamentID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	result := make([]*category_dto.WithChildrenResponse, 0)

	for rows.Next() {
		var res category_dto.WithChildrenResponse

		err := rows.Scan(
			&res.ID,
			&res.TournamentID,
			&res.ParentID,
			&res.Name,
			&res.IsActive,
			&res.CreatedAt,
			&res.UpdatedAt,
			&res.Children,
		)

		if err != nil {
			return nil, err
		}

		result = append(result, &res)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (repository *categoryRepository) Show(ctx context.Context, id int64) (*category_dto.Response, error) {
	var res category_dto.Response

	err := repository.db.QueryRow(ctx, `
		SELECT id, tournament_id, parent_id, name, is_active, created_at, updated_at
		FROM categories WHERE id = $1
	`, id).Scan(
		&res.ID,
		&res.TournamentID,
		&res.ParentID,
		&res.Name,
		&res.IsActive,
		&res.CreatedAt,
		&res.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (repository *categoryRepository) Update(ctx context.Context, id int64, req *category_dto.Update) (*category_dto.Response, error) {
	var res category_dto.Response

	err := repository.db.QueryRow(ctx, `
		UPDATE categories
		SET tournament_id = $1, parent_id = $2, name = $3, is_active = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $5
		RETURNING id, tournament_id, parent_id, name, is_active, created_at, updated_at
	`, req.TournamentID, req.ParentID, req.Name, req.IsActive, id).Scan(
		&res.ID,
		&res.TournamentID,
		&res.ParentID,
		&res.Name,
		&res.IsActive,
		&res.CreatedAt,
		&res.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (repository *categoryRepository) Delete(ctx context.Context, id int64) error {
	result, err := repository.db.Exec(ctx, "DELETE FROM categories WHERE id = $1", id)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}
