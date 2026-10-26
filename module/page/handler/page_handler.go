package page_handler

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
	page_dto "auth-login/module/page/dto"
	page_service "auth-login/module/page/service"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
)

type pageHandler struct {
	service *page_service.PageService
}

func NewPageHandler(router *httprouter.Router, group string, db *pgxpool.Pool) {
	handler := &pageHandler{
		service: page_service.NewPageService(db),
	}

	auth := auth_service.NewAuthService(db, helper.ENV("JWT_SECRET"))

	roles := []string{auth_dto.RoleSuperadmin, auth_dto.RoleAdmin}

	routes := group + "/pages"
	{
		router.POST(routes, auth_handler.RequireAuth(auth, handler.Create, roles...))
		router.GET(routes, auth_handler.RequireAuth(auth, handler.List))
		router.GET(routes+"-form", auth_handler.RequireAuth(auth, handler.Form))
		router.GET(routes+"/:id", auth_handler.RequireAuth(auth, handler.Show))
		router.PUT(routes+"/:id", auth_handler.RequireAuth(auth, handler.Update, roles...))
		router.DELETE(routes+"/:id", auth_handler.RequireAuth(auth, handler.Delete, roles...))
	}
}

func (handler *pageHandler) Create(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req page_dto.Create
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

func (handler *pageHandler) List(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	query := r.URL.Query()

	filter := page_dto.Filter{
		Name: strings.TrimSpace(query.Get("name")),
	}

	if query.Has("tournament_id") {
		value, err := strconv.ParseInt(query.Get("tournament_id"), 10, 64)
		{
			if err != nil || value < 1 {
				helper.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid tournament_id"})
				return
			}
		}

		filter.TournamentID = value
	}

	if query.Has("category_id") {
		value, err := strconv.ParseInt(query.Get("category_id"), 10, 64)
		{
			if err != nil || value < 1 {
				helper.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid category_id"})
				return
			}
		}

		filter.CategoryID = value
	}

	if query.Has("is_active") {
		raw := query.Get("is_active")

		if raw != "true" && raw != "false" {
			helper.JSON(w, http.StatusBadRequest, map[string]string{"error": "is_active must be true or false"})
			return
		}

		isActive := raw == "true"

		filter.IsActive = &isActive
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

func (handler *pageHandler) Form(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	query := r.URL.Query()

	tournamentID, err := strconv.ParseInt(query.Get("tournament_id"), 10, 64)
	{
		if err != nil || tournamentID < 1 {
			helper.JSON(w, http.StatusBadRequest, map[string]string{"error": "tournament_id must be a positive integer"})
			return
		}
	}

	categoryID, err := strconv.ParseInt(query.Get("category_id"), 10, 64)
	{
		if err != nil || categoryID < 1 {
			helper.JSON(w, http.StatusBadRequest, map[string]string{"error": "category_id must be a positive integer"})
			return
		}
	}

	res, err := handler.service.Form(r.Context(), tournamentID, categoryID)
	{
		if err != nil {
			writeServiceError(w, err)
			return
		}
	}

	helper.JSON(w, http.StatusOK, res)
}

func (handler *pageHandler) Show(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
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

func (handler *pageHandler) Update(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	id, ok := helper.ParseID(w, params)
	{
		if !ok {
			return
		}
	}

	var body page_dto.Create
	{
		if !parseBody(w, r, &body) {
			return
		}
	}

	req := page_dto.Update(body)

	res, err := handler.service.Update(r.Context(), id, &req)
	{
		if err != nil {
			writeServiceError(w, err)
			return
		}
	}

	helper.JSON(w, http.StatusOK, res)
}

func (handler *pageHandler) Delete(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
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

func parseBody(w http.ResponseWriter, r *http.Request, req *page_dto.Create) bool {
	{
		if err := helper.DecodeJSON(w, r, req); err != nil {
			helper.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return false
		}
	}

	req.Name = strings.TrimSpace(req.Name)

	if errs := helper.Validate(req); errs != nil {
		helper.JSON(w, http.StatusBadRequest, map[string]any{"errors": errs})
		return false
	}

	return true
}

func writeServiceError(w http.ResponseWriter, err error) {
	{
		if errors.Is(err, pgx.ErrNoRows) {
			helper.JSON(w, http.StatusNotFound, map[string]string{"error": "page not found"})
			return
		}

		if errors.Is(err, page_service.ErrTournamentNotFound) || errors.Is(err, page_service.ErrCategoryNotFound) {
			helper.JSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}

		if errors.Is(err, page_service.ErrInvalidPage) || errors.Is(err, page_service.ErrCategoryTournament) {
			helper.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		if errors.Is(err, page_service.ErrInactiveCategory) {
			helper.JSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
			return
		}
	}

	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			if pgErr.ConstraintName == "pages_slug_unique" {
				helper.JSON(w, http.StatusConflict, map[string]string{"error": "page slug already exists"})
				return
			}

			helper.JSON(w, http.StatusConflict, map[string]string{"error": "page sort already exists in this category"})
			return
		}

		if pgErr.Code == "23503" {
			if pgErr.ConstraintName == "fields_page_id_fkey" {
				helper.JSON(w, http.StatusConflict, map[string]string{"error": "page has fields; remove or move its fields first"})
				return
			}

			helper.JSON(w, http.StatusConflict, map[string]string{"error": "tournament or category changed; retry the request"})
			return
		}
	}

	log.Printf("Page service error: %v", err)

	helper.JSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
}
