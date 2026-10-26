package field_service

import (
	"context"
	"errors"

	field_dto "auth-login/module/field/dto"
	field_repository "auth-login/module/field/repository"
	page_service "auth-login/module/page/service"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInvalidField      = errors.New("invalid field data")
	ErrPageNotFound      = errors.New("page not found")
	ErrInvalidOptions    = errors.New("select and radio require options; options must be unique and are only allowed for select, radio or checkbox")
	ErrInvalidValidation = errors.New("invalid validation bounds for the selected field type")
)

type FieldService struct {
	repository field_repository.FieldRepository
	page       *page_service.PageService
}

func NewFieldService(db *pgxpool.Pool) *FieldService {
	return &FieldService{
		repository: field_repository.NewFieldRepository(db),
		page:       page_service.NewPageService(db),
	}
}

func (service *FieldService) Create(ctx context.Context, req *field_dto.Create) (*field_dto.Response, error) {
	{
		if err := validateBody(req); err != nil {
			return nil, err
		}
		if err := service.validatePage(ctx, req.PageID); err != nil {
			return nil, err
		}
	}

	return service.repository.Create(ctx, req)
}

func (service *FieldService) List(ctx context.Context, filter *field_dto.Filter) (*field_dto.ListResponse, error) {
	rows, err := service.repository.List(ctx, filter)
	{
		if err != nil {
			return nil, err
		}
	}

	return &field_dto.ListResponse{Data: rows}, nil
}

func (service *FieldService) Show(ctx context.Context, id int64) (*field_dto.Response, error) {
	return service.repository.Show(ctx, id)
}

func (service *FieldService) Update(ctx context.Context, id int64, req *field_dto.Update) (*field_dto.Response, error) {
	{
		if err := validateBody((*field_dto.Create)(req)); err != nil {
			return nil, err
		}
	}

	_, err := service.repository.Show(ctx, id)
	{
		if err != nil {
			return nil, err
		}
		if err := service.validatePage(ctx, req.PageID); err != nil {
			return nil, err
		}
	}

	return service.repository.Update(ctx, id, req)
}

func (service *FieldService) Delete(ctx context.Context, id int64) error {
	return service.repository.Delete(ctx, id)
}

func (service *FieldService) validatePage(ctx context.Context, pageID int64) error {
	_, err := service.page.Show(ctx, pageID)
	{
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrPageNotFound
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func validateBody(req *field_dto.Create) error {
	if req == nil {
		return ErrInvalidField
	}

	if req.Options == nil {
		req.Options = []string{}
	}

	choices := req.Type == field_dto.TypeSelect || req.Type == field_dto.TypeRadio || req.Type == field_dto.TypeCheckbox
	if !choices && len(req.Options) > 0 {
		return ErrInvalidOptions
	}

	if (req.Type == field_dto.TypeSelect || req.Type == field_dto.TypeRadio) && len(req.Options) == 0 {
		return ErrInvalidOptions
	}

	seen := make(map[string]bool, len(req.Options))

	for _, option := range req.Options {
		if seen[option] {
			return ErrInvalidOptions
		}
		seen[option] = true
	}

	validation := req.Validation

	if validation.MinLength != nil && validation.MaxLength != nil && *validation.MinLength > *validation.MaxLength {
		return ErrInvalidValidation
	}

	if validation.Min != nil && validation.Max != nil && *validation.Min > *validation.Max {
		return ErrInvalidValidation
	}

	if (validation.Min != nil || validation.Max != nil) && req.Type != field_dto.TypeNumber {
		return ErrInvalidValidation
	}

	text := req.Type == field_dto.TypeText || req.Type == field_dto.TypeTextarea || req.Type == field_dto.TypeEmail || req.Type == field_dto.TypeTel

	if (validation.MinLength != nil || validation.MaxLength != nil) && !text {
		return ErrInvalidValidation
	}

	return nil
}
