package category_service

import (
	"context"
	"errors"
	"strings"

	"auth-login/helper"
	category_dto "auth-login/module/category/dto"
	category_repository "auth-login/module/category/repository"
	tournament_service "auth-login/module/tournament/service"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInvalidCategory    = errors.New("invalid category data")
	ErrTournamentNotFound = errors.New("tournament not found")
	ErrParentNotFound     = errors.New("parent category not found")
	ErrParentTournament   = errors.New("parent category must belong to the same tournament")
	ErrParentCycle        = errors.New("parent category must not be itself or a descendant")
	ErrHierarchyConflict  = errors.New("category has child categories or its parent is no longer available")
	ErrConcurrentChange   = errors.New("category hierarchy changed concurrently; retry the request")
	ErrPageConflict       = errors.New("category has pages; remove or move its pages first")
)

type CategoryService struct {
	repository category_repository.CategoryRepository
	tournament *tournament_service.TournamentService
}

func NewCategoryService(db *pgxpool.Pool) *CategoryService {
	return &CategoryService{
		repository: category_repository.NewCategoryRepository(db),
		tournament: tournament_service.NewTournamentService(db),
	}
}

func (service *CategoryService) Create(ctx context.Context, req *category_dto.Create) (*category_dto.Response, error) {

	if err := validateBody(req); err != nil {
		return nil, err
	}

	if err := service.validateTournament(ctx, req.TournamentID); err != nil {
		return nil, err
	}

	if err := service.validateParent(ctx, 0, req.TournamentID, req.ParentID); err != nil {
		return nil, err
	}

	res, err := service.repository.Create(ctx, req)
	{
		if err != nil {
			return nil, serviceError(err)
		}
	}

	return res, nil
}

func (service *CategoryService) List(ctx context.Context, filter *category_dto.Filter) (*category_dto.ListResponse, error) {
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 10
	}

	rows, err := service.repository.List(ctx, filter)
	{
		if err != nil {
			return nil, err
		}
	}

	res := &category_dto.ListResponse{Data: rows, HasNext: len(rows) > filter.Limit}

	if res.HasNext {
		res.Data = rows[:filter.Limit]

		lastID := res.Data[len(res.Data)-1].ID

		res.NextCursor = &lastID
	}

	return res, nil
}

func (service *CategoryService) WithChildren(ctx context.Context, tournamentID int64) (*category_dto.WithChildrenListResponse, error) {

	if err := service.validateTournament(ctx, tournamentID); err != nil {
		return nil, err
	}

	rows, err := service.repository.WithChildren(ctx, tournamentID)
	{
		if err != nil {
			return nil, err
		}
	}

	return &category_dto.WithChildrenListResponse{Data: rows}, nil
}

func (service *CategoryService) Show(ctx context.Context, id int64) (*category_dto.Response, error) {
	return service.repository.Show(ctx, id)
}

func (service *CategoryService) Update(ctx context.Context, id int64, req *category_dto.Update) (*category_dto.Response, error) {

	if err := validateBody((*category_dto.Create)(req)); err != nil {
		return nil, err
	}

	if err := service.validateTournament(ctx, req.TournamentID); err != nil {
		return nil, err
	}

	_, err := service.repository.Show(ctx, id)
	{
		if err != nil {
			return nil, serviceError(err)
		}
	}

	if err := service.validateParent(ctx, id, req.TournamentID, req.ParentID); err != nil {
		return nil, serviceError(err)
	}

	res, err := service.repository.Update(ctx, id, req)
	{
		if err != nil {
			return nil, serviceError(err)
		}
	}

	return res, nil
}

func (service *CategoryService) Delete(ctx context.Context, id int64) error {
	return serviceError(service.repository.Delete(ctx, id))
}

func (service *CategoryService) validateTournament(ctx context.Context, tournamentID int64) error {
	_, err := service.tournament.Show(ctx, tournamentID)
	{
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrTournamentNotFound
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (service *CategoryService) validateParent(ctx context.Context, id, tournamentID int64, parentID *int64) error {
	visited := make(map[int64]bool)

	for parentID != nil {
		if *parentID == id || visited[*parentID] {
			return ErrParentCycle
		}

		visited[*parentID] = true

		parent, err := service.repository.Show(ctx, *parentID)
		{
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrParentNotFound
			}

			if err != nil {
				return err
			}
		}

		if parent.TournamentID != tournamentID {
			return ErrParentTournament
		}

		parentID = parent.ParentID
	}

	return nil
}

func validateBody(req *category_dto.Create) error {
	if req == nil {
		return ErrInvalidCategory
	}

	req.Name = strings.TrimSpace(req.Name)

	if helper.Validate(req) != nil {
		return ErrInvalidCategory
	}

	return nil
}

func serviceError(err error) error {
	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		if pgErr.Code == "40001" || pgErr.Code == "40P01" {
			return ErrConcurrentChange
		}

		if pgErr.Code == "23503" {
			switch pgErr.ConstraintName {
			case "categories_tournament_fk":
				return ErrTournamentNotFound
			case "categories_parent_tournament_fk":
				return ErrHierarchyConflict
			case "pages_category_tournament_fk":
				return ErrPageConflict
			}
		}
	}

	return err
}
