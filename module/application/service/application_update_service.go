package application_service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"time"

	application_dto "auth-login/module/application/dto"
	application_repository "auth-login/module/application/repository"
)

func (service *ApplicationService) Update(ctx context.Context, id int64, req *application_dto.Update) (*application_dto.UpdateResponse, error) {
	if req == nil {
		return nil, ErrInvalidApplication
	}

	birthDate, err := time.Parse(time.DateOnly, req.BirthDate)
	{
		if err != nil || birthDate.Year() < 1 || birthDate.After(time.Now()) {
			return nil, ErrInvalidBirthDate
		}
	}

	verificationMutex.Lock()
	defer verificationMutex.Unlock()

	application, err := service.repository.Show(ctx, id)
	{
		if err != nil {
			return nil, err
		}
	}

	form, err := service.page.Form(ctx, application.TournamentID, req.CategoryID)
	{
		if err != nil {
			return nil, err
		}
	}

	pages := make(map[string]int, len(form.Data))

	for i, page := range form.Data {
		if page.ID == nil {
			return nil, errors.New("form page has no id")
		}
		pages[page.Slug] = i
	}

	seen := make(map[string]bool, len(req.Pages))

	values := make([][]application_dto.DataValue, len(form.Data))

	errs := make(map[string]string)

	for _, page := range req.Pages {

		key := "pages." + page.Slug
		
		index, exists := pages[page.Slug]

		if !exists {
			errs[key] = "page does not belong to this active form"
			continue
		}

		if seen[page.Slug] {
			errs[key] = "page must be submitted only once"
			continue
		}

		seen[page.Slug] = true

		values[index], err = validatePageData(form.Data[index].Fields, page.Fields)
		{
			if err != nil {
				var validationErr *DataValidationError
				if !errors.As(err, &validationErr) {
					return nil, err
				}
				for name, message := range validationErr.Fields {
					errs[key+"."+name] = message
				}
			}
		}
	}

	for _, page := range form.Data {
		if !seen[page.Slug] {
			errs["pages."+page.Slug] = "page is required"
		}
	}

	if len(errs) > 0 {
		return nil, &DataValidationError{Fields: errs}
	}

	var previousVerification *application_dto.Verification

	var verification *application_dto.Verification

	if req.Email != application.Email || (req.CategoryID != application.CategoryID && application.EmailVerifiedAt == nil) {

		previousVerification, err = service.verification.Show(ctx, id)
		{
			if err != nil && !errors.Is(err, application_repository.ErrVerificationNotFound) {
				return nil, err
			}
		}

		number, err := rand.Int(rand.Reader, big.NewInt(1000000))
		{
			if err != nil {
				return nil, err
			}
		}

		verification = &application_dto.Verification{
			TournamentID:  application.TournamentID,
			CategoryID:    req.CategoryID,
			ApplicationID: id,
			Code:          fmt.Sprintf("%06d", number.Int64()),
			ExpiresAt:     time.Now().UTC().Add(verificationTTL),
		}
	}

	tx, err := service.repository.BeginData(ctx)
	{
		if err != nil {
			return nil, err
		}
	}

	defer tx.Rollback(context.Background())

	res, err := service.repository.Update(ctx, tx, id, req)
	{
		if err != nil {
			return nil, err
		}
	}

	if err := service.repository.DeleteData(ctx, tx, id); err != nil {
		return nil, err
	}

	for i, page := range form.Data {
		data := application_dto.SaveData{
			ApplicationID: id,
			TournamentID:  application.TournamentID,
			CategoryID:    req.CategoryID,
			Slug:          page.Slug,
		}
		_, err = service.repository.SaveData(ctx, tx, &data, *page.ID, values[i])
		{
			if err != nil {
				return nil, err
			}
		}
	}

	committed := false
	if verification != nil {
		if err := service.verification.Save(ctx, verification); err != nil {
			return nil, err
		}

		// The file is outside PostgreSQL; restore its previous entry if commit fails.
		defer func() {
			if committed {
				return
			}
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			var cleanupErr error
			if previousVerification == nil {
				cleanupErr = service.verification.Delete(cleanupCtx, id)
			} else {
				cleanupErr = service.verification.Save(cleanupCtx, previousVerification)
			}
			if cleanupErr != nil {
				log.Printf("Application %d verification restore error: %v", id, cleanupErr)
			}
		}()
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	committed = true

	response := &application_dto.UpdateResponse{Response: *res}
	if verification != nil {
		response.CodeExpiresAt = &verification.ExpiresAt
		log.Printf("Application verification: tournament_id=%d category_id=%d application_id=%d code=%s",
			verification.TournamentID, verification.CategoryID, verification.ApplicationID, verification.Code)
	}

	return response, nil
}
