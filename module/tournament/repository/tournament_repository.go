package tournament_repository

import (
	"context"
	"fmt"

	tournament_dto "auth-login/module/tournament/dto"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TournamentRepository interface {
	Create(ctx context.Context, req *tournament_dto.Create) (*tournament_dto.Response, error)
	List(ctx context.Context, filter *tournament_dto.Filter) ([]tournament_dto.Response, error)
	HasMore(ctx context.Context, filter *tournament_dto.Filter, cursor int64, direction string) (bool, error)
	Show(ctx context.Context, id int64) (*tournament_dto.Response, error)
	Update(ctx context.Context, id int64, req *tournament_dto.Update) (*tournament_dto.Response, error)
	Delete(ctx context.Context, id int64) error
}

type tournamentRepository struct {
	db *pgxpool.Pool
}

func NewTournamentRepository(db *pgxpool.Pool) TournamentRepository {
	return &tournamentRepository{db: db}
}

func (repository *tournamentRepository) Create(ctx context.Context, req *tournament_dto.Create) (*tournament_dto.Response, error) {
	var res tournament_dto.Response

	err := repository.db.QueryRow(ctx, `
		INSERT INTO tournaments(name, description, location, starts_at, ends_at, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, name, description, location, starts_at, ends_at, status, created_at, updated_at
	`, req.Name, req.Description, req.Location, req.StartsAt, req.EndsAt, req.Status).Scan(
		&res.ID,
		&res.Name,
		&res.Description,
		&res.Location,
		&res.StartsAt,
		&res.EndsAt,
		&res.Status,
		&res.CreatedAt,
		&res.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (repository *tournamentRepository) List(ctx context.Context, filter *tournament_dto.Filter) ([]tournament_dto.Response, error) {
	condition := "($1::BIGINT = 0 OR id < $1)"
	order := "DESC"

	if filter.Direction == tournament_dto.DirectionPrev {
		condition = "id > $1"
		order = "ASC"
	}

	query := fmt.Sprintf(`
		SELECT id, name, description, location, starts_at, ends_at, status, created_at, updated_at
		FROM tournaments
		WHERE %s
		  AND ($2::TEXT = '' OR name ILIKE '%%' || $2 || '%%')
		  AND ($3::TEXT = '' OR status = $3)
		  AND ($4::TEXT = '' OR description ILIKE '%%' || $4 || '%%')
		ORDER BY id %s
		LIMIT $5
	`, condition, order)

	rows, err := repository.db.Query(ctx, query, filter.Cursor, filter.Name, filter.Status, filter.Description, filter.Limit+1)
	{
		if err != nil {
			return nil, err
		}
	}

	defer rows.Close()

	result := make([]tournament_dto.Response, 0, filter.Limit+1)

	for rows.Next() {
		var res tournament_dto.Response

		err := rows.Scan(
			&res.ID,
			&res.Name,
			&res.Description,
			&res.Location,
			&res.StartsAt,
			&res.EndsAt,
			&res.Status,
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

func (repository *tournamentRepository) HasMore(ctx context.Context, filter *tournament_dto.Filter, cursor int64, direction string) (bool, error) {
	operator := "<"

	if direction == tournament_dto.DirectionPrev {
		operator = ">"
	}

	query := fmt.Sprintf(`
		SELECT EXISTS (
			SELECT 1 FROM tournaments
			WHERE id %s $1
			  AND ($2::TEXT = '' OR name ILIKE '%%' || $2 || '%%')
			  AND ($3::TEXT = '' OR status = $3)
			  AND ($4::TEXT = '' OR description ILIKE '%%' || $4 || '%%')
		)
	`, operator)

	var exists bool

	err := repository.db.QueryRow(ctx, query, cursor, filter.Name, filter.Status, filter.Description).Scan(&exists)
	{
		if err != nil {
			return false, err
		}
	}

	return exists, nil
}

func (repository *tournamentRepository) Show(ctx context.Context, id int64) (*tournament_dto.Response, error) {
	var res tournament_dto.Response

	err := repository.db.QueryRow(ctx, `
		SELECT id, name, description, location, starts_at, ends_at, status, created_at, updated_at
		FROM tournaments WHERE id = $1
	`, id).Scan(
		&res.ID,
		&res.Name,
		&res.Description,
		&res.Location,
		&res.StartsAt,
		&res.EndsAt,
		&res.Status,
		&res.CreatedAt,
		&res.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (repository *tournamentRepository) Update(ctx context.Context, id int64, req *tournament_dto.Update) (*tournament_dto.Response, error) {
	var res tournament_dto.Response

	err := repository.db.QueryRow(ctx, `
		UPDATE tournaments
		SET name = $1, description = $2, location = $3, starts_at = $4, ends_at = $5, status = $6, updated_at = CURRENT_TIMESTAMP
		WHERE id = $7
		RETURNING id, name, description, location, starts_at, ends_at, status, created_at, updated_at
	`, req.Name, req.Description, req.Location, req.StartsAt, req.EndsAt, req.Status, id).Scan(
		&res.ID,
		&res.Name,
		&res.Description,
		&res.Location,
		&res.StartsAt,
		&res.EndsAt,
		&res.Status,
		&res.CreatedAt,
		&res.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (repository *tournamentRepository) Delete(ctx context.Context, id int64) error {
	result, err := repository.db.Exec(ctx, "DELETE FROM tournaments WHERE id = $1", id)
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
