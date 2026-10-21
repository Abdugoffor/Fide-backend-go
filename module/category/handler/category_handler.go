package category_handler

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
	category_dto "auth-login/module/category/dto"
	category_service "auth-login/module/category/service"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
)

type categoryHandler struct {
	service *category_service.CategoryService
}

func NewCategoryHandler(router *httprouter.Router, group string, db *pgxpool.Pool) {
	handler := &categoryHandler{
		service: category_service.NewCategoryService(db),
	}

	auth := auth_service.NewAuthService(db, helper.ENV("JWT_SECRET"))

	roles := []string{auth_dto.RoleSuperadmin, auth_dto.RoleAdmin}

	routes := group + "/categories"
	{
		router.GET(routes, auth_handler.RequireAuth(auth, handler.List))
		router.GET(routes+"-with-childr", auth_handler.RequireAuth(auth, handler.WithChildren))
		router.GET(routes+"/:id", auth_handler.RequireAuth(auth, handler.Show))
		router.POST(routes, auth_handler.RequireAuth(auth, handler.Create, roles...))
		router.PUT(routes+"/:id", auth_handler.RequireAuth(auth, handler.Update, roles...))
		router.DELETE(routes+"/:id", auth_handler.RequireAuth(auth, handler.Delete, roles...))
	}
}

func (handler *categoryHandler) Create(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req category_dto.Create
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

func (handler *categoryHandler) List(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	query := r.URL.Query()

	filter := category_dto.Filter{
		Limit: 10,
		Name:  strings.TrimSpace(query.Get("name")),
	}

	if query.Has("cursor") {
		cursor, err := strconv.ParseInt(query.Get("cursor"), 10, 64)
		{
			if err != nil || cursor < 1 {
				helper.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid cursor"})
				return
			}
		}

		filter.Cursor = cursor
	}

	if query.Has("limit") {
		limit, err := strconv.Atoi(query.Get("limit"))
		{
			if err != nil || limit < 1 || limit > 100 {
				helper.JSON(w, http.StatusBadRequest, map[string]string{"error": "limit must be between 1 and 100"})
				return
			}
		}

		filter.Limit = limit
	}

	if query.Has("tournament_id") {
		tournamentID, err := strconv.ParseInt(query.Get("tournament_id"), 10, 64)
		{
			if err != nil || tournamentID < 1 {
				helper.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid tournament_id"})
				return
			}
		}

		filter.TournamentID = tournamentID
	}

	if query.Has("parent_id") {
		parentID, err := strconv.ParseInt(query.Get("parent_id"), 10, 64)
		{
			if err != nil || parentID < 0 {
				helper.JSON(w, http.StatusBadRequest, map[string]string{"error": "parent_id must be zero or a positive integer"})
				return
			}
		}

		filter.ParentID = &parentID
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

func (handler *categoryHandler) Show(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
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

func (handler *categoryHandler) WithChildren(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	raw := r.URL.Query().Get("tournament_id")

	if raw == "" {
		helper.JSON(w, http.StatusBadRequest, map[string]string{"error": "tournament_id is required"})
		return
	}

	tournamentID, err := strconv.ParseInt(raw, 10, 64)
	{
		if err != nil || tournamentID < 1 {
			helper.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid tournament_id"})
			return
		}
	}

	res, err := handler.service.WithChildren(r.Context(), tournamentID)
	{
		if err != nil {
			writeServiceError(w, err)
			return
		}
	}

	helper.JSON(w, http.StatusOK, res)
}

func (handler *categoryHandler) Update(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	id, ok := helper.ParseID(w, params)
	{
		if !ok {
			return
		}
	}

	var body category_dto.Create
	{
		if !parseBody(w, r, &body) {
			return
		}
	}

	req := category_dto.Update(body)

	res, err := handler.service.Update(r.Context(), id, &req)
	{
		if err != nil {
			writeServiceError(w, err)
			return
		}
	}

	helper.JSON(w, http.StatusOK, res)
}

func (handler *categoryHandler) Delete(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
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

func parseBody(w http.ResponseWriter, r *http.Request, req *category_dto.Create) bool {
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
			helper.JSON(w, http.StatusNotFound, map[string]string{"error": "category not found"})
			return
		}

		if errors.Is(err, category_service.ErrTournamentNotFound) || errors.Is(err, category_service.ErrParentNotFound) {
			helper.JSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}

		if errors.Is(err, category_service.ErrInvalidCategory) || errors.Is(err, category_service.ErrParentTournament) || errors.Is(err, category_service.ErrParentCycle) {
			helper.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		if errors.Is(err, category_service.ErrHierarchyConflict) || errors.Is(err, category_service.ErrConcurrentChange) || errors.Is(err, category_service.ErrPageConflict) {
			helper.JSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
			return
		}
	}

	log.Printf("Category service error: %v", err)

	helper.JSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
}
