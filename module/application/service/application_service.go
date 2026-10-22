package application_service

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"
	"sync"
	"time"

	"auth-login/helper"
	application_dto "auth-login/module/application/dto"
	application_repository "auth-login/module/application/repository"
	page_dto "auth-login/module/page/dto"
	page_service "auth-login/module/page/service"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	verificationTTL         = 10 * time.Minute
	verificationMaxAttempts = 5
)

var (
	ErrInvalidApplication  = errors.New("invalid application data")
	ErrInvalidBirthDate    = errors.New("birth_date must be a valid YYYY-MM-DD date and must not be in the future")
	ErrInvalidVerification = errors.New("invalid verification details or code")
	ErrCodeExpired         = errors.New("verification code has expired")
	ErrTooManyAttempts     = errors.New("too many verification attempts")
	ErrAlreadyVerified     = errors.New("application email is already verified")
	verificationMutex      sync.Mutex
)

type ApplicationService struct {
	repository   application_repository.ApplicationRepository
	verification application_repository.VerificationRepository
	page         *page_service.PageService
}

func NewApplicationService(db *pgxpool.Pool) *ApplicationService {
	path := strings.TrimSpace(helper.ENV("APPLICATION_CODES_FILE"))
	if path == "" {
		path = ".local/application_codes.json"
	}

	return &ApplicationService{
		repository:   application_repository.NewApplicationRepository(db),
		verification: application_repository.NewVerificationRepository(path),
		page:         page_service.NewPageService(db),
	}
}

func (service *ApplicationService) Create(ctx context.Context, req *application_dto.Create) (*application_dto.CreateResponse, error) {

	if req == nil {
		return nil, ErrInvalidApplication
	}

	req.Name = strings.TrimSpace(req.Name)

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	req.BirthDate = strings.TrimSpace(req.BirthDate)
	{
		if helper.Validate(req) != nil {
			return nil, ErrInvalidApplication
		}
	}

	birthDate, err := time.Parse(time.DateOnly, req.BirthDate)
	{
		if err != nil || birthDate.Year() < 1 || birthDate.After(time.Now()) {
			return nil, ErrInvalidBirthDate
		}
	}

	if err := service.page.ValidateCategory(ctx, req.TournamentID, req.CategoryID, true); err != nil {
		return nil, err
	}

	number, err := rand.Int(rand.Reader, big.NewInt(1000000))
	{
		if err != nil {
			return nil, err
		}
	}

	res, err := service.repository.Create(ctx, req)
	{
		if err != nil {
			return nil, err
		}
	}

	verification := &application_dto.Verification{
		TournamentID:  res.TournamentID,
		CategoryID:    res.CategoryID,
		ApplicationID: res.ID,
		Code:          fmt.Sprintf("%06d", number.Int64()),
		ExpiresAt:     time.Now().UTC().Add(verificationTTL),
	}

	if err := service.verification.Save(ctx, verification); err != nil {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if cleanupErr := service.repository.Delete(cleanupCtx, res.ID); cleanupErr != nil {
			log.Printf("Application %d cleanup error: %v", res.ID, cleanupErr)
		}

		return nil, err
	}

	log.Printf("Application verification: tournament_id=%d category_id=%d application_id=%d code=%s",
		verification.TournamentID, verification.CategoryID, verification.ApplicationID, verification.Code)

	return &application_dto.CreateResponse{Response: *res, CodeExpiresAt: verification.ExpiresAt}, nil
}

func (service *ApplicationService) VerifyEmail(ctx context.Context, req *application_dto.VerifyEmail) (*page_dto.FormResponse, error) {
	if req == nil {
		return nil, ErrInvalidVerification
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	req.Code = strings.TrimSpace(req.Code)
	{
		if helper.Validate(req) != nil {
			return nil, ErrInvalidVerification
		}
	}

	verificationMutex.Lock()
	defer verificationMutex.Unlock()

	application, err := service.repository.Show(ctx, req.ApplicationID)
	{
		if err != nil {
			return nil, err
		}
	}

	if application.TournamentID != req.TournamentID || application.Email != req.Email {
		return nil, ErrInvalidVerification
	}

	if application.EmailVerifiedAt != nil {
		return nil, ErrAlreadyVerified
	}

	verification, err := service.verification.Show(ctx, application.ID)
	{
		if errors.Is(err, application_repository.ErrVerificationNotFound) {
			return nil, ErrInvalidVerification
		}
		if err != nil {
			return nil, err
		}
	}

	if verification.TournamentID != application.TournamentID || verification.CategoryID != application.CategoryID {
		return nil, ErrInvalidVerification
	}

	if !time.Now().Before(verification.ExpiresAt) {
		return nil, ErrCodeExpired
	}

	if verification.Attempts >= verificationMaxAttempts {
		return nil, ErrTooManyAttempts
	}

	if subtle.ConstantTimeCompare([]byte(verification.Code), []byte(req.Code)) != 1 {
		verification.Attempts++
		{
			if err := service.verification.Save(ctx, verification); err != nil {
				return nil, err
			}
		}

		if verification.Attempts >= verificationMaxAttempts {
			return nil, ErrTooManyAttempts
		}

		return nil, ErrInvalidVerification
	}

	pages, err := service.page.Form(ctx, application.TournamentID, application.CategoryID)
	{
		if err != nil {
			return nil, err
		}
	}

	if err := service.repository.MarkEmailVerified(ctx, application.ID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAlreadyVerified
		}

		return nil, err
	}

	cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := service.verification.Delete(cleanupCtx, application.ID); err != nil {
		log.Printf("Application %d verification cleanup error: %v", application.ID, err)
	}

	return pages, nil
}
