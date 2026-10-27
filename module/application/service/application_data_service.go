package application_service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"auth-login/helper"
	application_dto "auth-login/module/application/dto"
	field_dto "auth-login/module/field/dto"
	page_dto "auth-login/module/page/dto"
)

var ErrPageNotFound = errors.New("active page not found in the selected tournament and category")
var ErrPreviousPagesIncomplete = errors.New("avval oldingi sahifalarni to'ldiring")

type DataValidationError struct {
	Fields map[string]string
}

func (err *DataValidationError) Error() string {
	return "invalid application page fields"
}

func (service *ApplicationService) SaveData(ctx context.Context, req *application_dto.SaveData) (*application_dto.DataResponse, error) {
	if req == nil {
		return nil, ErrInvalidApplication
	}

	application, err := service.repository.Show(ctx, req.ApplicationID)
	{
		if err != nil {
			return nil, err
		}
	}

	if application.TournamentID != req.TournamentID || application.CategoryID != req.CategoryID {
		return nil, ErrInvalidApplication
	}

	form, err := service.page.Form(ctx, req.TournamentID, req.CategoryID)
	{
		if err != nil {
			return nil, err
		}
	}

	var page *page_dto.FormPage

	page_slugs := make([]string, 0, len(form.Data))

	for i := range form.Data {
		if form.Data[i].Slug == req.Slug {
			page = &form.Data[i]
			break
		}

		page_slugs = append(page_slugs, form.Data[i].Slug)

	}

	if len(page_slugs) > 0 {
		count, err := service.repository.CountPages(ctx, req.ApplicationID, page_slugs)
		{
			if err != nil {
				return nil, err
			}
		}

		if count != len(page_slugs) {
			return nil, ErrPreviousPagesIncomplete
		}
	}

	if page == nil || page.ID == nil {
		return nil, ErrPageNotFound
	}

	values, err := validatePageData(page.Fields, req.Fields)
	{
		if err != nil {
			return nil, err
		}
	}

	tx, err := service.repository.BeginData(ctx)
	{
		if err != nil {
			return nil, err
		}
	}

	defer tx.Rollback(context.Background())

	created, err := service.repository.SaveData(ctx, tx, req, *page.ID, values)
	{
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &application_dto.DataResponse{
		ApplicationID: application.ID,
		Page:          page.Slug,
		Data:          values,
		Created:       created,
	}, nil
}

func validatePageData(fields []field_dto.FormField, input map[string]json.RawMessage) ([]application_dto.DataValue, error) {
	errs := make(map[string]string)
	known := make(map[string]bool, len(fields))
	result := make([]application_dto.DataValue, 0)

	for _, field := range fields {
		known[field.Name] = true
		raw, exists := input[field.Name]
		if !exists {
			if field.Validation.Required {
				errs[field.Name] = "field is required"
			}
			continue
		}

		values, err := validateFieldValue(field, raw)
		{
			if err != nil {
				errs[field.Name] = err.Error()
				continue
			}
		}

		if field.ID == nil {
			return nil, errors.New("form field has no id")
		}

		for _, value := range values {
			result = append(result, application_dto.DataValue{Field: *field.ID, Name: field.Name, Value: value})
		}
	}

	for name := range input {
		if !known[name] {
			errs[name] = "field does not belong to this active page"
		}
	}

	if len(errs) > 0 {
		return nil, &DataValidationError{Fields: errs}
	}

	return result, nil
}

func validateFieldValue(field field_dto.FormField, raw json.RawMessage) ([]string, error) {
	if strings.TrimSpace(string(raw)) == "null" {
		return nil, errors.New("null is not allowed; omit optional fields instead")
	}

	values := make([]string, 0)
	if field.Type == field_dto.TypeCheckbox {
		if len(field.Options) == 0 {
			var checked bool
			if err := json.Unmarshal(raw, &checked); err != nil {
				return nil, errors.New("checkbox without options must be a boolean")
			}
			if checked {
				values = append(values, "true")
			}
		} else {
			if err := json.Unmarshal(raw, &values); err != nil {
				return nil, errors.New("checkbox must be an array of option strings")
			}

			seen := make(map[string]bool, len(values))
			for i, value := range values {
				value = strings.TrimSpace(value)
				if !slices.Contains(field.Options, value) || seen[value] {
					return nil, errors.New("checkbox options must be allowed and unique")
				}
				seen[value] = true
				values[i] = value
			}
		}
	} else {
		var value string
		if err := json.Unmarshal(raw, &value); err != nil {
			if field.Type != field_dto.TypeNumber {
				return nil, errors.New("value must be a string")
			}

			var number json.Number
			if err := json.Unmarshal(raw, &number); err != nil {
				return nil, errors.New("value must be a number or numeric string")
			}
			value = number.String()
		}

		value = strings.TrimSpace(value)
		if value != "" {
			if err := validateScalarValue(field, value); err != nil {
				return nil, err
			}
			values = append(values, value)
		}
	}

	if field.Validation.Required && len(values) == 0 {
		return nil, errors.New("field is required")
	}

	return values, nil
}

func validateScalarValue(field field_dto.FormField, value string) error {
	if strings.ContainsRune(value, '\x00') {
		return errors.New("value must not contain null characters")
	}

	validation := field.Validation
	length := utf8.RuneCountInString(value)
	if validation.MinLength != nil && length < *validation.MinLength {
		return fmt.Errorf("must contain at least %d characters", *validation.MinLength)
	}
	if validation.MaxLength != nil && length > *validation.MaxLength {
		return fmt.Errorf("must contain at most %d characters", *validation.MaxLength)
	}

	switch field.Type {
	case field_dto.TypeNumber:
		number, err := strconv.ParseFloat(value, 64)
		if err != nil || math.IsNaN(number) || math.IsInf(number, 0) {
			return errors.New("must be a finite number")
		}
		if validation.Min != nil && number < *validation.Min {
			return fmt.Errorf("must be at least %g", *validation.Min)
		}
		if validation.Max != nil && number > *validation.Max {
			return fmt.Errorf("must be at most %g", *validation.Max)
		}
	case field_dto.TypeEmail:
		if helper.Validate(struct {
			Email string `json:"email" validate:"email,max=254"`
		}{Email: value}) != nil {
			return errors.New("invalid email address")
		}
	case field_dto.TypeDate:
		date, err := time.Parse(time.DateOnly, value)
		if err != nil || date.Year() < 1 {
			return errors.New("must be a valid YYYY-MM-DD date")
		}
	case field_dto.TypeSelect, field_dto.TypeRadio:
		if !slices.Contains(field.Options, value) {
			return errors.New("value must be one of the allowed options")
		}
	case field_dto.TypeText, field_dto.TypeTextarea, field_dto.TypeTel, field_dto.TypeFile:
	default:
		return errors.New("unsupported field type")
	}

	return nil
}
