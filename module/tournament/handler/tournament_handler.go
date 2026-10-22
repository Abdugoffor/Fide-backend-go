package tournament_handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"auth-login/helper"
	auth_dto "auth-login/module/auth/dto"
	auth_handler "auth-login/module/auth/handler"
	auth_service "auth-login/module/auth/service"
	tournament_dto "auth-login/module/tournament/dto"
	tournament_service "auth-login/module/tournament/service"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
)

type tournamentHandler struct {
	service *tournament_service.TournamentService
}

func NewTournamentHandler(router *httprouter.Router, group string, db *pgxpool.Pool) {
	handler := &tournamentHandler{
		service: tournament_service.NewTournamentService(db),
	}

	auth := auth_service.NewAuthService(db, helper.ENV("JWT_SECRET"))

	roles := []string{auth_dto.RoleSuperadmin, auth_dto.RoleAdmin}

	routes := group + "/tournaments"
	{
		router.POST(routes, auth_handler.RequireAuth(auth, handler.Create, roles...))
		router.GET(routes, auth_handler.RequireAuth(auth, handler.List))
		router.GET(routes+"/:id", auth_handler.RequireAuth(auth, handler.Show))
		router.PUT(routes+"/:id", auth_handler.RequireAuth(auth, handler.Update, roles...))
		router.DELETE(routes+"/:id", auth_handler.RequireAuth(auth, handler.Delete, roles...))
	}
}

func (handler *tournamentHandler) Create(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req tournament_dto.Create
	{
		if !parseBody(w, r, &req) {
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

func (handler *tournamentHandler) List(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	query := r.URL.Query()

	filter := tournament_dto.Filter{
		Limit:       10,
		Name:        strings.TrimSpace(query.Get("name")),
		Description: strings.TrimSpace(query.Get("description")),
		Status:      query.Get("status"),
		Direction:   query.Get("direction"),
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

	if filter.Status != "" && filter.Status != "planned" && filter.Status != "ongoing" && filter.Status != "completed" && filter.Status != "cancelled" {
		helper.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid status"})
		return
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

func (handler *tournamentHandler) Show(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
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

func (handler *tournamentHandler) Update(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	id, ok := helper.ParseID(w, params)
	{
		if !ok {
			return
		}
	}

	var body tournament_dto.Create
	{
		if !parseBody(w, r, &body) {
			return
		}
	}

	req := tournament_dto.Update(body)

	res, err := handler.service.Update(r.Context(), id, &req)
	{
		if err != nil {
			writeServiceError(w, err)
			return
		}
	}

	helper.JSON(w, http.StatusOK, res)
}

func (handler *tournamentHandler) Delete(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	id, ok := helper.ParseID(w, params)
	{
		if !ok {
			return
		}

		if err := handler.service.Delete(r.Context(), id); err != nil {
			writeServiceError(w, err)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseBody(w http.ResponseWriter, r *http.Request, req *tournament_dto.Create) bool {
	{
		if err := helper.DecodeJSON(w, r, req); err != nil {
			helper.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return false
		}
	}

	req.Name = strings.TrimSpace(req.Name)

	req.Description = strings.TrimSpace(req.Description)

	req.Location = strings.TrimSpace(req.Location)

	if req.Status == "" {
		req.Status = tournament_dto.StatusPlanned
	}

	if errs := helper.Validate(req); errs != nil {
		helper.JSON(w, http.StatusBadRequest, map[string]any{"errors": errs})
		return false
	}

	return true
}

func writeServiceError(w http.ResponseWriter, err error) {
	{
		if errors.Is(err, pgx.ErrNoRows) {
			helper.JSON(w, http.StatusNotFound, map[string]string{"error": "tournament not found"})
			return
		}

		if errors.Is(err, tournament_service.ErrInvalidDates) {
			helper.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
	}

	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) && pgErr.Code == "23503" && pgErr.ConstraintName == "categories_tournament_fk" {
		helper.JSON(w, http.StatusConflict, map[string]string{"error": "tournament has categories"})
		return
	}

	if errors.As(err, &pgErr) && pgErr.Code == "23503" && pgErr.ConstraintName == "pages_tournament_fk" {
		helper.JSON(w, http.StatusConflict, map[string]string{"error": "tournament has pages"})
		return
	}

	if errors.As(err, &pgErr) && pgErr.Code == "23503" && pgErr.ConstraintName == "applications_tournament_fk" {
		helper.JSON(w, http.StatusConflict, map[string]string{"error": "tournament has applications"})
		return
	}

	log.Printf("Tournament service error: %v", err)

	helper.JSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
}
