package application_handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"auth-login/helper"
	application_dto "auth-login/module/application/dto"
	application_service "auth-login/module/application/service"
	auth_handler "auth-login/module/auth/handler"
	auth_service "auth-login/module/auth/service"
	page_service "auth-login/module/page/service"
	tournament_dto "auth-login/module/tournament/dto"

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

	auth := auth_service.NewAuthService(db, helper.ENV("JWT_SECRET"))

	routes := group + "/applications"
	{
		router.POST(routes, handler.Create)
		router.GET(routes, auth_handler.RequireAuth(auth, handler.List))
		router.GET(routes+"/:id", auth_handler.RequireAuth(auth, handler.Show))
		router.POST(routes+"/verify-email", handler.VerifyEmail)
		router.POST(routes+"/data/:slug", handler.SaveData)
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

func (handler *applicationHandler) List(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	query := r.URL.Query()

	filter := application_dto.Filter{
		Limit:     10,
		Email:     strings.TrimSpace(query.Get("email")),
		Name:      strings.TrimSpace(query.Get("name")),
		Direction: query.Get("direction"),
	}

	if filter.Direction == "" {
		filter.Direction = tournament_dto.DirectionNext
	}

	if filter.Direction != tournament_dto.DirectionNext && filter.Direction != tournament_dto.DirectionPrev {
		helper.JSON(w, http.StatusBadRequest, map[string]string{"error": "direction must be next or prev"})
		return
	}

	if raw := query.Get("cursor"); raw != "" {
		cursor, err := strconv.ParseInt(raw, 10, 64)
		{
			if err != nil || cursor < 1 {
				helper.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid cursor"})
				return
			}
		}

		filter.Cursor = cursor
	}

	if filter.Direction == tournament_dto.DirectionPrev && filter.Cursor == 0 {
		helper.JSON(w, http.StatusBadRequest, map[string]string{"error": "cursor is required for prev direction"})
		return
	}

	if raw := query.Get("limit"); raw != "" {
		limit, err := strconv.Atoi(raw)
		{
			if err != nil || limit < 1 || limit > 100 {
				helper.JSON(w, http.StatusBadRequest, map[string]string{"error": "limit must be between 1 and 100"})
				return
			}
		}

		filter.Limit = limit
	}

	if raw := query.Get("tournament_id"); raw != "" {
		tournamentID, err := strconv.ParseInt(raw, 10, 64)
		{
			if err != nil || tournamentID < 1 {
				helper.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid tournament_id"})
				return
			}
		}

		filter.TournamentID = tournamentID
	}

	if raw := query.Get("category_id"); raw != "" {
		categoryID, err := strconv.ParseInt(raw, 10, 64)
		{
			if err != nil || categoryID < 1 {
				helper.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid category_id"})
				return
			}
		}

		filter.CategoryID = categoryID
	}

	res, err := handler.service.List(r.Context(), &filter)
	{
		if err != nil {
			writeServiceError(w, err)
			return
		}
	}

	helper.JSON(w, http.StatusOK, res)
}

func (handler *applicationHandler) Show(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	id, ok := helper.ParseID(w, params)
	{
		if !ok {
			return
		}
	}

	res, err := handler.service.Show(r.Context(), id)
	{
		if err != nil {
			writeServiceError(w, err)
			return
		}
	}

	helper.JSON(w, http.StatusOK, res)
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

func (handler *applicationHandler) SaveData(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	var req application_dto.SaveData
	{
		if err := helper.DecodeJSON(w, r, &req); err != nil {
			helper.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
	}

	req.Slug = strings.TrimSpace(params.ByName("slug"))
	{
		if errs := helper.Validate(req); errs != nil {
			helper.JSON(w, http.StatusBadRequest, map[string]any{"errors": errs})
			return
		}
	}

	res, err := handler.service.SaveData(r.Context(), &req)
	{
		if err != nil {
			writeServiceError(w, err)
			return
		}
	}

	status := http.StatusOK
	if res.Created {
		status = http.StatusCreated
	}

	helper.JSON(w, status, res)
}

func writeServiceError(w http.ResponseWriter, err error) {
	var validationErr *application_service.DataValidationError
	if errors.As(err, &validationErr) {
		helper.JSON(w, http.StatusBadRequest, map[string]any{"errors": validationErr.Fields})
		return
	}

	{
		if errors.Is(err, application_service.ErrPageNotFound) {
			helper.JSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}

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
