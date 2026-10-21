package field_repository

import (
	"context"

	field_dto "auth-login/module/field/dto"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FieldRepository interface {
	Create(ctx context.Context, req *field_dto.Create) (*field_dto.Response, error)
	List(ctx context.Context, filter *field_dto.Filter) ([]field_dto.Response, error)
	Show(ctx context.Context, id int64) (*field_dto.Response, error)
	Update(ctx context.Context, id int64, req *field_dto.Update) (*field_dto.Response, error)
	Delete(ctx context.Context, id int64) error
}

type fieldRepository struct {
	db *pgxpool.Pool
}

func NewFieldRepository(db *pgxpool.Pool) FieldRepository {
	return &fieldRepository{db: db}
}

func (repository *fieldRepository) Create(ctx context.Context, req *field_dto.Create) (*field_dto.Response, error) {
	var res field_dto.Response

	err := repository.db.QueryRow(ctx, `
		INSERT INTO fields(page_id, sort, type, options, name, validation, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, page_id, sort, type, options, name, validation, is_active, created_at, updated_at
	`, req.PageID, req.Sort, req.Type, req.Options, req.Name, req.Validation, req.IsActive).Scan(
		&res.ID,
		&res.PageID,
		&res.Sort,
		&res.Type,
		&res.Options,
		&res.Name,
		&res.Validation,
		&res.IsActive,
		&res.CreatedAt,
		&res.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (repository *fieldRepository) List(ctx context.Context, filter *field_dto.Filter) ([]field_dto.Response, error) {
	rows, err := repository.db.Query(ctx, `
		SELECT id, page_id, sort, type, options, name, validation, is_active, created_at, updated_at
		FROM fields
		WHERE ($1::BIGINT = 0 OR page_id = $1)
		  AND ($2::TEXT = '' OR name ILIKE '%' || $2 || '%')
		  AND ($3::TEXT = '' OR type = $3)
		  AND ($4::BOOLEAN IS NULL OR is_active = $4)
		ORDER BY page_id, sort, id
	`, filter.PageID, filter.Name, filter.Type, filter.IsActive)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	result := make([]field_dto.Response, 0)

	for rows.Next() {
		var res field_dto.Response

		err := rows.Scan(
			&res.ID,
			&res.PageID,
			&res.Sort,
			&res.Type,
			&res.Options,
			&res.Name,
			&res.Validation,
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

func (repository *fieldRepository) Show(ctx context.Context, id int64) (*field_dto.Response, error) {
	var res field_dto.Response

	err := repository.db.QueryRow(ctx, `
		SELECT id, page_id, sort, type, options, name, validation, is_active, created_at, updated_at
		FROM fields WHERE id = $1
	`, id).Scan(
		&res.ID,
		&res.PageID,
		&res.Sort,
		&res.Type,
		&res.Options,
		&res.Name,
		&res.Validation,
		&res.IsActive,
		&res.CreatedAt,
		&res.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (repository *fieldRepository) Update(ctx context.Context, id int64, req *field_dto.Update) (*field_dto.Response, error) {
	var res field_dto.Response

	err := repository.db.QueryRow(ctx, `
		UPDATE fields
		SET page_id = $1, sort = $2, type = $3, options = $4, name = $5, validation = $6, is_active = $7, updated_at = CURRENT_TIMESTAMP
		WHERE id = $8
		RETURNING id, page_id, sort, type, options, name, validation, is_active, created_at, updated_at
	`, req.PageID, req.Sort, req.Type, req.Options, req.Name, req.Validation, req.IsActive, id).Scan(
		&res.ID,
		&res.PageID,
		&res.Sort,
		&res.Type,
		&res.Options,
		&res.Name,
		&res.Validation,
		&res.IsActive,
		&res.CreatedAt,
		&res.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (repository *fieldRepository) Delete(ctx context.Context, id int64) error {
	result, err := repository.db.Exec(ctx, "DELETE FROM fields WHERE id = $1", id)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}
