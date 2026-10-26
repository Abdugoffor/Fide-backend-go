package page_service

import (
	"context"
	"errors"
	"slices"
	"strings"

	category_service "auth-login/module/category/service"
	page_dto "auth-login/module/page/dto"
	page_repository "auth-login/module/page/repository"
	tournament_service "auth-login/module/tournament/service"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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

	if req == nil {
		return nil, ErrInvalidPage
	}

	if err := service.ValidateCategory(ctx, req.TournamentID, req.CategoryID, false); err != nil {
		return nil, err
	}

	baseSlug := pageSlug(req.Name)

	for attempt := 0; ; attempt++ {
		slug := baseSlug

		if attempt > 0 {
			suffix := pageSlugSuffix(attempt)
			slug = strings.TrimRight(baseSlug[:min(len(baseSlug), 199-len(suffix))], "-") + "-" + suffix
		}

		res, err := service.repository.Create(ctx, req, slug)
		{
			if err != nil {
				var pgErr *pgconn.PgError

				if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "pages_slug_unique" {
					continue
				}

				return nil, err
			}
		}

		return res, nil
	}
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

	if err := service.ValidateCategory(ctx, tournamentID, categoryID, true); err != nil {
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
		if req == nil {
			return nil, ErrInvalidPage
		}
	}

	_, err := service.repository.Show(ctx, id)
	{
		if err != nil {
			return nil, err
		}
		if err := service.ValidateCategory(ctx, req.TournamentID, req.CategoryID, false); err != nil {
			return nil, err
		}
	}

	return service.repository.Update(ctx, id, req)
}

func (service *PageService) Delete(ctx context.Context, id int64) error {
	return service.repository.Delete(ctx, id)
}

func (service *PageService) ValidateCategory(ctx context.Context, tournamentID, categoryID int64, activeOnly bool) error {
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

func pageSlug(name string) string {
	var slug strings.Builder
	separator := false

	for _, character := range strings.ToLower(name) {
		if character >= 'a' && character <= 'z' {
			if separator && slug.Len() > 0 {
				slug.WriteByte('-')
			}

			slug.WriteRune(character)
			separator = false
		} else if character != '\'' && character != '`' && character != '’' && character != '‘' && character != 'ʻ' && character != 'ʼ' {
			separator = true
		}
	}

	if slug.Len() == 0 {
		return "page"
	}

	return slug.String()
}

func pageSlugSuffix(number int) string {
	var suffix []byte

	for number > 0 {
		number--
		suffix = append(suffix, 'a'+byte(number%26))
		number /= 26
	}

	slices.Reverse(suffix)

	return string(suffix)
}
