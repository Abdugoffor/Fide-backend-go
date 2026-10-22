package application_handler

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"auth-login/helper"
	application_dto "auth-login/module/application/dto"
	application_service "auth-login/module/application/service"
	page_service "auth-login/module/page/service"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
)

type applicationHandler struct {
	service *application_service.ApplicationService
}

func NewApplicationHandler(router *httprouter.Router, group string, db *pgxpool.Pool) {
	handler := &applicationHandler{
		service: application_service.NewApplicationService(db),
	}

	routes := group + "/applications"
	{
		router.POST(routes, handler.Create)
		router.POST(routes+"/verify-email", handler.VerifyEmail)
	}
}

func (handler *applicationHandler) Create(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req application_dto.Create
	{
		if err := helper.DecodeJSON(w, r, &req); err != nil {
			helper.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
	}

	req.Name = strings.TrimSpace(req.Name)

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	req.BirthDate = strings.TrimSpace(req.BirthDate)
	{
		if errs := helper.Validate(req); errs != nil {
			helper.JSON(w, http.StatusBadRequest, map[string]any{"errors": errs})
			return
		}
	}

	res, err := handler.service.Create(r.Context(), &req)
	{
		if err != nil {
			writeServiceError(w, err)
			return
		}
	}

	helper.JSON(w, http.StatusCreated, res)
}

func (handler *applicationHandler) VerifyEmail(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req application_dto.VerifyEmail
	{
		if err := helper.DecodeJSON(w, r, &req); err != nil {
			helper.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	req.Code = strings.TrimSpace(req.Code)
	{
		if errs := helper.Validate(req); errs != nil {
			helper.JSON(w, http.StatusBadRequest, map[string]any{"errors": errs})
			return
		}
	}

	res, err := handler.service.VerifyEmail(r.Context(), &req)
	{
		if err != nil {
			writeServiceError(w, err)
			return
		}
	}

	helper.JSON(w, http.StatusOK, res)
}

func writeServiceError(w http.ResponseWriter, err error) {
	{
		if errors.Is(err, pgx.ErrNoRows) {
			helper.JSON(w, http.StatusNotFound, map[string]string{"error": "application not found"})
			return
		}

		if errors.Is(err, page_service.ErrTournamentNotFound) || errors.Is(err, page_service.ErrCategoryNotFound) {
			helper.JSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}

		if errors.Is(err, application_service.ErrInvalidApplication) || errors.Is(err, application_service.ErrInvalidBirthDate) ||
			errors.Is(err, application_service.ErrInvalidVerification) || errors.Is(err, application_service.ErrCodeExpired) ||
			errors.Is(err, page_service.ErrCategoryTournament) || errors.Is(err, page_service.ErrInvalidPage) {
			helper.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		if errors.Is(err, application_service.ErrAlreadyVerified) || errors.Is(err, page_service.ErrInactiveCategory) {
			helper.JSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
			return
		}

		if errors.Is(err, application_service.ErrTooManyAttempts) {
			helper.JSON(w, http.StatusTooManyRequests, map[string]string{"error": err.Error()})
			return
		}
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		helper.JSON(w, http.StatusConflict, map[string]string{"error": "tournament or category changed; retry the request"})
		return
	}

	log.Printf("Application service error: %v", err)

	helper.JSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
}
