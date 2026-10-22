package page_service

import (
	"context"
	"errors"
	"strings"

	"auth-login/helper"
	category_service "auth-login/module/category/service"
	page_dto "auth-login/module/page/dto"
	page_repository "auth-login/module/page/repository"
	tournament_service "auth-login/module/tournament/service"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInvalidPage        = errors.New("invalid page data; custom page sort must be at least 3")
	ErrTournamentNotFound = errors.New("tournament not found")
	ErrCategoryNotFound   = errors.New("category not found")
	ErrCategoryTournament = errors.New("category must belong to the selected tournament")
	ErrInactiveCategory   = errors.New("category is inactive")
)

type PageService struct {
	repository page_repository.PageRepository
	tournament *tournament_service.TournamentService
	category   *category_service.CategoryService
}

func NewPageService(db *pgxpool.Pool) *PageService {
	return &PageService{
		repository: page_repository.NewPageRepository(db),
		tournament: tournament_service.NewTournamentService(db),
		category:   category_service.NewCategoryService(db),
	}
}

func (service *PageService) Create(ctx context.Context, req *page_dto.Create) (*page_dto.Response, error) {

	if err := validateBody(req); err != nil {
		return nil, err
	}

	if err := service.validateCategory(ctx, req.TournamentID, req.CategoryID, false); err != nil {
		return nil, err
	}

	return service.repository.Create(ctx, req)
}

func (service *PageService) List(ctx context.Context, filter *page_dto.Filter) (*page_dto.ListResponse, error) {
	rows, err := service.repository.List(ctx, filter)
	{
		if err != nil {
			return nil, err
		}
	}

	return &page_dto.ListResponse{Data: rows}, nil
}

func (service *PageService) Form(ctx context.Context, tournamentID, categoryID int64) (*page_dto.FormResponse, error) {

	if tournamentID < 1 || categoryID < 1 {
		return nil, ErrInvalidPage
	}

	if err := service.validateCategory(ctx, tournamentID, categoryID, true); err != nil {
		return nil, err
	}

	pages, err := service.repository.Form(ctx, tournamentID, categoryID)
	{
		if err != nil {
			return nil, err
		}
	}

	return &page_dto.FormResponse{Total: len(pages), Data: pages}, nil
}

func (service *PageService) Show(ctx context.Context, id int64) (*page_dto.Response, error) {
	return service.repository.Show(ctx, id)
}

func (service *PageService) Update(ctx context.Context, id int64, req *page_dto.Update) (*page_dto.Response, error) {
	{
		if err := validateBody((*page_dto.Create)(req)); err != nil {
			return nil, err
		}
	}

	_, err := service.repository.Show(ctx, id)
	{
		if err != nil {
			return nil, err
		}
		if err := service.validateCategory(ctx, req.TournamentID, req.CategoryID, false); err != nil {
			return nil, err
		}
	}

	return service.repository.Update(ctx, id, req)
}

func (service *PageService) Delete(ctx context.Context, id int64) error {
	return service.repository.Delete(ctx, id)
}

func (service *PageService) validateCategory(ctx context.Context, tournamentID, categoryID int64, activeOnly bool) error {
	_, err := service.tournament.Show(ctx, tournamentID)
	{
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrTournamentNotFound
		}
		if err != nil {
			return err
		}
	}

	category, err := service.category.Show(ctx, categoryID)
	{
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrCategoryNotFound
		}
		if err != nil {
			return err
		}
	}

	if category.TournamentID != tournamentID {
		return ErrCategoryTournament
	}

	if activeOnly && !category.IsActive {
		return ErrInactiveCategory
	}

	return nil
}

func validateBody(req *page_dto.Create) error {
	if req == nil {
		return ErrInvalidPage
	}

	req.Name = strings.TrimSpace(req.Name)
	{
		if helper.Validate(req) != nil {
			return ErrInvalidPage
		}
	}

	return nil
}
